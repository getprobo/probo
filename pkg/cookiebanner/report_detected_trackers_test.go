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

package cookiebanner

import (
	"context"
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/uri"
)

// TestReportDetectedTrackers_SkipsOversizedIdentifier asserts that a
// storage key longer than MaxTrackerIdentifierLength is ignored
// (PostgreSQL btree cannot index ~4KB values on
// idx_tracker_patterns_unique_pattern_per_banner) while a normal peer
// in the same report still inserts.
func TestReportDetectedTrackers_SkipsOversizedIdentifier(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)

	normalCookie := "_ga"
	oversizedKey := strings.Repeat("a", MaxTrackerIdentifierLength+1)
	source := coredata.CookieSourceScript

	require.NoError(
		t,
		svc.ReportDetectedTrackers(
			ctx,
			fx.banner.ID,
			ReportDetectedTrackersRequest{
				Cookies: []DetectedCookie{
					{
						Name:   normalCookie,
						Source: coredata.CookieSourceScript,
					},
				},
				Storage: []DetectedStorageItem{
					{
						Key:         oversizedKey,
						StorageType: coredata.TrackerTypeLocalStorage,
						Source:      &source,
					},
				},
			},
		),
	)

	var normalPattern coredata.TrackerPattern

	require.NoError(
		t,
		client.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				return normalPattern.LoadByBannerIDTypeAndPattern(
					ctx,
					conn,
					fx.scope,
					fx.banner.ID,
					coredata.TrackerTypeCookie,
					normalCookie,
					nil,
				)
			},
		),
	)
	assert.Equal(t, normalCookie, normalPattern.Pattern)

	var oversizedPattern coredata.TrackerPattern

	err := client.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return oversizedPattern.LoadByBannerIDTypeAndPattern(
				ctx,
				conn,
				fx.scope,
				fx.banner.ID,
				coredata.TrackerTypeLocalStorage,
				oversizedKey,
				nil,
			)
		},
	)
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)

	var patterns coredata.TrackerPatterns

	require.NoError(
		t,
		client.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				loaded, err := page.LoadAll(
					ctx,
					page.OrderBy[coredata.TrackerPatternOrderField]{
						Field:     coredata.TrackerPatternOrderFieldCreatedAt,
						Direction: page.OrderDirectionAsc,
					},
					func(ctx context.Context, cursor *page.Cursor[coredata.TrackerPatternOrderField]) ([]*coredata.TrackerPattern, error) {
						var batch coredata.TrackerPatterns
						if err := batch.LoadByCookieBannerID(ctx, conn, fx.scope, fx.banner.ID, cursor, nil); err != nil {
							return nil, err
						}

						return batch, nil
					},
				)
				if err != nil {
					return err
				}

				patterns = loaded

				return nil
			},
		),
	)
	assert.Len(t, patterns, 1, "only the normal cookie pattern must be inserted")
}

// TestReportDetectedTrackers_AcceptsMaxLengthIdentifier pins the
// boundary: an identifier of exactly MaxTrackerIdentifierLength bytes
// is stored.
func TestReportDetectedTrackers_AcceptsMaxLengthIdentifier(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)

	maxLenKey := strings.Repeat("b", MaxTrackerIdentifierLength)
	source := coredata.CookieSourceScript

	require.NoError(
		t,
		svc.ReportDetectedTrackers(
			ctx,
			fx.banner.ID,
			ReportDetectedTrackersRequest{
				Storage: []DetectedStorageItem{
					{
						Key:         maxLenKey,
						StorageType: coredata.TrackerTypeSessionStorage,
						Source:      &source,
					},
				},
			},
		),
	)

	var pattern coredata.TrackerPattern

	require.NoError(
		t,
		client.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				return pattern.LoadByBannerIDTypeAndPattern(
					ctx,
					conn,
					fx.scope,
					fx.banner.ID,
					coredata.TrackerTypeSessionStorage,
					maxLenKey,
					nil,
				)
			},
		),
	)
	assert.Equal(t, maxLenKey, pattern.Pattern)
}

