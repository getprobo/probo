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
	ThirdPartyServiceOrderBy OrderBy[coredata.ThirdPartyServiceOrderField]
)

func NewThirdPartyServiceConnection(p *page.Page[*coredata.ThirdPartyService, coredata.ThirdPartyServiceOrderField]) *ThirdPartyServiceConnection {
	var edges = make([]*ThirdPartyServiceEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewThirdPartyServiceEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &ThirdPartyServiceConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
	}
}

func NewThirdPartyServiceEdge(s *coredata.ThirdPartyService, orderBy coredata.ThirdPartyServiceOrderField) *ThirdPartyServiceEdge {
	return &ThirdPartyServiceEdge{
		Cursor: s.CursorKey(orderBy),
		Node:   NewThirdPartyService(s),
	}
}

func NewThirdPartyService(s *coredata.ThirdPartyService) *ThirdPartyService {
	return &ThirdPartyService{
		ID: s.ID,
		ThirdParty: &ThirdParty{
			ID: s.ThirdPartyID,
		},
		Name:        s.Name,
		Description: s.Description,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}
