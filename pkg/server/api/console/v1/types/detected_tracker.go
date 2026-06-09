// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
