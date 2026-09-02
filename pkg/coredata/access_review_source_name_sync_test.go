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
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
)

// nameSyncEpoch backdates the fixtures far enough that they sort ahead of any
// row another test leaves behind. LoadNextUnsyncedNameForUpdateSkipLocked is
// deliberately cross-tenant and ordered by created_at, and only five of the
// packages sharing this database hold the global queue lock, so these tests
// have to win the ordering rather than clear the queue: deleting other
// tenants' claimable rows would pull fixtures out from under whichever
// package is running alongside.
var nameSyncEpoch = time.Date(1991, time.August, 6, 0, 0, 0, 0, time.UTC)

// seedNameSyncSource inserts an organization, a connector and one
// connector-backed access source with its name unresolved, which is the state
// the source-name worker claims from.
func seedNameSyncSource(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	createdAt time.Time,
) (coredata.Scoper, *coredata.AccessReviewSource) {
	t.Helper()

	scope, organizationID := seedConnectorOrg(t, ctx, client)

	var key cipher.EncryptionKey

	connectorID, err := insertConnector(ctx, client, scope, organizationID, coredata.ConnectorProviderMetabase, key)
	require.NoError(t, err)

	source := &coredata.AccessReviewSource{
		ID:             gid.New(scope.GetTenantID(), coredata.AccessReviewSourceEntityType),
		OrganizationID: organizationID,
		ConnectorID:    &connectorID,
		Name:           "Metabase",
		CreatedAt:      createdAt,
		UpdatedAt:      time.Now().UTC(),
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		inserted, err := source.Insert(ctx, tx, scope)
		if err != nil {
			return err
		}

		require.True(t, inserted)

		return nil
	}))

	// A backdated unsynced row left behind would be claimed ahead of every
	// later test's fixture, so it must not outlive this one.
	t.Cleanup(func() {
		err := client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			_, err := tx.Exec(ctx, `DELETE FROM access_review_sources WHERE id = $1`, source.ID)

			return err
		})
		assert.NoError(t, err, "cleanup: cannot delete seeded access review source %s", source.ID)
	})

	return scope, source
}

// TestRecordNameSyncAttemptHoldsSourceOutOfQueue pins the invariant the
// production incident turned on: a charged attempt must make the row
// invisible to the very next claim. Without it the worker re-claims the same
// failing source at process latency, because a failed Process transitions
// nothing.
func TestRecordNameSyncAttemptHoldsSourceOutOfQueue(t *testing.T) {
	ctx := context.Background()
	client := test.PGClient(t)

	scope, source := seedNameSyncSource(t, ctx, client, nameSyncEpoch)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var claimed coredata.AccessReviewSource

		require.NoError(t, claimed.LoadNextUnsyncedNameForUpdateSkipLocked(ctx, tx))
		require.Equal(t, source.ID, claimed.ID)
		assert.Equal(t, 0, claimed.NameSyncAttempts)

		return claimed.RecordNameSyncAttempt(ctx, tx, scope, time.Minute)
	}))

	// A second claim in a fresh transaction must not see it.
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var claimed coredata.AccessReviewSource

		err := claimed.LoadNextUnsyncedNameForUpdateSkipLocked(ctx, tx)
		if err == nil {
			assert.NotEqual(t, source.ID, claimed.ID, "a charged source must not be re-claimed before its backoff")

			return nil
		}

		require.ErrorIs(t, err, coredata.ErrNoAccessReviewSourceNameSyncAvailable)

		return nil
	}))

	var reloaded coredata.AccessReviewSource

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return reloaded.LoadByID(ctx, tx, scope, source.ID)
	}))

	assert.Equal(t, 1, reloaded.NameSyncAttempts)
	require.NotNil(t, reloaded.NameSyncNextAttemptAt)

	// The deadline is computed from the database clock, so it has to be read
	// against that same clock: comparing it to the host's would flake on any
	// skew between the two rather than on a regression.
	var dbNow time.Time

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return tx.QueryRow(ctx, `SELECT NOW()`).Scan(&dbNow)
	}))

	assert.WithinDuration(t, dbNow.Add(time.Minute), *reloaded.NameSyncNextAttemptAt, 30*time.Second)
}

