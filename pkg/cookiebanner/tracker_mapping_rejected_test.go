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

package cookiebanner

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

// insertReviewedParty writes a catalog row in a given review state, so the
// mapping worker's handling of a rejection can be exercised without going
// through the whole enrichment and review path.
func insertReviewedParty(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	review coredata.CommonThirdPartyReview,
	verdict *coredata.CommonTrackerPatternAttribution,
) gid.GID {
	t.Helper()

	now := time.Now()
	id := gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType)
	party := coredata.CommonThirdParty{
		ID:             id,
		Name:           "Reviewed " + id.String(),
		Slug:           "reviewed-" + id.String(),
		Category:       coredata.ThirdPartyCategoryOther,
		Certifications: []string{},
		Review:         review,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if verdict != nil {
		party.RejectedVerdict = verdict
		party.ReviewedAt = &now
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

	return id
}

// verdictFor runs rejectedVerdictFor in a transaction, which it now requires
// so the review is read where it is acted on.
func verdictFor(
	t *testing.T,
	ctx context.Context,
	h *trackerMappingHandler,
	id gid.GID,
) (*coredata.CommonTrackerPatternAttribution, bool, error) {
	t.Helper()

	var (
		verdict *coredata.CommonTrackerPatternAttribution
		gone    bool
		inner   error
	)

	if err := h.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		verdict, gone, inner = h.rejectedVerdictFor(ctx, tx, id)
		return nil
	}); err != nil {
		return nil, false, err
	}

	return verdict, gone, inner
}

// A rejected catalog row must divert the pattern to the verdict the review
// recorded rather than link the vendor whose name matched. Both terminal
// verdicts must survive the round trip, since which one applies is the whole
// reason the verdict is stored on the row.
func TestRejectedVerdictFor(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	h := newMappingHandler(client)

	firstParty := coredata.CommonTrackerPatternAttributionFirstParty
	notAttributable := coredata.CommonTrackerPatternAttributionNotAttributable

	t.Run("rejected as first party", func(t *testing.T) {
		id := insertReviewedParty(t, ctx, client, coredata.CommonThirdPartyReviewRejected, &firstParty)

		got, gone, err := verdictFor(t, ctx, h, id)
		require.NoError(t, err)
		assert.False(t, gone)
		require.NotNil(t, got, "a rejected row must yield a verdict")
		assert.Equal(t, firstParty, *got)
	})

	t.Run("rejected as not attributable", func(t *testing.T) {
		id := insertReviewedParty(t, ctx, client, coredata.CommonThirdPartyReviewRejected, &notAttributable)

		got, gone, err := verdictFor(t, ctx, h, id)
		require.NoError(t, err)
		assert.False(t, gone)
		require.NotNil(t, got)
		assert.Equal(t, notAttributable, *got,
			"the stored verdict must be used, not a FIRST_PARTY default")
	})

	t.Run("validated row yields no verdict", func(t *testing.T) {
		id := insertReviewedParty(t, ctx, client, coredata.CommonThirdPartyReviewValidated, nil)

		got, gone, err := verdictFor(t, ctx, h, id)
		require.NoError(t, err)
		assert.False(t, gone)
		assert.Nil(t, got, "a validated row must keep its vendor link")
	})

	t.Run("unreviewed row yields no verdict", func(t *testing.T) {
		id := insertReviewedParty(t, ctx, client, coredata.CommonThirdPartyReviewUnreviewed, nil)

		got, gone, err := verdictFor(t, ctx, h, id)
		require.NoError(t, err)
		assert.False(t, gone)
		assert.Nil(t, got,
			"an unreviewed row is the default state and must not divert anything")
	})
}

// The verdict read takes FOR UPDATE, so a review committed by another
// transaction cannot land between the read and the write. Without the lock
// the second transaction commits immediately and the worker persists a
// verdict the catalog has already withdrawn.
func TestRejectedVerdictFor_LocksAgainstConcurrentReview(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	h := newMappingHandler(client)

	firstParty := coredata.CommonTrackerPatternAttributionFirstParty
	id := insertReviewedParty(t, ctx, client, coredata.CommonThirdPartyReviewRejected, &firstParty)

	released := make(chan struct{})
	blocked := make(chan error, 1)

	// Hold the row inside a transaction that read it, then try to validate it
	// from a second connection: that update must wait for the first to finish.
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		verdict, gone, err := h.rejectedVerdictFor(ctx, tx, id)
		require.NoError(t, err)
		require.False(t, gone)
		require.NotNil(t, verdict)

		go func() {
			blocked <- client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
				return (coredata.CommonThirdParty{}).UpdateReview(
					ctx, tx, id,
					coredata.CommonThirdPartyReviewValidated,
					nil,
					"concurrent",
				)
			})

			close(released)
		}()

		// The competing update must still be in flight while the lock is held.
		select {
		case <-released:
			t.Error("the review update completed while the row was locked")
		case <-time.After(300 * time.Millisecond):
		}

		return nil
	}))

	select {
	case err := <-blocked:
		require.NoError(t, err, "the update must succeed once the lock is released")
	case <-time.After(5 * time.Second):
		t.Fatal("the review update never completed after the lock was released")
	}
}

// The id reaching rejectedVerdictFor comes from an earlier transaction, so a
// prune or a merge can delete the row in between. That must be reported as
// gone rather than as "not rejected": the caller has to drop the id, because
// phase three would otherwise try to resolve an org third party for a row
// that no longer exists.
func TestRejectedVerdictFor_ReportsDeletedRow(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	h := newMappingHandler(client)

	firstParty := coredata.CommonTrackerPatternAttributionFirstParty
	id := insertReviewedParty(t, ctx, client, coredata.CommonThirdPartyReviewRejected, &firstParty)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		_, err := tx.Exec(ctx, `DELETE FROM common_third_parties WHERE id = $1`, id)
		return err
	}))

	verdict, gone, err := verdictFor(t, ctx, h, id)
	require.NoError(t, err, "a deleted row is an expected state, not an error")
	assert.True(t, gone, "the caller must be told the row is gone")
	assert.Nil(t, verdict, "no verdict can be applied for a row that does not exist")
}
