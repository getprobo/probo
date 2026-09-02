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
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/riskmanagement"
)

type (
	TreatmentPlanOrderBy OrderBy[coredata.TreatmentPlanOrderField]

	TreatmentPlanConnection struct {
		TotalCount int
		Edges      []*TreatmentPlanEdge
		PageInfo   PageInfo
		Resolver   any
		ParentID   gid.GID
		Filters    *coredata.TreatmentPlanFilter
		AsOf       *time.Time
	}
)

func NewTreatmentPlanConnection(
	p *page.Page[*coredata.TreatmentPlan, coredata.TreatmentPlanOrderField],
	parentType any,
	parentID gid.GID,
	filters *coredata.TreatmentPlanFilter,
) *TreatmentPlanConnection {
	edges := make([]*TreatmentPlanEdge, len(p.Data))
	for i := range edges {
		edges[i] = NewTreatmentPlanEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &TreatmentPlanConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),
		Resolver: parentType,
		ParentID: parentID,
		Filters:  filters,
	}
}

func NewTreatmentPlanEdge(tp *coredata.TreatmentPlan, orderBy coredata.TreatmentPlanOrderField) *TreatmentPlanEdge {
	return &TreatmentPlanEdge{
		Cursor: tp.CursorKey(orderBy),
		Node:   NewTreatmentPlan(tp),
	}
}

func NewTreatmentPlan(tp *coredata.TreatmentPlan) *TreatmentPlan {
	plan := &TreatmentPlan{
		ID:                 tp.ID,
		Treatment:          tp.Treatment,
		InherentLikelihood: tp.InherentLikelihood,
		InherentImpact:     tp.InherentImpact,
		InherentRiskScore:  tp.InherentRiskScore,
		ResidualLikelihood: tp.ResidualLikelihood,
		ResidualImpact:     tp.ResidualImpact,
		ResidualRiskScore:  tp.ResidualRiskScore,
		Category:           tp.Category,
		Risk: &Risk{
			ID: tp.RiskID,
		},
		RiskAnalysis: &RiskAnalysis{
			ID: tp.RiskAnalysisID,
		},
		Organization: &Organization{
			ID: tp.OrganizationID,
		},
		Owner: &Profile{
			ID: tp.OwnerID,
		},
		CreatedAt: tp.CreatedAt,
		UpdatedAt: tp.UpdatedAt,
	}

	return plan
}

func NewTreatmentPlanAsOf(
	tp *coredata.TreatmentPlan,
	asOf time.Time,
	progress riskmanagement.TreatmentProgress,
) *TreatmentPlan {
	plan := NewTreatmentPlan(tp)
	plan.AsOf = &asOf
	plan.NetLikelihood, plan.NetImpact, plan.NetRiskScore = riskmanagement.NetScores(tp, progress)
	plan.Progress = &progress

	return plan
}

func NewTreatmentPlanConnectionAsOf(
	p *page.Page[*coredata.TreatmentPlan, coredata.TreatmentPlanOrderField],
	parentType any,
	parentID gid.GID,
	filters *coredata.TreatmentPlanFilter,
	asOf time.Time,
	totalCount int,
	progressByID map[gid.GID]riskmanagement.TreatmentProgress,
) *TreatmentPlanConnection {
	edges := make([]*TreatmentPlanEdge, len(p.Data))
	for i := range edges {
		edges[i] = &TreatmentPlanEdge{
			Cursor: p.Data[i].CursorKey(p.Cursor.OrderBy.Field),
			Node:   NewTreatmentPlanAsOf(p.Data[i], asOf, progressByID[p.Data[i].ID]),
		}
	}

	asOfCopy := asOf

	return &TreatmentPlanConnection{
		Edges:      edges,
		PageInfo:   *NewPageInfo(p),
		Resolver:   parentType,
		ParentID:   parentID,
		Filters:    filters,
		AsOf:       &asOfCopy,
		TotalCount: totalCount,
	}
}
