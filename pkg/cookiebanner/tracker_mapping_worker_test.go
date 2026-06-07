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
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// promotionFixture extends workerFixture with a CommonThirdParty and a
// CommonTrackerPattern linking the catalog to the test pattern. It is
// the minimum scaffolding resolveOrgThirdParty needs to run end-to-end.
type promotionFixture struct {
	workerFixture
	commonThirdParty   coredata.CommonThirdParty
	commonPatternID    gid.GID
	trackerPattern     coredata.TrackerPattern
	commonThirdPartyID gid.GID
}

func seedPromotionFixture(t *testing.T, ctx context.Context, client *pg.Client) promotionFixture {
	t.Helper()

	fx := seedWorkerFixture(t, ctx, client)
	now := time.Now().UTC().Truncate(time.Microsecond)

	// common_third_parties (name/slug) and common_tracker_patterns
	// (tracker_type, pattern, max_age_seconds) are global, NOT
	// tenant-scoped, and both carry unique indexes. Tests run in
	// parallel, so the catalog rows must be unique per fixture or
	// concurrent runs collide. The tenant id is unique per fixture and
	// makes a stable, collision-free suffix.
	suffix := fx.scope.GetTenantID().String()
	patternName := "_ga_" + suffix

	commonThirdPartyID := gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType)
	commonThirdParty := coredata.CommonThirdParty{
		ID:             commonThirdPartyID,
		Name:           "Google " + suffix,
		Slug:           "google-" + suffix,
		Category:       coredata.ThirdPartyCategoryAnalytics,
		WebsiteURL:     new("https://google.com"),
		Certifications: []string{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	commonPattern := coredata.CommonTrackerPattern{
		ID:                 gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
		CommonThirdPartyID: &commonThirdPartyID,
		TrackerType:        coredata.TrackerTypeCookie,
		Pattern:            patternName,
		MatchType:          coredata.TrackerPatternMatchTypeExact,
		Description:        "",
		Confidence:         0.9,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	pattern := coredata.TrackerPattern{
		ID:                     gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType),
		OrganizationID:         fx.organizationID,
		CookieBannerID:         fx.banner.ID,
		CookieCategoryID:       fx.normalCategoryID,
		CommonTrackerPatternID: &commonPattern.ID,
		TrackerType:            coredata.TrackerTypeCookie,
		Pattern:                patternName,
		MatchType:              coredata.TrackerPatternMatchTypeExact,
		DisplayName:            patternName,
		Description:            "",
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := commonThirdParty.Insert(ctx, tx); err != nil {
			return err
		}

		if _, err := commonPattern.Upsert(ctx, tx); err != nil {
			return err
		}

		return pattern.Insert(ctx, tx, fx.scope)
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			if _, err := tx.Exec(ctx, `DELETE FROM common_third_party_domains WHERE common_third_party_id = $1`, commonThirdPartyID); err != nil {
				return err
			}

			if _, err := tx.Exec(ctx, `DELETE FROM common_tracker_patterns WHERE id = $1`, commonPattern.ID); err != nil {
				return err
			}

			if _, err := tx.Exec(ctx, `DELETE FROM common_third_parties WHERE id = $1`, commonThirdPartyID); err != nil {
				return err
			}

			if _, err := tx.Exec(ctx, `DELETE FROM third_parties WHERE organization_id = $1`, fx.organizationID); err != nil {
				return err
			}

			return nil
		})
	})

	return promotionFixture{
		workerFixture:      fx,
		commonThirdParty:   commonThirdParty,
		commonPatternID:    commonPattern.ID,
		commonThirdPartyID: commonThirdPartyID,
		trackerPattern:     pattern,
	}
}

func newMappingHandler(client *pg.Client) *trackerMappingHandler {
	return &trackerMappingHandler{
		pg:     client,
		logger: log.NewLogger(log.WithOutput(io.Discard)),
	}
}

// promote runs resolveOrgThirdParty, which manages its own short
// transactions internally (creation gating is derived from the
// pattern's category, not passed in).
func promote(
	t *testing.T,
	ctx context.Context,
	h *trackerMappingHandler,
	tp coredata.TrackerPattern,
	commonThirdPartyID gid.GID,
) *gid.GID {
	t.Helper()

	got, err := h.resolveOrgThirdParty(ctx, tp, commonThirdPartyID)
	require.NoError(t, err)

	return got
}

func TestPromoteThirdParty_ExactCommonLink(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedPromotionFixture(t, ctx, client)

	now := time.Now().UTC().Truncate(time.Microsecond)
	existing := coredata.ThirdParty{
		ID:                 gid.New(fx.scope.GetTenantID(), coredata.ThirdPartyEntityType),
		OrganizationID:     fx.organizationID,
		CommonThirdPartyID: &fx.commonThirdPartyID,
		Name:               "Google LLC",
		Category:           coredata.ThirdPartyCategoryAnalytics,
		Certifications:     []string{},
		Countries:          coredata.CountryCodes{},
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return existing.Insert(ctx, tx, fx.scope)
	}))

	got := promote(t, ctx, newMappingHandler(client), fx.trackerPattern, fx.commonThirdPartyID)

	require.NotNil(t, got)
	assert.Equal(t, existing.ID, *got, "should return the existing org ThirdParty linked by common id")
}

func TestPromoteThirdParty_HeuristicMatch(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedPromotionFixture(t, ctx, client)

	now := time.Now().UTC().Truncate(time.Microsecond)
	// Append a corporate suffix to the catalog name so the heuristic
	// matches on the suffix-stripped name (score 0.9) rather than an
	// exact link.
	manualEntry := coredata.ThirdParty{
		ID:             gid.New(fx.scope.GetTenantID(), coredata.ThirdPartyEntityType),
		OrganizationID: fx.organizationID,
		Name:           fx.commonThirdParty.Name + " LLC",
		Category:       coredata.ThirdPartyCategoryAnalytics,
		Certifications: []string{},
		Countries:      coredata.CountryCodes{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return manualEntry.Insert(ctx, tx, fx.scope)
	}))

	got := promote(t, ctx, newMappingHandler(client), fx.trackerPattern, fx.commonThirdPartyID)

	require.NotNil(t, got)
	assert.Equal(t, manualEntry.ID, *got, "heuristic match should return the manually-entered ThirdParty")

	var reloaded coredata.ThirdParty

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloaded.LoadByID(ctx, conn, fx.scope, manualEntry.ID)
	}))

	require.NotNil(t, reloaded.CommonThirdPartyID, "matched row must be tagged with common_third_party_id")
	assert.Equal(t, fx.commonThirdPartyID, *reloaded.CommonThirdPartyID)
}

