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
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/gid"
)

func TestCursorKey_RoundTrip(t *testing.T) {
	t.Parallel()

	id := gid.New(gid.NewTenantID(), 1)
	stamp := time.Date(2024, 1, 15, 23, 30, 0, 0, time.FixedZone("EST", -5*60*60))

	tests := []struct {
		name string
		key  CursorKey
		kind CursorKind
	}{
		{
			name: "timestamp",
			key:  NewCursorKey(id, stamp),
			kind: CursorTime,
		},
		{
			name: "date",
			key:  NewDateCursorKey(id, &stamp),
			kind: CursorDate,
		},
		{
			name: "number",
			key:  NewCursorKey(id, 2),
			kind: CursorNumber,
		},
		{
			name: "text",
			key:  NewCursorKey(id, "A.10"),
			kind: CursorText,
		},
		{
			name: "null timestamp",
			key:  NewCursorKey(id, (*time.Time)(nil)),
			kind: CursorTime,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.kind, tt.key.Kind)

			parsed, err := ParseCursorKey(tt.key.String())

			require.NoError(t, err)
			assert.Equal(t, tt.key.String(), parsed.String())
			assert.Equal(t, tt.kind, parsed.Kind)
		})
	}
}

func TestCursorKey_RejectsMismatchedValue(t *testing.T) {
	t.Parallel()

	id := gid.New(gid.NewTenantID(), 1)

	tests := []struct {
		name  string
		kind  CursorKind
		value any
	}{
		{name: "date", kind: CursorDate, value: "' OR 1=1--"},
		{name: "timestamp", kind: CursorTime, value: "' OR 1=1--"},
		{name: "number", kind: CursorNumber, value: "' OR 1=1--"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := ParseCursorKey(cursorToken(t, id, tt.kind, tt.value))

			require.ErrorIs(t, err, ErrInvalidCursorValue)
		})
	}

	t.Run("a cursor without a type is rejected", func(t *testing.T) {
		t.Parallel()

		payload, err := json.Marshal([]any{id.String(), "' OR 1=1--"})
		require.NoError(t, err)

		_, err = ParseCursorKey(base64.RawURLEncoding.EncodeToString(payload))

		require.ErrorIs(t, err, ErrInvalidFormat)
	})
}

func TestCursorKey_DateBindsCalendarDay(t *testing.T) {
	t.Parallel()

	id := gid.New(gid.NewTenantID(), 1)
	stamp := time.Date(2024, 1, 15, 23, 30, 0, 0, time.FixedZone("EST", -5*60*60))
	key := NewDateCursorKey(id, &stamp)
	cursor := NewCursor(10, &key, Head, OrderBy[testOrderField]{
		Field:     "valid_from",
		Direction: OrderDirectionDesc,
	})

	args, err := cursor.SQLArguments()

	require.NoError(t, err)
	assert.Equal(t, pgtype.Date{
		Time:  time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
		Valid: true,
	}, args["cursor_field_value"])
}

func TestCursorKey_TextKeepsRawValue(t *testing.T) {
	t.Parallel()

	id := gid.New(gid.NewTenantID(), 1)
	key := NewCursorKey(id, "' OR 1=1--")
	cursor := NewCursor(10, &key, Head, OrderBy[testOrderField]{
		Field:     "name",
		Direction: OrderDirectionDesc,
	})

	args, err := cursor.SQLArguments()

	require.NoError(t, err)
	assert.Equal(t, "' OR 1=1--", args["cursor_field_value"])
}

func cursorToken(t *testing.T, id gid.GID, kind CursorKind, value any) string {
	t.Helper()

	payload, err := json.Marshal([]any{id.String(), string(kind), value})
	require.NoError(t, err)

	return base64.RawURLEncoding.EncodeToString(payload)
}
