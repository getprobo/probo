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

package console_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

// versionInfo is the (version, state) tuple returned by the latestVersion field.
type versionInfo struct {
	Version int
	State   string
}

func latestVersion(t *testing.T, c *testutil.Client, bannerID string) versionInfo {
	t.Helper()

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on CookieBanner {
					latestVersion {
						version
						state
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			LatestVersion *struct {
				Version int    `json:"version"`
				State   string `json:"state"`
			} `json:"latestVersion"`
		} `json:"node"`
	}

	require.NoError(t, c.Execute(query, map[string]any{"id": bannerID}, &result), "latestVersion query failed")
	require.NotNil(t, result.Node.LatestVersion, "expected latestVersion to be present")

	return versionInfo{
		Version: result.Node.LatestVersion.Version,
		State:   result.Node.LatestVersion.State,
	}
}

func publishBanner(t *testing.T, c *testutil.Client, bannerID string) versionInfo {
	t.Helper()

	const query = `
		mutation PublishCookieBannerVersion($input: PublishCookieBannerVersionInput!) {
			publishCookieBannerVersion(input: $input) {
				cookieBannerVersion {
					version
					state
				}
			}
		}
	`

	var result struct {
		PublishCookieBannerVersion struct {
			CookieBannerVersion struct {
				Version int    `json:"version"`
				State   string `json:"state"`
			} `json:"cookieBannerVersion"`
		} `json:"publishCookieBannerVersion"`
	}

	require.NoError(t, c.Execute(query, map[string]any{
		"input": map[string]any{"cookieBannerId": bannerID},
	}, &result), "publishCookieBannerVersion mutation failed")

	return versionInfo{
		Version: result.PublishCookieBannerVersion.CookieBannerVersion.Version,
		State:   result.PublishCookieBannerVersion.CookieBannerVersion.State,
	}
}

func setPatternExcluded(t *testing.T, c *testutil.Client, patternID string, excluded bool) {
	t.Helper()

	const query = `
		mutation UpdateTrackerPattern($input: UpdateTrackerPatternInput!) {
			updateTrackerPattern(input: $input) {
				trackerPattern { id excluded }
			}
		}
	`

	var result struct{}
	require.NoError(t, c.Execute(query, map[string]any{
		"input": map[string]any{
			"trackerPatternId": patternID,
			"excluded":         excluded,
		},
	}, &result), "updateTrackerPattern excluded mutation failed")
}

// upsertTranslation upserts a translation for a banner+language pair and
// returns nothing (we read the version separately).
func upsertTranslation(t *testing.T, c *testutil.Client, bannerID, language, translations string) {
	t.Helper()

	const query = `
		mutation UpsertCookieBannerTranslation($input: UpsertCookieBannerTranslationInput!) {
			upsertCookieBannerTranslation(input: $input) {
				cookieBannerTranslation { id }
			}
		}
	`

	var result struct{}
	require.NoError(t, c.Execute(query, map[string]any{
		"input": map[string]any{
			"cookieBannerId": bannerID,
			"language":       language,
			"translations":   translations,
		},
	}, &result), "upsertCookieBannerTranslation mutation failed")
}