// TestNameSyncClaimYieldsToOtherSources pins the head-of-line fix: a source
// inside its backoff must let a newer one through, and come back when the
// backoff expires.
func TestNameSyncClaimYieldsToOtherSources(t *testing.T) {
	ctx := context.Background()
	client := test.PGClient(t)

	olderScope, older := seedNameSyncSource(t, ctx, client, nameSyncEpoch)
	_, newer := seedNameSyncSource(t, ctx, client, nameSyncEpoch.Add(time.Hour))

	// Charge the older source, which is the one the failing provider owns.
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return older.RecordNameSyncAttempt(ctx, tx, olderScope, time.Hour)
	}))

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var claimed coredata.AccessReviewSource

		require.NoError(t, claimed.LoadNextUnsyncedNameForUpdateSkipLocked(ctx, tx))
		assert.Equal(t, newer.ID, claimed.ID, "a backing-off source must not block the queue")

		return nil
	}))

	// Once the backoff lapses it is claimable again, and oldest-first puts it
	// back at the head.
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return older.RecordNameSyncAttempt(ctx, tx, olderScope, -time.Minute)
	}))

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var claimed coredata.AccessReviewSource

		require.NoError(t, claimed.LoadNextUnsyncedNameForUpdateSkipLocked(ctx, tx))
		assert.Equal(t, older.ID, claimed.ID)

		return nil
	}))
}

// TestMarkNameSyncedYieldsToConcurrentReset pins that a reconnect landing
// during a resolve is not overwritten by the worker's own write, which would
// silently strip the fresh retry budget the reconnect just granted.
func TestMarkNameSyncedYieldsToConcurrentReset(t *testing.T) {
	ctx := context.Background()
	client := test.PGClient(t)

	scope, source := seedNameSyncSource(t, ctx, client, nameSyncEpoch)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return source.RecordNameSyncAttempt(ctx, tx, scope, time.Minute)
	}))

	// A reconnect resets the budget while the resolve is still in flight.
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		sources := coredata.AccessReviewSources{}

		return sources.ResetNameSyncByConnectorID(ctx, tx, scope, *source.ConnectorID)
	}))

	// The worker finishes and writes the name it resolved against the old
	// connector. It must not land.
	stale := *source
	stale.Name = "Metabase / Stale Instance"

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return stale.MarkNameSynced(ctx, tx, scope, time.Now().UTC())
	}))

	var reloaded coredata.AccessReviewSource

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return reloaded.LoadByID(ctx, tx, scope, source.ID)
	}))

	assert.Nil(t, reloaded.NameSyncedAt, "the reconnect must keep the source claimable")
	assert.Equal(t, 0, reloaded.NameSyncAttempts, "the fresh budget must survive")
	assert.Equal(t, "Metabase", reloaded.Name)
}

// TestResetNameSyncByConnectorIDReleasesBackedOffSource pins the widened
// predicate: a source still inside its backoff has name_synced_at NULL, so the
// old "IS NOT NULL" guard skipped exactly the rows a reconnect must free.
func TestResetNameSyncByConnectorIDReleasesBackedOffSource(t *testing.T) {
	ctx := context.Background()
	client := test.PGClient(t)

	scope, source := seedNameSyncSource(t, ctx, client, nameSyncEpoch)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return source.RecordNameSyncAttempt(ctx, tx, scope, time.Hour)
	}))

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		sources := coredata.AccessReviewSources{}

		return sources.ResetNameSyncByConnectorID(ctx, tx, scope, *source.ConnectorID)
	}))

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var claimed coredata.AccessReviewSource

		require.NoError(t, claimed.LoadNextUnsyncedNameForUpdateSkipLocked(ctx, tx))
		assert.Equal(t, source.ID, claimed.ID)
		assert.Equal(t, 0, claimed.NameSyncAttempts)

		return nil
	}))
}

