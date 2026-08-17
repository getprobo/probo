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

package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/probot/identitybinding"
)

func TestInstallationInitiateURL(t *testing.T) {
	t.Parallel()

	service := NewInstallationService(
		nil,
		cipher.EncryptionKey{},
		InstallationConfig{
			ClientID:    "client-id",
			RedirectURI: "https://app.example.com/callback",
			AuthURL:     "https://slack.example.com/oauth/v2/authorize",
			StateSecret: "state-secret",
		},
		log.NewLogger(),
	)

	organizationID := gid.New(gid.NewTenantID(), coredata.OrganizationEntityType)
	identityID := gid.New(gid.NilTenant, coredata.IdentityEntityType)
	rawURL, err := service.InitiateURL(
		organizationID,
		identityID,
		"/organizations/example",
	)
	require.NoError(t, err)

	u, err := url.Parse(rawURL)
	require.NoError(t, err)
	assert.Equal(t, "client-id", u.Query().Get("client_id"))
	assert.Equal(t, "https://app.example.com/callback", u.Query().Get("redirect_uri"))
	assert.Contains(t, u.Query().Get("scope"), "chat:write")
	assert.Contains(t, u.Query().Get("scope"), "commands")
	assert.NotContains(t, u.Query().Get("scope"), "incoming-webhook")

	payload, err := validateInstallState(
		"state-secret",
		u.Query().Get("state"),
	)
	require.NoError(t, err)
	assert.Equal(t, organizationID, payload.Data.OrganizationID)
	assert.Equal(t, identityID, payload.Data.IdentityID)
}

func TestInstallationRefreshToken(t *testing.T) {
	t.Parallel()

	var received url.Values

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			require.NoError(t, r.ParseForm())
			received = r.Form

			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
				"ok":            true,
				"access_token":  "new-access-token",
				"refresh_token": "new-refresh-token",
				"expires_in":    43200,
				"scope":         "chat:write,channels:read",
			}))
		},
	))
	defer server.Close()

	service := NewInstallationService(
		nil,
		cipher.EncryptionKey{},
		InstallationConfig{
			ClientID:     "client-id",
			ClientSecret: "client-secret",
			TokenURL:     server.URL,
		},
		log.NewLogger(),
	)
	service.httpClient = server.Client()

	token, err := service.refreshToken(t.Context(), "old-refresh-token")
	require.NoError(t, err)
	assert.Equal(t, "new-access-token", token.AccessToken)
	assert.Equal(t, "new-refresh-token", token.RefreshToken)
	assert.Equal(t, "refresh_token", received.Get("grant_type"))
	assert.Equal(t, "old-refresh-token", received.Get("refresh_token"))
}

func TestInstallationComplete_ReleasesStateAfterExchangeFailure(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			require.NoError(
				t,
				json.NewEncoder(w).Encode(
					map[string]any{
						"ok":    false,
						"error": "temporarily_unavailable",
					},
				),
			)
		},
	))
	defer server.Close()

	ctx := t.Context()
	pgClient := test.PGClient(t)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organization := coredata.Organization{
		ID:        gid.New(tenantID, coredata.OrganizationEntityType),
		TenantID:  tenantID,
		Name:      "Slack OAuth release test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	require.NoError(
		t,
		pgClient.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				return organization.Insert(ctx, tx)
			},
		),
	)
	t.Cleanup(func() {
		_ = pgClient.WithTx(
			context.Background(),
			func(ctx context.Context, tx pg.Tx) error {
				return organization.Delete(ctx, tx, organization.ID)
			},
		)
	})

	service := NewInstallationService(
		pgClient,
		cipher.EncryptionKey{},
		InstallationConfig{
			ClientID:     "client-id",
			ClientSecret: "client-secret",
			RedirectURI:  "https://app.example.com/callback",
			TokenURL:     server.URL,
			StateSecret:  "state-secret",
		},
		log.NewLogger(),
	)
	service.httpClient = server.Client()
	state, err := newInstallState(
		"state-secret",
		organization.ID,
		gid.New(gid.NilTenant, coredata.IdentityEntityType),
		"",
	)
	require.NoError(t, err)

	_, err = service.Complete(ctx, state, "authorization-code")
	require.Error(t, err)

	claim := coredata.NewSlackbotInstallStateClaim(organization.ID, state)
	claimed := false

	require.NoError(
		t,
		pgClient.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				var err error

				claimed, err = claim.Claim(
					ctx,
					tx,
					scope,
					"00000000-0000-7000-8000-000000000010",
					time.Now(),
					5*time.Minute,
				)

				return err
			},
		),
	)
	assert.True(t, claimed)
}

