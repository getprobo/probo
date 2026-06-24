// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	DocumentType string
)

const (
	DocumentTypeOther                    DocumentType = "OTHER"
	DocumentTypeGovernance               DocumentType = "GOVERNANCE"
	DocumentTypePolicy                   DocumentType = "POLICY"
	DocumentTypeProcedure                DocumentType = "PROCEDURE"
	DocumentTypePlan                     DocumentType = "PLAN"
	DocumentTypeRegister                 DocumentType = "REGISTER"
	DocumentTypeRecord                   DocumentType = "RECORD"
	DocumentTypeReport                   DocumentType = "REPORT"
	DocumentTypeTemplate                 DocumentType = "TEMPLATE"
	DocumentTypeStatementOfApplicability DocumentType = "STATEMENT_OF_APPLICABILITY"
)

var (
	_ fmt.Stringer             = DocumentType("")
	_ encoding.TextMarshaler   = DocumentType("")
	_ encoding.TextUnmarshaler = (*DocumentType)(nil)
)

func DocumentTypes() []DocumentType {
	return []DocumentType{
		DocumentTypeOther,
		DocumentTypeGovernance,
		DocumentTypePolicy,
		DocumentTypeProcedure,
		DocumentTypePlan,
		DocumentTypeRegister,
		DocumentTypeRecord,
		DocumentTypeReport,
		DocumentTypeTemplate,
		DocumentTypeStatementOfApplicability,
	}
}

func (v DocumentType) IsValid() bool {
	switch v {
	case
		DocumentTypeOther,
		DocumentTypeGovernance,
		DocumentTypePolicy,
		DocumentTypeProcedure,
		DocumentTypePlan,
		DocumentTypeRegister,
		DocumentTypeRecord,
		DocumentTypeReport,
		DocumentTypeTemplate,
		DocumentTypeStatementOfApplicability:
		return true
	}

	return false
}

func (v DocumentType) String() string {
	return string(v)
}

func (v DocumentType) Label() string {
	switch v {
	case DocumentTypeGovernance:
		return "Governance"
	case DocumentTypePolicy:
		return "Policy"
	case DocumentTypeProcedure:
		return "Procedure"
	case DocumentTypePlan:
		return "Plan"
	case DocumentTypeRegister:
		return "Register"
	case DocumentTypeRecord:
		return "Record"
	case DocumentTypeReport:
		return "Report"
	case DocumentTypeTemplate:
		return "Template"
	case DocumentTypeStatementOfApplicability:
		return "Statement of Applicability"
	default:
		return "Document"
	}
}

func (v DocumentType) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *DocumentType) UnmarshalText(text []byte) error {
	val := DocumentType(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid DocumentType value: %q", string(text))
	}

	*v = val

	return nil
}