func TestPromoteThirdParty_FallbackCreate(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedPromotionFixture(t, ctx, client)

	got := promote(t, ctx, newMappingHandler(client), fx.trackerPattern, fx.commonThirdPartyID)

	require.NotNil(t, got, "fallback should create a new ThirdParty")

	var reloaded coredata.ThirdParty

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloaded.LoadByID(ctx, conn, fx.scope, *got)
	}))

	assert.Equal(t, fx.organizationID, reloaded.OrganizationID)
	assert.Equal(t, fx.commonThirdParty.Name, reloaded.Name)
	require.NotNil(t, reloaded.CommonThirdPartyID)
	assert.Equal(t, fx.commonThirdPartyID, *reloaded.CommonThirdPartyID)
	assert.Equal(t, coredata.ThirdPartyCategoryAnalytics, reloaded.Category)
	assert.True(t, reloaded.FirstLevel)
	assert.False(t, reloaded.ShowOnTrustCenter)
}

// TestResolveOrgThirdParty_CreationGated asserts that when no existing
// org ThirdParty matches the catalog third party, creating a new one is
// suppressed for an uncategorised pattern (creation gating is derived
// from the pattern's category) and proceeds for a categorised one.
func TestResolveOrgThirdParty_CreationGated(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedPromotionFixture(t, ctx, client)

	gatedPattern := fx.trackerPattern
	gatedPattern.CookieCategoryID = fx.uncategorisedID

	gated := promote(t, ctx, newMappingHandler(client), gatedPattern, fx.commonThirdPartyID)
	assert.Nil(t, gated, "creation must be suppressed for an uncategorised pattern with nothing to link")

	allowed := promote(t, ctx, newMappingHandler(client), fx.trackerPattern, fx.commonThirdPartyID)
	require.NotNil(t, allowed, "creation must proceed for a categorised pattern")
}

// TestProcess_PreservesCatalogMappingOnReTrigger asserts that when
// Process is called for a pattern that already carries a
// common_tracker_pattern_id, the catalog pipeline is skipped and the
// existing catalog link is preserved verbatim.
func TestProcess_PreservesCatalogMappingOnReTrigger(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedPromotionFixture(t, ctx, client)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return fx.trackerPattern.SetMappingRequested(ctx, tx)
	}))

	h := newMappingHandler(client)
	require.NoError(t, h.Process(ctx, fx.trackerPattern))

	var reloaded coredata.TrackerPattern

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloaded.LoadByID(ctx, conn, fx.scope, fx.trackerPattern.ID)
	}))

	require.NotNil(t, reloaded.CommonTrackerPatternID, "common tracker pattern link must be preserved")
	assert.Equal(t, fx.commonPatternID, *reloaded.CommonTrackerPatternID)
	require.NotNil(t, reloaded.ThirdPartyID, "the worker should have promoted to an org ThirdParty")
}

// TestProcess_UncategorisedPatternIsNotPromoted asserts that a pattern
// still in the uncategorised category gets its catalog mapping
// resolved but is NOT promoted to an org ThirdParty.
func TestProcess_UncategorisedPatternIsNotPromoted(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedPromotionFixture(t, ctx, client)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		_, err := tx.Exec(
			ctx,
			`UPDATE tracker_patterns
			   SET cookie_category_id = $1,
			       mapping_requested_at = $2
			 WHERE id = $3`,
			fx.uncategorisedID,
			time.Now().UTC().Truncate(time.Microsecond),
			fx.trackerPattern.ID,
		)

		return err
	}))

	var reloadedBefore coredata.TrackerPattern

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloadedBefore.LoadByID(ctx, conn, fx.scope, fx.trackerPattern.ID)
	}))

	h := newMappingHandler(client)
	require.NoError(t, h.Process(ctx, reloadedBefore))

	var reloaded coredata.TrackerPattern

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloaded.LoadByID(ctx, conn, fx.scope, fx.trackerPattern.ID)
	}))

	require.NotNil(t, reloaded.CommonTrackerPatternID, "catalog mapping must still be resolved")
	assert.Equal(t, fx.commonPatternID, *reloaded.CommonTrackerPatternID)
	assert.Nil(t, reloaded.ThirdPartyID, "uncategorised pattern must not be promoted to org ThirdParty")
}

// TestProcess_ExtensionPatternIsNotPromoted asserts that even when a
// pattern has a catalog link, a Source=EXTENSION pattern stays
// un-promoted.
func TestProcess_ExtensionPatternIsNotPromoted(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedPromotionFixture(t, ctx, client)

	now := time.Now().UTC().Truncate(time.Microsecond)
	source := coredata.CookieSourceExtension

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		_, err := tx.Exec(
			ctx,
			`UPDATE tracker_patterns
			   SET source = $1,
			       mapping_requested_at = $2
			 WHERE id = $3`,
			source,
			now,
			fx.trackerPattern.ID,
		)

		return err
	}))

	var reloadedBefore coredata.TrackerPattern

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloadedBefore.LoadByID(ctx, conn, fx.scope, fx.trackerPattern.ID)
	}))

	h := newMappingHandler(client)
	require.NoError(t, h.Process(ctx, reloadedBefore))

	var reloaded coredata.TrackerPattern

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloaded.LoadByID(ctx, conn, fx.scope, fx.trackerPattern.ID)
	}))

	assert.Nil(t, reloaded.ThirdPartyID, "EXTENSION-sourced pattern must not be promoted")
}

// TestProcess_NoOpWhenAlreadyPromoted asserts that re-running the
// worker on a pattern that already has a third_party_id leaves the
// row alone (the guard in Process).
func TestProcess_NoOpWhenAlreadyPromoted(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedPromotionFixture(t, ctx, client)

	now := time.Now().UTC().Truncate(time.Microsecond)
	preExisting := coredata.ThirdParty{
		ID:                 gid.New(fx.scope.GetTenantID(), coredata.ThirdPartyEntityType),
		OrganizationID:     fx.organizationID,
		CommonThirdPartyID: &fx.commonThirdPartyID,
		Name:               "Google",
		Category:           coredata.ThirdPartyCategoryAnalytics,
		Certifications:     []string{},
		Countries:          coredata.CountryCodes{},
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := preExisting.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		fx.trackerPattern.ThirdPartyID = &preExisting.ID

		_, err := tx.Exec(
			ctx,
			`UPDATE tracker_patterns
			   SET third_party_id = $1,
			       mapping_requested_at = $2
			 WHERE id = $3`,
			preExisting.ID,
			now,
			fx.trackerPattern.ID,
		)

		return err
	}))

	var reloadedBefore coredata.TrackerPattern

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloadedBefore.LoadByID(ctx, conn, fx.scope, fx.trackerPattern.ID)
	}))

	h := newMappingHandler(client)
	require.NoError(t, h.Process(ctx, reloadedBefore))

	var reloaded coredata.TrackerPattern

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloaded.LoadByID(ctx, conn, fx.scope, fx.trackerPattern.ID)
	}))

	require.NotNil(t, reloaded.ThirdPartyID)
	assert.Equal(t, preExisting.ID, *reloaded.ThirdPartyID, "third_party_id must not be overwritten")
}

