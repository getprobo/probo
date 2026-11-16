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

package itam_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/cloud"
	aws "go.probo.inc/probo/pkg/cloud/aws"
	"go.probo.inc/probo/pkg/connectorclient"
	"go.probo.inc/probo/pkg/connectorprotocol"
	"go.probo.inc/probo/pkg/connectorprotocol/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/identityfederation"
	"go.probo.inc/probo/pkg/itam"
)

const (
	distributionA = "arn:aws:cloudfront::123456789012:distribution/AAAAAAAAAAAAAA"
	distributionB = "arn:aws:cloudfront::123456789012:distribution/BBBBBBBBBBBBBB"
)

func TestAssetSync_TwoAccountsProduceTwoSources(t *testing.T) {
	t.Parallel()

	pgClient, seeded := seedAWSAccount(t, "222222222222")
	second := insertAccount(t, pgClient, seeded, "333333333333")
	noteAccount(t, pgClient, seeded.scope, seeded.account)
	noteAccount(t, pgClient, seeded.scope, second)

	firstSource := loadSource(t, pgClient, seeded.scope, seeded.account)
	secondSource := loadSource(t, pgClient, seeded.scope, second)

	assert.NotEqual(t, firstSource.ID, secondSource.ID)
	require.NotNil(t, firstSource.ConnectorAccountID)
	require.NotNil(t, secondSource.ConnectorAccountID)
	assert.Equal(t, seeded.account.ID, *firstSource.ConnectorAccountID)
	assert.Equal(t, second.ID, *secondSource.ConnectorAccountID)

	handler := itam.NewAssetSyncHandlerForTest(pgClient, testOpener(t), emptyList)
	firstSync, err := itam.ClaimSourceForTest(handler, t.Context(), firstSource.ID)
	require.NoError(t, err)
	secondSync, err := itam.ClaimSourceForTest(handler, t.Context(), secondSource.ID)
	require.NoError(t, err)

	assert.NotEqual(t, firstSync.ID, secondSync.ID)
	assert.Equal(t, coredata.AssetSyncStatusRunning, firstSync.Status)
	assert.Equal(t, coredata.AssetSyncStatusRunning, secondSync.Status)
	assert.Equal(t, firstSource.ID, firstSync.AssetSourceID)
	assert.Equal(t, secondSource.ID, secondSync.AssetSourceID)
}

func TestAssetSync_SecondListUpdatesSameAsset(t *testing.T) {
	t.Parallel()

	pgClient, seeded := seedAWSAccount(t, "222222222222")
	noteAccount(t, pgClient, seeded.scope, seeded.account)
	source := loadSource(t, pgClient, seeded.scope, seeded.account)

	current := []itam.AssetObservation{distribution(distributionA, "one.example.net")}
	handler := itam.NewAssetSyncHandlerForTest(pgClient, testOpener(t), func(
		context.Context,
		*aws.Session,
	) ([]itam.AssetObservation, error) {
		return current, nil
	})
	_, err := itam.SyncSourceForTest(handler, t.Context(), source.ID)
	require.NoError(t, err)

	first := loadInventoryAsset(t, pgClient, seeded.scope, source.ID, distributionA)
	time.Sleep(2 * time.Millisecond)

	current[0] = distribution(distributionA, "two.example.net")
	makeDue(t, pgClient, seeded.scope, source)
	_, err = itam.SyncSourceForTest(handler, t.Context(), source.ID)
	require.NoError(t, err)

	second := loadInventoryAsset(t, pgClient, seeded.scope, source.ID, distributionA)
	assert.Equal(t, first.ID, second.ID)
	assert.Equal(t, "two.example.net", second.Name)
	assert.True(t, second.LastSeenAt.After(first.LastSeenAt))
	assert.Nil(t, second.Region)

	var configuration itam.CloudFrontDistributionConfiguration
	require.NoError(t, json.Unmarshal(second.Configuration, &configuration))
	assert.Equal(t, "two.example.net", configuration.DomainName)
}

func TestAssetSync_ReconnectReusesSource(t *testing.T) {
	t.Parallel()

	pgClient, seeded := seedAWSAccount(t, "222222222222")
	noteAccount(t, pgClient, seeded.scope, seeded.account)
	first := loadSource(t, pgClient, seeded.scope, seeded.account)

	reconnected := insertConnectorAccount(t, pgClient, seeded, "222222222222")
	noteAccount(t, pgClient, seeded.scope, reconnected)
	second := loadSource(t, pgClient, seeded.scope, reconnected)

	assert.Equal(t, first.ID, second.ID)
	require.NotNil(t, second.ConnectorAccountID)
	assert.Equal(t, reconnected.ID, *second.ConnectorAccountID)
}

