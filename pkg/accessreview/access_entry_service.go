// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
	AccessEntryService struct {
		pg    *pg.Client
		scope coredata.Scoper
	}

	RecordAccessEntryDecisionRequest struct {
		EntryID      gid.GID
		Decision     coredata.AccessEntryDecision
		DecisionNote *string
		DecidedByID  *gid.GID
	}

	FlagAccessEntryRequest struct {
		EntryID     gid.GID
		Flags       []coredata.AccessEntryFlag
		FlagReasons []string
	}
)

func (s AccessEntryService) Get(
	ctx context.Context,
	entryID gid.GID,
) (*coredata.AccessEntry, error) {
	entry := &coredata.AccessEntry{}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return entry.LoadByID(ctx, conn, s.scope, entryID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get access entry: %w", err)
	}

	return entry, nil
}

func (s AccessEntryService) RecordDecision(
	ctx context.Context,
	req RecordAccessEntryDecisionRequest,
) (*coredata.AccessEntry, error) {
	if req.Decision == coredata.AccessEntryDecisionPending {
		return nil, fmt.Errorf("cannot decide access entry: invalid decision %q", req.Decision)
	}

	if req.Decision != coredata.AccessEntryDecisionApproved {
		if req.DecisionNote == nil || strings.TrimSpace(*req.DecisionNote) == "" {
			return nil, fmt.Errorf("cannot decide access entry: note is required for non-approved decisions")
		}
	}

	entry := &coredata.AccessEntry{}

	err := s.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := entry.LoadByID(ctx, conn, s.scope, req.EntryID); err != nil {
				return fmt.Errorf("cannot load access entry: %w", err)
			}

			campaign := &coredata.AccessReviewCampaign{}
			if err := campaign.LoadByID(ctx, conn, s.scope, entry.AccessReviewCampaignID); err != nil {
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
				entry.Flags = []coredata.AccessEntryFlag{}
			}
			if entry.FlagReasons == nil {
				entry.FlagReasons = []string{}
			}
			if req.Decision == coredata.AccessEntryDecisionRevoke || req.Decision == coredata.AccessEntryDecisionEscalate {
				if len(entry.Flags) == 0 {
					entry.Flags = []coredata.AccessEntryFlag{coredata.AccessEntryFlagExcessive}
				}
			}

			if err := entry.Update(ctx, conn, s.scope); err != nil {
				return fmt.Errorf("cannot record access entry decision: %w", err)
			}

			history := &coredata.AccessEntryDecisionHistory{
				ID:             gid.New(s.scope.GetTenantID(), coredata.AccessEntryDecisionHistoryEntityType),
				OrganizationID: entry.OrganizationID,
				AccessEntry:    entry.ID,
				Decision:       entry.Decision,
				DecisionNote:   entry.DecisionNote,
				DecidedBy:      entry.DecidedBy,
				DecidedAt:      *entry.DecidedAt,
				CreatedAt:      now,
			}
			if err := history.Insert(ctx, conn, s.scope); err != nil {
				return fmt.Errorf("cannot insert decision history: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot record access entry decision: %w", err)
	}

	updatedEntry, err := s.Get(ctx, req.EntryID)
	if err != nil {
		return nil, fmt.Errorf("cannot reload access entry after decision: %w", err)
	}

	return updatedEntry, nil
}