func TestMatchBySiblingOrigin_SiblingWithThirdPartyID(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedPromotionFixture(t, ctx, client)

	now := time.Now().UTC().Truncate(time.Microsecond)

	orgThirdParty := coredata.ThirdParty{
		ID:                 gid.New(fx.scope.GetTenantID(), coredata.ThirdPartyEntityType),
		OrganizationID:     fx.organizationID,
		CommonThirdPartyID: &fx.commonThirdPartyID,
		Name:               "Google LLC",
		Category:           coredata.ThirdPartyCategoryAnalytics,
		Certifications:     []string{},
		Countries:          coredata.CountryCodes{},
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	siblingPattern := coredata.TrackerPattern{
		ID:                     gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType),
		OrganizationID:         fx.organizationID,
		CookieBannerID:         fx.banner.ID,
		CookieCategoryID:       fx.normalCategoryID,
		CommonTrackerPatternID: &fx.commonPatternID,
		ThirdPartyID:           &orgThirdParty.ID,
		TrackerType:            coredata.TrackerTypeCookie,
		Pattern:                "_gid",
		MatchType:              coredata.TrackerPatternMatchTypeExact,
		DisplayName:            "_gid",
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	unmappedPattern := coredata.TrackerPattern{
		ID:               gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType),
		OrganizationID:   fx.organizationID,
		CookieBannerID:   fx.banner.ID,
		CookieCategoryID: fx.normalCategoryID,
		TrackerType:      coredata.TrackerTypeCookie,
		Pattern:          "_ga_unknown",
		MatchType:        coredata.TrackerPatternMatchTypeExact,
		DisplayName:      "_ga_unknown",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// Detected trackers store the eTLD+1 (uri.ExtractDomain), so the
	// sibling lookup matches on that exact value. Use a vendor-specific
	// domain rather than shared infrastructure (e.g. googletagmanager.com),
	// which resolveDeterministic strips before sibling grouping.
	initiatorDomain := "google-analytics.com"
	siblingDetected := coredata.DetectedTracker{
		ID:               gid.New(fx.scope.GetTenantID(), coredata.DetectedTrackerEntityType),
		CookieBannerID:   fx.banner.ID,
		TrackerPatternID: &siblingPattern.ID,
		TrackerType:      coredata.TrackerTypeCookie,
		Identifier:       "_gid",
		InitiatorDomain:  &initiatorDomain,
		LastDetectedAt:   now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	unmappedDetected := coredata.DetectedTracker{
		ID:               gid.New(fx.scope.GetTenantID(), coredata.DetectedTrackerEntityType),
		CookieBannerID:   fx.banner.ID,
		TrackerPatternID: &unmappedPattern.ID,
		TrackerType:      coredata.TrackerTypeCookie,
		Identifier:       "_ga_unknown",
		InitiatorDomain:  &initiatorDomain,
		LastDetectedAt:   now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := orgThirdParty.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		if err := siblingPattern.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		if err := unmappedPattern.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		if _, err := siblingDetected.Upsert(ctx, tx, fx.scope); err != nil {
			return err
		}

		if _, err := unmappedDetected.Upsert(ctx, tx, fx.scope); err != nil {
			return err
		}

		return nil
	}))

	h := newMappingHandler(client)

	var got *catalogMatch

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var err error

		got, err = h.matchBySiblingOrigin(ctx, tx, unmappedPattern, []string{"google-analytics.com"})

		return err
	}))

	require.NotNil(t, got, "sibling origin match should return a catalog match")
	require.NotNil(t, got.commonPatternID, "sibling origin match should return a common tracker pattern ID")
	require.NotNil(t, got.thirdPartyID, "sibling origin match should surface the sibling's org third party directly")
	assert.Equal(t, orgThirdParty.ID, *got.thirdPartyID)

	var commonPattern coredata.CommonTrackerPattern

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return commonPattern.LoadByID(ctx, conn, *got.commonPatternID)
	}))

	require.NotNil(t, commonPattern.CommonThirdPartyID)
	assert.Equal(t, fx.commonThirdPartyID, *commonPattern.CommonThirdPartyID)
	assert.Equal(t, float32(0.7), commonPattern.Confidence)
}

