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

type ComplianceFrameworkOrderBy = OrderBy[coredata.ComplianceFrameworkOrderField]

type ComplianceFramework struct {
	ID          gid.GID                                `json:"id"`
	Framework   *Framework                             `json:"framework"`
	Rank        int                                    `json:"rank"`
	Visibility  coredata.ComplianceFrameworkVisibility `json:"visibility"`
	CreatedAt   time.Time                              `json:"createdAt"`
	UpdatedAt   time.Time                              `json:"updatedAt"`
	FrameworkID gid.GID                                `json:"-"`
}

func (ComplianceFramework) IsNode()          {}
func (c ComplianceFramework) GetID() gid.GID { return c.ID }

type ComplianceFrameworkConnection struct {
	Edges    []*ComplianceFrameworkEdge `json:"edges"`
	PageInfo *PageInfo                  `json:"pageInfo"`
}

func NewComplianceFramework(cf *coredata.ComplianceFramework) *ComplianceFramework {
	return &ComplianceFramework{
		ID:          cf.ID,
		FrameworkID: cf.FrameworkID,
		Rank:        cf.Rank,
		Visibility:  cf.Visibility,
		CreatedAt:   cf.CreatedAt,
		UpdatedAt:   cf.UpdatedAt,
	}
}

func NewComplianceFrameworkConnection(
	p *page.Page[*coredata.ComplianceFramework, coredata.ComplianceFrameworkOrderField],
) *ComplianceFrameworkConnection {
	edges := make([]*ComplianceFrameworkEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewComplianceFrameworkEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &ComplianceFrameworkConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
	}
}

func NewComplianceFrameworkEdge(cf *coredata.ComplianceFramework, orderBy coredata.ComplianceFrameworkOrderField) *ComplianceFrameworkEdge {
	return &ComplianceFrameworkEdge{
		Cursor: cf.CursorKey(orderBy),
		Node:   NewComplianceFramework(cf),
	}
}
