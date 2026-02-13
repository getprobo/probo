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
	ProfileOrderBy OrderBy[coredata.MembershipProfileOrderField]

	ProfileConnection struct {
		TotalCount int
		Edges      []*ProfileEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewProfileConnection(
	p *page.Page[*coredata.MembershipProfile, coredata.MembershipProfileOrderField],
	resolver any,
	parentID gid.GID,
) *ProfileConnection {
	edges := make([]*ProfileEdge, len(p.Data))
	for i, profile := range p.Data {
		edges[i] = NewProfileEdge(profile, p.Cursor.OrderBy.Field)
	}

	return &ProfileConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: resolver,
		ParentID: parentID,
	}
}

func NewProfileEdge(
	profile *coredata.MembershipProfile,
	orderField coredata.MembershipProfileOrderField,
) *ProfileEdge {
	return &ProfileEdge{
		Node:   NewProfile(profile),
		Cursor: profile.CursorKey(orderField),
	}
}

func NewProfile(profile *coredata.MembershipProfile) *Profile {
	return &Profile{
		ID:                       profile.ID,
		FullName:                 profile.FullName,
		AdditionalEmailAddresses: profile.AdditionalEmailAddresses,
		Kind:                     profile.Kind,
		Position:                 profile.Position,
		ContractStartDate:        profile.ContractStartDate,
		ContractEndDate:          profile.ContractEndDate,
		CreatedAt:                profile.CreatedAt,
		UpdatedAt:                profile.UpdatedAt,
		Identity: &Identity{
			ID: profile.IdentityID,
		},
		Organization: &Organization{
			ID: profile.OrganizationID,
		},
	}
}
