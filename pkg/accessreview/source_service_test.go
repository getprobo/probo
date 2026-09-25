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

package accessreview_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
)

func TestUpdateSource_RelinkWhenOtherAccountHasSource(t *testing.T) {
	t.Parallel()

	env := newAccessSourceEnv(t)
	connectorID := env.insertConnector(
		t,
		coredata.ConnectorProviderGitHub,
		&coredata.GitHubConnectorSettings{Organization: "acme"},
	)
	enabledID := env.insertAccount(t, connectorID, "enabled-account")

	env.insertAccount(t, connectorID, "acme")

	enabledSource := env.insertLinkedSource(t, enabledID, "enabled source")
	csvSource := insertAccessSource(t, env.ctx, env.client, env.scope, &coredata.AccessReviewSource{
		ID:             gid.New(env.scope.GetTenantID(), coredata.AccessReviewSourceEntityType),
		OrganizationID: env.organizationID,
		Name:           "csv source",
		CreatedAt:      env.now,
		UpdatedAt:      env.now,
	})
	secondCSV := insertAccessSource(t, env.ctx, env.client, env.scope, &coredata.AccessReviewSource{
		ID:             gid.New(env.scope.GetTenantID(), coredata.AccessReviewSourceEntityType),
		OrganizationID: env.organizationID,
		Name:           "second csv source",
		CreatedAt:      env.now,
		UpdatedAt:      env.now,
	})

	updated, err := env.svc.UpdateSource(
		env.ctx,
		env.scope,
		accessreview.UpdateAccessReviewSourceRequest{
			AccessReviewSourceID: csvSource.ID,
			ConnectorID:          new(&connectorID),
		},
	)
	require.NoError(t, err)
	require.NotNil(t, updated.ConnectorAccountID)
	assert.NotEqual(t, enabledID, *updated.ConnectorAccountID)

	initial := &coredata.ConnectorAccount{}

	require.NoError(t, env.client.WithConn(
		env.ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return initial.LoadByConnectorAndExternalID(ctx, conn, env.scope, connectorID, "acme")
		},
	))
	assert.Equal(t, initial.ID, *updated.ConnectorAccountID)

	stillEnabled := &coredata.AccessReviewSource{}

	require.NoError(t, env.client.WithConn(
		env.ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return stillEnabled.LoadByConnectorAccountID(ctx, conn, env.scope, enabledID)
		},
	))
	assert.Equal(t, enabledSource.ID, stillEnabled.ID)

	_, err = env.svc.UpdateSource(
		env.ctx,
		env.scope,
		accessreview.UpdateAccessReviewSourceRequest{
			AccessReviewSourceID: secondCSV.ID,
			ConnectorID:          new(&connectorID),
		},
	)
	require.Error(t, err)
	assert.ErrorContains(t, err, "connector account already referenced by another source")
}

