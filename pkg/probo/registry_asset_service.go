// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type RegistryAssetService struct {
	svc *Service
}

type CreateRegistryAssetRequest struct {
	OrganizationID    gid.GID
	Name              string
	Amount            int
	OwnerID           gid.GID
	RegistryAssetType coredata.RegistryAssetType
	DataTypesStored   string
	ThirdPartyIDs     []gid.GID
}

type UpdateRegistryAssetRequest struct {
	ID                gid.GID
	Name              *string
	Amount            *int
	OwnerID           *gid.GID
	RegistryAssetType *coredata.RegistryAssetType
	DataTypesStored   *string
	ThirdPartyIDs     []gid.GID
}

func (car *CreateRegistryAssetRequest) Validate() error {
	v := validator.New()

	v.Check(car.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(car.Name, "name", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(car.Amount, "amount", validator.Required(), validator.Min(1))
	v.Check(car.OwnerID, "owner_id", validator.Required(), validator.GID(coredata.MembershipProfileEntityType))
	v.Check(car.RegistryAssetType, "registry_asset_type", validator.Required(), validator.OneOfSlice(coredata.RegistryAssetTypes()))
	v.Check(car.DataTypesStored, "data_types_stored", validator.Required(), validator.SafeText(ContentMaxLength))
	v.CheckEach(car.ThirdPartyIDs, "third_party_ids", func(index int, item any) {
		v.Check(item, fmt.Sprintf("third_party_ids[%d]", index), validator.Required(), validator.GID(coredata.ThirdPartyEntityType))
	})

	return v.Error()
}

func (uar *UpdateRegistryAssetRequest) Validate() error {
	v := validator.New()

	v.Check(uar.ID, "id", validator.Required(), validator.GID(coredata.RegistryAssetEntityType))
	v.Check(uar.Name, "name", validator.SafeTextNoNewLine(NameMaxLength))
	v.Check(uar.Amount, "amount", validator.Min(1))
	v.Check(uar.OwnerID, "owner_id", validator.GID(coredata.MembershipProfileEntityType))
	v.Check(uar.RegistryAssetType, "registry_asset_type", validator.OneOfSlice(coredata.RegistryAssetTypes()))
	v.Check(uar.DataTypesStored, "data_types_stored", validator.SafeText(ContentMaxLength))
	v.CheckEach(uar.ThirdPartyIDs, "third_party_ids", func(index int, item any) {
		v.Check(item, fmt.Sprintf("third_party_ids[%d]", index), validator.GID(coredata.ThirdPartyEntityType))
	})

	return v.Error()
}

func (s RegistryAssetService) Get(
	ctx context.Context, scope coredata.Scoper,
	assetID gid.GID,
) (*coredata.RegistryAsset, error) {
	asset := &coredata.RegistryAsset{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return asset.LoadByID(ctx, conn, scope, assetID)
		},
	)
	if err != nil {
		return nil, err
	}

	return asset, nil
}

func (s RegistryAssetService) GetByIDs(
	ctx context.Context,
	scope coredata.Scoper,
	assetIDs ...gid.GID,
) (coredata.RegistryAssets, error) {
	var assets coredata.RegistryAssets

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := assets.LoadByIDs(ctx, conn, scope, assetIDs); err != nil &&
				!errors.Is(err, coredata.ErrResourceNotFound) {
				return fmt.Errorf("cannot load registry assets by ids: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return assets, nil
}

func (s RegistryAssetService) GetByOwnerID(
	ctx context.Context, scope coredata.Scoper,
	ownerID gid.GID,
) (*coredata.RegistryAsset, error) {
	asset := &coredata.RegistryAsset{OwnerID: ownerID}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return asset.LoadByOwnerID(ctx, conn, scope)
		},
	)
	if err != nil {
		return nil, err
	}

	return asset, nil
}

func (s RegistryAssetService) CountForOrganizationID(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			assets := coredata.RegistryAssets{}

			count, err = assets.CountByOrganizationID(ctx, conn, scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot count assets: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s RegistryAssetService) ListForOrganizationID(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.RegistryAssetOrderField],
) (*page.Page[*coredata.RegistryAsset, coredata.RegistryAssetOrderField], error) {
	var assets coredata.RegistryAssets

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return assets.LoadByOrganizationID(
				ctx,
				conn,
				scope,
				organizationID,
				cursor,
			)
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(assets, cursor), nil
}

func (s RegistryAssetService) Update(
	ctx context.Context, scope coredata.Scoper,
	req UpdateRegistryAssetRequest,
) (*coredata.RegistryAsset, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()
	asset := &coredata.RegistryAsset{ID: req.ID}
	assetThirdParties := &coredata.AssetThirdParties{}

	err := s.svc.pg.WithTx(ctx, func(ctx context.Context, conn pg.Tx) error {
		if err := asset.LoadByID(ctx, conn, scope, req.ID); err != nil {
			return fmt.Errorf("cannot load asset: %w", err)
		}

		asset.UpdatedAt = now
		if req.Name != nil {
			asset.Name = *req.Name
		}

		if req.Amount != nil {
			asset.Amount = *req.Amount
		}

		if req.OwnerID != nil {
			profile := &coredata.MembershipProfile{}
			if err := profile.LoadByID(ctx, conn, scope, *req.OwnerID); err != nil {
				return fmt.Errorf("cannot load owner profile: %w", err)
			}

			asset.OwnerID = *req.OwnerID
		}

		if req.RegistryAssetType != nil {
			asset.RegistryAssetType = *req.RegistryAssetType
		}

		if req.DataTypesStored != nil {
			asset.DataTypesStored = *req.DataTypesStored
		}

		if err := asset.Update(ctx, conn, scope); err != nil {
			return fmt.Errorf("cannot update asset: %w", err)
		}

		if err := coredata.UpsertSearchResource(
			ctx,
			conn,
			asset.ID,
			asset.OrganizationID,
			asset.Name,
			asset.SearchText(),
		); err != nil {
			return fmt.Errorf("cannot sync search: %w", err)
		}

		if req.ThirdPartyIDs != nil {
			if len(req.ThirdPartyIDs) > 0 {
				thirdParties := &coredata.ThirdParties{}
				if err := thirdParties.LoadByIDs(ctx, conn, scope, req.ThirdPartyIDs); err != nil {
					return fmt.Errorf("cannot load thirdParties: %w", err)
				}
			}

			if err := assetThirdParties.Merge(ctx, conn, scope, asset.ID, asset.OrganizationID, req.ThirdPartyIDs); err != nil {
				return fmt.Errorf("cannot update asset thirdParties: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return asset, nil
}

func (s RegistryAssetService) Create(
	ctx context.Context, scope coredata.Scoper,
	req CreateRegistryAssetRequest,
) (*coredata.RegistryAsset, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()
	assetID := gid.New(scope.GetTenantID(), coredata.RegistryAssetEntityType)
	assetThirdParties := &coredata.AssetThirdParties{}

	asset := &coredata.RegistryAsset{
		ID:                assetID,
		OrganizationID:    req.OrganizationID,
		Name:              req.Name,
		Amount:            req.Amount,
		OwnerID:           req.OwnerID,
		RegistryAssetType: req.RegistryAssetType,
		DataTypesStored:   req.DataTypesStored,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	err := s.svc.pg.WithTx(ctx, func(ctx context.Context, conn pg.Tx) error {
		profile := &coredata.MembershipProfile{}
		if err := profile.LoadByID(ctx, conn, scope, req.OwnerID); err != nil {
			return fmt.Errorf("cannot load owner profile: %w", err)
		}

		if err := asset.Insert(ctx, conn, scope); err != nil {
			return fmt.Errorf("cannot insert asset: %w", err)
		}

		if err := coredata.UpsertSearchResource(
			ctx,
			conn,
			asset.ID,
			asset.OrganizationID,
			asset.Name,
			asset.SearchText(),
		); err != nil {
			return fmt.Errorf("cannot sync search: %w", err)
		}

		if len(req.ThirdPartyIDs) > 0 {
			thirdParties := &coredata.ThirdParties{}
			if err := thirdParties.LoadByIDs(ctx, conn, scope, req.ThirdPartyIDs); err != nil {
				return fmt.Errorf("cannot load thirdParties: %w", err)
			}

			if err := assetThirdParties.Insert(ctx, conn, scope, asset.ID, asset.OrganizationID, req.ThirdPartyIDs); err != nil {
				return fmt.Errorf("cannot create asset thirdParties: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return asset, nil
}

func (s RegistryAssetService) Delete(
	ctx context.Context, scope coredata.Scoper,
	assetID gid.GID,
) error {
	asset := &coredata.RegistryAsset{ID: assetID}

	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := coredata.DeleteSearchResource(ctx, tx, asset.ID); err != nil {
				return fmt.Errorf("cannot sync search: %w", err)
			}

			return asset.Delete(ctx, tx, scope)
		},
	)
}
