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

package accessreview

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/connectorprotocol/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
)

func TestBackfillInventoryFromLatestCampaignEntries(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)

	fx := seedFetchFixture(t, ctx, client, "email,full_name,role\nunused@example.com,Unused,Admin\n")
	now := time.Now().UTC()

	insertCampaignEntry(
		t,
		ctx,
		client,
		fx,
		"alice@example.com",
		"Alice",
		[]string{"Admin", "Billing"},
		now,
	)
	insertCampaignEntry(
		t,
		ctx,
		client,
		fx,
		"bob@example.com",
		"Bob",
		nil,
		now,
	)

	completedAt := now
	attempt := &coredata.AccessReviewCampaignSourceFetchAttempt{
		ID:                           gid.New(fx.scope.GetTenantID(), coredata.AccessReviewCampaignSourceFetchAttemptEntityType),
		OrganizationID:               fx.organizationID,
		TenantID:                     fx.scope.GetTenantID(),
		AccessReviewCampaignSourceID: fx.campaignSourceID,
		Status:                       coredata.AccessReviewCampaignSourceFetchStatusSuccess,
		FetchedAccountsCount:         2,
		StartedAt:                    &now,
		CompletedAt:                  &completedAt,
		CreatedAt:                    now,
		UpdatedAt:                    now,
	}
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return attempt.Insert(ctx, tx, fx.scope)
	}))

	count, err := svc.BackfillInventoryFromLatestCampaignEntries(
		ctx,
		fx.scope,
		fx.organizationID,
	)
	require.NoError(t, err)
	assert.Equal(t, 3, count)

	inventoryFX := inventorySyncFixture{
		scope:          fx.scope,
		organizationID: fx.organizationID,
		sourceID:       fx.sourceID,
	}
	grants := loadInventoryGrants(t, ctx, client, inventoryFX)
	require.Len(t, grants, 3)

	byKey := make(map[string]*coredata.AccessInventoryGrant, len(grants))
	for _, grant := range grants {
		byKey[grant.PrincipalKey+"|"+grant.EntitlementKey] = grant
		assert.Equal(t, DefaultResourceKey, grant.ResourceKey)
		assert.Equal(t, coredata.AccessInventoryGrantPresencePresent, grant.Presence)
		assert.Equal(t, 0, grant.ConsecutiveMissingSyncs)
	}

	require.Contains(t, byKey, "alice@example.com|Admin")
	require.Contains(t, byKey, "alice@example.com|Billing")
	require.Contains(t, byKey, "bob@example.com|member")
	assert.Equal(t, ImplicitMemberEntitlementDisplayName, byKey["bob@example.com|member"].EntitlementDisplayName)

	assert.Empty(t, loadInventoryChangeEvents(t, ctx, client, inventoryFX))

	again, err := svc.BackfillInventoryFromLatestCampaignEntries(
		ctx,
		fx.scope,
		fx.organizationID,
	)
	require.NoError(t, err)
	assert.Equal(t, 0, again)
	assert.Len(t, loadInventoryGrants(t, ctx, client, inventoryFX), 3)
	assert.Empty(t, loadInventoryChangeEvents(t, ctx, client, inventoryFX))
}

func TestBackfillInventoryFromLatestCampaignEntries_SkipsWhenSyncSucceeded(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)

	fx := seedFetchFixture(t, ctx, client, "email,full_name,role\nunused@example.com,Unused,Admin\n")
	now := time.Now().UTC()

	insertCampaignEntry(
		t,
		ctx,
		client,
		fx,
		"carol@example.com",
		"Carol",
		[]string{"Viewer"},
		now,
	)

	completedAt := now
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		attempt := &coredata.AccessReviewCampaignSourceFetchAttempt{
			ID:                           gid.New(fx.scope.GetTenantID(), coredata.AccessReviewCampaignSourceFetchAttemptEntityType),
			OrganizationID:               fx.organizationID,
			TenantID:                     fx.scope.GetTenantID(),
			AccessReviewCampaignSourceID: fx.campaignSourceID,
			Status:                       coredata.AccessReviewCampaignSourceFetchStatusSuccess,
			FetchedAccountsCount:         1,
			StartedAt:                    &now,
			CompletedAt:                  &completedAt,
			CreatedAt:                    now,
			UpdatedAt:                    now,
		}
		if err := attempt.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		run := &coredata.AccessSourceSyncRun{
			ID:                   gid.New(fx.scope.GetTenantID(), coredata.AccessSourceSyncRunEntityType),
			OrganizationID:       fx.organizationID,
			AccessReviewSourceID: fx.sourceID,
			Status:               coredata.AccessSourceSyncRunStatusSuccess,
			GrantsUpsertedCount:  1,
			StartedAt:            &now,
			CompletedAt:          &completedAt,
			CreatedAt:            now,
			UpdatedAt:            now,
		}

		return run.Insert(ctx, tx, fx.scope)
	}))

	count, err := svc.BackfillInventoryFromLatestCampaignEntries(
		ctx,
		fx.scope,
		fx.organizationID,
	)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	inventoryFX := inventorySyncFixture{
		scope:          fx.scope,
		organizationID: fx.organizationID,
		sourceID:       fx.sourceID,
	}
	assert.Empty(t, loadInventoryGrants(t, ctx, client, inventoryFX))
	assert.Empty(t, loadInventoryChangeEvents(t, ctx, client, inventoryFX))
}

