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

package cookiebanner

import "go.probo.inc/probo/pkg/coredata"

var defaultCategories = []struct {
	Name        string
	Description string
	Kind        coredata.CookieCategoryKind
	Rank        int
}{
	{"Necessary", "Essential cookies required for the website to function properly.", coredata.CookieCategoryKindNecessary, 0},
	{"Analytics", "Cookies that help understand how visitors interact with the website.", coredata.CookieCategoryKindNormal, 1},
	{"Advertising", "Cookies used to deliver relevant advertisements and track campaigns.", coredata.CookieCategoryKindNormal, 2},
	{"Functional", "Cookies that enable enhanced functionality and personalization.", coredata.CookieCategoryKindNormal, 3},
	{"Uncategorised", "Cookies that have not been assigned to a category yet.", coredata.CookieCategoryKindUncategorised, 4},
}

var defaultUIStrings = map[string]string{
	"banner_title":             "Cookie Preferences",
	"banner_description":       "We use cookies to improve your experience and analyze site traffic. {{privacy_policy_link}}",
	"button_accept_all":        "Accept all",
	"button_reject_all":        "Reject all",
	"button_customize":         "Customize",
	"button_save":              "Save preferences",
	"panel_title":              "Customise Preferences",
	"panel_description":        "Choose which cookie categories to allow. {{necessary_category}} cookies are always active as they are needed for the site to work.",
	"label_description":        "Description: {{value}}",
	"label_duration":           "Duration: {{value}}",
	"aria_close":               "Close",
	"aria_show_details":        "Show cookie details",
	"aria_hide_details":        "Hide cookie details",
	"aria_cookie_settings":     "Cookie settings",
	"privacy_policy_link_text": "Privacy Policy",
	"placeholder_text":         "This content requires {{category}} cookies.",
	"placeholder_button":       "Manage cookie preferences",
	"duration_year_one":        "{{count}} year",
	"duration_year_other":      "{{count}} years",
	"duration_month_one":       "{{count}} month",
	"duration_month_other":     "{{count}} months",
	"duration_week_one":        "{{count}} week",
	"duration_week_other":      "{{count}} weeks",
	"duration_day_one":         "{{count}} day",
	"duration_day_other":       "{{count}} days",
	"duration_hour_one":        "{{count}} hour",
	"duration_hour_other":      "{{count}} hours",
	"duration_minute_one":      "{{count}} minute",
	"duration_minute_other":    "{{count}} minutes",
	"duration_session":         "session",
}
