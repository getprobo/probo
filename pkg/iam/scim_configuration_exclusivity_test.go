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

package iam

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

func TestCreateSCIMConfiguration_RefusesAuditConnector(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	svc := OrganizationService{Service: &Service{pg: client}}
	scope, organizationID, connectorID := seedAuditConnector(t, ctx, client)

	_, _, _, err := svc.CreateSCIMConfiguration(ctx, organizationID, &connectorID)
	require.ErrorIs(t, err, coredata.ErrResourceInUse)
	require.ErrorContains(t, err, "access review source")

	loaded := &coredata.SCIMConfiguration{}
	err = client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return loaded.LoadByOrganizationID(ctx, conn, scope, organizationID)
	})
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)
}

func TestDeleteSCIMConfiguration_KeepsConnector(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	svc := OrganizationService{Service: &Service{pg: client}}
	scope, organizationID, connectorID := seedPlainConnector(t, ctx, client)

	config, _, _, err := svc.CreateSCIMConfiguration(ctx, organizationID, &connectorID)
	require.NoError(t, err)

	require.NoError(t, svc.DeleteSCIMConfiguration(ctx, organizationID, config.ID))

	loaded := &coredata.Connector{}
	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return loaded.LoadMetadataByID(ctx, conn, scope, connectorID)
	}))
}

func seedAuditConnector(t *testing.T, ctx context.Context, client *pg.Client) (coredata.Scoper, gid.GID, gid.GID) {
	t.Helper()

	scope, organizationID, connectorID := seedPlainConnector(t, ctx, client)
	now := time.Now()
	accountID := gid.New(scope.GetTenantID(), coredata.ConnectorAccountEntityType)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		account := &coredata.ConnectorAccount{
			ID:                accountID,
			OrganizationID:    organizationID,
			ConnectorID:       connectorID,
			ExternalAccountID: "workspace",
			Name:              "workspace",
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		if _, err := account.Upsert(ctx, tx, scope); err != nil {
			return err
		}

		source := &coredata.AccessReviewSource{
			ID:                 gid.New(scope.GetTenantID(), coredata.AccessReviewSourceEntityType),
			OrganizationID:     organizationID,
			ConnectorAccountID: &accountID,
			Name:               "audit",
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		inserted, err := source.Insert(ctx, tx, scope)
		if err != nil {
			return err
		}
		if !inserted {
			return coredata.ErrResourceAlreadyExists
		}

		return nil
	}))

	return scope, organizationID, connectorID
}

func seedPlainConnector(t *testing.T, ctx context.Context, client *pg.Client) (coredata.Scoper, gid.GID, gid.GID) {
	t.Helper()

	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	connectorID := gid.New(tenantID, coredata.ConnectorEntityType)
	now := time.Now()
	var key cipher.EncryptionKey

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		org := &coredata.Organization{
			ID:        organizationID,
			TenantID:  tenantID,
			Name:      "SCIM exclusivity",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := org.Insert(ctx, tx); err != nil {
			return err
		}

		cnnctr := &coredata.Connector{
			ID:             connectorID,
			OrganizationID: organizationID,
			Provider:       coredata.ConnectorProviderGoogleWorkspace,
			Protocol:       coredata.ConnectorProtocolOAuth2,
			Connection: &connector.OAuth2Connection{
				AccessToken: "test-token",
				TokenType:   "Bearer",
			},
			CreatedAt: now,
			UpdatedAt: now,
		}

		return cnnctr.Insert(ctx, tx, scope, key)
	}))

	return scope, organizationID, connectorID
}
