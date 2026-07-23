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

type BusinessFunctionService struct {
	svc *Service
}

type (
	CreateBusinessFunctionRequest struct {
		OrganizationID  gid.GID
		ReferenceID     string
		Name            string
		Classification  coredata.BusinessFunctionClassification
		MTDMinutes      int
		RTOMinutes      int
		RPOMinutes      int
		ImpactTolerance *string
		Notes           *string
		OwnerID         *gid.GID
		AssetIDs        []gid.GID
		ThirdPartyIDs   []gid.GID
	}

	UpdateBusinessFunctionRequest struct {
		ID              gid.GID
		ReferenceID     *string
		Name            *string
		Classification  *coredata.BusinessFunctionClassification
		MTDMinutes      *int
		RTOMinutes      *int
		RPOMinutes      *int
		ImpactTolerance **string
		Notes           **string
		OwnerID         **gid.GID
		AssetIDs        *[]gid.GID
		ThirdPartyIDs   *[]gid.GID
	}
)

func (r *CreateBusinessFunctionRequest) Validate() error {
	v := validator.New()

	v.Check(r.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(r.ReferenceID, "reference_id", validator.Required(), validator.SafeText(NameMaxLength))
	v.Check(r.Name, "name", validator.Required(), validator.SafeText(TitleMaxLength))
	v.Check(r.Classification, "classification", validator.Required(), validator.OneOfSlice(coredata.BusinessFunctionClassifications()))
	v.Check(r.MTDMinutes, "mtd_minutes", validator.Required(), validator.Min(0))
	v.Check(r.RTOMinutes, "rto_minutes", validator.Required(), validator.Min(0))
	v.Check(r.RPOMinutes, "rpo_minutes", validator.Required(), validator.Min(0))
	v.Check(r.ImpactTolerance, "impact_tolerance", validator.SafeText(ContentMaxLength))
	v.Check(r.Notes, "notes", validator.SafeText(ContentMaxLength))
	v.Check(r.OwnerID, "owner_id", validator.GID(coredata.MembershipProfileEntityType))
	v.CheckEach(
		r.AssetIDs,
		"asset_ids",
		func(index int, item any) {
			v.Check(item, fmt.Sprintf("asset_ids[%d]", index), validator.Required(), validator.GID(coredata.AssetEntityType))
		},
	)
	v.CheckEach(
		r.ThirdPartyIDs,
		"third_party_ids",
		func(index int, item any) {
			v.Check(item, fmt.Sprintf("third_party_ids[%d]", index), validator.Required(), validator.GID(coredata.ThirdPartyEntityType))
		},
	)

	return v.Error()
}

func (r *UpdateBusinessFunctionRequest) Validate() error {
	v := validator.New()

	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.BusinessFunctionEntityType))
	v.Check(r.ReferenceID, "reference_id", validator.SafeText(NameMaxLength))
	v.Check(r.Name, "name", validator.SafeText(TitleMaxLength))
	v.Check(r.Classification, "classification", validator.OneOfSlice(coredata.BusinessFunctionClassifications()))
	v.Check(r.MTDMinutes, "mtd_minutes", validator.Min(0))
	v.Check(r.RTOMinutes, "rto_minutes", validator.Min(0))
	v.Check(r.RPOMinutes, "rpo_minutes", validator.Min(0))
	v.Check(r.ImpactTolerance, "impact_tolerance", validator.SafeText(ContentMaxLength))
	v.Check(r.Notes, "notes", validator.SafeText(ContentMaxLength))
	v.Check(r.OwnerID, "owner_id", validator.GID(coredata.MembershipProfileEntityType))
	v.CheckEach(
		r.AssetIDs,
		"asset_ids",
		func(index int, item any) {
			v.Check(item, fmt.Sprintf("asset_ids[%d]", index), validator.GID(coredata.AssetEntityType))
		},
	)
	v.CheckEach(
		r.ThirdPartyIDs,
		"third_party_ids",
		func(index int, item any) {
			v.Check(item, fmt.Sprintf("third_party_ids[%d]", index), validator.GID(coredata.ThirdPartyEntityType))
		},
	)

	return v.Error()
}

