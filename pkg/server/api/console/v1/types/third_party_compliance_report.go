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
	ThirdPartyComplianceReportOrderBy OrderBy[coredata.ThirdPartyComplianceReportOrderField]
)

func NewThirdPartyComplianceReportConnection(p *page.Page[*coredata.ThirdPartyComplianceReport, coredata.ThirdPartyComplianceReportOrderField]) *ThirdPartyComplianceReportConnection {
	var edges = make([]*ThirdPartyComplianceReportEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewThirdPartyComplianceReportEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &ThirdPartyComplianceReportConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
	}
}

func NewThirdPartyComplianceReportEdge(c *coredata.ThirdPartyComplianceReport, orderBy coredata.ThirdPartyComplianceReportOrderField) *ThirdPartyComplianceReportEdge {
	return &ThirdPartyComplianceReportEdge{
		Cursor: c.CursorKey(orderBy),
		Node:   NewThirdPartyComplianceReport(c),
	}
}

func NewThirdPartyComplianceReport(c *coredata.ThirdPartyComplianceReport) *ThirdPartyComplianceReport {
	object := &ThirdPartyComplianceReport{
		ID: c.ID,
		ThirdParty: &ThirdParty{
			ID: c.ThirdPartyID,
		},
		ReportDate: c.ReportDate,
		ValidUntil: c.ValidUntil,
		ReportName: c.ReportName,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
	}

	if c.ReportFileId != nil {
		object.File = &File{
			ID: *c.ReportFileId,
		}
	}

	return object
}
