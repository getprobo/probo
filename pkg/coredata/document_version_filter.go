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
	"go.probo.inc/probo/pkg/mail"
)

type (
	DocumentVersionFilter struct {
		statuses           []DocumentVersionStatus
		userEmail          *mail.Addr
		approverIdentityID *gid.GID
	}
)

func NewDocumentVersionFilter() *DocumentVersionFilter {
	return &DocumentVersionFilter{}
}

func (f *DocumentVersionFilter) WithStatuses(statuses ...DocumentVersionStatus) *DocumentVersionFilter {
	f.statuses = statuses
	return f
}

func (f *DocumentVersionFilter) WithUserEmail(userEmail *mail.Addr) *DocumentVersionFilter {
	f.userEmail = userEmail
	return f
}

func (f *DocumentVersionFilter) WithApproverIdentityID(identityID *gid.GID) *DocumentVersionFilter {
	f.approverIdentityID = identityID
	return f
}

func (f *DocumentVersionFilter) SQLArguments() pgx.StrictNamedArgs {
	var filterStatuses []string
	for _, s := range f.statuses {
		filterStatuses = append(filterStatuses, s.String())
	}

	return pgx.StrictNamedArgs{
		"filter_statuses":      filterStatuses,
		"user_email":           f.userEmail,
		"approver_identity_id": f.approverIdentityID,
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
		@user_email::text IS NULL
		OR EXISTS (
			SELECT 1
			FROM document_version_signatures dvs
			INNER JOIN iam_membership_profiles p ON dvs.signed_by_profile_id = p.id
			INNER JOIN identities i ON p.identity_id = i.id
			WHERE dvs.document_version_id = document_versions.id
				AND i.email_address = @user_email::CITEXT
				AND dvs.state IN ('REQUESTED', 'SIGNED')
		)
	)
	AND
	(
		@approver_identity_id::text IS NULL
		OR EXISTS (
			SELECT 1
			FROM document_version_approval_quorums dvaq
			INNER JOIN document_version_approval_decisions dvad ON dvad.quorum_id = dvaq.id
			INNER JOIN iam_membership_profiles p ON dvad.approver_id = p.id
			WHERE dvaq.version_id = document_versions.id
				AND p.identity_id = @approver_identity_id::text
		)
	)
)`
}
