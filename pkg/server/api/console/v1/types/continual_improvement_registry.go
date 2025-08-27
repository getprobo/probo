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
	ContinualImprovementRegistriesOrderBy OrderBy[coredata.ContinualImprovementRegistriesOrderField]

	ContinualImprovementRegistryConnection struct {
		TotalCount int
		Edges      []*ContinualImprovementRegistryEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewContinualImprovementRegistryConnection(
	p *page.Page[*coredata.ContinualImprovementRegistry, coredata.ContinualImprovementRegistriesOrderField],
	parentType any,
	parentID gid.GID,
) *ContinualImprovementRegistryConnection {
	edges := make([]*ContinualImprovementRegistryEdge, len(p.Data))
	for i, registry := range p.Data {
		edges[i] = NewContinualImprovementRegistryEdge(registry, p.Cursor.OrderBy.Field)
	}

	return &ContinualImprovementRegistryConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewContinualImprovementRegistry(cir *coredata.ContinualImprovementRegistry) *ContinualImprovementRegistry {
	return &ContinualImprovementRegistry{
		ID:          cir.ID,
		ReferenceID: cir.ReferenceID,
		Description: cir.Description,
		Source:      cir.Source,
		TargetDate:  cir.TargetDate,
		Status:      cir.Status,
		Priority:    cir.Priority,
		CreatedAt:   cir.CreatedAt,
		UpdatedAt:   cir.UpdatedAt,
	}
}

func NewContinualImprovementRegistryEdge(cir *coredata.ContinualImprovementRegistry, orderField coredata.ContinualImprovementRegistriesOrderField) *ContinualImprovementRegistryEdge {
	return &ContinualImprovementRegistryEdge{
		Node:   NewContinualImprovementRegistry(cir),
		Cursor: cir.CursorKey(orderField),
	}
}
