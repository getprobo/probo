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
	ProfileOrderBy OrderBy[coredata.MembershipProfileOrderField]

	ProfileConnection struct {
		TotalCount int
		Edges      []*ProfileEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filters  *coredata.MembershipProfileFilter
	}
)

func NewProfileConnection(
	p *page.Page[*coredata.MembershipProfile, coredata.MembershipProfileOrderField],
	resolver any,
	parentID gid.GID,
	filters *coredata.MembershipProfileFilter,
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
		Filters:  filters,
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
		EmailAddress:             profile.EmailAddress,
		Source:                   profile.Source.String(),
		State:                    profile.State,
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