func TestCookieBannerVersioning_NoOpUpdates(t *testing.T) {
	t.Parallel()

	t.Run("UpdateCookieBanner with all original values does not bump version", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner, factory.Attrs{
			"cookiePolicyUrl":   "https://example.com/cookies",
			"consentExpiryDays": 365,
			"consentMode":       "OPT_IN",
		})

		published := publishBanner(t, owner, bannerID)
		require.Equal(t, "PUBLISHED", published.State)
		baseline := published.Version

		const query = `
			mutation UpdateCookieBanner($input: UpdateCookieBannerInput!) {
				updateCookieBanner(input: $input) { cookieBanner { id } }
			}
		`

		var result struct{}
		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"cookieBannerId":    bannerID,
				"cookiePolicyUrl":   "https://example.com/cookies",
				"consentExpiryDays": 365,
				"consentMode":       "OPT_IN",
			},
		}, &result)
		require.NoError(t, err)

		got := latestVersion(t, owner, bannerID)
		assert.Equal(t, baseline, got.Version, "version should not change for no-op banner update")
		assert.Equal(t, "PUBLISHED", got.State, "no draft should be created")
	})

	t.Run("UpdateCookieBanner with only name change does not bump version", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		published := publishBanner(t, owner, bannerID)
		baseline := published.Version

		const query = `
			mutation UpdateCookieBanner($input: UpdateCookieBannerInput!) {
				updateCookieBanner(input: $input) { cookieBanner { id name } }
			}
		`

		var result struct {
			UpdateCookieBanner struct {
				CookieBanner struct {
					Name string `json:"name"`
				} `json:"cookieBanner"`
			} `json:"updateCookieBanner"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"cookieBannerId": bannerID,
				"name":           factory.SafeName("Renamed"),
			},
		}, &result)
		require.NoError(t, err)
		assert.NotEmpty(t, result.UpdateCookieBanner.CookieBanner.Name)

		got := latestVersion(t, owner, bannerID)
		assert.Equal(t, baseline, got.Version, "renaming should not affect the visitor-facing snapshot")
		assert.Equal(t, "PUBLISHED", got.State)
	})

	t.Run("UpdateCookieCategory with all original values does not bump version", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID, factory.Attrs{
			"name":        "Marketing",
			"slug":        "marketing-noop",
			"description": "Marketing cookies",
			"rank":        12,
		})

		published := publishBanner(t, owner, bannerID)
		baseline := published.Version

		const query = `
			mutation UpdateCookieCategory($input: UpdateCookieCategoryInput!) {
				updateCookieCategory(input: $input) { cookieCategory { id } }
			}
		`

		var result struct{}
		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"cookieCategoryId": categoryID,
				"name":             "Marketing",
				"slug":             "marketing-noop",
				"description":      "Marketing cookies",
			},
		}, &result)
		require.NoError(t, err)

		got := latestVersion(t, owner, bannerID)
		assert.Equal(t, baseline, got.Version)
		assert.Equal(t, "PUBLISHED", got.State)
	})

	t.Run("ReorderCookieCategory with current rank does not bump version", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID, factory.Attrs{
			"slug": "reorder-noop",
			"rank": 7,
		})

		published := publishBanner(t, owner, bannerID)
		baseline := published.Version

		const query = `
			mutation ReorderCookieCategory($input: ReorderCookieCategoryInput!) {
				reorderCookieCategory(input: $input) { cookieBanner { id } }
			}
		`

		var result struct{}
		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"cookieCategoryId": categoryID,
				"rank":             7,
			},
		}, &result)
		require.NoError(t, err)

		got := latestVersion(t, owner, bannerID)
		assert.Equal(t, baseline, got.Version)
		assert.Equal(t, "PUBLISHED", got.State)
	})

	t.Run("ReorderCookieCategory with new rank does not bump version", func(t *testing.T) {
		// Rank is admin-only metadata; the snapshot is sorted by
		// (Kind weight, ID), so a rank change is invisible to visitors.
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID, factory.Attrs{
			"slug": "reorder-real",
			"rank": 10,
		})

		published := publishBanner(t, owner, bannerID)
		baseline := published.Version

		const query = `
			mutation ReorderCookieCategory($input: ReorderCookieCategoryInput!) {
				reorderCookieCategory(input: $input) { cookieBanner { id } }
			}
		`

		var result struct{}
		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"cookieCategoryId": categoryID,
				"rank":             42,
			},
		}, &result)
		require.NoError(t, err)

		got := latestVersion(t, owner, bannerID)
		assert.Equal(t, baseline, got.Version, "real rank change must not bump the version")
		assert.Equal(t, "PUBLISHED", got.State)
	})

	t.Run("UpdateTrackerPattern on visible pattern with same value does not bump version", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID, factory.Attrs{"slug": "visible-noop"})
		patternID := factory.CreateTrackerPattern(owner, categoryID, factory.Attrs{
			"displayName": "GA Tracker",
			"description": "Original description",
		})

		published := publishBanner(t, owner, bannerID)
		baseline := published.Version

		const query = `
			mutation UpdateTrackerPattern($input: UpdateTrackerPatternInput!) {
				updateTrackerPattern(input: $input) { trackerPattern { id } }
			}
		`

		var result struct{}
		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"trackerPatternId": patternID,
				"displayName":      "GA Tracker",
				"description":      "Original description",
			},
		}, &result)
		require.NoError(t, err)

		got := latestVersion(t, owner, bannerID)
		assert.Equal(t, baseline, got.Version)
		assert.Equal(t, "PUBLISHED", got.State)
	})
}

func TestCookieBannerVersioning_ExcludedPattern(t *testing.T) {
	t.Parallel()

	t.Run("Update on excluded pattern does not bump version", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID, factory.Attrs{"slug": "excl-update"})
		patternID := factory.CreateTrackerPattern(owner, categoryID, factory.Attrs{
			"displayName": "Original",
		})

		setPatternExcluded(t, owner, patternID, true)
		published := publishBanner(t, owner, bannerID)
		baseline := published.Version

		const query = `
			mutation UpdateTrackerPattern($input: UpdateTrackerPatternInput!) {
				updateTrackerPattern(input: $input) {
					trackerPattern { id displayName description }
				}
			}
		`

		var result struct {
			UpdateTrackerPattern struct {
				TrackerPattern struct {
					DisplayName string `json:"displayName"`
					Description string `json:"description"`
				} `json:"trackerPattern"`
			} `json:"updateTrackerPattern"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"trackerPatternId": patternID,
				"displayName":      "Renamed Excluded",
				"description":      "Now with notes",
			},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, "Renamed Excluded", result.UpdateTrackerPattern.TrackerPattern.DisplayName)
		assert.Equal(t, "Now with notes", result.UpdateTrackerPattern.TrackerPattern.Description)

		got := latestVersion(t, owner, bannerID)
		assert.Equal(t, baseline, got.Version, "excluded pattern fields are invisible to visitors")
		assert.Equal(t, "PUBLISHED", got.State)
	})

	t.Run("Delete of excluded pattern does not bump version", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID, factory.Attrs{"slug": "excl-delete"})
		patternID := factory.CreateTrackerPattern(owner, categoryID)

		setPatternExcluded(t, owner, patternID, true)
		published := publishBanner(t, owner, bannerID)
		baseline := published.Version

		const query = `
			mutation DeleteTrackerPattern($input: DeleteTrackerPatternInput!) {
				deleteTrackerPattern(input: $input) {
					deletedTrackerPatternId
				}
			}
		`

		var result struct{}
		err := owner.Execute(query, map[string]any{
			"input": map[string]any{"trackerPatternId": patternID},
		}, &result)
		require.NoError(t, err)

		got := latestVersion(t, owner, bannerID)
		assert.Equal(t, baseline, got.Version)
		assert.Equal(t, "PUBLISHED", got.State)
	})

	t.Run("Move of excluded pattern does not bump version", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryA := factory.CreateCookieCategory(owner, bannerID, factory.Attrs{"slug": "excl-move-a"})
		categoryB := factory.CreateCookieCategory(owner, bannerID, factory.Attrs{"slug": "excl-move-b"})
		patternID := factory.CreateTrackerPattern(owner, categoryA)

		setPatternExcluded(t, owner, patternID, true)
		published := publishBanner(t, owner, bannerID)
		baseline := published.Version

		const query = `
			mutation MoveTrackerPatternToCategory($input: MoveTrackerPatternToCategoryInput!) {
				moveTrackerPatternToCategory(input: $input) {
					trackerPattern { id }
				}
			}
		`

		var result struct{}
		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"trackerPatternId":       patternID,
				"targetCookieCategoryId": categoryB,
			},
		}, &result)
		require.NoError(t, err)

		got := latestVersion(t, owner, bannerID)
		assert.Equal(t, baseline, got.Version)
		assert.Equal(t, "PUBLISHED", got.State)
	})
}

