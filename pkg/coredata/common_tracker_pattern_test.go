// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

// seedCommonThirdParty inserts a global catalog third party with a
// collision-free name/slug. The catalog is not tenant-scoped and carries
// unique indexes, so parallel tests must namespace their rows.
func seedCommonThirdParty(t *testing.T, ctx context.Context, client *pg.Client) coredata.CommonThirdParty {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Microsecond)
	id := gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType)
	suffix := id.String()

	party := coredata.CommonThirdParty{
		ID:             id,
		Name:           "Acme " + suffix,
		Slug:           "acme-" + suffix,
		Category:       coredata.ThirdPartyCategoryAnalytics,
		Certifications: []string{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return party.Insert(ctx, tx)
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			_, err := tx.Exec(ctx, `DELETE FROM common_third_parties WHERE id = $1`, id)
			return err
		})
	})

	return party
}

// insertCommonTrackerPattern inserts a catalog pattern row verbatim
// (Insert, not Upsert) so a test can stage an exact enrichment state.
func insertCommonTrackerPattern(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	cp coredata.CommonTrackerPattern,
) {
	t.Helper()

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return cp.Insert(ctx, tx)
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			_, err := tx.Exec(ctx, `DELETE FROM common_tracker_patterns WHERE id = $1`, cp.ID)
			return err
		})
	})
}

func loadCommonTrackerPattern(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	id gid.GID,
) coredata.CommonTrackerPattern {
	t.Helper()

	var reloaded coredata.CommonTrackerPattern

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloaded.LoadByID(ctx, conn, id)
	}))

	return reloaded
}

// TestCommonTrackerPattern_SetEnriched_AllowsEmptyDescription pins the
// no-fabrication contract: the enrichment worker records an empty
// description when it cannot substantiate a purpose, and the row is
// still marked terminally enriched so the stale-recovery loop never
// re-queues it.
func TestCommonTrackerPattern_SetEnriched_AllowsEmptyDescription(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	requestedAt := now.Add(-time.Minute)
	cp := coredata.CommonTrackerPattern{
		ID:                    gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
		TrackerType:           coredata.TrackerTypeLocalStorage,
		Pattern:               "blank_key_" + gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType).String(),
		MatchType:             coredata.TrackerPatternMatchTypeExact,
		Description:           "",
		Confidence:            0.5,
		EnrichmentRequestedAt: &requestedAt,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	insertCommonTrackerPattern(t, ctx, client, cp)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return cp.SetEnriched(ctx, tx, "", nil)
	}))

	reloaded := loadCommonTrackerPattern(t, ctx, client, cp.ID)
	assert.Equal(t, "", reloaded.Description, "blank description must stay blank")
	assert.NotNil(t, reloaded.EnrichedAt, "blank row must be marked enriched (terminal-for-now)")
	assert.Nil(t, reloaded.EnrichmentRequestedAt, "enriched row must leave the queue")

	// A blank but enriched row must NOT be re-queued by stale recovery:
	// enriched_at is set, so the stale sweep skips it.
	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return coredata.ResetStaleEnrichments(ctx, conn, 0)
	}))

	afterSweep := loadCommonTrackerPattern(t, ctx, client, cp.ID)
	assert.Nil(t, afterSweep.EnrichmentRequestedAt, "stale recovery must not re-queue an enriched blank row")
}

// TestCommonTrackerPattern_SetEnriched_LinksThirdPartyWithoutOverride
// pins the link-no-override contract: the enrichment worker links a
// resolved third party only when the row has none, and never clobbers an
// attribution the mapping pipeline already resolved.
func TestCommonTrackerPattern_SetEnriched_LinksThirdPartyWithoutOverride(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	party := seedCommonThirdParty(t, ctx, client)
	other := seedCommonThirdParty(t, ctx, client)

	now := time.Now().UTC().Truncate(time.Microsecond)

	t.Run("links when unset", func(t *testing.T) {
		cp := coredata.CommonTrackerPattern{
			ID:          gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
			TrackerType: coredata.TrackerTypeCookie,
			Pattern:     "link_unset_" + gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType).String(),
			MatchType:   coredata.TrackerPatternMatchTypeExact,
			Confidence:  0.5,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		insertCommonTrackerPattern(t, ctx, client, cp)

		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			return cp.SetEnriched(ctx, tx, "Analytics tracker.", &party.ID)
		}))

		reloaded := loadCommonTrackerPattern(t, ctx, client, cp.ID)
		require.NotNil(t, reloaded.CommonThirdPartyID)
		assert.Equal(t, party.ID, *reloaded.CommonThirdPartyID, "unlinked row must gain the resolved third party")
	})

	t.Run("does not override existing link", func(t *testing.T) {
		cp := coredata.CommonTrackerPattern{
			ID:                 gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
			CommonThirdPartyID: &party.ID,
			TrackerType:        coredata.TrackerTypeCookie,
			Pattern:            "link_set_" + gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType).String(),
			MatchType:          coredata.TrackerPatternMatchTypeExact,
			Confidence:         0.5,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		insertCommonTrackerPattern(t, ctx, client, cp)

		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			return cp.SetEnriched(ctx, tx, "Analytics tracker.", &other.ID)
		}))

		reloaded := loadCommonTrackerPattern(t, ctx, client, cp.ID)
		require.NotNil(t, reloaded.CommonThirdPartyID)
		assert.Equal(t, party.ID, *reloaded.CommonThirdPartyID, "existing third party link must not be overridden")
	})
}

