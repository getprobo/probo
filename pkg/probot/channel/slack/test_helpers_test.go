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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/probot"
)

func testEncryptionKey() cipher.EncryptionKey {
	return cipher.EncryptionKey{1, 2, 3}
}

func newTestInstallationService(
	t *testing.T,
	pgClient *pg.Client,
	apiBaseURL string,
) *InstallationService {
	t.Helper()

	if apiBaseURL == "" {
		server := httptest.NewServer(
			http.HandlerFunc(
				func(w http.ResponseWriter, _ *http.Request) {
					_, _ = w.Write([]byte(`{"ok":true,"channel":"C","ts":"1.1","messages":[]}`))
				},
			),
		)
		t.Cleanup(server.Close)
		apiBaseURL = server.URL + "/api"
	}

	service := NewInstallationService(
		pgClient,
		testEncryptionKey(),
		InstallationConfig{APIBaseURL: apiBaseURL},
		log.NewLogger(),
	)
	service.httpClient = httpclient.DefaultPooledClient(
		httpclient.WithSSRFProtection(),
		httpclient.WithSSRFAllowLoopback(),
	)

	return service
}

func insertActiveInstallation(
	t *testing.T,
	pgClient *pg.Client,
	organizationID gid.GID,
	teamID string,
	botUserID string,
) {
	t.Helper()

	insertInstallationWithKey(
		t,
		pgClient,
		organizationID,
		teamID,
		botUserID,
		testEncryptionKey(),
	)
}

func insertInstallationWithKey(
	t *testing.T,
	pgClient *pg.Client,
	organizationID gid.GID,
	teamID string,
	botUserID string,
	key cipher.EncryptionKey,
) {
	t.Helper()

	scope := coredata.NewScopeFromObjectID(organizationID)
	installation := coredata.NewSlackbotInstallation(scope, organizationID)
	installation.TeamID = teamID
	installation.BotUserID = botUserID
	installation.Scopes = []string{"chat:write"}

	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				_, err := installation.Upsert(
					ctx,
					tx,
					scope,
					key,
					coredata.SlackbotInstallationCredentials{AccessToken: testBotToken},
				)

				return err
			},
		),
	)
}

func newTestAgentProfiles(t *testing.T) *probot.AgentProfileRegistry {
	t.Helper()

	profiles := probot.NewAgentProfileRegistry()
	require.NoError(t, profiles.Register("probot", agent.New("Probot", nil)))

	return profiles
}

func uniqueSlackTeamID(t *testing.T) string {
	t.Helper()

	return "T-" + gid.New(gid.NilTenant, coredata.SlackbotInstallationEntityType).String()
}

func insertTestIdentity(t *testing.T, pgClient *pg.Client, identityID gid.GID) {
	t.Helper()

	now := time.Now()
	emailAddress, err := mail.ParseAddr(identityID.String() + "@example.com")
	require.NoError(t, err)
	identity := coredata.Identity{
		ID:                   identityID,
		EmailAddress:         emailAddress,
		FullName:             "Slack Test",
		EmailAddressVerified: true,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				return identity.Insert(ctx, tx)
			},
		),
	)
	t.Cleanup(
		func() {
			_ = pgClient.WithTx(
				context.Background(),
				func(ctx context.Context, tx pg.Tx) error {
					return identity.Delete(ctx, tx)
				},
			)
		},
	)
}