func (s AccessEntryService) RecordDecisions(
	ctx context.Context,
	decisions []RecordAccessEntryDecisionRequest,
) ([]*coredata.AccessEntry, error) {
	for _, d := range decisions {
		if d.Decision == coredata.AccessEntryDecisionPending {
			return nil, fmt.Errorf("cannot bulk decide access entries: invalid decision %q", d.Decision)
		}
		if d.Decision != coredata.AccessEntryDecisionApproved {
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
		func(conn pg.Conn) error {
			// Track verified campaigns to avoid repeated loads within the
			// same transaction.
			verifiedCampaigns := make(map[gid.GID]bool)

			for _, d := range decisions {
				entry := &coredata.AccessEntry{}
				if err := entry.LoadByID(ctx, conn, s.scope, d.EntryID); err != nil {
					return fmt.Errorf("cannot load access entry %s: %w", d.EntryID, err)
				}

				if !verifiedCampaigns[entry.AccessReviewCampaignID] {
					campaign := &coredata.AccessReviewCampaign{}
					if err := campaign.LoadByID(ctx, conn, s.scope, entry.AccessReviewCampaignID); err != nil {
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
					entry.Flags = []coredata.AccessEntryFlag{}
				}
				if entry.FlagReasons == nil {
					entry.FlagReasons = []string{}
				}
				if d.Decision == coredata.AccessEntryDecisionRevoke || d.Decision == coredata.AccessEntryDecisionEscalate {
					if len(entry.Flags) == 0 {
						entry.Flags = []coredata.AccessEntryFlag{coredata.AccessEntryFlagExcessive}
					}
				}

				if err := entry.Update(ctx, conn, s.scope); err != nil {
					return fmt.Errorf("cannot record decision for entry %s: %w", d.EntryID, err)
				}

				history := &coredata.AccessEntryDecisionHistory{
					ID:             gid.New(s.scope.GetTenantID(), coredata.AccessEntryDecisionHistoryEntityType),
					OrganizationID: entry.OrganizationID,
					AccessEntry:    entry.ID,
					Decision:       entry.Decision,
					DecisionNote:   entry.DecisionNote,
					DecidedBy:      entry.DecidedBy,
					DecidedAt:      *entry.DecidedAt,
					CreatedAt:      now,
				}
				if err := history.Insert(ctx, conn, s.scope); err != nil {
					return fmt.Errorf("cannot insert decision history for entry %s: %w", d.EntryID, err)
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot record access entry decisions: %w", err)
	}

	entries := make([]*coredata.AccessEntry, len(entryIDs))
	for i, id := range entryIDs {
		entry, err := s.Get(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("cannot reload access entry %s: %w", id, err)
		}
		entries[i] = entry
	}

	return entries, nil
}

func (s AccessEntryService) FlagEntry(
	ctx context.Context,
	req FlagAccessEntryRequest,
) (*coredata.AccessEntry, error) {
	entry := &coredata.AccessEntry{}

	err := s.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := entry.LoadByID(ctx, conn, s.scope, req.EntryID); err != nil {
				return fmt.Errorf("cannot load access entry: %w", err)
			}

			campaign := &coredata.AccessReviewCampaign{}
			if err := campaign.LoadByID(ctx, conn, s.scope, entry.AccessReviewCampaignID); err != nil {
				return fmt.Errorf("cannot load campaign: %w", err)
			}

			if campaign.Status != coredata.AccessReviewCampaignStatusPendingActions {
				return fmt.Errorf("cannot flag access entry: campaign status is %s, expected PENDING_ACTIONS", campaign.Status)
			}

			now := time.Now()
			entry.Flags = req.Flags
			if entry.Flags == nil {
				entry.Flags = []coredata.AccessEntryFlag{}
			}
			entry.FlagReasons = req.FlagReasons
			if entry.FlagReasons == nil {
				entry.FlagReasons = []string{}
			}
			entry.UpdatedAt = now

			return entry.UpdateFlags(ctx, conn, s.scope)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot flag access entry: %w", err)
	}

	return s.Get(ctx, req.EntryID)
}

func (s AccessEntryService) ListForCampaignID(
	ctx context.Context,
	campaignID gid.GID,
	cursor *page.Cursor[coredata.AccessEntryOrderField],
	filter *coredata.AccessEntryFilter,
) (*page.Page[*coredata.AccessEntry, coredata.AccessEntryOrderField], error) {
	var entries coredata.AccessEntries

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return entries.LoadByCampaignID(ctx, conn, s.scope, campaignID, cursor, filter)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list access entries: %w", err)
	}

	return page.NewPage(entries, cursor), nil
}

func (s AccessEntryService) ListForCampaignIDAndSourceID(
	ctx context.Context,
	campaignID gid.GID,
	sourceID gid.GID,
	cursor *page.Cursor[coredata.AccessEntryOrderField],
	filter *coredata.AccessEntryFilter,
) (*page.Page[*coredata.AccessEntry, coredata.AccessEntryOrderField], error) {
	var entries coredata.AccessEntries

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return entries.LoadByCampaignIDAndSourceID(ctx, conn, s.scope, campaignID, sourceID, cursor, filter)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list access entries: %w", err)
	}

	return page.NewPage(entries, cursor), nil
}

func (s AccessEntryService) CountForCampaignID(
	ctx context.Context,
	campaignID gid.GID,
	filter *coredata.AccessEntryFilter,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			entries := coredata.AccessEntries{}
			count, err = entries.CountByCampaignID(ctx, conn, s.scope, campaignID, filter)
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

func (s AccessEntryService) CountForCampaignIDAndSourceID(
	ctx context.Context,
	campaignID gid.GID,
	sourceID gid.GID,
	filter *coredata.AccessEntryFilter,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			entries := coredata.AccessEntries{}
			count, err = entries.CountByCampaignIDAndSourceID(ctx, conn, s.scope, campaignID, sourceID, filter)
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

func (s AccessEntryService) CountPendingForCampaignID(
	ctx context.Context,
	campaignID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			entries := coredata.AccessEntries{}
			count, err = entries.CountPendingByCampaignID(ctx, conn, s.scope, campaignID)
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

func (s AccessEntryService) DecisionHistory(
	ctx context.Context,
	entryID gid.GID,
) (coredata.AccessEntryDecisionHistories, error) {
	var histories coredata.AccessEntryDecisionHistories

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return histories.LoadByEntryID(ctx, conn, s.scope, entryID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load decision history: %w", err)
	}

	return histories, nil
}

func (s AccessEntryService) Statistics(
	ctx context.Context,
	campaignID gid.GID,
) (*coredata.AccessEntryStatistics, error) {
	stats := &coredata.AccessEntryStatistics{}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return stats.LoadByCampaignID(ctx, conn, s.scope, campaignID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load campaign statistics: %w", err)
	}

	return stats, nil
}

func (s AccessEntryService) StatisticsForSource(
	ctx context.Context,
	campaignID gid.GID,
	sourceID gid.GID,
) (*coredata.AccessEntryStatistics, error) {
	stats := &coredata.AccessEntryStatistics{}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return stats.LoadByCampaignIDAndSourceID(ctx, conn, s.scope, campaignID, sourceID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load source statistics: %w", err)
	}

	return stats, nil
}
