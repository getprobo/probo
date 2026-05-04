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

import "fmt"

type CookiePatternOrderField string

const (
	CookiePatternOrderFieldCreatedAt     CookiePatternOrderField = "CREATED_AT"
	CookiePatternOrderFieldName          CookiePatternOrderField = "NAME"
	CookiePatternOrderFieldLastMatchedAt CookiePatternOrderField = "LAST_MATCHED_AT"
	CookiePatternOrderFieldUpdatedAt     CookiePatternOrderField = "UPDATED_AT"
	CookiePatternOrderFieldSource        CookiePatternOrderField = "SOURCE"
)

func (p CookiePatternOrderField) Column() string {
	switch p {
	case CookiePatternOrderFieldCreatedAt:
		return "created_at"
	case CookiePatternOrderFieldName:
		return "display_name"
	case CookiePatternOrderFieldLastMatchedAt:
		return "COALESCE(last_matched_at, '0001-01-01T00:00:00Z'::timestamptz)"
	case CookiePatternOrderFieldUpdatedAt:
		return "updated_at"
	case CookiePatternOrderFieldSource:
		return "source"
	}
	panic(fmt.Sprintf("unsupported order by: %s", p))
}

func (p CookiePatternOrderField) IsValid() bool {
	switch p {
	case CookiePatternOrderFieldCreatedAt,
		CookiePatternOrderFieldName,
		CookiePatternOrderFieldLastMatchedAt,
		CookiePatternOrderFieldUpdatedAt,
		CookiePatternOrderFieldSource:
		return true
	}
	return false
}

func (p CookiePatternOrderField) String() string {
	return string(p)
}

func (p *CookiePatternOrderField) UnmarshalText(text []byte) error {
	*p = CookiePatternOrderField(text)
	if !p.IsValid() {
		return fmt.Errorf("%s is not a valid CookiePatternOrderField", string(text))
	}
	return nil
}

func (p CookiePatternOrderField) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}
