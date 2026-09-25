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
	"errors"
	"strings"
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
	"go.probo.inc/probo/pkg/validator"
)

func TestEnableAccounts_RejectsUnsafeAccountFields(t *testing.T) {
	t.Parallel()

	longName := strings.Repeat("a", TitleMaxLength+1)

	tests := []struct {
		name    string
		account EnableConnectorAccount
		field   string
	}{
		{
			name: "empty external account id",
			account: EnableConnectorAccount{
				Name: "Acme",
			},
			field: "accounts[0].external_account_id",
		},
		{
			name: "name over max length",
			account: EnableConnectorAccount{
				ExternalAccountID: "123456789012",
				Name:              longName,
			},
			field: "accounts[0].name",
		},
		{
			name: "html in name",
			account: EnableConnectorAccount{
				ExternalAccountID: "123456789012",
				Name:              "<script>alert(1)</script>",
			},
			field: "accounts[0].name",
		},
		{
			name: "newline in external account id",
			account: EnableConnectorAccount{
				ExternalAccountID: "123\n456",
				Name:              "Acme",
			},
			field: "accounts[0].external_account_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := &ConnectorService{}
			connectorID := gid.New(gid.NewTenantID(), coredata.ConnectorEntityType)

			_, err := service.EnableAccounts(
				t.Context(),
				nil,
				connectorID,
				[]EnableConnectorAccount{tt.account},
			)
			require.Error(t, err)

			validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
			require.True(t, ok)
			assert.Contains(t, validationErrors.Fields(), tt.field)
		})
	}
}

func TestEnableAccounts_RejectsDuplicateExternalAccountIDs(t *testing.T) {
	t.Parallel()

	service := &ConnectorService{}
	connectorID := gid.New(gid.NewTenantID(), coredata.ConnectorEntityType)

	_, err := service.EnableAccounts(
		t.Context(),
		nil,
		connectorID,
		[]EnableConnectorAccount{
			{ExternalAccountID: "111", Name: "Prod"},
			{ExternalAccountID: "111", Name: "Production"},
		},
	)
	require.Error(t, err)

	validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
	require.True(t, ok)

	duplicate := validationErrors.ByField("accounts.external_account_id")
	require.Len(t, duplicate, 1)
	assert.Equal(t, validator.ErrorCodeInvalidFormat, duplicate[0].Code)
}

func seedConnectorDeleteOrg(t *testing.T, client *pg.Client) (coredata.Scoper, gid.GID) {
	t.Helper()

	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now().UTC()

	require.NoError(t, client.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		org := &coredata.Organization{
			ID:        organizationID,
			TenantID:  tenantID,
			Name:      "Connector delete " + organizationID.String(),
			CreatedAt: now,
			UpdatedAt: now,
		}

		return org.Insert(ctx, tx)
	}))

	return scope, organizationID
}

func newConnectorForDelete(
	t *testing.T,
	client *pg.Client,
	scope coredata.Scoper,
	organizationID gid.GID,
) *coredata.Connector {
	t.Helper()

	service := ConnectorService{svc: &Service{pg: client}}

	cnnctr, err := service.Create(t.Context(), scope, CreateConnectorRequest{
		OrganizationID: organizationID,
		Provider:       coredata.ConnectorProviderBrex,
		Protocol:       coredata.ConnectorProtocolOAuth2,
		Connection: &connector.OAuth2Connection{
			AccessToken: "test-token",
			TokenType:   "Bearer",
		},
	})
	require.NoError(t, err)

	return cnnctr
}

func insertConnectorAccountForDelete(
	t *testing.T,
	client *pg.Client,
	scope coredata.Scoper,
	organizationID gid.GID,
	connectorID gid.GID,
) *coredata.ConnectorAccount {
	t.Helper()

	now := time.Now().UTC()
	account := &coredata.ConnectorAccount{
		ID:                gid.New(scope.GetTenantID(), coredata.ConnectorAccountEntityType),
		OrganizationID:    organizationID,
		ConnectorID:       connectorID,
		ExternalAccountID: "123456789012",
		Name:              "Production",
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	require.NoError(t, client.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		_, err := account.Upsert(ctx, tx, scope)

		return err
	}))

	return account
}

