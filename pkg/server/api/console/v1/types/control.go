// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	ControlOrderBy OrderBy[coredata.ControlOrderField]

	ControlConnection struct {
		TotalCount int
		Edges      []*ControlEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filters  *coredata.ControlFilter
	}
)

func NewControlConnection(
	p *page.Page[*coredata.Control, coredata.ControlOrderField],
	resolver any,
	parentID gid.GID,
	filter *coredata.ControlFilter,
) *ControlConnection {
	edges := make([]*ControlEdge, len(p.Data))
	for i, control := range p.Data {
		edges[i] = NewControlEdge(control, p.Cursor.OrderBy.Field)
	}

	return &ControlConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: resolver,
		ParentID: parentID,
		Filters:  filter,
	}
}

func NewControlEdge(control *coredata.Control, orderField coredata.ControlOrderField) *ControlEdge {
	return &ControlEdge{
		Node:   NewControl(control),
		Cursor: control.CursorKey(orderField),
	}
}

func NewControl(control *coredata.Control) *Control {
	return &Control{
		ID: control.ID,
		Organization: &Organization{
			ID: control.OrganizationID,
		},
		Framework: &Framework{
			ID: control.FrameworkID,
		},
		SectionTitle:                control.SectionTitle,
		Name:                        control.Name,
		Description:                 control.Description,
		BestPractice:                control.BestPractice,
		NotImplementedJustification: control.NotImplementedJustification,
		MaturityLevel:               control.MaturityLevel,
		CreatedAt:                   control.CreatedAt,
		UpdatedAt:                   control.UpdatedAt,
	}
}
