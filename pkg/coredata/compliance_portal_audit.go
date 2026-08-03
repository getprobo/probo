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
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
)

type (
	CompliancePortalAudit struct {
		ID                 gid.GID                    `db:"id"`
		OrganizationID     gid.GID                    `db:"organization_id"`
		CompliancePortalID gid.GID                    `db:"compliance_portal_id"`
		AuditID            gid.GID                    `db:"audit_id"`
		Visibility         CompliancePortalVisibility `db:"visibility"`
		CreatedAt          time.Time                  `db:"created_at"`
		UpdatedAt          time.Time                  `db:"updated_at"`
	}

	CompliancePortalAudits []*CompliancePortalAudit
)

func (cpa *CompliancePortalAudit) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM cp_audits WHERE id = ANY(@resource_ids::text[])`

	args := pgx.StrictNamedArgs{"resource_ids": resourceIDs}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query authorization attributes: %w", err)
	}

	defer rows.Close()

	attrsByID := make(policy.AttributesByID)

	for rows.Next() {
		var id, organizationID gid.GID

		if err := rows.Scan(&id, &organizationID); err != nil {
			return nil, fmt.Errorf("cannot scan authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{
			"organization_id": organizationID.String(),
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate authorization attributes: %w", err)
	}

	return attrsByID, nil
}

func (cpa *CompliancePortalAudit) LoadByCompliancePortalIDAndAuditID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalID gid.GID,
	auditID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	compliance_portal_id,
	audit_id,
	visibility,
	created_at,
	updated_at
FROM
	cp_audits
WHERE
	%s
	AND compliance_portal_id = @compliance_portal_id
	AND audit_id = @audit_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"compliance_portal_id": compliancePortalID,
		"audit_id":             auditID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust center audit: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CompliancePortalAudit])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect trust center audit: %w", err)
	}

	*cpa = row

	return nil
}

func (cpa *CompliancePortalAudit) LoadByCompliancePortalIDAndReportFileID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalID gid.GID,
	reportFileID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	compliance_portal_id,
	audit_id,
	visibility,
	created_at,
	updated_at
FROM
	cp_audits
WHERE
	%s
	AND compliance_portal_id = @compliance_portal_id
	AND audit_id IN (
		SELECT
			id
		FROM
			audits
		WHERE %s
			AND report_file_id = @report_file_id
	)
ORDER BY
	CASE
		WHEN visibility = @public_visibility::compliance_portal_visibility THEN 0
		ELSE 1
	END,
	audit_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment(), scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"compliance_portal_id": compliancePortalID,
		"report_file_id":       reportFileID,
		"public_visibility":    CompliancePortalVisibilityPublic,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust center audit: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CompliancePortalAudit])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect trust center audit: %w", err)
	}

	*cpa = row

	return nil
}

func (cpa *CompliancePortalAudit) LoadByCompliancePortalIDAndFrameworkID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalID gid.GID,
	frameworkID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	compliance_portal_id,
	audit_id,
	visibility,
	created_at,
	updated_at
FROM
	cp_audits
WHERE
	%s
	AND compliance_portal_id = @compliance_portal_id
	AND audit_id IN (
		SELECT
			id
		FROM
			audits
		WHERE %s
			AND framework_id = @framework_id
	)
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment(), scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"compliance_portal_id": compliancePortalID,
		"framework_id":         frameworkID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust center audit: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CompliancePortalAudit])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect trust center audit: %w", err)
	}

	*cpa = row

	return nil
}

// LoadCompliancePortalAuditsByCompliancePortalIDAndAuditIDs loads portal audit
// link rows for a bounded set of audit IDs.
func LoadCompliancePortalAuditsByCompliancePortalIDAndAuditIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalID gid.GID,
	auditIDs []gid.GID,
) (map[gid.GID]*CompliancePortalAudit, error) {
	if len(auditIDs) == 0 {
		return map[gid.GID]*CompliancePortalAudit{}, nil
	}

	q := `
SELECT
	id,
	organization_id,
	compliance_portal_id,
	audit_id,
	visibility,
	created_at,
	updated_at
FROM
	cp_audits
WHERE
	%s
	AND compliance_portal_id = @compliance_portal_id
	AND audit_id = ANY(@audit_ids::text[]);
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"compliance_portal_id": compliancePortalID,
		"audit_ids":            auditIDs,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query trust center audits: %w", err)
	}
	defer rows.Close()

	rowsByAuditID := map[gid.GID]*CompliancePortalAudit{}

	for rows.Next() {
		row, err := pgx.RowToStructByName[CompliancePortalAudit](rows)
		if err != nil {
			return nil, fmt.Errorf("cannot scan trust center audit: %w", err)
		}

		rowsByAuditID[row.AuditID] = &row
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot collect trust center audits: %w", err)
	}

	return rowsByAuditID, nil
}

func (cpa *CompliancePortalAudit) Upsert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO cp_audits (
	id,
	tenant_id,
	organization_id,
	compliance_portal_id,
	audit_id,
	visibility,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@compliance_portal_id,
	@audit_id,
	@visibility,
	@created_at,
	@updated_at
)
ON CONFLICT (compliance_portal_id, audit_id) DO UPDATE
SET
	visibility = EXCLUDED.visibility,
	updated_at = EXCLUDED.updated_at
RETURNING
	id,
	organization_id,
	compliance_portal_id,
	audit_id,
	visibility,
	created_at,
	updated_at
`

	args := pgx.StrictNamedArgs{
		"id":                   cpa.ID,
		"tenant_id":            scope.GetTenantID(),
		"organization_id":      cpa.OrganizationID,
		"compliance_portal_id": cpa.CompliancePortalID,
		"audit_id":             cpa.AuditID,
		"visibility":           cpa.Visibility,
		"created_at":           cpa.CreatedAt,
		"updated_at":           cpa.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot upsert trust center audit: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CompliancePortalAudit])
	if err != nil {
		return fmt.Errorf("cannot collect trust center audit: %w", err)
	}

	*cpa = row

	return nil
}

func DeleteCompliancePortalAuditByCompliancePortalIDAndAuditID(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	compliancePortalID gid.GID,
	auditID gid.GID,
) error {
	q := `
DELETE FROM cp_audits
WHERE
	%s
	AND compliance_portal_id = @compliance_portal_id
	AND audit_id = @audit_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"compliance_portal_id": compliancePortalID,
		"audit_id":             auditID,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete trust center audit: %w", err)
	}

	return nil
}

func DeleteCompliancePortalAuditByID(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	id gid.GID,
) error {
	q := `
DELETE FROM cp_audits
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete trust center audit: %w", err)
	}

	return nil
}
