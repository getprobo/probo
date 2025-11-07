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

type Task struct {
	ID             gid.GID        `json:"id"`
	OrganizationID gid.GID        `json:"-"`
	Name           string         `json:"name"`
	Description    *string        `json:"description,omitempty"`
	State          coredata.TaskState `json:"state"`
	TimeEstimate   *time.Duration `json:"timeEstimate,omitempty"`
	Deadline       *time.Time     `json:"deadline,omitempty"`
	AssignedTo     *People        `json:"assignedTo,omitempty"`
	Organization   *Organization  `json:"organization"`
	Measure        *Measure       `json:"measure,omitempty"`
	Evidences      *EvidenceConnection `json:"evidences"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
}

func (Task) IsNode()          {}
func (t Task) GetID() gid.GID { return t.ID }

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
	return &Task{
		ID:             t.ID,
		OrganizationID: t.OrganizationID,
		Name:           t.Name,
		Description:    t.Description,
		State:          t.State,
		TimeEstimate:   t.TimeEstimate,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
		Deadline:       t.Deadline,
	}
}
