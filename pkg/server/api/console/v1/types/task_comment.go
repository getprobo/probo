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
	TaskCommentOrderBy OrderBy[coredata.TaskCommentOrderField]

	TaskCommentConnection struct {
		TotalCount int
		Edges      []*TaskCommentEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewTaskCommentConnection(
	p *page.Page[*coredata.TaskComment, coredata.TaskCommentOrderField],
	parentType any,
	parentID gid.GID,
) *TaskCommentConnection {
	var edges = make([]*TaskCommentEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewTaskCommentEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &TaskCommentConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewTaskCommentEdge(
	c *coredata.TaskComment,
	orderBy coredata.TaskCommentOrderField,
) *TaskCommentEdge {
	return &TaskCommentEdge{
		Cursor: c.CursorKey(orderBy),
		Node:   NewTaskComment(c),
	}
}

func NewTaskComment(c *coredata.TaskComment) *TaskComment {
	return &TaskComment{
		ID: c.ID,
		Task: &Task{
			ID: c.TaskID,
		},
		Owner: &Profile{
			ID: c.OwnerID,
		},
		Content:   c.Content,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}
