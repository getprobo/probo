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
	"errors"
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
		ID:             gid.New(scope.GetTenantID(), coredata.AccessReviewCampaignEntityType),
		OrganizationID: req.OrganizationID,
		Name:           req.Name,
		Description:    req.Description,
		Status:         coredata.AccessReviewCampaignStatusDraft,
		CreatedAt:      now,
		UpdatedAt:      now,
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
					if errors.Is(err, coredata.ErrResourceNotFound) {
						return coredata.ErrResourceNotFound
					}

					return fmt.Errorf("cannot load access source: %w", err)
				}

				if source.OrganizationID != campaign.OrganizationID {
					return coredata.ErrResourceNotFound
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
				return fmt.Errorf("cannot update campaign: %w", CampaignStatusError(campaign.Status))
			}

			if req.Name != nil && *req.Name != nil {
				campaign.Name = **req.Name
			}

			if req.Description != nil && *req.Description != nil {
				campaign.Description = **req.Description
			}

			campaign.UpdatedAt = time.Now()

			if err := campaign.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update campaign: %w", err)
			}

			if req.AccessReviewSourceIDs != nil {
				if err := s.syncCampaignSources(ctx, conn, scope, campaign, *req.AccessReviewSourceIDs); err != nil {
					return err
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
				return fmt.Errorf("cannot add scope source: %w", CampaignStatusError(campaign.Status))
			}
			if err := source.LoadByID(ctx, conn, scope, req.AccessReviewSourceID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return coredata.ErrResourceNotFound
				}

				return fmt.Errorf("cannot load access source: %w", err)
			}

			if source.OrganizationID != campaign.OrganizationID {
				return coredata.ErrResourceNotFound
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
				return fmt.Errorf("cannot remove scope source: %w", CampaignStatusError(campaign.Status))
			}
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

func (s *Service) syncCampaignSources(
	ctx context.Context,
	conn pg.Tx,
	scope coredata.Scoper,
	campaign *coredata.AccessReviewCampaign,
	sourceIDs []gid.GID,
) error {
	var campaignSources coredata.AccessReviewCampaignSources
	if err := campaignSources.LoadByCampaignID(ctx, conn, scope, campaign.ID); err != nil {
		return fmt.Errorf("cannot load campaign sources: %w", err)
	}

	existingSourceIDs := make([]gid.GID, 0, len(campaignSources))
	for _, campaignSource := range campaignSources {
		if campaignSource.AccessReviewSourceID != nil {
			existingSourceIDs = append(existingSourceIDs, *campaignSource.AccessReviewSourceID)
		}
	}

	for _, sourceID := range sourceIDs {
		if containsGID(existingSourceIDs, sourceID) {
			continue
		}

		source := &coredata.AccessReviewSource{}
		if err := source.LoadByID(ctx, conn, scope, sourceID); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return coredata.ErrResourceNotFound
			}

			return fmt.Errorf("cannot load access source: %w", err)
		}

		if source.OrganizationID != campaign.OrganizationID {
			return coredata.ErrResourceNotFound
		}

		if err := s.upsertCampaignSource(ctx, conn, scope, campaign.ID, source); err != nil {
			return fmt.Errorf("cannot snapshot scope source: %w", err)
		}
	}

	for _, existingSourceID := range existingSourceIDs {
		if containsGID(sourceIDs, existingSourceID) {
			continue
		}

		campaignSource := &coredata.AccessReviewCampaignSource{}
		if err := campaignSource.DeleteByCampaignIDAndAccessReviewSourceID(
			ctx,
			conn,
			scope,
			campaign.ID,
			existingSourceID,
		); err != nil {
			return fmt.Errorf("cannot delete campaign source: %w", err)
		}
	}

	return nil
}

func containsGID(ids []gid.GID, id gid.GID) bool {
	for _, candidate := range ids {
		if candidate == id {
			return true
		}
	}

	return false
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
				return fmt.Errorf("cannot start campaign: %w", CampaignStatusError(campaign.Status))
			}

			var campaignSources coredata.AccessReviewCampaignSources
			if err := campaignSources.LoadByCampaignID(ctx, conn, scope, campaign.ID); err != nil {
				return fmt.Errorf("cannot load campaign sources: %w", err)
			}

			if len(campaignSources) == 0 {
				return ErrCampaignMissingSources
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
