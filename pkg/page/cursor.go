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
	direction := c.sqlDirection()
	nulls := c.sqlNulls()

	where := "TRUE"
	if c.Key != nil {
		where = cursorWhere(field, direction, nulls, c.Key.Value == nil)
	}

	orderBy := field + " " + direction + nullsSQL(nulls) + ", id " + direction

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

func fieldNulls[T OrderField](field T) Nulls {
	ordered, ok := any(field).(NullsOrderField)
	if !ok {
		return NullsDefault
	}

	return ordered.Nulls()
}

// Tail walks the order backwards, so FIRST and LAST swap.
func (c *Cursor[T]) sqlNulls() Nulls {
	nulls := fieldNulls(c.OrderBy.Field)
	if nulls == NullsDefault || c.Position == Head {
		return nulls
	}

	if nulls == NullsLast {
		return NullsFirst
	}

	return NullsLast
}

func nullsSQL(nulls Nulls) string {
	switch nulls {
	case NullsFirst:
		return " NULLS FIRST"
	case NullsLast:
		return " NULLS LAST"
	default:
		return ""
	}
}

func cursorWhere(field string, direction string, nulls Nulls, valueNull bool) string {
	cmp := ">="
	idExclude := "id < @cursor_id"
	if direction == "DESC" {
		cmp = "<="
		idExclude = "id > @cursor_id"
	}

	matched := "(" + field + " " + cmp + " @cursor_field_value) AND NOT (" + field + " = @cursor_field_value AND " + idExclude + ")"

	switch nulls {
	case NullsLast:
		if valueNull {
			return field + " IS NULL AND NOT (" + idExclude + ")"
		}

		return "(" + field + " IS NULL OR (" + matched + "))"
	case NullsFirst:
		if valueNull {
			return "NOT (" + field + " IS NULL AND " + idExclude + ")"
		}

		return matched
	default:
		return matched
	}
}
