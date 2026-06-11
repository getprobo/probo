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
			return entry.LoadByID(ctx, conn, scope, entryID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get access entry: %w", err)
	}

	return entry, nil
}

func (s *Service) RecordDecision(
	ctx context.Context,
	scope coredata.Scoper,
	req RecordAccessReviewEntryDecisionRequest,
) (*coredata.AccessReviewEntry, error) {
	if req.Decision == coredata.AccessReviewEntryDecisionPending {
		return nil, fmt.Errorf("cannot decide access entry: invalid decision %q", req.Decision)
	}

	if req.Decision != coredata.AccessReviewEntryDecisionApproved {
		if req.DecisionNote == nil || strings.TrimSpace(*req.DecisionNote) == "" {
			return nil, fmt.Errorf("cannot decide access entry: note is required for non-approved decisions")
		}
	}

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
				return fmt.Errorf("cannot decide access entry: campaign status is %s, expected PENDING_ACTIONS", campaign.Status)
			}

			now := time.Now()
			entry.Decision = req.Decision
			entry.DecisionNote = req.DecisionNote
			entry.DecidedBy = req.DecidedByID
			entry.DecidedAt = &now

			entry.UpdatedAt = now
			if entry.Flags == nil {
				entry.Flags = []coredata.AccessReviewEntryFlag{}
			}

			if entry.FlagReasons == nil {
				entry.FlagReasons = []string{}
			}

			if req.Decision == coredata.AccessReviewEntryDecisionRevoke || req.Decision == coredata.AccessReviewEntryDecisionEscalate {
				if len(entry.Flags) == 0 {
					entry.Flags = []coredata.AccessReviewEntryFlag{coredata.AccessReviewEntryFlagExcessive}
				}
			}

			if err := entry.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot record access entry decision: %w", err)
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
				return fmt.Errorf("cannot insert decision history: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot record access entry decision: %w", err)
	}

	updatedEntry, err := s.GetEntry(ctx, scope, req.EntryID)
	if err != nil {
		return nil, fmt.Errorf("cannot reload access entry after decision: %w", err)
	}

	return updatedEntry, nil
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

				now := time.Now()
				entry.Decision = d.Decision
				entry.DecisionNote = d.DecisionNote
				entry.DecidedBy = d.DecidedByID
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
				return fmt.Errorf("cannot flag access entry: campaign status is %s, expected PENDING_ACTIONS", campaign.Status)
			}

			now := time.Now()

			entry.Flags = req.Flags
			if entry.Flags == nil {
				entry.Flags = []coredata.AccessReviewEntryFlag{}
			}

			entry.FlagReasons = req.FlagReasons
			if entry.FlagReasons == nil {
				entry.FlagReasons = []string{}
			}

			entry.UpdatedAt = now

			return entry.UpdateFlags(ctx, conn, scope)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot flag access entry: %w", err)
	}

	return s.GetEntry(ctx, scope, req.EntryID)
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
			return entries.LoadByCampaignID(ctx, conn, scope, campaignID, cursor, filter)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list access entries: %w", err)
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
			return entries.LoadByCampaignIDAndSourceID(ctx, conn, scope, campaignID, sourceID, cursor, filter)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list access entries: %w", err)
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
		return 0, fmt.Errorf("cannot count access entries: %w", err)
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
		return 0, fmt.Errorf("cannot count access entries: %w", err)
	}

	return count, nil
}

func (s *Service) CountPendingEntriesForCampaignID(
	ctx context.Context,
	scope coredata.Scoper,
	campaignID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			entries := coredata.AccessReviewEntries{}

			count, err = entries.CountPendingByCampaignID(ctx, conn, scope, campaignID)
			if err != nil {
				return fmt.Errorf("cannot count pending access entries: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot count pending access entries: %w", err)
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
			return stats.LoadByCampaignID(ctx, conn, scope, campaignID)
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
			return stats.LoadByCampaignIDAndSourceID(ctx, conn, scope, campaignID, sourceID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load source statistics: %w", err)
	}

	return stats, nil
}
