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
	EmployeePortalOrderBy OrderBy[coredata.EmployeePortalOrderField]

	EmployeePortalConnection struct {
		TotalCount int
		Edges      []*EmployeePortalEdge
		PageInfo   *PageInfo

		ParentID gid.GID
	}
)

func NewEmployeePortal(p *coredata.EmployeePortal) *EmployeePortal {
	return &EmployeePortal{
		ID:           p.ID,
		Name:         p.Name,
		Capabilities: &p.Capabilities,
		Organization: &Organization{
			ID: p.OrganizationID,
		},
	}
}

func NewEmployeePortalConnection(
	p *page.Page[*coredata.EmployeePortal, coredata.EmployeePortalOrderField],
	parentID gid.GID,
) *EmployeePortalConnection {
	edges := make([]*EmployeePortalEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewEmployeePortalEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &EmployeePortalConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
		ParentID: parentID,
	}
}

func NewEmployeePortalEdge(
	p *coredata.EmployeePortal,
	orderBy coredata.EmployeePortalOrderField,
) *EmployeePortalEdge {
	return &EmployeePortalEdge{
		Cursor: p.CursorKey(orderBy),
		Node:   NewEmployeePortal(p),
	}
}
