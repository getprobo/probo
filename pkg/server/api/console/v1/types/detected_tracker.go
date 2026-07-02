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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	DetectedTrackerOrderBy OrderBy[coredata.DetectedTrackerOrderField]

	DetectedTrackerConnection struct {
		TotalCount int
		Edges      []*DetectedTrackerEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewDetectedTrackerConnection(
	p *page.Page[*coredata.DetectedTracker, coredata.DetectedTrackerOrderField],
	parentType any,
	parentID gid.GID,
) *DetectedTrackerConnection {
	edges := make([]*DetectedTrackerEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewDetectedTrackerEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &DetectedTrackerConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewDetectedTrackerEdge(
	dt *coredata.DetectedTracker,
	orderBy coredata.DetectedTrackerOrderField,
) *DetectedTrackerEdge {
	return &DetectedTrackerEdge{
		Cursor: dt.CursorKey(orderBy),
		Node:   NewDetectedTrackerNode(dt),
	}
}

func NewDetectedTrackerNode(dt *coredata.DetectedTracker) *DetectedTracker {
	return &DetectedTracker{
		ID:             dt.ID,
		Identifier:     dt.Identifier,
		InitiatorURL:   dt.InitiatorURL,
		MaxAgeSeconds:  dt.MaxAgeSeconds,
		Source:         dt.Source,
		LastDetectedAt: dt.LastDetectedAt,
		CreatedAt:      dt.CreatedAt,
	}
}
