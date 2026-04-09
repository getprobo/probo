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
	"encoding/json"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/page"
)

func newCookieBannerTheme(raw json.RawMessage) *CookieBannerTheme {
	if len(raw) == 0 {
		return nil
	}

	var theme CookieBannerTheme
	if err := json.Unmarshal(raw, &theme); err != nil {
		return nil
	}

	return &theme
}

func MergeThemeInput(existing json.RawMessage, input *CookieBannerTheme) json.RawMessage {
	if input == nil {
		return existing
	}

	current := make(map[string]any)
	if len(existing) > 0 {
		_ = json.Unmarshal(existing, &current)
	}

	if input.PrimaryColor != nil {
		current["primaryColor"] = *input.PrimaryColor
	}
	if input.PrimaryTextColor != nil {
		current["primaryTextColor"] = *input.PrimaryTextColor
	}
	if input.SecondaryColor != nil {
		current["secondaryColor"] = *input.SecondaryColor
	}
	if input.SecondaryTextColor != nil {
		current["secondaryTextColor"] = *input.SecondaryTextColor
	}
	if input.BackgroundColor != nil {
		current["backgroundColor"] = *input.BackgroundColor
	}
	if input.TextColor != nil {
		current["textColor"] = *input.TextColor
	}
	if input.SecondaryTextBodyColor != nil {
		current["secondaryTextBodyColor"] = *input.SecondaryTextBodyColor
	}
	if input.BorderColor != nil {
		current["borderColor"] = *input.BorderColor
	}
	if input.FontFamily != nil {
		current["fontFamily"] = *input.FontFamily
	}
	if input.BorderRadius != nil {
		current["borderRadius"] = *input.BorderRadius
	}
	if input.Position != nil {
		current["position"] = *input.Position
	}
	if input.RevisitPosition != nil {
		current["revisitPosition"] = *input.RevisitPosition
	}

	data, err := json.Marshal(current)
	if err != nil {
		return existing
	}

	return data
}

func NewCookieBanner(b *coredata.CookieBanner) *CookieBanner {
	return &CookieBanner{
		ID:                   b.ID,
		OrganizationID:       b.OrganizationID,
		Name:                 b.Name,
		Domain:               b.Domain,
		State:                b.State,
		Title:                b.Title,
		Description:          b.Description,
		AcceptAllLabel:       b.AcceptAllLabel,
		RejectAllLabel:       b.RejectAllLabel,
		SavePreferencesLabel: b.SavePreferencesLabel,
		PrivacyPolicyURL:     b.PrivacyPolicyURL,
		ConsentExpiryDays:    b.ConsentExpiryDays,
		ConsentMode:          b.ConsentMode,
		Version:              b.Version,
		Theme:                newCookieBannerTheme(b.Theme),
		CreatedAt:            b.CreatedAt,
		UpdatedAt:            b.UpdatedAt,
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

func NewCookieCategory(c *coredata.CookieCategory) *CookieCategory {
	cookies := c.Cookies
	if cookies == nil {
		cookies = make(coredata.CookieItems, 0)
	}

	return &CookieCategory{
		ID:             c.ID,
		CookieBannerID: c.CookieBannerID,
		Name:           c.Name,
		Description:    c.Description,
		Required:       c.Required,
		Rank:           c.Rank,
		Cookies:        cookies,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}

func NewListCookieCategoriesOutput(p *page.Page[*coredata.CookieCategory, coredata.CookieCategoryOrderField]) ListCookieCategoriesOutput {
	categories := make([]*CookieCategory, 0, len(p.Data))
	for _, c := range p.Data {
		categories = append(categories, NewCookieCategory(c))
	}

	var nextCursor *page.CursorKey
	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListCookieCategoriesOutput{
		NextCursor:       nextCursor,
		CookieCategories: categories,
	}
}

func NewConsentRecord(r *coredata.ConsentRecord) *ConsentRecord {
	return &ConsentRecord{
		ID:             r.ID,
		CookieBannerID: r.CookieBannerID,
		VisitorID:      r.VisitorID,
		IPAddress:      r.IPAddress,
		UserAgent:      r.UserAgent,
		ConsentData:    string(r.ConsentData),
		Action:         r.Action,
		BannerVersion:  r.BannerVersion,
		CreatedAt:      r.CreatedAt,
	}
}

func NewListConsentRecordsOutput(p *page.Page[*coredata.ConsentRecord, coredata.ConsentRecordOrderField]) ListConsentRecordsOutput {
	records := make([]*ConsentRecord, 0, len(p.Data))
	for _, r := range p.Data {
		records = append(records, NewConsentRecord(r))
	}

	var nextCursor *page.CursorKey
	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListConsentRecordsOutput{
		NextCursor:     nextCursor,
		ConsentRecords: records,
	}
}
