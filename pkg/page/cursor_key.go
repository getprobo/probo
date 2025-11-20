// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