func TestCookieBannerVersioning_TranslationChangesNeverBump(t *testing.T) {
	t.Parallel()

	t.Run("re-upserting identical JSON does not bump version", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)

		const customJSON = `{"banner_title":"Cookie Bar","button_accept_all":"Accept"}`
		upsertTranslation(t, owner, bannerID, "it", customJSON)
		published := publishBanner(t, owner, bannerID)
		baseline := published.Version

		upsertTranslation(t, owner, bannerID, "it", customJSON)

		got := latestVersion(t, owner, bannerID)
		assert.Equal(t, baseline, got.Version, "re-upserting identical JSON should not bump the version")
		assert.Equal(t, "PUBLISHED", got.State)
	})

	t.Run("changing translation content does not bump version", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)

		upsertTranslation(t, owner, bannerID, "it", `{"banner_title":"Cookie Bar"}`)
		published := publishBanner(t, owner, bannerID)
		baseline := published.Version

		upsertTranslation(t, owner, bannerID, "it", `{"banner_title":"Barra dei Cookie"}`)

		got := latestVersion(t, owner, bannerID)
		assert.Equal(t, baseline, got.Version, "translation content changes should not bump the version")
		assert.Equal(t, "PUBLISHED", got.State)
	})

	t.Run("adding a new language does not bump version", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		published := publishBanner(t, owner, bannerID)
		baseline := published.Version

		upsertTranslation(t, owner, bannerID, "fr", `{"banner_title":"Bandeau cookies"}`)

		got := latestVersion(t, owner, bannerID)
		assert.Equal(t, baseline, got.Version, "adding a new language should not bump the version")
		assert.Equal(t, "PUBLISHED", got.State)
	})
}

