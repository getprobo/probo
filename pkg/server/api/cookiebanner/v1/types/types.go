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

	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

type (
	ThemeResponse struct {
		PrimaryColor           string `json:"primary_color"`
		PrimaryTextColor       string `json:"primary_text_color"`
		SecondaryColor         string `json:"secondary_color"`
		SecondaryTextColor     string `json:"secondary_text_color"`
		BackgroundColor        string `json:"background_color"`
		TextColor              string `json:"text_color"`
		SecondaryTextBodyColor string `json:"secondary_text_body_color"`
		BorderColor            string `json:"border_color"`
		FontFamily             string `json:"font_family"`
		BorderRadius           int    `json:"border_radius"`
		Position               string `json:"position"`
		RevisitPosition        string `json:"revisit_position"`
	}

	CookieItemResponse struct {
		Name        string `json:"name"`
		Duration    string `json:"duration"`
		Description string `json:"description"`
	}

	CategoryResponse struct {
		ID          string               `json:"id"`
		Name        string               `json:"name"`
		Description string               `json:"description"`
		Required    bool                 `json:"required"`
		Rank        int                  `json:"rank"`
		Cookies     []CookieItemResponse `json:"cookies"`
	}

	ConfigResponse struct {
		ID                   string             `json:"id"`
		Title                string             `json:"title"`
		Description          string             `json:"description"`
		AcceptAllLabel       string             `json:"accept_all_label"`
		RejectAllLabel       string             `json:"reject_all_label"`
		SavePreferencesLabel string             `json:"save_preferences_label"`
		PrivacyPolicyURL     string             `json:"privacy_policy_url"`
		ConsentExpiryDays    int                `json:"consent_expiry_days"`
		ConsentMode          string             `json:"consent_mode"`
		Version              int                `json:"version"`
		Categories           []CategoryResponse `json:"categories"`
		Theme                *ThemeResponse     `json:"theme,omitempty"`
	}
)

var defaultTheme = ThemeResponse{
	PrimaryColor:           "#2563eb",
	PrimaryTextColor:       "#ffffff",
	SecondaryColor:         "#1a1a1a",
	SecondaryTextColor:     "#ffffff",
	BackgroundColor:        "#ffffff",
	TextColor:              "#1a1a1a",
	SecondaryTextBodyColor: "#4b5563",
	BorderColor:            "#e5e7eb",
	FontFamily:             "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif",
	BorderRadius:           8,
	Position:               "bottom",
	RevisitPosition:        "bottom-left",
}

func NewCategoryResponse(cat *coredata.CookieCategory) CategoryResponse {
	cookies := make([]CookieItemResponse, 0, len(cat.Cookies))
	for _, c := range cat.Cookies {
		cookies = append(cookies, CookieItemResponse{
			Name:        c.Name,
			Duration:    c.Duration,
			Description: c.Description,
		})
	}

	return CategoryResponse{
		ID:          cat.ID.String(),
		Name:        cat.Name,
		Description: cat.Description,
		Required:    cat.Required,
		Rank:        cat.Rank,
		Cookies:     cookies,
	}
}

func newThemeResponse(raw json.RawMessage) *ThemeResponse {
	theme := defaultTheme

	if len(raw) > 0 {
		var override struct {
			PrimaryColor           *string `json:"primary_color"`
			PrimaryTextColor       *string `json:"primary_text_color"`
			SecondaryColor         *string `json:"secondary_color"`
			SecondaryTextColor     *string `json:"secondary_text_color"`
			BackgroundColor        *string `json:"background_color"`
			TextColor              *string `json:"text_color"`
			SecondaryTextBodyColor *string `json:"secondary_text_body_color"`
			BorderColor            *string `json:"border_color"`
			FontFamily             *string `json:"font_family"`
			BorderRadius           *int    `json:"border_radius"`
			Position               *string `json:"position"`
			RevisitPosition        *string `json:"revisit_position"`
		}
		if err := json.Unmarshal(raw, &override); err == nil {
			if override.PrimaryColor != nil {
				theme.PrimaryColor = *override.PrimaryColor
			}
			if override.PrimaryTextColor != nil {
				theme.PrimaryTextColor = *override.PrimaryTextColor
			}
			if override.SecondaryColor != nil {
				theme.SecondaryColor = *override.SecondaryColor
			}
			if override.SecondaryTextColor != nil {
				theme.SecondaryTextColor = *override.SecondaryTextColor
			}
			if override.BackgroundColor != nil {
				theme.BackgroundColor = *override.BackgroundColor
			}
			if override.TextColor != nil {
				theme.TextColor = *override.TextColor
			}
			if override.SecondaryTextBodyColor != nil {
				theme.SecondaryTextBodyColor = *override.SecondaryTextBodyColor
			}
			if override.BorderColor != nil {
				theme.BorderColor = *override.BorderColor
			}
			if override.FontFamily != nil {
				theme.FontFamily = *override.FontFamily
			}
			if override.BorderRadius != nil {
				theme.BorderRadius = *override.BorderRadius
			}
			if override.Position != nil {
				theme.Position = *override.Position
			}
			if override.RevisitPosition != nil {
				theme.RevisitPosition = *override.RevisitPosition
			}
		}
	}

	return &theme
}

func NewConfigResponse(banner *coredata.CookieBanner, categories coredata.CookieCategories) ConfigResponse {
	cats := make([]CategoryResponse, 0, len(categories))
	for _, cat := range categories {
		cats = append(cats, NewCategoryResponse(cat))
	}

	return ConfigResponse{
		ID:                   banner.ID.String(),
		Title:                banner.Title,
		Description:          banner.Description,
		AcceptAllLabel:       banner.AcceptAllLabel,
		RejectAllLabel:       banner.RejectAllLabel,
		SavePreferencesLabel: banner.SavePreferencesLabel,
		PrivacyPolicyURL:     banner.PrivacyPolicyURL,
		ConsentExpiryDays:    banner.ConsentExpiryDays,
		ConsentMode:          strings.ReplaceAll(strings.ToLower(string(banner.ConsentMode)), "_", "-"),
		Version:              banner.Version,
		Categories:           cats,
		Theme:                newThemeResponse(banner.Theme),
	}
}
