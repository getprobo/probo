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

package types

import (
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	AssetSnapshot struct {
		ID         gid.GID   `json:"id"`
		CapturedAt time.Time `json:"capturedAt"`
		Actor      *Profile  `json:"actor,omitempty"`
		Name       *string   `json:"name,omitempty"`
		Note       *string   `json:"note,omitempty"`
	}

	AssetSnapshotOrderBy OrderBy[coredata.AssetSnapshotOrderField]

	AssetSnapshotConnection struct {
		TotalCount int
		Edges      []*AssetSnapshotEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}

	AssetSnapshotEdge struct {
		Cursor page.CursorKey `json:"cursor"`
		Node   *AssetSnapshot `json:"node"`
	}

	AssetSnapshotAsset struct {
		ID                  gid.GID              `json:"id"`
		AssetID             gid.GID              `json:"assetId"`
		Name                string               `json:"name"`
		Amount              int                  `json:"amount"`
		Owner               *Profile             `json:"owner,omitempty"`
		AssetType           coredata.AssetType   `json:"assetType"`
		DataTypesStored     *string              `json:"dataTypesStored,omitempty"`
		Kind                *string              `json:"kind,omitempty"`
		ExternalID          *string              `json:"externalId,omitempty"`
		Region              *string              `json:"region,omitempty"`
		Configuration       *string              `json:"configuration,omitempty"`
		SchemaVersion       *int                 `json:"schemaVersion,omitempty"`
		LastSeenAt          *time.Time           `json:"lastSeenAt,omitempty"`
		Status              coredata.AssetStatus `json:"status"`
		ConnectorID         *gid.GID             `json:"connectorId,omitempty"`
		ThirdPartyIDs       []gid.GID            `json:"thirdPartyIds"`
		BusinessFunctionIDs []gid.GID            `json:"businessFunctionIds"`
		TagIDs              []gid.GID            `json:"tagIds"`
		AssetCreatedAt      time.Time            `json:"assetCreatedAt"`
		AssetUpdatedAt      time.Time            `json:"assetUpdatedAt"`
	}

	AssetSnapshotAssetOrderBy OrderBy[coredata.AssetSnapshotAssetOrderField]

	AssetSnapshotAssetConnection struct {
		TotalCount int
		Edges      []*AssetSnapshotAssetEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}

	AssetSnapshotAssetEdge struct {
		Cursor page.CursorKey      `json:"cursor"`
		Node   *AssetSnapshotAsset `json:"node"`
	}
)

func (AssetSnapshot) IsNode()          {}
func (s AssetSnapshot) GetID() gid.GID { return s.ID }

func NewAssetSnapshotConnection(
	p *page.Page[*coredata.AssetSnapshot, coredata.AssetSnapshotOrderField],
	parentType any,
	parentID gid.GID,
) *AssetSnapshotConnection {
	edges := make([]*AssetSnapshotEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewAssetSnapshotEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &AssetSnapshotConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewAssetSnapshotEdge(
	snapshot *coredata.AssetSnapshot,
	orderBy coredata.AssetSnapshotOrderField,
) *AssetSnapshotEdge {
	return &AssetSnapshotEdge{
		Cursor: snapshot.CursorKey(orderBy),
		Node:   NewAssetSnapshot(snapshot),
	}
}

func NewAssetSnapshot(snapshot *coredata.AssetSnapshot) *AssetSnapshot {
	node := &AssetSnapshot{
		ID:         snapshot.ID,
		CapturedAt: snapshot.CapturedAt,
		Name:       snapshot.Name,
		Note:       snapshot.Note,
	}

	if snapshot.ActorID != nil {
		node.Actor = &Profile{ID: *snapshot.ActorID}
	}

	return node
}

func NewAssetSnapshotAssetConnection(
	p *page.Page[*coredata.AssetSnapshotAsset, coredata.AssetSnapshotAssetOrderField],
	parentType any,
	parentID gid.GID,
) *AssetSnapshotAssetConnection {
	edges := make([]*AssetSnapshotAssetEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewAssetSnapshotAssetEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &AssetSnapshotAssetConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewAssetSnapshotAssetEdge(
	asset *coredata.AssetSnapshotAsset,
	orderBy coredata.AssetSnapshotAssetOrderField,
) *AssetSnapshotAssetEdge {
	return &AssetSnapshotAssetEdge{
		Cursor: asset.CursorKey(orderBy),
		Node:   NewAssetSnapshotAsset(asset),
	}
}

func NewAssetSnapshotAsset(asset *coredata.AssetSnapshotAsset) *AssetSnapshotAsset {
	node := &AssetSnapshotAsset{
		ID:                  asset.ID,
		AssetID:             asset.AssetID,
		Name:                asset.Name,
		Amount:              asset.Amount,
		AssetType:           asset.AssetType,
		DataTypesStored:     asset.DataTypesStored,
		Kind:                asset.Kind,
		ExternalID:          asset.ExternalID,
		Region:              asset.Region,
		Configuration:       rawJSONText(asset.Configuration),
		SchemaVersion:       asset.SchemaVersion,
		LastSeenAt:          asset.LastSeenAt,
		Status:              asset.Status,
		ConnectorID:         asset.ConnectorID,
		ThirdPartyIDs:       mustStoredGIDs(asset.ThirdPartyIDs),
		BusinessFunctionIDs: mustStoredGIDs(asset.BusinessFunctionIDs),
		TagIDs:              mustStoredGIDs(asset.TagIDs),
		AssetCreatedAt:      asset.AssetCreatedAt,
		AssetUpdatedAt:      asset.AssetUpdatedAt,
	}

	if asset.OwnerID != nil {
		node.Owner = &Profile{ID: *asset.OwnerID}
	}

	if node.ThirdPartyIDs == nil {
		node.ThirdPartyIDs = []gid.GID{}
	}

	if node.BusinessFunctionIDs == nil {
		node.BusinessFunctionIDs = []gid.GID{}
	}

	if node.TagIDs == nil {
		node.TagIDs = []gid.GID{}
	}

	return node
}

func mustStoredGIDs(values []string) []gid.GID {
	ids := make([]gid.GID, len(values))
	for i, value := range values {
		ids[i] = gid.MustParseGID(value)
	}

	return ids
}
