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

		got, err := h.rejectedVerdictFor(ctx, id)
		require.NoError(t, err)
		require.NotNil(t, got, "a rejected row must yield a verdict")
		assert.Equal(t, firstParty, *got)
	})

	t.Run("rejected as not attributable", func(t *testing.T) {
		id := insertReviewedParty(t, ctx, client, coredata.CommonThirdPartyReviewRejected, &notAttributable)

		got, err := h.rejectedVerdictFor(ctx, id)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, notAttributable, *got,
			"the stored verdict must be used, not a FIRST_PARTY default")
	})

	t.Run("validated row yields no verdict", func(t *testing.T) {
		id := insertReviewedParty(t, ctx, client, coredata.CommonThirdPartyReviewValidated, nil)

		got, err := h.rejectedVerdictFor(ctx, id)
		require.NoError(t, err)
		assert.Nil(t, got, "a validated row must keep its vendor link")
	})

	t.Run("unreviewed row yields no verdict", func(t *testing.T) {
		id := insertReviewedParty(t, ctx, client, coredata.CommonThirdPartyReviewUnreviewed, nil)

		got, err := h.rejectedVerdictFor(ctx, id)
		require.NoError(t, err)
		assert.Nil(t, got,
			"an unreviewed row is the default state and must not divert anything")
	})
}
