// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
)

type ProcessingActivityDataProtectionImpactAssessment string

const (
	ProcessingActivityDataProtectionImpactAssessmentNeeded    ProcessingActivityDataProtectionImpactAssessment = "NEEDED"
	ProcessingActivityDataProtectionImpactAssessmentNotNeeded ProcessingActivityDataProtectionImpactAssessment = "NOT_NEEDED"
)

var (
	_ fmt.Stringer             = ProcessingActivityDataProtectionImpactAssessment("")
	_ encoding.TextMarshaler   = ProcessingActivityDataProtectionImpactAssessment("")
	_ encoding.TextUnmarshaler = (*ProcessingActivityDataProtectionImpactAssessment)(nil)
)

func ProcessingActivityDataProtectionImpactAssessments() []ProcessingActivityDataProtectionImpactAssessment {
	return []ProcessingActivityDataProtectionImpactAssessment{
		ProcessingActivityDataProtectionImpactAssessmentNeeded,
		ProcessingActivityDataProtectionImpactAssessmentNotNeeded,
	}
}

func (v ProcessingActivityDataProtectionImpactAssessment) IsValid() bool {
	switch v {
	case
		ProcessingActivityDataProtectionImpactAssessmentNeeded,
		ProcessingActivityDataProtectionImpactAssessmentNotNeeded:
		return true
	}

	return false
}

func (v ProcessingActivityDataProtectionImpactAssessment) String() string {
	return string(v)
}

func (v ProcessingActivityDataProtectionImpactAssessment) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *ProcessingActivityDataProtectionImpactAssessment) UnmarshalText(text []byte) error {
	val := ProcessingActivityDataProtectionImpactAssessment(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid ProcessingActivityDataProtectionImpactAssessment value: %q", string(text))
	}

	*v = val

	return nil
}
