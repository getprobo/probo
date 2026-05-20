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
	ThirdPartyContactOrderField string
)

const (
	ThirdPartyContactOrderFieldCreatedAt ThirdPartyContactOrderField = "CREATED_AT"
	ThirdPartyContactOrderFieldFullName  ThirdPartyContactOrderField = "FULL_NAME"
	ThirdPartyContactOrderFieldEmail     ThirdPartyContactOrderField = "EMAIL"
)

var (
	_ page.OrderField          = ThirdPartyContactOrderField("")
	_ fmt.Stringer             = ThirdPartyContactOrderField("")
	_ encoding.TextMarshaler   = ThirdPartyContactOrderField("")
	_ encoding.TextUnmarshaler = (*ThirdPartyContactOrderField)(nil)
)

func ThirdPartyContactOrderFields() []ThirdPartyContactOrderField {
	return []ThirdPartyContactOrderField{
		ThirdPartyContactOrderFieldCreatedAt,
		ThirdPartyContactOrderFieldFullName,
		ThirdPartyContactOrderFieldEmail,
	}
}

func (v ThirdPartyContactOrderField) IsValid() bool {
	switch v {
	case
		ThirdPartyContactOrderFieldCreatedAt,
		ThirdPartyContactOrderFieldFullName,
		ThirdPartyContactOrderFieldEmail:
		return true
	}

	return false
}

func (v ThirdPartyContactOrderField) String() string {
	return string(v)
}

func (v ThirdPartyContactOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *ThirdPartyContactOrderField) UnmarshalText(text []byte) error {
	val := ThirdPartyContactOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid ThirdPartyContactOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p ThirdPartyContactOrderField) Column() string {
	return string(p)
}
