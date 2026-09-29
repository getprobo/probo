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
	require.ErrorIs(t, err, coredata.ErrResourceInUse)
	assert.ErrorContains(t, err, "connector account already referenced by another source")
}

func TestUpdateSource_AllowsSameExternalAccountOnAnotherConnector(t *testing.T) {
	t.Parallel()

	env := newAccessSourceEnv(t)
	firstConnectorID := env.insertConnector(
		t,
		coredata.ConnectorProviderGitHub,
		&coredata.GitHubConnectorSettings{Organization: "acme"},
	)
	secondConnectorID := env.insertConnector(
		t,
		coredata.ConnectorProviderGitHub,
		&coredata.GitHubConnectorSettings{Organization: "acme"},
	)
	firstAccountID := env.insertAccount(t, firstConnectorID, "acme")
	secondAccountID := env.insertAccount(t, secondConnectorID, "acme")

	owned := env.insertLinkedSource(t, firstAccountID, "owned")
	other := insertAccessSource(t, env.ctx, env.client, env.scope, &coredata.AccessReviewSource{
		ID:             gid.New(env.scope.GetTenantID(), coredata.AccessReviewSourceEntityType),
		OrganizationID: env.organizationID,
		Name:           "other",
		CreatedAt:      env.now,
		UpdatedAt:      env.now,
	})

	updated, err := env.svc.UpdateSource(
		env.ctx,
		env.scope,
		accessreview.UpdateAccessReviewSourceRequest{
			AccessReviewSourceID: other.ID,
			ConnectorID:          new(&secondConnectorID),
		},
	)
	require.NoError(t, err)
	require.NotNil(t, updated.ConnectorAccountID)
	assert.Equal(t, secondAccountID, *updated.ConnectorAccountID)

	stillOwned := &coredata.AccessReviewSource{}

	require.NoError(t, env.client.WithConn(
		env.ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return stillOwned.LoadByID(ctx, conn, env.scope, owned.ID)
		},
	))
	require.NotNil(t, stillOwned.ConnectorAccountID)
	assert.Equal(t, firstAccountID, *stillOwned.ConnectorAccountID)
}

