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
)

// A literal built without setting Review must still land UNREVIEWED, and the
// round trip must decode the enum rather than an empty string.
func TestCommonThirdPartyReview_DefaultsOnInsert(t *testing.T) {
	t.Parallel()
	client := test.PGClient(t)
	ctx := context.Background()
	party := seedCommonThirdParty(t, ctx, client)

	var loaded coredata.CommonThirdParty

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return loaded.LoadBySlug(ctx, tx, party.Slug)
	}))

	require.NotNil(t, loaded.Review, "a stored row always has a concrete review state")
	assert.Equal(t, coredata.CommonThirdPartyReviewUnreviewed, *loaded.Review)
	assert.Nil(t, loaded.RejectedVerdict)
	assert.Nil(t, loaded.ReviewedAt)
}

// The CHECK must reject a rejection with no verdict, and accept one with.
func TestCommonThirdPartyReview_CheckConstraint(t *testing.T) {
	t.Parallel()
	client := test.PGClient(t)
	ctx := context.Background()
	party := seedCommonThirdParty(t, ctx, client)

	err := client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		_, err := tx.Exec(ctx,
			`UPDATE common_third_parties SET review = 'REJECTED' WHERE id = $1`, party.ID)

		return err
	})
	require.Error(t, err, "REJECTED without a verdict must violate the CHECK")

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		_, err := tx.Exec(ctx,
			`UPDATE common_third_parties
			 SET review = 'REJECTED', rejected_verdict = 'FIRST_PARTY', reviewed_at = $2
			 WHERE id = $1`, party.ID, time.Now())

		return err
	}))

	var loaded coredata.CommonThirdParty

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return loaded.LoadBySlug(ctx, tx, party.Slug)
	}))
	require.NotNil(t, loaded.Review)
	assert.Equal(t, coredata.CommonThirdPartyReviewRejected, *loaded.Review)
	require.NotNil(t, loaded.RejectedVerdict)
	assert.Equal(t, coredata.CommonTrackerPatternAttributionFirstParty, *loaded.RejectedVerdict)
	assert.True(t, loaded.RejectedVerdict.IsTerminal())

	// And a validated row must not be allowed to carry a verdict.
	err = client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		_, err := tx.Exec(ctx,
			`UPDATE common_third_parties
			 SET review = 'VALIDATED', rejected_verdict = 'FIRST_PARTY' WHERE id = $1`, party.ID)

		return err
	})
	require.Error(t, err, "a non-rejected row must not carry a verdict")
}

// Upsert writes review from the receiver, which is what lets the seed promote
// its curated entries to VALIDATED on every run. The risk that buys is a
// human verdict being reset by an unrelated patch, so pin both halves: a
// receiver carrying REJECTED keeps it, and one carrying VALIDATED promotes.
func TestCommonThirdPartyUpsert_PreservesReview(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	party := seedCommonThirdParty(t, ctx, client)
	notAttributable := coredata.CommonTrackerPatternAttributionNotAttributable

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return (coredata.CommonThirdParty{}).UpdateReview(
			ctx, tx, party.ID,
			coredata.CommonThirdPartyReviewRejected,
			&notAttributable,
			"tester",
		)
	}))

	// What proboctl upsert does: load the row, patch a field, write it back.
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		loaded := coredata.CommonThirdParty{}
		if err := loaded.LoadBySlug(ctx, tx, party.Slug); err != nil {
			return err
		}

		loaded.Category = coredata.ThirdPartyCategoryEngineering
		loaded.UpdatedAt = time.Now()

		_, err := loaded.Upsert(ctx, tx)

		return err
	}))

	after := coredata.CommonThirdParty{}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return after.LoadBySlug(ctx, tx, party.Slug)
	}))

	require.NotNil(t, after.Review)
	assert.Equal(t, coredata.CommonThirdPartyReviewRejected, *after.Review,
		"a category patch must not reset a human rejection")
	require.NotNil(t, after.RejectedVerdict)
	assert.Equal(t, notAttributable, *after.RejectedVerdict)
	assert.Equal(t, coredata.ThirdPartyCategoryEngineering, after.Category,
		"the patch itself must still land")

	// And a receiver that asserts curation promotes the row, which is the
	// seed's path on a second run.
	validated := coredata.CommonThirdPartyReviewValidated

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		reseeded := coredata.CommonThirdParty{
			ID:             gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType),
			Name:           party.Name,
			Slug:           party.Slug,
			Category:       coredata.ThirdPartyCategoryOther,
			Certifications: []string{},
			Review:         &validated,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		_, err := reseeded.Upsert(ctx, tx)

		return err
	}))

	promoted := coredata.CommonThirdParty{}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return promoted.LoadBySlug(ctx, tx, party.Slug)
	}))

	require.NotNil(t, promoted.Review)
	assert.Equal(t, coredata.CommonThirdPartyReviewValidated, *promoted.Review,
		"a seed run must promote its curated entry")
	assert.Nil(t, promoted.RejectedVerdict,
		"promotion must clear the verdict the rejection carried")
}

