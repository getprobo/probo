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
	"testing"

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

func TestCursor_SQLFragment_NullsLast(t *testing.T) {
	t.Parallel()

	id := gid.New(gid.NewTenantID(), 1)
	valueKey := NewCursorKey(id, "v")
	nullKey := NewCursorKey(id, nil)
	orderBy := OrderBy[testOrderField]{
		Field:     "due",
		Direction: OrderDirectionDesc,
	}

	tests := []struct {
		name     string
		position Position
		key      *CursorKey
		want     string
	}{
		{
			name:     "descending head includes nulls after the cursor value",
			position: Head,
			key:      &valueKey,
			want:     "(due IS NULL OR ((due <= @cursor_field_value) AND NOT (due = @cursor_field_value AND id > @cursor_id))) ORDER BY due DESC NULLS LAST, id DESC LIMIT @cursor_limit",
		},
		{
			name:     "descending tail walks with nulls first",
			position: Tail,
			key:      &valueKey,
			want:     "(due >= @cursor_field_value) AND NOT (due = @cursor_field_value AND id < @cursor_id) ORDER BY due ASC NULLS FIRST, id ASC LIMIT @cursor_limit",
		},
		{
			name:     "descending head on a null cursor stays in the nulls",
			position: Head,
			key:      &nullKey,
			want:     "due IS NULL AND NOT (id > @cursor_id) ORDER BY due DESC NULLS LAST, id DESC LIMIT @cursor_limit",
		},
		{
			name:     "descending tail on a null cursor keeps later values",
			position: Tail,
			key:      &nullKey,
			want:     "NOT (due IS NULL AND id < @cursor_id) ORDER BY due ASC NULLS FIRST, id ASC LIMIT @cursor_limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cursor := NewCursor(10, tt.key, tt.position, orderBy)
			assert.Equal(t, tt.want, cursor.SQLFragment())
		})
	}

	t.Run("ascending head includes nulls after the cursor value", func(t *testing.T) {
		t.Parallel()

		cursor := NewCursor(10, &valueKey, Head, OrderBy[testOrderField]{
			Field:     "due",
			Direction: OrderDirectionAsc,
		})

		assert.Equal(
			t,
			"(due IS NULL OR ((due >= @cursor_field_value) AND NOT (due = @cursor_field_value AND id < @cursor_id))) ORDER BY due ASC NULLS LAST, id ASC LIMIT @cursor_limit",
			cursor.SQLFragment(),
		)
	})
}

func TestCursor_SQLFragment_DefaultAndMapped(t *testing.T) {
	t.Parallel()

	id := gid.New(gid.NewTenantID(), 1)
	valueKey := NewCursorKey(id, "v")

	t.Run("a single-argument column function also wraps the raw cursor value", func(t *testing.T) {
		t.Parallel()

		cursor := NewCursor(10, &valueKey, Head, OrderBy[testOrderField]{
			Field:     "section_title_sort_key(section_title)",
			Direction: OrderDirectionDesc,
		})

		assert.Equal(
			t,
			"(section_title_sort_key(section_title) IS NULL OR ((section_title_sort_key(section_title) <= section_title_sort_key(@cursor_field_value)) AND NOT (section_title_sort_key(section_title) = section_title_sort_key(@cursor_field_value) AND id > @cursor_id))) ORDER BY section_title_sort_key(section_title) DESC NULLS LAST, id DESC LIMIT @cursor_limit",
			cursor.SQLFragment(),
		)
	})

	t.Run("a multi-argument expression compares the raw cursor value", func(t *testing.T) {
		t.Parallel()

		cursor := NewCursor(10, &valueKey, Head, OrderBy[testOrderField]{
			Field:     "LEAST(next_evidence_due, next_test_due)",
			Direction: OrderDirectionDesc,
		})

		assert.Equal(
			t,
			"(LEAST(next_evidence_due, next_test_due) IS NULL OR ((LEAST(next_evidence_due, next_test_due) <= @cursor_field_value) AND NOT (LEAST(next_evidence_due, next_test_due) = @cursor_field_value AND id > @cursor_id))) ORDER BY LEAST(next_evidence_due, next_test_due) DESC NULLS LAST, id DESC LIMIT @cursor_limit",
			cursor.SQLFragment(),
		)
	})

	t.Run("plain fields are unchanged", func(t *testing.T) {
		t.Parallel()

		cursor := NewCursor(10, &valueKey, Head, OrderBy[testOrderField]{
			Field:     "name",
			Direction: OrderDirectionDesc,
		})

		assert.Equal(
			t,
			"(name IS NULL OR ((name <= @cursor_field_value) AND NOT (name = @cursor_field_value AND id > @cursor_id))) ORDER BY name DESC NULLS LAST, id DESC LIMIT @cursor_limit",
			cursor.SQLFragment(),
		)
	})
}

func TestCursor_SQLArguments_NullsOmitsEmptyValue(t *testing.T) {
	t.Parallel()

	id := gid.New(gid.NewTenantID(), 1)
	key := NewCursorKey(id, nil)
	cursor := NewCursor(10, &key, Head, OrderBy[testOrderField]{
		Field:     "due",
		Direction: OrderDirectionDesc,
	})

	args := cursor.SQLArguments()
	_, hasValue := args["cursor_field_value"]
	assert.False(t, hasValue)
	assert.Equal(t, id, args["cursor_id"])
	assert.NotContains(t, cursor.SQLFragment(), "@cursor_field_value")
}
