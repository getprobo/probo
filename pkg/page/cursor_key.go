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
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.probo.inc/probo/pkg/gid"
)

type CursorKey struct {
	ID    gid.GID
	Kind  CursorKind
	Value any
}

// StringCursorKey is a cursor key for string IDs
type StringCursorKey struct {
	ID    string
	Value any
}

var (
	CursorKeyNil CursorKey

	ErrInvalidFormat = errors.New("invalid format")
)

func ParseCursorKey(s string) (CursorKey, error) {
	data, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return CursorKeyNil, ErrInvalidFormat
	}

	var arr []json.RawMessage
	if err := json.Unmarshal(data, &arr); err != nil {
		return CursorKeyNil, ErrInvalidFormat
	}

	key, err := cursorKeyFromArray(arr)
	if err != nil {
		return CursorKeyNil, err
	}

	return key, nil
}

func NewCursorKey(id gid.GID, value any) CursorKey {
	kind, typed := classifyCursorValue(value)

	return CursorKey{
		ID:    id,
		Kind:  kind,
		Value: typed,
	}
}

// NewDateCursorKey encodes a calendar day. A time.Time passed to NewCursorKey
// is a timestamp, and a date column cannot be compared with one.
func NewDateCursorKey(id gid.GID, value any) CursorKey {
	switch v := value.(type) {
	case nil:
		return CursorKey{ID: id, Kind: CursorDate}
	case time.Time:
		return CursorKey{ID: id, Kind: CursorDate, Value: dateCursorValue(v)}
	case *time.Time:
		if v == nil {
			return CursorKey{ID: id, Kind: CursorDate}
		}

		return CursorKey{ID: id, Kind: CursorDate, Value: dateCursorValue(*v)}
	default:
		panic(fmt.Sprintf("date cursor value must be a time, got %T", value))
	}
}

func (ck CursorKey) Bytes() []byte {
	data, _ := ck.MarshalBinary()
	return data
}

func (ck CursorKey) String() string {
	data, err := ck.MarshalBinary()
	if err != nil {
		return ""
	}

	return base64.RawURLEncoding.EncodeToString(data)
}

func (ck CursorKey) FieldValue() any {
	return ck.Value
}

func (ck CursorKey) MarshalText() ([]byte, error) {
	return []byte(ck.String()), nil
}

func (ck *CursorKey) UnmarshalText(data []byte) error {
	newCk, err := ParseCursorKey(string(data))
	if err != nil {
		return err
	}

	*ck = newCk

	return nil
}

func (ck CursorKey) MarshalBinary() ([]byte, error) {
	idJSON, err := json.Marshal(ck.ID.String())
	if err != nil {
		return nil, err
	}

	kindJSON, err := json.Marshal(string(ck.Kind))
	if err != nil {
		return nil, err
	}

	valueJSON, err := marshalCursorValue(ck.Kind, ck.Value)
	if err != nil {
		return nil, err
	}

	return json.Marshal([]json.RawMessage{idJSON, kindJSON, valueJSON})
}

func (ck *CursorKey) UnmarshalBinary(data []byte) error {
	var arr []json.RawMessage
	if err := json.Unmarshal(data, &arr); err != nil {
		return ErrInvalidFormat
	}

	key, err := cursorKeyFromArray(arr)
	if err != nil {
		return err
	}

	*ck = key

	return nil
}

func cursorKeyFromArray(arr []json.RawMessage) (CursorKey, error) {
	if len(arr) != 3 {
		return CursorKeyNil, ErrInvalidFormat
	}

	var idStr string
	if err := json.Unmarshal(arr[0], &idStr); err != nil {
		return CursorKeyNil, ErrInvalidFormat
	}

	id, err := gid.ParseGID(idStr)
	if err != nil {
		return CursorKeyNil, ErrInvalidFormat
	}

	var kind string
	if err := json.Unmarshal(arr[1], &kind); err != nil {
		return CursorKeyNil, ErrInvalidFormat
	}

	value, err := unmarshalCursorValue(CursorKind(kind), arr[2])
	if err != nil {
		return CursorKeyNil, err
	}

	return CursorKey{
		ID:    id,
		Kind:  CursorKind(kind),
		Value: value,
	}, nil
}

func (ck CursorKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(ck.String())
}

func (ck *CursorKey) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	parsed, err := ParseCursorKey(s)
	if err != nil {
		return err
	}

	*ck = parsed

	return nil
}
