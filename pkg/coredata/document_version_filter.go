// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	"github.com/jackc/pgx/v5"
	"go.probo.inc/probo/pkg/gid"
)

type EmployeeFilterMode string

const (
	EmployeeFilterModeSignature EmployeeFilterMode = "signature"
	EmployeeFilterModeApproval  EmployeeFilterMode = "approval"
)

type (
	DocumentVersionFilter struct {
		statuses            []DocumentVersionStatus
		employeeIdentityID  *gid.GID
		employeeFilterModes []EmployeeFilterMode
	}
)

func NewDocumentVersionFilter() *DocumentVersionFilter {
	return &DocumentVersionFilter{}
}

func (f *DocumentVersionFilter) WithStatuses(statuses ...DocumentVersionStatus) *DocumentVersionFilter {
	f.statuses = statuses
	return f
}

func (f *DocumentVersionFilter) WithEmployeeIdentityID(identityID *gid.GID, modes ...EmployeeFilterMode) *DocumentVersionFilter {
	f.employeeIdentityID = identityID
	f.employeeFilterModes = modes
	return f
}

func (f *DocumentVersionFilter) SQLArguments() pgx.StrictNamedArgs {
	var filterStatuses []string
	for _, s := range f.statuses {
		filterStatuses = append(filterStatuses, s.String())
	}

	var employeeFilterModes []string
	for _, m := range f.employeeFilterModes {
		employeeFilterModes = append(employeeFilterModes, string(m))
	}

	return pgx.StrictNamedArgs{
		"filter_statuses":       filterStatuses,
		"employee_identity_id":  f.employeeIdentityID,
		"employee_filter_modes": employeeFilterModes,
	}
}

func (f *DocumentVersionFilter) SQLFragment() string {
	return `
(
	(
		@filter_statuses::text[] IS NULL
		OR document_versions.status::text = ANY(@filter_statuses::text[])
	)
	AND
	(
		@employee_identity_id::text IS NULL
		OR (
			'signature' = ANY(@employee_filter_modes::text[]) AND EXISTS (
				SELECT 1
				FROM document_version_signatures dvs
				INNER JOIN iam_membership_profiles p ON dvs.signed_by_profile_id = p.id
				WHERE dvs.document_version_id = document_versions.id
					AND p.identity_id = @employee_identity_id::text
					AND dvs.state IN ('REQUESTED', 'SIGNED')
			)
		)
		OR (
			'approval' = ANY(@employee_filter_modes::text[]) AND EXISTS (
				SELECT 1
				FROM document_version_approval_quorums dvaq
				INNER JOIN document_version_approval_decisions dvad ON dvad.quorum_id = dvaq.id
				INNER JOIN iam_membership_profiles p ON dvad.approver_id = p.id
				WHERE dvaq.version_id = document_versions.id
					AND p.identity_id = @employee_identity_id::text
					AND NOT (dvad.state = 'APPROVED' AND dvad.electronic_signature_id IS NULL)
			)
		)
	)
)`
}
