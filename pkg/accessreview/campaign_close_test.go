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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestCloseCampaign_IgnoresPendingOnFailedSourceFetch(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedCampaignCloseFixture(t, ctx, client)

	tenantID := fx.scope.GetTenantID()
	now := time.Now().UTC().Truncate(time.Microsecond)
	failureMsg := "We couldn't fetch accounts from this source."

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		failedAttempt := &coredata.AccessReviewCampaignSourceFetchAttempt{
			ID:                           gid.New(tenantID, coredata.AccessReviewCampaignSourceFetchAttemptEntityType),
			OrganizationID:               fx.organizationID,
			AccessReviewCampaignSourceID: fx.failedCampaignSourceID,
			Status:                       coredata.AccessReviewCampaignSourceFetchStatusFailed,
			Error:                        &failureMsg,
			CompletedAt:                  &now,
			CreatedAt:                    now,
			UpdatedAt:                    now,
		}
		if err := failedAttempt.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		successAttempt := &coredata.AccessReviewCampaignSourceFetchAttempt{
			ID:                           gid.New(tenantID, coredata.AccessReviewCampaignSourceFetchAttemptEntityType),
			OrganizationID:               fx.organizationID,
			AccessReviewCampaignSourceID: fx.successCampaignSourceID,
			Status:                       coredata.AccessReviewCampaignSourceFetchStatusSuccess,
			FetchedAccountsCount:         1,
			CompletedAt:                  &now,
			CreatedAt:                    now,
			UpdatedAt:                    now,
		}
		if err := successAttempt.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		pendingOnFailed := &coredata.AccessReviewEntry{
			ID:                           gid.New(tenantID, coredata.AccessReviewEntryEntityType),
			OrganizationID:               fx.organizationID,
			AccessReviewCampaignID:       fx.campaignID,
			AccessReviewCampaignSourceID: fx.failedCampaignSourceID,
			Email:                        "orphan@example.com",
			FullName:                     "Orphan",
			Roles:                        []string{"member"},
			MFAStatus:                    coredata.MFAStatusUnknown,
			AuthMethod:                   coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType:                  coredata.AccessReviewEntryAccountTypeUser,
			ExternalID:                   "orphan",
			AccountKey:                   "orphan@example.com",
			IncrementalTag:               coredata.AccessReviewEntryIncrementalTagNew,
			Flags:                        []coredata.AccessReviewEntryFlag{},
			FlagReasons:                  []string{},
			Decision:                     coredata.AccessReviewEntryDecisionPending,
			CreatedAt:                    now,
			UpdatedAt:                    now,
		}
		if err := pendingOnFailed.Upsert(ctx, tx, fx.scope); err != nil {
			return err
		}

		decidedOnSuccess := &coredata.AccessReviewEntry{
			ID:                           gid.New(tenantID, coredata.AccessReviewEntryEntityType),
			OrganizationID:               fx.organizationID,
			AccessReviewCampaignID:       fx.campaignID,
			AccessReviewCampaignSourceID: fx.successCampaignSourceID,
			Email:                        "reviewed@example.com",
			FullName:                     "Reviewed",
			Roles:                        []string{"member"},
			MFAStatus:                    coredata.MFAStatusUnknown,
			AuthMethod:                   coredata.AccessReviewEntryAuthMethodUnknown,
			AccountType:                  coredata.AccessReviewEntryAccountTypeUser,
			ExternalID:                   "reviewed",
			AccountKey:                   "reviewed@example.com",
			IncrementalTag:               coredata.AccessReviewEntryIncrementalTagNew,
			Flags:                        []coredata.AccessReviewEntryFlag{},
			FlagReasons:                  []string{},
			Decision:                     coredata.AccessReviewEntryDecisionApproved,
			CreatedAt:                    now,
			UpdatedAt:                    now,
		}

		return decidedOnSuccess.Upsert(ctx, tx, fx.scope)
	}))

	svc := accessreview.NewService(client, nil, nil, nil, nil)

	campaign, err := svc.CloseCampaign(ctx, fx.scope, fx.campaignID)
	require.NoError(t, err)
	assert.Equal(t, coredata.AccessReviewCampaignStatusCompleted, campaign.Status)
}

