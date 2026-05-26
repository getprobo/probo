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

package cookiebanner

import (
	"context"
	"errors"
	"io"
	"net"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

const testPgDSNEnvVar = "PROBO_TEST_PG_URL"

func newTestPgClient(t *testing.T) *pg.Client {
	t.Helper()

	dsn := os.Getenv(testPgDSNEnvVar)
	if dsn == "" {
		t.Skipf("skipping: %s not set (requires a migrated test database)", testPgDSNEnvVar)
	}

	u, err := url.Parse(dsn)
	require.NoError(t, err, "invalid %s value", testPgDSNEnvVar)

	opts := []pg.Option{pg.WithRegisterer(prometheus.NewRegistry())}

	if u.Host != "" {
		host := u.Host
		if u.Port() == "" {
			host = net.JoinHostPort(u.Hostname(), "5432")
		}

		opts = append(opts, pg.WithAddr(host))
	}

	if u.User != nil {
		opts = append(opts, pg.WithUser(u.User.Username()))
		if password, ok := u.User.Password(); ok {
			opts = append(opts, pg.WithPassword(password))
		}
	}

	if len(u.Path) > 1 {
		opts = append(opts, pg.WithDatabase(u.Path[1:]))
	}

	client, err := pg.NewClient(opts...)
	require.NoError(t, err)

	t.Cleanup(func() {
		client.Close()
	})

	return client
}

// workerFixture bootstraps the parent rows the worker's transaction
// needs: an organization, a cookie banner, an uncategorised category,
// and a normal category. Patterns/detected trackers are seeded
// per-test.
type workerFixture struct {
	scope            *coredata.Scope
	organizationID   gid.GID
	banner           coredata.CookieBanner
	uncategorisedID  gid.GID
	normalCategoryID gid.GID
}

func seedWorkerFixture(t *testing.T, ctx context.Context, client *pg.Client) workerFixture {
	t.Helper()

	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	bannerID := gid.New(tenantID, coredata.CookieBannerEntityType)
	uncategorisedID := gid.New(tenantID, coredata.CookieCategoryEntityType)
	normalCategoryID := gid.New(tenantID, coredata.CookieCategoryEntityType)
	now := time.Now().UTC().Truncate(time.Microsecond)

	banner := coredata.CookieBanner{
		ID:                bannerID,
		OrganizationID:    organizationID,
		Name:              "Worker Test Banner",
		Origin:            "https://worker-test-" + bannerID.String() + ".example.com",
		State:             coredata.CookieBannerStateActive,
		CookiePolicyURL:   "https://worker-test.example.com/cookies",
		ConsentExpiryDays: 180,
		ShowBranding:      false,
		DefaultLanguage:   "en",
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		org := &coredata.Organization{
			ID:        organizationID,
			TenantID:  tenantID,
			Name:      "Worker Test Org",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := org.Insert(ctx, tx); err != nil {
			return err
		}

		if err := banner.Insert(ctx, tx, scope); err != nil {
			return err
		}

		uncategorised := &coredata.CookieCategory{
			ID:              uncategorisedID,
			OrganizationID:  organizationID,
			CookieBannerID:  bannerID,
			Name:            "Uncategorised",
			Slug:            "uncategorised",
			Description:     "",
			Kind:            coredata.CookieCategoryKindUncategorised,
			Rank:            0,
			GCMConsentTypes: []string{},
			PostHogConsent:  false,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := uncategorised.Insert(ctx, tx, scope); err != nil {
			return err
		}

		normal := &coredata.CookieCategory{
			ID:              normalCategoryID,
			OrganizationID:  organizationID,
			CookieBannerID:  bannerID,
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
		if err := normal.Insert(ctx, tx, scope); err != nil {
			return err
		}

		return nil
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			if _, err := tx.Exec(ctx, `DELETE FROM detected_trackers WHERE cookie_banner_id = $1`, bannerID); err != nil {
				return err
			}

			if _, err := tx.Exec(ctx, `DELETE FROM tracker_patterns WHERE cookie_banner_id = $1`, bannerID); err != nil {
				return err
			}

			if _, err := tx.Exec(ctx, `DELETE FROM cookie_banner_versions WHERE cookie_banner_id = $1`, bannerID); err != nil {
				return err
			}

			if _, err := tx.Exec(ctx, `DELETE FROM cookie_categories WHERE cookie_banner_id = $1`, bannerID); err != nil {
				return err
			}

			if _, err := tx.Exec(ctx, `DELETE FROM cookie_banners WHERE id = $1`, bannerID); err != nil {
				return err
			}

			if _, err := tx.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, organizationID); err != nil {
				return err
			}

			return nil
		})
	})

	return workerFixture{
		scope:            scope,
		organizationID:   organizationID,
		banner:           banner,
		uncategorisedID:  uncategorisedID,
		normalCategoryID: normalCategoryID,
	}
}

