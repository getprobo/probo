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
	IncidentRegistryOrderBy OrderBy[coredata.IncidentRegistryOrderField]

	IncidentRegistryConnection struct {
		TotalCount int
		Edges      []*IncidentRegistryEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewIncidentRegistryConnection(
	p *page.Page[*coredata.IncidentRegistry, coredata.IncidentRegistryOrderField],
	parentType any,
	parentID gid.GID,
) *IncidentRegistryConnection {
	edges := make([]*IncidentRegistryEdge, len(p.Data))
	for i, registry := range p.Data {
		edges[i] = NewIncidentRegistryEdge(registry, p.Cursor.OrderBy.Field)
	}

	return &IncidentRegistryConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewIncidentRegistry(ir *coredata.IncidentRegistry) *IncidentRegistry {
	return &IncidentRegistry{
		ID:           ir.ID,
		ReferenceID:  ir.ReferenceID,
		Title:        ir.Title,
		Description:  ir.Description,
		Source:       ir.Source,
		IncidentDate: ir.IncidentDate,
		ResolvedDate: ir.ResolvedDate,
		Status:       ir.Status,
		Priority:     ir.Priority,
		Severity:     ir.Severity,
		Category:     ir.Category,
		CreatedAt:    ir.CreatedAt,
		UpdatedAt:    ir.UpdatedAt,
	}
}

func NewIncidentRegistryEdge(ir *coredata.IncidentRegistry, orderField coredata.IncidentRegistryOrderField) *IncidentRegistryEdge {
	return &IncidentRegistryEdge{
		Node:   NewIncidentRegistry(ir),
		Cursor: ir.CursorKey(orderField),
	}
}