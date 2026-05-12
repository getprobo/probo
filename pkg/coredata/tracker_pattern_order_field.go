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

type TrackerPatternOrderField string

const (
	TrackerPatternOrderFieldCreatedAt     TrackerPatternOrderField = "CREATED_AT"
	TrackerPatternOrderFieldName          TrackerPatternOrderField = "NAME"
	TrackerPatternOrderFieldLastMatchedAt TrackerPatternOrderField = "LAST_MATCHED_AT"
	TrackerPatternOrderFieldUpdatedAt     TrackerPatternOrderField = "UPDATED_AT"
	TrackerPatternOrderFieldSource        TrackerPatternOrderField = "SOURCE"
)

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

func (p TrackerPatternOrderField) IsValid() bool {
	switch p {
	case TrackerPatternOrderFieldCreatedAt,
		TrackerPatternOrderFieldName,
		TrackerPatternOrderFieldLastMatchedAt,
		TrackerPatternOrderFieldUpdatedAt,
		TrackerPatternOrderFieldSource:
		return true
	}
	return false
}

func (p TrackerPatternOrderField) String() string {
	return string(p)
}

func (p *TrackerPatternOrderField) UnmarshalText(text []byte) error {
	*p = TrackerPatternOrderField(text)
	if !p.IsValid() {
		return fmt.Errorf("%s is not a valid TrackerPatternOrderField", string(text))
	}
	return nil
}

func (p TrackerPatternOrderField) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}
