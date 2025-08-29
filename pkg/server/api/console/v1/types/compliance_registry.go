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
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
)

type (
	ComplianceRegistryOrderBy OrderBy[coredata.ComplianceRegistryOrderField]

	ComplianceRegistryConnection struct {
		TotalCount int
		Edges      []*ComplianceRegistryEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filter   *ComplianceRegistryFilter
	}
)

func NewComplianceRegistryConnection(
	p *page.Page[*coredata.ComplianceRegistry, coredata.ComplianceRegistryOrderField],
	parentType any,
	parentID gid.GID,
	filter *ComplianceRegistryFilter,
) *ComplianceRegistryConnection {
	edges := make([]*ComplianceRegistryEdge, len(p.Data))
	for i, registry := range p.Data {
		edges[i] = NewComplianceRegistryEdge(registry, p.Cursor.OrderBy.Field)
	}

	return &ComplianceRegistryConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
		Filter:   filter,
	}
}

func NewComplianceRegistry(cr *coredata.ComplianceRegistry) *ComplianceRegistry {
	return &ComplianceRegistry{
		ID:                     cr.ID,
		SnapshotID:             cr.SnapshotID,
		SourceID:               cr.SourceID,
		ReferenceID:            cr.ReferenceID,
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

func NewComplianceRegistryEdge(cr *coredata.ComplianceRegistry, orderField coredata.ComplianceRegistryOrderField) *ComplianceRegistryEdge {
	return &ComplianceRegistryEdge{
		Node:   NewComplianceRegistry(cr),
		Cursor: cr.CursorKey(orderField),
	}
}