func TestMatchBySiblingOrigin_AmbiguousThirdParties(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedPromotionFixture(t, ctx, client)

	now := time.Now().UTC().Truncate(time.Microsecond)

	otherSuffix := fx.scope.GetTenantID().String()
	otherCommonThirdPartyID := gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType)
	otherCommonThirdParty := coredata.CommonThirdParty{
		ID:             otherCommonThirdPartyID,
		Name:           "Facebook " + otherSuffix,
		Slug:           "facebook-" + otherSuffix,
		Category:       coredata.ThirdPartyCategoryMarketing,
		Certifications: []string{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	orgThirdPartyA := coredata.ThirdParty{
		ID:                 gid.New(fx.scope.GetTenantID(), coredata.ThirdPartyEntityType),
		OrganizationID:     fx.organizationID,
		CommonThirdPartyID: &fx.commonThirdPartyID,
		Name:               "Google",
		Category:           coredata.ThirdPartyCategoryAnalytics,
		Certifications:     []string{},
		Countries:          coredata.CountryCodes{},
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	orgThirdPartyB := coredata.ThirdParty{
		ID:                 gid.New(fx.scope.GetTenantID(), coredata.ThirdPartyEntityType),
		OrganizationID:     fx.organizationID,
		CommonThirdPartyID: &otherCommonThirdPartyID,
		Name:               "Facebook",
		Category:           coredata.ThirdPartyCategoryMarketing,
		Certifications:     []string{},
		Countries:          coredata.CountryCodes{},
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	siblingA := coredata.TrackerPattern{
		ID:               gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType),
		OrganizationID:   fx.organizationID,
		CookieBannerID:   fx.banner.ID,
		CookieCategoryID: fx.normalCategoryID,
		ThirdPartyID:     &orgThirdPartyA.ID,
		TrackerType:      coredata.TrackerTypeCookie,
		Pattern:          "sibling_a",
		MatchType:        coredata.TrackerPatternMatchTypeExact,
		DisplayName:      "sibling_a",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	siblingB := coredata.TrackerPattern{
		ID:               gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType),
		OrganizationID:   fx.organizationID,
		CookieBannerID:   fx.banner.ID,
		CookieCategoryID: fx.normalCategoryID,
		ThirdPartyID:     &orgThirdPartyB.ID,
		TrackerType:      coredata.TrackerTypeCookie,
		Pattern:          "sibling_b",
		MatchType:        coredata.TrackerPatternMatchTypeExact,
		DisplayName:      "sibling_b",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	unmappedPattern := coredata.TrackerPattern{
		ID:               gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType),
		OrganizationID:   fx.organizationID,
		CookieBannerID:   fx.banner.ID,
		CookieCategoryID: fx.normalCategoryID,
		TrackerType:      coredata.TrackerTypeCookie,
		Pattern:          "ambiguous_test",
		MatchType:        coredata.TrackerPatternMatchTypeExact,
		DisplayName:      "ambiguous_test",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	sharedDomain := "shared-tracker.com"

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := otherCommonThirdParty.Insert(ctx, tx); err != nil {
			return err
		}

		if err := orgThirdPartyA.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		if err := orgThirdPartyB.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		if err := siblingA.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		if err := siblingB.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		if err := unmappedPattern.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		detA := coredata.DetectedTracker{
			ID:               gid.New(fx.scope.GetTenantID(), coredata.DetectedTrackerEntityType),
			CookieBannerID:   fx.banner.ID,
			TrackerPatternID: &siblingA.ID,
			TrackerType:      coredata.TrackerTypeCookie,
			Identifier:       "sibling_a",
			InitiatorDomain:  &sharedDomain,
			LastDetectedAt:   now,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if _, err := detA.Upsert(ctx, tx, fx.scope); err != nil {
			return err
		}

		detB := coredata.DetectedTracker{
			ID:               gid.New(fx.scope.GetTenantID(), coredata.DetectedTrackerEntityType),
			CookieBannerID:   fx.banner.ID,
			TrackerPatternID: &siblingB.ID,
			TrackerType:      coredata.TrackerTypeCookie,
			Identifier:       "sibling_b",
			InitiatorDomain:  &sharedDomain,
			LastDetectedAt:   now,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if _, err := detB.Upsert(ctx, tx, fx.scope); err != nil {
			return err
		}

		detUnmapped := coredata.DetectedTracker{
			ID:               gid.New(fx.scope.GetTenantID(), coredata.DetectedTrackerEntityType),
			CookieBannerID:   fx.banner.ID,
			TrackerPatternID: &unmappedPattern.ID,
			TrackerType:      coredata.TrackerTypeCookie,
			Identifier:       "ambiguous_test",
			InitiatorDomain:  &sharedDomain,
			LastDetectedAt:   now,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if _, err := detUnmapped.Upsert(ctx, tx, fx.scope); err != nil {
			return err
		}

		return nil
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			_, _ = tx.Exec(ctx, `DELETE FROM common_third_parties WHERE id = $1`, otherCommonThirdPartyID)

			return nil
		})
	})

	h := newMappingHandler(client)

	var got *catalogMatch

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var err error

		got, err = h.matchBySiblingOrigin(ctx, tx, unmappedPattern, []string{"shared-tracker.com"})

		return err
	}))

	assert.Nil(t, got, "ambiguous siblings mapping to different third parties should return nil")
}

func TestMatchBySiblingOrigin_NoSiblings(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)

	now := time.Now().UTC().Truncate(time.Microsecond)

	unmappedPattern := coredata.TrackerPattern{
		ID:               gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType),
		OrganizationID:   fx.organizationID,
		CookieBannerID:   fx.banner.ID,
		CookieCategoryID: fx.normalCategoryID,
		TrackerType:      coredata.TrackerTypeCookie,
		Pattern:          "lonely_cookie",
		MatchType:        coredata.TrackerPatternMatchTypeExact,
		DisplayName:      "lonely_cookie",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return unmappedPattern.Insert(ctx, tx, fx.scope)
	}))

	h := newMappingHandler(client)

	var got *catalogMatch

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var err error

		got, err = h.matchBySiblingOrigin(ctx, tx, unmappedPattern, []string{"unique-domain.com"})

		return err
	}))

	assert.Nil(t, got, "no siblings sharing the domain should return nil")
}

func TestMatchBySiblingOrigin_EmptyDomains(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)

	now := time.Now().UTC().Truncate(time.Microsecond)

	pattern := coredata.TrackerPattern{
		ID:               gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType),
		OrganizationID:   fx.organizationID,
		CookieBannerID:   fx.banner.ID,
		CookieCategoryID: fx.normalCategoryID,
		TrackerType:      coredata.TrackerTypeCookie,
		Pattern:          "no_domains",
		MatchType:        coredata.TrackerPatternMatchTypeExact,
		DisplayName:      "no_domains",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return pattern.Insert(ctx, tx, fx.scope)
	}))

	h := newMappingHandler(client)

	var got *catalogMatch

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var err error

		got, err = h.matchBySiblingOrigin(ctx, tx, pattern, nil)

		return err
	}))

	assert.Nil(t, got, "nil domains should immediately return nil")
}

