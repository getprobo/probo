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

import "github.com/jackc/pgx/v5"

type (
	CompliancePortalAccessFilter struct {
		query *string
	}
)

func NewCompliancePortalAccessFilter(query *string) *CompliancePortalAccessFilter {
	return &CompliancePortalAccessFilter{
		query: query,
	}
}

func (f *CompliancePortalAccessFilter) SQLArguments() pgx.StrictNamedArgs {
	args := pgx.StrictNamedArgs{
		"filter_query": nil,
	}

	if f.query != nil && *f.query != "" {
		args["filter_query"] = escapeLikePattern(*f.query)
	}

	return args
}

func (f *CompliancePortalAccessFilter) SQLFragment() string {
	return `
(
	CASE
		WHEN @filter_query::text IS NOT NULL AND @filter_query::text <> '' THEN
			EXISTS (
				SELECT 1
				FROM iam_membership_profiles p
				INNER JOIN identities i ON i.id = p.identity_id
				WHERE p.identity_id = cp_accesses.identity_id
					AND p.organization_id = cp_accesses.organization_id
					AND (
						p.full_name ILIKE '%' || @filter_query || '%' ESCAPE '\'
						OR i.email_address ILIKE '%' || @filter_query || '%' ESCAPE '\'
					)
			)
		ELSE TRUE
	END
)`
}