func TestInstallationComplete_PersistsInstallationOnOAuthSuccess(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	pgClient := test.PGClient(t)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organization := coredata.Organization{
		ID:        gid.New(tenantID, coredata.OrganizationEntityType),
		TenantID:  tenantID,
		Name:      "Slack OAuth complete test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	teamID := fmt.Sprintf("T-complete-%s", organization.ID)

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			require.NoError(t, r.ParseForm())
			assert.Equal(t, "authorization-code", r.Form.Get("code"))
			assert.Equal(t, "client-id", r.Form.Get("client_id"))

			require.NoError(
				t,
				json.NewEncoder(w).Encode(
					map[string]any{
						"ok":            true,
						"access_token":  "xoxb-access",
						"refresh_token": "xoxe-refresh",
						"expires_in":    43200,
						"scope":         "chat:write,commands",
						"bot_user_id":   "U-bot",
						"team": map[string]any{
							"id": teamID,
						},
					},
				),
			)
		},
	))
	defer server.Close()

	require.NoError(
		t,
		pgClient.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				return organization.Insert(ctx, tx)
			},
		),
	)
	t.Cleanup(func() {
		_ = pgClient.WithTx(
			context.Background(),
			func(ctx context.Context, tx pg.Tx) error {
				return organization.Delete(ctx, tx, organization.ID)
			},
		)
	})

	encryptionKey := cipher.EncryptionKey{1, 2, 3}
	service := NewInstallationService(
		pgClient,
		encryptionKey,
		InstallationConfig{
			ClientID:     "client-id",
			ClientSecret: "client-secret",
			RedirectURI:  "https://app.example.com/callback",
			TokenURL:     server.URL,
			StateSecret:  "state-secret",
		},
		log.NewLogger(),
	)
	service.httpClient = server.Client()
	state, err := newInstallState(
		"state-secret",
		organization.ID,
		gid.New(gid.NilTenant, coredata.IdentityEntityType),
		"/organizations/example",
	)
	require.NoError(t, err)

	result, err := service.Complete(ctx, state, "authorization-code")
	require.NoError(t, err)
	require.NotNil(t, result.Installation)
	assert.Equal(t, teamID, result.Installation.TeamID)
	assert.Equal(t, "U-bot", result.Installation.BotUserID)
	assert.Equal(t, coredata.SlackbotInstallationStatusActive, result.Installation.Status)
	assert.Equal(t, "/organizations/example", result.ContinueURL)

	loaded, err := service.GetByOrganizationID(ctx, scope, organization.ID)
	require.NoError(t, err)
	assert.Equal(t, result.Installation.ID, loaded.ID)
	assert.Equal(t, teamID, loaded.TeamID)

	claim := coredata.NewSlackbotInstallStateClaim(organization.ID, state)
	claimed := false

	require.NoError(
		t,
		pgClient.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				var err error

				claimed, err = claim.Claim(
					ctx,
					tx,
					scope,
					"00000000-0000-7000-8000-000000000011",
					time.Now(),
					5*time.Minute,
				)

				return err
			},
		),
	)
	assert.False(t, claimed)
}

