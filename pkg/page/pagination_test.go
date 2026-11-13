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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/gid"
)

// testOrderField is the sort column. Column() is the SQL expression the
// keyset compares; the cursor value stays the raw value from the row.
type testOrderField string

func (f testOrderField) Column() string { return string(f) }
func (f testOrderField) String() string { return string(f) }

// record is one row. CursorKey is the value the next page resumes from.
type record struct {
	id    gid.GID
	value any
}

func (r record) CursorKey(testOrderField) CursorKey {
	return NewCursorKey(r.id, r.value)
}

var (
	_ OrderField                = testOrderField("")
	_ Paginable[testOrderField] = record{}
)

// The loader over-fetches so NewPage can see the page boundary.
//
// Head, no cursor: at most size+1 rows in display order. The extra row is
// the lookahead. NewPage drops it and sets HasNext.
//
// Head, with a cursor: at most size+2 rows. The first row is the cursor
// row. The last row, when the slice is full, is the lookahead. NewPage
// drops the cursor row, drops the lookahead when present, and sets HasPrev
// only when that first row is the requested cursor.
//
// Tail is the same fetch with the scan reversed. NewPage restores display
// order, then trims. HasNext is the cursor-row match. HasPrev is the
// lookahead.
//
// NULLs sort after every real value. A tail scan flips that to NULLS FIRST
// because the comparison itself is flipped.

func TestNewCursor_Size(t *testing.T) {
	t.Parallel()

	orderBy := OrderBy[testOrderField]{
		Field:     "created_at",
		Direction: OrderDirectionAsc,
	}

	tests := []struct {
		name string
		size int
		want int
	}{
		{name: "negative size defaults to 25", size: -10, want: DefaultCursorSize},
		{name: "zero size defaults to 25", size: 0, want: DefaultCursorSize},
		{name: "a size inside the range is kept", size: 50, want: 50},
		{name: "the maximum size is kept", size: MaxCursorSize, want: MaxCursorSize},
		{name: "a size past the maximum is clamped", size: MaxCursorSize + 1, want: MaxCursorSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cursor := NewCursor[testOrderField](tt.size, nil, Head, orderBy)
			assert.Equal(t, tt.want, cursor.Size)
		})
	}
}

func TestNewPage(t *testing.T) {
	t.Parallel()

	rows := newRecords("a", "b", "c", "d")
	a, b, c, d := rows[0], rows[1], rows[2], rows[3]
	cursorA := a.CursorKey("name")
	cursorC := c.CursorKey("name")
	cursorD := d.CursorKey("name")
	other := NewCursorKey(gid.New(gid.NewTenantID(), 1), "a")

	tests := []struct {
		name     string
		position Position
		rows     []record
		key      *CursorKey
		want     []record
		hasNext  bool
		hasPrev  bool
	}{
		{
			name:     "head without a cursor returns a short page",
			position: Head,
			rows:     []record{a},
			want:     []record{a},
		},
		{
			name:     "head without a cursor keeps a full page that has no lookahead",
			position: Head,
			rows:     []record{a, b},
			want:     []record{a, b},
		},
		{
			name:     "head without a cursor drops the lookahead and reports a next page",
			position: Head,
			rows:     []record{a, b, c},
			want:     []record{a, b},
			hasNext:  true,
		},
		{
			name:     "head from a cursor drops the boundary row and reports a previous page",
			position: Head,
			rows:     []record{a, b},
			key:      &cursorA,
			want:     []record{b},
			hasPrev:  true,
		},
		{
			name:     "head from a cursor drops the boundary and the lookahead",
			position: Head,
			rows:     []record{a, b, c, d},
			key:      &cursorA,
			want:     []record{b, c},
			hasNext:  true,
			hasPrev:  true,
		},
		{
			name:     "head does not report a previous page when the first row is not the cursor",
			position: Head,
			rows:     []record{a, b, c, d},
			key:      &other,
			want:     []record{b, c},
			hasNext:  true,
		},
		{
			name:     "head from a cursor that is the only row is empty and has a previous page",
			position: Head,
			rows:     []record{a},
			key:      &cursorA,
			want:     []record{},
			hasPrev:  true,
		},
		{
			name:     "tail without a cursor reverses the scan and drops the lookahead",
			position: Tail,
			rows:     []record{c, b, a},
			want:     []record{b, c},
			hasPrev:  true,
		},
		{
			name:     "tail without a cursor keeps a short page",
			position: Tail,
			rows:     []record{b, a},
			want:     []record{a, b},
		},
		{
			name:     "tail from a cursor reverses, drops the boundary, and reports a next page",
			position: Tail,
			rows:     []record{c, b, a},
			key:      &cursorC,
			want:     []record{a, b},
			hasNext:  true,
		},
		{
			name:     "tail from a cursor drops both extras and reports both directions",
			position: Tail,
			rows:     []record{d, c, b, a},
			key:      &cursorD,
			want:     []record{b, c},
			hasNext:  true,
			hasPrev:  true,
		},
		{
			name:     "tail does not report a next page when the first scanned row is not the cursor",
			position: Tail,
			rows:     []record{c, b, a},
			key:      &other,
			want:     []record{a, b},
		},
		{
			name:     "tail from a cursor that is the only row is empty and has a next page",
			position: Tail,
			rows:     []record{a},
			key:      &cursorA,
			want:     []record{},
			hasNext:  true,
		},
		{
			name:     "a position other than head or tail is returned unchanged",
			position: Position("OTHER"),
			rows:     []record{a, b, c},
			key:      &cursorA,
			want:     []record{a, b, c},
		},
		{
			name:     "an empty result has no page cursor",
			position: Head,
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cursor := NewCursor(2, tt.key, tt.position, OrderBy[testOrderField]{
				Field:     "name",
				Direction: OrderDirectionAsc,
			})
			input := append([]record(nil), tt.rows...)
			got := NewPage(input, cursor)

			assertPage(t, got, cursor, tt.rows, tt.want, tt.hasNext, tt.hasPrev)
		})
	}
}