func newExactPattern(
	fx workerFixture,
	pattern string,
	categoryID gid.GID,
	source coredata.CookieSource,
	maxAge *int,
) *coredata.TrackerPattern {
	now := time.Now().UTC().Truncate(time.Microsecond)

	return &coredata.TrackerPattern{
		ID:                 gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType),
		OrganizationID:     fx.organizationID,
		CookieBannerID:     fx.banner.ID,
		CookieCategoryID:   categoryID,
		TrackerType:        coredata.TrackerTypeCookie,
		Pattern:            pattern,
		MatchType:          coredata.TrackerPatternMatchTypeExact,
		DisplayName:        pattern,
		Description:        "",
		MaxAgeSeconds:      maxAge,
		Source:             &source,
		MappingRequestedAt: &now,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

func newGlobInCategory(
	fx workerFixture,
	pattern string,
	categoryID gid.GID,
	source coredata.CookieSource,
	maxAge *int,
) *coredata.TrackerPattern {
	now := time.Now().UTC().Truncate(time.Microsecond)

	return &coredata.TrackerPattern{
		ID:                 gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType),
		OrganizationID:     fx.organizationID,
		CookieBannerID:     fx.banner.ID,
		CookieCategoryID:   categoryID,
		TrackerType:        coredata.TrackerTypeCookie,
		Pattern:            pattern,
		MatchType:          coredata.TrackerPatternMatchTypeGlob,
		DisplayName:        pattern,
		Description:        "",
		MaxAgeSeconds:      maxAge,
		Source:             &source,
		MappingRequestedAt: &now,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

func newTestHandler(client *pg.Client) *patternAnalysisHandler {
	return &patternAnalysisHandler{
		svc:    NewService(client, false),
		pg:     client,
		logger: log.NewLogger(log.WithOutput(io.Discard)),
	}
}

// TestPatternAnalysisWorker_PromotesSourceOnExistingGlob seeds a
// banner with a PRE_EXISTING `_ga_*` glob in the analytics category,
// adds three SCRIPT-source exacts that group under the same template,
// and asserts that running the worker promotes the glob's source to
// SCRIPT. This guards against the original regression where
// bestSource was only honoured on first insert and subsequent batches
// could never promote.
func TestPatternAnalysisWorker_PromotesSourceOnExistingGlob(t *testing.T) {
	t.Parallel()

	client := newTestPgClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)

	maxAge := 7 * 24 * 3600

	existingGlob := newGlobInCategory(
		fx,
		"_ga_*",
		fx.normalCategoryID,
		coredata.CookieSourcePreExisting,
		&maxAge,
	)
	exacts := []*coredata.TrackerPattern{
		newExactPattern(fx, "_ga_abc123", fx.normalCategoryID, coredata.CookieSourceScript, &maxAge),
		newExactPattern(fx, "_ga_def456", fx.normalCategoryID, coredata.CookieSourceScript, &maxAge),
		newExactPattern(fx, "_ga_xyz789", fx.normalCategoryID, coredata.CookieSourceScript, &maxAge),
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := existingGlob.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		for _, ep := range exacts {
			if err := ep.Insert(ctx, tx, fx.scope); err != nil {
				return err
			}
		}

		return nil
	}))

	h := newTestHandler(client)
	require.NoError(t, h.Process(ctx, fx.banner))

	loaded := &coredata.TrackerPattern{}

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return loaded.LoadByBannerIDTypeAndPattern(
			ctx,
			conn,
			fx.scope,
			fx.banner.ID,
			coredata.TrackerTypeCookie,
			"_ga_*",
			&maxAge,
		)
	}))

	require.NotNil(t, loaded.Source)
	assert.Equal(t, coredata.CookieSourceScript, *loaded.Source, "glob source must be promoted to SCRIPT")
	assert.Equal(t, existingGlob.ID, loaded.ID, "the existing glob row must be reused, not replaced")
	assert.Equal(t, fx.normalCategoryID, loaded.CookieCategoryID, "category must not be touched")

	var remainingExacts coredata.TrackerPatterns

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return remainingExacts.LoadAllByCookieBannerID(
			ctx,
			conn,
			fx.scope,
			fx.banner.ID,
			coredata.NewTrackerPatternFilter(new(coredata.TrackerPatternMatchTypeExact), nil, new(false)),
			nil,
		)
	}))
	assert.Empty(t, remainingExacts, "all three exacts must be relinked and deleted")
}

