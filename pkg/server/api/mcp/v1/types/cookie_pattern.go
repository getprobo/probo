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
	"go.probo.inc/probo/pkg/page"
)

func NewCookiePattern(p *coredata.CookiePattern) *CookiePattern {
	return &CookiePattern{
		ID:               p.ID,
		OrganizationID:   p.OrganizationID,
		CookieBannerID:   p.CookieBannerID,
		CookieCategoryID: p.CookieCategoryID,
		Pattern:          p.Pattern,
		MatchType:        CookiePatternMatchType(p.MatchType),
		DisplayName:      p.DisplayName,
		MaxAgeSeconds:    p.MaxAgeSeconds,
		Description:      p.Description,
		Source:           CookiePatternSource(p.Source),
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
	}
}

func NewListCookiePatternsOutput(pg *page.Page[*coredata.CookiePattern, coredata.CookiePatternOrderField]) ListCookiePatternsOutput {
	patterns := make([]*CookiePattern, 0, len(pg.Data))
	for _, p := range pg.Data {
		patterns = append(patterns, NewCookiePattern(p))
	}

	var nextCursor *page.CursorKey
	if len(pg.Data) > 0 {
		cursorKey := pg.Data[len(pg.Data)-1].CursorKey(pg.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListCookiePatternsOutput{
		NextCursor:     nextCursor,
		CookiePatterns: patterns,
	}
}
