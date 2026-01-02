package types

import (
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type Control struct {
	ID                     gid.GID                `json:"id"`
	OrganizationID         gid.GID                `json:"-"`
	SectionTitle           string                 `json:"sectionTitle"`
	Name                   string                 `json:"name"`
	Description            *string                `json:"description,omitempty"`
	Status                 coredata.ControlStatus `json:"status"`
	ExclusionJustification *string                `json:"exclusionJustification,omitempty"`
	BestPractice           bool                   `json:"bestPractice"`
	Framework              *Framework             `json:"framework"`
	Measures               *MeasureConnection     `json:"measures"`
	Documents              *DocumentConnection    `json:"documents"`
	Audits                 *AuditConnection       `json:"audits"`
	Snapshots              *SnapshotConnection    `json:"snapshots"`
	CreatedAt              time.Time              `json:"createdAt"`
	UpdatedAt              time.Time              `json:"updatedAt"`
}

func (Control) IsNode()          {}
func (c Control) GetID() gid.GID { return c.ID }

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

func NewControl(control *coredata.Control) *Control {
	return &Control{
		ID:                     control.ID,
		OrganizationID:         control.OrganizationID,
		SectionTitle:           control.SectionTitle,
		Name:                   control.Name,
		Description:            control.Description,
		Status:                 control.Status,
		ExclusionJustification: control.ExclusionJustification,
		BestPractice:           control.BestPractice,
		CreatedAt:              control.CreatedAt,
		UpdatedAt:              control.UpdatedAt,
	}
}

func NewControlEdge(control *coredata.Control, orderField coredata.ControlOrderField) *ControlEdge {
	return &ControlEdge{
		Node:   NewControl(control),
		Cursor: control.CursorKey(orderField),
	}
}