func TestMatchBySiblingOrigin_ConvergentSiblings(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedPromotionFixture(t, ctx, client)

	now := time.Now().UTC().Truncate(time.Microsecond)

	orgThirdParty := coredata.ThirdParty{
		ID:                 gid.New(fx.scope.GetTenantID(), coredata.ThirdPartyEntityType),
		OrganizationID:     fx.organizationID,
		CommonThirdPartyID: &fx.commonThirdPartyID,
		Name:               "Google",
		Category:           coredata.ThirdPartyCategoryAnalytics,
		Certifications:     []string{},
		Countries:          coredata.CountryCodes{},
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	siblingA := coredata.TrackerPattern{
		ID:               gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType),
		OrganizationID:   fx.organizationID,
		CookieBannerID:   fx.banner.ID,
		CookieCategoryID: fx.normalCategoryID,
		ThirdPartyID:     &orgThirdParty.ID,
		TrackerType:      coredata.TrackerTypeCookie,
		Pattern:          "converge_a",
		MatchType:        coredata.TrackerPatternMatchTypeExact,
		DisplayName:      "converge_a",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	siblingB := coredata.TrackerPattern{
		ID:               gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType),
		OrganizationID:   fx.organizationID,
		CookieBannerID:   fx.banner.ID,
		CookieCategoryID: fx.normalCategoryID,
		ThirdPartyID:     &orgThirdParty.ID,
		TrackerType:      coredata.TrackerTypeCookie,
		Pattern:          "converge_b",
		MatchType:        coredata.TrackerPatternMatchTypeExact,
		DisplayName:      "converge_b",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	unmappedPattern := coredata.TrackerPattern{
		ID:               gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType),
		OrganizationID:   fx.organizationID,
		CookieBannerID:   fx.banner.ID,
		CookieCategoryID: fx.normalCategoryID,
		TrackerType:      coredata.TrackerTypeCookie,
		Pattern:          "converge_target",
		MatchType:        coredata.TrackerPatternMatchTypeExact,
		DisplayName:      "converge_target",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	sharedDomain := "google.com"

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := orgThirdParty.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		if err := siblingA.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		if err := siblingB.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		if err := unmappedPattern.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		detA := coredata.DetectedTracker{
			ID:               gid.New(fx.scope.GetTenantID(), coredata.DetectedTrackerEntityType),
			CookieBannerID:   fx.banner.ID,
			TrackerPatternID: &siblingA.ID,
			TrackerType:      coredata.TrackerTypeCookie,
			Identifier:       "converge_a",
			InitiatorDomain:  &sharedDomain,
			LastDetectedAt:   now,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if _, err := detA.Upsert(ctx, tx, fx.scope); err != nil {
			return err
		}

		detB := coredata.DetectedTracker{
			ID:               gid.New(fx.scope.GetTenantID(), coredata.DetectedTrackerEntityType),
			CookieBannerID:   fx.banner.ID,
			TrackerPatternID: &siblingB.ID,
			TrackerType:      coredata.TrackerTypeCookie,
			Identifier:       "converge_b",
			InitiatorDomain:  &sharedDomain,
			LastDetectedAt:   now,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if _, err := detB.Upsert(ctx, tx, fx.scope); err != nil {
			return err
		}

		detUnmapped := coredata.DetectedTracker{
			ID:               gid.New(fx.scope.GetTenantID(), coredata.DetectedTrackerEntityType),
			CookieBannerID:   fx.banner.ID,
			TrackerPatternID: &unmappedPattern.ID,
			TrackerType:      coredata.TrackerTypeCookie,
			Identifier:       "converge_target",
			InitiatorDomain:  &sharedDomain,
			LastDetectedAt:   now,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if _, err := detUnmapped.Upsert(ctx, tx, fx.scope); err != nil {
			return err
		}

		return nil
	}))

	h := newMappingHandler(client)

	var got *catalogMatch

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var err error

		got, err = h.matchBySiblingOrigin(ctx, tx, unmappedPattern, []string{"google.com"})

		return err
	}))

	require.NotNil(t, got, "multiple siblings converging to same third party should succeed")
	require.NotNil(t, got.commonPatternID)

	var commonPattern coredata.CommonTrackerPattern

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return commonPattern.LoadByID(ctx, conn, *got.commonPatternID)
	}))

	require.NotNil(t, commonPattern.CommonThirdPartyID)
	assert.Equal(t, fx.commonThirdPartyID, *commonPattern.CommonThirdPartyID)
}

func TestPromoteThirdParty_ExactCommonLinkIgnoresSimilarUnlinked(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedPromotionFixture(t, ctx, client)

	now := time.Now().UTC().Truncate(time.Microsecond)

	manualEntry := coredata.ThirdParty{
		ID:             gid.New(fx.scope.GetTenantID(), coredata.ThirdPartyEntityType),
		OrganizationID: fx.organizationID,
		Name:           "Google LLC",
		Category:       coredata.ThirdPartyCategoryAnalytics,
		Certifications: []string{},
		Countries:      coredata.CountryCodes{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	linked := coredata.ThirdParty{
		ID:                 gid.New(fx.scope.GetTenantID(), coredata.ThirdPartyEntityType),
		OrganizationID:     fx.organizationID,
		CommonThirdPartyID: &fx.commonThirdPartyID,
		Name:               "Google",
		Category:           coredata.ThirdPartyCategoryAnalytics,
		Certifications:     []string{},
		Countries:          coredata.CountryCodes{},
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := manualEntry.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		return linked.Insert(ctx, tx, fx.scope)
	}))

	got := promote(t, ctx, newMappingHandler(client), fx.trackerPattern, fx.commonThirdPartyID)

	require.NotNil(t, got)
	assert.Equal(t, linked.ID, *got, "exact-link path must short-circuit before the heuristic fires")
}

// TestProcess_BackfillsCommonThirdPartyFromSibling asserts that a pattern
// linked to an unlinked catalog row (no common_third_party_id) gets its
// catalog row backfilled from a sibling signal, and is promoted directly
// to the sibling's existing org ThirdParty.
func TestProcess_BackfillsCommonThirdPartyFromSibling(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedPromotionFixture(t, ctx, client)

	now := time.Now().UTC().Truncate(time.Microsecond)

	orgThirdParty := coredata.ThirdParty{
		ID:                 gid.New(fx.scope.GetTenantID(), coredata.ThirdPartyEntityType),
		OrganizationID:     fx.organizationID,
		CommonThirdPartyID: &fx.commonThirdPartyID,
		Name:               "Google LLC",
		Category:           coredata.ThirdPartyCategoryAnalytics,
		Certifications:     []string{},
		Countries:          coredata.CountryCodes{},
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	siblingPattern := coredata.TrackerPattern{
		ID:                     gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType),
		OrganizationID:         fx.organizationID,
		CookieBannerID:         fx.banner.ID,
		CookieCategoryID:       fx.normalCategoryID,
		CommonTrackerPatternID: &fx.commonPatternID,
		ThirdPartyID:           &orgThirdParty.ID,
		TrackerType:            coredata.TrackerTypeCookie,
		Pattern:                "_gid_backfill",
		MatchType:              coredata.TrackerPatternMatchTypeExact,
		DisplayName:            "_gid_backfill",
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	unlinkedCommon := coredata.CommonTrackerPattern{
		ID:          gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
		TrackerType: coredata.TrackerTypeCookie,
		Pattern:     "_ga_backfill",
		MatchType:   coredata.TrackerPatternMatchTypeExact,
		Confidence:  0.5,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	target := coredata.TrackerPattern{
		ID:                     gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType),
		OrganizationID:         fx.organizationID,
		CookieBannerID:         fx.banner.ID,
		CookieCategoryID:       fx.normalCategoryID,
		CommonTrackerPatternID: &unlinkedCommon.ID,
		TrackerType:            coredata.TrackerTypeCookie,
		Pattern:                "_ga_backfill",
		MatchType:              coredata.TrackerPatternMatchTypeExact,
		DisplayName:            "_ga_backfill",
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	// A vendor-specific initiator domain: resolveDeterministic strips
	// shared infrastructure (e.g. googletagmanager.com) before sibling
	// grouping, so the backfill must be driven by a real vendor domain.
	initiatorDomain := "google-analytics.com"

	siblingDetected := coredata.DetectedTracker{
		ID:               gid.New(fx.scope.GetTenantID(), coredata.DetectedTrackerEntityType),
		CookieBannerID:   fx.banner.ID,
		TrackerPatternID: &siblingPattern.ID,
		TrackerType:      coredata.TrackerTypeCookie,
		Identifier:       "_gid_backfill",
		InitiatorDomain:  &initiatorDomain,
		LastDetectedAt:   now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	targetDetected := coredata.DetectedTracker{
		ID:               gid.New(fx.scope.GetTenantID(), coredata.DetectedTrackerEntityType),
		CookieBannerID:   fx.banner.ID,
		TrackerPatternID: &target.ID,
		TrackerType:      coredata.TrackerTypeCookie,
		Identifier:       "_ga_backfill",
		InitiatorDomain:  &initiatorDomain,
		LastDetectedAt:   now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if _, err := unlinkedCommon.Upsert(ctx, tx); err != nil {
			return err
		}

		if err := orgThirdParty.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		if err := siblingPattern.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		if err := target.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		if _, err := siblingDetected.Upsert(ctx, tx, fx.scope); err != nil {
			return err
		}

		if _, err := targetDetected.Upsert(ctx, tx, fx.scope); err != nil {
			return err
		}

		return nil
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			_, _ = tx.Exec(ctx, `DELETE FROM common_tracker_patterns WHERE id = $1`, unlinkedCommon.ID)

			return nil
		})
	})

	h := newMappingHandler(client)
	require.NoError(t, h.Process(ctx, target))

	var reloadedCommon coredata.CommonTrackerPattern

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloadedCommon.LoadByID(ctx, conn, unlinkedCommon.ID)
	}))

	require.NotNil(t, reloadedCommon.CommonThirdPartyID, "the unlinked catalog row must be backfilled")
	assert.Equal(t, fx.commonThirdPartyID, *reloadedCommon.CommonThirdPartyID)

	var reloadedTarget coredata.TrackerPattern

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloadedTarget.LoadByID(ctx, conn, fx.scope, target.ID)
	}))

	require.NotNil(t, reloadedTarget.ThirdPartyID, "target must be promoted to the sibling's org third party")
	assert.Equal(t, orgThirdParty.ID, *reloadedTarget.ThirdPartyID)
	require.NotNil(t, reloadedTarget.CommonTrackerPatternID)
	assert.Equal(t, unlinkedCommon.ID, *reloadedTarget.CommonTrackerPatternID, "the existing catalog link must be preserved")
}

