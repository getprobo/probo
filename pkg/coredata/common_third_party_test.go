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
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// TestCommonThirdPartyUpsert_SyncsReceiverToWrittenRow pins the invariant
// the common-third-parties seed relies on: after Upsert the receiver
// identifies the row that was actually written, so the seed can attach an
// entry's domains using party.ID without reloading.
//
// The seed used to reload by name on the update path. lower(name) is not
// unique, so that reload could return a different row and misattribute the
// entry's domains to it. Upsert conflicts on slug and returns the
// conflicting row, which is what makes the reload unnecessary.
func TestCommonThirdPartyUpsert_SyncsReceiverToWrittenRow(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	existing := seedCommonThirdParty(t, ctx, client)

	// A second row sharing the name but not the slug: the state that made
	// the old reload-by-name ambiguous.
	decoy := coredata.CommonThirdParty{
		ID:             gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType),
		Name:           existing.Name,
		Slug:           existing.Slug + "-decoy",
		Category:       coredata.ThirdPartyCategoryAnalytics,
		Certifications: []string{},
		CreatedAt:      existing.CreatedAt,
		UpdatedAt:      existing.UpdatedAt,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return decoy.Insert(ctx, tx)
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			_, err := tx.Exec(ctx, `DELETE FROM common_third_parties WHERE id = $1`, decoy.ID)
			return err
		})
	})

	// Re-seeding mints a fresh GID and conflicts on the existing slug,
	// exactly as the seed command does on a second run.
	now := time.Now().UTC().Truncate(time.Microsecond)
	reseeded := coredata.CommonThirdParty{
		ID:             gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType),
		Name:           existing.Name,
		Slug:           existing.Slug,
		Category:       coredata.ThirdPartyCategoryMarketing,
		Certifications: []string{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	var inserted bool

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var err error

		inserted, err = reseeded.Upsert(ctx, tx)

		return err
	}))

	assert.False(t, inserted, "conflicting slug must report an update, not an insert")
	assert.Equal(t, existing.ID, reseeded.ID, "receiver must carry the id of the row written")
	assert.NotEqual(t, decoy.ID, reseeded.ID, "receiver must not pick up the same-named decoy row")
	assert.Equal(t, existing.Slug, reseeded.Slug)
}

// TestLoadAllUnreferencedIDs_DomainsDoNotProtectAnEntry pins that an owned
// domain does not shield a catalog entry from pruning.
//
// A domain is part of the entry's own record, not something referencing it: it
// is enrichment output that means nothing once the entry is gone, and the
// foreign key cascades it away with the row. Counting it as a reference
// stranded exactly the entries most worth deleting, because an enriched entry
// almost always has one.
func TestLoadAllUnreferencedIDs_DomainsDoNotProtectAnEntry(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	withDomain := seedCommonThirdParty(t, ctx, client)

	domain := coredata.CommonThirdPartyDomain{
		ID:                 gid.New(gid.NilTenant, coredata.CommonThirdPartyDomainEntityType),
		CommonThirdPartyID: withDomain.ID,
		Domain:             "owned-" + withDomain.Slug + ".example",
		CreatedAt:          withDomain.CreatedAt,
		UpdatedAt:          withDomain.CreatedAt,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return domain.Insert(ctx, tx)
	}))

	// Back-date past the in-flight window so the age guard does not hide it.
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		_, err := tx.Exec(
			ctx,
			`UPDATE common_third_parties SET created_at = NOW() - interval '48 hours' WHERE id = $1`,
			withDomain.ID,
		)

		return err
	}))

	var ids []gid.GID

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		var parties coredata.CommonThirdParties

		var err error

		ids, err = parties.LoadAllUnreferencedIDs(ctx, conn, time.Now().Add(-time.Hour), false)

		return err
	}))

	assert.Contains(t, ids, withDomain.ID, "a domain must not protect an otherwise unreferenced entry")
}