func TestInstallationUninstall_RemovesIdentityBindings(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	pgClient := test.PGClient(t)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organization := coredata.Organization{
		ID:        gid.New(tenantID, coredata.OrganizationEntityType),
		TenantID:  tenantID,
		Name:      "Slack uninstall binding test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	teamID := fmt.Sprintf("T-uninst-%s", organization.ID)
	otherTeamID := fmt.Sprintf("T-keep-%s", organization.ID)
	encryptionKey := cipher.EncryptionKey{1, 2, 3}

	require.NoError(
		t,
		pgClient.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				if err := organization.Insert(ctx, tx); err != nil {
					return err
				}

				installation := coredata.NewSlackbotInstallation(
					scope,
					organization.ID,
				)
				installation.TeamID = teamID
				installation.BotUserID = "B-bot"
				installation.Status = coredata.SlackbotInstallationStatusDisabled
				installation.Scopes = []string{"chat:write"}
				_, err := installation.Upsert(
					ctx,
					tx,
					scope,
					encryptionKey,
					coredata.SlackbotInstallationCredentials{AccessToken: "x"},
				)

				return err
			},
		),
	)
	t.Cleanup(func() {
		_ = pgClient.WithTx(
			context.Background(),
			func(ctx context.Context, tx pg.Tx) error {
				return organization.Delete(ctx, tx, organization.ID)
			},
		)
	})

	baseURL, err := baseurl.Parse("https://console.example.com")
	require.NoError(t, err)

	bindings := identitybinding.NewService(pgClient, baseURL)

	identityID := seedUninstallIdentity(t, ctx, pgClient)
	otherIdentityID := seedUninstallIdentity(t, ctx, pgClient)
	subject := identitybinding.Subject{
		Provider:         ProviderName,
		ExternalTenantID: teamID,
		ExternalUserID:   "U-uninst",
	}
	otherSubject := identitybinding.Subject{
		Provider:         ProviderName,
		ExternalTenantID: otherTeamID,
		ExternalUserID:   "U-keep",
	}
	token := issueUninstallBindToken(t, ctx, bindings, subject, organization.ID)
	_, err = bindings.Confirm(ctx, identityID, token)
	require.NoError(t, err)

	pendingToken := issueUninstallBindToken(
		t,
		ctx,
		bindings,
		identitybinding.Subject{
			Provider:         ProviderName,
			ExternalTenantID: teamID,
			ExternalUserID:   "U-pending",
		},
		organization.ID,
	)
	otherToken := issueUninstallBindToken(
		t,
		ctx,
		bindings,
		otherSubject,
		organization.ID,
	)
	otherBinding, err := bindings.Confirm(ctx, otherIdentityID, otherToken)
	require.NoError(t, err)

	bindPrompt := NewBindPromptService(pgClient, encryptionKey, log.NewLogger())
	require.NoError(
		t,
		bindPrompt.RememberResponseURL(
			ctx,
			teamID,
			"U-uninst",
			"https://hooks.slack.com/commands/T123/1/abc",
		),
	)
	require.NoError(
		t,
		bindPrompt.RememberResponseURL(
			ctx,
			otherTeamID,
			"U-keep",
			"https://hooks.slack.com/commands/T456/1/abc",
		),
	)

	service := NewInstallationService(
		pgClient,
		encryptionKey,
		InstallationConfig{},
		log.NewLogger(),
	)
	require.NoError(t, service.Uninstall(ctx, scope, organization.ID))

	_, err = service.GetByOrganizationID(ctx, scope, organization.ID)
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)

	_, err = bindings.Lookup(ctx, subject)
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)
	_, err = bindings.Preview(ctx, token)
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)
	_, err = bindings.Preview(ctx, pendingToken)
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)

	loaded, err := bindings.Lookup(ctx, otherSubject)
	require.NoError(t, err)
	assert.Equal(t, otherBinding.ID, loaded.ID)

	var removed coredata.SlackbotBindCallback

	err = pgClient.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return removed.LoadByTeamAndUser(ctx, conn, teamID, "U-uninst")
		},
	)
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)

	var kept coredata.SlackbotBindCallback

	require.NoError(
		t,
		pgClient.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				return kept.LoadByTeamAndUser(ctx, conn, otherTeamID, "U-keep")
			},
		),
	)
	assert.Equal(t, otherTeamID, kept.TeamID)
}

func seedUninstallIdentity(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
) gid.GID {
	t.Helper()

	identityID := gid.New(gid.NilTenant, coredata.IdentityEntityType)
	emailAddress, err := mail.ParseAddr(fmt.Sprintf("%s@example.com", identityID))
	require.NoError(t, err)

	now := time.Now().UTC()

	require.NoError(
		t,
		client.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				return (&coredata.Identity{
					ID:                   identityID,
					EmailAddress:         emailAddress,
					FullName:             "Slack Uninstall Binding Test",
					EmailAddressVerified: true,
					CreatedAt:            now,
					UpdatedAt:            now,
				}).Insert(ctx, tx)
			},
		),
	)
	t.Cleanup(func() {
		_ = client.WithTx(
			context.Background(),
			func(ctx context.Context, tx pg.Tx) error {
				return (&coredata.Identity{ID: identityID}).Delete(ctx, tx)
			},
		)
	})

	return identityID
}

func issueUninstallBindToken(
	t *testing.T,
	ctx context.Context,
	service *identitybinding.Service,
	subject identitybinding.Subject,
	organizationID gid.GID,
) string {
	t.Helper()

	bindURL, err := service.BindURL(ctx, subject, organizationID)
	require.NoError(t, err)
	parsed, err := url.Parse(bindURL)
	require.NoError(t, err)

	token := parsed.Query().Get("token")
	require.NotEmpty(t, token)

	return token
}

func TestSplitScopes(t *testing.T) {
	t.Parallel()

	assert.Equal(
		t,
		[]string{"chat:write", "channels:read", "groups:read"},
		splitScopes("chat:write,channels:read groups:read"),
	)
}
