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
	ResourceTagOrderBy OrderBy[coredata.ResourceTagOrderField]

	ResourceTagConnection struct {
		TotalCount int
		Edges      []*ResourceTagEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewResourceTagConnection(
	p *page.Page[*coredata.ResourceTag, coredata.ResourceTagOrderField],
	parentType any,
	parentID gid.GID,
) *ResourceTagConnection {
	var edges = make([]*ResourceTagEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewResourceTagEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &ResourceTagConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewResourceTagEdge(tag *coredata.ResourceTag, orderBy coredata.ResourceTagOrderField) *ResourceTagEdge {
	return &ResourceTagEdge{
		Cursor: tag.CursorKey(orderBy),
		Node:   NewResourceTag(tag),
	}
}

func NewResourceTag(tag *coredata.ResourceTag) *ResourceTag {
	return &ResourceTag{
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
