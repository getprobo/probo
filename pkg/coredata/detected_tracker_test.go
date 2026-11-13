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

package coredata_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func newDetectedTracker(
	fx trackerPatternFixture,
	patternID gid.GID,
	identifier string,
	observedBy coredata.DetectedTrackerObserver,
) *coredata.DetectedTracker {
	now := time.Now().UTC().Truncate(time.Microsecond)

	return &coredata.DetectedTracker{
		ID:               gid.New(fx.scope.GetTenantID(), coredata.DetectedTrackerEntityType),
		CookieBannerID:   fx.cookieBannerID,
		TrackerPatternID: &patternID,
		TrackerType:      coredata.TrackerTypeCookie,
		Identifier:       identifier,
		Source:           new(coredata.CookieSourceHTTP),
		ObservedBy:       observedBy,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func TestDetectedTrackers_LoadObserversByTrackerPatternIDs(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedTrackerPatternFixture(t, ctx, client)

	both := seedTrackerPattern(t, ctx, client, fx, "both", coredata.TrackerPatternMatchTypeExact, coredata.CookieSourceScript)
	sdkOnly := seedTrackerPattern(t, ctx, client, fx, "sdk_only", coredata.TrackerPatternMatchTypeExact, coredata.CookieSourceScript)
	none := seedTrackerPattern(t, ctx, client, fx, "none", coredata.TrackerPatternMatchTypeExact, coredata.CookieSourceScript)

	insert := func(patternID gid.GID, identifier string, observedBy coredata.DetectedTrackerObserver) {
		t.Helper()

		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			_, err := newDetectedTracker(fx, patternID, identifier, observedBy).InsertIfNotExists(ctx, tx, fx.scope)

			return err
		}))
	}

	insert(both.ID, "both", coredata.DetectedTrackerObserverScanner)
	insert(both.ID, "both", coredata.DetectedTrackerObserverSDK)
	insert(sdkOnly.ID, "sdk_only", coredata.DetectedTrackerObserverSDK)

	var observers map[gid.GID][]coredata.DetectedTrackerObserver

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		var err error

		observers, err = new(coredata.DetectedTrackers).LoadObserversByTrackerPatternIDs(
			ctx,
			conn,
			fx.scope,
			[]gid.GID{both.ID, sdkOnly.ID, none.ID},
		)

		return err
	}))

	assert.Equal(t, []coredata.DetectedTrackerObserver{
		coredata.DetectedTrackerObserverScanner,
		coredata.DetectedTrackerObserverSDK,
	}, observers[both.ID])
	assert.Equal(t, []coredata.DetectedTrackerObserver{
		coredata.DetectedTrackerObserverSDK,
	}, observers[sdkOnly.ID])
	_, found := observers[none.ID]
	assert.False(t, found)
}

func TestDetectedTracker_InsertIfNotExists_SeparatesObservers(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedTrackerPatternFixture(t, ctx, client)
	tp := seedTrackerPattern(t, ctx, client, fx, "_ga", coredata.TrackerPatternMatchTypeExact, coredata.CookieSourceScript)

	insert := func(dt *coredata.DetectedTracker) bool {
		var inserted bool

		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			var err error

			inserted, err = dt.InsertIfNotExists(ctx, tx, fx.scope)

			return err
		}))

		return inserted
	}

	assert.True(t, insert(newDetectedTracker(fx, tp.ID, "_ga", coredata.DetectedTrackerObserverSDK)))
	assert.True(t, insert(newDetectedTracker(fx, tp.ID, "_ga", coredata.DetectedTrackerObserverScanner)), "the scanner row must not collide with the SDK row")
	assert.False(t, insert(newDetectedTracker(fx, tp.ID, "_ga", coredata.DetectedTrackerObserverSDK)), "a repeated SDK detection must not insert a second row")
}

