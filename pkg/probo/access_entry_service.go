// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package probo

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
		svc *TenantService
	}

	RecordAccessEntryDecisionRequest struct {
		EntryID      gid.GID
		Decision     coredata.AccessEntryDecision
		DecisionNote *string
		DecidedByID  *gid.GID
	}

	FlagAccessEntryRequest struct {
		EntryID    gid.GID
		Flag       coredata.AccessEntryFlag
		FlagReason *string
	}
)

func (s AccessEntryService) ResolveOrganizationID(
	ctx context.Context,
	entryID gid.GID,
) (gid.GID, error) {
	var organizationID gid.GID

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			var err error
			entry := &coredata.AccessEntry{}
			organizationID, err = entry.LoadOrganizationID(ctx, conn, entryID)
			return err
		},
	)
	if err != nil {
		return gid.GID{}, fmt.Errorf("cannot resolve organization id: %w", err)
	}

	return organizationID, nil
}

func (s AccessEntryService) Get(
	ctx context.Context,
	entryID gid.GID,
) (*coredata.AccessEntry, error) {
	entry := &coredata.AccessEntry{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return entry.LoadByID(ctx, conn, s.svc.scope, entryID)
		},
	)
	if err != nil {
		return nil, err
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

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := entry.LoadByID(ctx, conn, s.svc.scope, req.EntryID); err != nil {
				return fmt.Errorf("cannot load access entry: %w", err)
			}

			campaign := &coredata.AccessReviewCampaign{}
			if err := campaign.LoadByID(ctx, conn, s.svc.scope, entry.AccessReviewCampaignID); err != nil {
				return fmt.Errorf("cannot load campaign: %w", err)
			}

			if campaign.Status != coredata.AccessReviewCampaignStatusPendingActions {
				return fmt.Errorf("cannot decide access entry: campaign status is %s, expected PENDING_ACTIONS", campaign.Status)
			}

			if entry.Decision != coredata.AccessEntryDecisionPending {
				return fmt.Errorf("cannot decide access entry: invalid transition from %s", entry.Decision)
			}

			now := time.Now()
			entry.Decision = req.Decision
			entry.DecisionNote = req.DecisionNote
			entry.DecidedBy = req.DecidedByID
			entry.DecidedAt = &now
			entry.UpdatedAt = now
			if req.Decision == coredata.AccessEntryDecisionRevoke || req.Decision == coredata.AccessEntryDecisionEscalate {
				if entry.Flag == coredata.AccessEntryFlagNone {
					entry.Flag = coredata.AccessEntryFlagExcessive
				}
			}

			if err := entry.RecordDecision(ctx, conn, s.svc.scope); err != nil {
				return err
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	updatedEntry, err := s.Get(ctx, req.EntryID)
	if err != nil {
		return nil, err
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

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			// Track verified campaigns to avoid repeated loads within the
			// same transaction.
			verifiedCampaigns := make(map[gid.GID]bool)

			for _, d := range decisions {
				entry := &coredata.AccessEntry{}
				if err := entry.LoadByID(ctx, conn, s.svc.scope, d.EntryID); err != nil {
					return fmt.Errorf("cannot load access entry %s: %w", d.EntryID, err)
				}

				if !verifiedCampaigns[entry.AccessReviewCampaignID] {
					campaign := &coredata.AccessReviewCampaign{}
					if err := campaign.LoadByID(ctx, conn, s.svc.scope, entry.AccessReviewCampaignID); err != nil {
						return fmt.Errorf("cannot load campaign: %w", err)
					}
					if campaign.Status != coredata.AccessReviewCampaignStatusPendingActions {
						return fmt.Errorf("cannot decide access entry: campaign status is %s, expected PENDING_ACTIONS", campaign.Status)
					}
					verifiedCampaigns[entry.AccessReviewCampaignID] = true
				}

				if entry.Decision != coredata.AccessEntryDecisionPending {
					return fmt.Errorf(
						"cannot bulk decide access entries: invalid transition from %s for entry %s",
						entry.Decision,
						d.EntryID,
					)
				}

				now := time.Now()
				entry.Decision = d.Decision
				entry.DecisionNote = d.DecisionNote
				entry.DecidedBy = d.DecidedByID
				entry.DecidedAt = &now
				entry.UpdatedAt = now
				if d.Decision == coredata.AccessEntryDecisionRevoke || d.Decision == coredata.AccessEntryDecisionEscalate {
					if entry.Flag == coredata.AccessEntryFlagNone {
						entry.Flag = coredata.AccessEntryFlagExcessive
					}
				}

				if err := entry.RecordDecision(ctx, conn, s.svc.scope); err != nil {
					return fmt.Errorf("cannot record decision for entry %s: %w", d.EntryID, err)
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
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

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := entry.LoadByID(ctx, conn, s.svc.scope, req.EntryID); err != nil {
				return fmt.Errorf("cannot load access entry: %w", err)
			}

			campaign := &coredata.AccessReviewCampaign{}
			if err := campaign.LoadByID(ctx, conn, s.svc.scope, entry.AccessReviewCampaignID); err != nil {
				return fmt.Errorf("cannot load campaign: %w", err)
			}

			if campaign.Status != coredata.AccessReviewCampaignStatusPendingActions {
				return fmt.Errorf("cannot flag access entry: campaign status is %s, expected PENDING_ACTIONS", campaign.Status)
			}

			now := time.Now()
			entry.Flag = req.Flag
			entry.FlagReason = req.FlagReason
			entry.UpdatedAt = now

			return entry.UpdateFlag(ctx, conn, s.svc.scope)
		},
	)
	if err != nil {
		return nil, err
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

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return entries.LoadByCampaignID(ctx, conn, s.svc.scope, campaignID, cursor, filter)
		},
	)
	if err != nil {
		return nil, err
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

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return entries.LoadByCampaignIDAndSourceID(ctx, conn, s.svc.scope, campaignID, sourceID, cursor, filter)
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(entries, cursor), nil
}

func (s AccessEntryService) CountForCampaignID(
	ctx context.Context,
	campaignID gid.GID,
	filter *coredata.AccessEntryFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			entries := coredata.AccessEntries{}
			count, err = entries.CountByCampaignID(ctx, conn, s.svc.scope, campaignID, filter)
			return err
		},
	)
	if err != nil {
		return 0, err
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

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			entries := coredata.AccessEntries{}
			count, err = entries.CountByCampaignIDAndSourceID(ctx, conn, s.svc.scope, campaignID, sourceID, filter)
			return err
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s AccessEntryService) Statistics(
	ctx context.Context,
	campaignID gid.GID,
) (*coredata.AccessEntryStatistics, error) {
	stats := &coredata.AccessEntryStatistics{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return stats.LoadByCampaignID(ctx, conn, s.svc.scope, campaignID)
		},
	)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (s AccessEntryService) StatisticsForSource(
	ctx context.Context,
	campaignID gid.GID,
	sourceID gid.GID,
) (*coredata.AccessEntryStatistics, error) {
	stats := &coredata.AccessEntryStatistics{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return stats.LoadByCampaignIDAndSourceID(ctx, conn, s.svc.scope, campaignID, sourceID)
		},
	)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