func TestEnsureSource_StoresResolvedAccount(t *testing.T) {
	t.Parallel()

	env := newAccessSourceEnv(t)
	connectorID := env.insertConnector(t, coredata.ConnectorProviderGitHub)
	accountID := env.insertAccount(t, connectorID, "chosen")

	source, created, err := env.svc.EnsureSource(
		env.ctx,
		env.scope,
		accessreview.CreateAccessReviewSourceRequest{
			OrganizationID:     env.organizationID,
			ConnectorAccountID: &accountID,
			Name:               "account only",
		},
	)
	require.NoError(t, err)
	require.True(t, created)
	require.NotNil(t, source.ConnectorAccountID)
	assert.Equal(t, accountID, *source.ConnectorAccountID)

	again, created, err := env.svc.EnsureSource(
		env.ctx,
		env.scope,
		accessreview.CreateAccessReviewSourceRequest{
			OrganizationID:     env.organizationID,
			ConnectorAccountID: &accountID,
			Name:               "account only again",
		},
	)
	require.NoError(t, err)
	assert.False(t, created)
	assert.Equal(t, source.ID, again.ID)

	resolvedConnectorID := env.insertConnector(t, coredata.ConnectorProviderSlack)
	resolvedAccountID := env.insertAccount(t, resolvedConnectorID, "only")

	byConnector, created, err := env.svc.EnsureSource(
		env.ctx,
		env.scope,
		accessreview.CreateAccessReviewSourceRequest{
			OrganizationID: env.organizationID,
			ConnectorID:    &resolvedConnectorID,
			Name:           "connector only",
		},
	)
	require.NoError(t, err)
	require.True(t, created)
	require.NotNil(t, byConnector.ConnectorAccountID)
	assert.Equal(t, resolvedAccountID, *byConnector.ConnectorAccountID)

	env.insertAccount(t, resolvedConnectorID, "second")

	_, _, err = env.svc.EnsureSource(
		env.ctx,
		env.scope,
		accessreview.CreateAccessReviewSourceRequest{
			OrganizationID: env.organizationID,
			ConnectorID:    &resolvedConnectorID,
			Name:           "ambiguous",
		},
	)
	require.ErrorIs(t, err, coredata.ErrMultipleConnectorAccounts)

	freshConnectorID := env.insertConnector(t, coredata.ConnectorProviderLinear)
	freshAccountID := env.insertAccount(t, freshConnectorID, "fresh")
	unrelatedConnectorID := env.insertConnector(t, coredata.ConnectorProviderNotion)

	stored, created, err := env.svc.EnsureSource(
		env.ctx,
		env.scope,
		accessreview.CreateAccessReviewSourceRequest{
			OrganizationID:     env.organizationID,
			ConnectorID:        &unrelatedConnectorID,
			ConnectorAccountID: &freshAccountID,
			Name:               "account wins",
		},
	)
	require.NoError(t, err)
	require.True(t, created)
	require.NotNil(t, stored.ConnectorAccountID)
	assert.Equal(t, freshAccountID, *stored.ConnectorAccountID)
}

func TestEnsureSource_RejectsForeignOrganization(t *testing.T) {
	t.Parallel()

	env := newAccessSourceEnv(t)
	foreignOrganizationID := gid.New(env.scope.GetTenantID(), coredata.OrganizationEntityType)

	require.NoError(t, env.client.WithTx(
		env.ctx,
		func(ctx context.Context, tx pg.Tx) error {
			org := &coredata.Organization{
				ID:        foreignOrganizationID,
				TenantID:  env.scope.GetTenantID(),
				Name:      "Foreign organization",
				CreatedAt: env.now,
				UpdatedAt: env.now,
			}

			return org.Insert(ctx, tx)
		},
	))

	foreignConnectorID := gid.New(env.scope.GetTenantID(), coredata.ConnectorEntityType)

	var key cipher.EncryptionKey

	require.NoError(t, env.client.WithTx(
		env.ctx,
		func(ctx context.Context, tx pg.Tx) error {
			cnnctr := &coredata.Connector{
				ID:             foreignConnectorID,
				OrganizationID: foreignOrganizationID,
				Provider:       coredata.ConnectorProviderGitHub,
				Protocol:       coredata.ConnectorProtocolOAuth2,
				Connection: &connector.OAuth2Connection{
					AccessToken: "test-token",
					TokenType:   "Bearer",
				},
				CreatedAt: env.now,
				UpdatedAt: env.now,
			}

			return cnnctr.Insert(ctx, tx, env.scope, key)
		},
	))

	foreignAccount := &coredata.ConnectorAccount{
		ID:                gid.New(env.scope.GetTenantID(), coredata.ConnectorAccountEntityType),
		OrganizationID:    foreignOrganizationID,
		ConnectorID:       foreignConnectorID,
		ExternalAccountID: "foreign",
		Name:              "foreign",
		CreatedAt:         env.now,
		UpdatedAt:         env.now,
	}

	require.NoError(t, env.client.WithTx(
		env.ctx,
		func(ctx context.Context, tx pg.Tx) error {
			_, err := foreignAccount.Upsert(ctx, tx, env.scope)

			return err
		},
	))

	_, _, err := env.svc.EnsureSource(
		env.ctx,
		env.scope,
		accessreview.CreateAccessReviewSourceRequest{
			OrganizationID:     env.organizationID,
			ConnectorAccountID: &foreignAccount.ID,
			Name:               "foreign account",
		},
	)
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)

	_, _, err = env.svc.EnsureSource(
		env.ctx,
		env.scope,
		accessreview.CreateAccessReviewSourceRequest{
			OrganizationID: env.organizationID,
			ConnectorID:    &foreignConnectorID,
			Name:           "foreign connector",
		},
	)
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)

	var count int

	require.NoError(t, env.client.WithConn(
		env.ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var sources coredata.AccessReviewSources

			var countErr error

			count, countErr = sources.CountByOrganizationID(ctx, conn, env.scope, env.organizationID)

			return countErr
		},
	))
	assert.Equal(t, 0, count)

	csvSource := insertAccessSource(t, env.ctx, env.client, env.scope, &coredata.AccessReviewSource{
		ID:             gid.New(env.scope.GetTenantID(), coredata.AccessReviewSourceEntityType),
		OrganizationID: env.organizationID,
		Name:           "csv source",
		CreatedAt:      env.now,
		UpdatedAt:      env.now,
	})

	_, err = env.svc.UpdateSource(
		env.ctx,
		env.scope,
		accessreview.UpdateAccessReviewSourceRequest{
			AccessReviewSourceID: csvSource.ID,
			ConnectorID:          new(&foreignConnectorID),
		},
	)
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)

	reloaded := &coredata.AccessReviewSource{}

	require.NoError(t, env.client.WithConn(
		env.ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return reloaded.LoadByID(ctx, conn, env.scope, csvSource.ID)
		},
	))
	assert.Nil(t, reloaded.ConnectorAccountID)
}

