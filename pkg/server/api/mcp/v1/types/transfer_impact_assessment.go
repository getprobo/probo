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

func NewTransferImpactAssessment(t *coredata.TransferImpactAssessment) *TransferImpactAssessment {
	if t == nil {
		return nil
	}

	return &TransferImpactAssessment{
		ID:                    t.ID,
		OrganizationID:        t.OrganizationID,
		ProcessingActivityID:  t.ProcessingActivityID,
		DataSubjects:          t.DataSubjects,
		LegalMechanism:        t.LegalMechanism,
		Transfer:              t.Transfer,
		LocalLawRisk:          t.LocalLawRisk,
		SupplementaryMeasures: t.SupplementaryMeasures,
		CreatedAt:             t.CreatedAt,
		UpdatedAt:             t.UpdatedAt,
	}
}

func NewListTransferImpactAssessmentsOutput(pg *page.Page[*coredata.TransferImpactAssessment, coredata.TransferImpactAssessmentOrderField]) ListTransferImpactAssessmentsOutput {
	items := make([]*TransferImpactAssessment, 0, len(pg.Data))
	for _, v := range pg.Data {
		items = append(items, NewTransferImpactAssessment(v))
	}

	var nextCursor *page.CursorKey

	if len(pg.Data) > 0 {
		cursorKey := pg.Data[len(pg.Data)-1].CursorKey(pg.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListTransferImpactAssessmentsOutput{
		NextCursor:                nextCursor,
		TransferImpactAssessments: items,
	}
}