func TestBackfillInventoryFromLatestCampaignEntries_UsesLatestSuccessfulSnapshot(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)

	fx := seedFetchFixture(t, ctx, client, "email,full_name,role\nunused@example.com,Unused,Admin\n")
	older := time.Now().UTC().Add(-time.Hour)
	newer := older.Add(30 * time.Minute)

	insertCampaignEntry(t, ctx, client, fx, "old@example.com", "Old", []string{"Legacy"}, older)

	completedOlder := older
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		attempt := &coredata.AccessReviewCampaignSourceFetchAttempt{
			ID:                           gid.New(fx.scope.GetTenantID(), coredata.AccessReviewCampaignSourceFetchAttemptEntityType),
			OrganizationID:               fx.organizationID,
			TenantID:                     fx.scope.GetTenantID(),
			AccessReviewCampaignSourceID: fx.campaignSourceID,
			Status:                       coredata.AccessReviewCampaignSourceFetchStatusSuccess,
			FetchedAccountsCount:         1,
			StartedAt:                    &older,
			CompletedAt:                  &completedOlder,
			CreatedAt:                    older,
			UpdatedAt:                    older,
		}

		return attempt.Insert(ctx, tx, fx.scope)
	}))

	newerCampaignID := gid.New(fx.scope.GetTenantID(), coredata.AccessReviewCampaignEntityType)
	newerCampaignSourceID := gid.New(fx.scope.GetTenantID(), coredata.AccessReviewCampaignSourceEntityType)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		campaign := &coredata.AccessReviewCampaign{
			ID:             newerCampaignID,
			OrganizationID: fx.organizationID,
			Name:           "Newer campaign",
			Status:         coredata.AccessReviewCampaignStatusInProgress,
			StartedAt:      &newer,
			CreatedAt:      newer,
			UpdatedAt:      newer,
		}
		if err := campaign.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		campaignSource := &coredata.AccessReviewCampaignSource{
			ID:                     newerCampaignSourceID,
			OrganizationID:         fx.organizationID,
			AccessReviewCampaignID: newerCampaignID,
			AccessReviewSourceID:   &fx.sourceID,
			Name:                   "CSV",
			CreatedAt:              newer,
			UpdatedAt:              newer,
		}
		if err := campaignSource.Upsert(ctx, tx, fx.scope); err != nil {
			return err
		}

		entry := &coredata.AccessReviewEntry{
			ID:                           gid.New(fx.scope.GetTenantID(), coredata.AccessReviewEntryEntityType),
			OrganizationID:               fx.organizationID,
			AccessReviewCampaignID:       newerCampaignID,
			AccessReviewCampaignSourceID: newerCampaignSourceID,
			Email:                        "new@example.com",
			FullName:                     "New",
			Roles:                        []string{"Current"},
			JobTitle:                     "",
			MFAStatus:                    coredata.MFAStatusUnknown,
			AuthMethod:                   coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType:                  coredata.AccessReviewEntryAccountTypeUser,
			ExternalID:                   "",
			AccountKey:                   "new@example.com",
			Flags:                        []coredata.AccessReviewEntryFlag{},
			FlagReasons:                  []string{},
			Decision:                     coredata.AccessReviewEntryDecisionPending,
			CreatedAt:                    newer,
			UpdatedAt:                    newer,
		}
		if err := entry.Upsert(ctx, tx, fx.scope); err != nil {
			return err
		}

		completedNewer := newer
		attempt := &coredata.AccessReviewCampaignSourceFetchAttempt{
			ID:                           gid.New(fx.scope.GetTenantID(), coredata.AccessReviewCampaignSourceFetchAttemptEntityType),
			OrganizationID:               fx.organizationID,
			TenantID:                     fx.scope.GetTenantID(),
			AccessReviewCampaignSourceID: newerCampaignSourceID,
			Status:                       coredata.AccessReviewCampaignSourceFetchStatusSuccess,
			FetchedAccountsCount:         1,
			StartedAt:                    &newer,
			CompletedAt:                  &completedNewer,
			CreatedAt:                    newer,
			UpdatedAt:                    newer,
		}

		return attempt.Insert(ctx, tx, fx.scope)
	}))

	count, err := svc.BackfillInventoryFromLatestCampaignEntries(
		ctx,
		fx.scope,
		fx.organizationID,
	)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	inventoryFX := inventorySyncFixture{
		scope:          fx.scope,
		organizationID: fx.organizationID,
		sourceID:       fx.sourceID,
	}
	grants := loadInventoryGrants(t, ctx, client, inventoryFX)
	require.Len(t, grants, 1)
	assert.Equal(t, "new@example.com", grants[0].PrincipalKey)
	assert.Equal(t, "Current", grants[0].EntitlementKey)
	assert.Empty(t, loadInventoryChangeEvents(t, ctx, client, inventoryFX))
}

func insertCampaignEntry(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx fetchFixture,
	email string,
	fullName string,
	roles []string,
	at time.Time,
) {
	t.Helper()

	entry := &coredata.AccessReviewEntry{
		ID:                           gid.New(fx.scope.GetTenantID(), coredata.AccessReviewEntryEntityType),
		OrganizationID:               fx.organizationID,
		AccessReviewCampaignID:       fx.campaignID,
		AccessReviewCampaignSourceID: fx.campaignSourceID,
		Email:                        email,
		FullName:                     fullName,
		Roles:                        roles,
		JobTitle:                     "",
		MFAStatus:                    coredata.MFAStatusUnknown,
		AuthMethod:                   coredata.AccessReviewEntryAuthMethodUnknown,
		AccountType:                  coredata.AccessReviewEntryAccountTypeUser,
		ExternalID:                   "",
		AccountKey:                   email,
		Flags:                        []coredata.AccessReviewEntryFlag{},
		FlagReasons:                  []string{},
		Decision:                     coredata.AccessReviewEntryDecisionPending,
		CreatedAt:                    at,
		UpdatedAt:                    at,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return entry.Upsert(ctx, tx, fx.scope)
	}))
}
