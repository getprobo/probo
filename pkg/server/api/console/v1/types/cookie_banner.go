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
	CookieBannerOrderBy OrderBy[coredata.CookieBannerOrderField]

	CookieBannerConnection struct {
		TotalCount int
		Edges      []*CookieBannerEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewCookieBannerConnection(
	p *page.Page[*coredata.CookieBanner, coredata.CookieBannerOrderField],
	parentType any,
	parentID gid.GID,
) *CookieBannerConnection {
	var edges = make([]*CookieBannerEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewCookieBannerEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &CookieBannerConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewCookieBannerEdge(b *coredata.CookieBanner, orderBy coredata.CookieBannerOrderField) *CookieBannerEdge {
	return &CookieBannerEdge{
		Cursor: b.CursorKey(orderBy),
		Node:   NewCookieBanner(b),
	}
}

func NewCookieBanner(b *coredata.CookieBanner) *CookieBanner {
	return &CookieBanner{
		ID: b.ID,
		Organization: &Organization{
			ID: b.OrganizationID,
		},
		Name:              b.Name,
		Origin:            b.Origin,
		State:             b.State,
		PrivacyPolicyURL:  b.PrivacyPolicyURL,
		CookiePolicyURL:   b.CookiePolicyURL,
		ConsentExpiryDays: b.ConsentExpiryDays,
		ConsentMode:       b.ConsentMode,
		ShowBranding:      b.ShowBranding,
		DefaultLanguage:   b.DefaultLanguage,
		CreatedAt:         b.CreatedAt,
		UpdatedAt:         b.UpdatedAt,
	}
}

func NewCookieBannerTranslation(t *coredata.CookieBannerTranslation) *CookieBannerTranslation {
	return &CookieBannerTranslation{
		ID:           t.ID,
		Language:     t.Language,
		Translations: string(t.Translations),
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
	}
}
