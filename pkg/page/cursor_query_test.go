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
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/gid"
)

// These tests run the keyset SQL against Postgres. The expected rows are the
// page a caller sees: nulls after every real value, ties broken by id, and a
// cursor resumed from the raw stored value. A single-argument function is
// applied to that raw value, so section titles sort by the padded key rather
// than by the text itself.

const pageRowsDDL = `
CREATE TEMP TABLE page_rows (
	id text PRIMARY KEY,
	name text,
	section_title text,
	_title text,
	next_evidence_due timestamptz,
	next_test_due timestamptz
)`

const sortFunctionSQL = `
CREATE OR REPLACE FUNCTION %s(text) RETURNS text AS $fn$
DECLARE
	result text := '';
	matches text[];
	remainder text := $1;
BEGIN
	WHILE remainder ~ '\d+' LOOP
		result := result || substring(remainder FROM '^[^\d]*');
		matches := regexp_matches(remainder, '(\d+)', '');

		IF matches IS NOT NULL THEN
			result := result || lpad(matches[1], 10, '0');
			remainder := substring(remainder FROM '\d+(.*)$');
		ELSE
			EXIT;
		END IF;
	END LOOP;

	result := result || remainder;

	RETURN result;
END;
$fn$ LANGUAGE plpgsql IMMUTABLE STRICT`

type storedRow struct {
	id    gid.GID
	value any
	label string
}

func (r storedRow) CursorKey(testOrderField) CursorKey {
	return NewCursorKey(r.id, r.value)
}

func TestCursor_QueryText(t *testing.T) {
	t.Parallel()

	ascending, descending := textRows(t)

	tests := []struct {
		name      string
		direction OrderDirection
		want      []storedRow
	}{
		{name: "ascending", direction: OrderDirectionAsc, want: ascending},
		{name: "descending", direction: OrderDirectionDesc, want: descending},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			withPageDB(t, func(ctx context.Context, tx pgx.Tx) {
				insertColumn(t, ctx, tx, "name", tt.want)
				assertCursorPages(t, ctx, tx, "name", tt.direction, tt.want)
			})
		})
	}
}

func TestCursor_QueryFunction(t *testing.T) {
	t.Parallel()

	// A.2 sorts before A.10. Comparing the raw titles would put A.10 first.
	ascending, descending := titleRows(t)

	tests := []struct {
		name    string
		column  testOrderField
		fn      string
		inserts string
	}{
		{
			name:    "digits and an underscore in the function name",
			column:  "sort_key_1(_title)",
			fn:      "sort_key_1",
			inserts: "_title",
		},
		{
			name:    "uppercase letters in the function name",
			column:  "SortKey(section_title)",
			fn:      "SortKey",
			inserts: "section_title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			directions := []struct {
				name      string
				direction OrderDirection
				want      []storedRow
			}{
				{name: "ascending", direction: OrderDirectionAsc, want: ascending},
				{name: "descending", direction: OrderDirectionDesc, want: descending},
			}

			for _, direction := range directions {
				t.Run(direction.name, func(t *testing.T) {
					t.Parallel()

					withPageDB(t, func(ctx context.Context, tx pgx.Tx) {
						createSortFunction(t, ctx, tx, tt.fn)
						insertColumn(t, ctx, tx, tt.inserts, direction.want)
						assertCursorPages(t, ctx, tx, tt.column, direction.direction, direction.want)
					})
				})
			}
		})
	}
}

func TestCursor_QueryLeast(t *testing.T) {
	t.Parallel()

	// jan/1 has a March evidence date and a January test date. The page
	// follows the earlier of the two, so it sorts with the other January row.
	inserts, ascending, descending := dueRows(t)

	tests := []struct {
		name      string
		direction OrderDirection
		want      []storedRow
	}{
		{name: "ascending", direction: OrderDirectionAsc, want: ascending},
		{name: "descending", direction: OrderDirectionDesc, want: descending},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			withPageDB(t, func(ctx context.Context, tx pgx.Tx) {
				insertDue(t, ctx, tx, inserts)
				assertCursorPages(
					t,
					ctx,
					tx,
					"LEAST(next_evidence_due, next_test_due)",
					tt.direction,
					tt.want,
				)
			})
		})
	}
}

