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
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	TagOrderBy OrderBy[coredata.TagOrderField]

	TagAssignmentOrderBy OrderBy[coredata.TagAssignmentOrderField]

	TagConnection struct {
		TotalCount int
		Edges      []*TagEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}

	TagAssignmentConnection struct {
		TotalCount int
		Edges      []*TagAssignmentEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}

	// ResourceID is the tagged resource. GraphQL exposes it as resource, which
	// resolves that GID through Query.node.
	TagAssignment struct {
		ID         gid.GID
		ResourceID gid.GID
		CreatedAt  time.Time
	}
)

func NewTagConnection(
	p *page.Page[*coredata.Tag, coredata.TagOrderField],
	parentType any,
	parentID gid.GID,
) *TagConnection {
	var edges = make([]*TagEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewTagEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &TagConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewTagEdge(tag *coredata.Tag, orderBy coredata.TagOrderField) *TagEdge {
	return &TagEdge{
		Cursor: tag.CursorKey(orderBy),
		Node:   NewTag(tag),
	}
}

func NewTag(tag *coredata.Tag) *Tag {
	return &Tag{
		ID: tag.ID,
		Organization: &Organization{
			ID: tag.OrganizationID,
		},
		Key:       tag.Key,
		Value:     tag.Value,
		Color:     tag.Color,
		CreatedAt: tag.CreatedAt,
		UpdatedAt: tag.UpdatedAt,
	}
}

func NewTagAssignmentConnection(
	p *page.Page[*coredata.TagAssignment, coredata.TagAssignmentOrderField],
	parentType any,
	parentID gid.GID,
) *TagAssignmentConnection {
	var edges = make([]*TagAssignmentEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewTagAssignmentEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &TagAssignmentConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewTagAssignmentEdge(
	assignment *coredata.TagAssignment,
	orderBy coredata.TagAssignmentOrderField,
) *TagAssignmentEdge {
	return &TagAssignmentEdge{
		Cursor: assignment.CursorKey(orderBy),
		Node:   NewTagAssignment(assignment),
	}
}

func NewTagAssignment(assignment *coredata.TagAssignment) *TagAssignment {
	return &TagAssignment{
		ID:         assignment.ID,
		ResourceID: assignment.ResourceID,
		CreatedAt:  assignment.CreatedAt,
	}
}
