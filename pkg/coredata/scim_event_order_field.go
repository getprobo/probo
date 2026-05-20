// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
	SCIMEventOrderField string
)

const (
	SCIMEventOrderFieldCreatedAt SCIMEventOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = SCIMEventOrderField("")
	_ fmt.Stringer             = SCIMEventOrderField("")
	_ encoding.TextMarshaler   = SCIMEventOrderField("")
	_ encoding.TextUnmarshaler = (*SCIMEventOrderField)(nil)
)

func SCIMEventOrderFields() []SCIMEventOrderField {
	return []SCIMEventOrderField{
		SCIMEventOrderFieldCreatedAt,
	}
}

func (v SCIMEventOrderField) IsValid() bool {
	switch v {
	case
		SCIMEventOrderFieldCreatedAt:
		return true
	}

	return false
}

func (v SCIMEventOrderField) String() string {
	return string(v)
}

func (v SCIMEventOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *SCIMEventOrderField) UnmarshalText(text []byte) error {
	val := SCIMEventOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid SCIMEventOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p SCIMEventOrderField) Column() string {
	return string(p)
}