func TestReportDetectedTrackers_ResourceReportingDisabled(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)

	_, err := svc.UpdateCookieBanner(
		ctx,
		fx.scope,
		UpdateCookieBannerRequest{
			CookieBannerID: fx.banner.ID,
			Capabilities:   &coredata.CookieBannerCapabilitiesPatch{ResourceReporting: new(false)},
		},
	)
	require.NoError(t, err)

	require.NoError(
		t,
		svc.ReportDetectedTrackers(
			ctx,
			fx.banner.ID,
			ReportDetectedTrackersRequest{
				Cookies: []DetectedCookie{
					{
						Name:   "_ga",
						Source: coredata.CookieSourceScript,
					},
				},
				Resources: []DetectedResourceItem{
					{
						URL:          uri.URI("https://cdn.example.com/tracker.js"),
						ResourceType: coredata.TrackerResourceTypeScript,
					},
				},
			},
		),
	)

	var resource coredata.TrackerResource

	err = client.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return resource.LoadByBannerTypeOriginPath(
				ctx,
				conn,
				fx.scope,
				fx.banner.ID,
				coredata.TrackerResourceTypeScript,
				"https://cdn.example.com",
				"/tracker.js",
			)
		},
	)
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)

	var cookiePattern coredata.TrackerPattern

	require.NoError(
		t,
		client.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				return cookiePattern.LoadByBannerIDTypeAndPattern(
					ctx,
					conn,
					fx.scope,
					fx.banner.ID,
					coredata.TrackerTypeCookie,
					"_ga",
					nil,
				)
			},
		),
	)
	assert.Equal(t, "_ga", cookiePattern.Pattern)
}

func TestReportDetectedTrackers_ResourceReportingEnabled(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)

	require.NoError(
		t,
		svc.ReportDetectedTrackers(
			ctx,
			fx.banner.ID,
			ReportDetectedTrackersRequest{
				Resources: []DetectedResourceItem{
					{
						URL:          uri.URI("https://cdn.example.com/pixel.js"),
						ResourceType: coredata.TrackerResourceTypeScript,
					},
				},
			},
		),
	)

	var resource coredata.TrackerResource

	require.NoError(
		t,
		client.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				return resource.LoadByBannerTypeOriginPath(
					ctx,
					conn,
					fx.scope,
					fx.banner.ID,
					coredata.TrackerResourceTypeScript,
					"https://cdn.example.com",
					"/pixel.js",
				)
			},
		),
	)
	assert.Equal(t, "https://cdn.example.com", resource.Origin)
	assert.Equal(t, "/pixel.js", resource.Path)
}

// TestReportDetectedTrackers_SeparatesHostOnlyAndDomain asserts that
// a host-only (or unknown-domain) cookie and a Domain-scoped cookie
// with the same name are stored as two detections under one pattern.
func TestReportDetectedTrackers_SeparatesHostOnlyAndDomain(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)

	require.NoError(
		t,
		svc.ReportDetectedTrackers(
			ctx,
			fx.banner.ID,
			ReportDetectedTrackersRequest{
				Cookies: []DetectedCookie{
					{
						Name:   "_ga",
						Source: coredata.CookieSourceScript,
					},
				},
			},
		),
	)

	domain := "example.com"
	hostOnly := false

	require.NoError(
		t,
		svc.ReportDetectedTrackers(
			ctx,
			fx.banner.ID,
			ReportDetectedTrackersRequest{
				Cookies: []DetectedCookie{
					{
						Name:         "_ga",
						Source:       coredata.CookieSourcePreExisting,
						CookieDomain: &domain,
						HostOnly:     &hostOnly,
					},
				},
			},
		),
	)

	trackers := loadDetectedTrackersByPattern(t, ctx, client, fx, "_ga")
	require.Len(t, trackers, 2)

	var hostOnlyTracker, domainTracker *coredata.DetectedTracker
	for _, tracker := range trackers {
		if tracker.CookieDomain == nil {
			hostOnlyTracker = tracker
			continue
		}

		domainTracker = tracker
	}

	require.NotNil(t, hostOnlyTracker)
	require.NotNil(t, hostOnlyTracker.Source)
	assert.Equal(t, coredata.CookieSourceScript, *hostOnlyTracker.Source)

	require.NotNil(t, domainTracker)
	require.NotNil(t, domainTracker.Source)
	assert.Equal(t, coredata.CookieSourcePreExisting, *domainTracker.Source)
	require.NotNil(t, domainTracker.CookieDomain)
	assert.Equal(t, domain, *domainTracker.CookieDomain)
	require.NotNil(t, domainTracker.HostOnly)
	assert.False(t, *domainTracker.HostOnly)

	assert.NotEqual(t, hostOnlyTracker.ID, domainTracker.ID)
	require.NotNil(t, hostOnlyTracker.TrackerPatternID)
	require.NotNil(t, domainTracker.TrackerPatternID)
	assert.Equal(t, *hostOnlyTracker.TrackerPatternID, *domainTracker.TrackerPatternID)
}

