package types

import (
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	AssetOrderBy OrderBy[coredata.AssetOrderField]

	AssetConnection struct {
		TotalCount int
		Edges      []*AssetEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filter   *AssetFilter
	}
)

func NewAssetConnection(
	p *page.Page[*coredata.Asset, coredata.AssetOrderField],
	resolver any,
	parentID gid.GID,
	filter *AssetFilter,
) *AssetConnection {
	edges := make([]*AssetEdge, len(p.Data))
	for i, asset := range p.Data {
		edges[i] = NewAssetEdge(asset, p.Cursor.OrderBy.Field)
	}

	return &AssetConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: resolver,
		ParentID: parentID,
		Filter:   filter,
	}
}

func NewAssetEdge(asset *coredata.Asset, orderField coredata.AssetOrderField) *AssetEdge {
	return &AssetEdge{
		Node:   NewAsset(asset),
		Cursor: asset.CursorKey(orderField),
	}
}

func NewAsset(asset *coredata.Asset) *Asset {
	return &Asset{
		ID:         asset.ID,
		SnapshotID: asset.SnapshotID,
		Name:       asset.Name,
		Amount:     asset.Amount,
		Owner: &Profile{
			ID: asset.OwnerID,
		},
		AssetType:       asset.AssetType,
		DataTypesStored: asset.DataTypesStored,
		CreatedAt:       asset.CreatedAt,
		UpdatedAt:       asset.UpdatedAt,
	}
}
