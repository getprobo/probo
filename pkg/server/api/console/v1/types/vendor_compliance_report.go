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

type (
	VendorComplianceReportOrderBy OrderBy[coredata.VendorComplianceReportOrderField]
)

func NewVendorComplianceReportConnection(p *page.Page[*coredata.VendorComplianceReport, coredata.VendorComplianceReportOrderField]) *VendorComplianceReportConnection {
	var edges = make([]*VendorComplianceReportEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewVendorComplianceReportEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &VendorComplianceReportConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
	}
}

func NewVendorComplianceReportEdge(c *coredata.VendorComplianceReport, orderBy coredata.VendorComplianceReportOrderField) *VendorComplianceReportEdge {
	return &VendorComplianceReportEdge{
		Cursor: c.CursorKey(orderBy),
		Node:   NewVendorComplianceReport(c),
	}
}

func NewVendorComplianceReport(c *coredata.VendorComplianceReport) *VendorComplianceReport {
	object := &VendorComplianceReport{
		ID: c.ID,
		Vendor: &Vendor{
			ID: c.VendorID,
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
