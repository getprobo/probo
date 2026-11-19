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
)

type (
	CorrectiveActionPlanOrderBy OrderBy[coredata.CorrectiveActionPlanOrderField]

	CorrectiveActionPlanConnection struct {
		TotalCount int
		Edges      []*CorrectiveActionPlanEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewCorrectiveActionPlanConnection(
	p *page.Page[*coredata.CorrectiveActionPlan, coredata.CorrectiveActionPlanOrderField],
	parentType any,
	parentID gid.GID,
) *CorrectiveActionPlanConnection {
	edges := make([]*CorrectiveActionPlanEdge, len(p.Data))
	for i, plan := range p.Data {
		edges[i] = NewCorrectiveActionPlanEdge(plan, p.Cursor.OrderBy.Field)
	}

	return &CorrectiveActionPlanConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewCorrectiveActionPlanEdge(
	p *coredata.CorrectiveActionPlan,
	orderField coredata.CorrectiveActionPlanOrderField,
) *CorrectiveActionPlanEdge {
	return &CorrectiveActionPlanEdge{
		Cursor: p.CursorKey(orderField),
		Node:   NewCorrectiveActionPlan(p),
	}
}

func NewCorrectiveActionPlan(p *coredata.CorrectiveActionPlan) *CorrectiveActionPlan {
	plan := &CorrectiveActionPlan{
		ID:          p.ID,
		Title:       p.Title,
		Description: p.Description,
		DueDate:     p.DueDate,
		Finding: &Finding{
			ID: p.FindingID,
		},
		Organization: &Organization{
			ID: p.OrganizationID,
		},
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}

	if p.OwnerID != nil {
		plan.Owner = &Profile{
			ID: *p.OwnerID,
		}
	}

	return plan
}
