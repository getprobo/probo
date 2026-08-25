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

// The grouped loader exists so a reviewer can read every row's keys in one
// round trip: doing it per row is what makes reading the evidence slower than
// guessing from the name. Pin the grouping, the confidence ordering the
// caller's cap relies on, and that unlinked patterns stay out.
func TestLoadSummariesGroupedByCommonThirdPartyID(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	party := seedCommonThirdParty(t, ctx, client)
	other := seedCommonThirdParty(t, ctx, client)

	insert := func(owner *coredata.CommonThirdParty, key string, confidence float32) {
		t.Helper()

		now := time.Now().UTC().Truncate(time.Microsecond)
		pattern := coredata.CommonTrackerPattern{
			ID:          gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType),
			TrackerType: coredata.TrackerTypeCookie,
			Pattern:     key,
			MatchType:   coredata.TrackerPatternMatchTypeExact,
			Confidence:  confidence,
			Attribution: coredata.CommonTrackerPatternAttributionThirdParty,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if owner != nil {
			pattern.CommonThirdPartyID = &owner.ID
		}

		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			return pattern.Insert(ctx, tx)
		}))

		t.Cleanup(func() {
			_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
				_, err := tx.Exec(ctx, `DELETE FROM common_tracker_patterns WHERE id = $1`, pattern.ID)
				return err
			})
		})
	}

	insert(&party, "zzz-low-confidence", 0.20)
	insert(&party, "aaa-high-confidence", 0.90)
	insert(&other, "other-party-key", 0.80)
	insert(nil, "unlinked-key", 0.70)

	var byParty map[gid.GID][]coredata.PatternSummary

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		var patterns coredata.CommonTrackerPatterns

		var err error

		byParty, err = patterns.LoadSummariesGroupedByCommonThirdPartyID(ctx, conn)

		return err
	}))

	mine := byParty[party.ID]
	require.Len(t, mine, 2, "only this entry's patterns")

	assert.Equal(t, "aaa-high-confidence", mine[0].Pattern,
		"highest confidence first, since the caller shows only the first few")
	assert.Equal(t, "zzz-low-confidence", mine[1].Pattern)
	assert.Equal(t, coredata.TrackerTypeCookie, mine[0].TrackerType,
		"the storage kind is what separates an egressing cookie from local state")

	require.Len(t, byParty[other.ID], 1, "grouping must not leak across entries")

	for owner, summaries := range byParty {
		for _, s := range summaries {
			assert.NotEqual(t, "unlinked-key", s.Pattern,
				"a pattern with no vendor has no entry to group under (owner %s)", owner)
		}
	}
}
