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

func NewRisk(r *coredata.Risk) *Risk {
	return &Risk{
		ID:                 r.ID,
		OrganizationID:     r.OrganizationID,
		Name:               r.Name,
		Description:        r.Description,
		Category:           r.Category,
		Treatment:          r.Treatment,
		Note:               r.Note,
		OwnerID:            r.OwnerID,
		InherentLikelihood: r.InherentLikelihood,
		InherentImpact:     r.InherentImpact,
		InherentRiskScore:  r.InherentRiskScore,
		ResidualLikelihood: r.ResidualLikelihood,
		ResidualImpact:     r.ResidualImpact,
		ResidualRiskScore:  r.ResidualRiskScore,
		CreatedAt:          r.CreatedAt,
		UpdatedAt:          r.UpdatedAt,
	}
}

func NewListMeasureRisksOutput(riskPage *page.Page[*coredata.Risk, coredata.RiskOrderField]) ListMeasureRisksOutput {
	risks := make([]*Risk, 0, len(riskPage.Data))
	for _, v := range riskPage.Data {
		risks = append(risks, NewRisk(v))
	}

	var nextCursor *page.CursorKey

	if len(riskPage.Data) > 0 {
		cursorKey := riskPage.Data[len(riskPage.Data)-1].CursorKey(riskPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListMeasureRisksOutput{
		NextCursor: nextCursor,
		Risks:      risks,
	}
}

func NewListRisksOutput(riskPage *page.Page[*coredata.Risk, coredata.RiskOrderField]) ListRisksOutput {
	risks := make([]*Risk, 0, len(riskPage.Data))
	for _, v := range riskPage.Data {
		risks = append(risks, NewRisk(v))
	}

	var nextCursor *page.CursorKey

	if len(riskPage.Data) > 0 {
		cursorKey := riskPage.Data[len(riskPage.Data)-1].CursorKey(riskPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListRisksOutput{
		NextCursor: nextCursor,
		Risks:      risks,
	}
}
