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
	ElectronicSignatureEventType string
)

const (
	ElectronicSignatureEventTypeDocumentViewed       ElectronicSignatureEventType = "DOCUMENT_VIEWED"
	ElectronicSignatureEventTypeConsentGiven         ElectronicSignatureEventType = "CONSENT_GIVEN"
	ElectronicSignatureEventTypeFullNameTyped        ElectronicSignatureEventType = "FULL_NAME_TYPED"
	ElectronicSignatureEventTypeSignatureAccepted    ElectronicSignatureEventType = "SIGNATURE_ACCEPTED"
	ElectronicSignatureEventTypeSignatureCompleted   ElectronicSignatureEventType = "SIGNATURE_COMPLETED"
	ElectronicSignatureEventTypeSealComputed         ElectronicSignatureEventType = "SEAL_COMPUTED"
	ElectronicSignatureEventTypeTimestampRequested   ElectronicSignatureEventType = "TIMESTAMP_REQUESTED"
	ElectronicSignatureEventTypeCertificateGenerated ElectronicSignatureEventType = "CERTIFICATE_GENERATED"
	ElectronicSignatureEventTypeProcessingError      ElectronicSignatureEventType = "PROCESSING_ERROR"
)

func (t ElectronicSignatureEventType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t *ElectronicSignatureEventType) UnmarshalText(data []byte) error {
	val := string(data)

	switch val {
	case ElectronicSignatureEventTypeDocumentViewed.String():
		*t = ElectronicSignatureEventTypeDocumentViewed
	case ElectronicSignatureEventTypeConsentGiven.String():
		*t = ElectronicSignatureEventTypeConsentGiven
	case ElectronicSignatureEventTypeFullNameTyped.String():
		*t = ElectronicSignatureEventTypeFullNameTyped
	case ElectronicSignatureEventTypeSignatureAccepted.String():
		*t = ElectronicSignatureEventTypeSignatureAccepted
	case ElectronicSignatureEventTypeSignatureCompleted.String():
		*t = ElectronicSignatureEventTypeSignatureCompleted
	case ElectronicSignatureEventTypeSealComputed.String():
		*t = ElectronicSignatureEventTypeSealComputed
	case ElectronicSignatureEventTypeTimestampRequested.String():
		*t = ElectronicSignatureEventTypeTimestampRequested
	case ElectronicSignatureEventTypeCertificateGenerated.String():
		*t = ElectronicSignatureEventTypeCertificateGenerated
	case ElectronicSignatureEventTypeProcessingError.String():
		*t = ElectronicSignatureEventTypeProcessingError
	default:
		return fmt.Errorf("invalid ElectronicSignatureEventType value: %q", val)
	}

	return nil
}

func (t ElectronicSignatureEventType) String() string {
	return string(t)
}

func (t *ElectronicSignatureEventType) Scan(value any) error {
	val, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid scan source for ElectronicSignatureEventType, expected string got %T", value)
	}

	return t.UnmarshalText([]byte(val))
}

func (t ElectronicSignatureEventType) Value() (driver.Value, error) {
	return t.String(), nil
}