func TestSource_RefusesConnectorWithNoAccount(t *testing.T) {
	t.Parallel()

	env := newAccessSourceEnv(t)
	connectorID := env.insertConnector(t, coredata.ConnectorProviderGitHub)

	_, _, err := env.svc.EnsureSource(
		env.ctx,
		env.scope,
		accessreview.CreateAccessReviewSourceRequest{
			OrganizationID: env.organizationID,
			ConnectorID:    &connectorID,
			Name:           "missing account",
		},
	)
	require.ErrorIs(t, err, accessreview.ErrNoConnectorAccount)

	csvSource := insertAccessSource(t, env.ctx, env.client, env.scope, &coredata.AccessReviewSource{
		ID:             gid.New(env.scope.GetTenantID(), coredata.AccessReviewSourceEntityType),
		OrganizationID: env.organizationID,
		Name:           "csv source",
		CreatedAt:      env.now,
		UpdatedAt:      env.now,
	})

	_, err = env.svc.UpdateSource(
		env.ctx,
		env.scope,
		accessreview.UpdateAccessReviewSourceRequest{
			AccessReviewSourceID: csvSource.ID,
			ConnectorID:          new(&connectorID),
		},
	)
	require.ErrorIs(t, err, accessreview.ErrNoConnectorAccount)

	var count int

	require.NoError(t, env.client.WithConn(
		env.ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var accounts coredata.ConnectorAccounts

			var countErr error

			count, countErr = accounts.CountByConnectorID(ctx, conn, env.scope, connectorID)

			return countErr
		},
	))
	assert.Equal(t, 0, count)

	namedID := gid.New(env.scope.GetTenantID(), coredata.ConnectorEntityType)

	require.NoError(t, env.client.WithTx(
		env.ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var key cipher.EncryptionKey

			cnnctr := &coredata.Connector{
				ID:             namedID,
				OrganizationID: env.organizationID,
				Provider:       coredata.ConnectorProviderGitHub,
				Protocol:       coredata.ConnectorProtocolOAuth2,
				Connection: &connector.OAuth2Connection{
					AccessToken: "test-token",
					TokenType:   "Bearer",
				},
				CreatedAt: env.now,
				UpdatedAt: env.now,
			}

			if err := cnnctr.SetSettings(&coredata.GitHubConnectorSettings{Organization: "acme"}); err != nil {
				return err
			}

			return cnnctr.Insert(ctx, tx, env.scope, key)
		},
	))

	_, _, err = env.svc.EnsureSource(
		env.ctx,
		env.scope,
		accessreview.CreateAccessReviewSourceRequest{
			OrganizationID: env.organizationID,
			ConnectorID:    &namedID,
			Name:           "named but missing",
		},
	)
	require.ErrorIs(t, err, accessreview.ErrNoConnectorAccount)

	require.NoError(t, env.client.WithConn(
		env.ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var accounts coredata.ConnectorAccounts

			var countErr error

			count, countErr = accounts.CountByConnectorID(ctx, conn, env.scope, namedID)

			return countErr
		},
	))
	assert.Equal(t, 0, count)
}

