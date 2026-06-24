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

type DetectedTrackerOrderField string

const (
	DetectedTrackerOrderFieldInitiatorURL   DetectedTrackerOrderField = "INITIATOR_URL"
	DetectedTrackerOrderFieldLastDetectedAt DetectedTrackerOrderField = "LAST_DETECTED_AT"
)

var (
	_ page.OrderField          = DetectedTrackerOrderField("")
	_ fmt.Stringer             = DetectedTrackerOrderField("")
	_ encoding.TextMarshaler   = DetectedTrackerOrderField("")
	_ encoding.TextUnmarshaler = (*DetectedTrackerOrderField)(nil)
)

func DetectedTrackerOrderFields() []DetectedTrackerOrderField {
	return []DetectedTrackerOrderField{
		DetectedTrackerOrderFieldInitiatorURL,
		DetectedTrackerOrderFieldLastDetectedAt,
	}
}

func (v DetectedTrackerOrderField) IsValid() bool {
	switch v {
	case
		DetectedTrackerOrderFieldInitiatorURL,
		DetectedTrackerOrderFieldLastDetectedAt:
		return true
	}

	return false
}

func (v DetectedTrackerOrderField) String() string {
	return string(v)
}

func (v DetectedTrackerOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *DetectedTrackerOrderField) UnmarshalText(text []byte) error {
	val := DetectedTrackerOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid DetectedTrackerOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p DetectedTrackerOrderField) Column() string {
	switch p {
	case DetectedTrackerOrderFieldInitiatorURL:
		return "COALESCE(initiator_url, '')"
	case DetectedTrackerOrderFieldLastDetectedAt:
		return "last_detected_at"
	}

	panic(fmt.Sprintf("unsupported order by: %s", p))
}
