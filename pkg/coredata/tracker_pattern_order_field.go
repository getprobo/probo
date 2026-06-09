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

type TrackerPatternOrderField string

const (
	TrackerPatternOrderFieldCreatedAt     TrackerPatternOrderField = "CREATED_AT"
	TrackerPatternOrderFieldName          TrackerPatternOrderField = "NAME"
	TrackerPatternOrderFieldLastMatchedAt TrackerPatternOrderField = "LAST_MATCHED_AT"
	TrackerPatternOrderFieldUpdatedAt     TrackerPatternOrderField = "UPDATED_AT"
	TrackerPatternOrderFieldSource        TrackerPatternOrderField = "SOURCE"
)

var (
	_ page.OrderField          = TrackerPatternOrderField("")
	_ fmt.Stringer             = TrackerPatternOrderField("")
	_ encoding.TextMarshaler   = TrackerPatternOrderField("")
	_ encoding.TextUnmarshaler = (*TrackerPatternOrderField)(nil)
)

func TrackerPatternOrderFields() []TrackerPatternOrderField {
	return []TrackerPatternOrderField{
		TrackerPatternOrderFieldCreatedAt,
		TrackerPatternOrderFieldName,
		TrackerPatternOrderFieldLastMatchedAt,
		TrackerPatternOrderFieldUpdatedAt,
		TrackerPatternOrderFieldSource,
	}
}

func (v TrackerPatternOrderField) IsValid() bool {
	switch v {
	case
		TrackerPatternOrderFieldCreatedAt,
		TrackerPatternOrderFieldName,
		TrackerPatternOrderFieldLastMatchedAt,
		TrackerPatternOrderFieldUpdatedAt,
		TrackerPatternOrderFieldSource:
		return true
	}

	return false
}

func (v TrackerPatternOrderField) String() string {
	return string(v)
}

func (v TrackerPatternOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *TrackerPatternOrderField) UnmarshalText(text []byte) error {
	val := TrackerPatternOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid TrackerPatternOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p TrackerPatternOrderField) Column() string {
	switch p {
	case TrackerPatternOrderFieldCreatedAt:
		return "created_at"
	case TrackerPatternOrderFieldName:
		return "display_name"
	case TrackerPatternOrderFieldLastMatchedAt:
		return "COALESCE(last_matched_at, '0001-01-01T00:00:00Z'::timestamptz)"
	case TrackerPatternOrderFieldUpdatedAt:
		return "updated_at"
	case TrackerPatternOrderFieldSource:
		return "COALESCE(source, '')"
	}

	panic(fmt.Sprintf("unsupported order by: %s", p))
}