// TestDeleteIfUnreferenced_RefusesOnceReferenced pins the guard that closes the
// prune race. Candidate selection and deletion are separate statements, so an
// entry can gain a reference in between; deleting it anyway would clear that
// brand-new link through ON DELETE SET NULL.
func TestDeleteIfUnreferenced_RefusesOnceReferenced(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	t.Run("deletes while unreferenced", func(t *testing.T) {
		t.Parallel()

		party := seedCommonThirdParty(t, ctx, client)

		var gone bool

		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			var err error

			gone, err = coredata.CommonThirdParty{}.DeleteIfUnreferenced(ctx, tx, party.ID, false)

			return err
		}))

		assert.True(t, gone)
	})

	t.Run("refuses once a pattern references it", func(t *testing.T) {
		t.Parallel()

		party := seedCommonThirdParty(t, ctx, client)

		// Stands in for a reference created after the candidate selection.
		now := time.Now().UTC().Truncate(time.Microsecond)
		insertCommonTrackerPattern(t, ctx, client, coredata.CommonTrackerPattern{
			ID:                 gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
			CommonThirdPartyID: &party.ID,
			TrackerType:        coredata.TrackerTypeCookie,
			Pattern:            "race_" + party.Slug,
			MatchType:          coredata.TrackerPatternMatchTypeExact,
			Confidence:         0.9,
			Attribution:        coredata.CommonTrackerPatternAttributionThirdParty,
			CreatedAt:          now,
			UpdatedAt:          now,
		})

		var gone bool

		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			var err error

			gone, err = coredata.CommonThirdParty{}.DeleteIfUnreferenced(ctx, tx, party.ID, false)

			return err
		}))

		assert.False(t, gone, "must not delete an entry that gained a reference")

		var still coredata.CommonThirdParty

		require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
			return still.LoadByID(ctx, conn, party.ID)
		}))
	})
}

func markCommonThirdPartyEnriched(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	id gid.GID,
) {
	t.Helper()

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		_, err := tx.Exec(
			ctx,
			`UPDATE common_third_parties SET enrichment = $2 WHERE id = $1`,
			id,
			json.RawMessage(`{"status":"done"}`),
		)

		return err
	}))
}

// TestDeleteIfUnreferenced_RefusesOnceEnriched pins the --unenriched-only
// race: selection and deletion are separate statements, so enrichment can
// complete in between. The delete must then refuse, matching the filter the
// operator asked for.
func TestDeleteIfUnreferenced_RefusesOnceEnriched(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	t.Run("deletes unenriched when unenrichedOnly", func(t *testing.T) {
		t.Parallel()

		party := seedCommonThirdParty(t, ctx, client)

		var gone bool

		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			var err error

			gone, err = coredata.CommonThirdParty{}.DeleteIfUnreferenced(ctx, tx, party.ID, true)

			return err
		}))

		assert.True(t, gone)
	})

	t.Run("refuses once enrichment is set when unenrichedOnly", func(t *testing.T) {
		t.Parallel()

		party := seedCommonThirdParty(t, ctx, client)
		markCommonThirdPartyEnriched(t, ctx, client, party.ID)

		var gone bool

		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			var err error

			gone, err = coredata.CommonThirdParty{}.DeleteIfUnreferenced(ctx, tx, party.ID, true)

			return err
		}))

		assert.False(t, gone, "must not delete an entry that completed enrichment")

		var still coredata.CommonThirdParty

		require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
			return still.LoadByID(ctx, conn, party.ID)
		}))
	})

	t.Run("deletes enriched when not unenrichedOnly", func(t *testing.T) {
		t.Parallel()

		party := seedCommonThirdParty(t, ctx, client)
		markCommonThirdPartyEnriched(t, ctx, client, party.ID)

		var gone bool

		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			var err error

			gone, err = coredata.CommonThirdParty{}.DeleteIfUnreferenced(ctx, tx, party.ID, false)

			return err
		}))

		assert.True(t, gone)
	})
}
