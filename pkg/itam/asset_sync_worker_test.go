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
	"fmt"
	"slices"
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
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connectorprotocol"
	"go.probo.inc/probo/pkg/connectorprotocol/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/identityfederation"
	"go.probo.inc/probo/pkg/itam"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
)

func TestAssetSync_RecordsDiscoveredRow(t *testing.T) {
	t.Parallel()

	seeded := seedAWSAccount(t)
	externalID := "arn:aws:cloudfront::222222222222:distribution/E1"
	region := "us-east-1"

	syncAccount(
		t,
		seeded,
		map[string]itam.AssetLister{
			itam.CloudFrontDistributionKind: observations(
				itam.AssetObservation{
					ExternalID: externalID,
					Name:       "cdn",
					Region:     &region,
					Configuration: itam.CloudFrontDistributionConfiguration{
						DomainName:         "d.cloudfront.net",
						Enabled:            true,
						Status:             "Deployed",
						PriceClass:         "PriceClass_100",
						DefaultCertificate: true,
					},
					SchemaVersion: itam.CloudFrontDistributionSchemaVersion,
				},
			),
		},
	)

	asset := loadDiscovered(t, seeded, itam.CloudFrontDistributionKind, externalID)
	sync := loadLatestSync(t, seeded)

	assert.Equal(t, 1, asset.Amount)
	assert.Equal(t, coredata.AssetTypeVirtual, asset.Type)
	assert.Equal(t, coredata.AssetStatusActive, asset.Status)
	require.NotNil(t, asset.ConnectorID)
	assert.Equal(t, seeded.account.ConnectorID, *asset.ConnectorID)
	assert.NotEqual(t, seeded.account.ID, *asset.ConnectorID)
	assert.Nil(t, asset.OwnerID)
	assert.Nil(t, asset.DataTypesStored)
	assert.Equal(t, "cdn", asset.Name)
	require.NotNil(t, asset.Kind)
	assert.Equal(t, itam.CloudFrontDistributionKind, *asset.Kind)
	require.NotNil(t, asset.LastSeenAt)
	assert.Equal(t, coredata.AssetSyncStatusSucceeded, sync.Status)
	assert.Nil(t, sync.Error)
	assert.Equal(t, []string{
		itam.CloudFrontDistributionKind,
		itam.EC2InstanceKind,
		itam.RDSDBInstanceKind,
	}, sync.CompletedKinds)
	assert.Contains(t, string(asset.Configuration), "d.cloudfront.net")
}

func TestAssetSync_RefreshKeepsManualFields(t *testing.T) {
	t.Parallel()

	seeded := seedAWSAccount(t)
	externalID := "arn:aws:ec2:us-east-1:222222222222:instance/i-1"
	region := "us-east-1"
	listers := map[string]itam.AssetLister{
		itam.EC2InstanceKind: observations(itam.AssetObservation{
			ExternalID:    externalID,
			Name:          "web",
			Region:        &region,
			Configuration: itam.EC2InstanceConfiguration{InstanceType: "t3.micro", State: "running"},
			SchemaVersion: itam.EC2InstanceSchemaVersion,
		}),
	}

	syncAccount(t, seeded, listers)

	asset := loadDiscovered(t, seeded, itam.EC2InstanceKind, externalID)
	ownerID := insertOwner(t, seeded)
	dataTypes := "customer records"
	seen := time.Now().Add(-time.Hour).UTC().Truncate(time.Microsecond)
	setManualFields(t, seeded, asset.ID, ownerID, 4, dataTypes, coredata.AssetStatusDeactivated, seen)
	makeAccountDue(t, seeded)

	listers[itam.EC2InstanceKind] = observations(itam.AssetObservation{
		ExternalID:    externalID,
		Name:          "web-renamed",
		Region:        &region,
		Configuration: itam.EC2InstanceConfiguration{InstanceType: "t3.small", State: "running"},
		SchemaVersion: itam.EC2InstanceSchemaVersion,
	})
	syncAccount(t, seeded, listers)

	refreshed := loadDiscovered(t, seeded, itam.EC2InstanceKind, externalID)

	assert.Equal(t, asset.ID, refreshed.ID)
	assert.Equal(t, 4, refreshed.Amount)
	require.NotNil(t, refreshed.OwnerID)
	assert.Equal(t, ownerID, *refreshed.OwnerID)
	require.NotNil(t, refreshed.DataTypesStored)
	assert.Equal(t, dataTypes, *refreshed.DataTypesStored)
	assert.Equal(t, coredata.AssetStatusDeactivated, refreshed.Status)
	assert.Equal(t, "web-renamed", refreshed.Name)
	require.NotNil(t, refreshed.LastSeenAt)
	assert.True(t, refreshed.LastSeenAt.After(seen))
	assert.Contains(t, string(refreshed.Configuration), "t3.small")
}

