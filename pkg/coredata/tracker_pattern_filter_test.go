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

func TestTrackerPatternFilter_CategorizedAndExcluded(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedTrackerPatternFixture(t, ctx, client)

	onBanner := seedTrackerPattern(
		t,
		ctx,
		client,
		fx,
		"on_banner",
		coredata.TrackerPatternMatchTypeExact,
		coredata.CookieSourceScript,
	)

	excluded := seedTrackerPattern(
		t,
		ctx,
		client,
		fx,
		"excluded",
		coredata.TrackerPatternMatchTypeExact,
		coredata.CookieSourceScript,
	)
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var loaded coredata.TrackerPattern
		if err := loaded.LoadByID(ctx, tx, fx.scope, excluded.ID); err != nil {
			return err
		}

		loaded.Excluded = true
		loaded.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)

		return loaded.Update(ctx, tx, fx.scope)
	}))

	uncategorisedID := gid.New(fx.scope.GetTenantID(), coredata.CookieCategoryEntityType)
	now := time.Now().UTC()
	uncategorisedSource := coredata.CookieSourceScript

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		category := &coredata.CookieCategory{
			ID:              uncategorisedID,
			OrganizationID:  fx.organizationID,
			CookieBannerID:  fx.cookieBannerID,
			Name:            "Uncategorised",
			Slug:            "uncategorised",
			Description:     "",
			Kind:            coredata.CookieCategoryKindUncategorised,
			Rank:            99,
			GCMConsentTypes: []string{},
			TCFPurposeIDs:   []int{},
			PostHogConsent:  false,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := category.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		uncategorised := &coredata.TrackerPattern{
			ID:               gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType),
			OrganizationID:   fx.organizationID,
			CookieBannerID:   fx.cookieBannerID,
			CookieCategoryID: uncategorisedID,
			TrackerType:      coredata.TrackerTypeCookie,
			Pattern:          "uncategorised",
			MatchType:        coredata.TrackerPatternMatchTypeExact,
			Source:           &uncategorisedSource,
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		return uncategorised.Insert(ctx, tx, fx.scope)
	}))

	cursor := page.NewCursor(
		10,
		nil,
		page.Head,
		page.OrderBy[coredata.TrackerPatternOrderField]{
			Field:     coredata.TrackerPatternOrderFieldName,
			Direction: page.OrderDirectionAsc,
		},
	)

	load := func(t *testing.T, apply func(*coredata.TrackerPatternFilter)) []string {
		t.Helper()

		filter := coredata.NewTrackerPatternFilter(nil, nil, nil)
		apply(filter)

		var patterns coredata.TrackerPatterns

		require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
			return patterns.LoadByCookieBannerID(
				ctx,
				conn,
				fx.scope,
				fx.cookieBannerID,
				cursor,
				filter,
			)
		}))

		names := make([]string, 0, len(patterns))
		for _, pattern := range patterns {
			names = append(names, pattern.Pattern)
		}

		return names
	}

	t.Run("on banner is categorized and not excluded", func(t *testing.T) {
		t.Parallel()

		names := load(t, func(filter *coredata.TrackerPatternFilter) {
			filter.WithExcluded(new(false)).WithCategorized(new(true))
		})

		assert.Equal(t, []string{onBanner.Pattern}, names)
	})

	t.Run("uncategorised only", func(t *testing.T) {
		t.Parallel()

		names := load(t, func(filter *coredata.TrackerPatternFilter) {
			filter.WithCategorized(new(false))
		})

		assert.Equal(t, []string{"uncategorised"}, names)
	})
}

