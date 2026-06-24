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

package page

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/gid"
)

// loadAllItem is a minimal Paginable used to exercise LoadAll without a
// database. Values are unique so ordering is fully determined by value and
// the id tiebreak is never decisive.
type loadAllItem struct {
	id    gid.GID
	value int
}

func (i *loadAllItem) CursorKey(_ testOrderField) CursorKey {
	return CursorKey{ID: i.id, Value: i.value}
}

// newLoadAllStore builds n items pre-sorted ascending by value (value == i).
func newLoadAllStore(n int) []*loadAllItem {
	tenantID := gid.NewTenantID()

	store := make([]*loadAllItem, n)
	for i := range store {
		store[i] = &loadAllItem{id: gid.New(tenantID, 1), value: i}
	}

	return store
}

// keysetPage mimics the SQL a coredata LoadBy* method runs for a forward
// (ascending, Head) cursor: it returns the over-fetched window the cursor
// asks for (Size+1 rows with no key, Size+2 once a key is set, including the
// boundary row) so NewPage can trim it exactly as it would in production.
func keysetPage(store []*loadAllItem, cursor *Cursor[testOrderField]) []*loadAllItem {
	limit := cursor.Size + 1
	if cursor.Key != nil {
		limit = cursor.Size + 2
	}

	var out []*loadAllItem

	for _, it := range store {
		if cursor.Key != nil {
			keyValue := cursor.Key.Value.(int)
			if it.value < keyValue {
				continue
			}

			if it.value == keyValue && it.id.String() < cursor.Key.ID.String() {
				continue
			}
		}

		out = append(out, it)
		if len(out) == limit {
			break
		}
	}

	return out
}

func loadAllValues(items []*loadAllItem) []int {
	values := make([]int, len(items))
	for i, it := range items {
		values[i] = it.value
	}

	return values
}

func ascOrderBy() OrderBy[testOrderField] {
	return OrderBy[testOrderField]{
		Field:     testOrderField("value"),
		Direction: OrderDirectionAsc,
	}
}

func TestLoadAll(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		count          int
		expectedFetchs int
	}{
		{name: "empty result set", count: 0, expectedFetchs: 1},
		{name: "single short page", count: 10, expectedFetchs: 1},
		{name: "exactly one full page", count: MaxCursorSize, expectedFetchs: 1},
		{name: "two pages", count: MaxCursorSize + 1, expectedFetchs: 2},
		{name: "several pages", count: 2*MaxCursorSize + 7, expectedFetchs: 3},
		{name: "at the page-count limit", count: MaxLoadAllPages * MaxCursorSize, expectedFetchs: MaxLoadAllPages},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			store := newLoadAllStore(tt.count)

			fetchs := 0
			got, err := LoadAll(
				context.Background(),
				ascOrderBy(),
				func(_ context.Context, cursor *Cursor[testOrderField]) ([]*loadAllItem, error) {
					fetchs++
					return keysetPage(store, cursor), nil
				},
			)

			require.NoError(t, err)
			require.Len(t, got, tt.count)
			assert.Equal(t, loadAllValues(store), loadAllValues(got), "every row returned exactly once, in order")
			assert.Equal(t, tt.expectedFetchs, fetchs, "page count")
		})
	}
}

func TestLoadAllPropagatesFetchError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("boom")

	got, err := LoadAll(
		context.Background(),
		ascOrderBy(),
		func(_ context.Context, _ *Cursor[testOrderField]) ([]*loadAllItem, error) {
			return nil, sentinel
		},
	)

	require.ErrorIs(t, err, sentinel)
	assert.Nil(t, got)
}

func TestLoadAllRefusesUnboundedResultSet(t *testing.T) {
	t.Parallel()

	// One row past MaxLoadAllPages full pages keeps HasNext true after the
	// last allowed page, so LoadAll must bail instead of looping forever.
	store := newLoadAllStore(MaxLoadAllPages*MaxCursorSize + 1)

	fetchs := 0
	got, err := LoadAll(
		context.Background(),
		ascOrderBy(),
		func(_ context.Context, cursor *Cursor[testOrderField]) ([]*loadAllItem, error) {
			fetchs++
			return keysetPage(store, cursor), nil
		},
	)

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "result set exceeds")
	assert.Equal(t, MaxLoadAllPages, fetchs, "stops after walking the max number of pages")
}
