package types

import (
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type Asset struct {
	ID              gid.GID            `json:"id"`
	OrganizationID  gid.GID            `json:"-"`
	SnapshotID      *gid.GID           `json:"snapshotId,omitempty"`
	Name            string             `json:"name"`
	Amount          int                `json:"amount"`
	Owner           *People            `json:"owner"`
	Vendors         *VendorConnection  `json:"vendors"`
	AssetType       coredata.AssetType `json:"assetType"`
	DataTypesStored string             `json:"dataTypesStored"`
	Organization    *Organization      `json:"organization"`
	CreatedAt       time.Time          `json:"createdAt"`
	UpdatedAt       time.Time          `json:"updatedAt"`
}

func (Asset) IsNode()             {}
func (this Asset) GetID() gid.GID { return this.ID }

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
		OrganizationID:  asset.OrganizationID,
		SnapshotID:      asset.SnapshotID,
		Name:            asset.Name,
		Amount:          asset.Amount,
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
