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

	"go.probo.inc/probo/pkg/page"
)

type CompliancePortalAccessOrderField string

const (
	CompliancePortalAccessOrderFieldCreatedAt           CompliancePortalAccessOrderField = "CREATED_AT"
	CompliancePortalAccessOrderFieldPendingRequestCount CompliancePortalAccessOrderField = "PENDING_REQUEST_COUNT"
)

var (
	_ page.OrderField          = CompliancePortalAccessOrderField("")
	_ fmt.Stringer             = CompliancePortalAccessOrderField("")
	_ encoding.TextMarshaler   = CompliancePortalAccessOrderField("")
	_ encoding.TextUnmarshaler = (*CompliancePortalAccessOrderField)(nil)
)

func CompliancePortalAccessOrderFields() []CompliancePortalAccessOrderField {
	return []CompliancePortalAccessOrderField{
		CompliancePortalAccessOrderFieldCreatedAt,
		CompliancePortalAccessOrderFieldPendingRequestCount,
	}
}

func (v CompliancePortalAccessOrderField) IsValid() bool {
	return isValidOrderField(v, CompliancePortalAccessOrderFields())
}

func (v CompliancePortalAccessOrderField) String() string {
	return string(v)
}

func (v CompliancePortalAccessOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *CompliancePortalAccessOrderField) UnmarshalText(text []byte) error {
	return unmarshalOrderField(v, text, CompliancePortalAccessOrderFields())
}

func (tcaof CompliancePortalAccessOrderField) Column() string {
	switch tcaof {
	case CompliancePortalAccessOrderFieldCreatedAt:
		return "created_at"
	case CompliancePortalAccessOrderFieldPendingRequestCount:
		return `(
			SELECT COUNT(*)
			FROM cp_document_accesses
			WHERE compliance_portal_access_id = cp_accesses.id
			AND status = 'REQUESTED'::compliance_portal_document_access_status
		)`
	}

	panic(fmt.Sprintf("unsupported order by: %s", tcaof))
}
