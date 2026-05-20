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

type ProcessingActivityTransferImpactAssessment string

const (
	ProcessingActivityTransferImpactAssessmentNeeded    ProcessingActivityTransferImpactAssessment = "NEEDED"
	ProcessingActivityTransferImpactAssessmentNotNeeded ProcessingActivityTransferImpactAssessment = "NOT_NEEDED"
)

var (
	_ fmt.Stringer             = ProcessingActivityTransferImpactAssessment("")
	_ encoding.TextMarshaler   = ProcessingActivityTransferImpactAssessment("")
	_ encoding.TextUnmarshaler = (*ProcessingActivityTransferImpactAssessment)(nil)
)

func ProcessingActivityTransferImpactAssessments() []ProcessingActivityTransferImpactAssessment {
	return []ProcessingActivityTransferImpactAssessment{
		ProcessingActivityTransferImpactAssessmentNeeded,
		ProcessingActivityTransferImpactAssessmentNotNeeded,
	}
}

func (v ProcessingActivityTransferImpactAssessment) IsValid() bool {
	switch v {
	case
		ProcessingActivityTransferImpactAssessmentNeeded,
		ProcessingActivityTransferImpactAssessmentNotNeeded:
		return true
	}

	return false
}

func (v ProcessingActivityTransferImpactAssessment) String() string {
	return string(v)
}

func (v ProcessingActivityTransferImpactAssessment) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *ProcessingActivityTransferImpactAssessment) UnmarshalText(text []byte) error {
	val := ProcessingActivityTransferImpactAssessment(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid ProcessingActivityTransferImpactAssessment value: %q", string(text))
	}

	*v = val

	return nil
}