func TestAssetSync_MissingDistributionKeepsLastSeen(t *testing.T) {
	t.Parallel()

	pgClient, seeded := seedAWSAccount(t, "222222222222")
	noteAccount(t, pgClient, seeded.scope, seeded.account)
	source := loadSource(t, pgClient, seeded.scope, seeded.account)

	current := []itam.AssetObservation{
		distribution(distributionA, "one.example.net"),
		distribution(distributionB, "two.example.net"),
	}
	handler := itam.NewAssetSyncHandlerForTest(pgClient, testOpener(t), func(
		context.Context,
		*aws.Session,
	) ([]itam.AssetObservation, error) {
		return current, nil
	})
	_, err := itam.SyncSourceForTest(handler, t.Context(), source.ID)
	require.NoError(t, err)

	kept := loadInventoryAsset(t, pgClient, seeded.scope, source.ID, distributionB)
	time.Sleep(2 * time.Millisecond)

	current = current[:1]
	makeDue(t, pgClient, seeded.scope, source)
	finished, err := itam.SyncSourceForTest(handler, t.Context(), source.ID)
	require.NoError(t, err)

	still := loadInventoryAsset(t, pgClient, seeded.scope, source.ID, distributionB)
	seen := loadInventoryAsset(t, pgClient, seeded.scope, source.ID, distributionA)

	assert.True(t, still.LastSeenAt.Equal(kept.LastSeenAt))
	assert.True(t, seen.LastSeenAt.After(kept.LastSeenAt))
	assert.Equal(
		t,
		[]string{itam.CloudFrontDistributionKind},
		reloadSync(t, pgClient, seeded.scope, finished.ID).CompletedKinds,
	)
}

func TestAssetSync_PageCapDoesNotAdvanceLastSeen(t *testing.T) {
	t.Parallel()

	pgClient, seeded := seedAWSAccount(t, "222222222222")
	noteAccount(t, pgClient, seeded.scope, seeded.account)
	source := loadSource(t, pgClient, seeded.scope, seeded.account)

	listed := []itam.AssetObservation{distribution(distributionA, "one.example.net")}
	var listErr error
	handler := itam.NewAssetSyncHandlerForTest(pgClient, testOpener(t), func(
		context.Context,
		*aws.Session,
	) ([]itam.AssetObservation, error) {
		return listed, listErr
	})
	_, err := itam.SyncSourceForTest(handler, t.Context(), source.ID)
	require.NoError(t, err)
	before := loadInventoryAsset(t, pgClient, seeded.scope, source.ID, distributionA)

	listErr = errors.New("cloudfront boom secret")
	makeDue(t, pgClient, seeded.scope, source)
	failed, err := itam.SyncSourceForTest(handler, t.Context(), source.ID)
	require.NoError(t, err)
	after := loadInventoryAsset(t, pgClient, seeded.scope, source.ID, distributionA)
	stored := reloadSync(t, pgClient, seeded.scope, failed.ID)

	assert.True(t, after.LastSeenAt.Equal(before.LastSeenAt))
	assert.Equal(t, coredata.AssetSyncStatusFailed, stored.Status)
	require.NotNil(t, stored.Error)
	assert.Equal(t, itam.AssetSyncFailureMessageForTest(), *stored.Error)
	assert.NotContains(t, *stored.Error, "secret")
	assert.NotContains(t, stored.CompletedKinds, itam.CloudFrontDistributionKind)
	reloaded := loadSource(t, pgClient, seeded.scope, seeded.account)
	assert.WithinDuration(t, time.Now().Add(24*time.Hour), reloaded.NextSyncAt, 2*time.Minute)
}

func TestAssetSync_DeleteAccountClearsSource(t *testing.T) {
	t.Parallel()

	pgClient, seeded := seedAWSAccount(t, "222222222222")
	noteAccount(t, pgClient, seeded.scope, seeded.account)
	source := loadSource(t, pgClient, seeded.scope, seeded.account)
	handler := itam.NewAssetSyncHandlerForTest(pgClient, testOpener(t), func(
		context.Context,
		*aws.Session,
	) ([]itam.AssetObservation, error) {
		return []itam.AssetObservation{distribution(distributionA, "one.example.net")}, nil
	})
	_, err := itam.SyncSourceForTest(handler, t.Context(), source.ID)
	require.NoError(t, err)
	asset := loadInventoryAsset(t, pgClient, seeded.scope, source.ID, distributionA)

	require.NoError(t, pgClient.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		return seeded.account.Delete(ctx, tx, seeded.scope)
	}))

	cleared := loadSource(t, pgClient, seeded.scope, seeded.account)
	kept := loadInventoryAsset(t, pgClient, seeded.scope, source.ID, distributionA)

	assert.Nil(t, cleared.ConnectorAccountID)
	assert.Equal(t, asset.ID, kept.ID)
	assert.True(t, kept.LastSeenAt.Equal(asset.LastSeenAt))
}

