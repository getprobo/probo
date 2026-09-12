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

package oauth2

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/hash"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/uri"
)

func TestRevokeLeakedManualAccessToken_RevokesAndNotifiesOnce(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	pgClient := test.PGClient(t)
	baseURL := uri.URI("https://us.probo.com")
	tokenValue := newManualAccessToken(baseURL)
	identityID := gid.New(gid.NilTenant, coredata.IdentityEntityType)
	tokenID := gid.New(identityID.TenantID(), coredata.OAuth2AccessTokenEntityType)
	emailAddress, err := mail.ParseAddr(fmt.Sprintf("%s@example.com", identityID))
	require.NoError(t, err)

	now := time.Now().UTC()
	identity := &coredata.Identity{
		ID:                   identityID,
		EmailAddress:         emailAddress,
		FullName:             "Secret Scanning Test",
		EmailAddressVerified: true,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	accessToken := &coredata.OAuth2AccessToken{
		ID:          tokenID,
		Name:        "Leaked test token",
		HashedValue: hash.SHA256String(tokenValue),
		IdentityID:  identityID,
		Resources:   []uri.URI{baseURL},
		Scopes:      coredata.OAuth2Scopes{"v1:org:read"},
		CreatedAt:   now,
		ExpiresAt:   now.Add(time.Hour),
	}

	require.NoError(
		t,
		pgClient.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				if err := identity.Insert(ctx, tx); err != nil {
					return fmt.Errorf("cannot insert identity: %w", err)
				}

				if err := accessToken.Insert(ctx, tx); err != nil {
					return fmt.Errorf("cannot insert access token: %w", err)
				}

				return nil
			},
		),
	)
	t.Cleanup(
		func() {
			_ = pgClient.WithTx(
				context.Background(),
				func(ctx context.Context, tx pg.Tx) error {
					if _, err := tx.Exec(
						ctx,
						"DELETE FROM emails WHERE recipient_email = @recipient_email",
						pgx.StrictNamedArgs{"recipient_email": emailAddress},
					); err != nil {
						return err
					}

					return (&coredata.Identity{ID: identityID}).Delete(ctx, tx)
				},
			)
		},
	)

	service := NewService(pgClient, nil, baseURL, log.NewLogger())
	revoked, err := service.RevokeLeakedManualAccessToken(ctx, tokenValue)
	require.NoError(t, err)
	assert.True(t, revoked)

	_, err = service.LoadAccessToken(ctx, tokenValue)
	assert.ErrorIs(t, err, coredata.ErrResourceNotFound)
	assert.Equal(t, 1, countNotificationEmails(t, pgClient, emailAddress.String()))

	revoked, err = service.RevokeLeakedManualAccessToken(ctx, tokenValue)
	require.NoError(t, err)
	assert.False(t, revoked)
	assert.Equal(t, 1, countNotificationEmails(t, pgClient, emailAddress.String()))
}

func TestRevokeLeakedManualAccessToken_IgnoresSelfHostedToken(t *testing.T) {
	t.Parallel()

	service := NewService(nil, nil, "https://probo.example.com", log.NewLogger())

	revoked, err := service.RevokeLeakedManualAccessToken(
		t.Context(),
		"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
	)

	require.NoError(t, err)
	assert.False(t, revoked)
}

func countNotificationEmails(t *testing.T, pgClient *pg.Client, recipientEmail string) int {
	t.Helper()

	var count int
	err := pgClient.WithConn(
		t.Context(),
		func(ctx context.Context, conn pg.Querier) error {
			err := conn.QueryRow(
				ctx,
				`SELECT COUNT(id) FROM emails WHERE recipient_email = @recipient_email AND subject = @subject`,
				pgx.StrictNamedArgs{
					"recipient_email": recipientEmail,
					"subject":         "Your Probo API token was revoked",
				},
			).Scan(&count)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("cannot count notification emails: %w", err)
			}

			return nil
		},
	)
	require.NoError(t, err)

	return count
}