func TestDetectedTracker_Upsert_RefreshesScannerAttributes(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedTrackerPatternFixture(t, ctx, client)
	tp := seedTrackerPattern(t, ctx, client, fx, "_session", coredata.TrackerPatternMatchTypeExact, coredata.CookieSourceHTTP)

	scan := newPendingCookieBannerScan(fx)
	require.NoError(t, insertCookieBannerScan(t, ctx, client, fx, scan))

	first := newDetectedTracker(fx, tp.ID, "_session", coredata.DetectedTrackerObserverScanner)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		_, err := first.Upsert(ctx, tx, fx.scope)
		return err
	}))

	second := newDetectedTracker(fx, tp.ID, "_session", coredata.DetectedTrackerObserverScanner)
	second.HTTPOnly = new(true)
	second.SameSite = new(coredata.CookieSameSiteLax)
	second.SetByURL = new("https://tracker-pattern-test.example.com/login")
	second.CookieBannerScanID = &scan.ID

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		_, err := second.Upsert(ctx, tx, fx.scope)
		return err
	}))

	var trackers coredata.DetectedTrackers

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		cursor := page.NewCursor(
			10,
			nil,
			page.Head,
			page.OrderBy[coredata.DetectedTrackerOrderField]{
				Field:     coredata.DetectedTrackerOrderFieldCreatedAt,
				Direction: page.OrderDirectionAsc,
			},
		)

		return trackers.LoadByTrackerPatternID(ctx, conn, fx.scope, tp.ID, cursor)
	}))

	require.Len(t, trackers, 1)
	got := trackers[0]
	assert.Equal(t, first.ID, got.ID)
	assert.Equal(t, coredata.DetectedTrackerObserverScanner, got.ObservedBy)
	require.NotNil(t, got.HTTPOnly)
	assert.True(t, *got.HTTPOnly)
	require.NotNil(t, got.SameSite)
	assert.Equal(t, coredata.CookieSameSiteLax, *got.SameSite)
	require.NotNil(t, got.CookieBannerScanID)
	assert.Equal(t, scan.ID, *got.CookieBannerScanID)
}

func TestTrackerResource_Upsert_StampsScanOnlyWhenSet(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedTrackerPatternFixture(t, ctx, client)

	scan := newPendingCookieBannerScan(fx)
	require.NoError(t, insertCookieBannerScan(t, ctx, client, fx, scan))

	newResource := func(scanID *gid.GID) *coredata.TrackerResource {
		now := time.Now().UTC().Truncate(time.Microsecond)

		return &coredata.TrackerResource{
			ID:                 gid.New(fx.scope.GetTenantID(), coredata.TrackerResourceEntityType),
			OrganizationID:     fx.organizationID,
			CookieBannerID:     fx.cookieBannerID,
			CookieCategoryID:   fx.cookieCategoryID,
			ResourceType:       coredata.TrackerResourceTypeScript,
			Origin:             "https://www.googletagmanager.com",
			Path:               "/gtm.js",
			DisplayName:        "www.googletagmanager.com/gtm.js",
			CookieBannerScanID: scanID,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
	}

	upsert := func(tr *coredata.TrackerResource) bool {
		var inserted bool

		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			var err error

			inserted, err = tr.Upsert(ctx, tx, fx.scope)

			return err
		}))

		return inserted
	}

	sdkReport := newResource(nil)
	require.True(t, upsert(sdkReport))
	sdkUpdatedAt := sdkReport.UpdatedAt

	assert.False(t, upsert(newResource(nil)), "a repeated SDK report must stay a no-op")

	scannerReport := newResource(&scan.ID)
	scannerReport.UpdatedAt = sdkUpdatedAt.Add(time.Second)
	assert.False(t, upsert(scannerReport), "a scanner report of an existing resource is not an insert")

	var reloaded coredata.TrackerResource

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloaded.LoadByID(ctx, conn, fx.scope, sdkReport.ID)
	}))
	require.NotNil(t, reloaded.CookieBannerScanID)
	assert.Equal(t, scan.ID, *reloaded.CookieBannerScanID)
	assert.True(t, reloaded.UpdatedAt.After(sdkUpdatedAt), "stamping the scan must refresh updated_at")
}

func TestDetectedTracker_CountByCookieBannerScanID_SplitsCookiesAndStorage(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedTrackerPatternFixture(t, ctx, client)
	tp := seedTrackerPattern(t, ctx, client, fx, "_ga", coredata.TrackerPatternMatchTypeExact, coredata.CookieSourceHTTP)

	scan := newPendingCookieBannerScan(fx)
	require.NoError(t, insertCookieBannerScan(t, ctx, client, fx, scan))

	other := newPendingCookieBannerScan(fx)
	other.State = coredata.CookieBannerScanStateCompleted
	require.NoError(t, insertCookieBannerScan(t, ctx, client, fx, other))

	cookie := newDetectedTracker(fx, tp.ID, "_ga", coredata.DetectedTrackerObserverScanner)
	cookie.CookieBannerScanID = &scan.ID

	storage := newDetectedTracker(fx, tp.ID, "theme", coredata.DetectedTrackerObserverScanner)
	storage.TrackerType = coredata.TrackerTypeLocalStorage
	storage.CookieBannerScanID = &scan.ID

	session := newDetectedTracker(fx, tp.ID, "tab", coredata.DetectedTrackerObserverScanner)
	session.TrackerType = coredata.TrackerTypeSessionStorage
	session.CookieBannerScanID = &scan.ID

	sdk := newDetectedTracker(fx, tp.ID, "sdk_only", coredata.DetectedTrackerObserverSDK)
	sdk.CookieBannerScanID = &scan.ID

	elsewhere := newDetectedTracker(fx, tp.ID, "other_scan", coredata.DetectedTrackerObserverScanner)
	elsewhere.CookieBannerScanID = &other.ID

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		for _, dt := range []*coredata.DetectedTracker{cookie, storage, session, sdk, elsewhere} {
			if _, err := dt.Upsert(ctx, tx, fx.scope); err != nil {
				return err
			}
		}

		return nil
	}))

	var cookies, storageCount int

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		var err error

		cookies, storageCount, err = new(coredata.DetectedTrackers).CountByCookieBannerScanID(ctx, conn, fx.scope, scan.ID)

		return err
	}))
	assert.Equal(t, 1, cookies)
	assert.Equal(t, 2, storageCount)
}

