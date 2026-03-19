// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
		ID:         f.ID,
		SnapshotID: f.SnapshotID,
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