func TestDeleteSource_KeepsConnectorWhenOtherAccountHasSource(t *testing.T) {
	t.Parallel()

	env := newAccessSourceEnv(t)
	connectorID := env.insertConnector(t, coredata.ConnectorProviderGitHub)
	keptAccountID := env.insertAccount(t, connectorID, "kept")
	droppedAccountID := env.insertAccount(t, connectorID, "dropped")
	kept := env.insertLinkedSource(t, keptAccountID, "kept source")
	dropped := env.insertLinkedSource(t, droppedAccountID, "dropped source")

	require.NoError(t, env.svc.DeleteSource(env.ctx, env.scope, dropped.ID))

	requireMissingSource(t, env, dropped.ID)
	requirePresentSource(t, env, kept.ID)
	requireConnectorPresent(t, env, connectorID)
}

func TestDeleteSource_KeepsConnectorWhenLastSource(t *testing.T) {
	t.Parallel()

	env := newAccessSourceEnv(t)
	connectorID := env.insertConnector(t, coredata.ConnectorProviderGitHub)
	accountID := env.insertAccount(t, connectorID, "only")
	source := env.insertLinkedSource(t, accountID, "only source")

	require.NoError(t, env.svc.DeleteSource(env.ctx, env.scope, source.ID))

	requireMissingSource(t, env, source.ID)
	requireConnectorPresent(t, env, connectorID)
}

func TestUpdateSource_RelinkKeepsConnectorWhenOtherAccountHasSource(t *testing.T) {
	t.Parallel()

	env := newAccessSourceEnv(t)
	connectorID := env.insertConnector(t, coredata.ConnectorProviderGitHub)
	keptAccountID := env.insertAccount(t, connectorID, "kept")
	movedAccountID := env.insertAccount(t, connectorID, "moved")
	kept := env.insertLinkedSource(t, keptAccountID, "kept source")
	moved := env.insertLinkedSource(t, movedAccountID, "moved source")

	var unlinked *gid.GID

	updated, err := env.svc.UpdateSource(
		env.ctx,
		env.scope,
		accessreview.UpdateAccessReviewSourceRequest{
			AccessReviewSourceID: moved.ID,
			ConnectorID:          &unlinked,
		},
	)
	require.NoError(t, err)
	assert.Nil(t, updated.ConnectorAccountID)

	requirePresentSource(t, env, kept.ID)
	requireConnectorPresent(t, env, connectorID)
}

func TestUpdateSource_RelinkKeepsAbandonedConnectorWhenLastSource(t *testing.T) {
	t.Parallel()

	env := newAccessSourceEnv(t)
	connectorID := env.insertConnector(t, coredata.ConnectorProviderGitHub)
	accountID := env.insertAccount(t, connectorID, "only")
	source := env.insertLinkedSource(t, accountID, "only source")

	var unlinked *gid.GID

	updated, err := env.svc.UpdateSource(
		env.ctx,
		env.scope,
		accessreview.UpdateAccessReviewSourceRequest{
			AccessReviewSourceID: source.ID,
			ConnectorID:          &unlinked,
		},
	)
	require.NoError(t, err)
	assert.Nil(t, updated.ConnectorAccountID)

	requireConnectorPresent(t, env, connectorID)
}

func TestEnsureSource_RefusesBridgedConnector(t *testing.T) {
	t.Parallel()

	env := newAccessSourceEnv(t)
	connectorID := env.insertConnector(t, coredata.ConnectorProviderGoogleWorkspace)
	accountID := env.insertAccount(t, connectorID, "workspace")
	env.insertBridge(t, connectorID)

	_, _, err := env.svc.EnsureSource(
		env.ctx,
		env.scope,
		accessreview.CreateAccessReviewSourceRequest{
			OrganizationID:     env.organizationID,
			ConnectorAccountID: &accountID,
			Name:               "shared",
		},
	)
	require.ErrorIs(t, err, coredata.ErrResourceInUse)
	require.ErrorContains(t, err, "SCIM configuration")
	requireConnectorPresent(t, env, connectorID)
}

