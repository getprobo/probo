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

type CookieCategoryOrderField string

const (
	CookieCategoryOrderFieldRank CookieCategoryOrderField = "RANK"
)

var (
	_ page.OrderField          = CookieCategoryOrderField("")
	_ fmt.Stringer             = CookieCategoryOrderField("")
	_ encoding.TextMarshaler   = CookieCategoryOrderField("")
	_ encoding.TextUnmarshaler = (*CookieCategoryOrderField)(nil)
)

func CookieCategoryOrderFields() []CookieCategoryOrderField {
	return []CookieCategoryOrderField{
		CookieCategoryOrderFieldRank,
	}
}

func (v CookieCategoryOrderField) IsValid() bool {
	switch v {
	case
		CookieCategoryOrderFieldRank:
		return true
	}

	return false
}

func (v CookieCategoryOrderField) String() string {
	return string(v)
}

func (v CookieCategoryOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *CookieCategoryOrderField) UnmarshalText(text []byte) error {
	val := CookieCategoryOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid CookieCategoryOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p CookieCategoryOrderField) Column() string {
	switch p {
	case CookieCategoryOrderFieldRank:
		return "rank"
	}

	panic(fmt.Sprintf("unsupported order by: %s", p))
}
