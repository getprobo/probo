// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package page

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPageFirstLast(t *testing.T) {
	t.Parallel()

	empty := &Page[*loadAllItem, testOrderField]{}
	assert.Nil(t, empty.First())
	assert.Nil(t, empty.Last())

	a := &loadAllItem{id: mkID(0), value: 10}
	b := &loadAllItem{id: mkID(1), value: 20}
	full := &Page[*loadAllItem, testOrderField]{Data: []*loadAllItem{a, b}}
	assert.Same(t, a, full.First())
	assert.Same(t, b, full.Last())
}

// TestNewPage locks the over-fetch/trim contract: callers fetch one extra row
// with no key (Size+1) or two extra with a key (Size+2, the inclusive boundary
// row plus the look-ahead), and NewPage trims them while deriving HasNext/
// HasPrev. Tail fetches the window in reverse (boundary first) and NewPage
// flips it back to display order.
func TestNewPage(t *testing.T) {
	t.Parallel()

	// Fixed ascending dataset A<B<C<D<E by both value and id.
	a := &loadAllItem{id: mkID(0), value: 10}
	b := &loadAllItem{id: mkID(1), value: 20}
	c := &loadAllItem{id: mkID(2), value: 30}
	d := &loadAllItem{id: mkID(3), value: 40}
	e := &loadAllItem{id: mkID(4), value: 50}

	keyFor := func(it *loadAllItem) *CursorKey {
		k := CursorKey{ID: it.id, Value: it.value}
		return &k
	}

	const size = 2

	tests := []struct {
		name     string
		position Position
		key      *CursorKey
		data     []*loadAllItem // raw fetched window, in SQL return order
		want     []int
		wantNext bool
		wantPrev bool
	}{
		{
			name:     "head/no key/empty",
			position: Head,
			data:     nil,
			want:     []int{},
		},
		{
			name:     "head/no key/partial page",
			position: Head,
			data:     []*loadAllItem{a, b},
			want:     []int{10, 20},
		},
		{
			name:     "head/no key/full page has next",
			position: Head,
			data:     []*loadAllItem{a, b, c}, // Size+1
			want:     []int{10, 20},
			wantNext: true,
		},
		{
			name:     "head/key/full page has next and prev",
			position: Head,
			key:      keyFor(b),
			data:     []*loadAllItem{b, c, d, e}, // Size+2, boundary b first
			want:     []int{30, 40},
			wantNext: true,
			wantPrev: true,
		},
		{
			name:     "head/key/last page keeps prev only",
			position: Head,
			key:      keyFor(c),
			data:     []*loadAllItem{c, d, e}, // boundary + remainder
			want:     []int{40, 50},
			wantPrev: true,
		},
		{
			name:     "tail/no key/empty",
			position: Tail,
			data:     nil,
			want:     []int{},
		},
		{
			name:     "tail/no key/partial page",
			position: Tail,
			data:     []*loadAllItem{b, a}, // reversed order, fewer than Size+1
			want:     []int{10, 20},
		},
		{
			name:     "tail/no key/full last page has prev",
			position: Tail,
			data:     []*loadAllItem{e, d, c}, // Size+1, reversed
			want:     []int{40, 50},
			wantPrev: true,
		},
		{
			name:     "tail/key/full page has next and prev",
			position: Tail,
			key:      keyFor(d),
			data:     []*loadAllItem{d, c, b, a}, // Size+2, boundary d first
			want:     []int{20, 30},
			wantNext: true,
			wantPrev: true,
		},
		{
			name:     "tail/key/first page keeps next only",
			position: Tail,
			key:      keyFor(c),
			data:     []*loadAllItem{c, b, a}, // boundary + remainder
			want:     []int{10, 20},
			wantNext: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cur := NewCursor(size, tt.key, tt.position, ascOrderBy())
			pg := NewPage(tt.data, cur)

			assert.Equal(t, tt.want, loadAllValues(pg.Data), "trimmed page contents")
			assert.Equal(t, tt.wantNext, pg.Info.HasNext, "HasNext")
			assert.Equal(t, tt.wantPrev, pg.Info.HasPrev, "HasPrev")
		})
	}
}
