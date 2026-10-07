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

package probo

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
)

func TestCreate_RecordsImplicitAccount(t *testing.T) {
	t.Parallel()

	svc, scope, organizationID := newConnectorCreateEnv(t)

	created, err := svc.Create(
		t.Context(),
		scope,
		CreateConnectorRequest{
			OrganizationID: organizationID,
			Provider:       coredata.ConnectorProviderBrex,
			Protocol:       coredata.ConnectorProtocolAPIKey,
			Connection:     &connector.APIKeyConnection{APIKey: "bxt_test-key"},
		},
	)
	require.NoError(t, err)

	account := loadConnectorAccount(t, svc, scope, created.ID, created.ID.String())
	assert.Equal(t, "Brex", account.Name)
	assert.Equal(t, 1, countConnectorAccounts(t, svc, scope, created.ID))
}

func TestCreate_RecordsSettingsAccount(t *testing.T) {
	t.Parallel()

	svc, scope, organizationID := newConnectorCreateEnv(t)

	created, err := svc.Create(
		t.Context(),
		scope,
		CreateConnectorRequest{
			OrganizationID: organizationID,
			Provider:       coredata.ConnectorProviderAWS,
			Protocol:       coredata.ConnectorProtocolWorkloadIdentity,
			Connection:     &connector.WorkloadIdentityConnection{},
			RawSettings:    json.RawMessage(`{"role_arn":"arn:aws:iam::123456789012:role/ProboAudit"}`),
		},
	)
	require.NoError(t, err)

	account := loadConnectorAccount(t, svc, scope, created.ID, "123456789012")
	assert.Equal(t, "123456789012", account.Name)
	assert.Equal(t, 1, countConnectorAccounts(t, svc, scope, created.ID))
}

func TestCreate_LeavesOrganizationInstallWithoutAccount(t *testing.T) {
	t.Parallel()

	svc, scope, organizationID := newConnectorCreateEnv(t)

	created, err := svc.Create(
		t.Context(),
		scope,
		CreateConnectorRequest{
			OrganizationID: organizationID,
			Provider:       coredata.ConnectorProviderAWS,
			Protocol:       coredata.ConnectorProtocolWorkloadIdentity,
			Connection:     &connector.WorkloadIdentityConnection{},
		},
	)
	require.NoError(t, err)
	assert.Equal(t, 0, countConnectorAccounts(t, svc, scope, created.ID))
}

func newConnectorCreateEnv(t *testing.T) (*ConnectorService, coredata.Scoper, gid.GID) {
	t.Helper()

	client := test.PGClient(t)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now().UTC()

	require.NoError(t, client.WithTx(
		t.Context(),
		func(ctx context.Context, tx pg.Tx) error {
			org := &coredata.Organization{
				ID:        organizationID,
				TenantID:  tenantID,
				Name:      "Connector account seed",
				CreatedAt: now,
				UpdatedAt: now,
			}

			return org.Insert(ctx, tx)
		},
	))

	var key cipher.EncryptionKey

	return &ConnectorService{
		svc: &Service{
			pg:            client,
			encryptionKey: key,
		},
		providerRegistry: provider.NewBuiltinRegistry(),
	}, scope, organizationID
}

func loadConnectorAccount(
	t *testing.T,
	svc *ConnectorService,
	scope coredata.Scoper,
	connectorID gid.GID,
	externalID string,
) *coredata.ConnectorAccount {
	t.Helper()

	account := &coredata.ConnectorAccount{}

	require.NoError(t, svc.svc.pg.WithConn(
		t.Context(),
		func(ctx context.Context, conn pg.Querier) error {
			return account.LoadByConnectorAndExternalID(ctx, conn, scope, connectorID, externalID)
		},
	))

	return account
}

func countConnectorAccounts(
	t *testing.T,
	svc *ConnectorService,
	scope coredata.Scoper,
	connectorID gid.GID,
) int {
	t.Helper()

	var count int

	require.NoError(t, svc.svc.pg.WithConn(
		t.Context(),
		func(ctx context.Context, conn pg.Querier) error {
			var err error

			accounts := coredata.ConnectorAccounts{}
			count, err = accounts.CountByConnectorID(ctx, conn, scope, connectorID)

			return err
		},
	))

	return count
}
