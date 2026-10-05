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
	TaskPictureOrderBy OrderBy[coredata.TaskPictureOrderField]

	TaskPictureConnection struct {
		TotalCount int
		Edges      []*TaskPictureEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewTaskPictureConnection(
	p *page.Page[*coredata.TaskPicture, coredata.TaskPictureOrderField],
	parentType any,
	parentID gid.GID,
) *TaskPictureConnection {
	edges := make([]*TaskPictureEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewTaskPictureEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &TaskPictureConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewTaskPictureEdge(
	picture *coredata.TaskPicture,
	orderBy coredata.TaskPictureOrderField,
) *TaskPictureEdge {
	return &TaskPictureEdge{
		Cursor: picture.CursorKey(orderBy),
		Node:   NewTaskPicture(picture),
	}
}

func NewTaskPicture(picture *coredata.TaskPicture) *TaskPicture {
	return &TaskPicture{
		ID: picture.ID,
		Task: &Task{
			ID: picture.TaskID,
		},
		File: &File{
			ID: picture.FileID,
		},
		LinearAssetURL: picture.LinearAssetURL,
		CreatedAt:      picture.CreatedAt,
		UpdatedAt:      picture.UpdatedAt,
	}
}
