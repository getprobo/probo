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
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
)

type (
	PeopleOrderBy OrderBy[coredata.PeopleOrderField]

	PeopleConnection struct {
		TotalCount int
		Edges      []*PeopleEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filters  *coredata.PeopleFilter
	}
)

func NewPeopleConnection(
	p *page.Page[*coredata.People, coredata.PeopleOrderField],
	parentType any,
	parentID gid.GID,
	filters *coredata.PeopleFilter,
) *PeopleConnection {
	var edges = make([]*PeopleEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewPeopleEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &PeopleConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
		Filters:  filters,
	}
}

func NewPeopleEdge(p *coredata.People, orderBy coredata.PeopleOrderField) *PeopleEdge {
	return &PeopleEdge{
		Cursor: p.CursorKey(orderBy),
		Node:   NewPeople(p),
	}
}

func NewPeople(p *coredata.People) *People {
	var emailAddresses []*mail.Addr
	for _, emailAddress := range p.AdditionalEmailAddresses {
		emailAddresses = append(emailAddresses, &emailAddress)
	}

	return &People{
		ID:                       p.ID,
		FullName:                 p.FullName,
		PrimaryEmailAddress:      p.PrimaryEmailAddress,
		AdditionalEmailAddresses: emailAddresses,
		Kind:                     p.Kind,
		Position:                 p.Position,
		ContractStartDate:        p.ContractStartDate,
		ContractEndDate:          p.ContractEndDate,
		CreatedAt:                p.CreatedAt,
		UpdatedAt:                p.UpdatedAt,
	}
}
