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

type CommonTrackerPatternOrderField string

const (
	CommonTrackerPatternOrderFieldPattern                 CommonTrackerPatternOrderField = "PATTERN"
	CommonTrackerPatternOrderFieldConfidence              CommonTrackerPatternOrderField = "CONFIDENCE"
	CommonTrackerPatternOrderFieldCreatedAt               CommonTrackerPatternOrderField = "CREATED_AT"
	CommonTrackerPatternOrderFieldUpdatedAt               CommonTrackerPatternOrderField = "UPDATED_AT"
	CommonTrackerPatternOrderFieldLastEnrichmentAttemptAt CommonTrackerPatternOrderField = "LAST_ENRICHMENT_ATTEMPT_AT"
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
		CommonTrackerPatternOrderFieldLastEnrichmentAttemptAt,
	}
}

func (v CommonTrackerPatternOrderField) IsValid() bool {
	switch v {
	case
		CommonTrackerPatternOrderFieldPattern,
		CommonTrackerPatternOrderFieldConfidence,
		CommonTrackerPatternOrderFieldCreatedAt,
		CommonTrackerPatternOrderFieldUpdatedAt,
		CommonTrackerPatternOrderFieldLastEnrichmentAttemptAt:
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
	case CommonTrackerPatternOrderFieldLastEnrichmentAttemptAt:
		return "COALESCE(last_enrichment_attempt_at, '0001-01-01T00:00:00Z'::timestamptz)"
	}

	panic(fmt.Sprintf("unsupported order by: %s", v))
}