func TestAssetSync_SecondAccountUpdatesSameConnectorRow(t *testing.T) {
	t.Parallel()

	seeded := seedAWSAccount(t)
	second := addAccount(t, seeded, "333333333333")
	externalID := "arn:aws:rds:eu-west-1:222222222222:db:app"
	region := "eu-west-1"

	syncAccount(t, seeded, map[string]itam.AssetLister{
		itam.RDSDBInstanceKind: observations(itam.AssetObservation{
			ExternalID:    externalID,
			Name:          "app",
			Region:        &region,
			Configuration: itam.RDSDBInstanceConfiguration{Engine: "postgres", Status: "available"},
			SchemaVersion: itam.RDSDBInstanceSchemaVersion,
		}),
	})
	syncAccount(t, second, map[string]itam.AssetLister{
		itam.RDSDBInstanceKind: observations(itam.AssetObservation{
			ExternalID:    externalID,
			Name:          "app-from-second-account",
			Region:        &region,
			Configuration: itam.RDSDBInstanceConfiguration{Engine: "postgres", Status: "available"},
			SchemaVersion: itam.RDSDBInstanceSchemaVersion,
		}),
	})

	asset := loadDiscovered(t, seeded, itam.RDSDBInstanceKind, externalID)

	assert.Equal(t, "app-from-second-account", asset.Name)
	require.NotNil(t, asset.ConnectorID)
	assert.Equal(t, seeded.account.ConnectorID, *asset.ConnectorID)
}

func TestAssetSync_MissingAssetKeepsLastSeen(t *testing.T) {
	t.Parallel()

	seeded := seedAWSAccount(t)
	externalID := "arn:aws:cloudfront::222222222222:distribution/E2"

	syncAccount(t, seeded, map[string]itam.AssetLister{
		itam.CloudFrontDistributionKind: observations(itam.AssetObservation{
			ExternalID:    externalID,
			Name:          "old",
			Configuration: itam.CloudFrontDistributionConfiguration{DomainName: "old.cloudfront.net"},
			SchemaVersion: itam.CloudFrontDistributionSchemaVersion,
		}),
	})

	asset := loadDiscovered(t, seeded, itam.CloudFrontDistributionKind, externalID)
	seen := time.Now().Add(-2 * time.Hour).UTC().Truncate(time.Microsecond)
	setLastSeen(t, seeded, asset.ID, seen)
	makeAccountDue(t, seeded)

	syncAccount(t, seeded, map[string]itam.AssetLister{})

	kept := loadDiscovered(t, seeded, itam.CloudFrontDistributionKind, externalID)

	require.NotNil(t, kept.LastSeenAt)
	assert.True(t, kept.LastSeenAt.Equal(seen))
	assert.Equal(t, "old", kept.Name)
}

