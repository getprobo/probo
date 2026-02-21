package types

import (
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	ControlOrderBy OrderBy[coredata.ControlOrderField]

	ControlConnection struct {
		TotalCount int
		Edges      []*ControlEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filters  *coredata.ControlFilter
	}
)

func NewControlConnection(
	p *page.Page[*coredata.Control, coredata.ControlOrderField],
	resolver any,
	parentID gid.GID,
	filter *coredata.ControlFilter,
) *ControlConnection {
	edges := make([]*ControlEdge, len(p.Data))
	for i, control := range p.Data {
		edges[i] = NewControlEdge(control, p.Cursor.OrderBy.Field)
	}

	return &ControlConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: resolver,
		ParentID: parentID,
		Filters:  filter,
	}
}

func NewControlEdge(control *coredata.Control, orderField coredata.ControlOrderField) *ControlEdge {
	return &ControlEdge{
		Node:   NewControl(control),
		Cursor: control.CursorKey(orderField),
	}
}

func NewControl(control *coredata.Control) *Control {
	return &Control{
		ID: control.ID,
		Organization: &Organization{
			ID: control.OrganizationID,
		},
		Framework: &Framework{
			ID: control.FrameworkID,
		},
		SectionTitle: control.SectionTitle,
		Name:         control.Name,
		Description:  control.Description,
		BestPractice: control.BestPractice,
		CreatedAt:    control.CreatedAt,
		UpdatedAt:    control.UpdatedAt,
	}
}