func TestAssetSync_StaleCapWaitsADay(t *testing.T) {
	t.Parallel()

	pgClient, seeded := seedAWSAccount(t, "222222222222")
	noteAccount(t, pgClient, seeded.scope, seeded.account)
	source := loadSource(t, pgClient, seeded.scope, seeded.account)
	handler := itam.NewAssetSyncHandlerForTest(pgClient, testOpener(t), emptyList)

	for attempt := 1; attempt <= 3; attempt++ {
		claimed, err := itam.ClaimSourceForTest(handler, t.Context(), source.ID)
		require.NoError(t, err)
		assert.Equal(t, attempt, claimed.AttemptNumber)
		require.NoError(t, itam.BackdateAssetSyncForTest(
			t.Context(),
			pgClient,
			claimed.ID,
			time.Now().Add(-time.Hour),
		))
		require.NoError(t, itam.RecoverStaleAssetSyncsForTest(handler, t.Context()))
		stored := reloadSync(t, pgClient, seeded.scope, claimed.ID)
		assert.Equal(t, coredata.AssetSyncStatusFailed, stored.Status)
		require.NotNil(t, stored.Error)
		assert.Equal(t, itam.AssetSyncStaleMessageForTest(), *stored.Error)
	}

	reloaded := loadSource(t, pgClient, seeded.scope, seeded.account)
	assert.WithinDuration(t, time.Now().Add(24*time.Hour), reloaded.NextSyncAt, 2*time.Minute)

	itam.SetAssetSyncSourceForTest(handler, source.ID)
	_, err := itam.ClaimAssetSyncForTest(handler, t.Context())
	require.ErrorIs(t, err, worker.ErrNoTask)
}

func emptyList(context.Context, *aws.Session) ([]itam.AssetObservation, error) {
	return nil, nil
}

func distribution(externalID string, name string) itam.AssetObservation {
	return itam.AssetObservation{
		ExternalID: externalID,
		Name:       name,
		Configuration: itam.CloudFrontDistributionConfiguration{
			DomainName:         name,
			Enabled:            true,
			Status:             "Deployed",
			PriceClass:         "PriceClass_100",
			DefaultCertificate: true,
		},
		SchemaVersion: itam.CloudFrontDistributionSchemaVersion,
	}
}

type seededAWS struct {
	scope     coredata.Scoper
	connector *coredata.Connector
	account   *coredata.ConnectorAccount
}

func seedAWSAccount(t *testing.T, externalID string) (*pg.Client, seededAWS) {
	t.Helper()

	pgClient := test.PGClient(t)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID, gid.Nil)
	now := time.Now().UTC()
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)

	var key cipher.EncryptionKey
	cnnctr := &coredata.Connector{
		ID:             gid.New(tenantID, coredata.ConnectorEntityType),
		OrganizationID: organizationID,
		Provider:       coredata.ConnectorProviderAWS,
		Protocol:       coredata.ConnectorProtocolWorkloadIdentity,
		Name:           "Inventory",
		Connection:     &connectorprotocol.WorkloadIdentityConnection{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	account := &coredata.ConnectorAccount{
		ID:                gid.New(tenantID, coredata.ConnectorAccountEntityType),
		OrganizationID:    organizationID,
		ConnectorID:       cnnctr.ID,
		ExternalAccountID: externalID,
		Name:              externalID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	require.NoError(t, pgClient.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		org := &coredata.Organization{
			ID:        organizationID,
			TenantID:  tenantID,
			Name:      "Inventory test",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := org.Insert(ctx, tx); err != nil {
			return err
		}

		if err := cnnctr.Insert(ctx, tx, scope, key); err != nil {
			return err
		}

		_, err := account.Upsert(ctx, tx, scope)

		return err
	}))

	return pgClient, seededAWS{scope: scope, connector: cnnctr, account: account}
}

func insertAccount(
	t *testing.T,
	pgClient *pg.Client,
	seeded seededAWS,
	externalID string,
) *coredata.ConnectorAccount {
	t.Helper()

	now := time.Now().UTC()
	account := &coredata.ConnectorAccount{
		ID:                gid.New(seeded.scope.GetTenantID(), coredata.ConnectorAccountEntityType),
		OrganizationID:    seeded.account.OrganizationID,
		ConnectorID:       seeded.connector.ID,
		ExternalAccountID: externalID,
		Name:              externalID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	require.NoError(t, pgClient.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		_, err := account.Upsert(ctx, tx, seeded.scope)

		return err
	}))

	return account
}

