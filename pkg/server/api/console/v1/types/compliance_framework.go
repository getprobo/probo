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