func TestTrackerResource_CountByCookieBannerScanID_SkipsExcluded(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedTrackerPatternFixture(t, ctx, client)

	scan := newPendingCookieBannerScan(fx)
	require.NoError(t, insertCookieBannerScan(t, ctx, client, fx, scan))

	newResource := func(path string, scanID *gid.GID, excluded bool) *coredata.TrackerResource {
		now := time.Now().UTC().Truncate(time.Microsecond)

		return &coredata.TrackerResource{
			ID:                 gid.New(fx.scope.GetTenantID(), coredata.TrackerResourceEntityType),
			OrganizationID:     fx.organizationID,
			CookieBannerID:     fx.cookieBannerID,
			CookieCategoryID:   fx.cookieCategoryID,
			ResourceType:       coredata.TrackerResourceTypeScript,
			Origin:             "https://www.googletagmanager.com",
			Path:               path,
			DisplayName:        "www.googletagmanager.com" + path,
			Excluded:           excluded,
			CookieBannerScanID: scanID,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		for _, tr := range []*coredata.TrackerResource{
			newResource("/gtm.js", &scan.ID, false),
			newResource("/hidden.js", &scan.ID, true),
			newResource("/sdk.js", nil, false),
		} {
			if err := tr.Insert(ctx, tx, fx.scope); err != nil {
				return err
			}
		}

		return nil
	}))

	var count int

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		var err error

		count, err = new(coredata.TrackerResources).CountByCookieBannerScanID(ctx, conn, fx.scope, scan.ID)

		return err
	}))
	assert.Equal(t, 1, count)
}

func TestTrackerResource_LoadByCookieBannerScanID_SkipsExcluded(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedTrackerPatternFixture(t, ctx, client)

	scan := newPendingCookieBannerScan(fx)
	require.NoError(t, insertCookieBannerScan(t, ctx, client, fx, scan))

	newResource := func(path string, excluded bool) *coredata.TrackerResource {
		now := time.Now().UTC().Truncate(time.Microsecond)

		return &coredata.TrackerResource{
			ID:                 gid.New(fx.scope.GetTenantID(), coredata.TrackerResourceEntityType),
			OrganizationID:     fx.organizationID,
			CookieBannerID:     fx.cookieBannerID,
			CookieCategoryID:   fx.cookieCategoryID,
			ResourceType:       coredata.TrackerResourceTypeScript,
			Origin:             "https://www.googletagmanager.com",
			Path:               path,
			DisplayName:        "www.googletagmanager.com" + path,
			Excluded:           excluded,
			CookieBannerScanID: &scan.ID,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		for _, tr := range []*coredata.TrackerResource{
			newResource("/gtm.js", false),
			newResource("/hidden.js", true),
		} {
			if err := tr.Insert(ctx, tx, fx.scope); err != nil {
				return err
			}
		}

		return nil
	}))

	var resources coredata.TrackerResources

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		cursor := page.NewCursor(
			10,
			nil,
			page.Head,
			page.OrderBy[coredata.TrackerResourceOrderField]{
				Field:     coredata.TrackerResourceOrderFieldCreatedAt,
				Direction: page.OrderDirectionAsc,
			},
		)

		return resources.LoadByCookieBannerScanID(ctx, conn, fx.scope, scan.ID, cursor)
	}))

	require.Len(t, resources, 1)
	assert.Equal(t, "/gtm.js", resources[0].Path)
	assert.False(t, resources[0].Excluded)
}

