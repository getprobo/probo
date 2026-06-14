// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// TestResetBannerTrackers_FullRebuild seeds a banner with an
// uncategorised glob covering two detections, an uncategorised exact
// carrying catalog/vendor links, a categorised exact, and an excluded
// exact. A full reset must: decompose the glob into per-identifier
// exacts and relink its detections, clear links on the surviving
// uncategorised exact and re-arm its mapping, preserve the categorised
// and excluded patterns, and arm pattern analysis on the banner.
func TestResetBannerTrackers_FullRebuild(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)

	thirdPartyID := seedThirdParty(t, ctx, client, fx, "Reset Vendor")
	commonPatternID := seedCommonTrackerPattern(t, ctx, client, "ga_linked")

	glob := newGlobInCategory(fx, "_ga_*", fx.uncategorisedID, coredata.CookieSourceScript, nil)

	linkedExact := newExactPattern(fx, "linked_cookie", fx.uncategorisedID, coredata.CookieSourcePreExisting, nil)
	linkedExact.CommonTrackerPatternID = &commonPatternID
	linkedExact.ThirdPartyID = &thirdPartyID
	linkedExact.Description = "stale description"

	categorised := newExactPattern(fx, "categorised_cookie", fx.normalCategoryID, coredata.CookieSourceScript, nil)

	excluded := newExactPattern(fx, "excluded_cookie", fx.uncategorisedID, coredata.CookieSourceScript, nil)
	excluded.Excluded = true

	now := time.Now().UTC().Truncate(time.Microsecond)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		for _, p := range []*coredata.TrackerPattern{glob, linkedExact, categorised, excluded} {
			if err := p.Insert(ctx, tx, fx.predicate); err != nil {
				return err
			}
		}

		for _, identifier := range []string{"_ga_ABC", "_ga_DEF"} {
			detection := &coredata.DetectedTracker{
				ID:               gid.New(fx.predicate.GetTenantID(), coredata.DetectedTrackerEntityType),
				CookieBannerID:   fx.banner.ID,
				TrackerPatternID: &glob.ID,
				TrackerType:      coredata.TrackerTypeCookie,
				Identifier:       identifier,
				Source:           new(coredata.CookieSourceScript),
				LastDetectedAt:   now,
				CreatedAt:        now,
				UpdatedAt:        now,
			}

			if _, err := detection.Upsert(ctx, tx, fx.predicate); err != nil {
				return err
			}
		}

		return nil
	}))

	result, err := ResetBannerTrackers(ctx, client, fx.predicate, fx.banner.ID, false, nil)
	require.NoError(t, err)

	require.Equal(t, 1, result.GlobsDecomposed)
	require.Equal(t, 2, result.ExactsCreated)
	require.Equal(t, 2, result.DetectionsRelinked)
	require.True(t, result.AnalysisRequested)
	// linked_cookie + _ga_ABC + _ga_DEF (excluded and categorised are untouched).
	require.Equal(t, int64(3), result.PatternsReset)

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		// The glob is gone.
		var goneGlob coredata.TrackerPattern

		err := goneGlob.LoadByBannerIDTypeAndPattern(ctx, conn, fx.predicate, fx.banner.ID, coredata.TrackerTypeCookie, "_ga_*", nil)
		require.ErrorIs(t, err, coredata.ErrResourceNotFound)

		// Each detection identifier is now its own exact, with mapping armed
		// and no links, and the detection relinked to it.
		for _, identifier := range []string{"_ga_ABC", "_ga_DEF"} {
			var exact coredata.TrackerPattern
			require.NoError(t, exact.LoadByBannerIDTypeAndPattern(ctx, conn, fx.predicate, fx.banner.ID, coredata.TrackerTypeCookie, identifier, nil))
			require.Equal(t, coredata.TrackerPatternMatchTypeExact, exact.MatchType)
			require.Equal(t, fx.uncategorisedID, exact.CookieCategoryID)
			require.Nil(t, exact.CommonTrackerPatternID)
			require.Nil(t, exact.ThirdPartyID)
			require.NotNil(t, exact.MappingRequestedAt)

			var detections coredata.DetectedTrackers
			require.NoError(t, detections.LoadAllByTrackerPatternID(ctx, conn, fx.predicate, exact.ID))
			require.Len(t, detections, 1)
			require.Equal(t, identifier, detections[0].Identifier)
		}

		// The surviving uncategorised exact had its links and copied
		// description cleared and mapping re-armed.
		var survivor coredata.TrackerPattern
		require.NoError(t, survivor.LoadByBannerIDTypeAndPattern(ctx, conn, fx.predicate, fx.banner.ID, coredata.TrackerTypeCookie, "linked_cookie", nil))
		require.Nil(t, survivor.CommonTrackerPatternID)
		require.Nil(t, survivor.ThirdPartyID)
		require.Empty(t, survivor.Description)
		require.NotNil(t, survivor.MappingRequestedAt)

		// The categorised pattern is untouched.
		var categorisedRow coredata.TrackerPattern
		require.NoError(t, categorisedRow.LoadByBannerIDTypeAndPattern(ctx, conn, fx.predicate, fx.banner.ID, coredata.TrackerTypeCookie, "categorised_cookie", nil))
		require.Equal(t, fx.normalCategoryID, categorisedRow.CookieCategoryID)

		// The excluded pattern is preserved.
		var excludedRow coredata.TrackerPattern
		require.NoError(t, excludedRow.LoadByBannerIDTypeAndPattern(ctx, conn, fx.predicate, fx.banner.ID, coredata.TrackerTypeCookie, "excluded_cookie", nil))
		require.True(t, excludedRow.Excluded)

		// Pattern analysis is armed on the banner.
		var banner coredata.CookieBanner
		require.NoError(t, banner.LoadByID(ctx, conn, fx.predicate, fx.banner.ID))
		require.NotNil(t, banner.PatternAnalysisRequestedAt)

		return nil
	}))
}

func seedCommonTrackerPattern(t *testing.T, ctx context.Context, client *pg.Client, pattern string) gid.GID {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Microsecond)
	cp := coredata.CommonTrackerPattern{
		ID:          gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
		TrackerType: coredata.TrackerTypeCookie,
		Pattern:     pattern,
		MatchType:   coredata.TrackerPatternMatchTypeExact,
		Description: "seeded",
		Confidence:  1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return cp.Insert(ctx, tx)
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			_, err := tx.Exec(ctx, `DELETE FROM common_tracker_patterns WHERE id = $1`, cp.ID)
			return err
		})
	})

	return cp.ID
}
