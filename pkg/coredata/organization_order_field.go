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
	OrganizationOrderField string
)

const (
	OrganizationOrderFieldName      OrganizationOrderField = "NAME"
	OrganizationOrderFieldCreatedAt OrganizationOrderField = "CREATED_AT"
	OrganizationOrderFieldUpdatedAt OrganizationOrderField = "UPDATED_AT"
)

var (
	_ page.OrderField          = OrganizationOrderField("")
	_ fmt.Stringer             = OrganizationOrderField("")
	_ encoding.TextMarshaler   = OrganizationOrderField("")
	_ encoding.TextUnmarshaler = (*OrganizationOrderField)(nil)
)

func OrganizationOrderFields() []OrganizationOrderField {
	return []OrganizationOrderField{
		OrganizationOrderFieldName,
		OrganizationOrderFieldCreatedAt,
		OrganizationOrderFieldUpdatedAt,
	}
}

func (v OrganizationOrderField) IsValid() bool {
	switch v {
	case
		OrganizationOrderFieldName,
		OrganizationOrderFieldCreatedAt,
		OrganizationOrderFieldUpdatedAt:
		return true
	}

	return false
}

func (v OrganizationOrderField) String() string {
	return string(v)
}

func (v OrganizationOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *OrganizationOrderField) UnmarshalText(text []byte) error {
	val := OrganizationOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid OrganizationOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p OrganizationOrderField) Column() string {
	return string(p)
}
