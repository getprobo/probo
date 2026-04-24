// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
	CookieCategoryOrderBy OrderBy[coredata.CookieCategoryOrderField]

	CookieCategoryConnection struct {
		TotalCount int
		Edges      []*CookieCategoryEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
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
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
	}
}
