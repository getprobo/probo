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
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/gid"
)

func TestCursor_SQLArguments_DateColumn(t *testing.T) {
	t.Parallel()

	id := gid.New(gid.NewTenantID(), 1)

	t.Run("a non-date is refused", func(t *testing.T) {
		t.Parallel()

		cursor := dateCursor(id, "VALID_FROM", "' OR 1=1--")

		_, err := cursor.SQLArguments()

		require.ErrorIs(t, err, ErrInvalidCursorValue)
	})

	t.Run("an rfc3339 value becomes the utc calendar day", func(t *testing.T) {
		t.Parallel()

		cursor := dateCursor(id, "VALID_FROM", "2024-01-15T23:30:00-05:00")

		args, err := cursor.SQLArguments()

		require.NoError(t, err)
		assert.Equal(t, pgtype.Date{
			Time:  time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
			Valid: true,
		}, args["cursor_field_value"])
	})

	t.Run("a date-only value is kept", func(t *testing.T) {
		t.Parallel()

		cursor := dateCursor(id, "AUDIT_END_DATE", "2024-06-01")

		args, err := cursor.SQLArguments()

		require.NoError(t, err)
		assert.Equal(t, pgtype.Date{
			Time:  time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
			Valid: true,
		}, args["cursor_field_value"])
	})
}

func TestCursor_SQLArguments_OtherColumns(t *testing.T) {
	t.Parallel()

	id := gid.New(gid.NewTenantID(), 1)

	t.Run("a timestamp column refuses a non-timestamp", func(t *testing.T) {
		t.Parallel()

		cursor := dateCursor(id, "created_at", "' OR 1=1--")

		_, err := cursor.SQLArguments()

		require.ErrorIs(t, err, ErrInvalidCursorValue)
	})

	t.Run("a text column keeps the raw value", func(t *testing.T) {
		t.Parallel()

		cursor := dateCursor(id, "name", "' OR 1=1--")

		args, err := cursor.SQLArguments()

		require.NoError(t, err)
		assert.Equal(t, "' OR 1=1--", args["cursor_field_value"])
	})

	t.Run("a text expression of a timestamp keeps the raw value", func(t *testing.T) {
		t.Parallel()

		column := "(CASE WHEN state = 'DEACTIVATED' THEN '0' ELSE '1' END)" +
			" || to_char(timezone('UTC', created_at), 'YYYYMMDDHH24MISSUS')"
		cursor := dateCursor(id, testOrderField(column), "120240115000000000000")

		args, err := cursor.SQLArguments()

		require.NoError(t, err)
		assert.Equal(t, cursor.Key.Value, args["cursor_field_value"])
	})

	t.Run("a number column refuses a non-number", func(t *testing.T) {
		t.Parallel()

		cursor := dateCursor(id, "priority_rank", "' OR 1=1--")

		_, err := cursor.SQLArguments()

		require.ErrorIs(t, err, ErrInvalidCursorValue)
	})
}

func dateCursor(id gid.GID, field testOrderField, value any) *Cursor[testOrderField] {
	key := NewCursorKey(id, value)

	return NewCursor(10, &key, Head, OrderBy[testOrderField]{
		Field:     field,
		Direction: OrderDirectionDesc,
	})
}
