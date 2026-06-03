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

type CommonThirdPartyOrderField string

const (
	CommonThirdPartyOrderFieldName      CommonThirdPartyOrderField = "NAME"
	CommonThirdPartyOrderFieldCreatedAt CommonThirdPartyOrderField = "CREATED_AT"
	CommonThirdPartyOrderFieldUpdatedAt CommonThirdPartyOrderField = "UPDATED_AT"
)

var (
	_ page.OrderField          = CommonThirdPartyOrderField("")
	_ fmt.Stringer             = CommonThirdPartyOrderField("")
	_ encoding.TextMarshaler   = CommonThirdPartyOrderField("")
	_ encoding.TextUnmarshaler = (*CommonThirdPartyOrderField)(nil)
)

func CommonThirdPartyOrderFields() []CommonThirdPartyOrderField {
	return []CommonThirdPartyOrderField{
		CommonThirdPartyOrderFieldName,
		CommonThirdPartyOrderFieldCreatedAt,
		CommonThirdPartyOrderFieldUpdatedAt,
	}
}

func (v CommonThirdPartyOrderField) IsValid() bool {
	switch v {
	case
		CommonThirdPartyOrderFieldName,
		CommonThirdPartyOrderFieldCreatedAt,
		CommonThirdPartyOrderFieldUpdatedAt:
		return true
	}

	return false
}

func (v CommonThirdPartyOrderField) String() string {
	return string(v)
}

func (v CommonThirdPartyOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *CommonThirdPartyOrderField) UnmarshalText(text []byte) error {
	val := CommonThirdPartyOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid CommonThirdPartyOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (v CommonThirdPartyOrderField) Column() string {
	switch v {
	case CommonThirdPartyOrderFieldName:
		return "name"
	case CommonThirdPartyOrderFieldCreatedAt:
		return "created_at"
	case CommonThirdPartyOrderFieldUpdatedAt:
		return "updated_at"
	}

	panic(fmt.Sprintf("unsupported order by: %s", v))
}