func (s BusinessFunctionService) Get(
	ctx context.Context,
	scope coredata.Scoper,
	businessFunctionID gid.GID,
) (*coredata.BusinessFunction, error) {
	businessFunction := &coredata.BusinessFunction{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return businessFunction.LoadByID(ctx, conn, scope, businessFunctionID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get business function: %w", err)
	}

	return businessFunction, nil
}

func (s *BusinessFunctionService) Create(
	ctx context.Context,
	scope coredata.Scoper,
	req *CreateBusinessFunctionRequest,
) (*coredata.BusinessFunction, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()

	businessFunction := &coredata.BusinessFunction{
		ID:              gid.New(scope.GetTenantID(), coredata.BusinessFunctionEntityType),
		OrganizationID:  req.OrganizationID,
		ReferenceID:     req.ReferenceID,
		Name:            req.Name,
		Classification:  req.Classification,
		MTDMinutes:      req.MTDMinutes,
		RTOMinutes:      req.RTOMinutes,
		RPOMinutes:      req.RPOMinutes,
		ImpactTolerance: req.ImpactTolerance,
		Notes:           req.Notes,
		OwnerID:         req.OwnerID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if req.OwnerID != nil {
				owner := &coredata.MembershipProfile{}
				if err := owner.LoadByID(ctx, conn, scope, *req.OwnerID); err != nil {
					return fmt.Errorf("cannot load owner profile: %w", err)
				}

				if owner.OrganizationID != req.OrganizationID {
					return fmt.Errorf("owner belongs to a different organization: %w", coredata.ErrResourceNotFound)
				}
			}

			for _, assetID := range req.AssetIDs {
				asset := &coredata.Asset{}
				if err := asset.LoadByID(ctx, conn, scope, assetID); err != nil {
					return fmt.Errorf("cannot load asset: %w", err)
				}

				if asset.OrganizationID != req.OrganizationID {
					return fmt.Errorf("asset belongs to a different organization: %w", coredata.ErrResourceNotFound)
				}
			}

			for _, thirdPartyID := range req.ThirdPartyIDs {
				thirdParty := &coredata.ThirdParty{}
				if err := thirdParty.LoadByID(ctx, conn, scope, thirdPartyID); err != nil {
					return fmt.Errorf("cannot load third party: %w", err)
				}

				if thirdParty.OrganizationID != req.OrganizationID {
					return fmt.Errorf("third party belongs to a different organization: %w", coredata.ErrResourceNotFound)
				}
			}

			if err := businessFunction.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot insert business function: %w", err)
			}

			if len(req.AssetIDs) > 0 {
				businessFunctionAssets := coredata.BusinessFunctionAssets{}
				if err := businessFunctionAssets.Insert(ctx, conn, scope, businessFunction.ID, req.OrganizationID, req.AssetIDs); err != nil {
					return fmt.Errorf("cannot insert business function assets: %w", err)
				}
			}

			if len(req.ThirdPartyIDs) > 0 {
				businessFunctionThirdParties := coredata.BusinessFunctionThirdParties{}
				if err := businessFunctionThirdParties.Insert(ctx, conn, scope, businessFunction.ID, req.OrganizationID, req.ThirdPartyIDs); err != nil {
					return fmt.Errorf("cannot insert business function third parties: %w", err)
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return businessFunction, nil
}

func (s *BusinessFunctionService) Update(
	ctx context.Context,
	scope coredata.Scoper,
	req *UpdateBusinessFunctionRequest,
) (*coredata.BusinessFunction, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	businessFunction := &coredata.BusinessFunction{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := businessFunction.LoadByID(ctx, conn, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load business function: %w", err)
			}

			if req.ReferenceID != nil {
				businessFunction.ReferenceID = *req.ReferenceID
			}

			if req.Name != nil {
				businessFunction.Name = *req.Name
			}

			if req.Classification != nil {
				businessFunction.Classification = *req.Classification
			}

			if req.MTDMinutes != nil {
				businessFunction.MTDMinutes = *req.MTDMinutes
			}

			if req.RTOMinutes != nil {
				businessFunction.RTOMinutes = *req.RTOMinutes
			}

			if req.RPOMinutes != nil {
				businessFunction.RPOMinutes = *req.RPOMinutes
			}

			if req.ImpactTolerance != nil {
				businessFunction.ImpactTolerance = *req.ImpactTolerance
			}

			if req.Notes != nil {
				businessFunction.Notes = *req.Notes
			}

			if req.OwnerID != nil {
				if *req.OwnerID != nil {
					owner := &coredata.MembershipProfile{}
					if err := owner.LoadByID(ctx, conn, scope, **req.OwnerID); err != nil {
						return fmt.Errorf("cannot load owner profile: %w", err)
					}

					if owner.OrganizationID != businessFunction.OrganizationID {
						return fmt.Errorf("owner belongs to a different organization: %w", coredata.ErrResourceNotFound)
					}
				}

				businessFunction.OwnerID = *req.OwnerID
			}

			if req.AssetIDs != nil {
				for _, assetID := range *req.AssetIDs {
					asset := &coredata.Asset{}
					if err := asset.LoadByID(ctx, conn, scope, assetID); err != nil {
						return fmt.Errorf("cannot load asset: %w", err)
					}

					if asset.OrganizationID != businessFunction.OrganizationID {
						return fmt.Errorf("asset belongs to a different organization: %w", coredata.ErrResourceNotFound)
					}
				}

				businessFunctionAssets := coredata.BusinessFunctionAssets{}
				if err := businessFunctionAssets.Merge(ctx, conn, scope, businessFunction.ID, businessFunction.OrganizationID, *req.AssetIDs); err != nil {
					return fmt.Errorf("cannot merge business function assets: %w", err)
				}
			}

			if req.ThirdPartyIDs != nil {
				for _, thirdPartyID := range *req.ThirdPartyIDs {
					thirdParty := &coredata.ThirdParty{}
					if err := thirdParty.LoadByID(ctx, conn, scope, thirdPartyID); err != nil {
						return fmt.Errorf("cannot load third party: %w", err)
					}

					if thirdParty.OrganizationID != businessFunction.OrganizationID {
						return fmt.Errorf("third party belongs to a different organization: %w", coredata.ErrResourceNotFound)
					}
				}

				businessFunctionThirdParties := coredata.BusinessFunctionThirdParties{}
				if err := businessFunctionThirdParties.Merge(ctx, conn, scope, businessFunction.ID, businessFunction.OrganizationID, *req.ThirdPartyIDs); err != nil {
					return fmt.Errorf("cannot merge business function third parties: %w", err)
				}
			}

			businessFunction.UpdatedAt = time.Now()

			if err := businessFunction.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update business function: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return businessFunction, nil
}

func (s BusinessFunctionService) Delete(
	ctx context.Context,
	scope coredata.Scoper,
	businessFunctionID gid.GID,
) error {
	businessFunction := coredata.BusinessFunction{ID: businessFunctionID}

	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			err := businessFunction.Delete(ctx, tx, scope)
			if err != nil {
				return fmt.Errorf("cannot delete business function: %w", err)
			}

			return nil
		},
	)
}

func (s BusinessFunctionService) ListForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.BusinessFunctionOrderField],
	filter *coredata.BusinessFunctionFilter,
) (*page.Page[*coredata.BusinessFunction, coredata.BusinessFunctionOrderField], error) {
	var businessFunctions coredata.BusinessFunctions

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := businessFunctions.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor, filter)
			if err != nil {
				return fmt.Errorf("cannot load business functions: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(businessFunctions, cursor), nil
}

func (s BusinessFunctionService) CountForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	filter *coredata.BusinessFunctionFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			businessFunctions := coredata.BusinessFunctions{}

			count, err = businessFunctions.CountByOrganizationID(ctx, conn, scope, organizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count business functions: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s BusinessFunctionService) ListAssets(
	ctx context.Context,
	scope coredata.Scoper,
	businessFunctionID gid.GID,
	cursor *page.Cursor[coredata.AssetOrderField],
) (*page.Page[*coredata.Asset, coredata.AssetOrderField], error) {
	var assets coredata.Assets

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := assets.LoadByBusinessFunctionID(ctx, conn, scope, businessFunctionID, cursor)
			if err != nil {
				return fmt.Errorf("cannot load business function assets: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(assets, cursor), nil
}

func (s BusinessFunctionService) CountAssets(
	ctx context.Context,
	scope coredata.Scoper,
	businessFunctionID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			assets := coredata.Assets{}

			count, err = assets.CountByBusinessFunctionID(ctx, conn, scope, businessFunctionID)
			if err != nil {
				return fmt.Errorf("cannot count business function assets: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s BusinessFunctionService) ListThirdParties(
	ctx context.Context,
	scope coredata.Scoper,
	businessFunctionID gid.GID,
	cursor *page.Cursor[coredata.ThirdPartyOrderField],
) (*page.Page[*coredata.ThirdParty, coredata.ThirdPartyOrderField], error) {
	var thirdParties coredata.ThirdParties

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := thirdParties.LoadByBusinessFunctionID(ctx, conn, scope, businessFunctionID, cursor)
			if err != nil {
				return fmt.Errorf("cannot load business function third parties: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(thirdParties, cursor), nil
}

func (s BusinessFunctionService) CountThirdParties(
	ctx context.Context,
	scope coredata.Scoper,
	businessFunctionID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			thirdParties := coredata.ThirdParties{}

			count, err = thirdParties.CountByBusinessFunctionID(ctx, conn, scope, businessFunctionID)
			if err != nil {
				return fmt.Errorf("cannot count business function third parties: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}
