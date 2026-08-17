// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package identitybinding

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/hash"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
)

func newTestService(t *testing.T) (*Service, *pg.Client) {
	t.Helper()

	client := test.PGClient(t)
	baseURL, err := baseurl.Parse("https://console.example.com")
	require.NoError(t, err)

	return NewService(client, baseURL), client
}

func seedIdentity(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
) gid.GID {
	t.Helper()

	identityID := gid.New(gid.NilTenant, coredata.IdentityEntityType)
	emailAddress, err := mail.ParseAddr(fmt.Sprintf("%s@example.com", identityID))
	require.NoError(t, err)

	now := time.Now().UTC()

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return (&coredata.Identity{
			ID:                   identityID,
			EmailAddress:         emailAddress,
			FullName:             "Probot Binding Test",
			EmailAddressVerified: true,
			CreatedAt:            now,
			UpdatedAt:            now,
		}).Insert(ctx, tx)
	}))
	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			return (&coredata.Identity{ID: identityID}).Delete(ctx, tx)
		})
	})

	return identityID
}

func testOrganizationID() gid.GID {
	return gid.New(gid.NewTenantID(), coredata.OrganizationEntityType)
}

func issueToken(
	t *testing.T,
	ctx context.Context,
	service *Service,
	subject Subject,
) string {
	t.Helper()

	bindURL, err := service.BindURL(ctx, subject, testOrganizationID())
	require.NoError(t, err)
	parsed, err := url.Parse(bindURL)
	require.NoError(t, err)

	token := parsed.Query().Get("token")
	require.NotEmpty(t, token)

	t.Cleanup(func() {
		_ = service.pg.WithConn(
			context.Background(),
			func(ctx context.Context, conn pg.Querier) error {
				_, err := conn.Exec(
					ctx,
					`DELETE FROM probot_identity_binding_challenges WHERE hashed_token = @hashed_token`,
					pgx.StrictNamedArgs{"hashed_token": hash.SHA256String(token)},
				)

				return err
			},
		)
	})

	return token
}

func TestClipDisplayName(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "acme", clipDisplayName("  acme  "))
	assert.Equal(t, "", clipDisplayName("   "))

	longName := strings.Repeat("é", displayNameMaxRunes+8)
	clipped := clipDisplayName(longName)
	assert.Equal(t, displayNameMaxRunes, len([]rune(clipped)))
	assert.Equal(t, strings.Repeat("é", displayNameMaxRunes), clipped)
}

func TestSubjectValidate(t *testing.T) {
	t.Parallel()

	assert.ErrorIs(t, (Subject{}).Validate(), ErrInvalidSubject)
	assert.ErrorIs(
		t,
		(Subject{Provider: "slack"}).Validate(),
		ErrInvalidSubject,
	)
	assert.NoError(
		t,
		(Subject{Provider: "email", ExternalUserID: "user@example.com"}).Validate(),
	)
}

func TestServiceBindingLifecycle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	service, client := newTestService(t)
	identityID := seedIdentity(t, ctx, client)
	subject := Subject{
		Provider:           "slack",
		ExternalTenantID:   "T123",
		ExternalUserID:     "U123",
		ExternalTenantName: "acme",
		ExternalUserName:   "ada",
	}
	token := issueToken(t, ctx, service, subject)

	preview, err := service.Preview(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, subject, *preview)

	binding, err := service.Confirm(ctx, identityID, token)
	require.NoError(t, err)
	assert.Equal(t, identityID, binding.IdentityID)
	assert.Equal(t, subject.Provider, binding.Provider)
	assert.Equal(t, subject.ExternalTenantID, binding.ExternalTenantID)
	assert.Equal(t, subject.ExternalUserID, binding.ExternalUserID)
	assert.Equal(t, subject.ExternalTenantName, binding.ExternalTenantName)
	assert.Equal(t, subject.ExternalUserName, binding.ExternalUserName)

	loaded, err := service.Lookup(ctx, subject)
	require.NoError(t, err)
	assert.Equal(t, binding.ID, loaded.ID)
	assert.Equal(t, subject.ExternalTenantName, loaded.ExternalTenantName)
	assert.Equal(t, subject.ExternalUserName, loaded.ExternalUserName)

	_, err = service.Preview(ctx, token)
	require.ErrorIs(t, err, ErrChallengeAlreadyUsed)
	_, err = service.Confirm(ctx, identityID, token)
	require.ErrorIs(t, err, ErrChallengeAlreadyUsed)

	require.NoError(t, service.Delete(ctx, identityID, binding.ID))
	_, err = service.Lookup(ctx, subject)
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)
}

func TestServiceConfirmNotifiesHandler(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	service, client := newTestService(t)
	identityID := seedIdentity(t, ctx, client)
	subject := Subject{
		Provider:         "slack",
		ExternalTenantID: "T-notify",
		ExternalUserID:   "U-notify",
	}
	token := issueToken(t, ctx, service, subject)
	handler := &recordingConfirmedHandler{}
	service.SetBindingConfirmedHandler(handler)

	_, err := service.Confirm(ctx, identityID, token)
	require.NoError(t, err)
	assert.Equal(t, subject, handler.subject)
}

