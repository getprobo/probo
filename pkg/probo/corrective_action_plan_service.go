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
package probo

import (
	"context"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type CorrectiveActionPlanService struct {
	svc *Service
}

type (
	CreateCorrectiveActionPlanRequest struct {
		FindingID   gid.GID
		Title       string
		Description *string
		OwnerID     *gid.GID
		DueDate     *time.Time
	}

	UpdateCorrectiveActionPlanRequest struct {
		ID          gid.GID
		Title       *string
		Description **string
		OwnerID     **gid.GID
		DueDate     **time.Time
	}
)

func (r *CreateCorrectiveActionPlanRequest) Validate() error {
	v := validator.New()

	v.Check(r.FindingID, "finding_id", validator.Required(), validator.GID(coredata.FindingEntityType))
	v.Check(r.Title, "title", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(r.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(r.OwnerID, "owner_id", validator.GID(coredata.MembershipProfileEntityType))

	return v.Error()
}

func (r *UpdateCorrectiveActionPlanRequest) Validate() error {
	v := validator.New()

	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.CorrectiveActionPlanEntityType))
	v.Check(r.Title, "title", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(r.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(r.OwnerID, "owner_id", validator.GID(coredata.MembershipProfileEntityType))

	return v.Error()
}

func (s CorrectiveActionPlanService) Get(
	ctx context.Context,
	scope coredata.Scoper,
	correctiveActionPlanID gid.GID,
) (*coredata.CorrectiveActionPlan, error) {
	plan := &coredata.CorrectiveActionPlan{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return plan.LoadByID(ctx, conn, scope, correctiveActionPlanID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get corrective action plan: %w", err)
	}

	return plan, nil
}

func (s CorrectiveActionPlanService) Create(
	ctx context.Context,
	scope coredata.Scoper,
	req *CreateCorrectiveActionPlanRequest,
) (*coredata.CorrectiveActionPlan, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()
	plan := &coredata.CorrectiveActionPlan{
		ID:          gid.New(scope.GetTenantID(), coredata.CorrectiveActionPlanEntityType),
		FindingID:   req.FindingID,
		Title:       req.Title,
		Description: req.Description,
		OwnerID:     req.OwnerID,
		DueDate:     req.DueDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			finding := &coredata.Finding{}
			if err := finding.LoadByID(ctx, tx, scope, req.FindingID); err != nil {
				return fmt.Errorf("cannot load finding: %w", err)
			}

			plan.OrganizationID = finding.OrganizationID

			if req.OwnerID != nil {
				if err := loadOrganizationProfile(ctx, tx, scope, *req.OwnerID, finding.OrganizationID); err != nil {
					return fmt.Errorf("cannot load owner profile: %w", err)
				}
			}

			if err := plan.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert corrective action plan: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return plan, nil
}

func (s CorrectiveActionPlanService) Update(
	ctx context.Context,
	scope coredata.Scoper,
	req *UpdateCorrectiveActionPlanRequest,
) (*coredata.CorrectiveActionPlan, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	plan := &coredata.CorrectiveActionPlan{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := plan.LoadByID(ctx, tx, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load corrective action plan: %w", err)
			}

			if req.Title != nil {
				plan.Title = *req.Title
			}

			if req.Description != nil {
				plan.Description = *req.Description
			}

			if req.OwnerID != nil {
				if *req.OwnerID != nil {
					if err := loadOrganizationProfile(ctx, tx, scope, **req.OwnerID, plan.OrganizationID); err != nil {
						return fmt.Errorf("cannot load owner profile: %w", err)
					}
				}

				plan.OwnerID = *req.OwnerID
			}

			if req.DueDate != nil {
				plan.DueDate = *req.DueDate
			}

			plan.UpdatedAt = time.Now()

			if err := plan.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update corrective action plan: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return plan, nil
}

func (s CorrectiveActionPlanService) Delete(
	ctx context.Context,
	scope coredata.Scoper,
	correctiveActionPlanID gid.GID,
) error {
	plan := &coredata.CorrectiveActionPlan{ID: correctiveActionPlanID}

	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := plan.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete corrective action plan: %w", err)
			}

			return nil
		},
	)
}

func (s CorrectiveActionPlanService) ListForFindingID(
	ctx context.Context,
	scope coredata.Scoper,
	findingID gid.GID,
	cursor *page.Cursor[coredata.CorrectiveActionPlanOrderField],
) (*page.Page[*coredata.CorrectiveActionPlan, coredata.CorrectiveActionPlanOrderField], error) {
	var plans coredata.CorrectiveActionPlans

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := plans.LoadByFindingID(ctx, conn, scope, findingID, cursor); err != nil {
				return fmt.Errorf("cannot load corrective action plans: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(plans, cursor), nil
}

func (s CorrectiveActionPlanService) CountForFindingID(
	ctx context.Context,
	scope coredata.Scoper,
	findingID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			plans := coredata.CorrectiveActionPlans{}

			count, err = plans.CountByFindingID(ctx, conn, scope, findingID)
			if err != nil {
				return fmt.Errorf("cannot count corrective action plans: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s CorrectiveActionPlanService) ListForInternalControlID(
	ctx context.Context,
	scope coredata.Scoper,
	internalControlID gid.GID,
	cursor *page.Cursor[coredata.CorrectiveActionPlanOrderField],
) (*page.Page[*coredata.CorrectiveActionPlan, coredata.CorrectiveActionPlanOrderField], error) {
	var plans coredata.CorrectiveActionPlans

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := plans.LoadByInternalControlID(ctx, conn, scope, internalControlID, cursor); err != nil {
				return fmt.Errorf("cannot load corrective action plans: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(plans, cursor), nil
}

func (s CorrectiveActionPlanService) CountForInternalControlID(
	ctx context.Context,
	scope coredata.Scoper,
	internalControlID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			plans := coredata.CorrectiveActionPlans{}

			count, err = plans.CountByInternalControlID(ctx, conn, scope, internalControlID)
			if err != nil {
				return fmt.Errorf("cannot count corrective action plans: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s CorrectiveActionPlanService) CreateInternalControlMapping(
	ctx context.Context,
	scope coredata.Scoper,
	correctiveActionPlanID gid.GID,
	internalControlID gid.GID,
) (*coredata.CorrectiveActionPlan, *coredata.InternalControl, error) {
	plan := &coredata.CorrectiveActionPlan{}
	internalControl := &coredata.InternalControl{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := plan.LoadByID(ctx, tx, scope, correctiveActionPlanID); err != nil {
				return fmt.Errorf("cannot load corrective action plan: %w", err)
			}

			if err := internalControl.LoadByID(ctx, tx, scope, internalControlID); err != nil {
				return fmt.Errorf("cannot load internal control: %w", err)
			}

			if internalControl.OrganizationID != plan.OrganizationID {
				return fmt.Errorf("cannot verify internal control organization: %w", coredata.ErrResourceNotFound)
			}

			mapping := coredata.CorrectiveActionPlanInternalControl{
				CorrectiveActionPlanID: plan.ID,
				InternalControlID:      internalControl.ID,
				OrganizationID:         plan.OrganizationID,
				CreatedAt:              time.Now(),
			}

			if err := mapping.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert corrective action plan internal control: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create corrective action plan internal control mapping: %w", err)
	}

	return plan, internalControl, nil
}

func (s CorrectiveActionPlanService) DeleteInternalControlMapping(
	ctx context.Context,
	scope coredata.Scoper,
	correctiveActionPlanID gid.GID,
	internalControlID gid.GID,
) (*coredata.CorrectiveActionPlan, *coredata.InternalControl, error) {
	plan := &coredata.CorrectiveActionPlan{}
	internalControl := &coredata.InternalControl{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := plan.LoadByID(ctx, tx, scope, correctiveActionPlanID); err != nil {
				return fmt.Errorf("cannot load corrective action plan: %w", err)
			}

			if err := internalControl.LoadByID(ctx, tx, scope, internalControlID); err != nil {
				return fmt.Errorf("cannot load internal control: %w", err)
			}

			mapping := coredata.CorrectiveActionPlanInternalControl{}
			if err := mapping.Delete(ctx, tx, scope, plan.ID, internalControl.ID); err != nil {
				return fmt.Errorf("cannot delete corrective action plan internal control: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot delete corrective action plan internal control mapping: %w", err)
	}

	return plan, internalControl, nil
}

func loadOrganizationProfile(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	profileID gid.GID,
	organizationID gid.GID,
) error {
	profile := &coredata.MembershipProfile{}
	if err := profile.LoadByID(ctx, conn, scope, profileID); err != nil {
		return fmt.Errorf("cannot load profile: %w", err)
	}

	if profile.OrganizationID != organizationID {
		return fmt.Errorf("cannot verify profile organization: %w", coredata.ErrResourceNotFound)
	}

	return nil
}
