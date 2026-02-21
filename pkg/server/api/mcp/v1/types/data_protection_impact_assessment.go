// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHOR BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN
// ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package types

import (
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/page"
)

func NewDataProtectionImpactAssessment(
	d *coredata.DataProtectionImpactAssessment,
) *DataProtectionImpactAssessment {
	return &DataProtectionImpactAssessment{
		ID:                          d.ID,
		OrganizationID:              d.OrganizationID,
		ProcessingActivityID:        d.ProcessingActivityID,
		Description:                 d.Description,
		NecessityAndProportionality: d.NecessityAndProportionality,
		PotentialRisk:               d.PotentialRisk,
		Mitigations:                 d.Mitigations,
		ResidualRisk:                d.ResidualRisk,
		CreatedAt:                   d.CreatedAt,
		UpdatedAt:                   d.UpdatedAt,
	}
}

func NewListDataProtectionImpactAssessmentsOutput(
	pg *page.Page[*coredata.DataProtectionImpactAssessment, coredata.DataProtectionImpactAssessmentOrderField],
) ListDataProtectionImpactAssessmentsOutput {
	items := make([]*DataProtectionImpactAssessment, 0, len(pg.Data))
	for _, v := range pg.Data {
		items = append(items, NewDataProtectionImpactAssessment(v))
	}
	var nextCursor *page.CursorKey
	if len(pg.Data) > 0 {
		cursorKey := pg.Data[len(pg.Data)-1].CursorKey(pg.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}
	return ListDataProtectionImpactAssessmentsOutput{
		NextCursor:                      nextCursor,
		DataProtectionImpactAssessments: items,
	}
}
