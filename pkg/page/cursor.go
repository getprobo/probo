// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

func (c *Cursor[T]) SQLFragment() string {
	field := c.OrderBy.Field.Column()
	value := cursorValue(field)
	direction := c.sqlDirection()
	// NULLs sort after every real value. A reverse page flips that with the scan.
	nullsLast := c.Position != Tail

	where := "TRUE"
	if c.Key != nil {
		where = cursorWhere(field, value, direction, nullsLast, c.Key.Value == nil)
	}

	nulls := " NULLS FIRST"
	if nullsLast {
		nulls = " NULLS LAST"
	}

	orderBy := field + " " + direction + nulls + ", id " + direction

	return where + " ORDER BY " + orderBy + " LIMIT @cursor_limit"
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

	if c.Key == nil {
		return arguments
	}

	arguments["cursor_id"] = c.Key.ID
	if strings.Contains(c.SQLFragment(), "@cursor_field_value") {
		arguments["cursor_field_value"] = c.Key.Value
	}

	return arguments
}

// cursorValue keeps the cursor key's raw value. When Column is a single-argument
// call such as section_title_sort_key(section_title), the comparison applies
// that same function to the bound value.
func cursorValue(column string) string {
	name, arg, ok := sqlCall(column)
	if !ok || !sqlIdent(name) || !sqlIdent(arg) {
		return "@cursor_field_value"
	}

	return name + "(@cursor_field_value)"
}

func sqlCall(column string) (string, string, bool) {
	open := strings.IndexByte(column, '(')
	if open <= 0 || !strings.HasSuffix(column, ")") || strings.Contains(column[open+1:], "(") {
		return "", "", false
	}

	return column[:open], column[open+1 : len(column)-1], true
}

func sqlIdent(s string) bool {
	if s == "" {
		return false
	}

	for i, r := range s {
		letter := r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')

		digit := r >= '0' && r <= '9'
		if !letter && (i == 0 || !digit) {
			return false
		}
	}

	return true
}

func (c *Cursor[T]) sqlDirection() string {
	switch {
	case c.OrderBy.Direction == OrderDirectionAsc && c.Position == Head:
		return "ASC"
	case c.OrderBy.Direction == OrderDirectionDesc && c.Position == Head:
		return "DESC"
	case c.OrderBy.Direction == OrderDirectionAsc && c.Position == Tail:
		return "DESC"
	case c.OrderBy.Direction == OrderDirectionDesc && c.Position == Tail:
		return "ASC"
	default:
		return ""
	}
}

func cursorWhere(field string, value string, direction string, nullsLast bool, valueNull bool) string {
	cmp := ">="
	idExclude := "id < @cursor_id"

	if direction == "DESC" {
		cmp = "<="
		idExclude = "id > @cursor_id"
	}

	matched := "(" + field + " " + cmp + " " + value + ") AND NOT (" + field + " = " + value + " AND " + idExclude + ")"

	if !nullsLast {
		if valueNull {
			return "NOT (" + field + " IS NULL AND " + idExclude + ")"
		}

		return matched
	}

	if valueNull {
		return field + " IS NULL AND NOT (" + idExclude + ")"
	}

	return "(" + field + " IS NULL OR (" + matched + "))"
}
