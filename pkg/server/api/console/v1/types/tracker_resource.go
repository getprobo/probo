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
	TrackerResourceOrderBy OrderBy[coredata.TrackerResourceOrderField]

	TrackerResourceConnection struct {
		TotalCount int
		Edges      []*TrackerResourceEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filter   *TrackerResourceFilter
	}

	TrackerResourceFilter struct {
		Query *string
		Type  *coredata.TrackerResourceType
	}
)

func NewTrackerResourceConnection(
	p *page.Page[*coredata.TrackerResource, coredata.TrackerResourceOrderField],
	parentType any,
	parentID gid.GID,
) *TrackerResourceConnection {
	edges := make([]*TrackerResourceEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewTrackerResourceEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &TrackerResourceConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewTrackerResourceConnectionWithFilter(
	p *page.Page[*coredata.TrackerResource, coredata.TrackerResourceOrderField],
	parentType any,
	parentID gid.GID,
	filter *TrackerResourceFilter,
) *TrackerResourceConnection {
	conn := NewTrackerResourceConnection(p, parentType, parentID)
	conn.Filter = filter

	return conn
}

func NewTrackerResourceEdge(tr *coredata.TrackerResource, orderBy coredata.TrackerResourceOrderField) *TrackerResourceEdge {
	return &TrackerResourceEdge{
		Cursor: tr.CursorKey(orderBy),
		Node:   NewTrackerResourceNode(tr),
	}
}

func NewTrackerResourceNode(tr *coredata.TrackerResource) *TrackerResource {
	return &TrackerResource{
		ID: tr.ID,
		CookieCategory: &CookieCategory{
			ID: tr.CookieCategoryID,
			CookieBanner: &CookieBanner{
				ID: tr.CookieBannerID,
			},
		},
		Type:           tr.ResourceType,
		Origin:         tr.Origin,
		Path:           tr.Path,
		DisplayName:    tr.DisplayName,
		Description:    tr.Description,
		Excluded:       tr.Excluded,
		LastDetectedAt: tr.LastDetectedAt,
		CreatedAt:      tr.CreatedAt,
		UpdatedAt:      tr.UpdatedAt,
	}
}
