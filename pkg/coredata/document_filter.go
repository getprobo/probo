// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
	"go.probo.inc/probo/pkg/mail"
)

type (
	DocumentFilter struct {
		query                   *string
		trustCenterVisibilities []TrustCenterVisibility
		published               *bool
		userEmail               *mail.Addr
	}
)

func NewDocumentFilter(query *string) *DocumentFilter {
	return &DocumentFilter{
		query: query,
	}
}

func NewDocumentTrustCenterFilter() *DocumentFilter {
	return &DocumentFilter{
		trustCenterVisibilities: []TrustCenterVisibility{
			TrustCenterVisibilityPrivate,
			TrustCenterVisibilityPublic,
		},
	}
}

func (f *DocumentFilter) WithPublished(published *bool) *DocumentFilter {
	f.published = published
	return f
}

func (f *DocumentFilter) WithUserEmail(userEmail *mail.Addr) *DocumentFilter {
	f.userEmail = userEmail
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
	return pgx.NamedArgs{
		"query":                     f.query,
		"trust_center_visibilities": visibilities,
		"published":                 f.published,
		"user_email":                f.userEmail,
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
		WHEN @published::boolean IS TRUE THEN current_published_version IS NOT NULL
		WHEN @published::boolean IS FALSE THEN current_published_version IS NULL
	END
	AND
	CASE
		WHEN @user_email::text IS NULL THEN TRUE
		ELSE EXISTS (
			SELECT 1
			FROM document_versions dv
			INNER JOIN document_version_signatures dvs ON dv.id = dvs.document_version_id
			INNER JOIN iam_membership_profiles p ON dvs.signed_by_profile_id = p.id
			INNER JOIN identities i ON p.identity_id = i.id
			WHERE dv.document_id = documents.id
				AND dv.status = 'PUBLISHED'
				AND i.email_address = @user_email::CITEXT
				AND dvs.state IN ('REQUESTED', 'SIGNED')
		)
	END
)`
}