func TestUpdateSource_RelinkRefusesBridgedConnector(t *testing.T) {
	t.Parallel()

	env := newAccessSourceEnv(t)
	bridgedID := env.insertConnector(t, coredata.ConnectorProviderGoogleWorkspace)
	env.insertAccount(t, bridgedID, "workspace")
	env.insertBridge(t, bridgedID)

	otherID := env.insertConnector(t, coredata.ConnectorProviderGitHub)
	otherAccountID := env.insertAccount(t, otherID, "acme")
	source := env.insertLinkedSource(t, otherAccountID, "movable")

	linked := &bridgedID

	updated, err := env.svc.UpdateSource(
		env.ctx,
		env.scope,
		accessreview.UpdateAccessReviewSourceRequest{
			AccessReviewSourceID: source.ID,
			ConnectorID:          &linked,
		},
	)
	require.ErrorIs(t, err, coredata.ErrResourceInUse)
	require.ErrorContains(t, err, "SCIM configuration")
	require.Nil(t, updated)
	requireConnectorPresent(t, env, bridgedID)
}

func TestDeleteSource_KeepsConnectorWhenBridgeReferencesIt(t *testing.T) {
	t.Parallel()

	env := newAccessSourceEnv(t)
	connectorID := env.insertConnector(t, coredata.ConnectorProviderGoogleWorkspace)
	accountID := env.insertAccount(t, connectorID, "only")
	source := env.insertLinkedSource(t, accountID, "only source")
	env.insertBridge(t, connectorID)

	require.NoError(t, env.svc.DeleteSource(env.ctx, env.scope, source.ID))

	requireMissingSource(t, env, source.ID)
	requireConnectorPresent(t, env, connectorID)
}

func insertAccessSource(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	scope coredata.Scoper,
	source *coredata.AccessReviewSource,
) *coredata.AccessReviewSource {
	t.Helper()

	require.NoError(t, client.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			inserted, err := source.Insert(ctx, tx, scope)
			if err != nil {
				return err
			}

			if !inserted {
				return fmt.Errorf("access source was not inserted")
			}

			return nil
		},
	))

	return source
}

type accessSourceEnv struct {
	ctx            context.Context
	client         *pg.Client
	scope          coredata.Scoper
	organizationID gid.GID
	now            time.Time
	svc            *accessreview.Service
}

func newAccessSourceEnv(t *testing.T) *accessSourceEnv {
	t.Helper()

	client := test.PGClient(t)
	ctx := context.Background()
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now().UTC()

	require.NoError(t, client.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			org := &coredata.Organization{
				ID:        organizationID,
				TenantID:  tenantID,
				Name:      "Access source connector ownership",
				CreatedAt: now,
				UpdatedAt: now,
			}

			return org.Insert(ctx, tx)
		},
	))

	var key cipher.EncryptionKey

	return &accessSourceEnv{
		ctx:            ctx,
		client:         client,
		scope:          scope,
		organizationID: organizationID,
		now:            now,
		svc: accessreview.NewService(
			client,
			key,
			nil,
			provider.NewBuiltinRegistry(),
			log.NewLogger(log.WithName("test")),
		),
	}
}

func (e *accessSourceEnv) insertConnector(
	t *testing.T,
	connectorProvider coredata.ConnectorProvider,
	settings ...any,
) gid.GID {
	t.Helper()

	require.LessOrEqual(t, len(settings), 1)

	connectorID := gid.New(e.scope.GetTenantID(), coredata.ConnectorEntityType)

	var key cipher.EncryptionKey

	require.NoError(t, e.client.WithTx(
		e.ctx,
		func(ctx context.Context, tx pg.Tx) error {
			cnnctr := &coredata.Connector{
				ID:             connectorID,
				OrganizationID: e.organizationID,
				Provider:       connectorProvider,
				Protocol:       coredata.ConnectorProtocolOAuth2,
				Connection: &connector.OAuth2Connection{
					AccessToken: "test-token",
					TokenType:   "Bearer",
				},
				CreatedAt: e.now,
				UpdatedAt: e.now,
			}

			if len(settings) == 1 {
				if err := cnnctr.SetSettings(settings[0]); err != nil {
					return err
				}
			}

			return cnnctr.Insert(ctx, tx, e.scope, key)
		},
	))

	return connectorID
}