// TestPatternAnalysisWorker_AdoptionTriggersDraftVersion seeds a
// banner with a categorised `_ga_*` glob and uncategorised exacts
// that match it. The merge loop must skip the relink (different
// category) but adoptUncategorisedPatterns must re-home the exacts;
// the worker must then create a draft banner version reflecting the
// consent state change. This guards against the original bug where
// the adopted bool was discarded.
func TestPatternAnalysisWorker_AdoptionTriggersDraftVersion(t *testing.T) {
	t.Parallel()

	client := newTestPgClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)

	maxAge := 7 * 24 * 3600

	existingGlob := newGlobInCategory(
		fx,
		"_ga_*",
		fx.normalCategoryID,
		coredata.CookieSourceScript,
		&maxAge,
	)
	exacts := []*coredata.TrackerPattern{
		newExactPattern(fx, "_ga_abc123", fx.uncategorisedID, coredata.CookieSourcePreExisting, &maxAge),
		newExactPattern(fx, "_ga_def456", fx.uncategorisedID, coredata.CookieSourcePreExisting, &maxAge),
		newExactPattern(fx, "_ga_xyz789", fx.uncategorisedID, coredata.CookieSourcePreExisting, &maxAge),
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := existingGlob.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		for _, ep := range exacts {
			if err := ep.Insert(ctx, tx, fx.scope); err != nil {
				return err
			}
		}

		return nil
	}))

	h := newTestHandler(client)
	require.NoError(t, h.Process(ctx, fx.banner))

	loaded := &coredata.TrackerPattern{}

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return loaded.LoadByBannerIDTypeAndPattern(
			ctx,
			conn,
			fx.scope,
			fx.banner.ID,
			coredata.TrackerTypeCookie,
			"_ga_*",
			&maxAge,
		)
	}))
	assert.Equal(t, fx.normalCategoryID, loaded.CookieCategoryID, "user-set category must not be overwritten by the worker")
	assert.Equal(t, existingGlob.ID, loaded.ID, "existing glob row must be reused")

	var remainingExacts coredata.TrackerPatterns

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return remainingExacts.LoadAllByCookieBannerID(
			ctx,
			conn,
			fx.scope,
			fx.banner.ID,
			coredata.NewTrackerPatternFilter(new(coredata.TrackerPatternMatchTypeExact), nil, new(false)),
			nil,
		)
	}))
	assert.Empty(t, remainingExacts, "adoptUncategorisedPatterns must absorb the uncategorised exacts into the existing glob")

	latest := &coredata.CookieBannerVersion{}

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return latest.LoadLatestByCookieBannerID(ctx, conn, fx.scope, fx.banner.ID)
	}))
	assert.Equal(t, coredata.CookieBannerVersionStateDraft, latest.State, "adoption must trigger a draft version")
}

