// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