func TestDetectedTrackerFilter_Identifiers(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedTrackerPatternFixture(t, ctx, client)

	keep := seedTrackerPattern(t, ctx, client, fx, "keep", coredata.TrackerPatternMatchTypeExact, coredata.CookieSourcePreExisting)
	drop := seedTrackerPattern(t, ctx, client, fx, "drop", coredata.TrackerPatternMatchTypeExact, coredata.CookieSourceScript)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if _, err := newDetectedTracker(fx, keep.ID, "keep", coredata.DetectedTrackerObserverSDK).InsertIfNotExists(ctx, tx, fx.scope); err != nil {
			return err
		}

		_, err := newDetectedTracker(fx, drop.ID, "drop", coredata.DetectedTrackerObserverSDK).InsertIfNotExists(ctx, tx, fx.scope)

		return err
	}))

	cursor := page.NewCursor(
		10,
		nil,
		page.Head,
		page.OrderBy[coredata.DetectedTrackerOrderField]{
			Field:     coredata.DetectedTrackerOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
	)

	load := func(t *testing.T, identifiers []string, set bool) []string {
		t.Helper()

		filter := coredata.NewDetectedTrackerFilter([]coredata.TrackerType{coredata.TrackerTypeCookie})
		if set {
			filter.WithIdentifiers(identifiers)
		}

		var detections coredata.DetectedTrackers

		require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
			return detections.LoadByCookieBannerID(
				ctx,
				conn,
				fx.scope,
				fx.cookieBannerID,
				filter,
				cursor,
			)
		}))

		names := make([]string, 0, len(detections))
		for _, detection := range detections {
			names = append(names, detection.Identifier)
		}

		return names
	}

	t.Run("one name", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []string{"keep"}, load(t, []string{"keep"}, true))
	})

	t.Run("nil selection", func(t *testing.T) {
		t.Parallel()

		assert.ElementsMatch(t, []string{"drop", "keep"}, load(t, nil, false))
	})

	t.Run("empty selection", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, load(t, []string{}, true))
	})
}

func TestDetectedTrackerFilter_UpdatedSince(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedTrackerPatternFixture(t, ctx, client)

	pattern := seedTrackerPattern(t, ctx, client, fx, "keep", coredata.TrackerPatternMatchTypeExact, coredata.CookieSourceScript)
	older := time.Now().UTC().Add(-time.Hour).Truncate(time.Microsecond)
	newer := older.Add(time.Minute)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		stale := newDetectedTracker(fx, pattern.ID, "stale", coredata.DetectedTrackerObserverSDK)
		stale.UpdatedAt = older
		stale.CreatedAt = older
		if _, err := stale.InsertIfNotExists(ctx, tx, fx.scope); err != nil {
			return err
		}

		fresh := newDetectedTracker(fx, pattern.ID, "fresh", coredata.DetectedTrackerObserverSDK)
		fresh.UpdatedAt = newer
		fresh.CreatedAt = newer
		_, err := fresh.InsertIfNotExists(ctx, tx, fx.scope)

		return err
	}))

	cursor := page.NewCursor(
		10,
		nil,
		page.Head,
		page.OrderBy[coredata.DetectedTrackerOrderField]{
			Field:     coredata.DetectedTrackerOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
	)

	load := func(t *testing.T, since *time.Time) []string {
		t.Helper()

		filter := coredata.NewDetectedTrackerFilter([]coredata.TrackerType{coredata.TrackerTypeCookie}).
			WithUpdatedSince(since)

		var detections coredata.DetectedTrackers

		require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
			return detections.LoadByCookieBannerID(
				ctx,
				conn,
				fx.scope,
				fx.cookieBannerID,
				filter,
				cursor,
			)
		}))

		names := make([]string, 0, len(detections))
		for _, detection := range detections {
			names = append(names, detection.Identifier)
		}

		return names
	}

	t.Run("from the newer row", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []string{"fresh"}, load(t, &newer))
	})

	t.Run("nil bound", func(t *testing.T) {
		t.Parallel()

		assert.ElementsMatch(t, []string{"fresh", "stale"}, load(t, nil))
	})
}

func TestDetectedTracker_Upsert_KeepsTheLatestUpdatedAt(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedTrackerPatternFixture(t, ctx, client)
	pattern := seedTrackerPattern(t, ctx, client, fx, "kept", coredata.TrackerPatternMatchTypeExact, coredata.CookieSourceScript)

	later := time.Now().UTC().Truncate(time.Microsecond)
	earlier := later.Add(-time.Hour)

	first := newDetectedTracker(fx, pattern.ID, "kept", coredata.DetectedTrackerObserverSDK)
	first.UpdatedAt = later
	first.CreatedAt = later

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		_, err := first.Upsert(ctx, tx, fx.scope)

		return err
	}))

	second := newDetectedTracker(fx, pattern.ID, "kept", coredata.DetectedTrackerObserverSDK)
	second.UpdatedAt = earlier
	second.CreatedAt = earlier

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		_, err := second.Upsert(ctx, tx, fx.scope)

		return err
	}))

	var trackers coredata.DetectedTrackers

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		cursor := page.NewCursor(
			10,
			nil,
			page.Head,
			page.OrderBy[coredata.DetectedTrackerOrderField]{
				Field:     coredata.DetectedTrackerOrderFieldCreatedAt,
				Direction: page.OrderDirectionAsc,
			},
		)

		return trackers.LoadByTrackerPatternID(ctx, conn, fx.scope, pattern.ID, cursor)
	}))

	require.Len(t, trackers, 1)
	assert.True(t, trackers[0].UpdatedAt.Equal(later))
}