func TestTrackerPatternFilter_ObservedBy(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedTrackerPatternFixture(t, ctx, client)

	sdkPattern := seedTrackerPattern(
		t,
		ctx,
		client,
		fx,
		"sdk_only",
		coredata.TrackerPatternMatchTypeExact,
		coredata.CookieSourceScript,
	)
	scanPattern := seedTrackerPattern(
		t,
		ctx,
		client,
		fx,
		"scan_only",
		coredata.TrackerPatternMatchTypeExact,
		coredata.CookieSourceHTTP,
	)
	bothPattern := seedTrackerPattern(
		t,
		ctx,
		client,
		fx,
		"both",
		coredata.TrackerPatternMatchTypeExact,
		coredata.CookieSourceScript,
	)

	insert := func(patternID gid.GID, identifier string, observedBy coredata.DetectedTrackerObserver) {
		t.Helper()

		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			_, err := newDetectedTracker(fx, patternID, identifier, observedBy).InsertIfNotExists(ctx, tx, fx.scope)

			return err
		}))
	}

	seedTrackerPattern(
		t,
		ctx,
		client,
		fx,
		"unseen",
		coredata.TrackerPatternMatchTypeExact,
		coredata.CookieSourceScript,
	)

	insert(sdkPattern.ID, "sdk_only", coredata.DetectedTrackerObserverSDK)
	insert(scanPattern.ID, "scan_only", coredata.DetectedTrackerObserverScanner)
	insert(bothPattern.ID, "both", coredata.DetectedTrackerObserverSDK)
	insert(bothPattern.ID, "both", coredata.DetectedTrackerObserverScanner)

	cursor := page.NewCursor(
		10,
		nil,
		page.Head,
		page.OrderBy[coredata.TrackerPatternOrderField]{
			Field:     coredata.TrackerPatternOrderFieldName,
			Direction: page.OrderDirectionAsc,
		},
	)

	load := func(t *testing.T, observedBy []coredata.DetectedTrackerObserver) []string {
		t.Helper()

		filter := coredata.NewTrackerPatternFilter(nil, nil, nil).
			WithObservedBy(observedBy)

		var patterns coredata.TrackerPatterns

		require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
			return patterns.LoadByCookieBannerID(
				ctx,
				conn,
				fx.scope,
				fx.cookieBannerID,
				cursor,
				filter,
			)
		}))

		names := make([]string, 0, len(patterns))
		for _, pattern := range patterns {
			names = append(names, pattern.Pattern)
		}

		return names
	}

	t.Run("sdk detections", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []string{"both", "sdk_only"}, load(t, []coredata.DetectedTrackerObserver{
			coredata.DetectedTrackerObserverSDK,
		}))
	})

	t.Run("scanner detections", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []string{"both", "scan_only"}, load(t, []coredata.DetectedTrackerObserver{
			coredata.DetectedTrackerObserverScanner,
		}))
	})

	t.Run("either observer", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []string{"both", "scan_only", "sdk_only"}, load(t, []coredata.DetectedTrackerObserver{
			coredata.DetectedTrackerObserverSDK,
			coredata.DetectedTrackerObserverScanner,
		}))
	})

	t.Run("empty selection", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []string{"both", "scan_only", "sdk_only", "unseen"}, load(t, nil))
	})
}

func TestTrackerPatternFilter_Identifiers(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedTrackerPatternFixture(t, ctx, client)

	seedTrackerPattern(t, ctx, client, fx, "keep", coredata.TrackerPatternMatchTypeExact, coredata.CookieSourcePreExisting)
	seedTrackerPattern(t, ctx, client, fx, "drop", coredata.TrackerPatternMatchTypeExact, coredata.CookieSourceScript)

	cursor := page.NewCursor(
		10,
		nil,
		page.Head,
		page.OrderBy[coredata.TrackerPatternOrderField]{
			Field:     coredata.TrackerPatternOrderFieldName,
			Direction: page.OrderDirectionAsc,
		},
	)

	load := func(t *testing.T, identifiers []string, set bool) []string {
		t.Helper()

		filter := coredata.NewTrackerPatternFilter(nil, nil, nil)
		if set {
			filter.WithIdentifiers(identifiers)
		}

		var patterns coredata.TrackerPatterns

		require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
			return patterns.LoadByCookieBannerID(
				ctx,
				conn,
				fx.scope,
				fx.cookieBannerID,
				cursor,
				filter,
			)
		}))

		names := make([]string, 0, len(patterns))
		for _, pattern := range patterns {
			names = append(names, pattern.Pattern)
		}

		return names
	}

	t.Run("one name", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []string{"keep"}, load(t, []string{"keep"}, true))
	})

	t.Run("nil selection", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []string{"drop", "keep"}, load(t, nil, false))
	})

	t.Run("empty selection", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, load(t, []string{}, true))
	})
}
