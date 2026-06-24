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

	"go.probo.inc/probo/pkg/gid"
)

type CursorKey struct {
	ID    gid.GID
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

	if len(arr) != 2 {
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

	var value any
	if err := json.Unmarshal(arr[1], &value); err != nil {
		return CursorKeyNil, ErrInvalidFormat
	}

	return CursorKey{
		ID:    id,
		Value: value,
	}, nil
}

func NewCursorKey(id gid.GID, value any) CursorKey {
	return CursorKey{
		ID:    id,
		Value: value,
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
	arr := []any{ck.ID.String(), ck.Value}
	return json.Marshal(arr)
}

func (ck *CursorKey) UnmarshalBinary(data []byte) error {
	var arr []json.RawMessage
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}

	if len(arr) != 2 {
		return ErrInvalidFormat
	}

	// Parse the ID
	var idStr string
	if err := json.Unmarshal(arr[0], &idStr); err != nil {
		return ErrInvalidFormat
	}

	id, err := gid.ParseGID(idStr)
	if err != nil {
		return ErrInvalidFormat
	}

	var value any
	if err := json.Unmarshal(arr[1], &value); err != nil {
		return ErrInvalidFormat
	}

	ck.ID = id
	ck.Value = value

	return nil
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
