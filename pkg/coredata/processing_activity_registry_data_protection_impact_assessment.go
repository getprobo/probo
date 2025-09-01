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

type ProcessingActivityRegistryDataProtectionImpactAssessment string

const (
	ProcessingActivityRegistryDataProtectionImpactAssessmentNeeded    ProcessingActivityRegistryDataProtectionImpactAssessment = "NEEDED"
	ProcessingActivityRegistryDataProtectionImpactAssessmentNotNeeded ProcessingActivityRegistryDataProtectionImpactAssessment = "NOT_NEEDED"
)

func (p ProcessingActivityRegistryDataProtectionImpactAssessment) String() string {
	return string(p)
}

func (p *ProcessingActivityRegistryDataProtectionImpactAssessment) Scan(value any) error {
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("unsupported type for ProcessingActivityRegistryDataProtectionImpactAssessment: %T", value)
	}

	switch s {
	case "NEEDED":
		*p = ProcessingActivityRegistryDataProtectionImpactAssessmentNeeded
	case "NOT_NEEDED":
		*p = ProcessingActivityRegistryDataProtectionImpactAssessmentNotNeeded
	default:
		return fmt.Errorf("invalid ProcessingActivityRegistryDataProtectionImpactAssessment value: %q", s)
	}
	return nil
}

func (p ProcessingActivityRegistryDataProtectionImpactAssessment) Value() (driver.Value, error) {
	return p.String(), nil
}
