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

package connector

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/connectorprotocol"
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

			service := &Service{}
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

	service := &Service{}
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
	scope := coredata.NewScope(tenantID, gid.Nil)
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

	service := Service{pg: client}

	cnnctr, err := service.Create(t.Context(), scope, CreateConnectorRequest{
		OrganizationID: organizationID,
		Name:           "Test",
		Provider:       coredata.ConnectorProviderBrex,
		Protocol:       coredata.ConnectorProtocolOAuth2,
		Connection: &connectorprotocol.OAuth2Connection{
			AccessToken: "test-token",
			TokenType:   "Bearer",
		},
	})
	require.NoError(t, err)

	return cnnctr
}

func loadStandaloneAccountForDelete(
	t *testing.T,
	client *pg.Client,
	scope coredata.Scoper,
	connectorID gid.GID,
) *coredata.ConnectorAccount {
	t.Helper()

	account := &coredata.ConnectorAccount{}

	require.NoError(t, client.WithConn(t.Context(), func(ctx context.Context, conn pg.Querier) error {
		return account.LoadStandaloneByConnectorID(ctx, conn, scope, connectorID)
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
		service := Service{pg: client}
		cnnctr := newConnectorForDelete(t, client, scope, organizationID)
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
		service := Service{pg: client}
		cnnctr := newConnectorForDelete(t, client, scope, organizationID)
		account := loadStandaloneAccountForDelete(t, client, scope, cnnctr.ID)

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
		service := Service{pg: client}
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

		err := service.Delete(t.Context(), scope, cnnctr.ID)
		require.Error(t, err)
		assert.ErrorIs(t, err, coredata.ErrResourceInUse)
		assert.ErrorContains(t, err, "SCIM configuration")

		_, err = service.Get(t.Context(), scope, cnnctr.ID)
		assert.NoError(t, err)
		assert.Len(t, loadConnectorAccountsForDelete(t, client, scope, cnnctr.ID), 1)
	})
}

func TestConnectorName_RejectsEmptyAndUnsafeValues(t *testing.T) {
	t.Parallel()

	longName := strings.Repeat("a", NameMaxLength+1)

	tests := []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "blank", value: "   "},
		{name: "too long", value: longName},
		{name: "html", value: "<b>Production</b>"},
		{name: "newline", value: "Production\norg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NormalizeConnectorName(tt.value)
			require.Error(t, err)

			validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
			require.True(t, ok)
			assert.Contains(t, validationErrors.Fields(), "name")
		})
	}
}

func TestCreateAndUpdate_StoreTrimmedConnectorName(t *testing.T) {
	t.Parallel()

	svc, scope, organizationID := newConnectorCreateEnv(t)

	created, err := svc.Create(t.Context(), scope, CreateConnectorRequest{
		OrganizationID: organizationID,
		Name:           "  Production  ",
		Provider:       coredata.ConnectorProviderBrex,
		Protocol:       coredata.ConnectorProtocolAPIKey,
		Connection:     &connectorprotocol.APIKeyConnection{APIKey: "bxt_test-key"},
	})
	require.NoError(t, err)
	assert.Equal(t, "Production", created.Name)

	updated, err := svc.Update(t.Context(), scope, UpdateConnectorRequest{
		ConnectorID: created.ID,
		Name:        "  QA  ",
	})
	require.NoError(t, err)
	assert.Equal(t, "QA", updated.Name)

	_, err = svc.Update(t.Context(), scope, UpdateConnectorRequest{
		ConnectorID: created.ID,
		Name:        "   ",
	})
	require.Error(t, err)

	validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
	require.True(t, ok)
	assert.Contains(t, validationErrors.Fields(), "name")
}

