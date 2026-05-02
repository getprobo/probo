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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestSnapshotsEqual(t *testing.T) {
	t.Parallel()

	baseSnapshot := func() coredata.CookieBannerVersionSnapshot {
		policy := "https://example.com/privacy"
		maxAge := 3600
		return coredata.CookieBannerVersionSnapshot{
			PrivacyPolicyURL:  &policy,
			CookiePolicyURL:   "https://example.com/cookies",
			ConsentExpiryDays: 180,
			ConsentMode:       "OPT_IN",
			DefaultLanguage:   "en",
			Categories: []coredata.CookieBannerVersionSnapshotCategory{
				{
					Name:        "Analytics",
					Slug:        "analytics",
					Description: "Analytics cookies",
					Kind:        coredata.CookieCategoryKindNormal,
					Cookies: coredata.CookieItems{
						{Name: "_ga", MaxAgeSeconds: &maxAge, Description: "Google Analytics"},
					},
					GCMConsentTypes: []string{"analytics_storage"},
					PostHogConsent:  false,
				},
			},
		}
	}

	t.Run("two identical snapshots are equal", func(t *testing.T) {
		t.Parallel()

		a := baseSnapshot()
		b := baseSnapshot()

		assert.True(t, snapshotsEqual(a, b))
	})

	t.Run("snapshot equals itself after json roundtrip", func(t *testing.T) {
		t.Parallel()

		a := baseSnapshot()

		raw, err := json.Marshal(a)
		require.NoError(t, err)

		var b coredata.CookieBannerVersionSnapshot
		require.NoError(t, json.Unmarshal(raw, &b))

		assert.True(t, snapshotsEqual(a, b))
	})

	t.Run("differing CookiePolicyURL is not equal", func(t *testing.T) {
		t.Parallel()

		a := baseSnapshot()
		b := baseSnapshot()
		b.CookiePolicyURL = "https://other.example.com/cookies"

		assert.False(t, snapshotsEqual(a, b))
	})

	t.Run("differing category name is not equal", func(t *testing.T) {
		t.Parallel()

		a := baseSnapshot()
		b := baseSnapshot()
		b.Categories[0].Name = "Tracking"

		assert.False(t, snapshotsEqual(a, b))
	})

	t.Run("differing GCMConsentTypes order is not equal", func(t *testing.T) {
		t.Parallel()

		a := baseSnapshot()
		a.Categories[0].GCMConsentTypes = []string{"analytics_storage", "ad_storage"}
		b := baseSnapshot()
		b.Categories[0].GCMConsentTypes = []string{"ad_storage", "analytics_storage"}

		assert.False(t, snapshotsEqual(a, b))
	})

	t.Run("nil vs set PrivacyPolicyURL is not equal", func(t *testing.T) {
		t.Parallel()

		a := baseSnapshot()
		b := baseSnapshot()
		b.PrivacyPolicyURL = nil

		assert.False(t, snapshotsEqual(a, b))
	})

	t.Run("zero-value snapshots are equal", func(t *testing.T) {
		t.Parallel()

		a := coredata.CookieBannerVersionSnapshot{}
		b := coredata.CookieBannerVersionSnapshot{}

		assert.True(t, snapshotsEqual(a, b))
	})
}

func TestBuildSnapshot_RankInvariant(t *testing.T) {
	t.Parallel()

	tenant := gid.NewTenantID()
	bannerID := gid.New(tenant, coredata.CookieBannerEntityType)

	necessaryID := gid.New(tenant, coredata.CookieCategoryEntityType)
	analyticsID := gid.New(tenant, coredata.CookieCategoryEntityType)
	advertisingID := gid.New(tenant, coredata.CookieCategoryEntityType)
	uncategorisedID := gid.New(tenant, coredata.CookieCategoryEntityType)

	mkCategories := func(necessaryRank, analyticsRank, advertisingRank, uncategorisedRank int) coredata.CookieCategories {
		return coredata.CookieCategories{
			{
				ID:              necessaryID,
				CookieBannerID:  bannerID,
				Name:            "Necessary",
				Slug:            "necessary",
				Description:     "Required.",
				Kind:            coredata.CookieCategoryKindNecessary,
				Rank:            necessaryRank,
				GCMConsentTypes: []string{"security_storage"},
			},
			{
				ID:              analyticsID,
				CookieBannerID:  bannerID,
				Name:            "Analytics",
				Slug:            "analytics",
				Description:     "Analytics.",
				Kind:            coredata.CookieCategoryKindNormal,
				Rank:            analyticsRank,
				GCMConsentTypes: []string{"analytics_storage"},
			},
			{
				ID:              advertisingID,
				CookieBannerID:  bannerID,
				Name:            "Advertising",
				Slug:            "advertising",
				Description:     "Ads.",
				Kind:            coredata.CookieCategoryKindNormal,
				Rank:            advertisingRank,
				GCMConsentTypes: []string{"ad_storage"},
			},
			{
				ID:              uncategorisedID,
				CookieBannerID:  bannerID,
				Name:            "Uncategorised",
				Slug:            "uncategorised",
				Description:     "Misc.",
				Kind:            coredata.CookieCategoryKindUncategorised,
				Rank:            uncategorisedRank,
				GCMConsentTypes: nil,
			},
		}
	}

	banner := &coredata.CookieBanner{
		ID:                bannerID,
		CookiePolicyURL:   "https://example.com/cookies",
		ConsentExpiryDays: 365,
		ConsentMode:       coredata.CookieConsentModeOptIn,
		DefaultLanguage:   "en",
	}

	t.Run("snapshot is identical regardless of rank values", func(t *testing.T) {
		t.Parallel()

		original := buildSnapshot(banner, mkCategories(0, 1, 2, 3), nil)
		shuffled := buildSnapshot(banner, mkCategories(99, 50, 25, 10), nil)

		assert.True(t, snapshotsEqual(original, shuffled), "rank changes must not affect the snapshot")
	})

	t.Run("snapshot is identical regardless of input slice order", func(t *testing.T) {
		t.Parallel()

		ordered := mkCategories(0, 1, 2, 3)
		reversed := coredata.CookieCategories{ordered[3], ordered[2], ordered[1], ordered[0]}

		a := buildSnapshot(banner, ordered, nil)
		b := buildSnapshot(banner, reversed, nil)

		assert.True(t, snapshotsEqual(a, b))
	})

	t.Run("Necessary comes first and Uncategorised comes last", func(t *testing.T) {
		t.Parallel()

		snap := buildSnapshot(banner, mkCategories(0, 1, 2, 3), nil)

		require.Len(t, snap.Categories, 4)
		assert.Equal(t, coredata.CookieCategoryKindNecessary, snap.Categories[0].Kind)
		assert.Equal(t, coredata.CookieCategoryKindUncategorised, snap.Categories[3].Kind)
	})
}
