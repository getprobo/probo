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

type (
	DocumentFilter struct {
		query                   *string
		trustCenterVisibilities []TrustCenterVisibility
		published               *bool
		employeeIdentityID      *gid.GID
		employeeFilterModes     []EmployeeFilterMode
		documentTypes           []DocumentType
		classifications         []DocumentClassification
		status                  []DocumentStatus
	}
)

func NewDocumentFilter(query *string) *DocumentFilter {
	return &DocumentFilter{
		query: query,
	}
}

func NewDocumentTrustCenterFilter() *DocumentFilter {
	published := true
	return &DocumentFilter{
		trustCenterVisibilities: []TrustCenterVisibility{
			TrustCenterVisibilityPrivate,
			TrustCenterVisibilityPublic,
		},
		published: &published,
		status:    []DocumentStatus{DocumentStatusActive},
	}
}

func (f *DocumentFilter) WithPublished(published *bool) *DocumentFilter {
	f.published = published
	return f
}

func (f *DocumentFilter) WithEmployeeIdentityID(identityID *gid.GID, modes ...EmployeeFilterMode) *DocumentFilter {
	f.employeeIdentityID = identityID
	f.employeeFilterModes = modes
	return f
}

func (f *DocumentFilter) WithDocumentTypes(documentTypes []DocumentType) *DocumentFilter {
	f.documentTypes = documentTypes
	return f
}

func (f *DocumentFilter) WithClassifications(classifications []DocumentClassification) *DocumentFilter {
	f.classifications = classifications
	return f
}

func (f *DocumentFilter) WithStatus(status []DocumentStatus) *DocumentFilter {
	f.status = status
	return f
}

func (f *DocumentFilter) SQLArguments() pgx.NamedArgs {
	var visibilities []string
	if f.trustCenterVisibilities != nil {
		visibilities = make([]string, len(f.trustCenterVisibilities))
		for i, v := range f.trustCenterVisibilities {
			visibilities[i] = v.String()
		}
	}

	var documentTypes []string
	if f.documentTypes != nil {
		documentTypes = make([]string, len(f.documentTypes))
		for i, dt := range f.documentTypes {
			documentTypes[i] = dt.String()
		}
	}

	var classifications []string
	if f.classifications != nil {
		classifications = make([]string, len(f.classifications))
		for i, c := range f.classifications {
			classifications[i] = c.String()
		}
	}

	var status []string
	if f.status != nil {
		status = make([]string, len(f.status))
		for i, s := range f.status {
			status[i] = s.String()
		}
	}

	var employeeFilterModes []string
	for _, m := range f.employeeFilterModes {
		employeeFilterModes = append(employeeFilterModes, string(m))
	}

	return pgx.NamedArgs{
		"query":                     f.query,
		"trust_center_visibilities": visibilities,
		"published":                 f.published,
		"employee_identity_id":      f.employeeIdentityID,
		"employee_filter_modes":     employeeFilterModes,
		"document_types":            documentTypes,
		"classifications":           classifications,
		"document_status":           status,
	}
}

func (f *DocumentFilter) SQLFragment() string {
	return `
(
	CASE
		WHEN @query::text IS NOT NULL AND @query::text != '' THEN
			search_vector @@ (
				SELECT to_tsquery('simple', string_agg(lexeme || ':*', ' & '))
				FROM unnest(regexp_split_to_array(trim(@query::text), '\s+')) AS lexeme
			)
		ELSE TRUE
	END
	AND
	CASE
		WHEN @trust_center_visibilities::trust_center_visibility[] IS NOT NULL THEN
			trust_center_visibility = ANY(@trust_center_visibilities::trust_center_visibility[])
		ELSE TRUE
	END
	AND
	CASE
		WHEN @published::boolean IS NULL THEN TRUE
		WHEN @published::boolean IS TRUE THEN current_published_major IS NOT NULL
		WHEN @published::boolean IS FALSE THEN current_published_major IS NULL
	END
	AND
	CASE
		WHEN @employee_identity_id::text IS NULL THEN TRUE
		ELSE (
			(
				'signature' = ANY(@employee_filter_modes::text[]) AND EXISTS (
					SELECT 1
					FROM document_versions dv
					INNER JOIN document_version_signatures dvs ON dv.id = dvs.document_version_id
					INNER JOIN iam_membership_profiles p ON dvs.signed_by_profile_id = p.id
					WHERE dv.document_id = documents.id
						AND p.identity_id = @employee_identity_id::text
						AND dvs.state IN ('REQUESTED', 'SIGNED')
				)
			)
			OR (
				'approval' = ANY(@employee_filter_modes::text[]) AND EXISTS (
					SELECT 1
					FROM document_versions dv
					INNER JOIN document_version_approval_quorums dvaq ON dvaq.version_id = dv.id
					INNER JOIN document_version_approval_decisions dvad ON dvad.quorum_id = dvaq.id
					INNER JOIN iam_membership_profiles p ON dvad.approver_id = p.id
					WHERE dv.document_id = documents.id
						AND p.identity_id = @employee_identity_id::text
				)
			)
		)
	END
	AND
	CASE
		WHEN @document_types::document_type[] IS NOT NULL THEN
			document_type = ANY(@document_types::document_type[])
		ELSE TRUE
	END
	AND
	CASE
		WHEN @classifications::document_classification[] IS NOT NULL THEN
			(
				SELECT dv.classification
				FROM document_versions dv
				WHERE dv.document_id = documents.id
				ORDER BY dv.major DESC, dv.minor DESC
				LIMIT 1
			) = ANY(@classifications::document_classification[])
		ELSE TRUE
	END
	AND
	CASE
		WHEN @document_status::text[] IS NULL THEN TRUE
		ELSE status::text = ANY(@document_status::text[])
	END
)`
}
