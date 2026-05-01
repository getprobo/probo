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

func NewCookieBanner(b *coredata.CookieBanner) *CookieBanner {
	return &CookieBanner{
		ID:                b.ID,
		OrganizationID:    b.OrganizationID,
		Name:              b.Name,
		Origin:            b.Origin,
		State:             CookieBannerState(b.State),
		PrivacyPolicyURL:  b.PrivacyPolicyURL,
		CookiePolicyURL:   b.CookiePolicyURL,
		ConsentExpiryDays: b.ConsentExpiryDays,
		ConsentMode:       CookieBannerConsentMode(b.ConsentMode),
		ShowBranding:      b.ShowBranding,
		DefaultLanguage:   b.DefaultLanguage,
		CreatedAt:         b.CreatedAt,
		UpdatedAt:         b.UpdatedAt,
	}
}

func NewListCookieBannersOutput(p *page.Page[*coredata.CookieBanner, coredata.CookieBannerOrderField]) ListCookieBannersOutput {
	banners := make([]*CookieBanner, 0, len(p.Data))
	for _, b := range p.Data {
		banners = append(banners, NewCookieBanner(b))
	}

	var nextCursor *page.CursorKey
	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListCookieBannersOutput{
		NextCursor:    nextCursor,
		CookieBanners: banners,
	}
}

func NewCookieBannerTranslation(t *coredata.CookieBannerTranslation) *CookieBannerTranslation {
	return &CookieBannerTranslation{
		ID:             t.ID,
		CookieBannerID: t.CookieBannerID,
		Language:       t.Language,
		Translations:   string(t.Translations),
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
	}
}
