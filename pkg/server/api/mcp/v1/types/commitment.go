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

func NewCommitmentGroup(g *coredata.CompliancePortalCommitmentGroup) *CommitmentGroup {
	return &CommitmentGroup{
		ID:          g.ID,
		Title:       g.Title,
		Description: g.Description,
		Rank:        g.Rank,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
	}
}

func NewListCommitmentGroupsOutput(
	p *page.Page[*coredata.CompliancePortalCommitmentGroup, coredata.CompliancePortalCommitmentGroupOrderField],
) ListCommitmentGroupsOutput {
	groups := make([]*CommitmentGroup, 0, len(p.Data))
	for _, g := range p.Data {
		groups = append(groups, NewCommitmentGroup(g))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListCommitmentGroupsOutput{
		NextCursor:       nextCursor,
		CommitmentGroups: groups,
	}
}

func NewCommitment(c *coredata.CompliancePortalCommitment) *Commitment {
	return &Commitment{
		ID:          c.ID,
		GroupID:     c.GroupID,
		Icon:        c.Icon,
		Eyebrow:     c.Eyebrow,
		Title:       c.Title,
		Description: c.Description,
		Rank:        c.Rank,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func NewListCommitmentsOutput(
	p *page.Page[*coredata.CompliancePortalCommitment, coredata.CompliancePortalCommitmentOrderField],
) ListCommitmentsOutput {
	commitments := make([]*Commitment, 0, len(p.Data))
	for _, c := range p.Data {
		commitments = append(commitments, NewCommitment(c))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListCommitmentsOutput{
		NextCursor:  nextCursor,
		Commitments: commitments,
	}
}
