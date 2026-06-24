// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