// TestMarkNameSyncedYieldsToAReclaimAtTheSameAttemptCount pins the claim
// generation. A reconnect resets the budget to zero and the next claim charges
// it straight back to one, so the attempt count alone cannot tell the two
// claims apart: a resolve still in flight from before the reconnect would
// match the guard again and overwrite the fresh claim with a stale name.
func TestMarkNameSyncedYieldsToAReclaimAtTheSameAttemptCount(t *testing.T) {
	ctx := context.Background()
	client := test.PGClient(t)

	scope, source := seedNameSyncSource(t, ctx, client, nameSyncEpoch)

	// The in-flight worker claims and starts resolving.
	inFlight := *source

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return inFlight.RecordNameSyncAttempt(ctx, tx, scope, time.Minute)
	}))
	require.Equal(t, 1, inFlight.NameSyncAttempts)

	// A reconnect lands, then a second worker claims the freed row. Both
	// claims now sit at attempt 1.
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		sources := coredata.AccessReviewSources{}

		return sources.ResetNameSyncByConnectorID(ctx, tx, scope, *source.ConnectorID)
	}))

	reclaimed := *source

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return reclaimed.RecordNameSyncAttempt(ctx, tx, scope, time.Minute)
	}))
	require.Equal(t, inFlight.NameSyncAttempts, reclaimed.NameSyncAttempts, "the count alone cannot separate the claims")

	// The first worker finishes and writes the name it resolved against the
	// connector that has since been replaced.
	inFlight.Name = "Metabase / Stale Instance"

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return inFlight.MarkNameSynced(ctx, tx, scope, time.Now().UTC())
	}))

	var reloaded coredata.AccessReviewSource

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return reloaded.LoadByID(ctx, tx, scope, source.ID)
	}))

	assert.Nil(t, reloaded.NameSyncedAt, "the stale write must not retire the fresh claim")
	assert.Equal(t, "Metabase", reloaded.Name, "the stale name must not land")

	// The claim that actually owns the row still retires it.
	reclaimed.Name = "Metabase / Current Instance"

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return reclaimed.MarkNameSynced(ctx, tx, scope, time.Now().UTC())
	}))

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return reloaded.LoadByID(ctx, tx, scope, source.ID)
	}))

	assert.NotNil(t, reloaded.NameSyncedAt)
	assert.Equal(t, "Metabase / Current Instance", reloaded.Name)
}

// TestNameSyncBudgetIsSpentOverFiveAttempts walks the whole budget the way a
// failing provider does, so the schedule the worker relies on to stop is
// pinned end to end rather than only in the backoff helper.
func TestNameSyncBudgetIsSpentOverFiveAttempts(t *testing.T) {
	ctx := context.Background()
	client := test.PGClient(t)

	scope, source := seedNameSyncSource(t, ctx, client, nameSyncEpoch)

	// The schedule the worker charges: 1m, 2m, 4m, 8m, 16m.
	backoffs := []time.Duration{time.Minute, 2 * time.Minute, 4 * time.Minute, 8 * time.Minute, 16 * time.Minute}

	for attempt, backoff := range backoffs {
		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			var claimed coredata.AccessReviewSource

			// Each attempt starts from a claim, so the row must still be
			// visible to the queue at every step below the ceiling.
			if err := claimed.LoadNextUnsyncedNameForUpdateSkipLocked(ctx, tx); err != nil {
				return err
			}

			require.Equal(t, source.ID, claimed.ID)
			require.Equal(t, attempt, claimed.NameSyncAttempts)

			// Charge with the deadline already elapsed, so the next iteration
			// can claim without waiting out the real backoff.
			return claimed.RecordNameSyncAttempt(ctx, tx, scope, -backoff)
		}))
	}

	var reloaded coredata.AccessReviewSource

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return reloaded.LoadByID(ctx, tx, scope, source.ID)
	}))

	// Five charges, which is the point at which the worker keeps the generic
	// name instead of asking the provider again.
	assert.Equal(t, len(backoffs), reloaded.NameSyncAttempts)
	assert.Nil(t, reloaded.NameSyncedAt, "the budget alone must not retire the row; the worker does that")

	// Retiring it at the exhausted count is what takes it out of the queue.
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return reloaded.MarkNameSynced(ctx, tx, scope, time.Now().UTC())
	}))

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var claimed coredata.AccessReviewSource

		err := claimed.LoadNextUnsyncedNameForUpdateSkipLocked(ctx, tx)
		if err == nil {
			assert.NotEqual(t, source.ID, claimed.ID, "a retired source must leave the queue")

			return nil
		}

		require.ErrorIs(t, err, coredata.ErrNoAccessReviewSourceNameSyncAvailable)

		return nil
	}))
}
