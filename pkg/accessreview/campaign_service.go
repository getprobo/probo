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

package accessreview

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.gearno.de/x/ref"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

// CampaignService implements access review campaign operations.
type CampaignService struct {
	pg    *pg.Client
	scope coredata.Scoper
}

// NewCampaignService creates a new campaign service.
func NewCampaignService(pgClient *pg.Client, scope coredata.Scoper) *CampaignService {
	return &CampaignService{
		pg:    pgClient,
		scope: scope,
	}
}

// Ensure CampaignService implements accessreview.AccessReviewCampaignService.
var _ AccessReviewCampaignService = (*CampaignService)(nil)

func (s *CampaignService) Create(
	ctx context.Context,
	req CreateAccessReviewCampaignRequest,
) (*coredata.AccessReviewCampaign, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	campaign := &coredata.AccessReviewCampaign{
		ID:                gid.New(s.scope.GetTenantID(), coredata.AccessReviewCampaignEntityType),
		AccessReviewID:    req.AccessReviewID,
		Name:              req.Name,
		Status:            coredata.AccessReviewCampaignStatusDraft,
		FrameworkControls: req.FrameworkControls,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	err := s.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			review := &coredata.AccessReview{}
			if err := review.LoadByID(ctx, conn, s.scope, req.AccessReviewID); err != nil {
				return fmt.Errorf("cannot load access review: %w", err)
			}

			if err := campaign.Insert(ctx, conn, s.scope); err != nil {
				return fmt.Errorf("cannot insert access review campaign: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create access review campaign: %w", err)
	}

	return campaign, nil
}

func (s *CampaignService) Get(
	ctx context.Context,
	campaignID gid.GID,
) (*coredata.AccessReviewCampaign, error) {
	campaign := &coredata.AccessReviewCampaign{}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return campaign.LoadByID(ctx, conn, s.scope, campaignID)
		},
	)
	if err != nil {
		return nil, err
	}

	return campaign, nil
}

func (s *CampaignService) Update(
	ctx context.Context,
	req UpdateAccessReviewCampaignRequest,
) (*coredata.AccessReviewCampaign, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	campaign := &coredata.AccessReviewCampaign{}

	err := s.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := campaign.LoadByID(ctx, conn, s.scope, req.CampaignID); err != nil {
				return fmt.Errorf("cannot load campaign: %w", err)
			}

			if req.Name != nil {
				campaign.Name = *req.Name
			}

			if req.FrameworkControls != nil {
				campaign.FrameworkControls = *req.FrameworkControls
			}

			campaign.UpdatedAt = time.Now()

			if err := campaign.Update(ctx, conn, s.scope); err != nil {
				return fmt.Errorf("cannot update campaign: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return campaign, nil
}

func (s *CampaignService) Delete(
	ctx context.Context,
	campaignID gid.GID,
) error {
	campaign := &coredata.AccessReviewCampaign{ID: campaignID}

	return s.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			return campaign.Delete(ctx, conn, s.scope)
		},
	)
}

