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
	"context"
	"fmt"
	"maps"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	CompliancePortalAccessResource struct {
		ID       gid.GID                               `db:"id"`
		Kind     CompliancePortalAccessResourceKind    `db:"kind"`
		Name     string                                `db:"name"`
		Category string                                `db:"category"`
		Status   *CompliancePortalDocumentAccessStatus `db:"status"`
	}

	CompliancePortalAccessResources []*CompliancePortalAccessResource
)

func (r *CompliancePortalAccessResource) accessStatusRank() int {
	if r.Status == nil {
		return 3
	}

	switch *r.Status {
	case CompliancePortalDocumentAccessStatusRequested:
		return 1
	case CompliancePortalDocumentAccessStatusGranted:
		return 2
	case CompliancePortalDocumentAccessStatusRevoked:
		return 4
	case CompliancePortalDocumentAccessStatusRejected:
		return 5
	}

	return 3
}

func (r *CompliancePortalAccessResource) CursorKey(
	orderBy CompliancePortalAccessResourceOrderField,
) page.CursorKey {
	switch orderBy {
	case CompliancePortalAccessResourceOrderFieldAccessStatus:
		return page.NewCursorKey(r.ID, r.accessStatusRank())
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func compliancePortalAccessResourcesUnionSQL() string {
	return `
SELECT
    documents.id,
    @kind_document::text AS kind,
    published_versions.title AS name,
    published_versions.document_type::text AS category,
    da.status
FROM (
    SELECT
        id,
        current_published_major,
        current_published_minor
    FROM
        documents
    WHERE
        %s
        AND deleted_at IS NULL
        AND organization_id = @organization_id
        AND status = @document_status_active::document_status
        AND current_published_major IS NOT NULL
) documents
INNER JOIN document_versions published_versions
    ON published_versions.document_id = documents.id
    AND published_versions.major = documents.current_published_major
    AND published_versions.minor = documents.current_published_minor
INNER JOIN cp_documents cpd
    ON cpd.document_id = documents.id
    AND cpd.compliance_portal_id = @compliance_portal_id
    AND cpd.visibility = @visibility_restricted::compliance_portal_visibility
LEFT JOIN cp_document_accesses da
    ON da.compliance_portal_access_id = @compliance_portal_access_id
    AND da.document_id = documents.id

UNION ALL

SELECT
    audits.report_file_id AS id,
    @kind_report::text AS kind,
    files.file_name AS name,
    frameworks.name AS category,
    da.status
FROM (
    SELECT
        id,
        report_file_id,
        framework_id
    FROM
        audits
    WHERE
        %s
        AND organization_id = @organization_id
        AND report_file_id IS NOT NULL
) audits
INNER JOIN cp_audits cpa
    ON cpa.audit_id = audits.id
    AND cpa.compliance_portal_id = @compliance_portal_id
    AND cpa.visibility = @visibility_restricted::compliance_portal_visibility
INNER JOIN files
    ON files.id = audits.report_file_id
INNER JOIN frameworks
    ON frameworks.id = audits.framework_id
LEFT JOIN cp_document_accesses da
    ON da.compliance_portal_access_id = @compliance_portal_access_id
    AND da.report_file_id = audits.report_file_id

UNION ALL

SELECT
    cp_files.id,
    @kind_file::text AS kind,
    cp_files.name,
    cp_files.category,
    da.status
FROM (
    SELECT
        id,
        name,
        category
    FROM
        cp_files
    WHERE
        %s
        AND compliance_portal_id = @compliance_portal_id
        AND compliance_portal_visibility = ANY(@file_visibilities::compliance_portal_visibility[])
) cp_files
LEFT JOIN cp_document_accesses da
    ON da.compliance_portal_access_id = @compliance_portal_access_id
    AND da.compliance_portal_file_id = cp_files.id
`
}

func compliancePortalAccessResourceUnionArgs(
	scope Scoper,
	compliancePortalAccessID gid.GID,
	organizationID gid.GID,
	compliancePortalID gid.GID,
) pgx.StrictNamedArgs {
	args := pgx.StrictNamedArgs{
		"compliance_portal_access_id": compliancePortalAccessID,
		"organization_id":             organizationID,
		"compliance_portal_id":        compliancePortalID,
		"kind_document":               CompliancePortalAccessResourceKindDocument,
		"kind_report":                 CompliancePortalAccessResourceKindReport,
		"kind_file":                   CompliancePortalAccessResourceKindFile,
		"document_status_active":      DocumentStatusActive,
		"visibility_restricted":       CompliancePortalVisibilityRestricted,
		"file_visibilities": []string{
			CompliancePortalVisibilityRestricted.String(),
			CompliancePortalVisibilityNone.String(),
		},
	}
	maps.Copy(args, scope.SQLArguments())

	return args
}

func compliancePortalAccessResourceOrderArgs() pgx.StrictNamedArgs {
	return pgx.StrictNamedArgs{
		"status_requested": CompliancePortalDocumentAccessStatusRequested,
		"status_granted":   CompliancePortalDocumentAccessStatusGranted,
		"status_revoked":   CompliancePortalDocumentAccessStatusRevoked,
		"status_rejected":  CompliancePortalDocumentAccessStatusRejected,
	}
}

func (r *CompliancePortalAccessResources) LoadByCompliancePortalAccessID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalAccessID gid.GID,
	organizationID gid.GID,
	compliancePortalID gid.GID,
	cursor *page.Cursor[CompliancePortalAccessResourceOrderField],
	filter *CompliancePortalAccessResourceFilter,
) error {
	q := `
SELECT
    id,
    kind,
    name,
    category,
    status
FROM (
` + compliancePortalAccessResourcesUnionSQL() + `
) resources
WHERE
    %s
    AND %s
`

	q = fmt.Sprintf(
		q,
		scope.SQLFragment(),
		scope.SQLFragment(),
		scope.SQLFragment(),
		filter.SQLFragment(),
		cursor.SQLFragment(),
	)

	args := compliancePortalAccessResourceUnionArgs(
		scope,
		compliancePortalAccessID,
		organizationID,
		compliancePortalID,
	)
	maps.Copy(args, compliancePortalAccessResourceOrderArgs())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query compliance portal access resources: %w", err)
	}

	resources, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CompliancePortalAccessResource])
	if err != nil {
		return fmt.Errorf("cannot collect compliance portal access resources: %w", err)
	}

	*r = resources

	return nil
}

func (r *CompliancePortalAccessResources) CountByCompliancePortalAccessID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalAccessID gid.GID,
	organizationID gid.GID,
	compliancePortalID gid.GID,
	filter *CompliancePortalAccessResourceFilter,
) (int, error) {
	q := `
SELECT
    COUNT(*)
FROM (
` + compliancePortalAccessResourcesUnionSQL() + `
) resources
WHERE
    %s
`

	q = fmt.Sprintf(
		q,
		scope.SQLFragment(),
		scope.SQLFragment(),
		scope.SQLFragment(),
		filter.SQLFragment(),
	)

	args := compliancePortalAccessResourceUnionArgs(
		scope,
		compliancePortalAccessID,
		organizationID,
		compliancePortalID,
	)
	maps.Copy(args, filter.SQLArguments())

	var count int
	err := conn.QueryRow(ctx, q, args).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count compliance portal access resources: %w", err)
	}

	return count, nil
}