func TestAssetSync_FailedKindDoesNotMoveLastSeen(t *testing.T) {
	t.Parallel()

	seeded := seedAWSAccount(t)
	distributionID := "arn:aws:cloudfront::222222222222:distribution/E3"
	instanceID := "arn:aws:ec2:us-east-1:222222222222:instance/i-2"
	region := "us-east-1"

	syncAccount(t, seeded, map[string]itam.AssetLister{
		itam.CloudFrontDistributionKind: observations(itam.AssetObservation{
			ExternalID:    distributionID,
			Name:          "cdn",
			Configuration: itam.CloudFrontDistributionConfiguration{DomainName: "cdn.cloudfront.net"},
			SchemaVersion: itam.CloudFrontDistributionSchemaVersion,
		}),
		itam.EC2InstanceKind: observations(itam.AssetObservation{
			ExternalID:    instanceID,
			Name:          "web",
			Region:        &region,
			Configuration: itam.EC2InstanceConfiguration{State: "running"},
			SchemaVersion: itam.EC2InstanceSchemaVersion,
		}),
	})

	distribution := loadDiscovered(t, seeded, itam.CloudFrontDistributionKind, distributionID)
	instance := loadDiscovered(t, seeded, itam.EC2InstanceKind, instanceID)
	seen := time.Now().Add(-3 * time.Hour).UTC().Truncate(time.Microsecond)
	setLastSeen(t, seeded, distribution.ID, seen)
	setLastSeen(t, seeded, instance.ID, seen)
	makeAccountDue(t, seeded)

	syncAccount(t, seeded, map[string]itam.AssetLister{
		itam.CloudFrontDistributionKind: failingLister(fmt.Errorf("denied secret-token")),
		itam.EC2InstanceKind: observations(itam.AssetObservation{
			ExternalID:    instanceID,
			Name:          "web-renamed",
			Region:        &region,
			Configuration: itam.EC2InstanceConfiguration{State: "stopped"},
			SchemaVersion: itam.EC2InstanceSchemaVersion,
		}),
	})

	kept := loadDiscovered(t, seeded, itam.CloudFrontDistributionKind, distributionID)
	moved := loadDiscovered(t, seeded, itam.EC2InstanceKind, instanceID)
	sync := loadLatestSync(t, seeded)

	require.NotNil(t, kept.LastSeenAt)
	assert.True(t, kept.LastSeenAt.Equal(seen))
	assert.Equal(t, "cdn", kept.Name)
	require.NotNil(t, moved.LastSeenAt)
	assert.True(t, moved.LastSeenAt.After(seen))
	assert.Equal(t, "web-renamed", moved.Name)
	assert.Equal(t, coredata.AssetSyncStatusFailed, sync.Status)
	require.NotNil(t, sync.Error)
	assert.Equal(t, itam.AssetSyncFailureMessageForTest(), *sync.Error)
	assert.NotContains(t, *sync.Error, "secret-token")
	assert.Equal(t, []string{
		itam.EC2InstanceKind,
		itam.RDSDBInstanceKind,
	}, sync.CompletedKinds)
	assert.Equal(t, 1, countAssetChangeEvents(t, seeded, distribution.ID))
	assert.Equal(t, 2, countAssetChangeEvents(t, seeded, instance.ID))

	_, err := itam.ClaimAssetSyncForTest(
		itam.NewAssetSyncHandlerWithListersForTest(seeded.pg, seeded.opener, nil),
		t.Context(),
		seeded.account.ID,
	)
	require.ErrorIs(t, err, worker.ErrNoTask)
}

