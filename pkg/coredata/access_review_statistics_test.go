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

package coredata_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestAccessReviewStatistics_CountsByCampaignAndSource(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedAccessReviewEntryFixture(t, ctx, client)

	tenantID := fx.scope.GetTenantID()
	otherCampaignSourceID := gid.New(tenantID, coredata.AccessReviewCampaignSourceEntityType)
	now := time.Now().UTC()

	entries := []struct {
		campaignSourceID gid.GID
		decision         coredata.AccessReviewEntryDecision
		flags            []coredata.AccessReviewEntryFlag
	}{
		{
			campaignSourceID: fx.campaignSourceID,
			decision:         coredata.AccessReviewEntryDecisionPending,
			flags:            []coredata.AccessReviewEntryFlag{coredata.AccessReviewEntryFlagInactive},
		},
		{
			campaignSourceID: fx.campaignSourceID,
			decision:         coredata.AccessReviewEntryDecisionApproved,
			flags:            []coredata.AccessReviewEntryFlag{},
		},
		{
			campaignSourceID: otherCampaignSourceID,
			decision:         coredata.AccessReviewEntryDecisionPending,
			flags:            []coredata.AccessReviewEntryFlag{coredata.AccessReviewEntryFlagInactive, coredata.AccessReviewEntryFlagExcessive},
		},
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		otherCampaignSource := &coredata.AccessReviewCampaignSource{
			ID:                     otherCampaignSourceID,
			OrganizationID:         fx.organizationID,
			TenantID:               tenantID,
			AccessReviewCampaignID: fx.campaignID,
			Name:                   "Statistics Test Other Source",
			CreatedAt:              now,
			UpdatedAt:              now,
		}
		if err := otherCampaignSource.Upsert(ctx, tx, fx.scope); err != nil {
			return err
		}

		for i, e := range entries {
			email := fmt.Sprintf("user%d@example.com", i)

			entry := &coredata.AccessReviewEntry{
				ID:                           gid.New(tenantID, coredata.AccessReviewEntryEntityType),
				OrganizationID:               fx.organizationID,
				AccessReviewCampaignID:       fx.campaignID,
				AccessReviewCampaignSourceID: e.campaignSourceID,
				Email:                        email,
				MFAStatus:                    coredata.MFAStatusUnknown,
				AuthMethod:                   coredata.AccessReviewEntryAuthMethodUnknown,
				AccountType:                  coredata.AccessReviewEntryAccountTypeUser,
				AccountKey:                   email,
				Flags:                        e.flags,
				FlagReasons:                  []string{},
				Decision:                     e.decision,
				CreatedAt:                    now,
				UpdatedAt:                    now,
			}
			if err := entry.Upsert(ctx, tx, fx.scope); err != nil {
				return err
			}
		}

		return nil
	}))

	var campaign, source coredata.AccessReviewStatistics

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		if err := campaign.LoadByCampaignID(ctx, conn, fx.scope, fx.campaignID); err != nil {
			return err
		}

		return source.LoadByCampaignIDAndSourceID(ctx, conn, fx.scope, fx.campaignID, fx.campaignSourceID)
	}))

	assert.Equal(t, 3, campaign.TotalCount)
	assert.Equal(t, 2, campaign.DecisionCounts[coredata.AccessReviewEntryDecisionPending])
	assert.Equal(t, 1, campaign.DecisionCounts[coredata.AccessReviewEntryDecisionApproved])
	assert.Equal(t, 2, campaign.FlagCounts[coredata.AccessReviewEntryFlagInactive])
	assert.Equal(t, 1, campaign.FlagCounts[coredata.AccessReviewEntryFlagExcessive])

	assert.Equal(t, 2, source.TotalCount)
	assert.Equal(t, 1, source.DecisionCounts[coredata.AccessReviewEntryDecisionPending])
	assert.Equal(t, 1, source.DecisionCounts[coredata.AccessReviewEntryDecisionApproved])
	assert.Equal(t, 1, source.FlagCounts[coredata.AccessReviewEntryFlagInactive])
	assert.Zero(t, source.FlagCounts[coredata.AccessReviewEntryFlagExcessive])
}
