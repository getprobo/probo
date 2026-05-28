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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

//go:fix inline
func ptr[T any](v T) *T { return new(v) }

// promotionFixture extends workerFixture with a CommonThirdParty and a
// CommonTrackerPattern linking the catalog to the test pattern. It is
// the minimum scaffolding promoteThirdParty needs to run end-to-end.
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

	commonThirdPartyID := gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType)
	commonThirdParty := coredata.CommonThirdParty{
		ID:             commonThirdPartyID,
		Name:           "Google",
		Slug:           "google",
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
		Pattern:            "_ga",
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
		Pattern:                "_ga",
		MatchType:              coredata.TrackerPatternMatchTypeExact,
		DisplayName:            "_ga",
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

// promote runs promoteThirdParty inside its own transaction so each
// test case starts from a clean state.
func promote(
	t *testing.T,
	ctx context.Context,
	h *trackerMappingHandler,
	client *pg.Client,
	tp coredata.TrackerPattern,
	commonPatternID gid.GID,
) *gid.GID {
	t.Helper()

	var got *gid.GID

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var err error

		got, err = h.promoteThirdParty(ctx, tx, tp, commonPatternID)

		return err
	}))

	return got
}

func TestPromoteThirdParty_ExactCommonLink(t *testing.T) {
	t.Parallel()

	client := newTestPgClient(t)
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

	got := promote(t, ctx, newMappingHandler(client), client, fx.trackerPattern, fx.commonPatternID)

	require.NotNil(t, got)
	assert.Equal(t, existing.ID, *got, "should return the existing org ThirdParty linked by common id")
}

func TestPromoteThirdParty_HeuristicMatch(t *testing.T) {
	t.Parallel()

	client := newTestPgClient(t)
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

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return manualEntry.Insert(ctx, tx, fx.scope)
	}))

	got := promote(t, ctx, newMappingHandler(client), client, fx.trackerPattern, fx.commonPatternID)

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

	client := newTestPgClient(t)
	ctx := context.Background()
	fx := seedPromotionFixture(t, ctx, client)

	got := promote(t, ctx, newMappingHandler(client), client, fx.trackerPattern, fx.commonPatternID)

	require.NotNil(t, got, "fallback should create a new ThirdParty")

	var reloaded coredata.ThirdParty

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloaded.LoadByID(ctx, conn, fx.scope, *got)
	}))

	assert.Equal(t, fx.organizationID, reloaded.OrganizationID)
	assert.Equal(t, "Google", reloaded.Name)
	require.NotNil(t, reloaded.CommonThirdPartyID)
	assert.Equal(t, fx.commonThirdPartyID, *reloaded.CommonThirdPartyID)
	assert.Equal(t, coredata.ThirdPartyCategoryAnalytics, reloaded.Category)
	assert.False(t, reloaded.FirstLevel)
	assert.False(t, reloaded.ShowOnTrustCenter)
}

func TestPromoteThirdParty_NoCommonThirdPartyOnPattern(t *testing.T) {
	t.Parallel()

	client := newTestPgClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)

	now := time.Now().UTC().Truncate(time.Microsecond)
	commonPattern := coredata.CommonTrackerPattern{
		ID:          gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
		TrackerType: coredata.TrackerTypeCookie,
		Pattern:     "unknown_xyz",
		MatchType:   coredata.TrackerPatternMatchTypeExact,
		Description: "",
		Confidence:  0.5,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	pattern := coredata.TrackerPattern{
		ID:                     gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType),
		OrganizationID:         fx.organizationID,
		CookieBannerID:         fx.banner.ID,
		CookieCategoryID:       fx.normalCategoryID,
		CommonTrackerPatternID: &commonPattern.ID,
		TrackerType:            coredata.TrackerTypeCookie,
		Pattern:                "unknown_xyz",
		MatchType:              coredata.TrackerPatternMatchTypeExact,
		DisplayName:            "unknown_xyz",
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if _, err := commonPattern.Upsert(ctx, tx); err != nil {
			return err
		}

		return pattern.Insert(ctx, tx, fx.scope)
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			_, err := tx.Exec(ctx, `DELETE FROM common_tracker_patterns WHERE id = $1`, commonPattern.ID)

			return err
		})
	})

	got := promote(t, ctx, newMappingHandler(client), client, pattern, commonPattern.ID)

	assert.Nil(t, got, "patterns whose catalog row has no CommonThirdPartyID should not be promoted")
}

// TestProcess_PreservesCatalogMappingOnReTrigger asserts that when
// Process is called for a pattern that already carries a
// common_tracker_pattern_id, the catalog pipeline is skipped and the
// existing catalog link is preserved verbatim.
func TestProcess_PreservesCatalogMappingOnReTrigger(t *testing.T) {
	t.Parallel()

	client := newTestPgClient(t)
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

	client := newTestPgClient(t)
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

	client := newTestPgClient(t)
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

	client := newTestPgClient(t)
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

func TestPromoteThirdParty_ExactCommonLinkIgnoresSimilarUnlinked(t *testing.T) {
	t.Parallel()

	client := newTestPgClient(t)
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

	got := promote(t, ctx, newMappingHandler(client), client, fx.trackerPattern, fx.commonPatternID)

	require.NotNil(t, got)
	assert.Equal(t, linked.ID, *got, "exact-link path must short-circuit before the heuristic fires")
}