func TestCursor_QueryGroupedColumn(t *testing.T) {
	t.Parallel()

	ascending, _ := textRows(t)

	withPageDB(t, func(ctx context.Context, tx pgx.Tx) {
		insertColumn(t, ctx, tx, "name", ascending)
		assertCursorPages(t, ctx, tx, "(name)", OrderDirectionAsc, ascending)
	})
}

func TestCursor_QueryEmpty(t *testing.T) {
	t.Parallel()

	withPageDB(t, func(ctx context.Context, tx pgx.Tx) {
		assertCursorPages(t, ctx, tx, "name", OrderDirectionAsc, nil)
	})
}

func TestCursor_QuerySingleRow(t *testing.T) {
	t.Parallel()

	row := storedRow{id: newTestIDs(t, 1)[0], value: "only", label: "only"}

	for _, direction := range []OrderDirection{OrderDirectionAsc, OrderDirectionDesc} {
		t.Run(string(direction), func(t *testing.T) {
			t.Parallel()

			withPageDB(t, func(ctx context.Context, tx pgx.Tx) {
				insertColumn(t, ctx, tx, "name", []storedRow{row})
				assertCursorPages(t, ctx, tx, "name", direction, []storedRow{row})
			})
		})
	}
}

func assertCursorPages(
	t *testing.T,
	ctx context.Context,
	tx pgx.Tx,
	field testOrderField,
	direction OrderDirection,
	want []storedRow,
) {
	t.Helper()

	orderBy := OrderBy[testOrderField]{Field: field, Direction: direction}
	index := indexRows(want)
	load := func(cursor *Cursor[testOrderField]) []storedRow {
		t.Helper()

		return queryPage(t, ctx, tx, cursor, index)
	}

	// One page holds the whole table. Head and tail both come back in
	// display order: tail scans backward, and NewPage reverses it.
	pageSize := len(want)
	if pageSize == 0 {
		pageSize = 1
	}

	head := NewCursor(pageSize, nil, Head, orderBy)
	assert.Equal(t, labels(want), labels(NewPage(load(head), head).Data), "full head page")

	tail := NewCursor(pageSize, nil, Tail, orderBy)
	assert.Equal(t, labels(want), labels(NewPage(load(tail), tail).Data), "full tail page")

	for i, row := range want {
		key := row.CursorKey(field)

		forward := NewCursor(pageSize, &key, Head, orderBy)
		forwardPage := NewPage(load(forward), forward)
		assert.Equal(t, labels(want[i+1:]), labels(forwardPage.Data), "head from %s", row.label)
		assert.True(t, forwardPage.Info.HasPrev, "head from %s", row.label)
		assert.False(t, forwardPage.Info.HasNext, "head from %s", row.label)

		backward := NewCursor(pageSize, &key, Tail, orderBy)
		backwardPage := NewPage(load(backward), backward)
		assert.Equal(t, labels(want[:i]), labels(backwardPage.Data), "tail from %s", row.label)
		assert.True(t, backwardPage.Info.HasNext, "tail from %s", row.label)
		assert.False(t, backwardPage.Info.HasPrev, "tail from %s", row.label)
	}

	assert.Equal(t, labels(want), walkPages(t, load, orderBy, Head), "forward walk")
	assert.Equal(t, labels(want), walkPages(t, load, orderBy, Tail), "backward walk")
}

