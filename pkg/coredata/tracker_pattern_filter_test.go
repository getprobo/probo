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
			DisplayName:      "uncategorised",
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

	load := func(apply func(*coredata.TrackerPatternFilter)) []string {
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

		names := load(func(filter *coredata.TrackerPatternFilter) {
			filter.WithExcluded(new(false)).WithCategorized(new(true))
		})

		assert.Equal(t, []string{onBanner.Pattern}, names)
	})

	t.Run("uncategorised only", func(t *testing.T) {
		t.Parallel()

		names := load(func(filter *coredata.TrackerPatternFilter) {
			filter.WithCategorized(new(false))
		})

		assert.Equal(t, []string{"uncategorised"}, names)
	})
}