// TestReportDetectedTrackers_ExtensionSkipsMappingRequest asserts that
// a new EXTENSION pattern stays uncategorised and does not arm the
// mapping worker.
func TestReportDetectedTrackers_ExtensionSkipsMappingRequest(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)

	require.NoError(
		t,
		svc.ReportDetectedTrackers(
			ctx,
			fx.banner.ID,
			ReportDetectedTrackersRequest{
				Cookies: []DetectedCookie{
					{
						Name:   "ext_session",
						Source: coredata.CookieSourceExtension,
					},
				},
			},
		),
	)

	var pattern coredata.TrackerPattern

	require.NoError(
		t,
		client.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				return pattern.LoadByBannerIDTypeAndPattern(
					ctx,
					conn,
					fx.scope,
					fx.banner.ID,
					coredata.TrackerTypeCookie,
					"ext_session",
					nil,
				)
			},
		),
	)

	require.NotNil(t, pattern.Source)
	assert.Equal(t, coredata.CookieSourceExtension, *pattern.Source)
	assert.Equal(t, fx.uncategorisedID, pattern.CookieCategoryID)
	assert.False(t, pattern.Excluded)
	assert.Nil(t, pattern.MappingRequestedAt)
	assert.Nil(t, pattern.CommonTrackerPatternID)
}

func TestInt4OrNil(t *testing.T) {
	t.Parallel()

	assert.Nil(t, int4OrNil(nil))
	assert.Nil(t, int4OrNil(new(0)))
	assert.Nil(t, int4OrNil(new(-1)))
	assert.Nil(t, int4OrNil(new(251610986978)))

	got := int4OrNil(new(math.MaxInt32))
	require.NotNil(t, got)
	assert.Equal(t, math.MaxInt32, *got)

	got = int4OrNil(new(3600))
	require.NotNil(t, got)
	assert.Equal(t, 3600, *got)
}

// TestReportDetectedTrackers_DropsOversizedMaxAge asserts that a
// cookie Max-Age larger than PostgreSQL INTEGER does not fail the
// upsert. The tracker is stored and the duration is omitted.
func TestReportDetectedTrackers_DropsOversizedMaxAge(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)
	oversized := 251610986978

	require.NoError(
		t,
		svc.ReportDetectedTrackers(
			ctx,
			fx.banner.ID,
			ReportDetectedTrackersRequest{
				Cookies: []DetectedCookie{
					{
						Name:          "_ga",
						MaxAgeSeconds: &oversized,
						Source:        coredata.CookieSourceScript,
					},
				},
			},
		),
	)

	trackers := loadDetectedTrackersByPattern(t, ctx, client, fx, "_ga")
	require.Len(t, trackers, 1)

	assert.Nil(t, trackers[0].MaxAgeSeconds)
	require.NotNil(t, trackers[0].Source)
	assert.Equal(t, coredata.CookieSourceScript, *trackers[0].Source)
}

func loadDetectedTrackersByPattern(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx workerFixture,
	pattern string,
) []*coredata.DetectedTracker {
	t.Helper()

	var trackers []*coredata.DetectedTracker

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		var tp coredata.TrackerPattern
		if err := tp.LoadByBannerIDTypeAndPattern(
			ctx,
			conn,
			fx.scope,
			fx.banner.ID,
			coredata.TrackerTypeCookie,
			pattern,
			nil,
		); err != nil {
			return err
		}

		loaded, err := page.LoadAll(
			ctx,
			page.OrderBy[coredata.DetectedTrackerOrderField]{
				Field:     coredata.DetectedTrackerOrderFieldLastDetectedAt,
				Direction: page.OrderDirectionAsc,
			},
			func(ctx context.Context, cursor *page.Cursor[coredata.DetectedTrackerOrderField]) ([]*coredata.DetectedTracker, error) {
				var batch coredata.DetectedTrackers
				if err := batch.LoadByTrackerPatternID(ctx, conn, fx.scope, tp.ID, cursor); err != nil {
					return nil, err
				}

				return batch, nil
			},
		)
		if err != nil {
			return err
		}

		trackers = loaded

		return nil
	}))

	return trackers
}
