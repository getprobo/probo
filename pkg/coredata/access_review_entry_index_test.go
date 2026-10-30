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
)

func TestAccessReviewEntryIndexes_LeadWithFilterColumns(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)

	for _, tc := range []struct {
		table  string
		column string
	}{
		{table: "access_review_entries", column: "access_review_campaign_id"},
		{table: "access_review_entry_decision_history", column: "access_review_entry_id"},
	} {
		t.Run(
			tc.table+" by "+tc.column,
			func(t *testing.T) {
				t.Parallel()

				var exists bool

				require.NoError(
					t,
					client.WithConn(
						context.Background(),
						func(ctx context.Context, conn pg.Querier) error {
							return conn.QueryRow(
								ctx,
								`
SELECT EXISTS (
    SELECT 1
    FROM pg_index i
    JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = i.indkey[0]
    WHERE i.indrelid = $1::regclass
        AND a.attname = $2
        AND i.indisvalid
        AND i.indpred IS NULL
)`,
								tc.table,
								tc.column,
							).Scan(&exists)
						},
					),
				)

				assert.True(t, exists, "no index on %s leads with %s", tc.table, tc.column)
			},
		)
	}
}
