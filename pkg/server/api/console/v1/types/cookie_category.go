// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	resolver any,
	parentID gid.GID,
) *CookieCategoryConnection {
	edges := make([]*CookieCategoryEdge, len(p.Data))
	for i, category := range p.Data {
		edges[i] = NewCookieCategoryEdge(category, p.Cursor.OrderBy.Field)
	}

	return &CookieCategoryConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: resolver,
		ParentID: parentID,
	}
}

func NewCookieCategoryEdge(
	category *coredata.CookieCategory,
	orderField coredata.CookieCategoryOrderField,
) *CookieCategoryEdge {
	return &CookieCategoryEdge{
		Node:   NewCookieCategory(category),
		Cursor: category.CursorKey(orderField),
	}
}

func NewCookieCategory(category *coredata.CookieCategory) *CookieCategory {
	cookies := make([]*coredata.CookieItem, 0, len(category.Cookies))
	for i := range category.Cookies {
		cookies = append(cookies, &category.Cookies[i])
	}

	return &CookieCategory{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		Required:    category.Required,
		Rank:        category.Rank,
		Cookies:     cookies,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	}
}
