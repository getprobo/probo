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
	"fmt"
	"strings"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	RecordAccessReviewEntryDecisionRequest struct {
		EntryID      gid.GID
		Decision     coredata.AccessReviewEntryDecision
		DecisionNote *string
		DecidedByID  *gid.GID
	}

	FlagAccessReviewEntryRequest struct {
		EntryID     gid.GID
		Flags       []coredata.AccessReviewEntryFlag
		FlagReasons []string
	}
)

func (s *Service) GetEntry(
	ctx context.Context,
	scope coredata.Scoper,
	entryID gid.GID,
) (*coredata.AccessReviewEntry, error) {
	entry := &coredata.AccessReviewEntry{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := entry.LoadByID(ctx, conn, scope, entryID); err != nil {
				return fmt.Errorf("cannot load access entry: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

func (s *Service) RecordDecision(
	ctx context.Context,
	scope coredata.Scoper,
	req RecordAccessReviewEntryDecisionRequest,
) (*coredata.AccessReviewEntry, error) {
	entries, err := s.RecordDecisions(ctx, scope, []RecordAccessReviewEntryDecisionRequest{req})
	if err != nil {
		return nil, err
	}

	return entries[0], nil
}

func (s *Service) RecordDecisions(
	ctx context.Context,
	scope coredata.Scoper,
	decisions []RecordAccessReviewEntryDecisionRequest,
) ([]*coredata.AccessReviewEntry, error) {
	for _, d := range decisions {
		if d.Decision == coredata.AccessReviewEntryDecisionPending {
			return nil, fmt.Errorf("cannot bulk decide access entries: invalid decision %q", d.Decision)
		}

		if d.Decision != coredata.AccessReviewEntryDecisionApproved {
			if d.DecisionNote == nil || strings.TrimSpace(*d.DecisionNote) == "" {
				return nil, fmt.Errorf(
					"cannot bulk decide access entries: note is required for non-approved decisions on entry %s",
					d.EntryID,
				)
			}
		}
	}

	entryIDs := make([]gid.GID, len(decisions))
	for i, d := range decisions {
		entryIDs[i] = d.EntryID
	}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			// Track verified campaigns to avoid repeated loads within the
			// same transaction.
			verifiedCampaigns := make(map[gid.GID]bool)
			decidedByCache := make(map[gid.GID]*gid.GID)

			for _, d := range decisions {
				entry := &coredata.AccessReviewEntry{}
				if err := entry.LoadByID(ctx, conn, scope, d.EntryID); err != nil {
					return fmt.Errorf("cannot load access entry %s: %w", d.EntryID, err)
				}

				if !verifiedCampaigns[entry.AccessReviewCampaignID] {
					campaign := &coredata.AccessReviewCampaign{}
					if err := campaign.LoadByID(ctx, conn, scope, entry.AccessReviewCampaignID); err != nil {
						return fmt.Errorf("cannot load campaign: %w", err)
					}

					if campaign.Status != coredata.AccessReviewCampaignStatusPendingActions {
						return fmt.Errorf("cannot decide access entry: campaign status is %s, expected PENDING_ACTIONS", campaign.Status)
					}

					verifiedCampaigns[entry.AccessReviewCampaignID] = true
				}

				decidedByID, ok := decidedByCache[entry.OrganizationID]
				if !ok {
					decidedByID = nil

					if d.DecidedByID != nil {
						profile := &coredata.MembershipProfile{}
						if err := profile.LoadByIdentityIDAndOrganizationID(ctx, conn, scope, *d.DecidedByID, entry.OrganizationID); err == nil {
							decidedByID = &profile.ID
						}
					}

					decidedByCache[entry.OrganizationID] = decidedByID
				}

				now := time.Now()
				entry.Decision = d.Decision
				entry.DecisionNote = d.DecisionNote
				entry.DecidedBy = decidedByID
				entry.DecidedAt = &now

				entry.UpdatedAt = now
				if entry.Flags == nil {
					entry.Flags = []coredata.AccessReviewEntryFlag{}
				}

				if entry.FlagReasons == nil {
					entry.FlagReasons = []string{}
				}

				if d.Decision == coredata.AccessReviewEntryDecisionRevoke || d.Decision == coredata.AccessReviewEntryDecisionEscalate {
					if len(entry.Flags) == 0 {
						entry.Flags = []coredata.AccessReviewEntryFlag{coredata.AccessReviewEntryFlagExcessive}
					}
				}

				if err := entry.Update(ctx, conn, scope); err != nil {
					return fmt.Errorf("cannot record decision for entry %s: %w", d.EntryID, err)
				}

				history := &coredata.AccessReviewEntryDecisionHistory{
					ID:                gid.New(scope.GetTenantID(), coredata.AccessReviewEntryDecisionHistoryEntityType),
					OrganizationID:    entry.OrganizationID,
					AccessReviewEntry: entry.ID,
					Decision:          entry.Decision,
					DecisionNote:      entry.DecisionNote,
					DecidedBy:         entry.DecidedBy,
					DecidedAt:         *entry.DecidedAt,
					CreatedAt:         now,
				}
				if err := history.Insert(ctx, conn, scope); err != nil {
					return fmt.Errorf("cannot insert decision history for entry %s: %w", d.EntryID, err)
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot record access entry decisions: %w", err)
	}

	entries := make([]*coredata.AccessReviewEntry, len(entryIDs))
	for i, id := range entryIDs {
		entry, err := s.GetEntry(ctx, scope, id)
		if err != nil {
			return nil, fmt.Errorf("cannot reload access entry %s: %w", id, err)
		}

		entries[i] = entry
	}

	return entries, nil
}

func (s *Service) FlagEntry(
	ctx context.Context,
	scope coredata.Scoper,
	req FlagAccessReviewEntryRequest,
) (*coredata.AccessReviewEntry, error) {
	entry := &coredata.AccessReviewEntry{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := entry.LoadByID(ctx, conn, scope, req.EntryID); err != nil {
				return fmt.Errorf("cannot load access entry: %w", err)
			}

			campaign := &coredata.AccessReviewCampaign{}
			if err := campaign.LoadByID(ctx, conn, scope, entry.AccessReviewCampaignID); err != nil {
				return fmt.Errorf("cannot load campaign: %w", err)
			}

			if campaign.Status != coredata.AccessReviewCampaignStatusPendingActions {
				return fmt.Errorf("cannot flag access entry: campaign status is %q, expected PENDING_ACTIONS", campaign.Status)
			}

			entry.Flags = req.Flags
			if entry.Flags == nil {
				entry.Flags = []coredata.AccessReviewEntryFlag{}
			}

			entry.FlagReasons = req.FlagReasons
			if entry.FlagReasons == nil {
				entry.FlagReasons = []string{}
			}

			entry.UpdatedAt = time.Now()

			if err := entry.UpdateFlags(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update access entry flags: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

func (s *Service) ListEntriesForCampaignID(
	ctx context.Context,
	scope coredata.Scoper,
	campaignID gid.GID,
	cursor *page.Cursor[coredata.AccessReviewEntryOrderField],
	filter *coredata.AccessReviewEntryFilter,
) (*page.Page[*coredata.AccessReviewEntry, coredata.AccessReviewEntryOrderField], error) {
	var entries coredata.AccessReviewEntries

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := entries.LoadByCampaignID(
				ctx,
				conn,
				scope,
				campaignID,
				cursor,
				filter,
			); err != nil {
				return fmt.Errorf("cannot load access entries: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(entries, cursor), nil
}

func (s *Service) ListEntriesForCampaignIDAndSourceID(
	ctx context.Context,
	scope coredata.Scoper,
	campaignID gid.GID,
	sourceID gid.GID,
	cursor *page.Cursor[coredata.AccessReviewEntryOrderField],
	filter *coredata.AccessReviewEntryFilter,
) (*page.Page[*coredata.AccessReviewEntry, coredata.AccessReviewEntryOrderField], error) {
	var entries coredata.AccessReviewEntries

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := entries.LoadByCampaignIDAndSourceID(
				ctx,
				conn,
				scope,
				campaignID,
				sourceID,
				cursor,
				filter,
			); err != nil {
				return fmt.Errorf("cannot load access entries: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(entries, cursor), nil
}

func (s *Service) CountEntriesForCampaignID(
	ctx context.Context,
	scope coredata.Scoper,
	campaignID gid.GID,
	filter *coredata.AccessReviewEntryFilter,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			entries := coredata.AccessReviewEntries{}

			count, err = entries.CountByCampaignID(ctx, conn, scope, campaignID, filter)
			if err != nil {
				return fmt.Errorf("cannot count access entries by campaign: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) CountEntriesForCampaignIDAndSourceID(
	ctx context.Context,
	scope coredata.Scoper,
	campaignID gid.GID,
	sourceID gid.GID,
	filter *coredata.AccessReviewEntryFilter,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			entries := coredata.AccessReviewEntries{}

			count, err = entries.CountByCampaignIDAndSourceID(ctx, conn, scope, campaignID, sourceID, filter)
			if err != nil {
				return fmt.Errorf("cannot count access entries by campaign and source: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) EntryDecisionHistory(
	ctx context.Context,
	scope coredata.Scoper,
	entryID gid.GID,
) (coredata.AccessReviewEntryDecisionHistories, error) {
	var histories coredata.AccessReviewEntryDecisionHistories

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return histories.LoadByEntryID(ctx, conn, scope, entryID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load decision history: %w", err)
	}

	return histories, nil
}

func (s *Service) CampaignStatistics(
	ctx context.Context,
	scope coredata.Scoper,
	campaignID gid.GID,
) (*coredata.AccessReviewStatistics, error) {
	stats := &coredata.AccessReviewStatistics{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := stats.LoadByCampaignID(ctx, conn, scope, campaignID); err != nil {
				return fmt.Errorf("cannot load campaign statistics: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load campaign statistics: %w", err)
	}

	return stats, nil
}

func (s *Service) CampaignSourceStatistics(
	ctx context.Context,
	scope coredata.Scoper,
	campaignID gid.GID,
	sourceID gid.GID,
) (*coredata.AccessReviewStatistics, error) {
	stats := &coredata.AccessReviewStatistics{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := stats.LoadByCampaignIDAndSourceID(ctx, conn, scope, campaignID, sourceID); err != nil {
				return fmt.Errorf("cannot load source statistics: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
