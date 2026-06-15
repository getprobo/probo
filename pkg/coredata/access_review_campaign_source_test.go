// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package coredata_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func insertAccessReviewEntry(t *testing.T, ctx context.Context, client *pg.Client, fx accessEntryFixture, accountKey string) gid.GID {
	t.Helper()

	tenantID := fx.scope.GetTenantID()
	now := time.Now().UTC().Truncate(time.Microsecond)
	entryID := gid.New(tenantID, coredata.AccessReviewEntryEntityType)

	entry := &coredata.AccessReviewEntry{
		ID:                           entryID,
		OrganizationID:               fx.organizationID,
		AccessReviewCampaignID:       fx.campaignID,
		AccessReviewCampaignSourceID: fx.campaignSourceID,
		Email:                        accountKey,
		FullName:                     "Snapshot User",
		Role:                         "member",
		MFAStatus:                    coredata.MFAStatusUnknown,
		AuthMethod:                   coredata.AccessReviewEntryAuthMethodUnknown,
		AccountType:                  coredata.AccessReviewEntryAccountTypeUser,
		ExternalID:                   "ext-snapshot",
		AccountKey:                   accountKey,
		IncrementalTag:               coredata.AccessReviewEntryIncrementalTagNew,
		Flags:                        []coredata.AccessReviewEntryFlag{},
		FlagReasons:                  []string{},
		Decision:                     coredata.AccessReviewEntryDecisionPending,
		CreatedAt:                    now,
		UpdatedAt:                    now,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return entry.Upsert(ctx, tx, fx.scope)
	}))

	return entryID
}

// TestAccessReviewSourceDeletion_PreservesSnapshotAndEntries verifies the core
// archival guarantee: deleting the live access source nulls the snapshot link
// (ON DELETE SET NULL) but keeps the per-campaign snapshot and its entries.
func TestAccessReviewSourceDeletion_PreservesSnapshotAndEntries(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedAccessReviewEntryFixture(t, ctx, client)

	entryID := insertAccessReviewEntry(t, ctx, client, fx, "preserve-me@example.com")

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		_, err := tx.Exec(ctx, `DELETE FROM access_review_sources WHERE id = $1`, fx.sourceID)
		return err
	}))

	loadedEntry := &coredata.AccessReviewEntry{}
	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return loadedEntry.LoadByID(ctx, conn, fx.scope, entryID)
	}))
	assert.Equal(t, "preserve-me@example.com", loadedEntry.Email, "entry must survive source deletion")

	loadedSource := &coredata.AccessReviewCampaignSource{}
	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return loadedSource.LoadByID(ctx, conn, fx.scope, fx.campaignSourceID)
	}))
	assert.Nil(t, loadedSource.AccessReviewSourceID, "snapshot link must be nulled, not cascaded")
	assert.Equal(t, "Upsert Freeze Test Source", loadedSource.Name, "snapshot name must be preserved")
}

// TestSourceFetchAttempts_AppendOnly verifies attempts accumulate as an
// append-only log and that the latest attempt reflects the most recent run.
func TestSourceFetchAttempts_AppendOnly(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedAccessReviewEntryFixture(t, ctx, client)

	tenantID := fx.scope.GetTenantID()
	now := time.Now().UTC().Truncate(time.Microsecond)
	failureMsg := "We couldn't fetch accounts from this source."

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		first := &coredata.AccessReviewCampaignSourceFetchAttempt{
			ID:                           gid.New(tenantID, coredata.AccessReviewCampaignSourceFetchAttemptEntityType),
			AccessReviewCampaignSourceID: fx.campaignSourceID,
			Status:                       coredata.AccessReviewCampaignSourceFetchStatusFailed,
			Error:                        &failureMsg,
			CompletedAt:                  &now,
			CreatedAt:                    now,
			UpdatedAt:                    now,
		}
		if err := first.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		second := &coredata.AccessReviewCampaignSourceFetchAttempt{
			ID:                           gid.New(tenantID, coredata.AccessReviewCampaignSourceFetchAttemptEntityType),
			AccessReviewCampaignSourceID: fx.campaignSourceID,
			Status:                       coredata.AccessReviewCampaignSourceFetchStatusSuccess,
			FetchedAccountsCount:         7,
			CompletedAt:                  &now,
			CreatedAt:                    now.Add(time.Minute),
			UpdatedAt:                    now.Add(time.Minute),
		}
		if err := second.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		return nil
	}))

	var history coredata.AccessReviewCampaignSourceFetchAttempts
	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return history.LoadAllByCampaignSourceID(ctx, conn, fx.scope, fx.campaignSourceID)
	}))
	require.Len(t, history, 2, "both attempts must be retained")
	assert.Equal(t, coredata.AccessReviewCampaignSourceFetchStatusSuccess, history[0].Status, "history is newest first")
	assert.Equal(t, coredata.AccessReviewCampaignSourceFetchStatusFailed, history[1].Status)
	require.NotNil(t, history[1].Error)
	assert.Equal(t, failureMsg, *history[1].Error, "the failed attempt's error is retained")

	var latest coredata.AccessReviewCampaignSourceFetchAttempts
	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return latest.LoadLatestByCampaignID(ctx, conn, fx.scope, fx.campaignID)
	}))
	require.Len(t, latest, 1, "one latest attempt per snapshot")
	assert.Equal(t, coredata.AccessReviewCampaignSourceFetchStatusSuccess, latest[0].Status)
	assert.Equal(t, 7, latest[0].FetchedAccountsCount)
}
