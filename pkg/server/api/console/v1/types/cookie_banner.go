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
	"fmt"
	"regexp"
	"slices"

	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
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

	CookieBannerTheme struct {
		PrimaryColor           string `json:"primaryColor"`
		PrimaryTextColor       string `json:"primaryTextColor"`
		SecondaryColor         string `json:"secondaryColor"`
		SecondaryTextColor     string `json:"secondaryTextColor"`
		BackgroundColor        string `json:"backgroundColor"`
		TextColor              string `json:"textColor"`
		SecondaryTextBodyColor string `json:"secondaryTextBodyColor"`
		BorderColor            string `json:"borderColor"`
		FontFamily             string `json:"fontFamily"`
		BorderRadius           int    `json:"borderRadius"`
		Position               string `json:"position"`
		RevisitPosition        string `json:"revisitPosition"`
	}

	ConsentAnalytics struct {
		TotalRecords        int `json:"totalRecords"`
		AcceptAllCount      int `json:"acceptAllCount"`
		RejectAllCount      int `json:"rejectAllCount"`
		CustomizeCount      int `json:"customizeCount"`
		AcceptCategoryCount int `json:"acceptCategoryCount"`
		GPCCount            int `json:"gpcCount"`
	}

	UpdateCookieBannerThemeInput struct {
		PrimaryColor           *string `json:"primaryColor,omitempty"`
		PrimaryTextColor       *string `json:"primaryTextColor,omitempty"`
		SecondaryColor         *string `json:"secondaryColor,omitempty"`
		SecondaryTextColor     *string `json:"secondaryTextColor,omitempty"`
		BackgroundColor        *string `json:"backgroundColor,omitempty"`
		TextColor              *string `json:"textColor,omitempty"`
		SecondaryTextBodyColor *string `json:"secondaryTextBodyColor,omitempty"`
		BorderColor            *string `json:"borderColor,omitempty"`
		FontFamily             *string `json:"fontFamily,omitempty"`
		BorderRadius           *int    `json:"borderRadius,omitempty"`
		Position               *string `json:"position,omitempty"`
		RevisitPosition        *string `json:"revisitPosition,omitempty"`
	}
)

var (
	hexColorRegex         = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	validPositions        = []string{"bottom", "bottom-left", "bottom-right", "center"}
	validRevisitPositions = []string{"bottom-left", "bottom-right"}
)

func (t *UpdateCookieBannerThemeInput) Validate() error {
	var errs validator.ValidationErrors

	colorFields := []struct {
		field string
		value *string
	}{
		{"theme.primaryColor", t.PrimaryColor},
		{"theme.primaryTextColor", t.PrimaryTextColor},
		{"theme.secondaryColor", t.SecondaryColor},
		{"theme.secondaryTextColor", t.SecondaryTextColor},
		{"theme.backgroundColor", t.BackgroundColor},
		{"theme.textColor", t.TextColor},
		{"theme.secondaryTextBodyColor", t.SecondaryTextBodyColor},
		{"theme.borderColor", t.BorderColor},
	}

	for _, cf := range colorFields {
		if cf.value != nil && !hexColorRegex.MatchString(*cf.value) {
			errs = append(errs, &validator.ValidationError{
				Field:   cf.field,
				Code:    validator.ErrorCodeInvalidFormat,
				Message: "must be a valid hex color (e.g. #ff0000)",
			})
		}
	}

	if t.Position != nil && !slices.Contains(validPositions, *t.Position) {
		errs = append(errs, &validator.ValidationError{
			Field:   "theme.position",
			Code:    validator.ErrorCodeInvalidFormat,
			Message: "must be one of: bottom, bottom-left, bottom-right, center",
		})
	}

	if t.RevisitPosition != nil && !slices.Contains(validRevisitPositions, *t.RevisitPosition) {
		errs = append(errs, &validator.ValidationError{
			Field:   "theme.revisitPosition",
			Code:    validator.ErrorCodeInvalidFormat,
			Message: "must be one of: bottom-left, bottom-right",
		})
	}

	if t.BorderRadius != nil && (*t.BorderRadius < 0 || *t.BorderRadius > 24) {
		errs = append(errs, &validator.ValidationError{
			Field:   "theme.borderRadius",
			Code:    validator.ErrorCodeInvalidFormat,
			Message: "must be between 0 and 24",
		})
	}

	if t.FontFamily != nil && len(*t.FontFamily) > 200 {
		errs = append(errs, &validator.ValidationError{
			Field:   "theme.fontFamily",
			Code:    validator.ErrorCodeInvalidFormat,
			Message: "must be 200 characters or fewer",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func NewCookieBannerConnection(
	p *page.Page[*coredata.CookieBanner, coredata.CookieBannerOrderField],
	resolver any,
	parentID gid.GID,
	baseURL *baseurl.BaseURL,
) *CookieBannerConnection {
	edges := make([]*CookieBannerEdge, len(p.Data))
	for i, banner := range p.Data {
		edges[i] = NewCookieBannerEdge(banner, p.Cursor.OrderBy.Field, baseURL)
	}

	return &CookieBannerConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: resolver,
		ParentID: parentID,
	}
}

func NewCookieBannerEdge(
	banner *coredata.CookieBanner,
	orderField coredata.CookieBannerOrderField,
	baseURL *baseurl.BaseURL,
) *CookieBannerEdge {
	return &CookieBannerEdge{
		Node:   NewCookieBanner(banner, baseURL),
		Cursor: banner.CursorKey(orderField),
	}
}

func NewCookieBanner(banner *coredata.CookieBanner, baseURL *baseurl.BaseURL) *CookieBanner {
	widgetURL := baseURL.WithPath("/api/cookie-banner/v1/widget.js").MustString()

	return &CookieBanner{
		ID:                   banner.ID,
		Name:                 banner.Name,
		Domain:               banner.Domain,
		State:                banner.State,
		Title:                banner.Title,
		Description:          banner.Description,
		AcceptAllLabel:       banner.AcceptAllLabel,
		RejectAllLabel:       banner.RejectAllLabel,
		SavePreferencesLabel: banner.SavePreferencesLabel,
		PrivacyPolicyURL:     banner.PrivacyPolicyURL,
		ConsentExpiryDays:    banner.ConsentExpiryDays,
		ConsentMode:          banner.ConsentMode,
		Version:              banner.Version,
		EmbedSnippet:         fmt.Sprintf(`<script src="%s" data-banner-id="%s" defer></script>`, widgetURL, banner.ID),
		CreatedAt:            banner.CreatedAt,
		UpdatedAt:            banner.UpdatedAt,
	}
}

var defaultCookieBannerTheme = CookieBannerTheme{
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

func NewCookieBannerTheme(raw json.RawMessage) *CookieBannerTheme {
	theme := defaultCookieBannerTheme

	if len(raw) > 0 {
		var override struct {
			PrimaryColor           *string `json:"primaryColor"`
			PrimaryTextColor       *string `json:"primaryTextColor"`
			SecondaryColor         *string `json:"secondaryColor"`
			SecondaryTextColor     *string `json:"secondaryTextColor"`
			BackgroundColor        *string `json:"backgroundColor"`
			TextColor              *string `json:"textColor"`
			SecondaryTextBodyColor *string `json:"secondaryTextBodyColor"`
			BorderColor            *string `json:"borderColor"`
			FontFamily             *string `json:"fontFamily"`
			BorderRadius           *int    `json:"borderRadius"`
			Position               *string `json:"position"`
			RevisitPosition        *string `json:"revisitPosition"`
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

func MergeThemeInput(existing json.RawMessage, input *UpdateCookieBannerThemeInput) json.RawMessage {
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
