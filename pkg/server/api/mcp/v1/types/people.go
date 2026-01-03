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

package types

import (
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
)

func NewPeople(p *coredata.People) *People {
	additionalEmails := p.AdditionalEmailAddresses
	if additionalEmails == nil {
		additionalEmails = []mail.Addr{}
	}
	return &People{
		FullName:                 p.FullName,
		ID:                       p.ID,
		OrganizationID:           p.OrganizationID,
		Kind:                     p.Kind,
		PrimaryEmailAddress:      p.PrimaryEmailAddress,
		AdditionalEmailAddresses: additionalEmails,
		Position:                 p.Position,
		ContractStartDate:        p.ContractStartDate,
		ContractEndDate:          p.ContractEndDate,
		CreatedAt:                p.CreatedAt,
		UpdatedAt:                p.UpdatedAt,
	}
}

func NewListPeopleOutput(peoplePage *page.Page[*coredata.People, coredata.PeopleOrderField]) ListPeopleOutput {
	people := make([]*People, 0, len(peoplePage.Data))
	for _, p := range peoplePage.Data {
		people = append(people, NewPeople(p))
	}

	var nextCursor *page.CursorKey
	if len(peoplePage.Data) > 0 {
		cursorKey := peoplePage.Data[len(peoplePage.Data)-1].CursorKey(peoplePage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListPeopleOutput{
		NextCursor: nextCursor,
		People:     people,
	}
}
