// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
	"database/sql/driver"
	"fmt"
)

type ProcessingActivityTransferImpactAssessment string

const (
	ProcessingActivityTransferImpactAssessmentNeeded    ProcessingActivityTransferImpactAssessment = "NEEDED"
	ProcessingActivityTransferImpactAssessmentNotNeeded ProcessingActivityTransferImpactAssessment = "NOT_NEEDED"
)

func ProcessingActivityTransferImpactAssessments() []ProcessingActivityTransferImpactAssessment {
	return []ProcessingActivityTransferImpactAssessment{
		ProcessingActivityTransferImpactAssessmentNeeded,
		ProcessingActivityTransferImpactAssessmentNotNeeded,
	}
}

func (p ProcessingActivityTransferImpactAssessment) String() string {
	return string(p)
}

func (p *ProcessingActivityTransferImpactAssessment) Scan(value any) error {
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("unsupported type for ProcessingActivityTransferImpactAssessment: %T", value)
	}

	switch s {
	case "NEEDED":
		*p = ProcessingActivityTransferImpactAssessmentNeeded
	case "NOT_NEEDED":
		*p = ProcessingActivityTransferImpactAssessmentNotNeeded
	default:
		return fmt.Errorf("invalid ProcessingActivityTransferImpactAssessment value: %q", s)
	}
	return nil
}

func (p ProcessingActivityTransferImpactAssessment) Value() (driver.Value, error) {
	return p.String(), nil
}