func (s *CampaignService) Start(
	ctx context.Context,
	req StartAccessReviewCampaignRequest,
) (*coredata.AccessReviewCampaign, error) {
	campaign := &coredata.AccessReviewCampaign{}

	err := s.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := s.lockCampaignForUpdate(ctx, conn, req.CampaignID); err != nil {
				return err
			}

			if err := campaign.LoadByID(ctx, conn, s.scope, req.CampaignID); err != nil {
				return fmt.Errorf("cannot load campaign: %w", err)
			}

			if campaign.Status != coredata.AccessReviewCampaignStatusDraft && campaign.Status != coredata.AccessReviewCampaignStatusFailed {
				return fmt.Errorf("campaign must be in DRAFT or FAILED status to start, current status: %s", campaign.Status)
			}

			// Insert scope systems (selected access sources for this campaign).
			// When provided, only the selected sources will be used during snapshot.
			// When empty, all sources for the access review are used (default behavior).
			for _, sourceID := range req.AccessSourceIDs {
				// Verify the access source exists and belongs to this access review
				source := &coredata.AccessSource{}
				if err := source.LoadByID(ctx, conn, s.scope, sourceID); err != nil {
					return fmt.Errorf("cannot load access source %s: %w", sourceID, err)
				}

				if source.AccessReviewID != campaign.AccessReviewID {
					return fmt.Errorf("access source %s does not belong to the same access review as this campaign", sourceID)
				}

				_, err := conn.Exec(ctx, `
INSERT INTO access_review_campaign_scope_systems (access_review_campaign_id, access_source_id)
VALUES ($1, $2)
ON CONFLICT (access_review_campaign_id, access_source_id) DO NOTHING
`, campaign.ID, sourceID)
				if err != nil {
					return fmt.Errorf("cannot insert scope system: %w", err)
				}
			}

			now := time.Now()
			campaign.Status = coredata.AccessReviewCampaignStatusInProgress
			campaign.StartedAt = &now
			campaign.UpdatedAt = now

			if err := campaign.Update(ctx, conn, s.scope); err != nil {
				return fmt.Errorf("cannot update campaign: %w", err)
			}

			if err := s.enqueueSourceFetches(ctx, conn, campaign.ID); err != nil {
				return fmt.Errorf("cannot queue source fetches: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return campaign, nil
}

func (s *CampaignService) RetryStart(
	ctx context.Context,
	campaignID gid.GID,
) (*coredata.AccessReviewCampaign, error) {
	return s.Start(ctx, StartAccessReviewCampaignRequest{
		CampaignID: campaignID,
	})
}

func (s *CampaignService) Close(
	ctx context.Context,
	campaignID gid.GID,
) (*coredata.AccessReviewCampaign, error) {
	campaign := &coredata.AccessReviewCampaign{}

	err := s.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := s.lockCampaignForUpdate(ctx, conn, campaignID); err != nil {
				return err
			}

			if err := campaign.LoadByID(ctx, conn, s.scope, campaignID); err != nil {
				return fmt.Errorf("cannot load campaign: %w", err)
			}

			if campaign.Status != coredata.AccessReviewCampaignStatusInProgress &&
				campaign.Status != coredata.AccessReviewCampaignStatusPendingActions {
				return fmt.Errorf("campaign must be IN_PROGRESS or PENDING_ACTIONS to close, current status: %s", campaign.Status)
			}

			checkpointCount := 0
			args := pgx.StrictNamedArgs{
				"campaign_id": campaignID,
			}
			maps.Copy(args, s.scope.SQLArguments())
			if err := conn.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(id)
FROM access_review_campaign_validation_checkpoints
WHERE %s
  AND access_review_campaign_id = @campaign_id
`, s.scope.SQLFragment()), args).Scan(&checkpointCount); err != nil {
				return fmt.Errorf("cannot count validation checkpoints: %w", err)
			}
			if checkpointCount == 0 {
				return fmt.Errorf("campaign validation checkpoint is required before close")
			}

			// Check all entries have been decided
			entries := coredata.AccessEntries{}
			pendingCount, err := entries.CountPendingByCampaignID(ctx, conn, s.scope, campaignID)
			if err != nil {
				return fmt.Errorf("cannot count pending entries: %w", err)
			}

			if pendingCount > 0 {
				return fmt.Errorf("cannot close campaign: %d entries still pending", pendingCount)
			}

			openTasks := 0
			if err := conn.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(id)
FROM access_review_remediation_tasks
WHERE %s
  AND access_review_campaign_id = @campaign_id
  AND (
    status = 'OPEN'
    OR (status = 'CANCELLED' AND (status_note IS NULL OR btrim(status_note) = ''))
  )
`, s.scope.SQLFragment()), args).Scan(&openTasks); err != nil {
				return fmt.Errorf("cannot count blocking remediation tasks: %w", err)
			}
			if openTasks > 0 {
				return fmt.Errorf("cannot close campaign: %d remediation tasks are still open or missing cancellation note", openTasks)
			}

			campaign.Status = coredata.AccessReviewCampaignStatusCompleted
			campaign.CompletedAt = ref.Ref(time.Now())
			campaign.UpdatedAt = time.Now()

			if err := campaign.Update(ctx, conn, s.scope); err != nil {
				return fmt.Errorf("cannot update campaign: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return campaign, nil
}

func (s *CampaignService) ValidateForClose(
	ctx context.Context,
	campaignID gid.GID,
	validatedBy *gid.GID,
	note *string,
) error {
	return s.pg.WithTx(ctx, func(conn pg.Conn) error {
		if err := s.lockCampaignForUpdate(ctx, conn, campaignID); err != nil {
			return err
		}

		campaign := &coredata.AccessReviewCampaign{}
		if err := campaign.LoadByID(ctx, conn, s.scope, campaignID); err != nil {
			return fmt.Errorf("cannot load campaign: %w", err)
		}
		if campaign.Status != coredata.AccessReviewCampaignStatusInProgress &&
			campaign.Status != coredata.AccessReviewCampaignStatusPendingActions {
			return fmt.Errorf("campaign must be IN_PROGRESS or PENDING_ACTIONS to validate, current status: %s", campaign.Status)
		}

		checkpointID := gid.New(s.scope.GetTenantID(), coredata.AccessReviewCampaignEntityType)
		now := time.Now()
		if _, err := conn.Exec(ctx, `
INSERT INTO access_review_campaign_validation_checkpoints (
    id,
    tenant_id,
    access_review_campaign_id,
    validated_by,
    note,
    validated_at,
    created_at
)
VALUES ($1,$2,$3,$4,$5,$6,$7)
ON CONFLICT (access_review_campaign_id) DO UPDATE SET
    validated_by = EXCLUDED.validated_by,
    note = EXCLUDED.note,
    validated_at = EXCLUDED.validated_at
`,
			checkpointID,
			s.scope.GetTenantID(),
			campaignID,
			validatedBy,
			note,
			now,
			now,
		); err != nil {
			return fmt.Errorf("cannot create validation checkpoint: %w", err)
		}

		return nil
	})
}

func (s *CampaignService) ExportEvidence(
	ctx context.Context,
	campaignID gid.GID,
) (string, string, error) {
	type decisionSummaryRow struct {
		Decision coredata.AccessEntryDecision `json:"decision"`
		Count    int                          `json:"count"`
	}

	returnPayload := ""
	returnChecksum := ""
	err := s.pg.WithTx(ctx, func(conn pg.Conn) error {
		if err := s.lockCampaignForUpdate(ctx, conn, campaignID); err != nil {
			return err
		}

		campaign := &coredata.AccessReviewCampaign{}
		if err := campaign.LoadByID(ctx, conn, s.scope, campaignID); err != nil {
			return fmt.Errorf("cannot load campaign: %w", err)
		}

		args := pgx.StrictNamedArgs{
			"campaign_id": campaignID,
		}
		maps.Copy(args, s.scope.SQLArguments())

		var validationCount int
		if err := conn.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(id)
FROM access_review_campaign_validation_checkpoints
WHERE %s
  AND access_review_campaign_id = @campaign_id
`, s.scope.SQLFragment()), args).Scan(&validationCount); err != nil {
			return fmt.Errorf("cannot count validation checkpoints: %w", err)
		}
		if validationCount == 0 {
			return fmt.Errorf("campaign validation checkpoint is required before export")
		}

		total := 0
		pending := 0
		if err := conn.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(id)
FROM access_entries
WHERE %s
  AND access_review_campaign_id = @campaign_id
`, s.scope.SQLFragment()), args).Scan(&total); err != nil {
			return fmt.Errorf("cannot count entries: %w", err)
		}
		if err := conn.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(id)
FROM access_entries
WHERE %s
  AND access_review_campaign_id = @campaign_id
  AND decision = 'PENDING'
`, s.scope.SQLFragment()), args).Scan(&pending); err != nil {
			return fmt.Errorf("cannot count pending entries: %w", err)
		}

		rows, err := conn.Query(ctx, fmt.Sprintf(`
SELECT decision, COUNT(id)
FROM access_entries
WHERE %s
  AND access_review_campaign_id = @campaign_id
GROUP BY decision
ORDER BY decision
`, s.scope.SQLFragment()), args)
		if err != nil {
			return fmt.Errorf("cannot query decision summary: %w", err)
		}
		defer rows.Close()

		decisions := make([]decisionSummaryRow, 0)
		for rows.Next() {
			var row decisionSummaryRow
			if err := rows.Scan(&row.Decision, &row.Count); err != nil {
				return fmt.Errorf("cannot scan decision summary: %w", err)
			}
			decisions = append(decisions, row)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("cannot iterate decision summary rows: %w", err)
		}

		payloadObj := map[string]any{
			"campaign_id":     campaign.ID,
			"campaign_status": campaign.Status,
			"started_at":      campaign.StartedAt,
			"completed_at":    campaign.CompletedAt,
			"generated_at":    time.Now().UTC(),
			"entry_summary": map[string]any{
				"total":   total,
				"pending": pending,
			},
			"decision_summary": decisions,
		}

		payloadBytes, err := json.Marshal(payloadObj)
		if err != nil {
			return fmt.Errorf("cannot marshal evidence payload: %w", err)
		}
		hash := sha256.Sum256(payloadBytes)
		checksum := hex.EncodeToString(hash[:])

		snapshotID := gid.New(s.scope.GetTenantID(), coredata.AccessReviewCampaignEntityType)
		now := time.Now()
		if _, err := conn.Exec(ctx, `
INSERT INTO access_review_campaign_evidence_snapshots (
    id,
    tenant_id,
    access_review_campaign_id,
    payload,
    checksum_sha256,
    signature,
    created_at
)
VALUES ($1,$2,$3,$4::jsonb,$5,NULL,$6)
ON CONFLICT (access_review_campaign_id) DO UPDATE SET
    payload = EXCLUDED.payload,
    checksum_sha256 = EXCLUDED.checksum_sha256
`,
			snapshotID,
			s.scope.GetTenantID(),
			campaignID,
			string(payloadBytes),
			checksum,
			now,
		); err != nil {
			return fmt.Errorf("cannot store evidence snapshot: %w", err)
		}

		returnPayload = string(payloadBytes)
		returnChecksum = checksum
		return nil
	})
	if err != nil {
		return "", "", err
	}

	return returnPayload, returnChecksum, nil
}

func (s *CampaignService) lockCampaignForUpdate(ctx context.Context, conn pg.Conn, campaignID gid.GID) error {
	q := `
SELECT id
FROM access_review_campaigns
WHERE %s
  AND id = @id
FOR UPDATE
`
	q = fmt.Sprintf(q, s.scope.SQLFragment())
	args := pgx.StrictNamedArgs{"id": campaignID}
	maps.Copy(args, s.scope.SQLArguments())

	var id gid.GID
	if err := conn.QueryRow(ctx, q, args).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return coredata.ErrResourceNotFound
		}
		return fmt.Errorf("cannot lock campaign: %w", err)
	}

	return nil
}

func (s *CampaignService) enqueueSourceFetches(
	ctx context.Context,
	conn pg.Conn,
	campaignID gid.GID,
) error {
	var sources coredata.AccessSources
	if err := sources.LoadScopeSourcesByCampaignID(ctx, conn, s.scope, campaignID); err != nil {
		return fmt.Errorf("cannot load scope sources: %w", err)
	}

	now := time.Now()
	for _, source := range sources {
		_, err := conn.Exec(ctx, `
INSERT INTO access_review_campaign_source_fetches (
	tenant_id,
	access_review_campaign_id,
	access_source_id,
	status,
	fetched_accounts_count,
	attempt_count,
	last_error,
	started_at,
	completed_at,
	created_at,
	updated_at
) VALUES (
	$1,$2,$3,'QUEUED',0,0,NULL,NULL,NULL,$4,$4
)
ON CONFLICT (access_review_campaign_id, access_source_id) DO UPDATE SET
	status = 'QUEUED',
	fetched_accounts_count = 0,
	attempt_count = 0,
	last_error = NULL,
	started_at = NULL,
	completed_at = NULL,
	updated_at = EXCLUDED.updated_at
`, s.scope.GetTenantID(), campaignID, source.ID, now)
		if err != nil {
			return fmt.Errorf("cannot queue source fetch %s: %w", source.ID, err)
		}
	}

	return nil
}

func (s *CampaignService) Cancel(
	ctx context.Context,
	campaignID gid.GID,
) (*coredata.AccessReviewCampaign, error) {
	campaign := &coredata.AccessReviewCampaign{}

	err := s.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := campaign.LoadByID(ctx, conn, s.scope, campaignID); err != nil {
				return fmt.Errorf("cannot load campaign: %w", err)
			}

			if campaign.Status == coredata.AccessReviewCampaignStatusCompleted ||
				campaign.Status == coredata.AccessReviewCampaignStatusCancelled {
				return fmt.Errorf("campaign is already %s", campaign.Status)
			}

			campaign.Status = coredata.AccessReviewCampaignStatusCancelled
			campaign.CompletedAt = ref.Ref(time.Now())
			campaign.UpdatedAt = time.Now()

			if err := campaign.Update(ctx, conn, s.scope); err != nil {
				return fmt.Errorf("cannot update campaign: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return campaign, nil
}

func (s *CampaignService) ListForAccessReviewID(
	ctx context.Context,
	accessReviewID gid.GID,
	cursor *page.Cursor[coredata.AccessReviewCampaignOrderField],
) (*page.Page[*coredata.AccessReviewCampaign, coredata.AccessReviewCampaignOrderField], error) {
	var campaigns coredata.AccessReviewCampaigns

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return campaigns.LoadByAccessReviewID(ctx, conn, s.scope, accessReviewID, cursor)
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(campaigns, cursor), nil
}

func (s *CampaignService) ListSourceFetches(
	ctx context.Context,
	campaignID gid.GID,
) (coredata.AccessReviewCampaignSourceFetches, error) {
	var fetches coredata.AccessReviewCampaignSourceFetches

	err := s.pg.WithConn(ctx, func(conn pg.Conn) error {
		return fetches.LoadByCampaignID(ctx, conn, s.scope, campaignID)
	})
	if err != nil {
		return nil, err
	}

	return fetches, nil
}

func (s *CampaignService) CountForAccessReviewID(
	ctx context.Context,
	accessReviewID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			campaigns := coredata.AccessReviewCampaigns{}
			count, err = campaigns.CountByAccessReviewID(ctx, conn, s.scope, accessReviewID)
			return err
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}