func TestPage_ForwardAndBack(t *testing.T) {
	t.Parallel()

	rows := newRecords("a", "b", "c", "d", "e")
	orderBy := OrderBy[testOrderField]{Field: "name", Direction: OrderDirectionAsc}

	// Windows a size-2 head scan would return, including the boundary row
	// and the lookahead. Each window starts at the previous page's last row.
	forward := [][]record{
		rows[0:3],
		rows[1:5],
		rows[3:5],
	}

	var key *CursorKey

	var seen []string

	for i, window := range forward {
		cursor := NewCursor(2, key, Head, orderBy)
		got := NewPage(append([]record(nil), window...), cursor)

		for _, row := range got.Data {
			seen = append(seen, row.value.(string))
		}

		if !got.Info.HasNext {
			assert.Equal(t, len(forward)-1, i)
			break
		}

		next := got.Last().CursorKey(orderBy.Field)
		key = &next
	}

	assert.Equal(t, []string{"a", "b", "c", "d", "e"}, seen)

	// The last page was d, e. Walking back starts with no cursor (the end
	// of the list) and then resumes from each page's first row.
	backward := []struct {
		window []record
		key    *CursorKey
	}{
		{window: []record{rows[4], rows[3], rows[2]}},
		{window: []record{rows[3], rows[2], rows[1], rows[0]}, key: keyOf(rows[3])},
		{window: []record{rows[1], rows[0]}, key: keyOf(rows[1])},
	}

	seen = nil

	for _, step := range backward {
		cursor := NewCursor(2, step.key, Tail, orderBy)
		got := NewPage(append([]record(nil), step.window...), cursor)
		page := make([]string, len(got.Data))

		for i, row := range got.Data {
			page[i] = row.value.(string)
		}

		seen = append(page, seen...)
	}

	assert.Equal(t, []string{"a", "b", "c", "d", "e"}, seen)
}