func walkPages(
	t *testing.T,
	load func(*Cursor[testOrderField]) []storedRow,
	orderBy OrderBy[testOrderField],
	position Position,
) []string {
	t.Helper()

	var key *CursorKey

	var pages [][]string

	finished := false

	for range 20 {
		cursor := NewCursor(2, key, position, orderBy)
		got := NewPage(load(cursor), cursor)

		if key == nil && position == Head {
			assert.False(t, got.Info.HasPrev)
		}

		if key == nil && position == Tail {
			assert.False(t, got.Info.HasNext)
		}

		if key != nil && position == Head {
			assert.True(t, got.Info.HasPrev)
		}

		if key != nil && position == Tail {
			assert.True(t, got.Info.HasNext)
		}

		pages = append(pages, labels(got.Data))

		advance := got.Info.HasNext
		if position == Tail {
			advance = got.Info.HasPrev
		}

		if !advance {
			finished = true

			break
		}

		require.NotEmpty(t, got.Data)

		if position == Head {
			next := got.Last().CursorKey(orderBy.Field)
			key = &next

			continue
		}

		prev := got.First().CursorKey(orderBy.Field)
		key = &prev
	}

	require.True(t, finished, "paging did not reach the end")

	if position == Head {
		return joinPages(pages)
	}

	reversed := make([][]string, len(pages))
	for i, page := range pages {
		reversed[len(pages)-1-i] = page
	}

	return joinPages(reversed)
}

func queryPage(
	t *testing.T,
	ctx context.Context,
	tx pgx.Tx,
	cursor *Cursor[testOrderField],
	index map[string]storedRow,
) []storedRow {
	t.Helper()

	statement := "SELECT id FROM page_rows WHERE " + cursor.SQLFragment()
	rows, err := tx.Query(ctx, statement, cursor.SQLArguments())
	require.NoError(t, err, statement)

	defer rows.Close()

	var got []storedRow

	for rows.Next() {
		var id string

		require.NoError(t, rows.Scan(&id))

		row, ok := index[id]
		require.True(t, ok, "query returned unknown id %s", id)

		got = append(got, row)
	}

	require.NoError(t, rows.Err())

	return got
}

func withPageDB(t *testing.T, fn func(context.Context, pgx.Tx)) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	conn := connectPageDB(t, ctx)
	tx, err := conn.Begin(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = tx.Rollback(context.Background())
	})

	_, err = tx.Exec(ctx, pageRowsDDL)
	require.NoError(t, err)

	fn(ctx, tx)
}

func connectPageDB(t *testing.T, ctx context.Context) *pgx.Conn {
	t.Helper()

	dsn := os.Getenv("PROBO_TEST_PG_URL")
	configured := dsn != ""

	if dsn == "" {
		dsn = "postgres://probod:probod@localhost:5432/probod_test"
	}

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		if configured {
			t.Fatalf("cannot connect to configured test database: %v", err)
		}

		t.Skipf("cannot connect to test database: %v", err)
	}

	t.Cleanup(func() {
		_ = conn.Close(context.Background())
	})

	return conn
}

func createSortFunction(t *testing.T, ctx context.Context, tx pgx.Tx, name string) {
	t.Helper()

	_, err := tx.Exec(ctx, fmt.Sprintf(sortFunctionSQL, name))
	require.NoError(t, err)
}

func insertColumn(t *testing.T, ctx context.Context, tx pgx.Tx, column string, rows []storedRow) {
	t.Helper()

	var statement string

	switch column {
	case "name":
		statement = `INSERT INTO page_rows (id, name) VALUES ($1, $2)`
	case "_title":
		statement = `INSERT INTO page_rows (id, _title) VALUES ($1, $2)`
	case "section_title":
		statement = `INSERT INTO page_rows (id, section_title) VALUES ($1, $2)`
	default:
		t.Fatalf("unsupported column %s", column)
	}

	for _, row := range rows {
		_, err := tx.Exec(ctx, statement, row.id, row.value)
		require.NoError(t, err)
	}
}

func insertDue(t *testing.T, ctx context.Context, tx pgx.Tx, rows []dueInsert) {
	t.Helper()

	for _, row := range rows {
		_, err := tx.Exec(
			ctx,
			`INSERT INTO page_rows (id, next_evidence_due, next_test_due) VALUES ($1, $2, $3)`,
			row.row.id,
			row.evidence,
			row.testDue,
		)
		require.NoError(t, err)
	}
}