// An auto-create reaches Upsert's conflict branch only by losing a race:
// it looked the slug up, found nothing, and by the time it wrote, another
// writer had created the row. It has no verdict of its own to record, so a
// nil Review must leave whatever is stored alone — otherwise a concurrent
// seed run silently resets a human rejection to UNREVIEWED.
func TestCommonThirdPartyUpsert_NilReviewPreservesStoredVerdict(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	party := seedCommonThirdParty(t, ctx, client)
	notAttributable := coredata.CommonTrackerPatternAttributionNotAttributable

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return (coredata.CommonThirdParty{}).UpdateReview(
			ctx, tx, party.ID,
			coredata.CommonThirdPartyReviewRejected,
			&notAttributable,
			"tester",
		)
	}))

	// What an auto-create does: fresh id, no review asserted, same slug.
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		autoCreated := coredata.CommonThirdParty{
			ID:             gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType),
			Name:           party.Name,
			Slug:           party.Slug,
			Category:       coredata.ThirdPartyCategoryOther,
			Certifications: []string{},
			Review:         nil,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		_, err := autoCreated.Upsert(ctx, tx)

		return err
	}))

	after := coredata.CommonThirdParty{}
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return after.LoadBySlug(ctx, tx, party.Slug)
	}))

	require.NotNil(t, after.Review)
	assert.Equal(t, coredata.CommonThirdPartyReviewRejected, *after.Review,
		"an auto-create must not reset a human rejection")
	require.NotNil(t, after.RejectedVerdict,
		"the verdict must survive alongside the state, or the CHECK would fail")
	assert.Equal(t, notAttributable, *after.RejectedVerdict)
	assert.Equal(t, coredata.ThirdPartyCategoryOther, after.Category,
		"the rest of the upsert must still land")
}

// A nil Review on a genuine insert has nothing to preserve, so the row must
// still land in a valid state rather than violating the NOT NULL column.
func TestCommonThirdPartyUpsert_NilReviewDefaultsOnInsert(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	id := gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType)

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			_, err := tx.Exec(ctx, `DELETE FROM common_third_parties WHERE id = $1`, id)
			return err
		})
	})

	fresh := coredata.CommonThirdParty{
		ID:             id,
		Name:           "Auto " + id.String(),
		Slug:           "auto-" + id.String(),
		Category:       coredata.ThirdPartyCategoryOther,
		Certifications: []string{},
		Review:         nil,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		_, err := fresh.Upsert(ctx, tx)
		return err
	}))

	require.NotNil(t, fresh.Review)
	assert.Equal(t, coredata.CommonThirdPartyReviewUnreviewed, *fresh.Review,
		"a fresh row with no assertion belongs in the backlog")
}