func TestUpdateAPIKey_ReplacesCustomerKey(t *testing.T) {
	t.Parallel()

	svc, scope, organizationID := newConnectorCreateEnv(t)

	created, err := svc.Create(t.Context(), scope, CreateConnectorRequest{
		OrganizationID: organizationID,
		Name:           "Production",
		Provider:       coredata.ConnectorProviderSentry,
		Protocol:       coredata.ConnectorProtocolAPIKey,
		Connection: &connectorprotocol.APIKeyConnection{
			APIKey: "old-key",
			Header: "x-stale",
		},
		RawSettings: json.RawMessage(`{"organization_slug":"acme"}`),
	})
	require.NoError(t, err)

	updated, err := svc.UpdateAPIKey(t.Context(), scope, UpdateAPIKeyConnectorRequest{
		ConnectorID: created.ID,
		Connection: &connectorprotocol.APIKeyConnection{
			APIKey: "  next-key  ",
			Header: "x-api-key",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "Production", updated.Name)

	loaded, err := svc.GetWithConnection(t.Context(), scope, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Production", loaded.Name)
	assert.JSONEq(t, `{"organization_slug":"acme"}`, string(loaded.RawSettings))

	conn, ok := loaded.Connection.(*connectorprotocol.APIKeyConnection)
	require.True(t, ok)
	assert.Equal(t, "next-key", conn.APIKey)
	assert.Equal(t, "x-api-key", conn.Header)

	replaced, err := svc.UpdateAPIKey(t.Context(), scope, UpdateAPIKeyConnectorRequest{
		ConnectorID: created.ID,
		Connection:  &connectorprotocol.APIKeyConnection{APIKey: "rotated-key"},
		RawSettings: json.RawMessage(`{"organization_slug":"other"}`),
	})
	require.NoError(t, err)
	assert.Equal(t, "Production", replaced.Name)

	loaded, err = svc.GetWithConnection(t.Context(), scope, created.ID)
	require.NoError(t, err)
	assert.JSONEq(t, `{"organization_slug":"other"}`, string(loaded.RawSettings))
	conn, ok = loaded.Connection.(*connectorprotocol.APIKeyConnection)
	require.True(t, ok)
	assert.Equal(t, "rotated-key", conn.APIKey)
	assert.Empty(t, conn.Header)
}

func TestUpdateAPIKey_RefusesOtherProtocolsAndManagedKeys(t *testing.T) {
	t.Parallel()

	svc, scope, organizationID := newConnectorCreateEnv(t)

	oauth, err := svc.Create(t.Context(), scope, CreateConnectorRequest{
		OrganizationID: organizationID,
		Name:           "OAuth",
		Provider:       coredata.ConnectorProviderBrex,
		Protocol:       coredata.ConnectorProtocolOAuth2,
		Connection: &connectorprotocol.OAuth2Connection{
			AccessToken: "keep-me",
			TokenType:   "Bearer",
		},
	})
	require.NoError(t, err)

	_, err = svc.UpdateAPIKey(t.Context(), scope, UpdateAPIKeyConnectorRequest{
		ConnectorID: oauth.ID,
		Connection:  &connectorprotocol.APIKeyConnection{APIKey: "should-not-stick"},
	})
	require.ErrorIs(t, err, ErrAPIKeyNotReplaceable)

	loaded, err := svc.GetWithConnection(t.Context(), scope, oauth.ID)
	require.NoError(t, err)

	oauthConn, ok := loaded.Connection.(*connectorprotocol.OAuth2Connection)
	require.True(t, ok)
	assert.Equal(t, "keep-me", oauthConn.AccessToken)

	managed, err := svc.Create(t.Context(), scope, CreateConnectorRequest{
		OrganizationID: organizationID,
		Name:           "Crisp",
		Provider:       coredata.ConnectorProviderCrisp,
		Protocol:       coredata.ConnectorProtocolAPIKey,
		Connection:     &connectorprotocol.APIKeyConnection{APIKey: "plugin-token"},
		RawSettings:    json.RawMessage(`{"website_id":"site-1"}`),
	})
	require.NoError(t, err)

	_, err = svc.UpdateAPIKey(t.Context(), scope, UpdateAPIKeyConnectorRequest{
		ConnectorID: managed.ID,
		Connection:  &connectorprotocol.APIKeyConnection{APIKey: "customer-paste"},
	})
	require.ErrorIs(t, err, ErrAPIKeyNotReplaceable)

	loaded, err = svc.GetWithConnection(t.Context(), scope, managed.ID)
	require.NoError(t, err)

	managedConn, ok := loaded.Connection.(*connectorprotocol.APIKeyConnection)
	require.True(t, ok)
	assert.Equal(t, "plugin-token", managedConn.APIKey)
}
