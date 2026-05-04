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

import (
	"bytes"
	"encoding/json"
	"reflect"
	"slices"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// resolveTranslations converts raw DB translations into the resolved map
// used by buildBannerConfig at serve time. Categories are filtered and sorted
// in snapshot order so the positional category translations align.
func resolveTranslations(
	translations coredata.CookieBannerTranslations,
	categories coredata.CookieCategories,
) map[string]coredata.CookieBannerVersionSnapshotTranslation {
	sortConsentCategories(categories)
	return buildSnapshotTranslations(translations, categories)
}

// snapshotsEqual reports whether two version snapshots are visitor-identical.
// buildSnapshot already normalises empty slices and nil maps, so reflect.DeepEqual
// is sufficient and is the single chokepoint we'd extend if we ever wanted to
// ignore particular fields.
func snapshotsEqual(a, b coredata.CookieBannerVersionSnapshot) bool {
	return reflect.DeepEqual(a, b)
}

// snapshotCategoryKindOrder returns a stable weight per Kind so the snapshot
// keeps the visitor-facing layout invariants (NECESSARY first, then NORMAL
// sorted by ID) without depending on the admin-controlled rank.
func snapshotCategoryKindOrder(k coredata.CookieCategoryKind) int {
	switch k {
	case coredata.CookieCategoryKindNecessary:
		return 0
	case coredata.CookieCategoryKindNormal:
		return 1
	default:
		return 2
	}
}

// sortConsentCategories sorts categories in snapshot order: NECESSARY first,
// then NORMAL sorted by ID. The caller is expected to have already excluded
// UNCATEGORISED at load time.
func sortConsentCategories(categories coredata.CookieCategories) {
	slices.SortStableFunc(categories, func(a, b *coredata.CookieCategory) int {
		if d := snapshotCategoryKindOrder(a.Kind) - snapshotCategoryKindOrder(b.Kind); d != 0 {
			return d
		}
		return bytes.Compare(a.ID[:], b.ID[:])
	})
}

func buildSnapshot(
	banner *coredata.CookieBanner,
	categories coredata.CookieCategories,
	allPatterns coredata.CookiePatterns,
) coredata.CookieBannerVersionSnapshot {
	sortConsentCategories(categories)

	cookiesByCategory := make(map[gid.GID]coredata.CookieItems)
	for _, p := range allPatterns {
		cookiesByCategory[p.CookieCategoryID] = append(
			cookiesByCategory[p.CookieCategoryID],
			coredata.CookieItem{
				Name:          p.DisplayName,
				MaxAgeSeconds: p.MaxAgeSeconds,
				Description:   p.Description,
			},
		)
	}

	snapshotCategories := make([]coredata.CookieBannerVersionSnapshotCategory, len(categories))
	for i, c := range categories {
		cookies := cookiesByCategory[c.ID]
		if cookies == nil {
			cookies = coredata.CookieItems{}
		}
		gcmConsentTypes := c.GCMConsentTypes
		if gcmConsentTypes == nil {
			gcmConsentTypes = []string{}
		}
		snapshotCategories[i] = coredata.CookieBannerVersionSnapshotCategory{
			Name:            c.Name,
			Slug:            c.Slug,
			Description:     c.Description,
			Kind:            c.Kind,
			Cookies:         cookies,
			GCMConsentTypes: gcmConsentTypes,
			PostHogConsent:  c.PostHogConsent,
		}
	}

	return coredata.CookieBannerVersionSnapshot{
		PrivacyPolicyURL:  banner.PrivacyPolicyURL,
		CookiePolicyURL:   banner.CookiePolicyURL,
		ConsentExpiryDays: banner.ConsentExpiryDays,
		ConsentMode:       string(banner.ConsentMode),
		DefaultLanguage:   banner.DefaultLanguage,
		Categories:        snapshotCategories,
	}
}

func buildSnapshotTranslations(
	translations coredata.CookieBannerTranslations,
	categories coredata.CookieCategories,
) map[string]coredata.CookieBannerVersionSnapshotTranslation {
	if len(translations) == 0 {
		return nil
	}

	result := make(map[string]coredata.CookieBannerVersionSnapshotTranslation, len(translations))

	for _, t := range translations {
		var raw struct {
			Categories map[string]struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			} `json:"categories"`
		}
		_ = json.Unmarshal(t.Translations, &raw)

		ui := make(map[string]string)
		var flat map[string]json.RawMessage
		_ = json.Unmarshal(t.Translations, &flat)
		for k, v := range flat {
			if k == "categories" || k == "cookies" {
				continue
			}
			var s string
			if json.Unmarshal(v, &s) == nil {
				ui[k] = s
			}
		}

		catTranslations := make([]coredata.CookieBannerVersionSnapshotCategoryTranslation, len(categories))
		for i, c := range categories {
			if raw.Categories != nil {
				if ct, ok := raw.Categories[c.ID.String()]; ok {
					catTranslations[i] = coredata.CookieBannerVersionSnapshotCategoryTranslation{
						Name:        ct.Name,
						Description: ct.Description,
					}
					continue
				}
			}
			catTranslations[i] = coredata.CookieBannerVersionSnapshotCategoryTranslation{
				Name:        c.Name,
				Description: c.Description,
			}
		}

		result[t.Language] = coredata.CookieBannerVersionSnapshotTranslation{
			UI:         ui,
			Categories: catTranslations,
		}
	}

	return result
}