func loadConnectorAccountsForDelete(
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

func TestConnectorService_Delete(t *testing.T) {
	t.Parallel()

	t.Run("a connector with an account and no source is deleted", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		scope, organizationID := seedConnectorDeleteOrg(t, client)
		service := ConnectorService{svc: &Service{pg: client}}
		cnnctr := newConnectorForDelete(t, client, scope, organizationID)
		insertConnectorAccountForDelete(t, client, scope, organizationID, cnnctr.ID)
		require.Len(t, loadConnectorAccountsForDelete(t, client, scope, cnnctr.ID), 1)

		require.NoError(t, service.Delete(t.Context(), scope, cnnctr.ID))

		_, err := service.Get(t.Context(), scope, cnnctr.ID)
		assert.ErrorIs(t, err, coredata.ErrResourceNotFound)
		assert.Empty(t, loadConnectorAccountsForDelete(t, client, scope, cnnctr.ID))
	})

	t.Run("an access review source refuses the delete by name", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		scope, organizationID := seedConnectorDeleteOrg(t, client)
		service := ConnectorService{svc: &Service{pg: client}}
		cnnctr := newConnectorForDelete(t, client, scope, organizationID)
		account := insertConnectorAccountForDelete(t, client, scope, organizationID, cnnctr.ID)

		now := time.Now().UTC()
		source := &coredata.AccessReviewSource{
			ID:                 gid.New(scope.GetTenantID(), coredata.AccessReviewSourceEntityType),
			OrganizationID:     organizationID,
			ConnectorAccountID: &account.ID,
			Name:               "Production",
			CreatedAt:          now,
			UpdatedAt:          now,
		}

		require.NoError(t, client.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
			_, err := source.Insert(ctx, tx, scope)

			return err
		}))

		err := service.Delete(t.Context(), scope, cnnctr.ID)
		require.Error(t, err)
		assert.ErrorIs(t, err, coredata.ErrResourceInUse)
		assert.ErrorContains(t, err, "access review source")

		_, err = service.Get(t.Context(), scope, cnnctr.ID)
		assert.NoError(t, err)
		assert.Len(t, loadConnectorAccountsForDelete(t, client, scope, cnnctr.ID), 1)
	})

	t.Run("a SCIM bridge refuses the delete by name", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		scope, organizationID := seedConnectorDeleteOrg(t, client)
		service := ConnectorService{svc: &Service{pg: client}}
		cnnctr := newConnectorForDelete(t, client, scope, organizationID)
		insertConnectorAccountForDelete(t, client, scope, organizationID, cnnctr.ID)

		now := time.Now().UTC()

		require.NoError(t, client.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
			config := &coredata.SCIMConfiguration{
				ID:             gid.New(scope.GetTenantID(), coredata.SCIMConfigurationEntityType),
				OrganizationID: organizationID,
				HashedToken:    []byte{0x01},
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := config.Insert(ctx, tx, scope); err != nil {
				return err
			}

			bridge := &coredata.SCIMBridge{
				ID:                  gid.New(scope.GetTenantID(), coredata.SCIMBridgeEntityType),
				OrganizationID:      organizationID,
				ScimConfigurationID: config.ID,
				ConnectorID:         &cnnctr.ID,
				Type:                coredata.SCIMBridgeTypeGoogleWorkspace,
				State:               coredata.SCIMBridgeStateActive,
				ExcludedUserNames:   []string{},
				CreatedAt:           now,
				UpdatedAt:           now,
			}

			return bridge.Insert(ctx, tx, scope)
		}))

		err := service.Delete(t.Context(), scope, cnnctr.ID)
		require.Error(t, err)
		assert.ErrorIs(t, err, coredata.ErrResourceInUse)
		assert.ErrorContains(t, err, "SCIM configuration")

		_, err = service.Get(t.Context(), scope, cnnctr.ID)
		assert.NoError(t, err)
		assert.Len(t, loadConnectorAccountsForDelete(t, client, scope, cnnctr.ID), 1)
	})
}

func TestConnectorService_Modules(t *testing.T) {
	t.Parallel()

	t.Run("an unused connector has no modules", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		scope, organizationID := seedConnectorDeleteOrg(t, client)
		service := ConnectorService{svc: &Service{pg: client}}
		cnnctr := newConnectorForDelete(t, client, scope, organizationID)

		modules, err := service.Modules(t.Context(), scope, cnnctr.ID)
		require.NoError(t, err)
		assert.Empty(t, modules)
	})

	t.Run("an access review source names the access review module", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		scope, organizationID := seedConnectorDeleteOrg(t, client)
		service := ConnectorService{svc: &Service{pg: client}}
		cnnctr := newConnectorForDelete(t, client, scope, organizationID)
		account := insertConnectorAccountForDelete(t, client, scope, organizationID, cnnctr.ID)

		now := time.Now().UTC()
		source := &coredata.AccessReviewSource{
			ID:                 gid.New(scope.GetTenantID(), coredata.AccessReviewSourceEntityType),
			OrganizationID:     organizationID,
			ConnectorAccountID: &account.ID,
			Name:               "Production",
			CreatedAt:          now,
			UpdatedAt:          now,
		}

		require.NoError(t, client.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
			_, err := source.Insert(ctx, tx, scope)

			return err
		}))

		modules, err := service.Modules(t.Context(), scope, cnnctr.ID)
		require.NoError(t, err)
		assert.Equal(t, []ConnectorModule{ConnectorModuleAccessReview}, modules)
	})

	t.Run("a SCIM bridge names the SCIM module", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		scope, organizationID := seedConnectorDeleteOrg(t, client)
		service := ConnectorService{svc: &Service{pg: client}}
		cnnctr := newConnectorForDelete(t, client, scope, organizationID)

		now := time.Now().UTC()

		require.NoError(t, client.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
			config := &coredata.SCIMConfiguration{
				ID:             gid.New(scope.GetTenantID(), coredata.SCIMConfigurationEntityType),
				OrganizationID: organizationID,
				HashedToken:    []byte{0x01},
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := config.Insert(ctx, tx, scope); err != nil {
				return err
			}

			bridge := &coredata.SCIMBridge{
				ID:                  gid.New(scope.GetTenantID(), coredata.SCIMBridgeEntityType),
				OrganizationID:      organizationID,
				ScimConfigurationID: config.ID,
				ConnectorID:         &cnnctr.ID,
				Type:                coredata.SCIMBridgeTypeGoogleWorkspace,
				State:               coredata.SCIMBridgeStateActive,
				ExcludedUserNames:   []string{},
				CreatedAt:           now,
				UpdatedAt:           now,
			}

			return bridge.Insert(ctx, tx, scope)
		}))

		modules, err := service.Modules(t.Context(), scope, cnnctr.ID)
		require.NoError(t, err)
		assert.Equal(t, []ConnectorModule{ConnectorModuleSCIM}, modules)
	})
}
