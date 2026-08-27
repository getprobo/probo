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

	"github.com/jackc/pgx/v5"
)

type (
	CompliancePortalAccessResourceStatusFilter string

	CompliancePortalAccessResourceFilter struct {
		status *CompliancePortalAccessResourceStatusFilter
	}
)

const (
	CompliancePortalAccessResourceStatusFilterRequested CompliancePortalAccessResourceStatusFilter = "REQUESTED"
	CompliancePortalAccessResourceStatusFilterGranted   CompliancePortalAccessResourceStatusFilter = "GRANTED"
	CompliancePortalAccessResourceStatusFilterRejected  CompliancePortalAccessResourceStatusFilter = "REJECTED"
	CompliancePortalAccessResourceStatusFilterRevoked   CompliancePortalAccessResourceStatusFilter = "REVOKED"
	CompliancePortalAccessResourceStatusFilterNone      CompliancePortalAccessResourceStatusFilter = "NONE"
)

var (
	_ fmt.Stringer             = CompliancePortalAccessResourceStatusFilter("")
	_ encoding.TextMarshaler   = CompliancePortalAccessResourceStatusFilter("")
	_ encoding.TextUnmarshaler = (*CompliancePortalAccessResourceStatusFilter)(nil)
)

func CompliancePortalAccessResourceStatusFilters() []CompliancePortalAccessResourceStatusFilter {
	return []CompliancePortalAccessResourceStatusFilter{
		CompliancePortalAccessResourceStatusFilterRequested,
		CompliancePortalAccessResourceStatusFilterGranted,
		CompliancePortalAccessResourceStatusFilterRejected,
		CompliancePortalAccessResourceStatusFilterRevoked,
		CompliancePortalAccessResourceStatusFilterNone,
	}
}

func (v CompliancePortalAccessResourceStatusFilter) IsValid() bool {
	switch v {
	case
		CompliancePortalAccessResourceStatusFilterRequested,
		CompliancePortalAccessResourceStatusFilterGranted,
		CompliancePortalAccessResourceStatusFilterRejected,
		CompliancePortalAccessResourceStatusFilterRevoked,
		CompliancePortalAccessResourceStatusFilterNone:
		return true
	}

	return false
}

func (v CompliancePortalAccessResourceStatusFilter) String() string {
	return string(v)
}

func (v CompliancePortalAccessResourceStatusFilter) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *CompliancePortalAccessResourceStatusFilter) UnmarshalText(text []byte) error {
	val := CompliancePortalAccessResourceStatusFilter(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid CompliancePortalAccessResourceStatusFilter value: %q", string(text))
	}

	*v = val

	return nil
}

func NewCompliancePortalAccessResourceFilter(
	status *CompliancePortalAccessResourceStatusFilter,
) *CompliancePortalAccessResourceFilter {
	return &CompliancePortalAccessResourceFilter{
		status: status,
	}
}

func (f *CompliancePortalAccessResourceFilter) SQLArguments() pgx.StrictNamedArgs {
	args := pgx.StrictNamedArgs{
		"filter_status_none": false,
		"filter_status":      nil,
	}

	if f.status == nil {
		return args
	}

	if *f.status == CompliancePortalAccessResourceStatusFilterNone {
		args["filter_status_none"] = true
		return args
	}

	args["filter_status"] = f.status.String()

	return args
}

func (f *CompliancePortalAccessResourceFilter) SQLFragment() string {
	return `
(
	CASE
		WHEN @filter_status_none::boolean THEN
			status IS NULL
		WHEN @filter_status::text IS NOT NULL THEN
			status = @filter_status::compliance_portal_document_access_status
		ELSE TRUE
	END
)`
}
