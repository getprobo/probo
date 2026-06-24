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
	CookieCategoryOrderBy OrderBy[coredata.CookieCategoryOrderField]

	CookieCategoryFilter struct {
		ExcludeKind *coredata.CookieCategoryKind
	}

	CookieCategoryConnection struct {
		TotalCount int
		Edges      []*CookieCategoryEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filter   *coredata.CookieCategoryFilter
	}
)

func NewCookieCategoryConnection(
	p *page.Page[*coredata.CookieCategory, coredata.CookieCategoryOrderField],
	parentType any,
	parentID gid.GID,
) *CookieCategoryConnection {
	var edges = make([]*CookieCategoryEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewCookieCategoryEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &CookieCategoryConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewCookieCategoryConnectionWithFilter(
	p *page.Page[*coredata.CookieCategory, coredata.CookieCategoryOrderField],
	parentType any,
	parentID gid.GID,
	filter *coredata.CookieCategoryFilter,
) *CookieCategoryConnection {
	conn := NewCookieCategoryConnection(p, parentType, parentID)
	conn.Filter = filter

	return conn
}

func NewCookieCategoryEdge(c *coredata.CookieCategory, orderBy coredata.CookieCategoryOrderField) *CookieCategoryEdge {
	return &CookieCategoryEdge{
		Cursor: c.CursorKey(orderBy),
		Node:   NewCookieCategory(c),
	}
}

func NewCookieCategory(c *coredata.CookieCategory) *CookieCategory {
	gcmConsentTypes := c.GCMConsentTypes
	if gcmConsentTypes == nil {
		gcmConsentTypes = []string{}
	}

	return &CookieCategory{
		ID: c.ID,
		CookieBanner: &CookieBanner{
			ID: c.CookieBannerID,
		},
		Name:            c.Name,
		Slug:            c.Slug,
		Description:     c.Description,
		Kind:            c.Kind,
		Rank:            c.Rank,
		GcmConsentTypes: gcmConsentTypes,
		PosthogConsent:  c.PostHogConsent,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
	}
}
