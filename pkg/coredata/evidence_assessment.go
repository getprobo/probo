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
	"encoding/json"
	"fmt"
)

// SetAssessment marshals a typed assessment into Evidence.Assessment.
// Passing nil — or any value that marshals to JSON null, including a
// typed-nil pointer — clears the field rather than persisting JSON
// null. The shape of the assessment is defined by its producer (see
// pkg/evidenceassessor.EvidenceAssessment); this package is
// intentionally agnostic about the schema and only owns the raw JSONB
// round-trip.
func (e *Evidence) SetAssessment(v any) error {
	if v == nil {
		e.Assessment = nil
		return nil
	}

	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("cannot marshal evidence assessment: %w", err)
	}

	if string(data) == "null" {
		e.Assessment = nil
		return nil
	}

	e.Assessment = data

	return nil
}

// GetAssessment unmarshals Evidence.Assessment into dst. It is a no-op
// when the column is NULL/empty, leaving dst untouched.
func (e *Evidence) GetAssessment(dst any) error {
	if len(e.Assessment) == 0 {
		return nil
	}

	if err := json.Unmarshal(e.Assessment, dst); err != nil {
		return fmt.Errorf("cannot unmarshal evidence assessment: %w", err)
	}

	return nil
}
