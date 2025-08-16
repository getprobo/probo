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
	NonconformityRegistryOrderBy OrderBy[coredata.NonconformityRegistryOrderField]

	NonconformityRegistryConnection struct {
		TotalCount int
		Edges      []*NonconformityRegistryEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewNonconformityRegistryConnection(
	p *page.Page[*coredata.NonconformityRegistry, coredata.NonconformityRegistryOrderField],
	parentType any,
	parentID gid.GID,
) *NonconformityRegistryConnection {
	edges := make([]*NonconformityRegistryEdge, len(p.Data))
	for i, registry := range p.Data {
		edges[i] = NewNonconformityRegistryEdge(registry, p.Cursor.OrderBy.Field)
	}

	return &NonconformityRegistryConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewNonconformityRegistry(nr *coredata.NonconformityRegistry) *NonconformityRegistry {
	return &NonconformityRegistry{
		ID:                 nr.ID,
		ReferenceID:        nr.ReferenceID,
		Description:        nr.Description,
		DateIdentified:     nr.DateIdentified,
		RootCause:          nr.RootCause,
		CorrectiveAction:   nr.CorrectiveAction,
		DueDate:            nr.DueDate,
		Status:             nr.Status,
		EffectivenessCheck: nr.EffectivenessCheck,
		CreatedAt:          nr.CreatedAt,
		UpdatedAt:          nr.UpdatedAt,
	}
}

func NewNonconformityRegistryEdge(nr *coredata.NonconformityRegistry, orderField coredata.NonconformityRegistryOrderField) *NonconformityRegistryEdge {
	return &NonconformityRegistryEdge{
		Node:   NewNonconformityRegistry(nr),
		Cursor: nr.CursorKey(orderField),
	}
}