// TestCommonTrackerPattern_Upsert_RequeuesBlankRowOnThirdPartyLink pins
// the re-trigger contract: when a blank, unlinked catalog row later
// gains a third party through the mapping pipeline's Upsert, enrichment
// is re-armed so the now-known vendor gets a second, better-informed
// description attempt.
func TestCommonTrackerPattern_Upsert_RequeuesBlankRowOnThirdPartyLink(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	party := seedCommonThirdParty(t, ctx, client)

	now := time.Now().UTC().Truncate(time.Microsecond)
	enrichedAt := now.Add(-time.Hour)
	pattern := "requeue_" + gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType).String()

	// Stage a terminal blank row: enriched, no description, no vendor.
	blank := coredata.CommonTrackerPattern{
		ID:          gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
		TrackerType: coredata.TrackerTypeCookie,
		Pattern:     pattern,
		MatchType:   coredata.TrackerPatternMatchTypeExact,
		Description: "",
		Confidence:  0.5,
		EnrichedAt:  &enrichedAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	insertCommonTrackerPattern(t, ctx, client, blank)

	// The mapping pipeline upserts the same key now carrying a vendor.
	linking := coredata.CommonTrackerPattern{
		ID:                 gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
		CommonThirdPartyID: &party.ID,
		TrackerType:        coredata.TrackerTypeCookie,
		Pattern:            pattern,
		MatchType:          coredata.TrackerPatternMatchTypeExact,
		Description:        "",
		Confidence:         0.7,
		CreatedAt:          now,
		UpdatedAt:          now.Add(time.Minute),
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		inserted, err := linking.Upsert(ctx, tx)
		if err != nil {
			return err
		}

		assert.False(t, inserted, "Upsert must hit the existing blank row, not insert a new one")

		return nil
	}))

	reloaded := loadCommonTrackerPattern(t, ctx, client, blank.ID)
	require.NotNil(t, reloaded.CommonThirdPartyID)
	assert.Equal(t, party.ID, *reloaded.CommonThirdPartyID, "blank row must gain the linked third party")
	assert.NotNil(t, reloaded.EnrichmentRequestedAt, "linking a vendor must re-queue the blank row for enrichment")
	assert.Nil(t, reloaded.EnrichedAt, "re-queued row must no longer be terminal")
}

// TestCommonTrackerPattern_Upsert_KeepsDescribedRowTerminal pins the
// negative case: a row that already has a description is not re-queued
// when its third party changes, since it already carries a substantiated
// purpose.
func TestCommonTrackerPattern_Upsert_KeepsDescribedRowTerminal(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	party := seedCommonThirdParty(t, ctx, client)

	now := time.Now().UTC().Truncate(time.Microsecond)
	enrichedAt := now.Add(-time.Hour)
	pattern := "described_" + gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType).String()

	described := coredata.CommonTrackerPattern{
		ID:          gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
		TrackerType: coredata.TrackerTypeCookie,
		Pattern:     pattern,
		MatchType:   coredata.TrackerPatternMatchTypeExact,
		Description: "An established analytics cookie.",
		Confidence:  0.9,
		EnrichedAt:  &enrichedAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	insertCommonTrackerPattern(t, ctx, client, described)

	linking := coredata.CommonTrackerPattern{
		ID:                 gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
		CommonThirdPartyID: &party.ID,
		TrackerType:        coredata.TrackerTypeCookie,
		Pattern:            pattern,
		MatchType:          coredata.TrackerPatternMatchTypeExact,
		Description:        "",
		Confidence:         0.7,
		CreatedAt:          now,
		UpdatedAt:          now.Add(time.Minute),
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		_, err := linking.Upsert(ctx, tx)
		return err
	}))

	reloaded := loadCommonTrackerPattern(t, ctx, client, described.ID)
	assert.Equal(t, "An established analytics cookie.", reloaded.Description, "existing description must be preserved")
	assert.Nil(t, reloaded.EnrichmentRequestedAt, "described row must not be re-queued")
	assert.NotNil(t, reloaded.EnrichedAt, "described row must stay terminal")
}
