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
)

var (
	_ fmt.Stringer             = ElectronicSignatureDocumentType("")
	_ encoding.TextMarshaler   = ElectronicSignatureDocumentType("")
	_ encoding.TextUnmarshaler = (*ElectronicSignatureDocumentType)(nil)
)

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
