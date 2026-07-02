// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
