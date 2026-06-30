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
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func (s *Service) CreateCampaign(
	ctx context.Context,
	scope coredata.Scoper,
	req CreateAccessReviewCampaignRequest,
) (*coredata.AccessReviewCampaign, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	campaign := &coredata.AccessReviewCampaign{
		ID:                gid.New(scope.GetTenantID(), coredata.AccessReviewCampaignEntityType),
		OrganizationID:    req.OrganizationID,
		Name:              req.Name,
		Description:       req.Description,
		Status:            coredata.AccessReviewCampaignStatusDraft,
		FrameworkControls: req.FrameworkControls,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := campaign.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot insert access review campaign: %w", err)
			}

			for _, sourceID := range req.AccessReviewSourceIDs {
				source := &coredata.AccessReviewSource{}
				if err := source.LoadByID(ctx, conn, scope, sourceID); err != nil {
					return fmt.Errorf("cannot load access source %s: %w", sourceID, err)
				}

				if source.OrganizationID != campaign.OrganizationID {
					return fmt.Errorf("cannot create campaign: access source %s does not belong to the same organization", sourceID)
				}

				if err := s.upsertCampaignSource(ctx, conn, scope, campaign.ID, source); err != nil {
					return fmt.Errorf("cannot snapshot scope source: %w", err)
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return campaign, nil
}

func (s *Service) GetCampaign(
	ctx context.Context,
	scope coredata.Scoper,
	campaignID gid.GID,
) (*coredata.AccessReviewCampaign, error) {
	campaign := &coredata.AccessReviewCampaign{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := campaign.LoadByID(ctx, conn, scope, campaignID); err != nil {
				return fmt.Errorf("cannot load campaign: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return campaign, nil
}

func (s *Service) GetCampaignSource(
	ctx context.Context,
	scope coredata.Scoper,
	campaignSourceID gid.GID,
) (*coredata.AccessReviewCampaignSource, error) {
	campaignSource := &coredata.AccessReviewCampaignSource{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := campaignSource.LoadByID(ctx, conn, scope, campaignSourceID); err != nil {
				return fmt.Errorf("cannot load campaign source: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return campaignSource, nil
}

func (s *Service) UpdateCampaign(
	ctx context.Context,
	scope coredata.Scoper,
	req UpdateAccessReviewCampaignRequest,
) (*coredata.AccessReviewCampaign, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("cannot validate update campaign request: %w", err)
	}

	campaign := &coredata.AccessReviewCampaign{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := lockCampaignForUpdate(ctx, conn, scope, req.CampaignID); err != nil {
				return fmt.Errorf("cannot lock campaign: %w", err)
			}

			if err := campaign.LoadByID(ctx, conn, scope, req.CampaignID); err != nil {
				return fmt.Errorf("cannot load campaign: %w", err)
			}

			if campaign.Status != coredata.AccessReviewCampaignStatusDraft {
				return fmt.Errorf("cannot update campaign: status is %s, expected DRAFT", campaign.Status)
			}

			if req.Name != nil && *req.Name != nil {
				campaign.Name = **req.Name
			}

			if req.Description != nil && *req.Description != nil {
				campaign.Description = **req.Description
			}

			if req.FrameworkControls != nil {
				campaign.FrameworkControls = *req.FrameworkControls
			}

			campaign.UpdatedAt = time.Now()

			if err := campaign.Update(ctx, conn, scope); err != nil {
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

func (s *Service) DeleteCampaign(
	ctx context.Context,
	scope coredata.Scoper,
	campaignID gid.GID,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := lockCampaignForUpdate(ctx, conn, scope, campaignID); err != nil {
				return fmt.Errorf("cannot lock campaign: %w", err)
			}

			campaign := &coredata.AccessReviewCampaign{}
			if err := campaign.LoadByID(ctx, conn, scope, campaignID); err != nil {
				return fmt.Errorf("cannot load campaign: %w", err)
			}

			if campaign.Status != coredata.AccessReviewCampaignStatusDraft &&
				campaign.Status != coredata.AccessReviewCampaignStatusCancelled {
				return fmt.Errorf("cannot delete campaign: status is %s, expected %s or %s", campaign.Status, coredata.AccessReviewCampaignStatusDraft, coredata.AccessReviewCampaignStatusCancelled)
			}

			if err := campaign.Delete(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot delete campaign: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) AddCampaignSource(
	ctx context.Context,
	scope coredata.Scoper,
	req AddCampaignSourceRequest,
) (*coredata.AccessReviewCampaign, error) {
	campaign := &coredata.AccessReviewCampaign{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := lockCampaignForUpdate(ctx, conn, scope, req.CampaignID); err != nil {
				return fmt.Errorf("cannot lock campaign: %w", err)
			}

			if err := campaign.LoadByID(ctx, conn, scope, req.CampaignID); err != nil {
				return fmt.Errorf("cannot load campaign: %w", err)
			}

			if campaign.Status != coredata.AccessReviewCampaignStatusDraft {
				return fmt.Errorf("cannot add scope source: campaign status is %s, expected %s", campaign.Status, coredata.AccessReviewCampaignStatusDraft)
			}

			source := &coredata.AccessReviewSource{}
			if err := source.LoadByID(ctx, conn, scope, req.AccessReviewSourceID); err != nil {
				return fmt.Errorf("cannot load access source %s: %w", req.AccessReviewSourceID, err)
			}

			if source.OrganizationID != campaign.OrganizationID {
				return fmt.Errorf("cannot add scope source: access source %q does not belong to the same organization", req.AccessReviewSourceID)
			}

			if err := s.upsertCampaignSource(ctx, conn, scope, campaign.ID, source); err != nil {
				return fmt.Errorf("cannot snapshot scope source: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return campaign, nil
}

func (s *Service) RemoveCampaignSource(
	ctx context.Context,
	scope coredata.Scoper,
	req RemoveCampaignSourceRequest,
) (*coredata.AccessReviewCampaign, error) {
	campaign := &coredata.AccessReviewCampaign{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := lockCampaignForUpdate(ctx, conn, scope, req.CampaignID); err != nil {
				return fmt.Errorf("cannot lock campaign: %w", err)
			}

			if err := campaign.LoadByID(ctx, conn, scope, req.CampaignID); err != nil {
				return fmt.Errorf("cannot load campaign: %w", err)
			}

			if campaign.Status != coredata.AccessReviewCampaignStatusDraft {
				return fmt.Errorf("cannot remove scope source: campaign status is %s, expected DRAFT", campaign.Status)
			}

			campaignSource := &coredata.AccessReviewCampaignSource{}
			if err := campaignSource.DeleteByCampaignIDAndAccessReviewSourceID(ctx, conn, scope, campaign.ID, req.AccessReviewSourceID); err != nil {
				return fmt.Errorf("cannot delete campaign source: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return campaign, nil
}

func (s *Service) StartCampaign(
	ctx context.Context,
	scope coredata.Scoper,
	campaignID gid.GID,
) (*coredata.AccessReviewCampaign, error) {
	campaign := &coredata.AccessReviewCampaign{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := lockCampaignForUpdate(ctx, conn, scope, campaignID); err != nil {
				return fmt.Errorf("cannot lock campaign: %w", err)
			}

			if err := campaign.LoadByID(ctx, conn, scope, campaignID); err != nil {
				return fmt.Errorf("cannot load campaign: %w", err)
			}

			if campaign.Status != coredata.AccessReviewCampaignStatusDraft {
				return fmt.Errorf("cannot start campaign: status is %s, expected %s", campaign.Status, coredata.AccessReviewCampaignStatusDraft)
			}

			var campaignSources coredata.AccessReviewCampaignSources
			if err := campaignSources.LoadByCampaignID(ctx, conn, scope, campaign.ID); err != nil {
				return fmt.Errorf("cannot load campaign sources: %w", err)
			}

			if len(campaignSources) == 0 {
				return fmt.Errorf("cannot start campaign: no scope sources configured")
			}

			now := time.Now()
			campaign.Status = coredata.AccessReviewCampaignStatusInProgress
			campaign.StartedAt = &now
			campaign.UpdatedAt = now

			if err := campaign.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update campaign: %w", err)
			}

			if err := s.enqueueSourceFetches(ctx, conn, scope, campaignSources); err != nil {
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

func (s *Service) CloseCampaign(
	ctx context.Context,
	scope coredata.Scoper,
	campaignID gid.GID,
) (*coredata.AccessReviewCampaign, error) {
	campaign := &coredata.AccessReviewCampaign{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := lockCampaignForUpdate(ctx, conn, scope, campaignID); err != nil {
				return fmt.Errorf("cannot lock campaign: %w", err)
			}

			if err := campaign.LoadByID(ctx, conn, scope, campaignID); err != nil {
				return fmt.Errorf("cannot load campaign: %w", err)
			}

			if campaign.Status != coredata.AccessReviewCampaignStatusPendingActions {
				return fmt.Errorf("cannot close campaign: status is %s, expected %s", campaign.Status, coredata.AccessReviewCampaignStatusPendingActions)
			}

			entries := coredata.AccessReviewEntries{}
			filter := &coredata.AccessReviewEntryFilter{
				Decision: new(coredata.AccessReviewEntryDecisionPending),
			}

			pendingCount, err := entries.CountByCampaignID(
				ctx,
				conn,
				scope,
				campaignID,
				filter,
			)
			if err != nil {
				return fmt.Errorf("cannot count pending entries: %w", err)
			}

			if pendingCount > 0 {
				return fmt.Errorf("cannot close campaign: %d entries still pending", pendingCount)
			}

			now := time.Now()
			campaign.Status = coredata.AccessReviewCampaignStatusCompleted
			campaign.CompletedAt = &now
			campaign.UpdatedAt = now

			if err := campaign.Update(ctx, conn, scope); err != nil {
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

func lockCampaignForUpdate(ctx context.Context, tx pg.Tx, scope coredata.Scoper, campaignID gid.GID) error {
	c := &coredata.AccessReviewCampaign{ID: campaignID}
	if err := c.LockForUpdate(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot lock campaign for update: %w", err)
	}

	return nil
}

// upsertCampaignSource snapshots a live access source into the campaign's scope
// so the review keeps the source identity even if the source is later deleted.
func (s *Service) upsertCampaignSource(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	campaignID gid.GID,
	source *coredata.AccessReviewSource,
) error {
	now := time.Now()
	sourceID := source.ID

	campaignSource := &coredata.AccessReviewCampaignSource{
		ID:                     gid.New(scope.GetTenantID(), coredata.AccessReviewCampaignSourceEntityType),
		OrganizationID:         source.OrganizationID,
		AccessReviewCampaignID: campaignID,
		AccessReviewSourceID:   &sourceID,
		Name:                   source.Name,
		ConnectorID:            source.ConnectorID,
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	if err := campaignSource.Upsert(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot upsert campaign source %s: %w", source.ID, err)
	}

	return nil
}

func (s *Service) enqueueSourceFetches(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	campaignSources coredata.AccessReviewCampaignSources,
) error {
	now := time.Now()

	for _, campaignSource := range campaignSources {
		attempt := &coredata.AccessReviewCampaignSourceFetchAttempt{
			ID:                           gid.New(scope.GetTenantID(), coredata.AccessReviewCampaignSourceFetchAttemptEntityType),
			OrganizationID:               campaignSource.OrganizationID,
			AccessReviewCampaignSourceID: campaignSource.ID,
			Status:                       coredata.AccessReviewCampaignSourceFetchStatusQueued,
			CreatedAt:                    now,
			UpdatedAt:                    now,
		}
		if err := attempt.Insert(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot queue source fetch %s: %w", campaignSource.ID, err)
		}
	}

	return nil
}

func (s *Service) CancelCampaign(
	ctx context.Context,
	scope coredata.Scoper,
	campaignID gid.GID,
) (*coredata.AccessReviewCampaign, error) {
	campaign := &coredata.AccessReviewCampaign{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := lockCampaignForUpdate(ctx, conn, scope, campaignID); err != nil {
				return fmt.Errorf("cannot lock campaign: %w", err)
			}

			if err := campaign.LoadByID(ctx, conn, scope, campaignID); err != nil {
				return fmt.Errorf("cannot load campaign: %w", err)
			}

			if campaign.Status == coredata.AccessReviewCampaignStatusCompleted ||
				campaign.Status == coredata.AccessReviewCampaignStatusCancelled {
				return fmt.Errorf("cannot update campaign: already %s", campaign.Status)
			}

			now := time.Now()
			campaign.Status = coredata.AccessReviewCampaignStatusCancelled
			campaign.CompletedAt = &now
			campaign.UpdatedAt = now

			if err := campaign.Update(ctx, conn, scope); err != nil {
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

func (s *Service) ListCampaignsForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.AccessReviewCampaignOrderField],
) (*page.Page[*coredata.AccessReviewCampaign, coredata.AccessReviewCampaignOrderField], error) {
	var campaigns coredata.AccessReviewCampaigns

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := campaigns.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor); err != nil {
				return fmt.Errorf("cannot load campaigns by organization: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(campaigns, cursor), nil
}

func (s *Service) ListCampaignSources(
	ctx context.Context,
	scope coredata.Scoper,
	campaignID gid.GID,
) (coredata.AccessReviewCampaignSources, error) {
	var sources coredata.AccessReviewCampaignSources

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := sources.LoadByCampaignID(ctx, conn, scope, campaignID); err != nil {
				return fmt.Errorf("cannot load campaign sources: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return sources, nil
}

func (s *Service) ListFetchAttemptsForCampaignSourceID(
	ctx context.Context,
	scope coredata.Scoper,
	campaignSourceID gid.GID,
	cursor *page.Cursor[coredata.AccessReviewCampaignSourceFetchAttemptOrderField],
) (*page.Page[*coredata.AccessReviewCampaignSourceFetchAttempt, coredata.AccessReviewCampaignSourceFetchAttemptOrderField], error) {
	var attempts coredata.AccessReviewCampaignSourceFetchAttempts

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := attempts.LoadByCampaignSourceID(ctx, conn, scope, campaignSourceID, cursor); err != nil {
				return fmt.Errorf("cannot load fetch attempts: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(attempts, cursor), nil
}

func (s *Service) CountFetchAttemptsForCampaignSourceID(
	ctx context.Context,
	scope coredata.Scoper,
	campaignSourceID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var attempts coredata.AccessReviewCampaignSourceFetchAttempts

			var err error

			count, err = attempts.CountByCampaignSourceID(ctx, conn, scope, campaignSourceID)
			if err != nil {
				return fmt.Errorf("cannot count fetch attempts: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) CountCampaignsForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			campaigns := coredata.AccessReviewCampaigns{}

			count, err = campaigns.CountByOrganizationID(ctx, conn, scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot count campaigns by organization: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}
