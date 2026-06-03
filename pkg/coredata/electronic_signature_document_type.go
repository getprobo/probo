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
	ElectronicSignatureDocumentType string
)

const (
	ElectronicSignatureDocumentTypeNDA                      ElectronicSignatureDocumentType = "NDA"
	ElectronicSignatureDocumentTypeDPA                      ElectronicSignatureDocumentType = "DPA"
	ElectronicSignatureDocumentTypeMSA                      ElectronicSignatureDocumentType = "MSA"
	ElectronicSignatureDocumentTypeSOW                      ElectronicSignatureDocumentType = "SOW"
	ElectronicSignatureDocumentTypeSLA                      ElectronicSignatureDocumentType = "SLA"
	ElectronicSignatureDocumentTypeTOS                      ElectronicSignatureDocumentType = "TOS"
	ElectronicSignatureDocumentTypePrivacyPolicy            ElectronicSignatureDocumentType = "PRIVACY_POLICY"
	ElectronicSignatureDocumentTypeGovernance               ElectronicSignatureDocumentType = "GOVERNANCE"
	ElectronicSignatureDocumentTypePolicy                   ElectronicSignatureDocumentType = "POLICY"
	ElectronicSignatureDocumentTypeProcedure                ElectronicSignatureDocumentType = "PROCEDURE"
	ElectronicSignatureDocumentTypePlan                     ElectronicSignatureDocumentType = "PLAN"
	ElectronicSignatureDocumentTypeRegister                 ElectronicSignatureDocumentType = "REGISTER"
	ElectronicSignatureDocumentTypeRecord                   ElectronicSignatureDocumentType = "RECORD"
	ElectronicSignatureDocumentTypeReport                   ElectronicSignatureDocumentType = "REPORT"
	ElectronicSignatureDocumentTypeTemplate                 ElectronicSignatureDocumentType = "TEMPLATE"
	ElectronicSignatureDocumentTypeStatementOfApplicability ElectronicSignatureDocumentType = "STATEMENT_OF_APPLICABILITY"
	ElectronicSignatureDocumentTypeOther                    ElectronicSignatureDocumentType = "OTHER"

	ESignProcessConsentText = "By typing my full name and clicking Accept, I consent to sign this document electronically and agree that my electronic signature has the same legal validity as a handwritten signature."
)

var (
	_ fmt.Stringer             = ElectronicSignatureDocumentType("")
	_ encoding.TextMarshaler   = ElectronicSignatureDocumentType("")
	_ encoding.TextUnmarshaler = (*ElectronicSignatureDocumentType)(nil)
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
		ElectronicSignatureDocumentTypeGovernance,
		ElectronicSignatureDocumentTypePolicy,
		ElectronicSignatureDocumentTypeProcedure,
		ElectronicSignatureDocumentTypePlan,
		ElectronicSignatureDocumentTypeRegister,
		ElectronicSignatureDocumentTypeRecord,
		ElectronicSignatureDocumentTypeReport,
		ElectronicSignatureDocumentTypeTemplate,
		ElectronicSignatureDocumentTypeStatementOfApplicability,
		ElectronicSignatureDocumentTypeOther,
	}
}

func (v ElectronicSignatureDocumentType) IsValid() bool {
	switch v {
	case
		ElectronicSignatureDocumentTypeNDA,
		ElectronicSignatureDocumentTypeDPA,
		ElectronicSignatureDocumentTypeMSA,
		ElectronicSignatureDocumentTypeSOW,
		ElectronicSignatureDocumentTypeSLA,
		ElectronicSignatureDocumentTypeTOS,
		ElectronicSignatureDocumentTypePrivacyPolicy,
		ElectronicSignatureDocumentTypeGovernance,
		ElectronicSignatureDocumentTypePolicy,
		ElectronicSignatureDocumentTypeProcedure,
		ElectronicSignatureDocumentTypePlan,
		ElectronicSignatureDocumentTypeRegister,
		ElectronicSignatureDocumentTypeRecord,
		ElectronicSignatureDocumentTypeReport,
		ElectronicSignatureDocumentTypeTemplate,
		ElectronicSignatureDocumentTypeStatementOfApplicability,
		ElectronicSignatureDocumentTypeOther:
		return true
	}

	return false
}

