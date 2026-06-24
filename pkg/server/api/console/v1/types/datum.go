// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
)

type Datum struct {
	ID                 gid.GID                     `json:"id"`
	OrganizationID     gid.GID                     `json:"-"`
	Name               string                      `json:"name"`
	DataClassification coredata.DataClassification `json:"dataClassification"`
	Owner              *Profile                    `json:"owner"`
	ThirdParties       *ThirdPartyConnection       `json:"third_parties"`
	Organization       *Organization               `json:"organization"`
	CreatedAt          time.Time                   `json:"createdAt"`
	UpdatedAt          time.Time                   `json:"updatedAt"`
}

func (Datum) IsNode()          {}
func (d Datum) GetID() gid.GID { return d.ID }

type (
	DatumOrderBy OrderBy[coredata.DatumOrderField]

	DatumConnection struct {
		TotalCount int
		Edges      []*DatumEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewDataConnection(
	p *page.Page[*coredata.Datum, coredata.DatumOrderField],
	parentType any,
	parentID gid.GID,
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
	}
}

func NewDatum(d *coredata.Datum) *Datum {
	return &Datum{
		ID: d.ID,
		Organization: &Organization{
			ID: d.OrganizationID,
		},
		Owner: &Profile{
			ID: d.OwnerID,
		},
		OrganizationID:     d.OrganizationID,
		Name:               d.Name,
		DataClassification: d.DataClassification,
		CreatedAt:          d.CreatedAt,
		UpdatedAt:          d.UpdatedAt,
	}
}

func NewDatumEdge(d *coredata.Datum, orderField coredata.DatumOrderField) *DatumEdge {
	return &DatumEdge{
		Node:   NewDatum(d),
		Cursor: d.CursorKey(orderField),
	}
}