func TestAssetSync_StaleFailureRetriesThenWaits(t *testing.T) {
	t.Parallel()

	seeded := seedAWSAccount(t)
	handler := itam.NewAssetSyncHandlerWithListersForTest(seeded.pg, seeded.opener, nil)

	first, err := itam.ClaimAssetSyncForTest(handler, t.Context(), seeded.account.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, first.AttemptNumber)
	backdateSync(t, seeded, first.ID, time.Now().Add(-20*time.Minute))
	require.NoError(t, itam.RecoverStaleAssetSyncsForTest(handler, t.Context()))

	failed := loadLatestSync(t, seeded)
	assert.Equal(t, coredata.AssetSyncStatusFailed, failed.Status)
	assert.Equal(t, 0, countOrganizationChangeEvents(t, seeded))
	require.NotNil(t, failed.Error)
	assert.Equal(t, itam.AssetSyncStaleMessageForTest(), *failed.Error)

	second, err := itam.ClaimAssetSyncForTest(handler, t.Context(), seeded.account.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, second.AttemptNumber)
	backdateSync(t, seeded, second.ID, time.Now().Add(-20*time.Minute))
	require.NoError(t, itam.RecoverStaleAssetSyncsForTest(handler, t.Context()))

	third, err := itam.ClaimAssetSyncForTest(handler, t.Context(), seeded.account.ID)
	require.NoError(t, err)
	assert.Equal(t, 3, third.AttemptNumber)
	backdateSync(t, seeded, third.ID, time.Now().Add(-20*time.Minute))
	require.NoError(t, itam.RecoverStaleAssetSyncsForTest(handler, t.Context()))

	_, err = itam.ClaimAssetSyncForTest(handler, t.Context(), seeded.account.ID)
	require.ErrorIs(t, err, worker.ErrNoTask)

	backdateSync(t, seeded, third.ID, time.Now().Add(-25*time.Hour))

	retried, err := itam.ClaimAssetSyncForTest(handler, t.Context(), seeded.account.ID)
	require.NoError(t, err)
	assert.Equal(t, seeded.account.ID, retried.ConnectorAccountID)
}

func TestAssetSync_SkipsOrganizationWhenInventoryDisabled(t *testing.T) {
	t.Parallel()

	seeded := seedAWSAccount(t)
	handler := itam.NewAssetSyncHandlerWithListersForTest(seeded.pg, seeded.opener, nil)

	setAssetInventoryEnabled(t, seeded, false)

	_, err := itam.ClaimAssetSyncForTest(handler, t.Context(), seeded.account.ID)
	require.ErrorIs(t, err, worker.ErrNoTask)

	setAssetInventoryEnabled(t, seeded, true)

	claimed, err := itam.ClaimAssetSyncForTest(handler, t.Context(), seeded.account.ID)
	require.NoError(t, err)
	assert.Equal(t, seeded.account.ID, claimed.ConnectorAccountID)
}

func TestAssetSync_ExistingListOmitsDiscoveredRows(t *testing.T) {
	t.Parallel()

	seeded := seedAWSAccount(t)
	now := time.Now().UTC()
	manual := &coredata.Asset{
		ID:             gid.New(seeded.account.ID.TenantID(), coredata.AssetEntityType),
		Name:           "Laptop",
		Amount:         1,
		OrganizationID: seeded.account.OrganizationID,
		Type:           coredata.AssetTypePhysical,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	require.NoError(t, seeded.pg.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		return manual.Insert(ctx, tx, seeded.scope)
	}))

	externalID := "arn:aws:cloudfront::222222222222:distribution/EHIDE"
	region := "us-east-1"
	syncAccount(
		t,
		seeded,
		map[string]itam.AssetLister{
			itam.CloudFrontDistributionKind: observations(
				itam.AssetObservation{
					ExternalID: externalID,
					Name:       "cdn",
					Region:     &region,
					Configuration: itam.CloudFrontDistributionConfiguration{
						DomainName:         "d.cloudfront.net",
						Enabled:            true,
						Status:             "Deployed",
						PriceClass:         "PriceClass_100",
						DefaultCertificate: true,
					},
					SchemaVersion: itam.CloudFrontDistributionSchemaVersion,
				},
			),
		},
	)

	visible := listOrganizationAssets(t, seeded)
	assert.Equal(t, []string{"Laptop"}, assetNames(visible))
	assert.Equal(t, 1, countOrganizationAssets(t, seeded))

	setAssetInventoryEnabled(t, seeded, false)

	hidden := listOrganizationAssets(t, seeded)
	assert.Equal(t, []string{"Laptop"}, assetNames(hidden))
	assert.Equal(t, 1, countOrganizationAssets(t, seeded))

	kept := loadDiscovered(t, seeded, itam.CloudFrontDistributionKind, externalID)
	assert.Equal(t, "cdn", kept.Name)
}

