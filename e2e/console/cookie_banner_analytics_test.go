// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package console_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

type browserFamilyCount struct {
	Family string `json:"family"`
	Count  int    `json:"count"`
}

func TestCookieBanner_DailyPageLoads(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	const query = `
		query($id: ID!, $days: Int) {
			node(id: $id) {
				... on CookieBanner {
					dailyPageLoads(days: $days) {
						day
						pageLoads {
							family
							count
						}
					}
				}
			}
		}
	`

	type dailyPageLoadsResult struct {
		Node struct {
			DailyPageLoads []struct {
				Day       string               `json:"day"`
				PageLoads []browserFamilyCount `json:"pageLoads"`
			} `json:"dailyPageLoads"`
		} `json:"node"`
	}

	t.Run("returns one entry per day oldest first", func(t *testing.T) {
		t.Parallel()

		bannerID := factory.CreateCookieBanner(owner)
		today := utcToday()

		seedDailyPageLoad(t, bannerID, today, "CHROME", "OPT_IN", 2)
		seedDailyPageLoad(t, bannerID, today, "CHROME", "OPT_OUT", 1)
		seedDailyPageLoad(t, bannerID, today.AddDate(0, 0, -6), "FIREFOX", "OPT_IN", 4)
		seedDailyPageLoad(t, bannerID, today.AddDate(0, 0, -7), "SAFARI", "OPT_IN", 5)

		var result dailyPageLoadsResult
		require.NoError(t, owner.Execute(query, map[string]any{"id": bannerID}, &result))
		require.Len(t, result.Node.DailyPageLoads, 7)

		for i, entry := range result.Node.DailyPageLoads {
			day, err := time.Parse(time.RFC3339Nano, entry.Day)
			require.NoError(t, err)
			assert.True(t, day.Equal(today.AddDate(0, 0, i-6)), "entry %d", i)
		}

		assert.Equal(
			t,
			[]browserFamilyCount{{Family: "FIREFOX", Count: 4}},
			result.Node.DailyPageLoads[0].PageLoads,
		)
		for _, entry := range result.Node.DailyPageLoads[1:6] {
			assert.Empty(t, entry.PageLoads)
		}
		assert.Equal(
			t,
			[]browserFamilyCount{{Family: "CHROME", Count: 3}},
			result.Node.DailyPageLoads[6].PageLoads,
		)

		var wider dailyPageLoadsResult
		require.NoError(t, owner.Execute(query, map[string]any{"id": bannerID, "days": 8}, &wider))
		require.Len(t, wider.Node.DailyPageLoads, 8)
		assert.Equal(
			t,
			[]browserFamilyCount{{Family: "SAFARI", Count: 5}},
			wider.Node.DailyPageLoads[0].PageLoads,
		)
	})

	t.Run("rejects days out of range", func(t *testing.T) {
		t.Parallel()

		bannerID := factory.CreateCookieBanner(owner)

		for _, days := range []int{0, 91} {
			err := owner.Execute(query, map[string]any{"id": bannerID, "days": days}, &dailyPageLoadsResult{})
			testutil.RequireErrorCode(t, err, "INVALID", "days=%d must be rejected", days)
		}
	})
}

