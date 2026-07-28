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
	"database/sql/driver"
	"encoding"
	"fmt"
)

type (
	EvidenceAssessmentStatus string
)

const (
	EvidenceAssessmentStatusPending    EvidenceAssessmentStatus = "PENDING"
	EvidenceAssessmentStatusProcessing EvidenceAssessmentStatus = "PROCESSING"
	EvidenceAssessmentStatusCompleted  EvidenceAssessmentStatus = "COMPLETED"
	EvidenceAssessmentStatusFailed     EvidenceAssessmentStatus = "FAILED"
)

var (
	_ fmt.Stringer             = EvidenceAssessmentStatus("")
	_ encoding.TextMarshaler   = EvidenceAssessmentStatus("")
	_ encoding.TextUnmarshaler = (*EvidenceAssessmentStatus)(nil)
	_ driver.Valuer            = EvidenceAssessmentStatus("")
)

func EvidenceAssessmentStatuses() []EvidenceAssessmentStatus {
	return []EvidenceAssessmentStatus{
		EvidenceAssessmentStatusPending,
		EvidenceAssessmentStatusProcessing,
		EvidenceAssessmentStatusCompleted,
		EvidenceAssessmentStatusFailed,
	}
}

func (v EvidenceAssessmentStatus) IsValid() bool {
	switch v {
	case EvidenceAssessmentStatusPending,
		EvidenceAssessmentStatusProcessing,
		EvidenceAssessmentStatusCompleted,
		EvidenceAssessmentStatusFailed:
		return true
	}

	return false
}

func (v EvidenceAssessmentStatus) String() string {
	return string(v)
}

func (v EvidenceAssessmentStatus) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *EvidenceAssessmentStatus) UnmarshalText(text []byte) error {
	val := EvidenceAssessmentStatus(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid EvidenceAssessmentStatus value: %q", string(text))
	}

	*v = val

	return nil
}

func (v *EvidenceAssessmentStatus) Scan(value any) error {
	switch val := value.(type) {
	case string:
		return v.UnmarshalText([]byte(val))
	case []byte:
		return v.UnmarshalText(val)
	default:
		return fmt.Errorf("invalid scan source for EvidenceAssessmentStatus, expected string or []byte got %T", value)
	}
}

func (v EvidenceAssessmentStatus) Value() (driver.Value, error) {
	return v.String(), nil
}
