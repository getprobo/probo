// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	fieldName := c.OrderBy.Field.Column()

	var orderDirection string
	switch {
	case c.OrderBy.Direction == OrderDirectionAsc && c.Position == Head:
		orderDirection = "ASC"
	case c.OrderBy.Direction == OrderDirectionDesc && c.Position == Head:
		orderDirection = "DESC"
	case c.OrderBy.Direction == OrderDirectionAsc && c.Position == Tail:
		orderDirection = "DESC"
	case c.OrderBy.Direction == OrderDirectionDesc && c.Position == Tail:
		orderDirection = "ASC"
	}

	whereClause := "TRUE"
	if c.Key != nil && orderDirection == "DESC" {
		whereClause = "(" + fieldName + " <= @cursor_field_value) AND NOT (" + fieldName + " = @cursor_field_value AND id > @cursor_id)"
	} else if c.Key != nil && orderDirection == "ASC" {
		whereClause = "(" + fieldName + " >= @cursor_field_value) AND NOT (" + fieldName + " = @cursor_field_value AND id < @cursor_id)"
	}

	orderByClause := fieldName + " " + orderDirection + ", id " + orderDirection

	return whereClause + " ORDER BY " + orderByClause + " LIMIT @cursor_limit"
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
		arguments["cursor_field_value"] = c.Key.Value
	}

	return arguments
}
