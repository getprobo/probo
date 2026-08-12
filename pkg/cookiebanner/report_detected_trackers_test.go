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
	svc := NewService(client, false)

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
	svc := NewService(client, false)

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
	svc := NewService(client, false)

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
	svc := NewService(client, false)

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
