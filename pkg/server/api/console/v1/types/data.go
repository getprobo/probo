package types

import (
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/page"
)

func NewDatum(d *coredata.Data) *Datum {
	return &Datum{
		ID:              d.ID,
		Name:            d.Name,
		DataSensitivity: d.DataSensitivity,
		CreatedAt:       d.CreatedAt,
		UpdatedAt:       d.UpdatedAt,
		Organization:    &Organization{ID: d.OrganizationID},
	}
}

func NewDatumEdge(d *coredata.Data, orderField coredata.DatumOrderField) *DatumEdge {
	return &DatumEdge{
		Node:   NewDatum(d),
		Cursor: d.CursorKey(orderField),
	}
}

func NewDataConnection(page *page.Page[*coredata.Data, coredata.DatumOrderField]) *DatumConnection {
	edges := make([]*DatumEdge, len(page.Data))
	for i, data := range page.Data {
		edges[i] = NewDatumEdge(data, page.Cursor.OrderBy.Field)
	}

	return &DatumConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(page),
	}
}