// TestProcess_UncategorisedLinksExistingThirdParty asserts that an
// uncategorised pattern is still linked to an already-existing matching
// org ThirdParty (linking to an existing party is ungated); only the
// creation of a new party stays gated, as covered by
// TestProcess_UncategorisedPatternIsNotPromoted.
func TestProcess_UncategorisedLinksExistingThirdParty(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedPromotionFixture(t, ctx, client)

	now := time.Now().UTC().Truncate(time.Microsecond)
	existing := coredata.ThirdParty{
		ID:                 gid.New(fx.scope.GetTenantID(), coredata.ThirdPartyEntityType),
		OrganizationID:     fx.organizationID,
		CommonThirdPartyID: &fx.commonThirdPartyID,
		Name:               "Google LLC",
		Category:           coredata.ThirdPartyCategoryAnalytics,
		Certifications:     []string{},
		Countries:          coredata.CountryCodes{},
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := existing.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		_, err := tx.Exec(
			ctx,
			`UPDATE tracker_patterns
			   SET cookie_category_id = $1,
			       mapping_requested_at = $2
			 WHERE id = $3`,
			fx.uncategorisedID,
			now,
			fx.trackerPattern.ID,
		)

		return err
	}))

	var reloadedBefore coredata.TrackerPattern

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloadedBefore.LoadByID(ctx, conn, fx.scope, fx.trackerPattern.ID)
	}))

	h := newMappingHandler(client)
	require.NoError(t, h.Process(ctx, reloadedBefore))

	var reloaded coredata.TrackerPattern

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloaded.LoadByID(ctx, conn, fx.scope, fx.trackerPattern.ID)
	}))

	require.NotNil(t, reloaded.ThirdPartyID, "uncategorised pattern must still link to an existing org third party")
	assert.Equal(t, existing.ID, *reloaded.ThirdPartyID)
}

// TestProcess_SiblingPromotionOnFirstPartyOrigin asserts that a pattern
// detected on the banner's own (first-party) origin is still grouped with
// its siblings sharing that origin. Sibling matching is an org-local
// co-occurrence signal and must not be defeated by the first-party domain
// filter that only protects the global catalog (domain) match.
func TestProcess_SiblingPromotionOnFirstPartyOrigin(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedPromotionFixture(t, ctx, client)

	now := time.Now().UTC().Truncate(time.Microsecond)

	orgThirdParty := coredata.ThirdParty{
		ID:                 gid.New(fx.scope.GetTenantID(), coredata.ThirdPartyEntityType),
		OrganizationID:     fx.organizationID,
		CommonThirdPartyID: &fx.commonThirdPartyID,
		Name:               "Google LLC",
		Category:           coredata.ThirdPartyCategoryAnalytics,
		Certifications:     []string{},
		Countries:          coredata.CountryCodes{},
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	siblingPattern := coredata.TrackerPattern{
		ID:                     gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType),
		OrganizationID:         fx.organizationID,
		CookieBannerID:         fx.banner.ID,
		CookieCategoryID:       fx.normalCategoryID,
		CommonTrackerPatternID: &fx.commonPatternID,
		ThirdPartyID:           &orgThirdParty.ID,
		TrackerType:            coredata.TrackerTypeCookie,
		Pattern:                "_sibling_fp",
		MatchType:              coredata.TrackerPatternMatchTypeExact,
		DisplayName:            "_sibling_fp",
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	unlinkedCommon := coredata.CommonTrackerPattern{
		ID:          gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
		TrackerType: coredata.TrackerTypeCookie,
		Pattern:     "__support__",
		MatchType:   coredata.TrackerPatternMatchTypeExact,
		Confidence:  0.5,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	target := coredata.TrackerPattern{
		ID:                     gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType),
		OrganizationID:         fx.organizationID,
		CookieBannerID:         fx.banner.ID,
		CookieCategoryID:       fx.normalCategoryID,
		CommonTrackerPatternID: &unlinkedCommon.ID,
		TrackerType:            coredata.TrackerTypeCookie,
		Pattern:                "__support__",
		MatchType:              coredata.TrackerPatternMatchTypeExact,
		DisplayName:            "__support__",
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	// The banner origin in seedWorkerFixture is an *.example.com host, so
	// its eTLD+1 (the first-party domain) is "example.com". Detecting both
	// patterns on that domain means uri.FilterFirstPartyDomains would strip
	// it — the regression this test guards against.
	firstPartyDomain := "example.com"

	siblingDetected := coredata.DetectedTracker{
		ID:               gid.New(fx.scope.GetTenantID(), coredata.DetectedTrackerEntityType),
		CookieBannerID:   fx.banner.ID,
		TrackerPatternID: &siblingPattern.ID,
		TrackerType:      coredata.TrackerTypeCookie,
		Identifier:       "_sibling_fp",
		InitiatorDomain:  &firstPartyDomain,
		LastDetectedAt:   now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	targetDetected := coredata.DetectedTracker{
		ID:               gid.New(fx.scope.GetTenantID(), coredata.DetectedTrackerEntityType),
		CookieBannerID:   fx.banner.ID,
		TrackerPatternID: &target.ID,
		TrackerType:      coredata.TrackerTypeCookie,
		Identifier:       "__support__",
		InitiatorDomain:  &firstPartyDomain,
		LastDetectedAt:   now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if _, err := unlinkedCommon.Upsert(ctx, tx); err != nil {
			return err
		}

		if err := orgThirdParty.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		if err := siblingPattern.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		if err := target.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		if _, err := siblingDetected.Upsert(ctx, tx, fx.scope); err != nil {
			return err
		}

		if _, err := targetDetected.Upsert(ctx, tx, fx.scope); err != nil {
			return err
		}

		return nil
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			_, _ = tx.Exec(ctx, `DELETE FROM common_tracker_patterns WHERE id = $1`, unlinkedCommon.ID)

			return nil
		})
	})

	h := newMappingHandler(client)
	require.NoError(t, h.Process(ctx, target))

	var reloadedCommon coredata.CommonTrackerPattern

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloadedCommon.LoadByID(ctx, conn, unlinkedCommon.ID)
	}))

	require.NotNil(t, reloadedCommon.CommonThirdPartyID, "catalog row must be backfilled from the first-party sibling")
	assert.Equal(t, fx.commonThirdPartyID, *reloadedCommon.CommonThirdPartyID)

	var reloadedTarget coredata.TrackerPattern

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloadedTarget.LoadByID(ctx, conn, fx.scope, target.ID)
	}))

	require.NotNil(t, reloadedTarget.ThirdPartyID, "target sharing a first-party origin must be promoted via its sibling")
	assert.Equal(t, orgThirdParty.ID, *reloadedTarget.ThirdPartyID)
}

