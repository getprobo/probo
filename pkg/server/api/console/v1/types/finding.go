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
	FindingEdge struct {
		Cursor page.CursorKey `json:"cursor"`
		Node   *Finding       `json:"node"`
	}

	FindingConnection struct {
		TotalCount int
		Edges      []*FindingEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filter   *FindingFilter
	}
)

func NewFindingConnection(
	p *page.Page[*coredata.Finding, coredata.FindingOrderField],
	parentType any,
	parentID gid.GID,
	filter *FindingFilter,
) *FindingConnection {
	edges := make([]*FindingEdge, len(p.Data))
	for i, finding := range p.Data {
		edges[i] = NewFindingEdge(finding, p.Cursor.OrderBy.Field)
	}

	return &FindingConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
		Filter:   filter,
	}
}

func NewFindingEdge(f *coredata.Finding, orderField coredata.FindingOrderField) *FindingEdge {
	return &FindingEdge{
		Node:   NewFinding(f),
		Cursor: f.CursorKey(orderField),
	}
}

func NewFinding(f *coredata.Finding) *Finding {
	finding := &Finding{
		ID: f.ID,
		Organization: &Organization{
			ID: f.OrganizationID,
		},
		Kind:               f.Kind,
		ReferenceID:        f.ReferenceID,
		Description:        f.Description,
		Source:             f.Source,
		IdentifiedOn:       f.IdentifiedOn,
		RootCause:          f.RootCause,
		CorrectiveAction:   f.CorrectiveAction,
		DueDate:            f.DueDate,
		Status:             f.Status,
		Priority:           f.Priority,
		EffectivenessCheck: f.EffectivenessCheck,
		CreatedAt:          f.CreatedAt,
		UpdatedAt:          f.UpdatedAt,
	}

	if f.OwnerID != nil {
		finding.Owner = &Profile{
			ID: *f.OwnerID,
		}
	}

	if f.RiskID != nil {
		finding.Risk = &Risk{
			ID: *f.RiskID,
		}
	}

	return finding
}
