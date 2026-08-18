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
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/riskmanagement"
)

func NewTreatmentPlan(tp *coredata.TreatmentPlan, progress riskmanagement.TreatmentProgress) *TreatmentPlan {
	netLikelihood, netImpact, netRiskScore := riskmanagement.NetScores(tp, progress)

	return &TreatmentPlan{
		ID:                     tp.ID,
		OrganizationID:         tp.OrganizationID,
		RiskID:                 tp.RiskID,
		RiskAnalysisID:         tp.RiskAnalysisID,
		Treatment:              tp.Treatment,
		OwnerID:                tp.OwnerID,
		InherentLikelihood:     tp.InherentLikelihood,
		InherentImpact:         tp.InherentImpact,
		InherentRiskScore:      tp.InherentRiskScore,
		ResidualLikelihood:     tp.ResidualLikelihood,
		ResidualImpact:         tp.ResidualImpact,
		ResidualRiskScore:      tp.ResidualRiskScore,
		NetLikelihood:          netLikelihood,
		NetImpact:              netImpact,
		NetRiskScore:           netRiskScore,
		MeasuresTotal:          progress.Total,
		MeasuresDone:           progress.Done,
		MeasuresInProgress:     progress.InProgress,
		MeasuresNotImplemented: progress.NotImplemented,
		CreatedAt:              tp.CreatedAt,
		UpdatedAt:              tp.UpdatedAt,
	}
}

func NewListTreatmentPlansOutput(
	p *page.Page[*coredata.TreatmentPlan, coredata.TreatmentPlanOrderField],
	progressByID map[gid.GID]riskmanagement.TreatmentProgress,
) ListTreatmentPlansOutput {
	items := make([]*TreatmentPlan, 0, len(p.Data))
	for _, v := range p.Data {
		items = append(items, NewTreatmentPlan(v, progressByID[v.ID]))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListTreatmentPlansOutput{
		NextCursor:     nextCursor,
		TreatmentPlans: items,
	}
}