func (e *accessSourceEnv) insertAccount(
	t *testing.T,
	connectorID gid.GID,
	externalID string,
) gid.GID {
	t.Helper()

	account := &coredata.ConnectorAccount{
		ID:                gid.New(e.scope.GetTenantID(), coredata.ConnectorAccountEntityType),
		OrganizationID:    e.organizationID,
		ConnectorID:       connectorID,
		ExternalAccountID: externalID,
		Name:              externalID,
		CreatedAt:         e.now,
		UpdatedAt:         e.now,
	}

	require.NoError(t, e.client.WithTx(
		e.ctx,
		func(ctx context.Context, tx pg.Tx) error {
			_, err := account.Upsert(ctx, tx, e.scope)

			return err
		},
	))

	return account.ID
}

func (e *accessSourceEnv) insertLinkedSource(
	t *testing.T,
	accountID gid.GID,
	name string,
) *coredata.AccessReviewSource {
	t.Helper()

	return insertAccessSource(t, e.ctx, e.client, e.scope, &coredata.AccessReviewSource{
		ID:                 gid.New(e.scope.GetTenantID(), coredata.AccessReviewSourceEntityType),
		OrganizationID:     e.organizationID,
		ConnectorAccountID: &accountID,
		Name:               name,
		CreatedAt:          e.now,
		UpdatedAt:          e.now,
	})
}

func (e *accessSourceEnv) insertBridge(t *testing.T, connectorID gid.GID) {
	t.Helper()

	require.NoError(t, e.client.WithTx(
		e.ctx,
		func(ctx context.Context, tx pg.Tx) error {
			config := &coredata.SCIMConfiguration{
				ID:             gid.New(e.scope.GetTenantID(), coredata.SCIMConfigurationEntityType),
				OrganizationID: e.organizationID,
				HashedToken:    []byte{0x01},
				CreatedAt:      e.now,
				UpdatedAt:      e.now,
			}

			if err := config.Insert(ctx, tx, e.scope); err != nil {
				return err
			}

			bridge := &coredata.SCIMBridge{
				ID:                  gid.New(e.scope.GetTenantID(), coredata.SCIMBridgeEntityType),
				OrganizationID:      e.organizationID,
				ScimConfigurationID: config.ID,
				ConnectorID:         &connectorID,
				Type:                coredata.SCIMBridgeTypeGoogleWorkspace,
				State:               coredata.SCIMBridgeStateActive,
				ExcludedUserNames:   []string{},
				CreatedAt:           e.now,
				UpdatedAt:           e.now,
			}

			return bridge.Insert(ctx, tx, e.scope)
		},
	))
}

func requirePresentSource(t *testing.T, env *accessSourceEnv, sourceID gid.GID) {
	t.Helper()

	loaded := &coredata.AccessReviewSource{}

	require.NoError(t, env.client.WithConn(
		env.ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return loaded.LoadByID(ctx, conn, env.scope, sourceID)
		},
	))
}

func requireMissingSource(t *testing.T, env *accessSourceEnv, sourceID gid.GID) {
	t.Helper()

	loaded := &coredata.AccessReviewSource{}

	err := env.client.WithConn(
		env.ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return loaded.LoadByID(ctx, conn, env.scope, sourceID)
		},
	)
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)
}

func requireConnectorPresent(t *testing.T, env *accessSourceEnv, connectorID gid.GID) {
	t.Helper()

	loaded := &coredata.Connector{}

	require.NoError(t, env.client.WithConn(
		env.ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return loaded.LoadMetadataByID(ctx, conn, env.scope, connectorID)
		},
	))
}
