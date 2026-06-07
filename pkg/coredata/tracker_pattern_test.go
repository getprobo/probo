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
	"go.probo.inc/probo/internal/test"
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
			Name:      "TrackerPattern Test Org",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := org.Insert(ctx, tx); err != nil {
			return err
		}

		banner := &coredata.CookieBanner{
			ID:                cookieBannerID,
			OrganizationID:    organizationID,
			Name:              "TrackerPattern Test Banner",
			Origin:            "https://tracker-pattern-test.example.com",
			State:             coredata.CookieBannerStateActive,
			CookiePolicyURL:   "https://tracker-pattern-test.example.com/cookies",
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

// TestTrackerPattern_Update_WritesSource pins the source-promotion
// path now folded into Update: load the row, bump Source, call
// Update, and verify the new value lands in the DB. This is the
// invariant the pattern-analysis worker and reportDetectedTracker
// rely on when promoting PRE_EXISTING → SCRIPT/EXTENSION; if Update
// stops writing `source`, every "ratchet the signal" call site
// silently regresses without the wrapping `shouldPromoteSource`
// gate noticing.
func TestTrackerPattern_Update_WritesSource(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
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
	newSource := coredata.CookieSourceScript

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var loaded coredata.TrackerPattern
		if err := loaded.LoadByID(ctx, tx, fx.scope, tp.ID); err != nil {
			return err
		}

		loaded.Source = &newSource
		loaded.UpdatedAt = bumpedAt

		return loaded.Update(ctx, tx, fx.scope)
	}))

	reloaded := &coredata.TrackerPattern{}

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloaded.LoadByID(ctx, conn, fx.scope, tp.ID)
	}))

	require.NotNil(t, reloaded.Source)
	assert.Equal(t, coredata.CookieSourceScript, *reloaded.Source, "DB row must reflect the new source")
	assert.True(t, reloaded.UpdatedAt.Equal(bumpedAt), "DB row must reflect the new updated_at")
}

// TestTrackerPattern_Update_NotFoundForMissingRow pins the
// ErrResourceNotFound contract: callers like the worker and
// reportDetectedTracker assume an unmatched UPDATE surfaces as
// ErrResourceNotFound so they can distinguish "row vanished mid-txn"
// from arbitrary pg errors.
func TestTrackerPattern_Update_NotFoundForMissingRow(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedTrackerPatternFixture(t, ctx, client)

	maxAge := 3600
	source := coredata.CookieSourceScript
	now := time.Now().UTC().Truncate(time.Microsecond)
	tp := &coredata.TrackerPattern{
		ID:               gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType),
		OrganizationID:   fx.organizationID,
		CookieBannerID:   fx.cookieBannerID,
		CookieCategoryID: fx.cookieCategoryID,
		TrackerType:      coredata.TrackerTypeCookie,
		Pattern:          "*_ghost",
		MatchType:        coredata.TrackerPatternMatchTypeGlob,
		DisplayName:      "*_ghost",
		MaxAgeSeconds:    &maxAge,
		Source:           &source,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	err := client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return tp.Update(ctx, tx, fx.scope)
	})

	assert.ErrorIs(t, err, coredata.ErrResourceNotFound)
}

// TestResetStaleMappings pins the mapping stale-recovery contract: a row
// dequeued (mapping_requested_at IS NULL) but never finished (no
// common_tracker_pattern_id) and idle past the window is re-armed, while
// a recently claimed row (clock not yet elapsed) and a completed row
// (catalog row assigned) are left untouched. Without this sweep a crash
// between Process phases would strand the pattern unmapped forever.
func TestResetStaleMappings(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedTrackerPatternFixture(t, ctx, client)

	now := time.Now().UTC().Truncate(time.Microsecond)
	old := now.Add(-time.Hour)
	maxAge := 3600
	source := coredata.CookieSourceScript

	newPattern := func(pattern string, updatedAt time.Time, commonID *gid.GID) *coredata.TrackerPattern {
		tp := &coredata.TrackerPattern{
			ID:                     gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType),
			OrganizationID:         fx.organizationID,
			CookieBannerID:         fx.cookieBannerID,
			CookieCategoryID:       fx.cookieCategoryID,
			CommonTrackerPatternID: commonID,
			TrackerType:            coredata.TrackerTypeCookie,
			Pattern:                pattern,
			MatchType:              coredata.TrackerPatternMatchTypeExact,
			DisplayName:            pattern,
			MaxAgeSeconds:          &maxAge,
			Source:                 &source,
			CreatedAt:              old,
			UpdatedAt:              updatedAt,
		}

		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			return tp.Insert(ctx, tx, fx.scope)
		}))

		return tp
	}

	commonPattern := coredata.CommonTrackerPattern{
		ID:          gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
		TrackerType: coredata.TrackerTypeCookie,
		Pattern:     "mapped_catalog_" + gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType).String(),
		MatchType:   coredata.TrackerPatternMatchTypeExact,
		Confidence:  0.5,
		CreatedAt:   old,
		UpdatedAt:   old,
	}
	insertCommonTrackerPattern(t, ctx, client, commonPattern)

	stale := newPattern("stale_unfinished", old, nil)
	fresh := newPattern("fresh_unfinished", now, nil)
	completed := newPattern("completed_mapping", old, &commonPattern.ID)

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return coredata.ResetStaleMappings(ctx, conn, 10*time.Minute)
	}))

	load := func(id gid.GID) coredata.TrackerPattern {
		var reloaded coredata.TrackerPattern

		require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
			return reloaded.LoadByID(ctx, conn, fx.scope, id)
		}))

		return reloaded
	}

	assert.NotNil(t, load(stale.ID).MappingRequestedAt, "claimed-but-unfinished idle row must be re-armed")
	assert.Nil(t, load(fresh.ID).MappingRequestedAt, "recently claimed row must not be re-armed before the window elapses")
	assert.Nil(t, load(completed.ID).MappingRequestedAt, "completed mapping (catalog row assigned) must never be re-armed")
}
