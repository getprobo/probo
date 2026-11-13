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

package coredata

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func TestPrepareAuditDateCursor(t *testing.T) {
	t.Parallel()

	id := gid.New(gid.NewTenantID(), AuditEntityType)

	t.Run("a sql injection payload is refused", func(t *testing.T) {
		t.Parallel()

		cursor := auditCursor(id, AuditOrderFieldValidFrom, "' OR 1=1--")

		err := prepareAuditDateCursor(cursor)

		require.ErrorIs(t, err, ErrInvalidCursorValue)
		assert.Equal(t, "' OR 1=1--", cursor.Key.Value)
	})

	t.Run("an rfc3339 cursor becomes the utc calendar day", func(t *testing.T) {
		t.Parallel()

		cursor := auditCursor(id, AuditOrderFieldValidFrom, "2024-01-15T00:00:00Z")

		err := prepareAuditDateCursor(cursor)

		require.NoError(t, err)
		assert.Equal(t, "2024-01-15", cursor.Key.Value)
	})

	t.Run("a date-only cursor is kept", func(t *testing.T) {
		t.Parallel()

		cursor := auditCursor(id, AuditOrderFieldAuditEndDate, "2024-06-01")

		err := prepareAuditDateCursor(cursor)

		require.NoError(t, err)
		assert.Equal(t, "2024-06-01", cursor.Key.Value)
	})

	t.Run("a time value becomes the utc calendar day", func(t *testing.T) {
		t.Parallel()

		instant := time.Date(2024, 1, 15, 23, 30, 0, 0, time.FixedZone("EST", -5*60*60))
		cursor := auditCursor(id, AuditOrderFieldValidUntil, instant)

		err := prepareAuditDateCursor(cursor)

		require.NoError(t, err)
		assert.Equal(t, "2024-01-16", cursor.Key.Value)
	})

	t.Run("a non-date order keeps the raw value", func(t *testing.T) {
		t.Parallel()

		cursor := auditCursor(id, AuditOrderFieldCreatedAt, "' OR 1=1--")

		err := prepareAuditDateCursor(cursor)

		require.NoError(t, err)
		assert.Equal(t, "' OR 1=1--", cursor.Key.Value)
	})
}

func auditCursor(id gid.GID, field AuditOrderField, value any) *page.Cursor[AuditOrderField] {
	key := page.NewCursorKey(id, value)

	return page.NewCursor(10, &key, page.Head, page.OrderBy[AuditOrderField]{
		Field:     field,
		Direction: page.OrderDirectionDesc,
	})
}
