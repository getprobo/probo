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

type ComplianceFramework struct {
	ID          gid.GID    `json:"id"`
	Framework   *Framework `json:"framework"`
	FrameworkID gid.GID    `json:"-"`
}

func (ComplianceFramework) IsNode()             {}
func (this ComplianceFramework) GetID() gid.GID { return this.ID }

type ComplianceFrameworkConnection struct {
	Edges    []*ComplianceFrameworkEdge `json:"edges"`
	PageInfo *PageInfo                  `json:"pageInfo"`
}

type ComplianceFrameworkEdge struct {
	Cursor page.CursorKey       `json:"cursor"`
	Node   *ComplianceFramework `json:"node"`
}

func NewComplianceFramework(cf *coredata.ComplianceFramework) *ComplianceFramework {
	return &ComplianceFramework{
		ID:          cf.ID,
		FrameworkID: cf.FrameworkID,
	}
}

func NewComplianceFrameworkEdge(cf *coredata.ComplianceFramework) *ComplianceFrameworkEdge {
	return &ComplianceFrameworkEdge{
		Cursor: cf.CursorKey(coredata.ComplianceFrameworkOrderFieldRank),
		Node:   NewComplianceFramework(cf),
	}
}

func NewComplianceFrameworkConnection(
	p *page.Page[*coredata.ComplianceFramework, coredata.ComplianceFrameworkOrderField],
) *ComplianceFrameworkConnection {
	edges := make([]*ComplianceFrameworkEdge, len(p.Data))
	for i, cf := range p.Data {
		edges[i] = NewComplianceFrameworkEdge(cf)
	}

	return &ComplianceFrameworkConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
	}
}
