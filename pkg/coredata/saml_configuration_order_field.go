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

type (
	SAMLConfigurationOrderField string
)

const (
	SAMLConfigurationOrderFieldCreatedAt SAMLConfigurationOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = SAMLConfigurationOrderField("")
	_ fmt.Stringer             = SAMLConfigurationOrderField("")
	_ encoding.TextMarshaler   = SAMLConfigurationOrderField("")
	_ encoding.TextUnmarshaler = (*SAMLConfigurationOrderField)(nil)
)

func SAMLConfigurationOrderFields() []SAMLConfigurationOrderField {
	return []SAMLConfigurationOrderField{
		SAMLConfigurationOrderFieldCreatedAt,
	}
}

func (v SAMLConfigurationOrderField) IsValid() bool {
	switch v {
	case
		SAMLConfigurationOrderFieldCreatedAt:
		return true
	}

	return false
}

func (v SAMLConfigurationOrderField) String() string {
	return string(v)
}

func (v SAMLConfigurationOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *SAMLConfigurationOrderField) UnmarshalText(text []byte) error {
	val := SAMLConfigurationOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid SAMLConfigurationOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p SAMLConfigurationOrderField) Column() string {
	return string(p)
}