func insertConnectorAccount(
	t *testing.T,
	pgClient *pg.Client,
	seeded seededAWS,
	externalID string,
) *coredata.ConnectorAccount {
	t.Helper()

	now := time.Now().UTC()
	var key cipher.EncryptionKey
	cnnctr := &coredata.Connector{
		ID:             gid.New(seeded.scope.GetTenantID(), coredata.ConnectorEntityType),
		OrganizationID: seeded.account.OrganizationID,
		Provider:       coredata.ConnectorProviderAWS,
		Protocol:       coredata.ConnectorProtocolWorkloadIdentity,
		Name:           "Reconnected",
		Connection:     &connectorprotocol.WorkloadIdentityConnection{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	account := &coredata.ConnectorAccount{
		ID:                gid.New(seeded.scope.GetTenantID(), coredata.ConnectorAccountEntityType),
		OrganizationID:    seeded.account.OrganizationID,
		ConnectorID:       cnnctr.ID,
		ExternalAccountID: externalID,
		Name:              externalID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	require.NoError(t, pgClient.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		if err := cnnctr.Insert(ctx, tx, seeded.scope, key); err != nil {
			return err
		}

		_, err := account.Upsert(ctx, tx, seeded.scope)

		return err
	}))

	return account
}

func noteAccount(
	t *testing.T,
	pgClient *pg.Client,
	scope coredata.Scoper,
	account *coredata.ConnectorAccount,
) {
	t.Helper()

	require.NoError(t, pgClient.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		return itam.NoteConnectorAccount(ctx, tx, scope, coredata.ConnectorProviderAWS, account)
	}))
}

func loadSource(
	t *testing.T,
	pgClient *pg.Client,
	scope coredata.Scoper,
	account *coredata.ConnectorAccount,
) *coredata.AssetSource {
	t.Helper()

	source := &coredata.AssetSource{}
	require.NoError(t, pgClient.WithConn(t.Context(), func(ctx context.Context, conn pg.Querier) error {
		return source.LoadByIdentity(
			ctx,
			conn,
			scope,
			account.OrganizationID,
			coredata.ConnectorProviderAWS,
			account.ExternalAccountID,
		)
	}))

	return source
}

func loadInventoryAsset(
	t *testing.T,
	pgClient *pg.Client,
	scope coredata.Scoper,
	sourceID gid.GID,
	externalID string,
) *coredata.InventoryAsset {
	t.Helper()

	asset := &coredata.InventoryAsset{}
	require.NoError(t, pgClient.WithConn(t.Context(), func(ctx context.Context, conn pg.Querier) error {
		return asset.LoadBySourceKindExternalID(
			ctx,
			conn,
			scope,
			sourceID,
			itam.CloudFrontDistributionKind,
			externalID,
		)
	}))

	return asset
}

func reloadSync(
	t *testing.T,
	pgClient *pg.Client,
	scope coredata.Scoper,
	id gid.GID,
) *coredata.AssetSync {
	t.Helper()

	syncRow := &coredata.AssetSync{}
	require.NoError(t, pgClient.WithConn(t.Context(), func(ctx context.Context, conn pg.Querier) error {
		return syncRow.LoadByID(ctx, conn, scope, id)
	}))

	return syncRow
}

func makeDue(
	t *testing.T,
	pgClient *pg.Client,
	scope coredata.Scoper,
	source *coredata.AssetSource,
) {
	t.Helper()

	require.NoError(t, pgClient.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		return source.SetNextSyncAt(ctx, tx, scope, time.Now().Add(-time.Minute))
	}))
}

func testOpener(t *testing.T) *connectorclient.Client {
	t.Helper()

	registries := provider.NewRegistry()
	err := registries.Register(&provider.Registration{
		Provider:    coredata.ConnectorProviderAWS,
		DisplayName: "Amazon Web Services",
		InitialAccountFunc: func(*coredata.Connector) (string, string, error) {
			return "111111111111", "hub", nil
		},
		WorkloadIdentity: &provider.WorkloadIdentityConfig{
			NewSession: func(
				context.Context,
				*identityfederation.Issuer,
				*coredata.Connector,
				string,
			) (cloud.Session, error) {
				return aws.NewSessionFromConfig("222222222222", "aws", awssdk.Config{}), nil
			},
			NewDriver: func(
				context.Context,
				cloud.Session,
				*coredata.Connector,
				*log.Logger,
			) (drivers.Driver, error) {
				return nil, nil
			},
		},
	})
	require.NoError(t, err)

	return connectorclient.New(
		cipher.EncryptionKey{},
		nil,
		registries,
		connectorprotocol.NewConnectorRegistry(),
	)
}
