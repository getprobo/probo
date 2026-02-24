// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

	"github.com/jackc/pgx/v5"
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
)

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
	if req.Decision == coredata.AccessEntryDecisionPending || req.Decision == coredata.AccessEntryDecisionModify {
		return nil, fmt.Errorf("invalid decision %q", req.Decision)
	}

	if req.Decision != coredata.AccessEntryDecisionApproved {
		if req.DecisionNote == nil || strings.TrimSpace(*req.DecisionNote) == "" {
			return nil, fmt.Errorf("decision note is required for non-approved decisions")
		}
	}

	entry := &coredata.AccessEntry{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := entry.LoadByID(ctx, conn, s.svc.scope, req.EntryID); err != nil {
				return fmt.Errorf("cannot load access entry: %w", err)
			}

			if entry.Decision != coredata.AccessEntryDecisionPending {
				return fmt.Errorf("invalid decision transition from %s", entry.Decision)
			}

			now := time.Now()
			previousDecision := entry.Decision
			entry.Decision = req.Decision
			entry.DecisionNote = req.DecisionNote
			entry.DecidedBy = req.DecidedByID
			entry.DecidedAt = &now
			entry.UpdatedAt = now
			if req.Decision == coredata.AccessEntryDecisionRevoke || req.Decision == coredata.AccessEntryDecisionEscalate {
				entry.Flag = coredata.AccessEntryFlagExcessive
			}

			q := `
UPDATE access_entries
SET
    flag = @flag,
    flag_reason = @flag_reason,
    decision = @decision,
    decision_note = @decision_note,
    decided_by = @decided_by,
    decided_at = @decided_at,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id
    AND decision = 'PENDING'
`
			q = fmt.Sprintf(q, s.svc.scope.SQLFragment())
			args := pgx.StrictNamedArgs{
				"id":            entry.ID,
				"flag":          entry.Flag,
				"flag_reason":   entry.FlagReason,
				"decision":      entry.Decision,
				"decision_note": entry.DecisionNote,
				"decided_by":    entry.DecidedBy,
				"decided_at":    entry.DecidedAt,
				"updated_at":    entry.UpdatedAt,
			}
			for k, v := range s.svc.scope.SQLArguments() {
				args[k] = v
			}

			result, err := conn.Exec(ctx, q, args)
			if err != nil {
				return fmt.Errorf("cannot update access entry: %w", err)
			}
			if result.RowsAffected() == 0 {
				return fmt.Errorf("invalid decision transition from pending")
			}

			decisionEventID := gid.New(s.svc.scope.GetTenantID(), coredata.AccessEntryEntityType)
			_, err = conn.Exec(ctx, `
INSERT INTO access_entry_decision_events (
    id,
    tenant_id,
    access_entry_id,
    access_review_campaign_id,
    previous_decision,
    new_decision,
    decision_note,
    decided_by,
    decided_at,
    created_at
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
`,
				decisionEventID,
				s.svc.scope.GetTenantID(),
				entry.ID,
				entry.AccessReviewCampaignID,
				previousDecision,
				entry.Decision,
				entry.DecisionNote,
				entry.DecidedBy,
				entry.DecidedAt,
				now,
			)
			if err != nil {
				return fmt.Errorf("cannot insert decision event: %w", err)
			}

			if req.Decision == coredata.AccessEntryDecisionRevoke || req.Decision == coredata.AccessEntryDecisionEscalate {
				note := ""
				if entry.DecisionNote != nil {
					note = *entry.DecisionNote
				}
				taskID := gid.New(s.svc.scope.GetTenantID(), coredata.AccessEntryEntityType)
				_, err = conn.Exec(ctx, `
INSERT INTO access_review_remediation_tasks (
    id,
    tenant_id,
    access_review_campaign_id,
    access_entry_id,
    decided_by,
    status,
    status_note,
    created_at,
    updated_at
)
VALUES ($1,$2,$3,$4,$5,'OPEN',$6,$7,$8)
ON CONFLICT (access_entry_id) DO NOTHING
`,
					taskID,
					s.svc.scope.GetTenantID(),
					entry.AccessReviewCampaignID,
					entry.ID,
					entry.DecidedBy,
					note,
					now,
					now,
				)
				if err != nil {
					return fmt.Errorf("cannot create remediation task: %w", err)
				}

				campaignArgs := pgx.StrictNamedArgs{
					"updated_at":   now,
					"campaign_id": entry.AccessReviewCampaignID,
				}
				for k, v := range s.svc.scope.SQLArguments() {
					campaignArgs[k] = v
				}
				if _, err := conn.Exec(ctx, fmt.Sprintf(`
UPDATE access_review_campaigns
SET status = 'PENDING_ACTIONS', updated_at = @updated_at
WHERE %s
  AND id = @campaign_id
  AND status = 'IN_PROGRESS'
`, s.svc.scope.SQLFragment()), campaignArgs); err != nil {
					return fmt.Errorf("cannot transition campaign to pending actions: %w", err)
				}
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

func (s AccessEntryService) ListForCampaignID(
	ctx context.Context,
	campaignID gid.GID,
	cursor *page.Cursor[coredata.AccessEntryOrderField],
) (*page.Page[*coredata.AccessEntry, coredata.AccessEntryOrderField], error) {
	var entries coredata.AccessEntries

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return entries.LoadByCampaignID(ctx, conn, s.svc.scope, campaignID, cursor)
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
) (*page.Page[*coredata.AccessEntry, coredata.AccessEntryOrderField], error) {
	var entries coredata.AccessEntries

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return entries.LoadByCampaignIDAndSourceID(ctx, conn, s.svc.scope, campaignID, sourceID, cursor)
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
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			entries := coredata.AccessEntries{}
			count, err = entries.CountByCampaignID(ctx, conn, s.svc.scope, campaignID)
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
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			entries := coredata.AccessEntries{}
			count, err = entries.CountByCampaignIDAndSourceID(ctx, conn, s.svc.scope, campaignID, sourceID)
			return err
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}
