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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
)

// The review filters are what let a consumer stop offering a rejected row
// without deleting it, so both directions need to hold: WithReview selects
// a state and WithoutReview excludes one.
func TestCommonThirdPartyFilter_Review(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	unreviewed := seedCommonThirdParty(t, ctx, client)
	rejected := seedCommonThirdParty(t, ctx, client)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		_, err := tx.Exec(ctx,
			`UPDATE common_third_parties
			 SET review = 'REJECTED', rejected_verdict = 'NOT_ATTRIBUTABLE'
			 WHERE id = $1`, rejected.ID)

		return err
	}))

	load := func(mutate func(*coredata.CommonThirdPartyFilter)) []string {
		t.Helper()

		filter := coredata.NewCommonThirdPartyFilter(nil)
		mutate(filter)

		var parties coredata.CommonThirdParties

		require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
			return parties.LoadAll(ctx, conn, filter)
		}))

		slugs := make([]string, 0, len(parties))
		for _, p := range parties {
			if p.Slug == unreviewed.Slug || p.Slug == rejected.Slug {
				slugs = append(slugs, p.Slug)
			}
		}

		return slugs
	}

	rejectedState := coredata.CommonThirdPartyReviewRejected
	unreviewedState := coredata.CommonThirdPartyReviewUnreviewed

	assert.Equal(t,
		[]string{rejected.Slug},
		load(func(f *coredata.CommonThirdPartyFilter) { f.WithReview(&rejectedState) }),
		"WithReview(REJECTED) must select only the rejected row")

	assert.Equal(t,
		[]string{unreviewed.Slug},
		load(func(f *coredata.CommonThirdPartyFilter) { f.WithReview(&unreviewedState) }),
		"WithReview(UNREVIEWED) must select only the unreviewed row")

	assert.Equal(t,
		[]string{unreviewed.Slug},
		load(func(f *coredata.CommonThirdPartyFilter) { f.WithoutReview(&rejectedState) }),
		"WithoutReview(REJECTED) must hide the rejected row and keep the rest")

	assert.ElementsMatch(t,
		[]string{unreviewed.Slug, rejected.Slug},
		load(func(f *coredata.CommonThirdPartyFilter) {}),
		"an unfiltered load must still see both")
}
