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

type TrackerResourceOrderField string

const (
	TrackerResourceOrderFieldCreatedAt      TrackerResourceOrderField = "CREATED_AT"
	TrackerResourceOrderFieldLastDetectedAt TrackerResourceOrderField = "LAST_DETECTED_AT"
	TrackerResourceOrderFieldOrigin         TrackerResourceOrderField = "ORIGIN"
	TrackerResourceOrderFieldUpdatedAt      TrackerResourceOrderField = "UPDATED_AT"
)

var (
	_ page.OrderField          = TrackerResourceOrderField("")
	_ fmt.Stringer             = TrackerResourceOrderField("")
	_ encoding.TextMarshaler   = TrackerResourceOrderField("")
	_ encoding.TextUnmarshaler = (*TrackerResourceOrderField)(nil)
)

func TrackerResourceOrderFields() []TrackerResourceOrderField {
	return []TrackerResourceOrderField{
		TrackerResourceOrderFieldCreatedAt,
		TrackerResourceOrderFieldLastDetectedAt,
		TrackerResourceOrderFieldOrigin,
		TrackerResourceOrderFieldUpdatedAt,
	}
}

func (v TrackerResourceOrderField) IsValid() bool {
	switch v {
	case
		TrackerResourceOrderFieldCreatedAt,
		TrackerResourceOrderFieldLastDetectedAt,
		TrackerResourceOrderFieldOrigin,
		TrackerResourceOrderFieldUpdatedAt:
		return true
	}

	return false
}

func (v TrackerResourceOrderField) String() string {
	return string(v)
}

func (v TrackerResourceOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *TrackerResourceOrderField) UnmarshalText(text []byte) error {
	val := TrackerResourceOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid TrackerResourceOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p TrackerResourceOrderField) Column() string {
	switch p {
	case TrackerResourceOrderFieldCreatedAt:
		return "created_at"
	case TrackerResourceOrderFieldLastDetectedAt:
		return "COALESCE(last_detected_at, '0001-01-01T00:00:00Z'::timestamptz)"
	case TrackerResourceOrderFieldOrigin:
		return "origin"
	case TrackerResourceOrderFieldUpdatedAt:
		return "updated_at"
	}

	panic(fmt.Sprintf("unsupported order by: %s", p))
}
