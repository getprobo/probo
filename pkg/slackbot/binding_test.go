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

package slackbot

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
)

func seedTestIdentity(t *testing.T, ctx context.Context, client *pg.Client) gid.GID {
	t.Helper()

	identityID := gid.New(gid.NilTenant, coredata.IdentityEntityType)
	now := time.Now().UTC()
	emailAddress, err := mail.ParseAddr(fmt.Sprintf("%s@example.com", identityID))
	require.NoError(t, err)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		identity := coredata.Identity{
			ID:                   identityID,
			EmailAddress:         emailAddress,
			FullName:             "Slack Bind Test",
			EmailAddressVerified: true,
			CreatedAt:            now,
			UpdatedAt:            now,
		}

		return identity.Insert(ctx, tx)
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			identity := coredata.Identity{ID: identityID}

			return identity.Delete(ctx, tx)
		})
	})

	return identityID
}

func deleteSlackBinding(t *testing.T, ctx context.Context, client *pg.Client, bindingID gid.GID) {
	t.Helper()

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		binding := coredata.SlackIdentityBinding{ID: bindingID}

		return binding.Delete(ctx, tx)
	}))
}

func TestBindingService_Confirm(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := test.PGClient(t)
	const secret = "binding-service-test-secret"

	baseURL, err := baseurl.Parse("https://console.example.com")
	require.NoError(t, err)

	service := NewBindingService(client, secret, baseURL)

	t.Run(
		"creates binding from valid token",
		func(t *testing.T) {
			t.Parallel()

			identityID := seedTestIdentity(t, ctx, client)
			teamID := fmt.Sprintf("T-%s", identityID.String())
			slackUserID := fmt.Sprintf("U-%s", identityID.String())

			token, err := newBindToken(secret, teamID, slackUserID)
			require.NoError(t, err)

			binding, err := service.Confirm(ctx, identityID, token)
			require.NoError(t, err)
			assert.Equal(t, teamID, binding.TeamID)
			assert.Equal(t, slackUserID, binding.SlackUserID)
			assert.Equal(t, identityID, binding.IdentityID)

			t.Cleanup(func() {
				deleteSlackBinding(t, ctx, client, binding.ID)
			})
		},
	)

	t.Run(
		"idempotent for same identity",
		func(t *testing.T) {
			t.Parallel()

			identityID := seedTestIdentity(t, ctx, client)
			teamID := fmt.Sprintf("T-idem-%s", identityID.String())
			slackUserID := fmt.Sprintf("U-idem-%s", identityID.String())

			token, err := newBindToken(secret, teamID, slackUserID)
			require.NoError(t, err)

			first, err := service.Confirm(ctx, identityID, token)
			require.NoError(t, err)

			second, err := service.Confirm(ctx, identityID, token)
			require.NoError(t, err)
			assert.Equal(t, first.ID, second.ID)

			t.Cleanup(func() {
				deleteSlackBinding(t, ctx, client, first.ID)
			})
		},
	)

	t.Run(
		"rejects when slack user already bound to another identity",
		func(t *testing.T) {
			t.Parallel()

			ownerID := seedTestIdentity(t, ctx, client)
			otherID := seedTestIdentity(t, ctx, client)
			teamID := fmt.Sprintf("T-conflict-%s", ownerID.String())
			slackUserID := fmt.Sprintf("U-conflict-%s", ownerID.String())

			ownerToken, err := newBindToken(secret, teamID, slackUserID)
			require.NoError(t, err)

			binding, err := service.Confirm(ctx, ownerID, ownerToken)
			require.NoError(t, err)

			t.Cleanup(func() {
				deleteSlackBinding(t, ctx, client, binding.ID)
			})

			// Fresh token for the same Slack user — distinct from token reuse.
			otherToken, err := newBindToken(secret, teamID, slackUserID)
			require.NoError(t, err)

			_, err = service.Confirm(ctx, otherID, otherToken)
			require.ErrorIs(t, err, ErrSlackIdentityAlreadyBound)
		},
	)

	t.Run(
		"rejects when identity already bound to another slack user in team",
		func(t *testing.T) {
			t.Parallel()

			identityID := seedTestIdentity(t, ctx, client)
			teamID := fmt.Sprintf("T-team-%s", identityID.String())

			firstToken, err := newBindToken(secret, teamID, fmt.Sprintf("U1-%s", identityID.String()))
			require.NoError(t, err)

			first, err := service.Confirm(ctx, identityID, firstToken)
			require.NoError(t, err)

			t.Cleanup(func() {
				deleteSlackBinding(t, ctx, client, first.ID)
			})

			secondToken, err := newBindToken(
				secret,
				teamID,
				fmt.Sprintf("U2-%s", identityID.String()),
			)
			require.NoError(t, err)

			_, err = service.Confirm(ctx, identityID, secondToken)
			require.ErrorIs(t, err, ErrSlackIdentityAlreadyBound)
		},
	)

	t.Run(
		"rejects reused token for a different identity",
		func(t *testing.T) {
			t.Parallel()

			ownerID := seedTestIdentity(t, ctx, client)
			otherID := seedTestIdentity(t, ctx, client)
			teamID := fmt.Sprintf("T-reuse-%s", ownerID.String())
			slackUserID := fmt.Sprintf("U-reuse-%s", ownerID.String())

			token, err := newBindToken(secret, teamID, slackUserID)
			require.NoError(t, err)

			binding, err := service.Confirm(ctx, ownerID, token)
			require.NoError(t, err)

			t.Cleanup(func() {
				deleteSlackBinding(t, ctx, client, binding.ID)
			})

			_, err = service.Confirm(ctx, otherID, token)
			require.ErrorIs(t, err, ErrBindTokenAlreadyUsed)
		},
	)
}