// TestPatternAnalysisWorker_AdoptionPromotesSourceCrossCategory
// seeds a banner with a PRE_EXISTING `ph_phc_*_posthog` glob already
// placed in a user-set category (analytics) and a single SCRIPT-source
// exact `ph_phc_<id>_posthog` in uncategorised. The merge loop hits
// the cross-category skip branch (the existing slot belongs to a
// recategorised glob), so promotion can only happen via the
// adoption path. This guards against the gap where last_matched_at
// was refreshed on the glob but its source stayed at PRE_EXISTING
// despite the new SDK-observed signal.
func TestPatternAnalysisWorker_AdoptionPromotesSourceCrossCategory(t *testing.T) {
	t.Parallel()

	client := newTestPgClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)

	maxAge := 365 * 24 * 3600

	existingGlob := newGlobInCategory(
		fx,
		"ph_phc_*_posthog",
		fx.normalCategoryID,
		coredata.CookieSourcePreExisting,
		&maxAge,
	)
	uncategorisedExact := newExactPattern(
		fx,
		"ph_phc_XBwJ2pHAf0MoYgh3TNZK32Qk7zLlTldhk4p9llGtZMN_posthog",
		fx.uncategorisedID,
		coredata.CookieSourceScript,
		&maxAge,
	)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := existingGlob.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		return uncategorisedExact.Insert(ctx, tx, fx.scope)
	}))

	h := newTestHandler(client)
	require.NoError(t, h.Process(ctx, fx.banner))

	loaded := &coredata.TrackerPattern{}

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return loaded.LoadByBannerIDTypeAndPattern(
			ctx,
			conn,
			fx.scope,
			fx.banner.ID,
			coredata.TrackerTypeCookie,
			"ph_phc_*_posthog",
			&maxAge,
		)
	}))

	require.NotNil(t, loaded.Source)
	assert.Equal(t, coredata.CookieSourceScript, *loaded.Source, "adoption must promote PRE_EXISTING glob to SCRIPT when a stronger exact is absorbed")
	assert.Equal(t, existingGlob.ID, loaded.ID, "the existing glob row must be reused, not replaced")
	assert.Equal(t, fx.normalCategoryID, loaded.CookieCategoryID, "user-set category must not be overwritten by the worker")

	var remainingExacts coredata.TrackerPatterns

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return remainingExacts.LoadAllByCookieBannerID(
			ctx,
			conn,
			fx.scope,
			fx.banner.ID,
			coredata.NewTrackerPatternFilter(new(coredata.TrackerPatternMatchTypeExact), nil, new(false)),
			nil,
		)
	}))
	assert.Empty(t, remainingExacts, "the uncategorised exact must be adopted into the existing glob")
}

