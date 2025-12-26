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
	InvitationOrderBy OrderBy[coredata.InvitationOrderField]

	InvitationConnection struct {
		TotalCount int
		Edges      []*InvitationEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filters  *coredata.InvitationFilter
	}
)

func NewInvitationConnection(
	p *page.Page[*coredata.Invitation, coredata.InvitationOrderField],
	resolver any,
	parentID gid.GID,
	filters *coredata.InvitationFilter,
) *InvitationConnection {
	edges := make([]*InvitationEdge, len(p.Data))
	for i, invitation := range p.Data {
		edges[i] = NewInvitationEdge(invitation, p.Cursor.OrderBy.Field)
	}

	return &InvitationConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: resolver,
		ParentID: parentID,
		Filters:  filters,
	}
}

func NewInvitationEdge(invitation *coredata.Invitation, orderField coredata.InvitationOrderField) *InvitationEdge {
	return &InvitationEdge{
		Node:   NewInvitation(invitation),
		Cursor: invitation.CursorKey(orderField),
	}
}

func NewInvitation(invitation *coredata.Invitation) *Invitation {
	return &Invitation{
		ID:         invitation.ID,
		Email:      invitation.Email,
		Role:       invitation.Role,
		FullName:   invitation.FullName,
		ExpiresAt:  invitation.ExpiresAt,
		AcceptedAt: invitation.AcceptedAt,
		CreatedAt:  invitation.CreatedAt,
		Status:     invitation.Status,
		Organization: &Organization{
			ID: invitation.OrganizationID,
		},
	}
}
