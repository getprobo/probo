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
	"encoding"
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

var (
	_ fmt.Stringer             = ElectronicSignatureEventType("")
	_ encoding.TextMarshaler   = ElectronicSignatureEventType("")
	_ encoding.TextUnmarshaler = (*ElectronicSignatureEventType)(nil)
)

func ElectronicSignatureEventTypes() []ElectronicSignatureEventType {
	return []ElectronicSignatureEventType{
		ElectronicSignatureEventTypeDocumentViewed,
		ElectronicSignatureEventTypeConsentGiven,
		ElectronicSignatureEventTypeFullNameTyped,
		ElectronicSignatureEventTypeSignatureAccepted,
		ElectronicSignatureEventTypeSignatureCompleted,
		ElectronicSignatureEventTypeSealComputed,
		ElectronicSignatureEventTypeTimestampRequested,
		ElectronicSignatureEventTypeCertificateGenerated,
		ElectronicSignatureEventTypeProcessingError,
	}
}

func (v ElectronicSignatureEventType) IsValid() bool {
	switch v {
	case
		ElectronicSignatureEventTypeDocumentViewed,
		ElectronicSignatureEventTypeConsentGiven,
		ElectronicSignatureEventTypeFullNameTyped,
		ElectronicSignatureEventTypeSignatureAccepted,
		ElectronicSignatureEventTypeSignatureCompleted,
		ElectronicSignatureEventTypeSealComputed,
		ElectronicSignatureEventTypeTimestampRequested,
		ElectronicSignatureEventTypeCertificateGenerated,
		ElectronicSignatureEventTypeProcessingError:
		return true
	}

	return false
}

func (v ElectronicSignatureEventType) String() string {
	return string(v)
}

func (v ElectronicSignatureEventType) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *ElectronicSignatureEventType) UnmarshalText(text []byte) error {
	val := ElectronicSignatureEventType(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid ElectronicSignatureEventType value: %q", string(text))
	}

	*v = val

	return nil
}
