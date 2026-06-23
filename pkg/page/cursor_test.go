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
	"encoding/binary"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.probo.inc/probo/pkg/gid"
)

type testOrderField string

func (f testOrderField) Column() string { return string(f) }
func (f testOrderField) String() string { return string(f) }

func TestNewCursor(t *testing.T) {
	t.Parallel()

	orderBy := OrderBy[testOrderField]{
		Field:     testOrderField("created_at"),
		Direction: OrderDirectionAsc,
	}

	tests := []struct {
		name         string
		size         int
		expectedSize int
	}{
		{
			name:         "negative size defaults to DefaultCursorSize",
			size:         -10,
			expectedSize: DefaultCursorSize,
		},
		{
			name:         "zero size defaults to DefaultCursorSize",
			size:         0,
			expectedSize: DefaultCursorSize,
		},
		{
			name:         "valid size is kept as-is",
			size:         50,
			expectedSize: 50,
		},
		{
			name:         "oversized value is clamped to MaxCursorSize",
			size:         500,
			expectedSize: MaxCursorSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cursor := NewCursor[testOrderField](tt.size, nil, Head, orderBy)
			assert.Equal(t, tt.expectedSize, cursor.Size)
		})
	}
}

const testField = "valid_from"

func newKey(value any) *CursorKey {
	k := NewCursorKey(mkID(0), value)
	return &k
}

// TestSQLFragment locks the exact SQL emitted for every direction/position/key
// combination. keyKind picks the boundary: none (first page), a non-NULL value,
// or a NULL value. @cursor_field_value must appear only in the non-NULL seek.
func TestSQLFragment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		direction OrderDirection
		position  Position
		keyKind   string
		want      string
	}{
		{
			name:      "Head ASC without key",
			direction: OrderDirectionAsc,
			position:  Head,
			keyKind:   "none",
			want:      "TRUE ORDER BY valid_from ASC NULLS LAST, id ASC LIMIT @cursor_limit",
		},
		{
			name:      "Head DESC without key",
			direction: OrderDirectionDesc,
			position:  Head,
			keyKind:   "none",
			want:      "TRUE ORDER BY valid_from DESC NULLS LAST, id DESC LIMIT @cursor_limit",
		},
		{
			name:      "Head ASC, non-NULL boundary (NULLS LAST)",
			direction: OrderDirectionAsc,
			position:  Head,
			keyKind:   "nonnull",
			want: "( valid_from > @cursor_field_value OR ( valid_from = @cursor_field_value AND id >= @cursor_id ) OR valid_from IS NULL ) " +
				"ORDER BY valid_from ASC NULLS LAST, id ASC LIMIT @cursor_limit",
		},
		{
			name:      "Head DESC, non-NULL boundary (NULLS LAST)",
			direction: OrderDirectionDesc,
			position:  Head,
			keyKind:   "nonnull",
			want: "( valid_from < @cursor_field_value OR ( valid_from = @cursor_field_value AND id <= @cursor_id ) OR valid_from IS NULL ) " +
				"ORDER BY valid_from DESC NULLS LAST, id DESC LIMIT @cursor_limit",
		},
		{
			name:      "Head ASC, NULL boundary (NULLS LAST)",
			direction: OrderDirectionAsc,
			position:  Head,
			keyKind:   "null",
			want:      "( valid_from IS NULL AND id >= @cursor_id ) ORDER BY valid_from ASC NULLS LAST, id ASC LIMIT @cursor_limit",
		},
		{
			name:      "Head DESC, NULL boundary (NULLS LAST)",
			direction: OrderDirectionDesc,
			position:  Head,
			keyKind:   "null",
			want:      "( valid_from IS NULL AND id <= @cursor_id ) ORDER BY valid_from DESC NULLS LAST, id DESC LIMIT @cursor_limit",
		},
		{
			name:      "Tail DESC, non-NULL boundary (reversed to ASC NULLS FIRST)",
			direction: OrderDirectionDesc,
			position:  Tail,
			keyKind:   "nonnull",
			want: "( valid_from > @cursor_field_value OR ( valid_from = @cursor_field_value AND id >= @cursor_id ) ) " +
				"ORDER BY valid_from ASC NULLS FIRST, id ASC LIMIT @cursor_limit",
		},
		{
			name:      "Tail ASC, non-NULL boundary (reversed to DESC NULLS FIRST)",
			direction: OrderDirectionAsc,
			position:  Tail,
			keyKind:   "nonnull",
			want: "( valid_from < @cursor_field_value OR ( valid_from = @cursor_field_value AND id <= @cursor_id ) ) " +
				"ORDER BY valid_from DESC NULLS FIRST, id DESC LIMIT @cursor_limit",
		},
		{
			name:      "Tail DESC, NULL boundary (reversed to ASC NULLS FIRST)",
			direction: OrderDirectionDesc,
			position:  Tail,
			keyKind:   "null",
			want: "( valid_from IS NULL AND id >= @cursor_id OR valid_from IS NOT NULL ) " +
				"ORDER BY valid_from ASC NULLS FIRST, id ASC LIMIT @cursor_limit",
		},
		{
			name:      "Tail ASC, NULL boundary (reversed to DESC NULLS FIRST)",
			direction: OrderDirectionAsc,
			position:  Tail,
			keyKind:   "null",
			want: "( valid_from IS NULL AND id <= @cursor_id OR valid_from IS NOT NULL ) " +
				"ORDER BY valid_from DESC NULLS FIRST, id DESC LIMIT @cursor_limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var key *CursorKey

			switch tt.keyKind {
			case "nonnull":
				key = newKey(time.Unix(0, 0))
			case "null":
				key = newKey(nil)
			}

			cur := NewCursor[testOrderField](
				10,
				key,
				tt.position,
				OrderBy[testOrderField]{Field: testField, Direction: tt.direction},
			)

			assert.Equal(t, tt.want, cur.SQLFragment())
		})
	}
}

// TestSQLArguments checks that @cursor_field_value is bound only for the
// non-NULL seek that references it, keeping StrictNamedArgs consistent with
// the generated SQL.
func TestSQLArguments(t *testing.T) {
	t.Parallel()

	orderBy := OrderBy[testOrderField]{Field: testField, Direction: OrderDirectionDesc}

	noKey := NewCursor[testOrderField](10, nil, Head, orderBy).SQLArguments()
	assert.NotContains(t, noKey, "cursor_id")
	assert.NotContains(t, noKey, "cursor_field_value")
	// One look-ahead row over the page size when there is no boundary.
	assert.Equal(t, 11, noKey["cursor_limit"])

	nonNull := NewCursor[testOrderField](10, newKey(time.Unix(0, 0)), Head, orderBy).SQLArguments()
	assert.Contains(t, nonNull, "cursor_id")
	assert.Contains(t, nonNull, "cursor_field_value")
	// Inclusive boundary row plus one look-ahead row.
	assert.Equal(t, 12, nonNull["cursor_limit"])

	nullVal := NewCursor[testOrderField](10, newKey(nil), Tail, orderBy).SQLArguments()
	assert.Contains(t, nullVal, "cursor_id")
	assert.NotContains(t, nullVal, "cursor_field_value")
	assert.Equal(t, 12, nullVal["cursor_limit"])
}

func mkID(n int) gid.GID {
	var id gid.GID
	binary.BigEndian.PutUint64(id[16:24], uint64(n+1))

	return id
}
