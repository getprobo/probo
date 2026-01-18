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
	MembershipOrderBy OrderBy[coredata.MembershipOrderField]

	MembershipConnection struct {
		TotalCount int
		Edges      []*MembershipEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewMembershipConnection(
	p *page.Page[*coredata.Membership, coredata.MembershipOrderField],
	resolver any,
	parentID gid.GID,
) *MembershipConnection {
	edges := make([]*MembershipEdge, len(p.Data))
	for i, membership := range p.Data {
		edges[i] = NewMembershipEdge(membership, p.Cursor.OrderBy.Field)
	}

	return &MembershipConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: resolver,
		ParentID: parentID,
	}
}

func NewMembershipEdge(membership *coredata.Membership, orderField coredata.MembershipOrderField) *MembershipEdge {
	return &MembershipEdge{
		Node:   NewMembership(membership),
		Cursor: membership.CursorKey(orderField),
	}
}

func NewMembership(membership *coredata.Membership) *Membership {
	return &Membership{
		ID:        membership.ID,
		CreatedAt: membership.CreatedAt,
		Identity: &Identity{
			ID: membership.IdentityID,
		},
		Organization: &Organization{
			ID: membership.OrganizationID,
		},
		Role:   membership.Role,
		Source: membership.Source,
		State:  membership.State,
	}
}
