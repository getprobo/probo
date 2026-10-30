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

type EmployeePortalDocumentFilter struct {
	published *bool
}

func NewEmployeePortalDocumentFilter(published *bool) *EmployeePortalDocumentFilter {
	return &EmployeePortalDocumentFilter{published: published}
}

func (f *EmployeePortalDocumentFilter) SQLFragment() string {
	return `
(
	CASE
		WHEN @filter_published::boolean IS NULL THEN TRUE
		ELSE (
			document_id IN (
				SELECT documents.id
				FROM documents
				WHERE
					documents.tenant_id = employee_portal_documents.tenant_id
					AND documents.deleted_at IS NULL
					AND documents.status = @filter_document_status::document_status
					AND documents.current_published_major IS NOT NULL
					AND documents.current_published_minor IS NOT NULL
			)
		) = @filter_published::boolean
	END
)`
}

func (f *EmployeePortalDocumentFilter) SQLArguments() pgx.StrictNamedArgs {
	var published any
	if f != nil && f.published != nil {
		published = *f.published
	}

	return pgx.StrictNamedArgs{
		"filter_published":       published,
		"filter_document_status": DocumentStatusActive,
	}
}
