// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
