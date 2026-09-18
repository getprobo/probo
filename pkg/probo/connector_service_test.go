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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func seedConnectorAccountOrg(t *testing.T, client *pg.Client) (coredata.Scoper, gid.GID) {
	t.Helper()

	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now().UTC()

	require.NoError(t, client.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		org := &coredata.Organization{
			ID:        organizationID,
			TenantID:  tenantID,
			Name:      "Connector Account Test Org",
			CreatedAt: now,
			UpdatedAt: now,
		}

		return org.Insert(ctx, tx)
	}))

	return scope, organizationID
}

func loadConnectorAccounts(
	t *testing.T,
	client *pg.Client,
	scope coredata.Scoper,
	connectorID gid.GID,
) []*coredata.ConnectorAccount {
	t.Helper()

	var accounts []*coredata.ConnectorAccount

	require.NoError(t, client.WithConn(t.Context(), func(ctx context.Context, conn pg.Querier) error {
		var err error

		accounts, err = page.LoadAll(
			ctx,
			page.OrderBy[coredata.ConnectorAccountOrderField]{
				Field:     coredata.ConnectorAccountOrderFieldCreatedAt,
				Direction: page.OrderDirectionAsc,
			},
			func(
				ctx context.Context,
				cursor *page.Cursor[coredata.ConnectorAccountOrderField],
			) ([]*coredata.ConnectorAccount, error) {
				var batch coredata.ConnectorAccounts
				if err := batch.LoadByConnectorID(ctx, conn, scope, connectorID, cursor); err != nil {
					return nil, err
				}

				return batch, nil
			},
		)

		return err
	}))

	return accounts
}

// TestConnectorService_Create_ImpliedAccount covers the branch in the create
// path: a connector whose settings name an account, and an organization
// credential whose accounts come from discovery instead.
//
// The provider-by-provider shapes and the whole enable/disable lifecycle are
// covered through the API in e2e/console/connector_account_test.go.
func TestConnectorService_Create_ImpliedAccount(t *testing.T) {
	t.Parallel()

	t.Run("a cloud connector carries the account its settings name", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		scope, organizationID := seedConnectorAccountOrg(t, client)
		service := ConnectorService{svc: &Service{pg: client}}

		cnnctr, err := service.Create(t.Context(), scope, CreateConnectorRequest{
			OrganizationID: organizationID,
			Provider:       coredata.ConnectorProviderAWS,
			Protocol:       coredata.ConnectorProtocolWorkloadIdentity,
			Connection:     &connector.WorkloadIdentityConnection{},
			RawSettings:    []byte(`{"role_arn":"arn:aws:iam::123456789012:role/ProboAudit"}`),
		})
		require.NoError(t, err)

		accounts := loadConnectorAccounts(t, client, scope, cnnctr.ID)
		require.Len(t, accounts, 1)
		require.NotNil(t, accounts[0].ExternalAccountID)
		assert.Equal(t, "123456789012", *accounts[0].ExternalAccountID)
	})

	t.Run("an organization connector starts at zero accounts", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		scope, organizationID := seedConnectorAccountOrg(t, client)
		service := ConnectorService{svc: &Service{pg: client}}

		cnnctr, err := service.Create(t.Context(), scope, CreateConnectorRequest{
			OrganizationID:     organizationID,
			Provider:           coredata.ConnectorProviderAWS,
			Protocol:           coredata.ConnectorProtocolWorkloadIdentity,
			Connection:         &connector.WorkloadIdentityConnection{},
			RawSettings:        []byte(`{"role_arn":"arn:aws:iam::123456789012:role/ProboAudit","member_role_name":"ProboAudit"}`),
			SkipImpliedAccount: true,
		})
		require.NoError(t, err)

		accounts := loadConnectorAccounts(t, client, scope, cnnctr.ID)
		assert.Empty(t, accounts, "discovery decides which accounts an organization credential covers")
	})
}