// TestProcess_ReenqueuesUnmappedSiblingOnResolve asserts that when a
// pattern newly resolves a catalog third party, same-banner siblings
// that share an initiator domain but are still unpromoted get their
// mapping re-armed (backward propagation), while the already-promoted
// sibling that supplied the resolution is left untouched.
func TestProcess_ReenqueuesUnmappedSiblingOnResolve(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedPromotionFixture(t, ctx, client)

	now := time.Now().UTC().Truncate(time.Microsecond)

	orgThirdParty := coredata.ThirdParty{
		ID:                 gid.New(fx.scope.GetTenantID(), coredata.ThirdPartyEntityType),
		OrganizationID:     fx.organizationID,
		CommonThirdPartyID: &fx.commonThirdPartyID,
		Name:               "Google LLC",
		Category:           coredata.ThirdPartyCategoryAnalytics,
		Certifications:     []string{},
		Countries:          coredata.CountryCodes{},
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	mappedSibling := coredata.TrackerPattern{
		ID:                     gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType),
		OrganizationID:         fx.organizationID,
		CookieBannerID:         fx.banner.ID,
		CookieCategoryID:       fx.normalCategoryID,
		CommonTrackerPatternID: &fx.commonPatternID,
		ThirdPartyID:           &orgThirdParty.ID,
		TrackerType:            coredata.TrackerTypeCookie,
		Pattern:                "_gid_reenq",
		MatchType:              coredata.TrackerPatternMatchTypeExact,
		DisplayName:            "_gid_reenq",
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	target := coredata.TrackerPattern{
		ID:               gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType),
		OrganizationID:   fx.organizationID,
		CookieBannerID:   fx.banner.ID,
		CookieCategoryID: fx.normalCategoryID,
		TrackerType:      coredata.TrackerTypeCookie,
		Pattern:          "_ga_reenq_target",
		MatchType:        coredata.TrackerPatternMatchTypeExact,
		DisplayName:      "_ga_reenq_target",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	unmappedSibling := coredata.TrackerPattern{
		ID:               gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType),
		OrganizationID:   fx.organizationID,
		CookieBannerID:   fx.banner.ID,
		CookieCategoryID: fx.normalCategoryID,
		TrackerType:      coredata.TrackerTypeCookie,
		Pattern:          "_unmapped_reenq",
		MatchType:        coredata.TrackerPatternMatchTypeExact,
		DisplayName:      "_unmapped_reenq",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	sharedDomain := "reenq-tracker.com"

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := orgThirdParty.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		for _, p := range []coredata.TrackerPattern{mappedSibling, target, unmappedSibling} {
			if err := p.Insert(ctx, tx, fx.scope); err != nil {
				return err
			}
		}

		for id, identifier := range map[gid.GID]string{
			mappedSibling.ID:   "_gid_reenq",
			target.ID:          "_ga_reenq_target",
			unmappedSibling.ID: "_unmapped_reenq",
		} {
			patternID := id
			det := coredata.DetectedTracker{
				ID:               gid.New(fx.scope.GetTenantID(), coredata.DetectedTrackerEntityType),
				CookieBannerID:   fx.banner.ID,
				TrackerPatternID: &patternID,
				TrackerType:      coredata.TrackerTypeCookie,
				Identifier:       identifier,
				InitiatorDomain:  &sharedDomain,
				LastDetectedAt:   now,
				CreatedAt:        now,
				UpdatedAt:        now,
			}

			if _, err := det.Upsert(ctx, tx, fx.scope); err != nil {
				return err
			}
		}

		return nil
	}))

	h := newMappingHandler(client)
	require.NoError(t, h.Process(ctx, target))

	var reloadedTarget coredata.TrackerPattern

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloadedTarget.LoadByID(ctx, conn, fx.scope, target.ID)
	}))

	require.NotNil(t, reloadedTarget.ThirdPartyID, "target must resolve via its promoted sibling")

	var reloadedUnmapped coredata.TrackerPattern

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloadedUnmapped.LoadByID(ctx, conn, fx.scope, unmappedSibling.ID)
	}))

	require.NotNil(t, reloadedUnmapped.MappingRequestedAt, "unmapped sibling sharing the origin must be re-enqueued")

	var reloadedMapped coredata.TrackerPattern

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloadedMapped.LoadByID(ctx, conn, fx.scope, mappedSibling.ID)
	}))

	assert.Nil(t, reloadedMapped.MappingRequestedAt, "already-promoted sibling must not be re-enqueued")
}

