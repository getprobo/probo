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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// ErrInvalidCursorValue is a cursor value that does not match the type stored
// in the cursor. Callers refuse it before querying.
var ErrInvalidCursorValue = errors.New("invalid cursor")

// CursorKind is the type of the value stored in a cursor key.
type CursorKind string

const (
	CursorText   CursorKind = "text"
	CursorTime   CursorKind = "time"
	CursorDate   CursorKind = "date"
	CursorNumber CursorKind = "number"
)

func classifyCursorValue(value any) (CursorKind, any) {
	if value == nil {
		return CursorText, nil
	}

	switch v := value.(type) {
	case time.Time:
		return CursorTime, v
	case *time.Time:
		if v == nil {
			return CursorTime, nil
		}

		return CursorTime, *v
	case pgtype.Date:
		if !v.Valid {
			return CursorDate, nil
		}

		return CursorDate, v
	}

	rv := reflect.ValueOf(value)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return kindOf(rv.Type().Elem()), nil
		}

		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.String:
		return CursorText, rv.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return CursorNumber, rv.Interface()
	}

	if s, ok := rv.Interface().(fmt.Stringer); ok {
		return CursorText, s.String()
	}

	return CursorText, fmt.Sprint(rv.Interface())
}

func kindOf(t reflect.Type) CursorKind {
	if t == reflect.TypeOf(time.Time{}) {
		return CursorTime
	}

	switch t.Kind() {
	case reflect.String:
		return CursorText
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return CursorNumber
	default:
		return CursorText
	}
}

func dateCursorValue(value time.Time) pgtype.Date {
	year, month, day := value.UTC().Date()

	return pgtype.Date{
		Time:  time.Date(year, month, day, 0, 0, 0, 0, time.UTC),
		Valid: true,
	}
}

func marshalCursorValue(kind CursorKind, value any) ([]byte, error) {
	if value == nil {
		return []byte("null"), nil
	}

	switch kind {
	case CursorTime:
		t, ok := value.(time.Time)
		if !ok {
			return nil, ErrInvalidCursorValue
		}

		return json.Marshal(t.Format(time.RFC3339Nano))
	case CursorDate:
		d, ok := value.(pgtype.Date)
		if !ok || !d.Valid {
			return nil, ErrInvalidCursorValue
		}

		return json.Marshal(d.Time.Format(time.DateOnly))
	case CursorNumber, CursorText:
		return json.Marshal(value)
	default:
		return nil, ErrInvalidCursorValue
	}
}

func unmarshalCursorValue(kind CursorKind, raw []byte) (any, error) {
	if bytes.Equal(raw, []byte("null")) {
		return nil, nil
	}

	switch kind {
	case CursorTime:
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, ErrInvalidCursorValue
		}

		t, err := time.Parse(time.RFC3339Nano, s)
		if err != nil {
			return nil, ErrInvalidCursorValue
		}

		return t, nil
	case CursorDate:
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, ErrInvalidCursorValue
		}

		t, err := time.Parse(time.DateOnly, s)
		if err != nil {
			return nil, ErrInvalidCursorValue
		}

		return dateCursorValue(t), nil
	case CursorNumber:
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.UseNumber()

		var n json.Number
		if err := dec.Decode(&n); err != nil {
			return nil, ErrInvalidCursorValue
		}

		if i, err := n.Int64(); err == nil {
			return i, nil
		}

		f, err := n.Float64()
		if err != nil {
			return nil, ErrInvalidCursorValue
		}

		return f, nil
	case CursorText:
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, ErrInvalidCursorValue
		}

		return s, nil
	default:
		return nil, ErrInvalidCursorValue
	}
}
