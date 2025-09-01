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
	DatumOrderBy OrderBy[coredata.DatumOrderField]

	DatumConnection struct {
		TotalCount int
		Edges      []*DatumEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filter   *DatumFilter
	}
)

func NewDataConnection(
	p *page.Page[*coredata.Datum, coredata.DatumOrderField],
	parentType any,
	parentID gid.GID,
	filter *DatumFilter,
) *DatumConnection {
	edges := make([]*DatumEdge, len(p.Data))
	for i, datum := range p.Data {
		edges[i] = NewDatumEdge(datum, p.Cursor.OrderBy.Field)
	}

	return &DatumConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
		Filter:   filter,
	}
}

func NewDatum(d *coredata.Datum) *Datum {
	return &Datum{
		ID:                 d.ID,
		Name:               d.Name,
		SnapshotID:         d.SnapshotID,
		DataClassification: d.DataClassification,
		CreatedAt:          d.CreatedAt,
		UpdatedAt:          d.UpdatedAt,
		Organization:       &Organization{ID: d.OrganizationID},
	}
}

func NewDatumEdge(d *coredata.Datum, orderField coredata.DatumOrderField) *DatumEdge {
	return &DatumEdge{
		Node:   NewDatum(d),
		Cursor: d.CursorKey(orderField),
	}
}
