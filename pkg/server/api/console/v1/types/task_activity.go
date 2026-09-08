// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	TaskActivityOrderBy OrderBy[coredata.TaskActivityOrderField]

	TaskActivityConnection struct {
		TotalCount int
		Edges      []*TaskActivityEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewTaskActivityConnection(
	p *page.Page[*coredata.TaskActivity, coredata.TaskActivityOrderField],
	parentType any,
	parentID gid.GID,
) *TaskActivityConnection {
	var edges = make([]*TaskActivityEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewTaskActivityEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &TaskActivityConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewTaskActivityEdge(
	e *coredata.TaskActivity,
	orderBy coredata.TaskActivityOrderField,
) *TaskActivityEdge {
	return &TaskActivityEdge{
		Cursor: e.CursorKey(orderBy),
		Node:   NewTaskActivity(e),
	}
}

func NewTaskActivity(e *coredata.TaskActivity) *TaskActivity {
	node := &TaskActivity{
		ID:           e.ID,
		ActivityType: e.ActivityType,
		Field:        e.Field,
		OldValue:     e.OldValue,
		NewValue:     e.NewValue,
		CreatedAt:    e.CreatedAt,
	}

	if e.ActorID != nil {
		node.Actor = &Profile{
			ID: *e.ActorID,
		}
	}

	return node
}
