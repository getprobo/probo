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

type Framework struct {
	ID             gid.GID            `json:"id"`
	OrganizationID gid.GID            `json:"-"`
	Name           string             `json:"name"`
	Description    *string            `json:"description,omitempty"`
	Organization   *Organization      `json:"organization"`
	Controls       *ControlConnection `json:"controls"`
	CreatedAt      time.Time          `json:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt"`
}

func (Framework) IsNode()             {}
func (this Framework) GetID() gid.GID { return this.ID }

type (
	FrameworkOrderBy OrderBy[coredata.FrameworkOrderField]

	FrameworkConnection struct {
		TotalCount int
		Edges      []*FrameworkEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewFrameworkConnection(
	p *page.Page[*coredata.Framework, coredata.FrameworkOrderField],
	parentType any,
	parentID gid.GID,
) *FrameworkConnection {
	var edges = make([]*FrameworkEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewFrameworkEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &FrameworkConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewFrameworkEdge(f *coredata.Framework, orderBy coredata.FrameworkOrderField) *FrameworkEdge {
	return &FrameworkEdge{
		Cursor: f.CursorKey(orderBy),
		Node:   NewFramework(f),
	}
}

func NewFramework(f *coredata.Framework) *Framework {
	return &Framework{
		ID:             f.ID,
		OrganizationID: f.OrganizationID,
		Name:           f.Name,
		Description:    f.Description,
		CreatedAt:      f.CreatedAt,
		UpdatedAt:      f.UpdatedAt,
	}
}
