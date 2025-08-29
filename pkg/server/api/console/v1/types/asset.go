package types

import (
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
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

func NewAsset(asset *coredata.Asset) *Asset {
	return &Asset{
		ID:              asset.ID,
		SnapshotID:      asset.SnapshotID,
		Name:            asset.Name,
		Amount:          asset.Amount,
		Criticity:       asset.Criticity,
		AssetType:       asset.AssetType,
		DataTypesStored: asset.DataTypesStored,
		CreatedAt:       asset.CreatedAt,
		UpdatedAt:       asset.UpdatedAt,
	}
}

func NewAssetEdge(asset *coredata.Asset, orderField coredata.AssetOrderField) *AssetEdge {
	return &AssetEdge{
		Node:   NewAsset(asset),
		Cursor: asset.CursorKey(orderField),
	}
}
