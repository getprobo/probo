// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	"go.probo.inc/probo/pkg/page"
)

type (
	ThirdPartyContactOrderBy OrderBy[coredata.ThirdPartyContactOrderField]
)

func NewThirdPartyContactConnection(p *page.Page[*coredata.ThirdPartyContact, coredata.ThirdPartyContactOrderField]) *ThirdPartyContactConnection {
	var edges = make([]*ThirdPartyContactEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewThirdPartyContactEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &ThirdPartyContactConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
	}
}

func NewThirdPartyContactEdge(c *coredata.ThirdPartyContact, orderBy coredata.ThirdPartyContactOrderField) *ThirdPartyContactEdge {
	return &ThirdPartyContactEdge{
		Cursor: c.CursorKey(orderBy),
		Node:   NewThirdPartyContact(c),
	}
}

func NewThirdPartyContact(c *coredata.ThirdPartyContact) *ThirdPartyContact {
	return &ThirdPartyContact{
		ID: c.ID,
		ThirdParty: &ThirdParty{
			ID: c.ThirdPartyID,
		},
		FullName:  c.FullName,
		Email:     c.Email,
		Phone:     c.Phone,
		Role:      c.Role,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}