func textRows(t *testing.T) ([]storedRow, []storedRow) {
	t.Helper()

	ids := newTestIDs(t, 6)
	a1 := storedRow{id: ids[0], value: "a", label: "a/1"}
	a2 := storedRow{id: ids[1], value: "a", label: "a/2"}
	b := storedRow{id: ids[2], value: "b", label: "b"}
	z := storedRow{id: ids[3], value: "z", label: "z"}
	firstNull := storedRow{id: ids[4], value: nil, label: "null/1"}
	secondNull := storedRow{id: ids[5], value: nil, label: "null/2"}

	ascending := []storedRow{a1, a2, b, z, firstNull, secondNull}
	descending := []storedRow{z, b, a2, a1, secondNull, firstNull}

	return ascending, descending
}

func titleRows(t *testing.T) ([]storedRow, []storedRow) {
	t.Helper()

	ids := newTestIDs(t, 6)
	a10First := storedRow{id: ids[0], value: "A.10", label: "A.10/1"}
	a2 := storedRow{id: ids[1], value: "A.2", label: "A.2"}
	a10Second := storedRow{id: ids[2], value: "A.10", label: "A.10/2"}
	b1 := storedRow{id: ids[3], value: "B.1", label: "B.1"}
	firstNull := storedRow{id: ids[4], value: nil, label: "null/1"}
	secondNull := storedRow{id: ids[5], value: nil, label: "null/2"}

	ascending := []storedRow{a2, a10First, a10Second, b1, firstNull, secondNull}
	descending := []storedRow{b1, a10Second, a10First, a2, secondNull, firstNull}

	return ascending, descending
}

type dueInsert struct {
	row      storedRow
	evidence *time.Time
	testDue  *time.Time
}

func dueRows(t *testing.T) ([]dueInsert, []storedRow, []storedRow) {
	t.Helper()

	ids := newTestIDs(t, 6)
	january := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	february := time.Date(2024, time.February, 1, 0, 0, 0, 0, time.UTC)
	march := time.Date(2024, time.March, 1, 0, 0, 0, 0, time.UTC)
	may := time.Date(2024, time.May, 1, 0, 0, 0, 0, time.UTC)
	june := time.Date(2024, time.June, 1, 0, 0, 0, 0, time.UTC)
	januaryFirst := storedRow{id: ids[0], value: january, label: "jan/1"}
	januarySecond := storedRow{id: ids[1], value: january, label: "jan/2"}
	februaryRow := storedRow{id: ids[2], value: february, label: "feb"}
	mayRow := storedRow{id: ids[3], value: may, label: "may"}
	firstNull := storedRow{id: ids[4], value: nil, label: "null/1"}
	secondNull := storedRow{id: ids[5], value: nil, label: "null/2"}

	inserts := []dueInsert{
		{row: januaryFirst, evidence: &march, testDue: &january},
		{row: januarySecond, evidence: &january, testDue: &june},
		{row: februaryRow, testDue: &february},
		{row: mayRow, evidence: &may},
		{row: firstNull},
		{row: secondNull},
	}
	ascending := []storedRow{januaryFirst, januarySecond, februaryRow, mayRow, firstNull, secondNull}
	descending := []storedRow{mayRow, februaryRow, januarySecond, januaryFirst, secondNull, firstNull}

	return inserts, ascending, descending
}

func newTestIDs(t *testing.T, n int) []gid.GID {
	t.Helper()

	ids := make([]gid.GID, n)
	prev := ""

	for i := range ids {
		ids[i][23] = byte(i + 1)
		encoded := ids[i].String()
		require.Greater(t, encoded, prev)

		prev = encoded
	}

	return ids
}

func indexRows(rows []storedRow) map[string]storedRow {
	index := make(map[string]storedRow, len(rows))

	for _, row := range rows {
		index[row.id.String()] = row
	}

	return index
}

func labels(rows []storedRow) []string {
	got := make([]string, len(rows))

	for i, row := range rows {
		got[i] = row.label
	}

	return got
}

func joinPages(pages [][]string) []string {
	got := []string{}

	for _, page := range pages {
		got = append(got, page...)
	}

	return got
}
