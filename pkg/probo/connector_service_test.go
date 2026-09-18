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

// TestConnectorAccountService pins the two rules that keep Settings and the
// modules from owning each other's state: enabling is idempotent per vendor
// identifier, and an account a source reviews cannot be removed out from
// under it.
func TestConnectorAccountService(t *testing.T) {
	t.Parallel()

	newAWSConnector := func(t *testing.T, client *pg.Client, scope coredata.Scoper, organizationID gid.GID) *coredata.Connector {
		t.Helper()

		service := ConnectorService{svc: &Service{pg: client}}

		cnnctr, err := service.Create(t.Context(), scope, CreateConnectorRequest{
			OrganizationID:     organizationID,
			Provider:           coredata.ConnectorProviderAWS,
			Protocol:           coredata.ConnectorProtocolWorkloadIdentity,
			Connection:         &connector.WorkloadIdentityConnection{},
			RawSettings:        []byte(`{"role_arn":"arn:aws:iam::111111111111:role/ProboAudit","member_role_name":"ProboAudit"}`),
			SkipImpliedAccount: true,
		})
		require.NoError(t, err)

		return cnnctr
	}

	t.Run("enabling the same account twice yields one row", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		scope, organizationID := seedConnectorAccountOrg(t, client)
		cnnctr := newAWSConnector(t, client, scope, organizationID)
		service := ConnectorAccountService{svc: &Service{pg: client}}

		req := EnableConnectorAccountsRequest{
			OrganizationID: organizationID,
			ConnectorID:    cnnctr.ID,
			Accounts:       []ConnectorAccountInput{{ExternalAccountID: "222222222222", Name: "Production"}},
		}

		first, err := service.Enable(t.Context(), scope, req)
		require.NoError(t, err)
		require.Len(t, first, 1)

		req.Accounts[0].Name = "Production renamed"

		second, err := service.Enable(t.Context(), scope, req)
		require.NoError(t, err)
		require.Len(t, second, 1)

		assert.Equal(t, first[0].ID, second[0].ID, "the row keeps its id, so a source pointing at it still does")
		assert.Equal(t, "Production renamed", second[0].Name)
		assert.Len(t, loadConnectorAccounts(t, client, scope, cnnctr.ID), 1)
	})

	t.Run("enabling records several accounts on one credential", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		scope, organizationID := seedConnectorAccountOrg(t, client)
		cnnctr := newAWSConnector(t, client, scope, organizationID)
		service := ConnectorAccountService{svc: &Service{pg: client}}

		_, err := service.Enable(t.Context(), scope, EnableConnectorAccountsRequest{
			OrganizationID: organizationID,
			ConnectorID:    cnnctr.ID,
			Accounts: []ConnectorAccountInput{
				{ExternalAccountID: "222222222222", Name: "Production"},
				{ExternalAccountID: "333333333333", Name: "Staging"},
			},
		})
		require.NoError(t, err)

		assert.Len(t, loadConnectorAccounts(t, client, scope, cnnctr.ID), 2)
	})

	t.Run("a connector in another organization is refused", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		scope, organizationID := seedConnectorAccountOrg(t, client)
		cnnctr := newAWSConnector(t, client, scope, organizationID)

		_, otherOrganizationID := seedConnectorAccountOrg(t, client)
		service := ConnectorAccountService{svc: &Service{pg: client}}

		_, err := service.Enable(t.Context(), scope, EnableConnectorAccountsRequest{
			OrganizationID: otherOrganizationID,
			ConnectorID:    cnnctr.ID,
			Accounts:       []ConnectorAccountInput{{ExternalAccountID: "222222222222"}},
		})
		require.Error(t, err)
	})

	t.Run("disabling an unused account removes it", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		scope, organizationID := seedConnectorAccountOrg(t, client)
		cnnctr := newAWSConnector(t, client, scope, organizationID)
		service := ConnectorAccountService{svc: &Service{pg: client}}

		enabled, err := service.Enable(t.Context(), scope, EnableConnectorAccountsRequest{
			OrganizationID: organizationID,
			ConnectorID:    cnnctr.ID,
			Accounts:       []ConnectorAccountInput{{ExternalAccountID: "222222222222"}},
		})
		require.NoError(t, err)

		require.NoError(t, service.Disable(t.Context(), scope, DisableConnectorAccountRequest{
			ConnectorAccountID: enabled[0].ID,
		}))

		assert.Empty(t, loadConnectorAccounts(t, client, scope, cnnctr.ID))
	})

	t.Run("an account a source reviews cannot be disabled", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		scope, organizationID := seedConnectorAccountOrg(t, client)
		cnnctr := newAWSConnector(t, client, scope, organizationID)
		service := ConnectorAccountService{svc: &Service{pg: client}}

		enabled, err := service.Enable(t.Context(), scope, EnableConnectorAccountsRequest{
			OrganizationID: organizationID,
			ConnectorID:    cnnctr.ID,
			Accounts:       []ConnectorAccountInput{{ExternalAccountID: "222222222222"}},
		})
		require.NoError(t, err)

		now := time.Now().UTC()
		source := &coredata.AccessReviewSource{
			ID:                 gid.New(scope.GetTenantID(), coredata.AccessReviewSourceEntityType),
			OrganizationID:     organizationID,
			ConnectorID:        &cnnctr.ID,
			ConnectorAccountID: &enabled[0].ID,
			Name:               "Production",
			CreatedAt:          now,
			UpdatedAt:          now,
		}

		require.NoError(t, client.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
			_, err := source.Insert(ctx, tx, scope)

			return err
		}))

		err = service.Disable(t.Context(), scope, DisableConnectorAccountRequest{
			ConnectorAccountID: enabled[0].ID,
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, coredata.ErrResourceInUse, "a review must not lose the account it reviews")
	})
}
