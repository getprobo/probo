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

type Nonconformity struct {
	ID                 gid.GID                      `json:"id"`
	OrganizationID     gid.GID                      `json:"-"`
	SnapshotID         *gid.GID                     `json:"snapshotId,omitempty"`
	Organization       *Organization                `json:"organization"`
	ReferenceID        string                       `json:"referenceId"`
	Description        *string                      `json:"description,omitempty"`
	Audit              *Audit                       `json:"audit"`
	DateIdentified     *time.Time                   `json:"dateIdentified,omitempty"`
	RootCause          string                       `json:"rootCause"`
	CorrectiveAction   *string                      `json:"correctiveAction,omitempty"`
	Owner              *People                      `json:"owner"`
	DueDate            *time.Time                   `json:"dueDate,omitempty"`
	Status             coredata.NonconformityStatus `json:"status"`
	EffectivenessCheck *string                      `json:"effectivenessCheck,omitempty"`
	CreatedAt          time.Time                    `json:"createdAt"`
	UpdatedAt          time.Time                    `json:"updatedAt"`
}

func (Nonconformity) IsNode()             {}
func (this Nonconformity) GetID() gid.GID { return this.ID }

type (
	NonconformityOrderBy OrderBy[coredata.NonconformityOrderField]

	NonconformityConnection struct {
		TotalCount int
		Edges      []*NonconformityEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filter   *NonconformityFilter
	}
)

func NewNonconformityConnection(
	p *page.Page[*coredata.Nonconformity, coredata.NonconformityOrderField],
	parentType any,
	parentID gid.GID,
	filter *NonconformityFilter,
) *NonconformityConnection {
	edges := make([]*NonconformityEdge, len(p.Data))
	for i, nonconformity := range p.Data {
		edges[i] = NewNonconformityEdge(nonconformity, p.Cursor.OrderBy.Field)
	}

	return &NonconformityConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
		Filter:   filter,
	}
}

func NewNonconformity(nr *coredata.Nonconformity) *Nonconformity {
	return &Nonconformity{
		ID:                 nr.ID,
		OrganizationID:     nr.OrganizationID,
		ReferenceID:        nr.ReferenceID,
		SnapshotID:         nr.SnapshotID,
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

func NewNonconformityEdge(nr *coredata.Nonconformity, orderField coredata.NonconformityOrderField) *NonconformityEdge {
	return &NonconformityEdge{
		Node:   NewNonconformity(nr),
		Cursor: nr.CursorKey(orderField),
	}
}
