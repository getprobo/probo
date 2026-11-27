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
	"encoding/json"
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	AssetChangeEvent struct {
		ID          gid.GID                  `json:"id"`
		ObservedAt  time.Time                `json:"observedAt"`
		AssetSyncID *gid.GID                 `json:"assetSyncId,omitempty"`
		Actor       *Profile                 `json:"actor,omitempty"`
		ChangeType  coredata.AssetChangeType `json:"changeType"`
		Before      *string                  `json:"before,omitempty"`
		After       *string                  `json:"after,omitempty"`
	}

	AssetChangeEventOrderBy OrderBy[coredata.AssetChangeEventOrderField]

	AssetChangeEventConnection struct {
		TotalCount int
		Edges      []*AssetChangeEventEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}

	AssetChangeEventEdge struct {
		Cursor page.CursorKey    `json:"cursor"`
		Node   *AssetChangeEvent `json:"node"`
	}
)

func (AssetChangeEvent) IsNode()          {}
func (e AssetChangeEvent) GetID() gid.GID { return e.ID }

func NewAssetChangeEventConnection(
	p *page.Page[*coredata.AssetChangeEvent, coredata.AssetChangeEventOrderField],
	parentType any,
	parentID gid.GID,
) *AssetChangeEventConnection {
	edges := make([]*AssetChangeEventEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewAssetChangeEventEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &AssetChangeEventConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewAssetChangeEventEdge(
	event *coredata.AssetChangeEvent,
	orderBy coredata.AssetChangeEventOrderField,
) *AssetChangeEventEdge {
	return &AssetChangeEventEdge{
		Cursor: event.CursorKey(orderBy),
		Node:   NewAssetChangeEvent(event),
	}
}

func NewAssetChangeEvent(event *coredata.AssetChangeEvent) *AssetChangeEvent {
	node := &AssetChangeEvent{
		ID:          event.ID,
		ObservedAt:  event.ObservedAt,
		AssetSyncID: event.AssetSyncID,
		ChangeType:  event.ChangeType,
		Before:      rawJSONText(event.Before),
		After:       rawJSONText(event.After),
	}

	if event.ActorID != nil {
		node.Actor = &Profile{ID: *event.ActorID}
	}

	return node
}

func rawJSONText(raw json.RawMessage) *string {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}

	value := string(raw)

	return &value
}