type recordingConfirmedHandler struct {
	subject Subject
}

func (h *recordingConfirmedHandler) BindingConfirmed(
	_ context.Context,
	subject Subject,
) error {
	h.subject = subject

	return nil
}

func TestServiceBindURLIsOpaque(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	service, _ := newTestService(t)
	subject := Subject{
		Provider:         "email",
		ExternalTenantID: "",
		ExternalUserID:   "private.user@example.com",
	}

	organizationID := testOrganizationID()
	bindURL, err := service.BindURL(ctx, subject, organizationID)
	require.NoError(t, err)
	assert.NotContains(t, bindURL, subject.ExternalUserID)
	assert.Contains(
		t,
		bindURL,
		"/organizations/"+organizationID.String()+"/employee/bind",
	)
}

func TestServiceBindURLRequiresOrganization(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	service, _ := newTestService(t)

	_, err := service.BindURL(
		ctx,
		Subject{Provider: "slack", ExternalUserID: "U1"},
		gid.Nil,
	)
	require.ErrorIs(t, err, ErrOrganizationRequired)
}

func TestServiceListByIdentity(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	service, client := newTestService(t)
	identityID := seedIdentity(t, ctx, client)
	otherID := seedIdentity(t, ctx, client)
	subject := Subject{
		Provider:         "slack",
		ExternalTenantID: fmt.Sprintf("T-%s", identityID),
		ExternalUserID:   "U-list",
	}
	token := issueToken(t, ctx, service, subject)
	binding, err := service.Confirm(ctx, identityID, token)
	require.NoError(t, err)

	listed, err := service.ListByIdentity(ctx, identityID)
	require.NoError(t, err)
	require.Len(t, listed, 1)
	assert.Equal(t, binding.ID, listed[0].ID)

	otherListed, err := service.ListByIdentity(ctx, otherID)
	require.NoError(t, err)
	assert.Empty(t, otherListed)
}

func TestServiceRejectsConflictsAndCrossIdentityReplay(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	service, client := newTestService(t)
	firstIdentityID := seedIdentity(t, ctx, client)
	secondIdentityID := seedIdentity(t, ctx, client)
	subject := Subject{
		Provider:         "slack",
		ExternalTenantID: fmt.Sprintf("T-%s", firstIdentityID),
		ExternalUserID:   "U1",
	}
	firstToken := issueToken(t, ctx, service, subject)
	binding, err := service.Confirm(ctx, firstIdentityID, firstToken)
	require.NoError(t, err)

	_, err = service.Confirm(ctx, secondIdentityID, firstToken)
	require.ErrorIs(t, err, ErrChallengeAlreadyUsed)

	secondToken := issueToken(t, ctx, service, subject)
	_, err = service.Confirm(ctx, secondIdentityID, secondToken)
	require.ErrorIs(t, err, ErrAlreadyBound)

	otherSubject := subject
	otherSubject.ExternalUserID = "U2"
	thirdToken := issueToken(t, ctx, service, otherSubject)
	_, err = service.Confirm(ctx, firstIdentityID, thirdToken)
	require.ErrorIs(t, err, ErrAlreadyBound)

	require.NoError(t, service.Delete(ctx, firstIdentityID, binding.ID))
}

func TestServiceRejectsExpiredChallenge(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	service, client := newTestService(t)
	identityID := seedIdentity(t, ctx, client)
	now := time.Now().UTC()
	service.now = func() time.Time { return now }
	token := issueToken(t, ctx, service, Subject{
		Provider:       "email",
		ExternalUserID: "user@example.com",
	})
	service.now = func() time.Time { return now.Add(challengeExpiry + time.Second) }

	_, err := service.Preview(ctx, token)
	require.ErrorIs(t, err, ErrChallengeExpired)
	_, err = service.Confirm(ctx, identityID, token)
	require.ErrorIs(t, err, ErrChallengeExpired)
}

func TestServiceBindURLPurgesRetainedExpiredChallenges(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	service, client := newTestService(t)
	now := time.Now().UTC()
	service.now = func() time.Time { return now }
	staleToken := "stale-binding-challenge"
	stale := &coredata.ProbotIdentityBindingChallenge{
		HashedToken:    hash.SHA256String(staleToken),
		Provider:       "email",
		ExternalUserID: "stale@example.com",
		ExpiresAt:      now.Add(-challengeRetention - time.Minute),
		CreatedAt:      now.Add(-challengeRetention - time.Hour),
	}

	require.NoError(t, client.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return stale.Insert(ctx, conn)
		},
	))

	_ = issueToken(t, ctx, service, Subject{
		Provider:       "email",
		ExternalUserID: "current@example.com",
	})

	err := client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return stale.LoadByHashedToken(ctx, conn, stale.HashedToken)
	})
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)
}

