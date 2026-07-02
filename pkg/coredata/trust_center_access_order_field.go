// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

type TrustCenterAccessOrderField string

const (
	TrustCenterAccessOrderFieldCreatedAt TrustCenterAccessOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = TrustCenterAccessOrderField("")
	_ fmt.Stringer             = TrustCenterAccessOrderField("")
	_ encoding.TextMarshaler   = TrustCenterAccessOrderField("")
	_ encoding.TextUnmarshaler = (*TrustCenterAccessOrderField)(nil)
)

func TrustCenterAccessOrderFields() []TrustCenterAccessOrderField {
	return []TrustCenterAccessOrderField{
		TrustCenterAccessOrderFieldCreatedAt,
	}
}

func (v TrustCenterAccessOrderField) IsValid() bool {
	switch v {
	case
		TrustCenterAccessOrderFieldCreatedAt:
		return true
	}

	return false
}

func (v TrustCenterAccessOrderField) String() string {
	return string(v)
}

func (v TrustCenterAccessOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *TrustCenterAccessOrderField) UnmarshalText(text []byte) error {
	val := TrustCenterAccessOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid TrustCenterAccessOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (tcaof TrustCenterAccessOrderField) Column() string {
	switch tcaof {
	case TrustCenterAccessOrderFieldCreatedAt:
		return "created_at"
	}

	panic(fmt.Sprintf("unsupported order by: %s", tcaof))
}
