// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