func reportDetectedCookies(t *testing.T, c *testutil.Client, bannerID string, names ...string) {
	t.Helper()

	type entry struct {
		Name   string `json:"name"`
		Source string `json:"source"`
	}
	cookies := make([]entry, len(names))
	for i, n := range names {
		cookies[i] = entry{Name: n, Source: "script"}
	}

	body, err := json.Marshal(map[string]any{"cookies": cookies})
	require.NoError(t, err)

	endpoint := fmt.Sprintf("%s/api/cookie-banner/v1/%s/detected-cookies", c.BaseURL(), bannerID)
	resp, err := c.HTTPClient().Post(endpoint, "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusNoContent, resp.StatusCode,
		"detected-cookies endpoint should return 204")
}

func TestCookieBannerVersioning_DetectedCookiesNeverBump(t *testing.T) {
	t.Parallel()

	t.Run("reporting detected cookies does not bump version", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		published := publishBanner(t, owner, bannerID)
		baseline := published.Version

		reportDetectedCookies(t, owner, bannerID, "_unknown_cookie", "_another")

		got := latestVersion(t, owner, bannerID)
		assert.Equal(t, baseline, got.Version, "detected cookies should not bump the version")
		assert.Equal(t, "PUBLISHED", got.State)
	})
}

func TestCookieBannerVersioning_RealChangesStillBumpVersion(t *testing.T) {
	t.Parallel()

	t.Run("UpdateCookieBanner consent change creates a new draft", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner, factory.Attrs{
			"consentExpiryDays": 365,
		})
		published := publishBanner(t, owner, bannerID)
		baseline := published.Version

		const query = `
			mutation UpdateCookieBanner($input: UpdateCookieBannerInput!) {
				updateCookieBanner(input: $input) { cookieBanner { id } }
			}
		`
		var result struct{}
		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"cookieBannerId":    bannerID,
				"consentExpiryDays": 90,
			},
		}, &result)
		require.NoError(t, err)

		got := latestVersion(t, owner, bannerID)
		assert.Greater(t, got.Version, baseline, "real consent change should bump the version")
		assert.Equal(t, "DRAFT", got.State)
	})

	t.Run("UpdateTrackerPattern displayName change on visible pattern creates a new draft", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID, factory.Attrs{"slug": "real-change"})
		patternID := factory.CreateTrackerPattern(owner, categoryID, factory.Attrs{
			"displayName": "Original",
		})

		published := publishBanner(t, owner, bannerID)
		baseline := published.Version

		const query = `
			mutation UpdateTrackerPattern($input: UpdateTrackerPatternInput!) {
				updateTrackerPattern(input: $input) { trackerPattern { id } }
			}
		`
		var result struct{}
		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"trackerPatternId": patternID,
				"displayName":      "Renamed",
			},
		}, &result)
		require.NoError(t, err)

		got := latestVersion(t, owner, bannerID)
		assert.Greater(t, got.Version, baseline)
		assert.Equal(t, "DRAFT", got.State)
	})
}
