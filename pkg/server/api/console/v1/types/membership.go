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
	"github.com/getprobo/probo/pkg/page"
)

type (
	MembershipOrderBy OrderBy[coredata.MembershipOrderField]
)

func NewMembershipConnection(p *page.Page[*coredata.Membership, coredata.MembershipOrderField]) *MembershipConnection {
	var edges = make([]*MembershipEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewMembershipEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &MembershipConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
	}
}

func NewMembershipEdge(membership *coredata.Membership, orderBy coredata.MembershipOrderField) *MembershipEdge {
	return &MembershipEdge{
		Cursor: membership.CursorKey(orderBy),
		Node:   NewMembership(membership),
	}
}

func NewMembership(m *coredata.Membership) *Membership {
	return &Membership{
		ID:             m.ID,
		UserID:         m.UserID,
		OrganizationID: m.OrganizationID,
		Role:           m.Role,
		FullName:       m.FullName,
		EmailAddress:   m.EmailAddress,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}