func TestTrackerPattern_Hits(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	const query = `
		query($id: ID!, $days: Int) {
			node(id: $id) {
				... on TrackerPattern {
					hits(days: $days) {
						family
						count
					}
					allowedHits(days: $days) {
						family
						count
					}
					deniedHits(days: $days) {
						family
						count
					}
				}
			}
		}
	`

	type hitsResult struct {
		Node struct {
			Hits        []browserFamilyCount `json:"hits"`
			AllowedHits []browserFamilyCount `json:"allowedHits"`
			DeniedHits  []browserFamilyCount `json:"deniedHits"`
		} `json:"node"`
	}

	t.Run("empty for a new pattern", func(t *testing.T) {
		t.Parallel()

		patternID := createAnalyticsTrackerPattern(owner)

		var result hitsResult
		require.NoError(t, owner.Execute(query, map[string]any{"id": patternID}, &result))
		assert.Empty(t, result.Node.Hits)
		assert.Empty(t, result.Node.AllowedHits)
		assert.Empty(t, result.Node.DeniedHits)
	})

	t.Run("splits the last seven days by consent", func(t *testing.T) {
		t.Parallel()

		patternID := createAnalyticsTrackerPattern(owner)
		today := utcToday()

		seedDailyPatternHit(t, patternID, today, "CHROME", true, 2)
		seedDailyPatternHit(t, patternID, today, "CHROME", false, 1)
		seedDailyPatternHit(t, patternID, today.AddDate(0, 0, -7), "FIREFOX", true, 5)

		var lastWeek hitsResult
		require.NoError(t, owner.Execute(query, map[string]any{"id": patternID}, &lastWeek))
		assert.Equal(t, []browserFamilyCount{{Family: "CHROME", Count: 3}}, lastWeek.Node.Hits)
		assert.Equal(t, []browserFamilyCount{{Family: "CHROME", Count: 2}}, lastWeek.Node.AllowedHits)
		assert.Equal(t, []browserFamilyCount{{Family: "CHROME", Count: 1}}, lastWeek.Node.DeniedHits)

		var lastEightDays hitsResult
		require.NoError(t, owner.Execute(query, map[string]any{"id": patternID, "days": 8}, &lastEightDays))
		assert.Equal(
			t,
			[]browserFamilyCount{
				{Family: "CHROME", Count: 3},
				{Family: "FIREFOX", Count: 5},
			},
			lastEightDays.Node.Hits,
		)
		assert.Equal(
			t,
			[]browserFamilyCount{
				{Family: "CHROME", Count: 2},
				{Family: "FIREFOX", Count: 5},
			},
			lastEightDays.Node.AllowedHits,
		)
		assert.Equal(t, []browserFamilyCount{{Family: "CHROME", Count: 1}}, lastEightDays.Node.DeniedHits)
	})

	t.Run("rejects days out of range", func(t *testing.T) {
		t.Parallel()

		patternID := createAnalyticsTrackerPattern(owner)

		for _, days := range []int{0, 91} {
			err := owner.Execute(query, map[string]any{"id": patternID, "days": days}, &hitsResult{})
			testutil.RequireErrorCode(t, err, "INVALID", "days=%d must be rejected", days)
		}
	})
}

func createAnalyticsTrackerPattern(owner *testutil.Client) string {
	bannerID := factory.CreateCookieBanner(owner)
	categoryID := factory.CreateCookieCategory(owner, bannerID)

	return factory.CreateTrackerPattern(owner, categoryID)
}

func utcToday() time.Time {
	year, month, day := time.Now().UTC().Date()

	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func seedDailyPageLoad(
	t *testing.T,
	cookieBannerID string,
	day time.Time,
	family string,
	presentation string,
	count int,
) {
	t.Helper()

	ctx := context.Background()
	conn := dialTestPg(t, ctx)
	t.Cleanup(func() { _ = conn.Close(ctx) })

	tag, err := conn.Exec(ctx, `
		INSERT INTO cookie_banner_daily_page_loads (
			cookie_banner_id, day, browser_family, presentation, consent_action, tenant_id, count, created_at, updated_at
		)
		SELECT id, $2::date, $3::text, $4::text, '', tenant_id, $5::integer, now(), now()
		FROM cookie_banners
		WHERE id = $1
	`, cookieBannerID, day, family, presentation, count)
	require.NoError(t, err)
	require.EqualValues(t, 1, tag.RowsAffected(), "cookie banner %s must exist", cookieBannerID)
}

func seedDailyPatternHit(
	t *testing.T,
	trackerPatternID string,
	day time.Time,
	family string,
	categoryGranted bool,
	count int,
) {
	t.Helper()

	ctx := context.Background()
	conn := dialTestPg(t, ctx)
	t.Cleanup(func() { _ = conn.Close(ctx) })

	tag, err := conn.Exec(ctx, `
		INSERT INTO tracker_pattern_daily_hits (
			tracker_pattern_id, day, browser_family, presentation, consent_action, category_granted, tenant_id, cookie_banner_id, count, created_at, updated_at
		)
		SELECT id, $2::date, $3::text, 'OPT_IN', '', $4::boolean, tenant_id, cookie_banner_id, $5::integer, now(), now()
		FROM tracker_patterns
		WHERE id = $1
	`, trackerPatternID, day, family, categoryGranted, count)
	require.NoError(t, err)
	require.EqualValues(t, 1, tag.RowsAffected(), "tracker pattern %s must exist", trackerPatternID)
}
