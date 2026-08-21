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

	assert.Equal(t, coredata.CommonThirdPartyReviewUnreviewed, loaded.Review)
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
	assert.Equal(t, coredata.CommonThirdPartyReviewRejected, loaded.Review)
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
