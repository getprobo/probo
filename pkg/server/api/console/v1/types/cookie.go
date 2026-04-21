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
	CookieOrderBy OrderBy[coredata.CookieOrderField]

	CookieConnection struct {
		TotalCount int
		Edges      []*CookieEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewCookieConnection(
	p *page.Page[*coredata.Cookie, coredata.CookieOrderField],
	parentType any,
	parentID gid.GID,
) *CookieConnection {
	edges := make([]*CookieEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewCookieEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &CookieConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewCookieEdge(c *coredata.Cookie, orderBy coredata.CookieOrderField) *CookieEdge {
	return &CookieEdge{
		Cursor: c.CursorKey(orderBy),
		Node:   NewCookie(c),
	}
}

func NewCookie(c *coredata.Cookie) *Cookie {
	return &Cookie{
		ID: c.ID,
		CookieCategory: &CookieCategory{
			ID: c.CookieCategoryID,
			CookieBanner: &CookieBanner{
				ID: c.CookieBannerID,
			},
		},
		Name:        c.Name,
		Duration:    c.Duration,
		Description: c.Description,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}