type campaignCloseFixture struct {
	scope                   coredata.Scoper
	organizationID          gid.GID
	sourceID                gid.GID
	campaignID              gid.GID
	successCampaignSourceID gid.GID
	failedCampaignSourceID  gid.GID
}

func seedCampaignCloseFixture(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
) campaignCloseFixture {
	t.Helper()

	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	campaignID := gid.New(tenantID, coredata.AccessReviewCampaignEntityType)
	sourceID := gid.New(tenantID, coredata.AccessReviewSourceEntityType)
	successCampaignSourceID := gid.New(tenantID, coredata.AccessReviewCampaignSourceEntityType)
	failedCampaignSourceID := gid.New(tenantID, coredata.AccessReviewCampaignSourceEntityType)
	now := time.Now().UTC()

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		org := &coredata.Organization{
			ID:        organizationID,
			TenantID:  tenantID,
			Name:      "Close Campaign Test Org",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := org.Insert(ctx, tx); err != nil {
			return err
		}

		source := &coredata.AccessReviewSource{
			ID:             sourceID,
			OrganizationID: organizationID,
			Name:           "Close Campaign Test Source",
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := source.Insert(ctx, tx, scope); err != nil {
			return err
		}

		campaign := &coredata.AccessReviewCampaign{
			ID:             campaignID,
			OrganizationID: organizationID,
			Name:           "Close Campaign Test",
			Status:         coredata.AccessReviewCampaignStatusPendingActions,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := campaign.Insert(ctx, tx, scope); err != nil {
			return err
		}

		successSource := &coredata.AccessReviewCampaignSource{
			ID:                     successCampaignSourceID,
			OrganizationID:         organizationID,
			AccessReviewCampaignID: campaignID,
			AccessReviewSourceID:   &sourceID,
			Name:                   "Success Snapshot",
			CreatedAt:              now,
			UpdatedAt:              now,
		}
		if err := successSource.Upsert(ctx, tx, scope); err != nil {
			return err
		}

		failedSource := &coredata.AccessReviewCampaignSource{
			ID:                     failedCampaignSourceID,
			OrganizationID:         organizationID,
			AccessReviewCampaignID: campaignID,
			AccessReviewSourceID:   &sourceID,
			Name:                   "Failed Snapshot",
			CreatedAt:              now,
			UpdatedAt:              now,
		}

		return failedSource.Upsert(ctx, tx, scope)
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			_, err := tx.Exec(ctx, `DELETE FROM access_review_campaign_source_fetch_attempts WHERE access_review_campaign_source_id = ANY($1)`, []gid.GID{successCampaignSourceID, failedCampaignSourceID})
			if err != nil {
				return err
			}

			_, err = tx.Exec(ctx, `DELETE FROM access_review_entries WHERE access_review_campaign_id = $1`, campaignID)
			if err != nil {
				return err
			}

			_, err = tx.Exec(ctx, `DELETE FROM access_review_campaign_sources WHERE access_review_campaign_id = $1`, campaignID)
			if err != nil {
				return err
			}

			_, err = tx.Exec(ctx, `DELETE FROM access_review_campaigns WHERE id = $1`, campaignID)
			if err != nil {
				return err
			}

			_, err = tx.Exec(ctx, `DELETE FROM access_review_sources WHERE id = $1`, sourceID)
			if err != nil {
				return err
			}

			_, err = tx.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, organizationID)

			return err
		})
	})

	return campaignCloseFixture{
		scope:                   scope,
		organizationID:          organizationID,
		sourceID:                sourceID,
		campaignID:              campaignID,
		successCampaignSourceID: successCampaignSourceID,
		failedCampaignSourceID:  failedCampaignSourceID,
	}
}
