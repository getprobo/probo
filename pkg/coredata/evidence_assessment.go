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

import (
	"encoding/json"
	"fmt"
)

// SetAssessment marshals a typed assessment into Evidence.Assessment.
// Passing nil clears the field. The shape of the assessment is defined
// by its producer (see pkg/evidencedescriber.EvidenceAssessment); this
// package is intentionally agnostic about the schema and only owns the
// raw JSONB round-trip.
func (e *Evidence) SetAssessment(v any) error {
	if v == nil {
		e.Assessment = nil
		return nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("cannot marshal evidence assessment: %w", err)
	}
	e.Assessment = data
	return nil
}

// AssessmentInto unmarshals Evidence.Assessment into dst. It is a no-op
// when the column is NULL/empty, leaving dst untouched.
func (e *Evidence) AssessmentInto(dst any) error {
	if len(e.Assessment) == 0 {
		return nil
	}
	if err := json.Unmarshal(e.Assessment, dst); err != nil {
		return fmt.Errorf("cannot unmarshal evidence assessment: %w", err)
	}
	return nil
}
