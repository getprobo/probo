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
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type CampaignService struct {
	pg    *pg.Client
	scope coredata.Scoper
}

func NewCampaignService(pgClient *pg.Client, scope coredata.Scoper) *CampaignService {
	return &CampaignService{
		pg:    pgClient,
		scope: scope,
	}
}

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
			if err := campaign.Insert(ctx, conn, s.scope); err != nil {
				return fmt.Errorf("cannot insert access review campaign: %w", err)
			}

			for _, sourceID := range req.AccessSourceIDs {
				source := &coredata.AccessSource{}
				if err := source.LoadByID(ctx, conn, s.scope, sourceID); err != nil {
					return fmt.Errorf("cannot load access source %s: %w", sourceID, err)
				}

				if source.OrganizationID != campaign.OrganizationID {
					return fmt.Errorf("cannot create campaign: access source %s does not belong to the same organization", sourceID)
				}

				scopeSystem := coredata.AccessReviewCampaignScopeSystem{
					AccessReviewCampaignID: campaign.ID,
					AccessSourceID:         sourceID,
				}
				if err := scopeSystem.Insert(ctx, conn, s.scope); err != nil {
					return fmt.Errorf("cannot insert scope system: %w", err)
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

func (s *CampaignService) Get(
	ctx context.Context,
	campaignID gid.GID,
) (*coredata.AccessReviewCampaign, error) {
	campaign := &coredata.AccessReviewCampaign{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := campaign.LoadByID(ctx, conn, s.scope, campaignID); err != nil {
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

func (s *CampaignService) Update(
	ctx context.Context,
	req UpdateAccessReviewCampaignRequest,
) (*coredata.AccessReviewCampaign, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("cannot validate update campaign request: %w", err)
	}

	campaign := &coredata.AccessReviewCampaign{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := lockCampaignForUpdate(ctx, conn, s.scope, req.CampaignID); err != nil {
				return fmt.Errorf("cannot lock campaign: %w", err)
			}

			if err := campaign.LoadByID(ctx, conn, s.scope, req.CampaignID); err != nil {
				return fmt.Errorf("cannot load campaign: %w", err)
			}

			if campaign.Status != coredata.AccessReviewCampaignStatusDraft {
				return fmt.Errorf("cannot update campaign: status is %s, expected DRAFT", campaign.Status)
			}

			if req.Name != nil {
				campaign.Name = *req.Name
			}

			if req.Description != nil {
				campaign.Description = *req.Description
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
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := lockCampaignForUpdate(ctx, conn, s.scope, campaignID); err != nil {
				return fmt.Errorf("cannot lock campaign: %w", err)
			}

			campaign := &coredata.AccessReviewCampaign{}
			if err := campaign.LoadByID(ctx, conn, s.scope, campaignID); err != nil {
				return fmt.Errorf("cannot load campaign: %w", err)
			}

			if campaign.Status != coredata.AccessReviewCampaignStatusDraft &&
				campaign.Status != coredata.AccessReviewCampaignStatusCancelled {
				return fmt.Errorf("cannot delete campaign: status is %s, expected %s or %s", campaign.Status, coredata.AccessReviewCampaignStatusDraft, coredata.AccessReviewCampaignStatusCancelled)
			}

			if err := campaign.Delete(ctx, conn, s.scope); err != nil {
				return fmt.Errorf("cannot delete campaign: %w", err)
			}
			return nil
		},
	)
}

func (s *CampaignService) AddScopeSource(
	ctx context.Context,
	req AddCampaignScopeSourceRequest,
) (*coredata.AccessReviewCampaign, error) {
	campaign := &coredata.AccessReviewCampaign{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := lockCampaignForUpdate(ctx, conn, s.scope, req.CampaignID); err != nil {
				return fmt.Errorf("cannot lock campaign: %w", err)
			}

			if err := campaign.LoadByID(ctx, conn, s.scope, req.CampaignID); err != nil {
				return fmt.Errorf("cannot load campaign: %w", err)
			}

			if campaign.Status != coredata.AccessReviewCampaignStatusDraft {
				return fmt.Errorf("cannot add scope source: campaign status is %s, expected %s", campaign.Status, coredata.AccessReviewCampaignStatusDraft)
			}

			source := &coredata.AccessSource{}
			if err := source.LoadByID(ctx, conn, s.scope, req.AccessSourceID); err != nil {
				return fmt.Errorf("cannot load access source %s: %w", req.AccessSourceID, err)
			}

			if source.OrganizationID != campaign.OrganizationID {
				return fmt.Errorf("cannot add scope source: access source %q does not belong to the same organization", req.AccessSourceID)
			}

			scopeSystem := coredata.AccessReviewCampaignScopeSystem{
				AccessReviewCampaignID: campaign.ID,
				AccessSourceID:         req.AccessSourceID,
			}
			if err := scopeSystem.Upsert(ctx, conn, s.scope); err != nil {
				return fmt.Errorf("cannot upsert scope system: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return campaign, nil
}

func (s *CampaignService) RemoveScopeSource(
	ctx context.Context,
	req RemoveCampaignScopeSourceRequest,
) (*coredata.AccessReviewCampaign, error) {
	campaign := &coredata.AccessReviewCampaign{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := lockCampaignForUpdate(ctx, conn, s.scope, req.CampaignID); err != nil {
				return fmt.Errorf("cannot lock campaign: %w", err)
			}

			if err := campaign.LoadByID(ctx, conn, s.scope, req.CampaignID); err != nil {
				return fmt.Errorf("cannot load campaign: %w", err)
			}

			if campaign.Status != coredata.AccessReviewCampaignStatusDraft {
				return fmt.Errorf("cannot remove scope source: campaign status is %s, expected DRAFT", campaign.Status)
			}

			scopeSystem := coredata.AccessReviewCampaignScopeSystem{
				AccessReviewCampaignID: campaign.ID,
				AccessSourceID:         req.AccessSourceID,
			}
			if err := scopeSystem.Delete(ctx, conn, s.scope); err != nil {
				return fmt.Errorf("cannot delete scope system: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return campaign, nil
}

func (s *CampaignService) Start(
	ctx context.Context,
	campaignID gid.GID,
) (*coredata.AccessReviewCampaign, error) {
	campaign := &coredata.AccessReviewCampaign{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := lockCampaignForUpdate(ctx, conn, s.scope, campaignID); err != nil {
				return fmt.Errorf("cannot lock campaign: %w", err)
			}

			if err := campaign.LoadByID(ctx, conn, s.scope, campaignID); err != nil {
				return fmt.Errorf("cannot load campaign: %w", err)
			}

			if campaign.Status != coredata.AccessReviewCampaignStatusDraft &&
				campaign.Status != coredata.AccessReviewCampaignStatusFailed {
				return fmt.Errorf("cannot start campaign: status is %s, expected %s or %s", campaign.Status, coredata.AccessReviewCampaignStatusDraft, coredata.AccessReviewCampaignStatusFailed)
			}

			var sources coredata.AccessSources
			if err := sources.LoadScopeSourcesByCampaignID(ctx, conn, s.scope, campaign.ID); err != nil {
				return fmt.Errorf("cannot load scope sources: %w", err)
			}

			if len(sources) == 0 {
				return fmt.Errorf("cannot start campaign: no scope sources configured")
			}

			now := time.Now()
			campaign.Status = coredata.AccessReviewCampaignStatusInProgress
			campaign.StartedAt = &now
			campaign.UpdatedAt = now

			if err := campaign.Update(ctx, conn, s.scope); err != nil {
				return fmt.Errorf("cannot update campaign: %w", err)
			}

			if err := s.enqueueSourceFetches(ctx, conn, campaign.ID, sources); err != nil {
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

func (s *CampaignService) Close(
	ctx context.Context,
	campaignID gid.GID,
) (*coredata.AccessReviewCampaign, error) {
	campaign := &coredata.AccessReviewCampaign{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := lockCampaignForUpdate(ctx, conn, s.scope, campaignID); err != nil {
				return fmt.Errorf("cannot lock campaign: %w", err)
			}

			if err := campaign.LoadByID(ctx, conn, s.scope, campaignID); err != nil {
				return fmt.Errorf("cannot load campaign: %w", err)
			}

			if campaign.Status != coredata.AccessReviewCampaignStatusPendingActions {
				return fmt.Errorf("cannot close campaign: status is %s, expected %s", campaign.Status, coredata.AccessReviewCampaignStatusPendingActions)
			}

			entries := coredata.AccessEntries{}
			pendingCount, err := entries.CountPendingByCampaignID(ctx, conn, s.scope, campaignID)
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

func lockCampaignForUpdate(ctx context.Context, tx pg.Tx, scope coredata.Scoper, campaignID gid.GID) error {
	c := &coredata.AccessReviewCampaign{ID: campaignID}
	if err := c.LockForUpdate(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot lock campaign for update: %w", err)
	}
	return nil
}

func (s *CampaignService) enqueueSourceFetches(
	ctx context.Context,
	tx pg.Tx,
	campaignID gid.GID,
	sources coredata.AccessSources,
) error {
	now := time.Now()
	for _, source := range sources {
		fetch := &coredata.AccessReviewCampaignSourceFetch{
			AccessReviewCampaignID: campaignID,
			AccessSourceID:         source.ID,
		}
		if err := fetch.UpsertQueued(ctx, tx, s.scope, now); err != nil {
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
		func(ctx context.Context, conn pg.Tx) error {
			if err := lockCampaignForUpdate(ctx, conn, s.scope, campaignID); err != nil {
				return fmt.Errorf("cannot lock campaign: %w", err)
			}

			if err := campaign.LoadByID(ctx, conn, s.scope, campaignID); err != nil {
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

func (s *CampaignService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.AccessReviewCampaignOrderField],
) (*page.Page[*coredata.AccessReviewCampaign, coredata.AccessReviewCampaignOrderField], error) {
	var campaigns coredata.AccessReviewCampaigns

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := campaigns.LoadByOrganizationID(ctx, conn, s.scope, organizationID, cursor); err != nil {
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

func (s *CampaignService) ListSourceFetches(
	ctx context.Context,
	campaignID gid.GID,
) (coredata.AccessReviewCampaignSourceFetches, error) {
	var fetches coredata.AccessReviewCampaignSourceFetches

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := fetches.LoadByCampaignID(ctx, conn, s.scope, campaignID); err != nil {
				return fmt.Errorf("cannot load source fetches by campaign: %w", err)
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return fetches, nil
}

func (s *CampaignService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			campaigns := coredata.AccessReviewCampaigns{}
			count, err = campaigns.CountByOrganizationID(ctx, conn, s.scope, organizationID)
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
