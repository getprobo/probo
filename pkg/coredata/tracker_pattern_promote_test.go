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

package coredata_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// trackerPatternFixture bootstraps the parent rows that a tracker
// pattern's FKs require: organization, cookie banner, and a normal
// cookie category.
type trackerPatternFixture struct {
	scope            *coredata.Scope
	organizationID   gid.GID
	cookieBannerID   gid.GID
	cookieCategoryID gid.GID
}

func seedTrackerPatternFixture(t *testing.T, ctx context.Context, client *pg.Client) trackerPatternFixture {
	t.Helper()

	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	cookieBannerID := gid.New(tenantID, coredata.CookieBannerEntityType)
	cookieCategoryID := gid.New(tenantID, coredata.CookieCategoryEntityType)
	now := time.Now().UTC()

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		org := &coredata.Organization{
			ID:        organizationID,
			TenantID:  tenantID,
			Name:      "TrackerPattern Promote Test Org",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := org.Insert(ctx, tx); err != nil {
			return err
		}

		banner := &coredata.CookieBanner{
			ID:                cookieBannerID,
			OrganizationID:    organizationID,
			Name:              "TrackerPattern Promote Test Banner",
			Origin:            "https://promote-test.example.com",
			State:             coredata.CookieBannerStateActive,
			CookiePolicyURL:   "https://promote-test.example.com/cookies",
			ConsentExpiryDays: 180,
			ShowBranding:      false,
			DefaultLanguage:   "en",
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		if err := banner.Insert(ctx, tx, scope); err != nil {
			return err
		}

		category := &coredata.CookieCategory{
			ID:              cookieCategoryID,
			OrganizationID:  organizationID,
			CookieBannerID:  cookieBannerID,
			Name:            "Analytics",
			Slug:            "analytics",
			Description:     "",
			Kind:            coredata.CookieCategoryKindNormal,
			Rank:            1,
			GCMConsentTypes: []string{},
			PostHogConsent:  false,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := category.Insert(ctx, tx, scope); err != nil {
			return err
		}

		return nil
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			if _, err := tx.Exec(ctx, `DELETE FROM tracker_patterns WHERE cookie_banner_id = $1`, cookieBannerID); err != nil {
				return err
			}

			if _, err := tx.Exec(ctx, `DELETE FROM cookie_categories WHERE cookie_banner_id = $1`, cookieBannerID); err != nil {
				return err
			}

			if _, err := tx.Exec(ctx, `DELETE FROM cookie_banners WHERE id = $1`, cookieBannerID); err != nil {
				return err
			}

			if _, err := tx.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, organizationID); err != nil {
				return err
			}

			return nil
		})
	})

	return trackerPatternFixture{
		scope:            scope,
		organizationID:   organizationID,
		cookieBannerID:   cookieBannerID,
		cookieCategoryID: cookieCategoryID,
	}
}

func seedTrackerPattern(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx trackerPatternFixture,
	pattern string,
	matchType coredata.TrackerPatternMatchType,
	source coredata.CookieSource,
) *coredata.TrackerPattern {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Microsecond)
	maxAge := 3600
	tp := &coredata.TrackerPattern{
		ID:               gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType),
		OrganizationID:   fx.organizationID,
		CookieBannerID:   fx.cookieBannerID,
		CookieCategoryID: fx.cookieCategoryID,
		TrackerType:      coredata.TrackerTypeCookie,
		Pattern:          pattern,
		MatchType:        matchType,
		DisplayName:      pattern,
		Description:      "",
		MaxAgeSeconds:    &maxAge,
		Source:           &source,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return tp.Insert(ctx, tx, fx.scope)
	}))

	return tp
}

func TestTrackerPattern_PromoteSource_OverwritesSource(t *testing.T) {
	t.Parallel()

	client := newTestPgClient(t)
	ctx := context.Background()
	fx := seedTrackerPatternFixture(t, ctx, client)

	tp := seedTrackerPattern(
		t,
		ctx,
		client,
		fx,
		"*_session",
		coredata.TrackerPatternMatchTypeGlob,
		coredata.CookieSourcePreExisting,
	)

	bumpedAt := time.Now().UTC().Add(time.Hour).Truncate(time.Microsecond)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return tp.PromoteSource(ctx, tx, fx.scope, coredata.CookieSourceScript, bumpedAt)
	}))

	require.NotNil(t, tp.Source)
	assert.Equal(t, coredata.CookieSourceScript, *tp.Source, "receiver must reflect the new source")
	assert.True(t, tp.UpdatedAt.Equal(bumpedAt), "receiver must reflect the new updated_at")

	loaded := &coredata.TrackerPattern{}

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return loaded.LoadByID(ctx, conn, fx.scope, tp.ID)
	}))

	require.NotNil(t, loaded.Source)
	assert.Equal(t, coredata.CookieSourceScript, *loaded.Source, "DB row must reflect the promoted source")
	assert.True(t, loaded.UpdatedAt.Equal(bumpedAt), "DB row must reflect the new updated_at")
}

func TestTrackerPattern_PromoteSource_OnlyTouchesSourceAndUpdatedAt(t *testing.T) {
	t.Parallel()

	client := newTestPgClient(t)
	ctx := context.Background()
	fx := seedTrackerPatternFixture(t, ctx, client)

	tp := seedTrackerPattern(
		t,
		ctx,
		client,
		fx,
		"*_token",
		coredata.TrackerPatternMatchTypeGlob,
		coredata.CookieSourcePreExisting,
	)

	originalCategory := tp.CookieCategoryID
	originalDisplay := tp.DisplayName
	originalMaxAge := tp.MaxAgeSeconds
	originalExcluded := tp.Excluded
	originalDescription := tp.Description

	bumpedAt := time.Now().UTC().Add(2 * time.Hour).Truncate(time.Microsecond)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return tp.PromoteSource(ctx, tx, fx.scope, coredata.CookieSourceExtension, bumpedAt)
	}))

	loaded := &coredata.TrackerPattern{}

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return loaded.LoadByID(ctx, conn, fx.scope, tp.ID)
	}))

	assert.Equal(t, originalCategory, loaded.CookieCategoryID, "category must be untouched")
	assert.Equal(t, originalDisplay, loaded.DisplayName, "display_name must be untouched")
	assert.Equal(t, originalMaxAge, loaded.MaxAgeSeconds, "max_age_seconds must be untouched")
	assert.Equal(t, originalExcluded, loaded.Excluded, "excluded must be untouched")
	assert.Equal(t, originalDescription, loaded.Description, "description must be untouched")
}

func TestTrackerPattern_PromoteSource_NotFoundForMissingRow(t *testing.T) {
	t.Parallel()

	client := newTestPgClient(t)
	ctx := context.Background()
	fx := seedTrackerPatternFixture(t, ctx, client)

	tp := &coredata.TrackerPattern{ID: gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType)}

	err := client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return tp.PromoteSource(ctx, tx, fx.scope, coredata.CookieSourceScript, time.Now().UTC())
	})

	assert.ErrorIs(t, err, coredata.ErrResourceNotFound)
}
