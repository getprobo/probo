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

package coredata_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
)

// seedConnectorOrg creates an organization for connector tests and
// returns its scope and ID.
func seedConnectorOrg(t *testing.T, ctx context.Context, client *pg.Client) (coredata.Scoper, gid.GID) {
	t.Helper()

	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now().UTC()

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		org := &coredata.Organization{
			ID:        organizationID,
			TenantID:  tenantID,
			Name:      "Connector Multi Test Org",
			CreatedAt: now,
			UpdatedAt: now,
		}

		return org.Insert(ctx, tx)
	}))

	return scope, organizationID
}

func insertConnector(
	ctx context.Context,
	client *pg.Client,
	scope coredata.Scoper,
	organizationID gid.GID,
	provider coredata.ConnectorProvider,
	key cipher.EncryptionKey,
) (gid.GID, error) {
	now := time.Now().UTC()
	id := gid.New(scope.GetTenantID(), coredata.ConnectorEntityType)

	err := client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		cnnctr := &coredata.Connector{
			ID:             id,
			OrganizationID: organizationID,
			Provider:       provider,
			Protocol:       coredata.ConnectorProtocolOAuth2,
			Connection: &connector.OAuth2Connection{
				AccessToken: "test-token",
				TokenType:   "Bearer",
			},
			CreatedAt: now,
			UpdatedAt: now,
		}

		return cnnctr.Insert(ctx, tx, scope, key)
	})

	return id, err
}

// TestConnectorInsert_MultipleConnectionsPerProvider pins the relaxed
// uniqueness: an organization may hold several connectors of the same
// provider — including Slack, whose messaging pipeline now picks its
// credential deterministically instead of relying on a unique row
// (pkg/coredata/migrations/20260819T142937Z.sql).
func TestConnectorInsert_MultipleConnectionsPerProvider(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	scope, organizationID := seedConnectorOrg(t, ctx, client)

	var key cipher.EncryptionKey

	for _, provider := range []coredata.ConnectorProvider{
		coredata.ConnectorProviderGitHub,
		coredata.ConnectorProviderGitHub,
		coredata.ConnectorProviderSlack,
		coredata.ConnectorProviderSlack,
	} {
		_, err := insertConnector(ctx, client, scope, organizationID, provider, key)
		require.NoError(t, err)
	}
}

// TestLoadSlackMessagingConnector pins the deterministic pick the
// messaging worker relies on with several Slack workspaces connected:
// a channel-configured connector (only the legacy messaging connect
// flow captured one) wins over an older plain one; with none
// configured, the oldest row wins.
func TestLoadSlackMessagingConnector(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	scope, organizationID := seedConnectorOrg(t, ctx, client)

	var key cipher.EncryptionKey

	insertSlack := func(createdAt time.Time, channelID string) gid.GID {
		id := gid.New(scope.GetTenantID(), coredata.ConnectorEntityType)
		conn := &connector.SlackConnection{
			OAuth2Connection: connector.OAuth2Connection{
				AccessToken: "test-token",
				TokenType:   "Bearer",
			},
		}
		// Insert copies the channel from the connection's settings into
		// the connector settings column the selector orders by.
		if channelID != "" {
			conn.Settings.Channel = "#compliance"
			conn.Settings.ChannelID = channelID
		}

		cnnctr := &coredata.Connector{
			ID:             id,
			OrganizationID: organizationID,
			Provider:       coredata.ConnectorProviderSlack,
			Protocol:       coredata.ConnectorProtocolOAuth2,
			Connection:     conn,
			CreatedAt:      createdAt,
			UpdatedAt:      createdAt,
		}

		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			return cnnctr.Insert(ctx, tx, scope, key)
		}))

		return id
	}

	now := time.Now().UTC()
	oldest := insertSlack(now.Add(-2*time.Hour), "")
	configured := insertSlack(now.Add(-1*time.Hour), "C123")

	picked := &coredata.Connector{}

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return picked.LoadSlackMessagingConnector(ctx, conn, scope, key, organizationID)
	}))
	require.Equal(t, configured, picked.ID, "channel-configured connector must win over an older plain one")

	// Remove the configured row: the pick falls back to the oldest.
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		toDelete := &coredata.Connector{ID: configured}
		return toDelete.Delete(ctx, tx, scope)
	}))

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return picked.LoadSlackMessagingConnector(ctx, conn, scope, key, organizationID)
	}))
	require.Equal(t, oldest, picked.ID, "oldest connector must win when none is channel-configured")
}
