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
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// ErrInvalidCursorValue is a cursor value that cannot be compared with the
// ordered column. Callers refuse it before querying.
var ErrInvalidCursorValue = errors.New("invalid cursor")

type cursorValueKind int

const (
	cursorValueRaw cursorValueKind = iota
	cursorValueDate
	cursorValueTimestamp
	cursorValueNumber
)

var cursorIdentPattern = regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]*`)

// coerceCursorValue parses a cursor key into the type PostgreSQL will compare
// it with. A text parameter compared with a date fails inside the database.
func coerceCursorValue(column string, value any) (any, error) {
	if value == nil || cursorValueNil(value) {
		return nil, nil
	}

	switch cursorColumnKind(column) {
	case cursorValueDate:
		return coerceCursorDate(value)
	case cursorValueTimestamp:
		return coerceCursorTimestamp(value)
	case cursorValueNumber:
		return coerceCursorNumber(value)
	default:
		return value, nil
	}
}

func cursorValueNil(value any) bool {
	switch v := value.(type) {
	case *time.Time:
		return v == nil
	default:
		return false
	}
}

func cursorColumnKind(column string) cursorValueKind {
	if strings.Contains(column, "||") || strings.Contains(strings.ToLower(column), "to_char") {
		return cursorValueRaw
	}

	var kind cursorValueKind
	for _, ident := range cursorIdentPattern.FindAllString(stripSQLQuotes(column), -1) {
		next := cursorIdentKind(strings.ToLower(ident))
		if next == cursorValueRaw {
			continue
		}

		if kind != cursorValueRaw && kind != next {
			return cursorValueRaw
		}

		kind = next
	}

	return kind
}

func stripSQLQuotes(column string) string {
	var b strings.Builder
	quoted := false
	for i := 0; i < len(column); i++ {
		if column[i] == '\'' {
			quoted = !quoted
			continue
		}

		if !quoted {
			b.WriteByte(column[i])
		}
	}

	return b.String()
}

func cursorIdentKind(ident string) cursorValueKind {
	switch ident {
	case "valid_from", "valid_until", "identified_on", "deadline":
		return cursorValueDate
	}

	switch {
	case strings.HasSuffix(ident, "_date"):
		return cursorValueDate
	case strings.HasSuffix(ident, "_at"):
		return cursorValueTimestamp
	case strings.HasSuffix(ident, "_due"):
		return cursorValueTimestamp
	case ident == "rank" || strings.HasSuffix(ident, "_rank") || strings.HasSuffix(ident, "_score"):
		return cursorValueNumber
	default:
		return cursorValueRaw
	}
}

func coerceCursorDate(value any) (pgtype.Date, error) {
	switch v := value.(type) {
	case time.Time:
		return cursorDate(v), nil
	case *time.Time:
		if v == nil {
			return pgtype.Date{}, nil
		}

		return cursorDate(*v), nil
	case string:
		if t, err := time.Parse(time.DateOnly, v); err == nil {
			return cursorDate(t), nil
		}

		if t, err := time.Parse(time.RFC3339Nano, v); err == nil {
			return cursorDate(t), nil
		}

		return pgtype.Date{}, ErrInvalidCursorValue
	default:
		return pgtype.Date{}, ErrInvalidCursorValue
	}
}

func cursorDate(t time.Time) pgtype.Date {
	year, month, day := t.UTC().Date()

	return pgtype.Date{
		Time:  time.Date(year, month, day, 0, 0, 0, 0, time.UTC),
		Valid: true,
	}
}

func coerceCursorTimestamp(value any) (any, error) {
	switch v := value.(type) {
	case time.Time:
		return v, nil
	case *time.Time:
		if v == nil {
			return nil, nil
		}

		return *v, nil
	case string:
		t, err := time.Parse(time.RFC3339Nano, v)
		if err != nil {
			return nil, ErrInvalidCursorValue
		}

		return t, nil
	default:
		return nil, ErrInvalidCursorValue
	}
}

func coerceCursorNumber(value any) (any, error) {
	switch v := value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return v, nil
	case string:
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n, nil
		}

		n, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, ErrInvalidCursorValue
		}

		return n, nil
	default:
		return nil, ErrInvalidCursorValue
	}
}