func TestServiceConcurrentConfirmationCreatesOneBinding(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	service, client := newTestService(t)
	identityID := seedIdentity(t, ctx, client)
	token := issueToken(t, ctx, service, Subject{
		Provider:         "slack",
		ExternalTenantID: fmt.Sprintf("T-%s", identityID),
		ExternalUserID:   "U-concurrent",
	})

	const workers = 4

	var wg sync.WaitGroup

	ids := make(chan gid.GID, workers)

	errs := make(chan error, workers)
	for range workers {
		wg.Go(func() {
			binding, err := service.Confirm(ctx, identityID, token)
			if err != nil {
				errs <- err

				return
			}

			ids <- binding.ID
		})
	}

	wg.Wait()
	close(ids)
	close(errs)

	for err := range errs {
		require.ErrorIs(t, err, ErrChallengeAlreadyUsed)
	}

	var firstID gid.GID

	successes := 0
	for id := range ids {
		successes++

		if firstID == (gid.GID{}) {
			firstID = id
		}

		assert.Equal(t, firstID, id)
	}

	require.Equal(t, 1, successes)
	require.NotEqual(t, gid.GID{}, firstID)
	require.NoError(t, service.Delete(ctx, identityID, firstID))
}

func TestDeleteByExternalTenant_RemovesProviderTenantRows(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	service, client := newTestService(t)
	identityID := seedIdentity(t, ctx, client)
	otherIdentityID := seedIdentity(t, ctx, client)
	target := Subject{
		Provider:         "slack",
		ExternalTenantID: fmt.Sprintf("T-purge-%s", identityID),
		ExternalUserID:   "U-purge",
	}
	other := Subject{
		Provider:         "slack",
		ExternalTenantID: fmt.Sprintf("T-keep-%s", otherIdentityID),
		ExternalUserID:   "U-keep",
	}
	targetToken := issueToken(t, ctx, service, target)
	_, err := service.Confirm(ctx, identityID, targetToken)
	require.NoError(t, err)

	pendingToken := issueToken(
		t,
		ctx,
		service,
		Subject{
			Provider:         target.Provider,
			ExternalTenantID: target.ExternalTenantID,
			ExternalUserID:   "U-pending",
		},
	)
	otherToken := issueToken(t, ctx, service, other)
	otherBinding, err := service.Confirm(ctx, otherIdentityID, otherToken)
	require.NoError(t, err)

	require.NoError(
		t,
		client.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				return DeleteByExternalTenant(
					ctx,
					tx,
					target.Provider,
					target.ExternalTenantID,
				)
			},
		),
	)

	_, err = service.Lookup(ctx, target)
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)
	_, err = service.Preview(ctx, targetToken)
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)
	_, err = service.Preview(ctx, pendingToken)
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)

	loaded, err := service.Lookup(ctx, other)
	require.NoError(t, err)
	assert.Equal(t, otherBinding.ID, loaded.ID)
	require.NoError(t, service.Delete(ctx, otherIdentityID, otherBinding.ID))
}

func TestDeleteByExternalTenant_RejectsEmptySubject(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	_, client := newTestService(t)

	err := client.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return DeleteByExternalTenant(ctx, conn, "slack", "")
		},
	)
	require.ErrorIs(t, err, ErrInvalidSubject)

	err = client.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return DeleteByExternalTenant(ctx, conn, "", "T-1")
		},
	)
	require.ErrorIs(t, err, ErrInvalidSubject)
}

func TestServiceDeleteHidesOtherIdentityBinding(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	service, client := newTestService(t)
	ownerID := seedIdentity(t, ctx, client)
	otherID := seedIdentity(t, ctx, client)
	token := issueToken(t, ctx, service, Subject{
		Provider:         "slack",
		ExternalTenantID: fmt.Sprintf("T-%s", ownerID),
		ExternalUserID:   "U-delete",
	})
	binding, err := service.Confirm(ctx, ownerID, token)
	require.NoError(t, err)

	err = service.Delete(ctx, otherID, binding.ID)
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)
	require.NoError(t, service.Delete(ctx, ownerID, binding.ID))
}

func TestServiceConfirmInvalidatesSiblingChallenges(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	service, client := newTestService(t)
	identityID := seedIdentity(t, ctx, client)
	subject := Subject{
		Provider:         "slack",
		ExternalTenantID: fmt.Sprintf("T-%s", identityID),
		ExternalUserID:   "U-siblings",
	}
	firstToken := issueToken(t, ctx, service, subject)
	secondToken := issueToken(t, ctx, service, subject)

	binding, err := service.Confirm(ctx, identityID, firstToken)
	require.NoError(t, err)
	assert.Equal(t, identityID, binding.IdentityID)

	_, err = service.Preview(ctx, firstToken)
	require.ErrorIs(t, err, ErrChallengeAlreadyUsed)
	_, err = service.Preview(ctx, secondToken)
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)
	_, err = service.Confirm(ctx, identityID, secondToken)
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)
}
