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

package page

import (
	"strings"
	"text/template"

	"github.com/jackc/pgx/v5"
)

type (
	Cursor[T OrderField] struct {
		Size     int
		Position Position
		Key      *CursorKey
		OrderBy  OrderBy[T]
	}

	Position string

	OrderBy[T OrderField] struct {
		Field     T
		Direction OrderDirection
	}
)

const (
	DefaultCursorSize = 25
	MaxCursorSize     = 500

	Tail Position = "TAIL"
	Head Position = "HEAD"
)

func NewCursor[T OrderField](size int, from *CursorKey, pos Position, orderBy OrderBy[T]) *Cursor[T] {
	if size <= 0 {
		size = DefaultCursorSize
	} else if size > MaxCursorSize {
		size = MaxCursorSize
	}

	return &Cursor[T]{
		Size:     size,
		Key:      from,
		Position: pos,
		OrderBy:  orderBy,
	}
}

// sqlFragmentData drives sqlFragmentTemplate. Boundary nullness and NULLS
// placement are resolved here (in Go) and select the matching branch, so
// @cursor_field_value only appears in the non-null seek where its type is
// inferable from the column (a bare "@cursor_field_value IS NULL" is not).
type sqlFragmentData struct {
	Column      string
	Cmp         string // strict comparison: < or >
	IDCmp       string // id tie-break: <= or >=
	Direction   string // ASC or DESC
	NullsFirst  bool   // NULLS FIRST when true, else NULLS LAST
	HasKey      bool
	ValueIsNull bool
}

// The null and non-null boundaries are different predicates: a non-null
// boundary spills into the whole NULL block (OR col IS NULL), while a null
// boundary is already inside it and must continue from the cursor id only
// (col IS NULL AND id <op> @cursor_id) to avoid re-seeing rows.
// The null branch also never references @cursor_field_value, whose type
// Postgres cannot infer from a bare NULL bind.
var sqlFragmentTemplate = template.Must(template.New("cursor").Parse(`
	{{if not .HasKey}}
		TRUE
	{{else if .ValueIsNull}}
		(
			{{.Column}} IS NULL AND id {{.IDCmp}} @cursor_id
			{{if .NullsFirst}} OR {{.Column}} IS NOT NULL {{end}}
		)
	{{else}}
		(
			{{.Column}} {{.Cmp}} @cursor_field_value
			OR ( {{.Column}} = @cursor_field_value AND id {{.IDCmp}} @cursor_id )
			{{if not .NullsFirst}} OR {{.Column}} IS NULL {{end}}
		)
	{{end}}
	ORDER BY {{.Column}} {{.Direction}} {{if .NullsFirst}} NULLS FIRST{{else}} NULLS LAST{{end}}, id {{.Direction}}
	LIMIT @cursor_limit
`))

func (c *Cursor[T]) SQLFragment() string {
	// Tail paging is the exact inverse of Head: flip direction and NULLs placement.
	ascending := c.OrderBy.Direction == OrderDirectionAsc
	if c.Position == Tail {
		ascending = !ascending
	}

	data := sqlFragmentData{
		Column:      c.OrderBy.Field.Column(),
		Cmp:         "<",
		IDCmp:       "<=",
		Direction:   "DESC",
		NullsFirst:  c.Position == Tail,
		HasKey:      c.Key != nil,
		ValueIsNull: c.Key != nil && c.Key.Value == nil,
	}

	if ascending {
		data.Cmp, data.IDCmp, data.Direction = ">", ">=", "ASC"
	}

	var buf strings.Builder
	if err := sqlFragmentTemplate.Execute(&buf, data); err != nil {
		panic("page: rendering cursor SQL fragment: " + err.Error())
	}

	return strings.Join(strings.Fields(buf.String()), " ")
}

func (c *Cursor[T]) SQLArguments() pgx.NamedArgs {
	var size = c.Size
	if c.Key == nil {
		size += 1
	} else {
		size += 2
	}

	arguments := pgx.NamedArgs{
		"cursor_limit": size,
	}

	if c.Key != nil {
		arguments["cursor_id"] = c.Key.ID

		if c.Key.Value != nil {
			arguments["cursor_field_value"] = c.Key.Value
		}
	}

	return arguments
}
