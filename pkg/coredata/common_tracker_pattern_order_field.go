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

type CommonTrackerPatternOrderField string

const (
	CommonTrackerPatternOrderFieldPattern    CommonTrackerPatternOrderField = "PATTERN"
	CommonTrackerPatternOrderFieldConfidence CommonTrackerPatternOrderField = "CONFIDENCE"
	CommonTrackerPatternOrderFieldCreatedAt  CommonTrackerPatternOrderField = "CREATED_AT"
	CommonTrackerPatternOrderFieldUpdatedAt  CommonTrackerPatternOrderField = "UPDATED_AT"
	CommonTrackerPatternOrderFieldEnrichedAt CommonTrackerPatternOrderField = "ENRICHED_AT"
)

var (
	_ page.OrderField          = CommonTrackerPatternOrderField("")
	_ fmt.Stringer             = CommonTrackerPatternOrderField("")
	_ encoding.TextMarshaler   = CommonTrackerPatternOrderField("")
	_ encoding.TextUnmarshaler = (*CommonTrackerPatternOrderField)(nil)
)

func CommonTrackerPatternOrderFields() []CommonTrackerPatternOrderField {
	return []CommonTrackerPatternOrderField{
		CommonTrackerPatternOrderFieldPattern,
		CommonTrackerPatternOrderFieldConfidence,
		CommonTrackerPatternOrderFieldCreatedAt,
		CommonTrackerPatternOrderFieldUpdatedAt,
		CommonTrackerPatternOrderFieldEnrichedAt,
	}
}

func (v CommonTrackerPatternOrderField) IsValid() bool {
	switch v {
	case
		CommonTrackerPatternOrderFieldPattern,
		CommonTrackerPatternOrderFieldConfidence,
		CommonTrackerPatternOrderFieldCreatedAt,
		CommonTrackerPatternOrderFieldUpdatedAt,
		CommonTrackerPatternOrderFieldEnrichedAt:
		return true
	}

	return false
}

func (v CommonTrackerPatternOrderField) String() string {
	return string(v)
}

func (v CommonTrackerPatternOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *CommonTrackerPatternOrderField) UnmarshalText(text []byte) error {
	val := CommonTrackerPatternOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid CommonTrackerPatternOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (v CommonTrackerPatternOrderField) Column() string {
	switch v {
	case CommonTrackerPatternOrderFieldPattern:
		return "pattern"
	case CommonTrackerPatternOrderFieldConfidence:
		return "confidence"
	case CommonTrackerPatternOrderFieldCreatedAt:
		return "created_at"
	case CommonTrackerPatternOrderFieldUpdatedAt:
		return "updated_at"
	case CommonTrackerPatternOrderFieldEnrichedAt:
		return "COALESCE(enriched_at, '0001-01-01T00:00:00Z'::timestamptz)"
	}

	panic(fmt.Sprintf("unsupported order by: %s", v))
}
