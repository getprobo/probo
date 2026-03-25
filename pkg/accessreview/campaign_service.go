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
		func(conn pg.Conn) error {
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
				if err := scopeSystem.Insert(ctx, conn); err != nil {
					return fmt.Errorf("cannot insert scope system: %w", err)
				}
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
			if err := s.lockCampaignForUpdate(ctx, conn, req.CampaignID); err != nil {
				return err
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
		func(conn pg.Conn) error {
			if err := s.lockCampaignForUpdate(ctx, conn, campaignID); err != nil {
				return err
			}

			campaign := &coredata.AccessReviewCampaign{}
			if err := campaign.LoadByID(ctx, conn, s.scope, campaignID); err != nil {
				return fmt.Errorf("cannot load campaign: %w", err)
			}

			if campaign.Status != coredata.AccessReviewCampaignStatusDraft &&
				campaign.Status != coredata.AccessReviewCampaignStatusCancelled {
				return fmt.Errorf("cannot delete campaign: status is %s, expected DRAFT or CANCELLED", campaign.Status)
			}

			return campaign.Delete(ctx, conn, s.scope)
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
		func(conn pg.Conn) error {
			if err := s.lockCampaignForUpdate(ctx, conn, req.CampaignID); err != nil {
				return err
			}

			if err := campaign.LoadByID(ctx, conn, s.scope, req.CampaignID); err != nil {
				return fmt.Errorf("cannot load campaign: %w", err)
			}

			if campaign.Status != coredata.AccessReviewCampaignStatusDraft {
				return fmt.Errorf("cannot add scope source: campaign status is %s, expected DRAFT", campaign.Status)
			}

			source := &coredata.AccessSource{}
			if err := source.LoadByID(ctx, conn, s.scope, req.AccessSourceID); err != nil {
				return fmt.Errorf("cannot load access source %s: %w", req.AccessSourceID, err)
			}

			if source.OrganizationID != campaign.OrganizationID {
				return fmt.Errorf("cannot add scope source: access source %s does not belong to the same organization", req.AccessSourceID)
			}

			scopeSystem := coredata.AccessReviewCampaignScopeSystem{
				AccessReviewCampaignID: campaign.ID,
				AccessSourceID:         req.AccessSourceID,
			}
			if err := scopeSystem.Upsert(ctx, conn); err != nil {
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
		func(conn pg.Conn) error {
			if err := s.lockCampaignForUpdate(ctx, conn, req.CampaignID); err != nil {
				return err
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
			if err := scopeSystem.Delete(ctx, conn); err != nil {
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
		func(conn pg.Conn) error {
			if err := s.lockCampaignForUpdate(ctx, conn, campaignID); err != nil {
				return err
			}

			if err := campaign.LoadByID(ctx, conn, s.scope, campaignID); err != nil {
				return fmt.Errorf("cannot load campaign: %w", err)
			}

			if campaign.Status != coredata.AccessReviewCampaignStatusDraft && campaign.Status != coredata.AccessReviewCampaignStatusFailed {
				return fmt.Errorf("cannot start campaign: status is %s, expected DRAFT or FAILED", campaign.Status)
			}

			now := time.Now()
			campaign.Status = coredata.AccessReviewCampaignStatusInProgress
			campaign.StartedAt = &now
			campaign.UpdatedAt = now

			if err := campaign.Update(ctx, conn, s.scope); err != nil {
				return fmt.Errorf("cannot update campaign: %w", err)
			}

			var sources coredata.AccessSources
			if err := sources.LoadScopeSourcesByCampaignID(ctx, conn, s.scope, campaign.ID); err != nil {
				return fmt.Errorf("cannot load scope sources: %w", err)
			}

			if len(sources) == 0 {
				return fmt.Errorf("cannot start campaign: no scope sources configured")
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

			if campaign.Status != coredata.AccessReviewCampaignStatusPendingActions {
				return fmt.Errorf("cannot close campaign: status is %s, expected PENDING_ACTIONS", campaign.Status)
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

func (s *CampaignService) lockCampaignForUpdate(ctx context.Context, conn pg.Conn, campaignID gid.GID) error {
	return lockCampaignForUpdate(ctx, conn, s.scope, campaignID)
}

func lockCampaignForUpdate(ctx context.Context, conn pg.Conn, scope coredata.Scoper, campaignID gid.GID) error {
	c := &coredata.AccessReviewCampaign{ID: campaignID}
	return c.LockForUpdate(ctx, conn, scope)
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
		fetch := &coredata.AccessReviewCampaignSourceFetch{
			AccessReviewCampaignID: campaignID,
			AccessSourceID:         source.ID,
		}
		if err := fetch.UpsertQueued(ctx, conn, s.scope.GetTenantID(), now); err != nil {
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
			if err := s.lockCampaignForUpdate(ctx, conn, campaignID); err != nil {
				return err
			}

			if err := campaign.LoadByID(ctx, conn, s.scope, campaignID); err != nil {
				return fmt.Errorf("cannot load campaign: %w", err)
			}

			if campaign.Status == coredata.AccessReviewCampaignStatusCompleted ||
				campaign.Status == coredata.AccessReviewCampaignStatusCancelled {
				return fmt.Errorf("cannot update campaign: already %s", campaign.Status)
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

func (s *CampaignService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.AccessReviewCampaignOrderField],
) (*page.Page[*coredata.AccessReviewCampaign, coredata.AccessReviewCampaignOrderField], error) {
	var campaigns coredata.AccessReviewCampaigns

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return campaigns.LoadByOrganizationID(ctx, conn, s.scope, organizationID, cursor)
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

func (s *CampaignService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			campaigns := coredata.AccessReviewCampaigns{}
			count, err = campaigns.CountByOrganizationID(ctx, conn, s.scope, organizationID)
			return err
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}
