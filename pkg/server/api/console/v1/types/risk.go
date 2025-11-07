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
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type Risk struct {
	ID                 gid.GID                `json:"id"`
	OrganizationID     gid.GID                `json:"-"`
	SnapshotID         *gid.GID               `json:"snapshotId,omitempty"`
	Name               string                 `json:"name"`
	Description        *string                `json:"description,omitempty"`
	Category           string                 `json:"category"`
	Treatment          coredata.RiskTreatment `json:"treatment"`
	InherentLikelihood int                    `json:"inherentLikelihood"`
	InherentImpact     int                    `json:"inherentImpact"`
	InherentRiskScore  int                    `json:"inherentRiskScore"`
	ResidualLikelihood int                    `json:"residualLikelihood"`
	ResidualImpact     int                    `json:"residualImpact"`
	ResidualRiskScore  int                    `json:"residualRiskScore"`
	Note               string                 `json:"note"`
	Owner              *People                `json:"owner,omitempty"`
	Organization       *Organization          `json:"organization"`
	Measures           *MeasureConnection     `json:"measures"`
	Documents          *DocumentConnection    `json:"documents"`
	Controls           *ControlConnection     `json:"controls"`
	Obligations        *ObligationConnection  `json:"obligations"`
	CreatedAt          time.Time              `json:"createdAt"`
	UpdatedAt          time.Time              `json:"updatedAt"`
}

func (Risk) IsNode()             {}
func (this Risk) GetID() gid.GID { return this.ID }

type (
	RiskOrderBy OrderBy[coredata.RiskOrderField]

	RiskConnection struct {
		TotalCount int
		Edges      []*RiskEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filters  *coredata.RiskFilter
	}
)

func NewRiskConnection(
	p *page.Page[*coredata.Risk, coredata.RiskOrderField],
	parentType any,
	parentID gid.GID,
	filters *coredata.RiskFilter,
) *RiskConnection {
	var edges = make([]*RiskEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewRiskEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &RiskConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
		Filters:  filters,
	}
}

func NewRiskEdge(r *coredata.Risk, orderBy coredata.RiskOrderField) *RiskEdge {
	return &RiskEdge{
		Cursor: r.CursorKey(orderBy),
		Node:   NewRisk(r),
	}
}

func NewRisk(r *coredata.Risk) *Risk {
	return &Risk{
		ID:                 r.ID,
		OrganizationID:     r.OrganizationID,
		Name:               r.Name,
		SnapshotID:         r.SnapshotID,
		Description:        r.Description,
		Treatment:          r.Treatment,
		InherentLikelihood: r.InherentLikelihood,
		InherentImpact:     r.InherentImpact,
		InherentRiskScore:  r.InherentRiskScore,
		ResidualLikelihood: r.ResidualLikelihood,
		ResidualImpact:     r.ResidualImpact,
		ResidualRiskScore:  r.ResidualRiskScore,
		Category:           r.Category,
		CreatedAt:          r.CreatedAt,
		UpdatedAt:          r.UpdatedAt,
		Note:               r.Note,
	}
}
