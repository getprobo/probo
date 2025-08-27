// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
)

type (
	SnapshotOrderBy OrderBy[coredata.SnapshotOrderField]

	SnapshotConnection struct {
		TotalCount int
		Edges      []*SnapshotEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewSnapshotConnection(
	p *page.Page[*coredata.Snapshot, coredata.SnapshotOrderField],
	parentType any,
	parentID gid.GID,
) *SnapshotConnection {
	edges := make([]*SnapshotEdge, len(p.Data))
	for i, snapshot := range p.Data {
		edges[i] = NewSnapshotEdge(snapshot, p.Cursor.OrderBy.Field)
	}

	return &SnapshotConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewSnapshot(s *coredata.Snapshot) *Snapshot {
	return &Snapshot{
		ID:          s.ID,
		Name:        s.Name,
		Type:        s.Type,
		Description: s.Description,
		CreatedAt:   s.CreatedAt,
	}
}

func NewSnapshotEdge(s *coredata.Snapshot, orderField coredata.SnapshotOrderField) *SnapshotEdge {
	return &SnapshotEdge{
		Node:   NewSnapshot(s),
		Cursor: s.CursorKey(orderField),
	}
}
