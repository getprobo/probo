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
	"go.probo.inc/probo/pkg/page"
)

func NewComplianceExternalURL(c *coredata.ComplianceExternalURL) *ComplianceExternalURL {
	return &ComplianceExternalURL{
		ID:   c.ID,
		Name: c.Name,
		URL:  c.URL,
		Rank: c.Rank,
	}
}

func NewComplianceExternalURLConnection(
	p *page.Page[*coredata.ComplianceExternalURL, coredata.ComplianceExternalURLOrderField],
) *ComplianceExternalURLConnection {
	edges := make([]*ComplianceExternalURLEdge, len(p.Data))

	for i, item := range p.Data {
		edges[i] = NewComplianceExternalURLEdge(item, p.Cursor.OrderBy.Field)
	}

	return &ComplianceExternalURLConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
	}
}

func NewComplianceExternalURLEdge(c *coredata.ComplianceExternalURL, orderBy coredata.ComplianceExternalURLOrderField) *ComplianceExternalURLEdge {
	return &ComplianceExternalURLEdge{
		Cursor: c.CursorKey(orderBy),
		Node:   NewComplianceExternalURL(c),
	}
}
