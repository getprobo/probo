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

type testDirectionalOrderField string

func (f testDirectionalOrderField) Column() string { return string(f) }
func (f testDirectionalOrderField) String() string { return string(f) }

func (f testDirectionalOrderField) DirectionalColumn(direction OrderDirection) string {
	return string(f) + "_" + string(direction)
}

func (f testDirectionalOrderField) DirectionalCursorValue(direction OrderDirection, placeholder string) string {
	return "map_" + string(direction) + "(" + placeholder + ")"
}

func TestCursor_SQLFragment_DirectionalOrderField(t *testing.T) {
	t.Parallel()

	key := NewCursorKey(gid.New(gid.NewTenantID(), 1), "v")

	t.Run("descending head uses the descending expression", func(t *testing.T) {
		t.Parallel()

		cursor := NewCursor(10, &key, Head, OrderBy[testDirectionalOrderField]{
			Field:     "due",
			Direction: OrderDirectionDesc,
		})

		assert.Equal(
			t,
			"(due_DESC <= map_DESC(@cursor_field_value)) AND NOT (due_DESC = map_DESC(@cursor_field_value) AND id > @cursor_id) ORDER BY due_DESC DESC, id DESC LIMIT @cursor_limit",
			cursor.SQLFragment(),
		)
	})

	t.Run("descending tail keeps the descending expression", func(t *testing.T) {
		t.Parallel()

		cursor := NewCursor(10, &key, Tail, OrderBy[testDirectionalOrderField]{
			Field:     "due",
			Direction: OrderDirectionDesc,
		})

		assert.Equal(
			t,
			"(due_DESC >= map_DESC(@cursor_field_value)) AND NOT (due_DESC = map_DESC(@cursor_field_value) AND id < @cursor_id) ORDER BY due_DESC ASC, id ASC LIMIT @cursor_limit",
			cursor.SQLFragment(),
		)
	})

	t.Run("plain fields are unchanged", func(t *testing.T) {
		t.Parallel()

		cursor := NewCursor(10, &key, Head, OrderBy[testOrderField]{
			Field:     "name",
			Direction: OrderDirectionDesc,
		})

		assert.Equal(
			t,
			"(name <= @cursor_field_value) AND NOT (name = @cursor_field_value AND id > @cursor_id) ORDER BY name DESC, id DESC LIMIT @cursor_limit",
			cursor.SQLFragment(),
		)
	})
}