func TestCursor_SQLFragment(t *testing.T) {
	t.Parallel()

	id := gid.New(gid.NewTenantID(), 1)
	valueKey := NewCursorKey(id, "v")
	nullKey := NewCursorKey(id, nil)

	// name is a plain column. Every scan direction and cursor shape is
	// listed so a change to the keyset predicate fails here.
	const (
		headAscValue = "(name IS NULL OR ((name >= @cursor_field_value) AND NOT (name = @cursor_field_value AND id < @cursor_id))) ORDER BY name ASC NULLS LAST, id ASC LIMIT @cursor_limit"
		headAscNull  = "name IS NULL AND NOT (id < @cursor_id) ORDER BY name ASC NULLS LAST, id ASC LIMIT @cursor_limit"
		headAscNone  = "TRUE ORDER BY name ASC NULLS LAST, id ASC LIMIT @cursor_limit"

		headDescValue = "(name IS NULL OR ((name <= @cursor_field_value) AND NOT (name = @cursor_field_value AND id > @cursor_id))) ORDER BY name DESC NULLS LAST, id DESC LIMIT @cursor_limit"
		headDescNull  = "name IS NULL AND NOT (id > @cursor_id) ORDER BY name DESC NULLS LAST, id DESC LIMIT @cursor_limit"
		headDescNone  = "TRUE ORDER BY name DESC NULLS LAST, id DESC LIMIT @cursor_limit"

		tailAscValue = "(name <= @cursor_field_value) AND NOT (name = @cursor_field_value AND id > @cursor_id) ORDER BY name DESC NULLS FIRST, id DESC LIMIT @cursor_limit"
		tailAscNull  = "NOT (name IS NULL AND id > @cursor_id) ORDER BY name DESC NULLS FIRST, id DESC LIMIT @cursor_limit"
		tailAscNone  = "TRUE ORDER BY name DESC NULLS FIRST, id DESC LIMIT @cursor_limit"

		tailDescValue = "(name >= @cursor_field_value) AND NOT (name = @cursor_field_value AND id < @cursor_id) ORDER BY name ASC NULLS FIRST, id ASC LIMIT @cursor_limit"
		tailDescNull  = "NOT (name IS NULL AND id < @cursor_id) ORDER BY name ASC NULLS FIRST, id ASC LIMIT @cursor_limit"
		tailDescNone  = "TRUE ORDER BY name ASC NULLS FIRST, id ASC LIMIT @cursor_limit"
	)

	tests := []struct {
		name      string
		column    testOrderField
		direction OrderDirection
		position  Position
		key       *CursorKey
		want      string
	}{
		{name: "head asc has no cursor", column: "name", direction: OrderDirectionAsc, position: Head, want: headAscNone},
		{name: "head asc continues after a value and still includes later nulls", column: "name", direction: OrderDirectionAsc, position: Head, key: &valueKey, want: headAscValue},
		{name: "head asc on a null cursor stays inside the nulls", column: "name", direction: OrderDirectionAsc, position: Head, key: &nullKey, want: headAscNull},

		{name: "head desc has no cursor", column: "name", direction: OrderDirectionDesc, position: Head, want: headDescNone},
		{name: "head desc continues after a value and still includes later nulls", column: "name", direction: OrderDirectionDesc, position: Head, key: &valueKey, want: headDescValue},
		{name: "head desc on a null cursor stays inside the nulls", column: "name", direction: OrderDirectionDesc, position: Head, key: &nullKey, want: headDescNull},

		{name: "tail asc scans backward with nulls first", column: "name", direction: OrderDirectionAsc, position: Tail, want: tailAscNone},
		{name: "tail asc has already passed the nulls", column: "name", direction: OrderDirectionAsc, position: Tail, key: &valueKey, want: tailAscValue},
		{name: "tail asc on a null cursor keeps the rows after those nulls", column: "name", direction: OrderDirectionAsc, position: Tail, key: &nullKey, want: tailAscNull},

		{name: "tail desc scans backward with nulls first", column: "name", direction: OrderDirectionDesc, position: Tail, want: tailDescNone},
		{name: "tail desc has already passed the nulls", column: "name", direction: OrderDirectionDesc, position: Tail, key: &valueKey, want: tailDescValue},
		{name: "tail desc on a null cursor keeps the rows after those nulls", column: "name", direction: OrderDirectionDesc, position: Tail, key: &nullKey, want: tailDescNull},

		{
			name:      "an unknown direction compares ascending and still sorts nulls last",
			column:    "name",
			direction: OrderDirection("SIDEWAYS"),
			position:  Head,
			key:       &valueKey,
			want:      "(name IS NULL OR ((name >= @cursor_field_value) AND NOT (name = @cursor_field_value AND id < @cursor_id))) ORDER BY name  NULLS LAST, id  LIMIT @cursor_limit",
		},
		{
			name:      "an unknown direction on a tail scan still flips the nulls",
			column:    "name",
			direction: OrderDirection("SIDEWAYS"),
			position:  Tail,
			key:       &valueKey,
			want:      "(name >= @cursor_field_value) AND NOT (name = @cursor_field_value AND id < @cursor_id) ORDER BY name  NULLS FIRST, id  LIMIT @cursor_limit",
		},

		{
			name:      "a single-argument function wraps the raw cursor value",
			column:    "section_title_sort_key(section_title)",
			direction: OrderDirectionDesc,
			position:  Head,
			key:       &valueKey,
			want:      "(section_title_sort_key(section_title) IS NULL OR ((section_title_sort_key(section_title) <= section_title_sort_key(@cursor_field_value)) AND NOT (section_title_sort_key(section_title) = section_title_sort_key(@cursor_field_value) AND id > @cursor_id))) ORDER BY section_title_sort_key(section_title) DESC NULLS LAST, id DESC LIMIT @cursor_limit",
		},
		{
			name:      "a function argument may contain digits and underscores",
			column:    "fn(a1_b)",
			direction: OrderDirectionDesc,
			position:  Head,
			key:       &valueKey,
			want:      "(fn(a1_b) IS NULL OR ((fn(a1_b) <= fn(@cursor_field_value)) AND NOT (fn(a1_b) = fn(@cursor_field_value) AND id > @cursor_id))) ORDER BY fn(a1_b) DESC NULLS LAST, id DESC LIMIT @cursor_limit",
		},
		{
			name:      "an identifier may start with an underscore",
			column:    "fn(_a)",
			direction: OrderDirectionDesc,
			position:  Head,
			key:       &valueKey,
			want:      "(fn(_a) IS NULL OR ((fn(_a) <= fn(@cursor_field_value)) AND NOT (fn(_a) = fn(@cursor_field_value) AND id > @cursor_id))) ORDER BY fn(_a) DESC NULLS LAST, id DESC LIMIT @cursor_limit",
		},
	}

	// These expressions are not a single-argument call, so the cursor value
	// is compared as stored.
	plain := []testOrderField{
		"LEAST(next_evidence_due, next_test_due)",
		"(name)",
		"fn(name",
		"fn(g(name))",
		"1fn(name)",
		"fn(1name)",
		"fn()",
		"fn(a-b)",
		"fn(aé)",
	}

	for _, column := range plain {
		tests = append(tests, struct {
			name      string
			column    testOrderField
			direction OrderDirection
			position  Position
			key       *CursorKey
			want      string
		}{
			name:      "compares the raw value for " + string(column),
			column:    column,
			direction: OrderDirectionDesc,
			position:  Head,
			key:       &valueKey,
			want:      strings.ReplaceAll(headDescValue, "name", string(column)),
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cursor := NewCursor(2, tt.key, tt.position, OrderBy[testOrderField]{
				Field:     tt.column,
				Direction: tt.direction,
			})

			assert.Equal(t, tt.want, cursor.SQLFragment())
			assertCursorArguments(t, cursor, id, tt.key)
		})
	}
}

