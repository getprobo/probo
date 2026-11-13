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
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

// ErrInvalidCursorValue is a cursor value that cannot be compared with the
// ordered column. Callers refuse it before querying.
var ErrInvalidCursorValue = errors.New("invalid cursor")

type (
	AuditFilter struct {
		compliancePortalVisibilities []CompliancePortalVisibility
		compliancePortalID           *gid.GID
	}
)

func NewAuditFilter() *AuditFilter {
	return &AuditFilter{}
}

func NewAuditCompliancePortalFilter() *AuditFilter {
	return &AuditFilter{
		compliancePortalVisibilities: []CompliancePortalVisibility{
			CompliancePortalVisibilityRestricted,
			CompliancePortalVisibilityPublic,
		},
	}
}

func (f *AuditFilter) WithCompliancePortalID(compliancePortalID gid.GID) *AuditFilter {
	clone := *f
	clone.compliancePortalID = &compliancePortalID

	return &clone
}

func (f *AuditFilter) WithCompliancePortalVisibilities(visibilities ...CompliancePortalVisibility) *AuditFilter {
	f.compliancePortalVisibilities = visibilities
	return f
}

func (f *AuditFilter) SQLArguments() pgx.NamedArgs {
	args := pgx.NamedArgs{
		"compliance_portal_id":           nil,
		"compliance_portal_visibilities": nil,
	}

	if f.compliancePortalID != nil {
		args["compliance_portal_id"] = *f.compliancePortalID
	}

	if f.compliancePortalVisibilities != nil {
		visibilities := make([]string, len(f.compliancePortalVisibilities))
		for i, visibility := range f.compliancePortalVisibilities {
			visibilities[i] = visibility.String()
		}

		args["compliance_portal_visibilities"] = visibilities
	}

	return args
}

// prepareAuditDateCursor rejects a date cursor PostgreSQL cannot cast, and
// rewrites RFC3339 to the UTC calendar day date columns accept.
func prepareAuditDateCursor(cursor *page.Cursor[AuditOrderField]) error {
	if cursor == nil || cursor.Key == nil || !auditDateOrderField(cursor.OrderBy.Field) {
		return nil
	}

	parsed, ok := auditCursorDate(cursor.Key.Value)
	if !ok {
		return ErrInvalidCursorValue
	}

	cursor.Key.Value = parsed

	return nil
}

func auditDateOrderField(field AuditOrderField) bool {
	switch field {
	case AuditOrderFieldValidFrom, AuditOrderFieldValidUntil,
		AuditOrderFieldAuditStartDate, AuditOrderFieldAuditEndDate:
		return true
	default:
		return false
	}
}

func auditCursorDate(value any) (any, bool) {
	switch v := value.(type) {
	case nil:
		return nil, true
	case time.Time:
		return v.UTC().Format(time.DateOnly), true
	case *time.Time:
		if v == nil {
			return nil, true
		}

		return v.UTC().Format(time.DateOnly), true
	case string:
		if t, err := time.Parse(time.DateOnly, v); err == nil {
			return t.Format(time.DateOnly), true
		}

		if t, err := time.Parse(time.RFC3339Nano, v); err == nil {
			return t.UTC().Format(time.DateOnly), true
		}

		return nil, false
	default:
		return nil, false
	}
}

func (f *AuditFilter) SQLFragment() string {
	return `
(
	CASE
		WHEN @compliance_portal_id::text IS NOT NULL
			AND @compliance_portal_visibilities::compliance_portal_visibility[] IS NOT NULL THEN
			EXISTS (
				SELECT 1
				FROM cp_audits
				WHERE cp_audits.audit_id = audits.id
					AND cp_audits.compliance_portal_id = @compliance_portal_id
					AND cp_audits.visibility = ANY(
						@compliance_portal_visibilities::compliance_portal_visibility[]
					)
			)
		WHEN @compliance_portal_visibilities::compliance_portal_visibility[] IS NOT NULL THEN
			FALSE
		ELSE TRUE
	END
)`
}
