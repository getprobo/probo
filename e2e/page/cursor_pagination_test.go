// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

// Package page_test exercises the keyset cursor (pkg/page) against a real
// Postgres instance, without going through GraphQL. It builds adversarial
// datasets (NULL boundaries, ties, NULL blocks larger than the page) that are
// hard to construct over the API, then walks every page forward and backward
// using the production Cursor.SQLFragment/SQLArguments and asserts the walk
// reconstructs the exact unpaged order with no row dropped or duplicated.
package page_test

import (
	"context"
	"fmt"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

// orderColumn is the nullable keyset column under test.
const orderColumn = orderField("valid_from")

type orderField string

func (f orderField) Column() string { return string(f) }
func (f orderField) String() string { return string(f) }

// item is a Paginable row backed by a (id, valid_from) tuple where valid_from
// may be NULL.
type item struct {
	id  gid.GID
	val *time.Time
}

func (i item) CursorKey(_ orderField) page.CursorKey {
	if i.val == nil {
		return page.NewCursorKey(i.id, nil)
	}

	return page.NewCursorKey(i.id, *i.val)
}

func TestCursorPaginationNullAware(t *testing.T) {
	t.Parallel()

	datasets := []struct {
		name string
		rows []item
	}{
		{"empty", build()},
		{"single non-null", build(at(10))},
		{"single null", build(null())},
		{"no nulls", build(at(0), at(3), at(6), at(9), at(12), at(15), at(18), at(21))},
		{"all nulls", build(null(), null(), null(), null(), null(), null())},
		// NULL block larger than the page: a boundary is forced to land on a
		// NULL value, the case the old predicate silently truncated.
		{"nulls exceed page", build(at(40), at(35), at(30), at(25), null(), null(), null(), null(), null(), null(), null(), null(), null(), null())},
		// NULL block smaller than the page.
		{"nulls below page", build(at(50), at(45), at(40), at(35), at(30), at(25), null(), null())},
		// Repeated values exercise the id tie-break alongside NULLs.
		{"ties with nulls", build(at(10), at(10), at(10), at(10), at(20), at(20), at(20), null(), null(), null(), null(), null())},
		// Ties straddling the NULL block boundary at every page size.
		{"ties at boundary", build(at(5), at(5), at(5), null(), null(), null())},
	}

	for _, ds := range datasets {
		t.Run(ds.name, func(t *testing.T) {
			t.Parallel()

			conn := connect(t)
			seed(t, conn, ds.rows)

			for _, dir := range []page.OrderDirection{page.OrderDirectionAsc, page.OrderDirectionDesc} {
				for _, size := range []int{1, 2, 3, 5, 100} {
					want := groundTruth(ds.rows, dir)

					t.Run(fmt.Sprintf("%s/size=%d/forward", dir, size), func(t *testing.T) {
						got := walk(t, conn, dir, size, page.Head)
						require.Equal(t, want, got, "forward walk must match unpaged order")
					})

					t.Run(fmt.Sprintf("%s/size=%d/backward", dir, size), func(t *testing.T) {
						got := walk(t, conn, dir, size, page.Tail)
						require.Equal(t, want, got, "backward (Tail) walk must reconstruct the same order")
					})
				}
			}
		})
	}
}

// walk paginates the whole set in pages of the given size using the production
// Cursor + NewPage, faithfully serializing/parsing the cursor between pages as
// the GraphQL layer does, and returns the collected ids in display order.
func walk(t *testing.T, conn *pgx.Conn, dir page.OrderDirection, size int, pos page.Position) []string {
	t.Helper()

	orderBy := page.OrderBy[orderField]{Field: orderColumn, Direction: dir}

	var key *page.CursorKey

	collected := []string{}

	for guard := 0; ; guard++ {
		require.Less(t, guard, 10000, "pagination did not terminate")

		cur := page.NewCursor[orderField](size, key, pos, orderBy)
		pg := page.NewPage[item, orderField](fetch(t, conn, cur), cur)

		ids := make([]string, 0, len(pg.Data))
		for _, r := range pg.Data {
			ids = append(ids, r.id.String())
		}

		switch pos {
		case page.Head:
			collected = append(collected, ids...)
			if !pg.Info.HasNext {
				return collected
			}

			next := roundTrip(t, pg.Last().CursorKey(orderColumn))
			key = &next
		case page.Tail:
			collected = append(ids, collected...)
			if !pg.Info.HasPrev {
				return collected
			}

			next := roundTrip(t, pg.First().CursorKey(orderColumn))
			key = &next
		}
	}
}

// roundTrip serializes a cursor key and parses it back, mirroring the
// client<->server round-trip so the boundary value arrives as the decoded JSON
// type (a string for timestamps), exactly as the API binds it.
func roundTrip(t *testing.T, k page.CursorKey) page.CursorKey {
	t.Helper()

	parsed, err := page.ParseCursorKey(k.String())
	require.NoError(t, err)

	return parsed
}

// fetch runs the production SQL fragment + named arguments against Postgres.
func fetch(t *testing.T, conn *pgx.Conn, cur *page.Cursor[orderField]) []item {
	t.Helper()

	sql := "SELECT id, valid_from FROM items WHERE " + cur.SQLFragment()

	rows, err := conn.Query(context.Background(), sql, cur.SQLArguments())
	require.NoError(t, err)

	defer rows.Close()

	var out []item

	for rows.Next() {
		var r item

		require.NoError(t, rows.Scan(&r.id, &r.val))

		out = append(out, r)
	}

	require.NoError(t, rows.Err())

	return out
}

// groundTruth is the independent oracle: the canonical unpaged order Postgres
// produces for "ORDER BY valid_from <dir> NULLS LAST, id <dir>".
func groundTruth(rows []item, dir page.OrderDirection) []string {
	asc := dir == page.OrderDirectionAsc

	sorted := append([]item(nil), rows...)
	sort.SliceStable(sorted, func(i, j int) bool {
		a, b := sorted[i], sorted[j]

		switch {
		case a.val == nil && b.val == nil:
			return idLess(a.id, b.id, asc)
		case a.val == nil:
			return false // NULLS LAST: a sorts after b.
		case b.val == nil:
			return true
		case a.val.Equal(*b.val):
			return idLess(a.id, b.id, asc)
		case asc:
			return a.val.Before(*b.val)
		default:
			return a.val.After(*b.val)
		}
	})

	ids := make([]string, 0, len(sorted))
	for _, r := range sorted {
		ids = append(ids, r.id.String())
	}

	return ids
}

func idLess(a, b gid.GID, asc bool) bool {
	if asc {
		return a.String() < b.String()
	}

	return a.String() > b.String()
}

func connect(t *testing.T) *pgx.Conn {
	t.Helper()

	url := os.Getenv("PROBO_TEST_PG_URL")
	if url == "" {
		url = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}

	conn, err := pgx.Connect(context.Background(), url)
	require.NoError(t, err, "connect to postgres (override with PROBO_TEST_PG_URL)")

	t.Cleanup(func() { _ = conn.Close(context.Background()) })

	return conn
}

// seed creates a per-session temp table (id ordered by byte value to match the
// Go oracle) and inserts the dataset.
func seed(t *testing.T, conn *pgx.Conn, rows []item) {
	t.Helper()

	ctx := context.Background()

	_, err := conn.Exec(ctx, `CREATE TEMP TABLE items (id TEXT COLLATE "C" PRIMARY KEY, valid_from TIMESTAMPTZ)`)
	require.NoError(t, err)

	for _, r := range rows {
		_, err := conn.Exec(ctx, `INSERT INTO items (id, valid_from) VALUES ($1, $2)`, r.id, r.val)
		require.NoError(t, err)
	}
}

// build materializes a dataset from cell specs.
func build(cells ...*time.Time) []item {
	rows := make([]item, 0, len(cells))
	for _, c := range cells {
		rows = append(rows, item{id: gid.New(gid.NewTenantID(), 1), val: c})
	}

	return rows
}

func at(sec int64) *time.Time {
	v := time.Unix(sec, 0).UTC()
	return &v
}

func null() *time.Time { return nil }
