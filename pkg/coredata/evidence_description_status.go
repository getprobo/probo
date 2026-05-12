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
	"database/sql/driver"
	"fmt"
)

type (
	EvidenceDescriptionStatus string
)

const (
	EvidenceDescriptionStatusPending    EvidenceDescriptionStatus = "PENDING"
	EvidenceDescriptionStatusProcessing EvidenceDescriptionStatus = "PROCESSING"
	EvidenceDescriptionStatusCompleted  EvidenceDescriptionStatus = "COMPLETED"
	EvidenceDescriptionStatusFailed     EvidenceDescriptionStatus = "FAILED"
)

func (s EvidenceDescriptionStatus) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *EvidenceDescriptionStatus) UnmarshalText(data []byte) error {
	val := string(data)

	switch val {
	case EvidenceDescriptionStatusPending.String():
		*s = EvidenceDescriptionStatusPending
	case EvidenceDescriptionStatusProcessing.String():
		*s = EvidenceDescriptionStatusProcessing
	case EvidenceDescriptionStatusCompleted.String():
		*s = EvidenceDescriptionStatusCompleted
	case EvidenceDescriptionStatusFailed.String():
		*s = EvidenceDescriptionStatusFailed
	default:
		return fmt.Errorf("invalid EvidenceDescriptionStatus value: %q", val)
	}

	return nil
}

func (s EvidenceDescriptionStatus) String() string {
	return string(s)
}

func (s *EvidenceDescriptionStatus) Scan(value any) error {
	val, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid scan source for EvidenceDescriptionStatus, expected string got %T", value)
	}

	return s.UnmarshalText([]byte(val))
}

func (s EvidenceDescriptionStatus) Value() (driver.Value, error) {
	return s.String(), nil
}
