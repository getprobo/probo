// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	SCIMConfigurationOrderField string
)

const (
	SCIMConfigurationOrderFieldCreatedAt SCIMConfigurationOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = SCIMConfigurationOrderField("")
	_ fmt.Stringer             = SCIMConfigurationOrderField("")
	_ encoding.TextMarshaler   = SCIMConfigurationOrderField("")
	_ encoding.TextUnmarshaler = (*SCIMConfigurationOrderField)(nil)
)

func SCIMConfigurationOrderFields() []SCIMConfigurationOrderField {
	return []SCIMConfigurationOrderField{
		SCIMConfigurationOrderFieldCreatedAt,
	}
}

func (v SCIMConfigurationOrderField) IsValid() bool {
	switch v {
	case
		SCIMConfigurationOrderFieldCreatedAt:
		return true
	}

	return false
}

func (v SCIMConfigurationOrderField) String() string {
	return string(v)
}

func (v SCIMConfigurationOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *SCIMConfigurationOrderField) UnmarshalText(text []byte) error {
	val := SCIMConfigurationOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid SCIMConfigurationOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p SCIMConfigurationOrderField) Column() string {
	return string(p)
}