func TestPage_FirstAndLastOnEmptyPage(t *testing.T) {
	t.Parallel()

	got := NewPage([]record{}, NewCursor(2, nil, Head, OrderBy[testOrderField]{
		Field:     "name",
		Direction: OrderDirectionAsc,
	}))

	assert.Equal(t, record{}, got.First())
	assert.Equal(t, record{}, got.Last())
	assert.False(t, got.Info.HasNext)
	assert.False(t, got.Info.HasPrev)
}

func newRecords(values ...string) []record {
	tenantID := gid.NewTenantID()
	rows := make([]record, len(values))

	for i, value := range values {
		rows[i] = record{id: gid.New(tenantID, 1), value: value}
	}

	return rows
}

func keyOf(row record) *CursorKey {
	key := row.CursorKey("name")
	return &key
}

func assertPage(
	t *testing.T,
	got *Page[record, testOrderField],
	cursor *Cursor[testOrderField],
	input []record,
	want []record,
	hasNext bool,
	hasPrev bool,
) {
	t.Helper()

	require.NotNil(t, got)
	require.NotNil(t, got.Info)
	assert.Equal(t, want, got.Data)
	assert.Equal(t, hasNext, got.Info.HasNext)
	assert.Equal(t, hasPrev, got.Info.HasPrev)

	var first, last record

	if len(want) > 0 {
		first = want[0]
		last = want[len(want)-1]
	}

	assert.Equal(t, first, got.First())
	assert.Equal(t, last, got.Last())

	if len(input) == 0 {
		assert.Nil(t, got.Cursor)
		return
	}

	assert.Same(t, cursor, got.Cursor)
}

func assertCursorArguments(t *testing.T, cursor *Cursor[testOrderField], id gid.GID, key *CursorKey) {
	t.Helper()

	args := cursor.SQLArguments()
	limit := 3

	if key != nil {
		limit = 4
	}

	assert.Equal(t, limit, args["cursor_limit"])

	if key == nil {
		_, hasID := args["cursor_id"]
		_, hasValue := args["cursor_field_value"]

		assert.False(t, hasID)
		assert.False(t, hasValue)

		return
	}

	assert.Equal(t, id, args["cursor_id"])

	if key.Value == nil {
		_, hasValue := args["cursor_field_value"]

		assert.False(t, hasValue)
		assert.NotContains(t, cursor.SQLFragment(), "@cursor_field_value")

		return
	}

	assert.Equal(t, key.Value, args["cursor_field_value"])
}
