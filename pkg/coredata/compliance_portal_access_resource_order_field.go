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

	"go.probo.inc/probo/pkg/page"
)

type CompliancePortalAccessResourceOrderField string

const (
	CompliancePortalAccessResourceOrderFieldAccessStatus CompliancePortalAccessResourceOrderField = "ACCESS_STATUS"
)

var (
	_ page.OrderField          = CompliancePortalAccessResourceOrderField("")
	_ fmt.Stringer             = CompliancePortalAccessResourceOrderField("")
	_ encoding.TextMarshaler   = CompliancePortalAccessResourceOrderField("")
	_ encoding.TextUnmarshaler = (*CompliancePortalAccessResourceOrderField)(nil)
)

func CompliancePortalAccessResourceOrderFields() []CompliancePortalAccessResourceOrderField {
	return []CompliancePortalAccessResourceOrderField{
		CompliancePortalAccessResourceOrderFieldAccessStatus,
	}
}

func (v CompliancePortalAccessResourceOrderField) IsValid() bool {
	return isValidOrderField(v, CompliancePortalAccessResourceOrderFields())
}

func (v CompliancePortalAccessResourceOrderField) String() string {
	return string(v)
}

func (v CompliancePortalAccessResourceOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *CompliancePortalAccessResourceOrderField) UnmarshalText(text []byte) error {
	return unmarshalOrderField(v, text, CompliancePortalAccessResourceOrderFields())
}

func (f CompliancePortalAccessResourceOrderField) Column() string {
	switch f {
	case CompliancePortalAccessResourceOrderFieldAccessStatus:
		return `(
			CASE
				WHEN status = @status_requested::compliance_portal_document_access_status THEN 1
				WHEN status = @status_granted::compliance_portal_document_access_status THEN 2
				WHEN status = @status_revoked::compliance_portal_document_access_status THEN 4
				WHEN status = @status_rejected::compliance_portal_document_access_status THEN 5
				ELSE 3
			END
		)`
	}

	panic(fmt.Sprintf("unsupported order by: %s", f))
}