// TestProcess_DoesNotReenqueuePromotedOrExtensionSiblings asserts that
// the re-enqueue skips siblings that are already promoted or
// EXTENSION-sourced, while still re-arming a plain unmapped sibling.
func TestProcess_DoesNotReenqueuePromotedOrExtensionSiblings(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedPromotionFixture(t, ctx, client)

	now := time.Now().UTC().Truncate(time.Microsecond)

	orgThirdParty := coredata.ThirdParty{
		ID:                 gid.New(fx.scope.GetTenantID(), coredata.ThirdPartyEntityType),
		OrganizationID:     fx.organizationID,
		CommonThirdPartyID: &fx.commonThirdPartyID,
		Name:               "Google LLC",
		Category:           coredata.ThirdPartyCategoryAnalytics,
		Certifications:     []string{},
		Countries:          coredata.CountryCodes{},
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	mappedSibling := coredata.TrackerPattern{
		ID:                     gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType),
		OrganizationID:         fx.organizationID,
		CookieBannerID:         fx.banner.ID,
		CookieCategoryID:       fx.normalCategoryID,
		CommonTrackerPatternID: &fx.commonPatternID,
		ThirdPartyID:           &orgThirdParty.ID,
		TrackerType:            coredata.TrackerTypeCookie,
		Pattern:                "_gid_guard",
		MatchType:              coredata.TrackerPatternMatchTypeExact,
		DisplayName:            "_gid_guard",
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	target := coredata.TrackerPattern{
		ID:               gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType),
		OrganizationID:   fx.organizationID,
		CookieBannerID:   fx.banner.ID,
		CookieCategoryID: fx.normalCategoryID,
		TrackerType:      coredata.TrackerTypeCookie,
		Pattern:          "_ga_guard_target",
		MatchType:        coredata.TrackerPatternMatchTypeExact,
		DisplayName:      "_ga_guard_target",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	plainSibling := coredata.TrackerPattern{
		ID:               gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType),
		OrganizationID:   fx.organizationID,
		CookieBannerID:   fx.banner.ID,
		CookieCategoryID: fx.normalCategoryID,
		TrackerType:      coredata.TrackerTypeCookie,
		Pattern:          "_plain_guard",
		MatchType:        coredata.TrackerPatternMatchTypeExact,
		DisplayName:      "_plain_guard",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	extensionSource := coredata.CookieSourceExtension
	extensionSibling := coredata.TrackerPattern{
		ID:               gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType),
		OrganizationID:   fx.organizationID,
		CookieBannerID:   fx.banner.ID,
		CookieCategoryID: fx.normalCategoryID,
		TrackerType:      coredata.TrackerTypeCookie,
		Pattern:          "_ext_guard",
		MatchType:        coredata.TrackerPatternMatchTypeExact,
		DisplayName:      "_ext_guard",
		Source:           &extensionSource,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	sharedDomain := "guard-tracker.com"

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := orgThirdParty.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		patterns := map[gid.GID]string{
			mappedSibling.ID:    "_gid_guard",
			target.ID:           "_ga_guard_target",
			plainSibling.ID:     "_plain_guard",
			extensionSibling.ID: "_ext_guard",
		}

		for _, p := range []coredata.TrackerPattern{mappedSibling, target, plainSibling, extensionSibling} {
			if err := p.Insert(ctx, tx, fx.scope); err != nil {
				return err
			}
		}

		for id, identifier := range patterns {
			patternID := id
			det := coredata.DetectedTracker{
				ID:               gid.New(fx.scope.GetTenantID(), coredata.DetectedTrackerEntityType),
				CookieBannerID:   fx.banner.ID,
				TrackerPatternID: &patternID,
				TrackerType:      coredata.TrackerTypeCookie,
				Identifier:       identifier,
				InitiatorDomain:  &sharedDomain,
				LastDetectedAt:   now,
				CreatedAt:        now,
				UpdatedAt:        now,
			}

			if _, err := det.Upsert(ctx, tx, fx.scope); err != nil {
				return err
			}
		}

		return nil
	}))

	h := newMappingHandler(client)
	require.NoError(t, h.Process(ctx, target))

	reload := func(id gid.GID) coredata.TrackerPattern {
		var p coredata.TrackerPattern

		require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
			return p.LoadByID(ctx, conn, fx.scope, id)
		}))

		return p
	}

	require.NotNil(t, reload(target.ID).ThirdPartyID, "target must resolve via its promoted sibling")
	require.NotNil(t, reload(plainSibling.ID).MappingRequestedAt, "plain unmapped sibling must be re-enqueued")
	assert.Nil(t, reload(mappedSibling.ID).MappingRequestedAt, "promoted sibling must not be re-enqueued")
	assert.Nil(t, reload(extensionSibling.ID).MappingRequestedAt, "EXTENSION-sourced sibling must not be re-enqueued")
}

// TestProcess_NoReenqueueWhenCommonThirdPartyPreexisted asserts that the
// re-trigger path, where the pattern's linked catalog row already carries
// a common third party, adds no new signal and therefore leaves unmapped
// siblings untouched (the cascade terminator).
func TestProcess_NoReenqueueWhenCommonThirdPartyPreexisted(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedPromotionFixture(t, ctx, client)

	now := time.Now().UTC().Truncate(time.Microsecond)

	unmappedSibling := coredata.TrackerPattern{
		ID:               gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType),
		OrganizationID:   fx.organizationID,
		CookieBannerID:   fx.banner.ID,
		CookieCategoryID: fx.normalCategoryID,
		TrackerType:      coredata.TrackerTypeCookie,
		Pattern:          "_unmapped_preexist",
		MatchType:        coredata.TrackerPatternMatchTypeExact,
		DisplayName:      "_unmapped_preexist",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	sharedDomain := "preexist-tracker.com"

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := unmappedSibling.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		for id, identifier := range map[gid.GID]string{
			fx.trackerPattern.ID: fx.trackerPattern.Pattern,
			unmappedSibling.ID:   "_unmapped_preexist",
		} {
			patternID := id
			det := coredata.DetectedTracker{
				ID:               gid.New(fx.scope.GetTenantID(), coredata.DetectedTrackerEntityType),
				CookieBannerID:   fx.banner.ID,
				TrackerPatternID: &patternID,
				TrackerType:      coredata.TrackerTypeCookie,
				Identifier:       identifier,
				InitiatorDomain:  &sharedDomain,
				LastDetectedAt:   now,
				CreatedAt:        now,
				UpdatedAt:        now,
			}

			if _, err := det.Upsert(ctx, tx, fx.scope); err != nil {
				return err
			}
		}

		return fx.trackerPattern.SetMappingRequested(ctx, tx)
	}))

	h := newMappingHandler(client)
	require.NoError(t, h.Process(ctx, fx.trackerPattern))

	var reloadedUnmapped coredata.TrackerPattern

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloadedUnmapped.LoadByID(ctx, conn, fx.scope, unmappedSibling.ID)
	}))

	assert.Nil(t, reloadedUnmapped.MappingRequestedAt, "re-trigger with a pre-existing common third party must not re-enqueue siblings")
}
