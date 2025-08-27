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

type TrustCenterAccessOrderBy = OrderBy[coredata.TrustCenterAccessOrderField]

func NewTrustCenterAccess(tca *coredata.TrustCenterAccess) *TrustCenterAccess {
	return &TrustCenterAccess{
		ID:        tca.ID,
		Email:     tca.Email,
		Name:      tca.Name,
		Active:    tca.Active,
		CreatedAt: tca.CreatedAt,
		UpdatedAt: tca.UpdatedAt,
	}
}

func NewTrustCenterAccessConnection(
	page *page.Page[*coredata.TrustCenterAccess, coredata.TrustCenterAccessOrderField],
) *TrustCenterAccessConnection {
	var edges = make([]*TrustCenterAccessEdge, len(page.Data))

	for i := range edges {
		edges[i] = NewTrustCenterAccessEdge(page.Data[i], page.Cursor.OrderBy.Field)
	}

	return &TrustCenterAccessConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(page),
	}
}

func NewTrustCenterAccessEdge(tca *coredata.TrustCenterAccess, orderBy coredata.TrustCenterAccessOrderField) *TrustCenterAccessEdge {
	return &TrustCenterAccessEdge{
		Cursor: tca.CursorKey(orderBy),
		Node:   NewTrustCenterAccess(tca),
	}
}

// Types are auto-generated in types.go - only helper functions remain here
