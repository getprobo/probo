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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	ContinualImprovementOrderBy OrderBy[coredata.ContinualImprovementOrderField]

	ContinualImprovementConnection struct {
		TotalCount int
		Edges      []*ContinualImprovementEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filter   *ContinualImprovementFilter
	}
)

func NewContinualImprovementConnection(
	p *page.Page[*coredata.ContinualImprovement, coredata.ContinualImprovementOrderField],
	parentType any,
	parentID gid.GID,
	filter *ContinualImprovementFilter,
) *ContinualImprovementConnection {
	edges := make([]*ContinualImprovementEdge, len(p.Data))
	for i, improvement := range p.Data {
		edges[i] = NewContinualImprovementEdge(improvement, p.Cursor.OrderBy.Field)
	}

	return &ContinualImprovementConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
		Filter:   filter,
	}
}

func NewContinualImprovementEdge(ci *coredata.ContinualImprovement, orderField coredata.ContinualImprovementOrderField) *ContinualImprovementEdge {
	return &ContinualImprovementEdge{
		Node:   NewContinualImprovement(ci),
		Cursor: ci.CursorKey(orderField),
	}
}

func NewContinualImprovement(ci *coredata.ContinualImprovement) *ContinualImprovement {
	return &ContinualImprovement{
		ID:         ci.ID,
		SnapshotID: ci.SnapshotID,
		Organization: &Organization{
			ID: ci.OrganizationID,
		},
		Owner: &Profile{
			ID: ci.OwnerID,
		},
		SourceID:    ci.SourceID,
		ReferenceID: ci.ReferenceID,
		Description: ci.Description,
		Source:      ci.Source,
		TargetDate:  ci.TargetDate,
		Status:      ci.Status,
		Priority:    ci.Priority,
		CreatedAt:   ci.CreatedAt,
		UpdatedAt:   ci.UpdatedAt,
	}
}
