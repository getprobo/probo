package probo

import (
	"context"
	"fmt"
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.gearno.de/kit/pg"
)

type AssetService struct {
	svc *TenantService
}

type CreateAssetRequest struct {
	OrganizationID  gid.GID
	Name            string
	Amount          int
	OwnerID         gid.GID
	AssetType       coredata.AssetType
	DataTypesStored string
	VendorIDs       []gid.GID
}

type UpdateAssetRequest struct {
	ID              gid.GID
	Name            *string
	Amount          *int
	OwnerID         *gid.GID
	AssetType       *coredata.AssetType
	DataTypesStored *string
	VendorIDs       []gid.GID
}

func (s AssetService) Get(
	ctx context.Context,
	assetID gid.GID,
) (*coredata.Asset, error) {
	asset := &coredata.Asset{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return asset.LoadByID(ctx, conn, s.svc.scope, assetID)
		},
	)

	if err != nil {
		return nil, err
	}

	return asset, nil
}

func (s AssetService) GetByOwnerID(
	ctx context.Context,
	ownerID gid.GID,
) (*coredata.Asset, error) {
	asset := &coredata.Asset{OwnerID: ownerID}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return asset.LoadByOwnerID(ctx, conn, s.svc.scope)
		},
	)

	if err != nil {
		return nil, err
	}

	return asset, nil
}

func (s AssetService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	filter *coredata.AssetFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			assets := coredata.Assets{}
			count, err = assets.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID, filter)
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

func (s AssetService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.AssetOrderField],
	filter *coredata.AssetFilter,
) (*page.Page[*coredata.Asset, coredata.AssetOrderField], error) {
	var assets coredata.Assets

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return assets.LoadByOrganizationID(
				ctx,
				conn,
				s.svc.scope,
				organizationID,
				cursor,
				filter,
			)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(assets, cursor), nil
}

func (s AssetService) Update(
	ctx context.Context,
	req UpdateAssetRequest,
) (*coredata.Asset, error) {
	now := time.Now()
	asset := &coredata.Asset{ID: req.ID}
	assetVendors := &coredata.AssetVendors{}

	err := s.svc.pg.WithTx(ctx, func(conn pg.Conn) error {
		if err := asset.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
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
			asset.OwnerID = *req.OwnerID
		}
		if req.AssetType != nil {
			asset.AssetType = *req.AssetType
		}
		if req.DataTypesStored != nil {
			asset.DataTypesStored = *req.DataTypesStored
		}

		if err := asset.Update(ctx, conn, s.svc.scope); err != nil {
			return fmt.Errorf("cannot update asset: %w", err)
		}

		if req.VendorIDs != nil {
			if err := assetVendors.Merge(ctx, conn, s.svc.scope, asset.ID, req.VendorIDs); err != nil {
				return fmt.Errorf("cannot update asset vendors: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return asset, nil
}

func (s AssetService) Create(
	ctx context.Context,
	req CreateAssetRequest,
) (*coredata.Asset, error) {
	now := time.Now()
	assetID := gid.New(s.svc.scope.GetTenantID(), coredata.AssetEntityType)
	assetVendors := &coredata.AssetVendors{}

	asset := &coredata.Asset{
		ID:              assetID,
		OrganizationID:  req.OrganizationID,
		Name:            req.Name,
		Amount:          req.Amount,
		OwnerID:         req.OwnerID,
		AssetType:       req.AssetType,
		DataTypesStored: req.DataTypesStored,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	err := s.svc.pg.WithTx(ctx, func(conn pg.Conn) error {
		if err := asset.Insert(ctx, conn, s.svc.scope); err != nil {
			return fmt.Errorf("cannot insert asset: %w", err)
		}

		if len(req.VendorIDs) > 0 {
			if err := assetVendors.Insert(ctx, conn, s.svc.scope, asset.ID, req.VendorIDs); err != nil {
				return fmt.Errorf("cannot create asset vendors: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return asset, nil
}

func (s AssetService) Delete(
	ctx context.Context,
	assetID gid.GID,
) error {
	asset := &coredata.Asset{ID: assetID}

	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return asset.Delete(ctx, conn, s.svc.scope)
		},
	)
}