func (v ElectronicSignatureDocumentType) String() string {
	return string(v)
}

func (v ElectronicSignatureDocumentType) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *ElectronicSignatureDocumentType) UnmarshalText(text []byte) error {
	val := ElectronicSignatureDocumentType(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid ElectronicSignatureDocumentType value: %q", string(text))
	}

	*v = val

	return nil
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
	case ElectronicSignatureDocumentTypeGovernance:
		return "Governance Document"
	case ElectronicSignatureDocumentTypePolicy:
		return "Policy"
	case ElectronicSignatureDocumentTypeProcedure:
		return "Procedure"
	case ElectronicSignatureDocumentTypePlan:
		return "Plan"
	case ElectronicSignatureDocumentTypeRegister:
		return "Register"
	case ElectronicSignatureDocumentTypeRecord:
		return "Record"
	case ElectronicSignatureDocumentTypeReport:
		return "Report"
	case ElectronicSignatureDocumentTypeTemplate:
		return "Template"
	case ElectronicSignatureDocumentTypeStatementOfApplicability:
		return "Statement of Applicability"
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
	case ElectronicSignatureDocumentTypeGovernance:
		docAgreement = "I acknowledge and agree to this Governance Document."
	case ElectronicSignatureDocumentTypePolicy:
		docAgreement = "I acknowledge and agree to this Policy."
	case ElectronicSignatureDocumentTypeProcedure:
		docAgreement = "I acknowledge and agree to this Procedure."
	case ElectronicSignatureDocumentTypePlan:
		docAgreement = "I acknowledge and agree to this Plan."
	case ElectronicSignatureDocumentTypeRegister:
		docAgreement = "I acknowledge and agree to this Register."
	case ElectronicSignatureDocumentTypeRecord:
		docAgreement = "I acknowledge and agree to this Record."
	case ElectronicSignatureDocumentTypeReport:
		docAgreement = "I acknowledge and agree to this Report."
	case ElectronicSignatureDocumentTypeTemplate:
		docAgreement = "I acknowledge and agree to this Template."
	case ElectronicSignatureDocumentTypeStatementOfApplicability:
		docAgreement = "I acknowledge and agree to this Statement of Applicability."
	case ElectronicSignatureDocumentTypeOther:
		return "", fmt.Errorf("cannot get consent text: document type OTHER requires explicit consent text")
	default:
		return "", fmt.Errorf("cannot get consent text: unknown document type %q", dt)
	}

	return docAgreement + " " + ESignProcessConsentText, nil
}

func ElectronicSignatureDocumentTypeFromDocumentType(dt DocumentType) ElectronicSignatureDocumentType {
	switch dt {
	case DocumentTypeGovernance:
		return ElectronicSignatureDocumentTypeGovernance
	case DocumentTypePolicy:
		return ElectronicSignatureDocumentTypePolicy
	case DocumentTypeProcedure:
		return ElectronicSignatureDocumentTypeProcedure
	case DocumentTypePlan:
		return ElectronicSignatureDocumentTypePlan
	case DocumentTypeRegister:
		return ElectronicSignatureDocumentTypeRegister
	case DocumentTypeRecord:
		return ElectronicSignatureDocumentTypeRecord
	case DocumentTypeReport:
		return ElectronicSignatureDocumentTypeReport
	case DocumentTypeTemplate:
		return ElectronicSignatureDocumentTypeTemplate
	case DocumentTypeStatementOfApplicability:
		return ElectronicSignatureDocumentTypeStatementOfApplicability
	default:
		return ElectronicSignatureDocumentTypeOther
	}
}
