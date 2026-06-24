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
