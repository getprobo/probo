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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	TaskOrderBy OrderBy[coredata.TaskOrderField]

	TaskConnection struct {
		TotalCount int
		Edges      []*TaskEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewTaskConnection(
	p *page.Page[*coredata.Task, coredata.TaskOrderField],
	parentType any,
	parentID gid.GID,
) *TaskConnection {
	var edges = make([]*TaskEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewTaskEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &TaskConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewTaskEdge(t *coredata.Task, orderBy coredata.TaskOrderField) *TaskEdge {
	return &TaskEdge{
		Cursor: t.CursorKey(orderBy),
		Node:   NewTask(t),
	}
}

func NewTask(t *coredata.Task) *Task {
	node := &Task{
		ID: t.ID,
		Organization: &Organization{
			ID: t.OrganizationID,
		},

		Name:         t.Name,
		Description:  t.Description,
		State:        t.State,
		TimeEstimate: t.TimeEstimate,
		Deadline:     t.Deadline,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
	}

	if t.MeasureID != nil {
		node.Measure = &Measure{
			ID: *t.MeasureID,
		}
	}

	if t.AssignedToID != nil {
		node.AssignedTo = &Profile{
			ID: *t.AssignedToID,
		}
	}

	return node
}
