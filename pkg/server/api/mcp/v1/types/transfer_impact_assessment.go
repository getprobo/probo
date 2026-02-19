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
		NextCursor:                 nextCursor,
		TransferImpactAssessments: items,
	}
}
