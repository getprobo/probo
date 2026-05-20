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

package coredata

import (
	"encoding"
	"fmt"

	"go.probo.inc/probo/pkg/page"
)

type (
	FrameworkOrderField string
)

const (
	FrameworkOrderFieldCreatedAt FrameworkOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = FrameworkOrderField("")
	_ fmt.Stringer             = FrameworkOrderField("")
	_ encoding.TextMarshaler   = FrameworkOrderField("")
	_ encoding.TextUnmarshaler = (*FrameworkOrderField)(nil)
)

func FrameworkOrderFields() []FrameworkOrderField {
	return []FrameworkOrderField{
		FrameworkOrderFieldCreatedAt,
	}
}

func (v FrameworkOrderField) IsValid() bool {
	switch v {
	case
		FrameworkOrderFieldCreatedAt:
		return true
	}

	return false
}

func (v FrameworkOrderField) String() string {
	return string(v)
}

func (v FrameworkOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *FrameworkOrderField) UnmarshalText(text []byte) error {
	val := FrameworkOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid FrameworkOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p FrameworkOrderField) Column() string {
	return string(p)
}
