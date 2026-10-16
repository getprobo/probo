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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/connectorprotocol"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
)

func TestConnector_UpdateConnection(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	scope, organizationID := seedConnectorOrg(t, ctx, client)

	var key cipher.EncryptionKey

	now := time.Now().UTC()
	cnnctr := &coredata.Connector{
		ID:             gid.New(scope.GetTenantID(), coredata.ConnectorEntityType),
		OrganizationID: organizationID,
		Provider:       coredata.ConnectorProviderPlanetScale,
		Protocol:       coredata.ConnectorProtocolOAuth2,
		Name:           "Acme PlanetScale",
		Connection: &connectorprotocol.OAuth2Connection{
			AccessToken:  "old-token",
			RefreshToken: "old-refresh",
			TokenType:    "Bearer",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, cnnctr.SetSettings(coredata.PlanetScaleConnectorSettings{Organization: "acme"}))

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return cnnctr.Insert(ctx, tx, scope, key)
	}))

	cnnctr.Name = "stale name"
	require.NoError(t, cnnctr.SetSettings(coredata.PlanetScaleConnectorSettings{Organization: "stale"}))
	cnnctr.Connection = &connectorprotocol.OAuth2Connection{
		AccessToken:  "new-token",
		RefreshToken: "new-refresh",
		TokenType:    "Bearer",
	}
	cnnctr.UpdatedAt = now.Add(time.Minute)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return cnnctr.UpdateConnection(ctx, tx, scope, key)
	}))

	loaded := &coredata.Connector{}

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return loaded.LoadByID(ctx, conn, scope, cnnctr.ID, key)
	}))

	conn, ok := loaded.Connection.(*connectorprotocol.OAuth2Connection)
	require.True(t, ok)
	assert.Equal(t, "new-token", conn.AccessToken)
	assert.Equal(t, "new-refresh", conn.RefreshToken)
	assert.Equal(t, "Acme PlanetScale", loaded.Name)

	settings, err := coredata.ConnectorSettings[coredata.PlanetScaleConnectorSettings](loaded)
	require.NoError(t, err)
	assert.Equal(t, "acme", settings.Organization)

	missing := &coredata.Connector{
		ID:         gid.New(scope.GetTenantID(), coredata.ConnectorEntityType),
		Connection: cnnctr.Connection,
	}
	err = client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return missing.UpdateConnection(ctx, tx, scope, key)
	})
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)
}
