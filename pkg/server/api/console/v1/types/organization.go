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
	OrganizationOrderBy OrderBy[coredata.OrganizationOrderField]
)

func NewOrganizationConnection(page *page.Page[*coredata.Organization, coredata.OrganizationOrderField]) *OrganizationConnection {
	var edges = make([]*OrganizationEdge, len(page.Data))

	for i := range edges {
		edges[i] = NewOrganizationEdge(page.Data[i], page.Cursor.OrderBy.Field)
	}

	return &OrganizationConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(page),
	}
}

func NewOrganizationEdge(o *coredata.Organization, orderBy coredata.OrganizationOrderField) *OrganizationEdge {
	return &OrganizationEdge{
		Cursor: o.CursorKey(orderBy),
		Node:   NewOrganization(o),
	}
}

func NewOrganization(o *coredata.Organization) *Organization {
	return &Organization{
		ID:                      o.ID,
		Name:                    o.Name,
		MailingAddress:          o.MailingAddress,
		TelephoneNumber:         o.TelephoneNumber,
		WebsiteURL:              o.WebsiteURL,
		SecurityComplianceEmail: o.SecurityComplianceEmail,
		CompanyDescription:      o.CompanyDescription,
		CompanyLegalName:        o.CompanyLegalName,
		CreatedAt:               o.CreatedAt,
		UpdatedAt:               o.UpdatedAt,
	}
}