// TestReportDetectedTrackers_PromotesSourceOnExistingGlob seeds a
// banner with a PRE_EXISTING `ph_phc_*_posthog` glob in
// uncategorised, then reports a SCRIPT-source cookie whose name
// globMatches the existing pattern (e.g. a posthog instance ID).
// FindMatchingPattern in reportDetectedTracker links the
// detected_tracker straight to the glob, so no new exact is
// created and the merge/adoption loops in
// patternAnalysisHandler.Process never see the new signal. Without
// the in-line PromoteSource call in reportDetectedTracker, only
// last_matched_at would advance — source would stay stuck at
// PRE_EXISTING despite the new SDK-observed evidence. This test
// pins the same-category promotion path; the cross-category gap
// is covered by TestPatternAnalysisWorker_AdoptionPromotesSourceCrossCategory.
func TestReportDetectedTrackers_PromotesSourceOnExistingGlob(t *testing.T) {
	t.Parallel()

	client := newTestPgClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)

	maxAge := 365 * 24 * 3600

	existingGlob := newGlobInCategory(
		fx,
		"ph_phc_*_posthog",
		fx.uncategorisedID,
		coredata.CookieSourcePreExisting,
		&maxAge,
	)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return existingGlob.Insert(ctx, tx, fx.scope)
	}))

	svc := NewService(client, false)

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		Cookies: []DetectedCookie{
			{
				Name:          "ph_phc_XBwJ2pHAf0MoYgh3TNZK32Qk7zLlTldhk4p9llGtZMN_posthog",
				MaxAgeSeconds: &maxAge,
				Source:        coredata.CookieSourceScript,
			},
		},
	}))

	loaded := &coredata.TrackerPattern{}

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return loaded.LoadByID(ctx, conn, fx.scope, existingGlob.ID)
	}))

	require.NotNil(t, loaded.Source)
	assert.Equal(t, coredata.CookieSourceScript, *loaded.Source, "reportDetectedTracker must promote a PRE_EXISTING glob to SCRIPT when a stronger detection matches it")
	assert.NotNil(t, loaded.LastMatchedAt, "matched detections must bump last_matched_at")
	assert.Equal(t, fx.uncategorisedID, loaded.CookieCategoryID, "category must not move")

	var exacts coredata.TrackerPatterns

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return exacts.LoadAllByCookieBannerID(
			ctx,
			conn,
			fx.scope,
			fx.banner.ID,
			coredata.NewTrackerPatternFilter(new(coredata.TrackerPatternMatchTypeExact), nil, new(false)),
			nil,
		)
	}))
	assert.Empty(t, exacts, "the detected cookie globMatches the existing glob; no exact pattern must be created")
}

// TestPatternAnalysisWorker_MergeWithoutAdoptionSkipsDraftVersion
// asserts the inverse: when the worker only consolidates exacts into
// a glob in their own category (no consent transition), no draft
// version is created. This guards against the prior over-eager
// consentChanged flag, which produced redundant draft versions on
// every merge into a non-uncategorised category.
func TestPatternAnalysisWorker_MergeWithoutAdoptionSkipsDraftVersion(t *testing.T) {
	t.Parallel()

	client := newTestPgClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)

	maxAge := 7 * 24 * 3600

	exacts := []*coredata.TrackerPattern{
		newExactPattern(fx, "_ga_abc123", fx.normalCategoryID, coredata.CookieSourceScript, &maxAge),
		newExactPattern(fx, "_ga_def456", fx.normalCategoryID, coredata.CookieSourceScript, &maxAge),
		newExactPattern(fx, "_ga_xyz789", fx.normalCategoryID, coredata.CookieSourceScript, &maxAge),
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		for _, ep := range exacts {
			if err := ep.Insert(ctx, tx, fx.scope); err != nil {
				return err
			}
		}

		return nil
	}))

	h := newTestHandler(client)
	require.NoError(t, h.Process(ctx, fx.banner))

	var globs coredata.TrackerPatterns

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return globs.LoadAllByCookieBannerID(
			ctx,
			conn,
			fx.scope,
			fx.banner.ID,
			coredata.NewTrackerPatternFilter(new(coredata.TrackerPatternMatchTypeGlob), nil, new(false)),
			nil,
		)
	}))
	require.Len(t, globs, 1, "the three exacts must consolidate into a single glob")
	assert.Equal(t, "_ga_*", globs[0].Pattern)
	assert.Equal(t, fx.normalCategoryID, globs[0].CookieCategoryID)

	latest := &coredata.CookieBannerVersion{}
	err := client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return latest.LoadLatestByCookieBannerID(ctx, conn, fx.scope, fx.banner.ID)
	})
	assert.ErrorIs(t, err, coredata.ErrResourceNotFound, "merge alone must not create a draft version")
	// Sanity-check the negative assertion: if the lookup unexpectedly
	// succeeds, fail with a clearer message than the bare error mismatch.
	if !errors.Is(err, coredata.ErrResourceNotFound) {
		t.Fatalf("merge-only run unexpectedly produced a banner version: state=%s", latest.State)
	}
}
