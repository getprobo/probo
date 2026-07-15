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
	"go.probo.inc/probo/pkg/page"
)

func NewCompliancePortalCommitmentGroup(g *coredata.CompliancePortalCommitmentGroup) *CompliancePortalCommitmentGroup {
	return &CompliancePortalCommitmentGroup{
		ID:          g.ID,
		Title:       g.Title,
		Description: g.Description,
	}
}

func NewCompliancePortalCommitmentGroupConnection(
	p *page.Page[*coredata.CompliancePortalCommitmentGroup, coredata.CompliancePortalCommitmentGroupOrderField],
) *CompliancePortalCommitmentGroupConnection {
	edges := make([]*CompliancePortalCommitmentGroupEdge, len(p.Data))

	for i, item := range p.Data {
		edges[i] = NewCompliancePortalCommitmentGroupEdge(item, p.Cursor.OrderBy.Field)
	}

	return &CompliancePortalCommitmentGroupConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
	}
}

func NewCompliancePortalCommitmentGroupEdge(
	g *coredata.CompliancePortalCommitmentGroup,
	orderBy coredata.CompliancePortalCommitmentGroupOrderField,
) *CompliancePortalCommitmentGroupEdge {
	return &CompliancePortalCommitmentGroupEdge{
		Cursor: g.CursorKey(orderBy),
		Node:   NewCompliancePortalCommitmentGroup(g),
	}
}

func NewCompliancePortalCommitment(c *coredata.CompliancePortalCommitment) *CompliancePortalCommitment {
	return &CompliancePortalCommitment{
		ID:          c.ID,
		Icon:        c.Icon,
		Eyebrow:     c.Eyebrow,
		Title:       c.Title,
		Description: c.Description,
	}
}

func NewCompliancePortalCommitmentConnection(
	p *page.Page[*coredata.CompliancePortalCommitment, coredata.CompliancePortalCommitmentOrderField],
) *CompliancePortalCommitmentConnection {
	edges := make([]*CompliancePortalCommitmentEdge, len(p.Data))

	for i, item := range p.Data {
		edges[i] = NewCompliancePortalCommitmentEdge(item, p.Cursor.OrderBy.Field)
	}

	return &CompliancePortalCommitmentConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
	}
}

func NewCompliancePortalCommitmentEdge(
	c *coredata.CompliancePortalCommitment,
	orderBy coredata.CompliancePortalCommitmentOrderField,
) *CompliancePortalCommitmentEdge {
	return &CompliancePortalCommitmentEdge{
		Cursor: c.CursorKey(orderBy),
		Node:   NewCompliancePortalCommitment(c),
	}
}