func TestEnsureSource_StoresResolvedAccount(t *testing.T) {
	t.Parallel()

	env := newAccessSourceEnv(t)
	connectorID := env.insertConnector(t, coredata.ConnectorProviderGitHub)
	accountID := env.insertAccount(t, connectorID, "chosen")

	source, created, err := ensureOne(
		t,
		env,
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

	again, created, err := ensureOne(
		t,
		env,
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

	byConnector, created, err := ensureOne(
		t,
		env,
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

	_, _, err = ensureOne(
		t,
		env,
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

	stored, created, err := ensureOne(
		t,
		env,
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

	_, _, err := ensureOne(
		t,
		env,
		accessreview.CreateAccessReviewSourceRequest{
			OrganizationID:     env.organizationID,
			ConnectorAccountID: &foreignAccount.ID,
			Name:               "foreign account",
		},
	)
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)

	_, _, err = ensureOne(
		t,
		env,
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

	_, _, err := ensureOne(
		t,
		env,
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

	_, _, err = ensureOne(
		t,
		env,
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

func TestDeleteSource_DeletesConnectorWhenLastSource(t *testing.T) {
	t.Parallel()

	env := newAccessSourceEnv(t)
	connectorID := env.insertConnector(t, coredata.ConnectorProviderGitHub)
	accountID := env.insertAccount(t, connectorID, "only")
	source := env.insertLinkedSource(t, accountID, "only source")

	require.NoError(t, env.svc.DeleteSource(env.ctx, env.scope, source.ID))

	requireMissingSource(t, env, source.ID)
	requireConnectorMissing(t, env, connectorID)
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

func TestUpdateSource_RelinkDeletesAbandonedConnectorWhenLastSource(t *testing.T) {
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

	requireConnectorMissing(t, env, connectorID)
}

func TestEnsureSource_RefusesBridgedConnector(t *testing.T) {
	t.Parallel()

	env := newAccessSourceEnv(t)
	connectorID := env.insertConnector(t, coredata.ConnectorProviderGoogleWorkspace)
	accountID := env.insertAccount(t, connectorID, "workspace")
	env.insertBridge(t, connectorID)

	_, _, err := ensureOne(
		t,
		env,
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

func TestEnsureSources_AllowsImplicitPickerAccount(t *testing.T) {
	t.Parallel()

	env := newAccessSourceEnv(t)
	connectorID := env.insertConnector(t, coredata.ConnectorProviderGitHub)
	accountID := env.insertAccount(t, connectorID, connectorID.String())

	source, created, err := ensureOne(
		t,
		env,
		accessreview.CreateAccessReviewSourceRequest{
			OrganizationID:     env.organizationID,
			ConnectorAccountID: &accountID,
			Name:               "github",
		},
	)
	require.NoError(t, err)
	require.True(t, created)
	require.NotNil(t, source.ConnectorAccountID)
	assert.Equal(t, accountID, *source.ConnectorAccountID)
}

func TestEnsureSources_CreatesDistinctAccountsWithSameExternalID(t *testing.T) {
	t.Parallel()

	env := newAccessSourceEnv(t)
	firstConnectorID := env.insertConnector(t, coredata.ConnectorProviderGitHub)
	secondConnectorID := env.insertConnector(t, coredata.ConnectorProviderGitHub)
	otherConnectorID := env.insertConnector(t, coredata.ConnectorProviderSlack)
	firstAccountID := env.insertAccount(t, firstConnectorID, "acme")
	secondAccountID := env.insertAccount(t, secondConnectorID, "acme")
	otherAccountID := env.insertAccount(t, otherConnectorID, "acme")
	csvData := "email,name\njane@example.com,Jane"

	results, err := env.svc.EnsureSources(
		env.ctx,
		env.scope,
		env.organizationID,
		[]accessreview.CreateAccessReviewSourceRequest{
			{ConnectorAccountID: &firstAccountID, Name: "first"},
			{ConnectorAccountID: &secondAccountID, Name: "second"},
			{ConnectorAccountID: &otherAccountID, Name: "other provider"},
			{Name: "csv", CsvData: &csvData},
		},
	)
	require.NoError(t, err)
	require.Len(t, results, 4)

	require.True(t, results[0].Created)
	require.NotNil(t, results[0].Source.ConnectorAccountID)
	assert.Equal(t, firstAccountID, *results[0].Source.ConnectorAccountID)

	require.True(t, results[1].Created)
	assert.NotEqual(t, results[0].Source.ID, results[1].Source.ID)
	require.NotNil(t, results[1].Source.ConnectorAccountID)
	assert.Equal(t, secondAccountID, *results[1].Source.ConnectorAccountID)

	require.True(t, results[2].Created)
	assert.NotEqual(t, results[0].Source.ID, results[2].Source.ID)
	require.NotNil(t, results[2].Source.ConnectorAccountID)
	assert.Equal(t, otherAccountID, *results[2].Source.ConnectorAccountID)

	require.True(t, results[3].Created)
	assert.Nil(t, results[3].Source.ConnectorAccountID)
	require.NotNil(t, results[3].Source.CsvData)

	again, err := env.svc.EnsureSources(
		env.ctx,
		env.scope,
		env.organizationID,
		[]accessreview.CreateAccessReviewSourceRequest{
			{ConnectorAccountID: &secondAccountID, Name: "again"},
		},
	)
	require.NoError(t, err)
	require.Len(t, again, 1)
	assert.False(t, again[0].Created)
	assert.Equal(t, results[1].Source.ID, again[0].Source.ID)
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

// Counts and the connector delete are separate statements. Hold the connector
// row so a bridge can commit after the counts and before the delete.
func TestDeleteSource_IgnoresConnectorReferenceAddedDuringCleanup(t *testing.T) {
	t.Parallel()

	env := newAccessSourceEnv(t)
	connectorID := env.insertConnector(t, coredata.ConnectorProviderGoogleWorkspace)
	accountID := env.insertAccount(t, connectorID, "only")
	source := env.insertLinkedSource(t, accountID, "only source")

	deleted := make(chan error, 1)

	err := env.client.WithTx(
		env.ctx,
		func(ctx context.Context, tx pg.Tx) error {
			cnnctr := &coredata.Connector{ID: connectorID}
			if err := cnnctr.LockByID(ctx, tx, env.scope); err != nil {
				return err
			}

			go func() {
				deleted <- env.svc.DeleteSource(context.Background(), env.scope, source.ID)
			}()

			deadline := time.Now().Add(5 * time.Second)

			for {
				var blocked bool

				err := tx.QueryRow(
					ctx,
					`
SELECT EXISTS (
    SELECT 1
    FROM pg_stat_activity
    WHERE pg_backend_pid() = ANY (pg_blocking_pids(pid))
)
`,
				).Scan(&blocked)
				if err != nil {
					return fmt.Errorf("cannot check connector lock waiters: %w", err)
				}

				if blocked {
					break
				}

				if time.Now().After(deadline) {
					return fmt.Errorf("connector delete never waited on the row lock")
				}

				time.Sleep(10 * time.Millisecond)
			}

			config := &coredata.SCIMConfiguration{
				ID:             gid.New(env.scope.GetTenantID(), coredata.SCIMConfigurationEntityType),
				OrganizationID: env.organizationID,
				HashedToken:    []byte{0x01},
				CreatedAt:      env.now,
				UpdatedAt:      env.now,
			}
			if err := config.Insert(ctx, tx, env.scope); err != nil {
				return err
			}

			bridge := &coredata.SCIMBridge{
				ID:                  gid.New(env.scope.GetTenantID(), coredata.SCIMBridgeEntityType),
				OrganizationID:      env.organizationID,
				ScimConfigurationID: config.ID,
				ConnectorID:         &connectorID,
				Type:                coredata.SCIMBridgeTypeGoogleWorkspace,
				State:               coredata.SCIMBridgeStateActive,
				ExcludedUserNames:   []string{},
				CreatedAt:           env.now,
				UpdatedAt:           env.now,
			}

			return bridge.Insert(ctx, tx, env.scope)
		},
	)
	require.NoError(t, err)

	select {
	case deleteErr := <-deleted:
		require.NoError(t, deleteErr)
	case <-time.After(5 * time.Second):
		t.Fatal("DeleteSource did not finish after the bridge committed")
	}

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

func TestSourceNeedsReconnect_Slack(t *testing.T) {
	t.Parallel()

	env := newAccessSourceEnv(t)

	insertSlack := func(t *testing.T, tokenType string, scope string, channelID string) gid.GID {
		t.Helper()

		connectorID := gid.New(env.scope.GetTenantID(), coredata.ConnectorEntityType)

		var key cipher.EncryptionKey

		require.NoError(t, env.client.WithTx(
			env.ctx,
			func(ctx context.Context, tx pg.Tx) error {
				conn := &connector.SlackConnection{
					AccessToken: "xox-test",
					TokenType:   tokenType,
					Scope:       scope,
				}
				// Insert stores the channel in the connector settings.
				conn.Settings.ChannelID = channelID

				cnnctr := &coredata.Connector{
					ID:             connectorID,
					OrganizationID: env.organizationID,
					Provider:       coredata.ConnectorProviderSlack,
					Protocol:       coredata.ConnectorProtocolOAuth2,
					Connection:     conn,
					CreatedAt:      env.now,
					UpdatedAt:      env.now,
				}

				return cnnctr.Insert(ctx, tx, env.scope, key)
			},
		))

		return connectorID
	}

	for name, tc := range map[string]struct {
		tokenType string
		scope     string
		channelID string
		want      bool
	}{
		"user token":           {tokenType: "user", scope: "users:read,users:read.email", want: false},
		"bot token":            {tokenType: "bot", scope: "users:read,users:read.email", want: true},
		"legacy messaging bot": {tokenType: "bot", scope: "chat:write,incoming-webhook", channelID: "C123", want: false},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			connectorID := insertSlack(t, tc.tokenType, tc.scope, tc.channelID)

			got, err := env.svc.SourceNeedsReconnect(env.ctx, env.scope, connectorID)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestSetConnectorOrganization_GitHubApp(t *testing.T) {
	t.Parallel()

	env := newAccessSourceEnv(t)
	connectorID := gid.New(env.scope.GetTenantID(), coredata.ConnectorEntityType)

	var key cipher.EncryptionKey

	require.NoError(t, env.client.WithTx(
		env.ctx,
		func(ctx context.Context, tx pg.Tx) error {
			cnnctr := &coredata.Connector{
				ID:             connectorID,
				OrganizationID: env.organizationID,
				Provider:       coredata.ConnectorProviderGitHub,
				Protocol:       coredata.ConnectorProtocolGitHubApp,
				Connection:     &connector.GitHubAppConnection{InstallationID: 42},
				CreatedAt:      env.now,
				UpdatedAt:      env.now,
			}

			if err := cnnctr.SetSettings(&coredata.GitHubConnectorSettings{Organization: "acme"}); err != nil {
				return err
			}

			return cnnctr.Insert(ctx, tx, env.scope, key)
		},
	))

	err := env.svc.SetConnectorOrganization(env.ctx, env.scope, connectorID, "other")
	require.ErrorIs(t, err, accessreview.ErrOrganizationPickerUnsupported)

	stored := &coredata.Connector{}

	require.NoError(t, env.client.WithConn(
		env.ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return stored.LoadMetadataByID(ctx, conn, env.scope, connectorID)
		},
	))

	settings, err := coredata.ConnectorSettings[coredata.GitHubConnectorSettings](stored)
	require.NoError(t, err)
	assert.Equal(t, "acme", settings.Organization)
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

func ensureOne(
	t *testing.T,
	env *accessSourceEnv,
	req accessreview.CreateAccessReviewSourceRequest,
) (*coredata.AccessReviewSource, bool, error) {
	t.Helper()

	results, err := env.svc.EnsureSources(
		env.ctx,
		env.scope,
		env.organizationID,
		[]accessreview.CreateAccessReviewSourceRequest{req},
	)
	if err != nil {
		return nil, false, err
	}

	require.Len(t, results, 1)

	return results[0].Source, results[0].Created, nil
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

func requireConnectorMissing(t *testing.T, env *accessSourceEnv, connectorID gid.GID) {
	t.Helper()

	loaded := &coredata.Connector{}

	err := env.client.WithConn(
		env.ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return loaded.LoadMetadataByID(ctx, conn, env.scope, connectorID)
		},
	)
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)
}
