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

type (
	RiskTreatment string
)

const (
	RiskTreatmentMitigated   RiskTreatment = "MITIGATED"
	RiskTreatmentAccepted    RiskTreatment = "ACCEPTED"
	RiskTreatmentAvoided     RiskTreatment = "AVOIDED"
	RiskTreatmentTransferred RiskTreatment = "TRANSFERRED"
)

func RiskTreatments() []RiskTreatment {
	return []RiskTreatment{
		RiskTreatmentMitigated,
		RiskTreatmentAccepted,
		RiskTreatmentAvoided,
		RiskTreatmentTransferred,
	}
}

func (rt RiskTreatment) MarshalText() ([]byte, error) {
	return []byte(rt.String()), nil
}

func (rt *RiskTreatment) UnmarshalText(data []byte) error {
	val := string(data)

	switch val {
	case RiskTreatmentMitigated.String():
		*rt = RiskTreatmentMitigated
	case RiskTreatmentAccepted.String():
		*rt = RiskTreatmentAccepted
	case RiskTreatmentAvoided.String():
		*rt = RiskTreatmentAvoided
	case RiskTreatmentTransferred.String():
		*rt = RiskTreatmentTransferred
	default:
		return fmt.Errorf("invalid RiskTreatment value: %q", val)
	}

	return nil
}

func (rt RiskTreatment) String() string {
	var val string

	switch rt {
	case RiskTreatmentMitigated:
		val = "MITIGATED"
	case RiskTreatmentAccepted:
		val = "ACCEPTED"
	case RiskTreatmentAvoided:
		val = "AVOIDED"
	case RiskTreatmentTransferred:
		val = "TRANSFERRED"
	}

	return val
}

func (rt *RiskTreatment) Scan(value any) error {
	val, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid scan source for RiskTreatment, expected string got %T", value)
	}

	return rt.UnmarshalText([]byte(val))
}

func (rt RiskTreatment) Value() (driver.Value, error) {
	return rt.String(), nil
}
