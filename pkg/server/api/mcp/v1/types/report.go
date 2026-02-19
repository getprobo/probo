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

func NewReport(r *coredata.Report) *Report {
	return &Report{
		ID:                    r.ID,
		Name:                  r.Name,
		OrganizationID:        r.OrganizationID,
		FrameworkID:           r.FrameworkID,
		FrameworkType:         r.FrameworkType,
		State:                 r.State,
		TrustCenterVisibility: r.TrustCenterVisibility,
		ValidFrom:             r.ValidFrom,
		ValidUntil:            r.ValidUntil,
		CreatedAt:             r.CreatedAt,
		UpdatedAt:             r.UpdatedAt,
	}
}

func NewListReportsOutput(reportPage *page.Page[*coredata.Report, coredata.ReportOrderField]) ListReportsOutput {
	reports := make([]*Report, 0, len(reportPage.Data))
	for _, v := range reportPage.Data {
		reports = append(reports, NewReport(v))
	}

	var nextCursor *page.CursorKey
	if len(reportPage.Data) > 0 {
		cursorKey := reportPage.Data[len(reportPage.Data)-1].CursorKey(reportPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListReportsOutput{
		NextCursor: nextCursor,
		Reports:    reports,
	}
}

