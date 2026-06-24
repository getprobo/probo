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
