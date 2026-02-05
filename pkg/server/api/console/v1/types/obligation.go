// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
	ObligationOrderBy OrderBy[coredata.ObligationOrderField]

	ObligationConnection struct {
		TotalCount int
		Edges      []*ObligationEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filter   *ObligationFilter
	}
)

func NewObligationConnection(
	p *page.Page[*coredata.Obligation, coredata.ObligationOrderField],
	parentType any,
	parentID gid.GID,
	filter *ObligationFilter,
) *ObligationConnection {
	edges := make([]*ObligationEdge, len(p.Data))
	for i, obligation := range p.Data {
		edges[i] = NewObligationEdge(obligation, p.Cursor.OrderBy.Field)
	}

	return &ObligationConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
		Filter:   filter,
	}
}

func NewObligation(cr *coredata.Obligation) *Obligation {
	return &Obligation{
		ID:         cr.ID,
		SnapshotID: cr.SnapshotID,
		SourceID:   cr.SourceID,
		Organization: &Organization{
			ID: cr.OrganizationID,
		},
		Owner: &Profile{
			ID: cr.OwnerID,
		},
		Area:                   cr.Area,
		Source:                 cr.Source,
		Requirement:            cr.Requirement,
		ActionsToBeImplemented: cr.ActionsToBeImplemented,
		Regulator:              cr.Regulator,
		LastReviewDate:         cr.LastReviewDate,
		DueDate:                cr.DueDate,
		Status:                 cr.Status,
		CreatedAt:              cr.CreatedAt,
		UpdatedAt:              cr.UpdatedAt,
	}
}

func NewObligationEdge(cr *coredata.Obligation, orderField coredata.ObligationOrderField) *ObligationEdge {
	return &ObligationEdge{
		Node:   NewObligation(cr),
		Cursor: cr.CursorKey(orderField),
	}
}