type seededAWS struct {
	pg        *pg.Client
	scope     coredata.Scoper
	connector *coredata.Connector
	account   *coredata.ConnectorAccount
	opener    *connector.Service
}

func seedAWSAccount(t *testing.T) seededAWS {
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
		Name:           "AWS",
		Connection:     &connectorprotocol.WorkloadIdentityConnection{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	account := &coredata.ConnectorAccount{
		ID:                gid.New(tenantID, coredata.ConnectorAccountEntityType),
		OrganizationID:    organizationID,
		ConnectorID:       cnnctr.ID,
		ExternalAccountID: "222222222222",
		Name:              "222222222222",
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	require.NoError(t, pgClient.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		org := &coredata.Organization{
			ID:                    organizationID,
			TenantID:              tenantID,
			Name:                  "Asset sync test",
			AssetInventoryEnabled: true,
			CreatedAt:             now,
			UpdatedAt:             now,
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

	opener := connector.NewService(
		pgClient,
		key,
		connectorprotocol.NewConnectorRegistry(),
		awsRegistry(t),
		log.NewLogger(log.WithName("asset-sync-test")),
	)

	return seededAWS{
		pg:        pgClient,
		scope:     scope,
		connector: cnnctr,
		account:   account,
		opener:    opener,
	}
}

func addAccount(t *testing.T, seeded seededAWS, externalID string) seededAWS {
	t.Helper()

	now := time.Now().UTC()
	account := &coredata.ConnectorAccount{
		ID:                gid.New(seeded.account.ID.TenantID(), coredata.ConnectorAccountEntityType),
		OrganizationID:    seeded.account.OrganizationID,
		ConnectorID:       seeded.connector.ID,
		ExternalAccountID: externalID,
		Name:              externalID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	require.NoError(t, seeded.pg.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		_, err := account.Upsert(ctx, tx, seeded.scope)

		return err
	}))

	copy := seeded
	copy.account = account

	return copy
}

func awsRegistry(t *testing.T) *provider.Registry {
	t.Helper()

	registries := provider.NewRegistry()
	err := registries.Register(&provider.Registration{
		Provider:    coredata.ConnectorProviderAWS,
		DisplayName: "Amazon Web Services",
		InitialAccountFunc: func(*coredata.Connector) (string, string, error) {
			return "111111111111", "111111111111", nil
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

	return registries
}

func observations(items ...itam.AssetObservation) itam.AssetLister {
	return func(context.Context, *aws.Session) ([]itam.AssetObservation, error) {
		return items, nil
	}
}

func failingLister(err error) itam.AssetLister {
	return func(context.Context, *aws.Session) ([]itam.AssetObservation, error) {
		return nil, err
	}
}

func syncAccount(t *testing.T, seeded seededAWS, listers map[string]itam.AssetLister) {
	t.Helper()

	handler := itam.NewAssetSyncHandlerWithListersForTest(seeded.pg, seeded.opener, listers)
	_, err := itam.SyncAccountForTest(handler, t.Context(), seeded.account.ID)
	require.NoError(t, err)
}

func TestAssetSync_UnchangedListWritesChangeEvent(t *testing.T) {
	t.Parallel()

	seeded := seedAWSAccount(t)
	externalID := "arn:aws:cloudfront::222222222222:distribution/E9"
	region := "us-east-1"
	observation := itam.AssetObservation{
		ExternalID: externalID,
		Name:       "cdn",
		Region:     &region,
		Configuration: itam.CloudFrontDistributionConfiguration{
			DomainName: "d.cloudfront.net",
		},
		SchemaVersion: itam.CloudFrontDistributionSchemaVersion,
	}
	listers := map[string]itam.AssetLister{
		itam.CloudFrontDistributionKind: observations(observation),
	}

	syncAccount(t, seeded, listers)
	asset := loadDiscovered(t, seeded, itam.CloudFrontDistributionKind, externalID)
	makeAccountDue(t, seeded)
	syncAccount(t, seeded, listers)

	events := loadAssetChangeEvents(t, seeded, asset.ID)
	require.Len(t, events, 2)
	assert.Equal(t, coredata.AssetChangeTypeCreated, events[0].ChangeType)
	assert.Equal(t, coredata.AssetChangeTypeUpdated, events[1].ChangeType)
	assert.Contains(t, string(events[1].After), "last_seen_at")
	assert.NotContains(t, string(events[1].After), `"name"`)
}

func countAssetChangeEvents(t *testing.T, seeded seededAWS, assetID gid.GID) int {
	t.Helper()

	var count int

	require.NoError(t, seeded.pg.WithConn(t.Context(), func(ctx context.Context, conn pg.Querier) error {
		events := coredata.AssetChangeEvents{}
		var err error
		count, err = events.CountByAssetID(ctx, conn, seeded.scope, assetID)

		return err
	}))

	return count
}

func countOrganizationChangeEvents(t *testing.T, seeded seededAWS) int {
	t.Helper()

	var count int

	require.NoError(t, seeded.pg.WithConn(t.Context(), func(ctx context.Context, conn pg.Querier) error {
		return conn.QueryRow(
			ctx,
			`SELECT COUNT(*) FROM asset_change_events WHERE organization_id = $1`,
			seeded.account.OrganizationID,
		).Scan(&count)
	}))

	return count
}

func loadAssetChangeEvents(
	t *testing.T,
	seeded seededAWS,
	assetID gid.GID,
) coredata.AssetChangeEvents {
	t.Helper()

	var events coredata.AssetChangeEvents

	require.NoError(t, seeded.pg.WithConn(t.Context(), func(ctx context.Context, conn pg.Querier) error {
		return events.LoadByAssetID(
			ctx,
			conn,
			seeded.scope,
			assetID,
			page.NewCursor(
				50,
				nil,
				page.Head,
				page.OrderBy[coredata.AssetChangeEventOrderField]{
					Field:     coredata.AssetChangeEventOrderFieldObservedAt,
					Direction: page.OrderDirectionDesc,
				},
			),
		)
	}))

	slices.Reverse(events)

	return events
}

func loadDiscovered(
	t *testing.T,
	seeded seededAWS,
	kind string,
	externalID string,
) *coredata.Asset {
	t.Helper()

	asset := &coredata.Asset{}

	require.NoError(t, seeded.pg.WithConn(t.Context(), func(ctx context.Context, conn pg.Querier) error {
		return asset.LoadByDiscoveredIdentity(
			ctx,
			conn,
			seeded.scope,
			seeded.account.ConnectorID,
			kind,
			externalID,
		)
	}))

	return asset
}

func setAssetInventoryEnabled(t *testing.T, seeded seededAWS, enabled bool) {
	t.Helper()

	require.NoError(t, seeded.pg.WithConn(t.Context(), func(ctx context.Context, conn pg.Querier) error {
		_, err := conn.Exec(
			ctx,
			`UPDATE organizations
			 SET asset_inventory_enabled = $1
			 WHERE id = $2`,
			enabled,
			seeded.account.OrganizationID,
		)

		return err
	}))
}

func listOrganizationAssets(t *testing.T, seeded seededAWS) coredata.Assets {
	t.Helper()

	var assets coredata.Assets

	cursor := page.NewCursor(
		50,
		nil,
		page.Head,
		page.OrderBy[coredata.AssetOrderField]{
			Field:     coredata.AssetOrderFieldCreatedAt,
			Direction: page.OrderDirectionDesc,
		},
	)

	require.NoError(t, seeded.pg.WithConn(t.Context(), func(ctx context.Context, conn pg.Querier) error {
		return assets.LoadByOrganizationID(
			ctx,
			conn,
			seeded.scope,
			seeded.account.OrganizationID,
			cursor,
		)
	}))

	return assets
}

func countOrganizationAssets(t *testing.T, seeded seededAWS) int {
	t.Helper()

	var count int

	require.NoError(t, seeded.pg.WithConn(t.Context(), func(ctx context.Context, conn pg.Querier) error {
		var err error

		assets := coredata.Assets{}
		count, err = assets.CountByOrganizationID(
			ctx,
			conn,
			seeded.scope,
			seeded.account.OrganizationID,
		)

		return err
	}))

	return count
}

func assetNames(assets coredata.Assets) []string {
	names := make([]string, len(assets))
	for i, asset := range assets {
		names[i] = asset.Name
	}

	return names
}

func loadLatestSync(t *testing.T, seeded seededAWS) *coredata.AssetSync {
	t.Helper()

	sync := &coredata.AssetSync{}

	require.NoError(t, seeded.pg.WithConn(t.Context(), func(ctx context.Context, conn pg.Querier) error {
		return sync.LoadLatestByConnectorAccountID(ctx, conn, seeded.scope, seeded.account.ID)
	}))

	return sync
}

func insertOwner(t *testing.T, seeded seededAWS) gid.GID {
	t.Helper()

	now := time.Now().UTC()
	profileID := gid.New(seeded.account.ID.TenantID(), coredata.MembershipProfileEntityType)
	emailAddress, err := mail.ParseAddr(fmt.Sprintf("%s@example.com", profileID))
	require.NoError(t, err)

	require.NoError(t, seeded.pg.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		identity := coredata.Identity{
			ID:                   gid.New(gid.NilTenant, coredata.IdentityEntityType),
			EmailAddress:         emailAddress,
			FullName:             "Owner",
			EmailAddressVerified: true,
			CreatedAt:            now,
			UpdatedAt:            now,
		}
		if err := identity.Insert(ctx, tx); err != nil {
			return err
		}

		membership := coredata.Membership{
			ID:             gid.New(seeded.account.ID.TenantID(), coredata.MembershipEntityType),
			IdentityID:     identity.ID,
			OrganizationID: seeded.account.OrganizationID,
			Role:           coredata.MembershipRoleOwner,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := membership.Insert(ctx, tx, seeded.scope); err != nil {
			return err
		}

		profile := coredata.MembershipProfile{
			ID:             profileID,
			IdentityID:     identity.ID,
			OrganizationID: seeded.account.OrganizationID,
			Source:         coredata.ProfileSourceManual,
			State:          coredata.ProfileStateActive,
			FullName:       "Owner",
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		return profile.Insert(ctx, tx)
	}))

	return profileID
}

func setManualFields(
	t *testing.T,
	seeded seededAWS,
	assetID gid.GID,
	ownerID gid.GID,
	amount int,
	dataTypes string,
	status coredata.AssetStatus,
	seen time.Time,
) {
	t.Helper()

	require.NoError(t, seeded.pg.WithConn(t.Context(), func(ctx context.Context, conn pg.Querier) error {
		_, err := conn.Exec(
			ctx,
			`
UPDATE assets
SET
    amount = $2,
    owner_profile_id = $3,
    data_types_stored = $4,
    status = $5,
    last_seen_at = $6
WHERE id = $1
`,
			assetID,
			amount,
			ownerID,
			dataTypes,
			status,
			seen,
		)

		return err
	}))
}

func setLastSeen(t *testing.T, seeded seededAWS, assetID gid.GID, seen time.Time) {
	t.Helper()

	require.NoError(t, seeded.pg.WithConn(t.Context(), func(ctx context.Context, conn pg.Querier) error {
		_, err := conn.Exec(
			ctx,
			`UPDATE assets SET last_seen_at = $2 WHERE id = $1`,
			assetID,
			seen,
		)

		return err
	}))
}

func makeAccountDue(t *testing.T, seeded seededAWS) {
	t.Helper()

	sync := loadLatestSync(t, seeded)
	backdateSync(t, seeded, sync.ID, time.Now().Add(-25*time.Hour))
}

func backdateSync(t *testing.T, seeded seededAWS, id gid.GID, at time.Time) {
	t.Helper()

	require.NoError(t, itam.BackdateAssetSyncForTest(t.Context(), seeded.pg, id, at))
}
