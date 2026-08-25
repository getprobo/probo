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

package commonthirdparty

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

// A reviewed row carries a decision, so prune must leave it alone even when
// nothing references it — which is exactly the state a row lands in once its
// misattributed patterns are detached.
func TestPruneSkipsReviewedRows(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	insert := func(review coredata.CommonThirdPartyReview, verdict *coredata.CommonTrackerPatternAttribution) coredata.CommonThirdParty {
		t.Helper()

		// Older than the 24h guard so age is not what keeps it.
		created := time.Now().Add(-72 * time.Hour)
		id := gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType)
		party := coredata.CommonThirdParty{
			ID:              id,
			Name:            "Prune " + id.String(),
			Slug:            "prune-" + id.String(),
			Category:        coredata.ThirdPartyCategoryOther,
			Certifications:  []string{},
			Review:          &review,
			RejectedVerdict: verdict,
			CreatedAt:       created,
			UpdatedAt:       created,
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

	firstParty := coredata.CommonTrackerPatternAttributionFirstParty

	validated := insert(coredata.CommonThirdPartyReviewValidated, nil)
	rejected := insert(coredata.CommonThirdPartyReviewRejected, &firstParty)
	unreviewed := insert(coredata.CommonThirdPartyReviewUnreviewed, nil)

	// LoadAllUnreferencedIDs is what prune selects from; all three qualify,
	// so the review check is the only thing that can separate them.
	var ids []gid.GID

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		var parties coredata.CommonThirdParties

		var err error

		ids, err = parties.LoadAllUnreferencedIDs(ctx, conn, time.Now().Add(-24*time.Hour), false)

		return err
	}))

	found := map[gid.GID]bool{}
	for _, id := range ids {
		found[id] = true
	}

	assert.True(t, found[validated.ID], "the query itself must see all three as unreferenced")
	assert.True(t, found[rejected.ID])
	assert.True(t, found[unreviewed.ID])

	// And the reviewed ones must carry a state prune can filter on.
	require.NotNil(t, validated.Review)
	assert.NotEqual(t, coredata.CommonThirdPartyReviewUnreviewed, *validated.Review)
	require.NotNil(t, rejected.Review)
	assert.NotEqual(t, coredata.CommonThirdPartyReviewUnreviewed, *rejected.Review)
	require.NotNil(t, unreviewed.Review)
	assert.Equal(t, coredata.CommonThirdPartyReviewUnreviewed, *unreviewed.Review)
}
