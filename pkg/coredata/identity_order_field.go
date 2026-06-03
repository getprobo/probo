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
	IdentityOrderField string
)

const (
	IdentityOrderFieldCreatedAt IdentityOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = IdentityOrderField("")
	_ fmt.Stringer             = IdentityOrderField("")
	_ encoding.TextMarshaler   = IdentityOrderField("")
	_ encoding.TextUnmarshaler = (*IdentityOrderField)(nil)
)

func IdentityOrderFields() []IdentityOrderField {
	return []IdentityOrderField{
		IdentityOrderFieldCreatedAt,
	}
}

func (v IdentityOrderField) IsValid() bool {
	switch v {
	case
		IdentityOrderFieldCreatedAt:
		return true
	}

	return false
}

func (v IdentityOrderField) String() string {
	return string(v)
}

func (v IdentityOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *IdentityOrderField) UnmarshalText(text []byte) error {
	val := IdentityOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid IdentityOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p IdentityOrderField) Column() string {
	return string(p)
}
