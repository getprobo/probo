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
	ElectronicSignatureDocumentType string
)

const (
	ElectronicSignatureDocumentTypeNDA           ElectronicSignatureDocumentType = "NDA"
	ElectronicSignatureDocumentTypeDPA           ElectronicSignatureDocumentType = "DPA"
	ElectronicSignatureDocumentTypeMSA           ElectronicSignatureDocumentType = "MSA"
	ElectronicSignatureDocumentTypeSOW           ElectronicSignatureDocumentType = "SOW"
	ElectronicSignatureDocumentTypeSLA           ElectronicSignatureDocumentType = "SLA"
	ElectronicSignatureDocumentTypeTOS           ElectronicSignatureDocumentType = "TOS"
	ElectronicSignatureDocumentTypePrivacyPolicy ElectronicSignatureDocumentType = "PRIVACY_POLICY"
	ElectronicSignatureDocumentTypeOther         ElectronicSignatureDocumentType = "OTHER"

	ESignProcessConsentText = "By typing my full name and clicking Accept, I consent to sign this document electronically and agree that my electronic signature has the same legal validity as a handwritten signature."
)

func ElectronicSignatureDocumentTypes() []ElectronicSignatureDocumentType {
	return []ElectronicSignatureDocumentType{
		ElectronicSignatureDocumentTypeNDA,
		ElectronicSignatureDocumentTypeDPA,
		ElectronicSignatureDocumentTypeMSA,
		ElectronicSignatureDocumentTypeSOW,
		ElectronicSignatureDocumentTypeSLA,
		ElectronicSignatureDocumentTypeTOS,
		ElectronicSignatureDocumentTypePrivacyPolicy,
		ElectronicSignatureDocumentTypeOther,
	}
}

func (dt ElectronicSignatureDocumentType) MarshalText() ([]byte, error) {
	return []byte(dt.String()), nil
}

func (dt *ElectronicSignatureDocumentType) UnmarshalText(data []byte) error {
	val := string(data)

	switch val {
	case ElectronicSignatureDocumentTypeNDA.String():
		*dt = ElectronicSignatureDocumentTypeNDA
	case ElectronicSignatureDocumentTypeDPA.String():
		*dt = ElectronicSignatureDocumentTypeDPA
	case ElectronicSignatureDocumentTypeMSA.String():
		*dt = ElectronicSignatureDocumentTypeMSA
	case ElectronicSignatureDocumentTypeSOW.String():
		*dt = ElectronicSignatureDocumentTypeSOW
	case ElectronicSignatureDocumentTypeSLA.String():
		*dt = ElectronicSignatureDocumentTypeSLA
	case ElectronicSignatureDocumentTypeTOS.String():
		*dt = ElectronicSignatureDocumentTypeTOS
	case ElectronicSignatureDocumentTypePrivacyPolicy.String():
		*dt = ElectronicSignatureDocumentTypePrivacyPolicy
	case ElectronicSignatureDocumentTypeOther.String():
		*dt = ElectronicSignatureDocumentTypeOther
	default:
		return fmt.Errorf("invalid ElectronicSignatureDocumentType value: %q", val)
	}

	return nil
}

func (dt ElectronicSignatureDocumentType) String() string {
	return string(dt)
}

func (dt *ElectronicSignatureDocumentType) Scan(value any) error {
	val, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid scan source for ElectronicSignatureDocumentType, expected string got %T", value)
	}

	return dt.UnmarshalText([]byte(val))
}

func (dt ElectronicSignatureDocumentType) Value() (driver.Value, error) {
	return dt.String(), nil
}

func (dt ElectronicSignatureDocumentType) DisplayName() string {
	switch dt {
	case ElectronicSignatureDocumentTypeNDA:
		return "Non-Disclosure Agreement"
	case ElectronicSignatureDocumentTypeDPA:
		return "Data Processing Agreement"
	case ElectronicSignatureDocumentTypeMSA:
		return "Master Service Agreement"
	case ElectronicSignatureDocumentTypeSOW:
		return "Statement of Work"
	case ElectronicSignatureDocumentTypeSLA:
		return "Service Level Agreement"
	case ElectronicSignatureDocumentTypeTOS:
		return "Terms of Service"
	case ElectronicSignatureDocumentTypePrivacyPolicy:
		return "Privacy Policy"
	default:
		return string(dt)
	}
}

func (dt ElectronicSignatureDocumentType) ConsentText() (string, error) {
	var docAgreement string
	switch dt {
	case ElectronicSignatureDocumentTypeNDA:
		docAgreement = "I agree to the terms of this Non-Disclosure Agreement."
	case ElectronicSignatureDocumentTypeDPA:
		docAgreement = "I agree to the terms of this Data Processing Agreement."
	case ElectronicSignatureDocumentTypeMSA:
		docAgreement = "I agree to the terms of this Master Service Agreement."
	case ElectronicSignatureDocumentTypeSOW:
		docAgreement = "I agree to the terms of this Statement of Work."
	case ElectronicSignatureDocumentTypeSLA:
		docAgreement = "I agree to the terms of this Service Level Agreement."
	case ElectronicSignatureDocumentTypeTOS:
		docAgreement = "I agree to these Terms of Service."
	case ElectronicSignatureDocumentTypePrivacyPolicy:
		docAgreement = "I agree to this Privacy Policy."
	case ElectronicSignatureDocumentTypeOther:
		return "", fmt.Errorf("document type OTHER requires explicit consent text")
	default:
		return "", fmt.Errorf("unknown document type %q", dt)
	}

	return docAgreement + " " + ESignProcessConsentText, nil
}
