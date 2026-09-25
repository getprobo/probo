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
	"context"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	CompliancePortalDocumentAccess struct {
		ID                       gid.GID                              `db:"id"`
		OrganizationID           gid.GID                              `db:"organization_id"`
		CompliancePortalAccessID gid.GID                              `db:"compliance_portal_access_id"`
		DocumentID               *gid.GID                             `db:"document_id"`
		ReportFileID             *gid.GID                             `db:"report_file_id"`
		CompliancePortalFileID   *gid.GID                             `db:"compliance_portal_file_id"`
		Status                   CompliancePortalDocumentAccessStatus `db:"status"`
		// RequestedAt is set when the visitor request flow creates the row and
		// preserved across later status changes. Nil for admin-created grants.
		RequestedAt *time.Time `db:"requested_at"`
		CreatedAt   time.Time  `db:"created_at"`
		UpdatedAt   time.Time  `db:"updated_at"`
	}

	CompliancePortalDocumentAccesses []*CompliancePortalDocumentAccess
)

func (tcda *CompliancePortalDocumentAccess) CursorKey(orderBy CompliancePortalDocumentAccessOrderField) page.CursorKey {
	switch orderBy {
	case CompliancePortalDocumentAccessOrderFieldCreatedAt:
		return page.NewCursorKey(tcda.ID, tcda.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (tcda *CompliancePortalDocumentAccess) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM cp_document_accesses WHERE id = ANY(@resource_ids::text[])`

	args := pgx.StrictNamedArgs{
		"resource_ids": resourceIDs,
	}

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

func (tcda *CompliancePortalDocumentAccess) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	accessID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    compliance_portal_access_id,
    document_id,
    report_file_id,
    compliance_portal_file_id,
    status,
    requested_at,
    created_at,
    updated_at
FROM
    cp_document_accesses
WHERE
    %s
    AND id = @access_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"access_id": accessID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query compliance portal document access: %w", err)
	}

	access, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CompliancePortalDocumentAccess])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect compliance portal document access: %w", err)
	}

	*tcda = access

	return nil
}

func (tcda *CompliancePortalDocumentAccess) LoadByCompliancePortalAccessIDAndDocumentID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalAccessID gid.GID,
	documentID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    compliance_portal_access_id,
    document_id,
    report_file_id,
    compliance_portal_file_id,
    status,
    requested_at,
    created_at,
    updated_at
FROM
    cp_document_accesses
WHERE
    %s
    AND compliance_portal_access_id = @compliance_portal_access_id
    AND document_id = @document_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"compliance_portal_access_id": compliancePortalAccessID,
		"document_id":                 documentID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query compliance portal document access: %w", err)
	}

	access, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CompliancePortalDocumentAccess])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect compliance portal document access: %w", err)
	}

	*tcda = access

	return nil
}

func (tcda *CompliancePortalDocumentAccess) LoadByCompliancePortalAccessIDAndReportFileID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalAccessID gid.GID,
	reportFileID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    compliance_portal_access_id,
    document_id,
    report_file_id,
    compliance_portal_file_id,
    status,
    requested_at,
    created_at,
    updated_at
FROM
    cp_document_accesses
WHERE
    %s
    AND compliance_portal_access_id = @compliance_portal_access_id
    AND report_file_id = @report_file_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"compliance_portal_access_id": compliancePortalAccessID,
		"report_file_id":              reportFileID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query compliance portal document access: %w", err)
	}

	access, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CompliancePortalDocumentAccess])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect compliance portal document access: %w", err)
	}

	*tcda = access

	return nil
}

func (tcda *CompliancePortalDocumentAccess) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO cp_document_accesses (
    id,
    tenant_id,
    organization_id,
    compliance_portal_access_id,
    document_id,
    report_file_id,
    compliance_portal_file_id,
    status,
    requested_at,
    created_at,
    updated_at
) VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @compliance_portal_access_id,
    @document_id,
    @report_file_id,
    @compliance_portal_file_id,
    @status::compliance_portal_document_access_status,
    @requested_at,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                          tcda.ID,
		"tenant_id":                   scope.GetTenantID(),
		"organization_id":             tcda.OrganizationID,
		"compliance_portal_access_id": tcda.CompliancePortalAccessID,
		"document_id":                 tcda.DocumentID,
		"report_file_id":              tcda.ReportFileID,
		"compliance_portal_file_id":   tcda.CompliancePortalFileID,
		"status":                      tcda.Status,
		"requested_at":                tcda.RequestedAt,
		"created_at":                  tcda.CreatedAt,
		"updated_at":                  tcda.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" {
				switch pgErr.ConstraintName {
				case "cp_document_accesses_compliance_portal_access_id_document_id_ke",
					"cp_document_accesses_compliance_portal_access_id_report_file_id",
					"cp_document_accesses_compliance_portal_file_id_key":
					return ErrResourceAlreadyExists
				}
			}
		}

		return fmt.Errorf("cannot insert compliance portal document access: %w", err)
	}

	return nil
}

func (tcda *CompliancePortalDocumentAccess) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE cp_document_accesses SET
    status = @status::compliance_portal_document_access_status,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":         tcda.ID,
		"status":     tcda.Status,
		"updated_at": tcda.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update compliance portal document access: %w", err)
	}

	return nil
}

func (tcda *CompliancePortalDocumentAccess) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM cp_document_accesses
WHERE
    %s
    AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id": tcda.ID,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete compliance portal document access: %w", err)
	}

	return nil
}

func (tcdas *CompliancePortalDocumentAccesses) CountByCompliancePortalAccessID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalAccessID gid.GID,
) (int, error) {
	q := `
SELECT
    COUNT(id)
FROM
    cp_document_accesses
WHERE
    %s
    AND compliance_portal_access_id = @compliance_portal_access_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"compliance_portal_access_id": compliancePortalAccessID,
	}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

// CountVisitorRequestedByCompliancePortalAccessID counts access rows stamped by
// the visitor request flow (requested_at IS NOT NULL), regardless of later
// status changes. Admin-created grants leave requested_at NULL.
func (tcdas *CompliancePortalDocumentAccesses) CountVisitorRequestedByCompliancePortalAccessID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalAccessID gid.GID,
) (int, error) {
	q := `
SELECT
    COUNT(id)
FROM
    cp_document_accesses
WHERE
    %s
    AND compliance_portal_access_id = @compliance_portal_access_id
    AND requested_at IS NOT NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"compliance_portal_access_id": compliancePortalAccessID,
	}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (tcdas *CompliancePortalDocumentAccesses) CountPendingRequestByCompliancePortalAccessID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalAccessID gid.GID,
) (int, error) {
	q := `
SELECT
    COUNT(id)
FROM
    cp_document_accesses
WHERE
    %s
    AND compliance_portal_access_id = @compliance_portal_access_id
    AND status = 'REQUESTED'::compliance_portal_document_access_status
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"compliance_portal_access_id": compliancePortalAccessID,
	}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (tcdas *CompliancePortalDocumentAccesses) CountActiveByCompliancePortalAccessID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalAccessID gid.GID,
) (int, error) {
	q := `
SELECT
    COUNT(id)
FROM
    cp_document_accesses
WHERE
    %s
    AND compliance_portal_access_id = @compliance_portal_access_id
    AND status = 'GRANTED'::compliance_portal_document_access_status
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"compliance_portal_access_id": compliancePortalAccessID,
	}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (tcdas *CompliancePortalDocumentAccesses) LoadByCompliancePortalAccessID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalAccessID gid.GID,
	cursor *page.Cursor[CompliancePortalDocumentAccessOrderField],
) error {
	q := `
SELECT
    id,
    organization_id,
    compliance_portal_access_id,
    document_id,
    report_file_id,
    compliance_portal_file_id,
    status,
    requested_at,
    created_at,
    updated_at
FROM
    cp_document_accesses
WHERE
    %s
    AND compliance_portal_access_id = @compliance_portal_access_id
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{
		"compliance_portal_access_id": compliancePortalAccessID,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query compliance portal document accesses: %w", err)
	}

	accesses, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CompliancePortalDocumentAccess])
	if err != nil {
		return fmt.Errorf("cannot collect compliance portal document accesses: %w", err)
	}

	*tcdas = accesses

	return nil
}

// RerequestByDocumentIDs reactivates REJECTED/REVOKED document access rows for
// a visitor retry (status back to REQUESTED) and stamps requested_at when
// missing on those rows or on pre-existing REQUESTED rows so
// viewerHasRequestedAccess stays consistent with the mutation.
func RerequestByDocumentIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalAccessID gid.GID,
	documentIDs []gid.GID,
	requestedAt time.Time,
) error {
	if len(documentIDs) == 0 {
		return nil
	}

	q := `
UPDATE cp_document_accesses
SET
    status = @requested_status::compliance_portal_document_access_status,
    requested_at = COALESCE(requested_at, @requested_at),
    updated_at = @requested_at
WHERE
    %s
    AND compliance_portal_access_id = @compliance_portal_access_id
    AND document_id = ANY(@document_ids)
    AND (
        status = ANY(@retryable_statuses::compliance_portal_document_access_status[])
        OR (
            status = @requested_status::compliance_portal_document_access_status
            AND requested_at IS NULL
        )
    )
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"compliance_portal_access_id": compliancePortalAccessID,
		"document_ids":                documentIDs,
		"requested_at":                requestedAt,
		"requested_status":            CompliancePortalDocumentAccessStatusRequested,
		"retryable_statuses": []CompliancePortalDocumentAccessStatus{
			CompliancePortalDocumentAccessStatusRejected,
			CompliancePortalDocumentAccessStatusRevoked,
		},
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot rerequest compliance portal document accesses: %w", err)
	}

	return nil
}

func GrantByDocumentIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalAccessID gid.GID,
	documentIDs []gid.GID,
	updatedAt time.Time,
) error {
	q := `
UPDATE cp_document_accesses
SET status = 'GRANTED'::compliance_portal_document_access_status, updated_at = @updated_at
WHERE
    %s
    AND compliance_portal_access_id = @compliance_portal_access_id
    AND document_id = ANY(@document_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"compliance_portal_access_id": compliancePortalAccessID,
		"document_ids":                documentIDs,
		"updated_at":                  updatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot grant compliance portal document accesses by document IDs: %w", err)
	}

	return nil
}

func RejectOrRevokeByDocumentIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalAccessID gid.GID,
	documentIDs []gid.GID,
	updatedAt time.Time,
) error {
	q := `
UPDATE cp_document_accesses
SET
    status = CASE
        WHEN status = 'GRANTED'::compliance_portal_document_access_status THEN 'REVOKED'::compliance_portal_document_access_status
        ELSE 'REJECTED'::compliance_portal_document_access_status
    END,
    updated_at = @updated_at
WHERE
    %s
    AND compliance_portal_access_id = @compliance_portal_access_id
    AND document_id = ANY(@document_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"compliance_portal_access_id": compliancePortalAccessID,
		"document_ids":                documentIDs,
		"updated_at":                  updatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot reject compliance portal document accesses by document IDs: %w", err)
	}

	return nil
}

// RerequestByReportFileIDs reactivates REJECTED/REVOKED report access rows for
// a visitor retry and stamps unstamped REQUESTED rows. See RerequestByDocumentIDs.
func RerequestByReportFileIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalAccessID gid.GID,
	reportFileIDs []gid.GID,
	requestedAt time.Time,
) error {
	if len(reportFileIDs) == 0 {
		return nil
	}

	q := `
UPDATE cp_document_accesses
SET
    status = @requested_status::compliance_portal_document_access_status,
    requested_at = COALESCE(requested_at, @requested_at),
    updated_at = @requested_at
WHERE
    %s
    AND compliance_portal_access_id = @compliance_portal_access_id
    AND report_file_id = ANY(@report_file_ids)
    AND (
        status = ANY(@retryable_statuses::compliance_portal_document_access_status[])
        OR (
            status = @requested_status::compliance_portal_document_access_status
            AND requested_at IS NULL
        )
    )
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"compliance_portal_access_id": compliancePortalAccessID,
		"report_file_ids":             reportFileIDs,
		"requested_at":                requestedAt,
		"requested_status":            CompliancePortalDocumentAccessStatusRequested,
		"retryable_statuses": []CompliancePortalDocumentAccessStatus{
			CompliancePortalDocumentAccessStatusRejected,
			CompliancePortalDocumentAccessStatusRevoked,
		},
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot rerequest compliance portal report accesses: %w", err)
	}

	return nil
}

func GrantByReportFileIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalAccessID gid.GID,
	reportFileIDs []gid.GID,
	updatedAt time.Time,
) error {
	q := `
UPDATE cp_document_accesses
SET status = 'GRANTED'::compliance_portal_document_access_status, updated_at = @updated_at
WHERE
    %s
    AND compliance_portal_access_id = @compliance_portal_access_id
    AND report_file_id = ANY(@report_file_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"compliance_portal_access_id": compliancePortalAccessID,
		"report_file_ids":             reportFileIDs,
		"updated_at":                  updatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot grant compliance portal document accesses by report file IDs: %w", err)
	}

	return nil
}

func RejectOrRevokeByReportFileIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalAccessID gid.GID,
	reportFileIDs []gid.GID,
	updatedAt time.Time,
) error {
	q := `
UPDATE cp_document_accesses
SET
    status = CASE
        WHEN status = 'GRANTED'::compliance_portal_document_access_status THEN 'REVOKED'::compliance_portal_document_access_status
        ELSE 'REJECTED'::compliance_portal_document_access_status
    END,
    updated_at = @updated_at
WHERE
    %s
    AND compliance_portal_access_id = @compliance_portal_access_id
    AND report_file_id = ANY(@report_file_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"compliance_portal_access_id": compliancePortalAccessID,
		"report_file_ids":             reportFileIDs,
		"updated_at":                  updatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot reject compliance portal document accesses by report file IDs: %w", err)
	}

	return nil
}

type UpsertCompliancePortalDocumentAccessesData struct {
	ID     gid.GID                              `json:"id"`
	Status CompliancePortalDocumentAccessStatus `json:"status"`
}

func (tcdas CompliancePortalDocumentAccesses) UpsertDocumentAccesses(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	compliancePortalAccessID gid.GID,
	data []UpsertCompliancePortalDocumentAccessesData,
) error {
	if len(data) == 0 {
		return nil
	}

	q := `
WITH data AS (
    SELECT
        t.*
    FROM json_to_recordset(@data)
        AS t(
            id text,
            status compliance_portal_document_access_status
        )
)
INSERT INTO cp_document_accesses (
    id,
    tenant_id,
    organization_id,
    compliance_portal_access_id,
    document_id,
    report_file_id,
    compliance_portal_file_id,
    status,
    created_at,
    updated_at
)
SELECT
    generate_gid(decode_base64_unpadded(@tenant_id), @compliance_portal_document_access_entity_type),
    @tenant_id,
    @organization_id,
    @compliance_portal_access_id,
    data.id,
    NULL,
    NULL,
    data.status,
    @now::timestamptz,
    @now::timestamptz
FROM data
ON CONFLICT (compliance_portal_access_id, document_id) DO UPDATE SET
    status = EXCLUDED.status,
    updated_at = EXCLUDED.updated_at
`

	args := pgx.StrictNamedArgs{
		"compliance_portal_document_access_entity_type": CompliancePortalDocumentAccessEntityType,
		"tenant_id":                   scope.GetTenantID(),
		"compliance_portal_access_id": compliancePortalAccessID,
		"organization_id":             organizationID,
		"now":                         time.Now(),
		"data":                        data,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot upsert document accesses: %w", err)
	}

	return nil
}

func (tcdas CompliancePortalDocumentAccesses) BulkInsertDocumentAccesses(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalAccessID gid.GID,
	organizationID gid.GID,
	documentIDs []gid.GID,
	status CompliancePortalDocumentAccessStatus,
	createdAt time.Time,
) error {
	if len(documentIDs) == 0 {
		return nil
	}

	q := `
WITH document_access_data AS (
    SELECT
        generate_gid(decode_base64_unpadded(@tenant_id), @compliance_portal_document_access_entity_type) AS id,
        @tenant_id AS tenant_id,
        @organization_id AS organization_id,
        @compliance_portal_access_id AS compliance_portal_access_id,
        unnest(@document_ids::text[]) AS document_id,
        null::text AS report_file_id,
        null::text AS compliance_portal_file_id,
        @status::compliance_portal_document_access_status AS status,
        @created_at::timestamptz AS requested_at,
        @created_at::timestamptz AS created_at,
        @updated_at::timestamptz AS updated_at
)
INSERT INTO cp_document_accesses (
    id,
    tenant_id,
    organization_id,
    compliance_portal_access_id,
    document_id,
    report_file_id,
    compliance_portal_file_id,
    status,
    requested_at,
    created_at,
    updated_at
)
SELECT * FROM document_access_data
ON CONFLICT DO NOTHING
`

	args := pgx.StrictNamedArgs{
		"tenant_id":                   scope.GetTenantID(),
		"organization_id":             organizationID,
		"compliance_portal_access_id": compliancePortalAccessID,
		"document_ids":                documentIDs,
		"compliance_portal_document_access_entity_type": CompliancePortalDocumentAccessEntityType,
		"status":     status,
		"created_at": createdAt,
		"updated_at": createdAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot bulk insert compliance portal document accesses: %w", err)
	}

	return nil
}

func (tcdas CompliancePortalDocumentAccesses) UpsertReportFileAccesses(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	compliancePortalAccessID gid.GID,
	data []UpsertCompliancePortalDocumentAccessesData,
) error {
	if len(data) == 0 {
		return nil
	}

	q := `
WITH data AS (
    SELECT
        t.*
    FROM json_to_recordset(@data)
        AS t(
            id text,
            status compliance_portal_document_access_status
        )
)
INSERT INTO cp_document_accesses (
    id,
    tenant_id,
    organization_id,
    compliance_portal_access_id,
    document_id,
    report_file_id,
    compliance_portal_file_id,
    status,
    created_at,
    updated_at
)
SELECT
    generate_gid(decode_base64_unpadded(@tenant_id), @compliance_portal_document_access_entity_type),
    @tenant_id,
    @organization_id,
    @compliance_portal_access_id,
    NULL,
    data.id,
    NULL,
    data.status,
    @now::timestamptz,
    @now::timestamptz
FROM data
ON CONFLICT (compliance_portal_access_id, report_file_id) DO UPDATE SET
    status = EXCLUDED.status,
    updated_at = EXCLUDED.updated_at
`

	args := pgx.StrictNamedArgs{
		"compliance_portal_document_access_entity_type": CompliancePortalDocumentAccessEntityType,
		"tenant_id":                   scope.GetTenantID(),
		"compliance_portal_access_id": compliancePortalAccessID,
		"organization_id":             organizationID,
		"now":                         time.Now(),
		"data":                        data,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot upsert report file accesses: %w", err)
	}

	return nil
}

func (tcdas CompliancePortalDocumentAccesses) BulkInsertReportFileAccesses(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalAccessID gid.GID,
	organizationID gid.GID,
	reportFileIDs []gid.GID,
	status CompliancePortalDocumentAccessStatus,
	createdAt time.Time,
) error {
	if len(reportFileIDs) == 0 {
		return nil
	}

	q := `
WITH report_file_access_data AS (
    SELECT
        generate_gid(decode_base64_unpadded(@tenant_id), @compliance_portal_document_access_entity_type) AS id,
        @tenant_id AS tenant_id,
        @organization_id AS organization_id,
        @compliance_portal_access_id AS compliance_portal_access_id,
        null::text AS document_id,
        unnest(@report_file_ids::text[]) AS report_file_id,
        null::text AS compliance_portal_file_id,
        @status::compliance_portal_document_access_status AS status,
        @created_at::timestamptz AS requested_at,
        @created_at::timestamptz AS created_at,
        @updated_at::timestamptz AS updated_at
)
INSERT INTO cp_document_accesses (
    id,
    tenant_id,
    organization_id,
    compliance_portal_access_id,
    document_id,
    report_file_id,
    compliance_portal_file_id,
    status,
    requested_at,
    created_at,
    updated_at
)
SELECT * FROM report_file_access_data
ON CONFLICT DO NOTHING
`

	args := pgx.StrictNamedArgs{
		"tenant_id":       scope.GetTenantID(),
		"organization_id": organizationID,
		"compliance_portal_document_access_entity_type": CompliancePortalDocumentAccessEntityType,
		"compliance_portal_access_id":                   compliancePortalAccessID,
		"report_file_ids":                               reportFileIDs,
		"status":                                        status,
		"created_at":                                    createdAt,
		"updated_at":                                    createdAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot bulk insert compliance portal report file accesses: %w", err)
	}

	return nil
}

func (tcda *CompliancePortalDocumentAccess) LoadByCompliancePortalAccessIDAndCompliancePortalFileID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalAccessID gid.GID,
	compliancePortalFileID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    compliance_portal_access_id,
    document_id,
    report_file_id,
    compliance_portal_file_id,
    status,
    requested_at,
    created_at,
    updated_at
FROM
    cp_document_accesses
WHERE
    %s
    AND compliance_portal_access_id = @compliance_portal_access_id
    AND compliance_portal_file_id = @compliance_portal_file_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"compliance_portal_access_id": compliancePortalAccessID,
		"compliance_portal_file_id":   compliancePortalFileID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query compliance portal document access: %w", err)
	}

	access, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CompliancePortalDocumentAccess])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect compliance portal document access: %w", err)
	}

	*tcda = access

	return nil
}

func (tcdas *CompliancePortalDocumentAccesses) LoadByCompliancePortalAccessIDAndDocumentIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalAccessID gid.GID,
	documentIDs []gid.GID,
) error {
	if len(documentIDs) == 0 {
		*tcdas = CompliancePortalDocumentAccesses{}
		return nil
	}

	q := `
SELECT
    id,
    organization_id,
    compliance_portal_access_id,
    document_id,
    report_file_id,
    compliance_portal_file_id,
    status,
    requested_at,
    created_at,
    updated_at
FROM
    cp_document_accesses
WHERE
    %s
    AND compliance_portal_access_id = @compliance_portal_access_id
    AND document_id = ANY(@document_ids::text[]);
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"compliance_portal_access_id": compliancePortalAccessID,
		"document_ids":                documentIDs,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query compliance portal document accesses: %w", err)
	}

	accesses, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CompliancePortalDocumentAccess])
	if err != nil {
		return fmt.Errorf("cannot collect compliance portal document accesses: %w", err)
	}

	*tcdas = accesses

	return nil
}

func (tcdas *CompliancePortalDocumentAccesses) LoadByCompliancePortalAccessIDAndReportFileIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalAccessID gid.GID,
	reportFileIDs []gid.GID,
) error {
	if len(reportFileIDs) == 0 {
		*tcdas = CompliancePortalDocumentAccesses{}
		return nil
	}

	q := `
SELECT
    id,
    organization_id,
    compliance_portal_access_id,
    document_id,
    report_file_id,
    compliance_portal_file_id,
    status,
    requested_at,
    created_at,
    updated_at
FROM
    cp_document_accesses
WHERE
    %s
    AND compliance_portal_access_id = @compliance_portal_access_id
    AND report_file_id = ANY(@report_file_ids::text[]);
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"compliance_portal_access_id": compliancePortalAccessID,
		"report_file_ids":             reportFileIDs,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query compliance portal document accesses: %w", err)
	}

	accesses, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CompliancePortalDocumentAccess])
	if err != nil {
		return fmt.Errorf("cannot collect compliance portal document accesses: %w", err)
	}

	*tcdas = accesses

	return nil
}

func (tcdas *CompliancePortalDocumentAccesses) LoadByCompliancePortalAccessIDAndCompliancePortalFileIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalAccessID gid.GID,
	compliancePortalFileIDs []gid.GID,
) error {
	if len(compliancePortalFileIDs) == 0 {
		*tcdas = CompliancePortalDocumentAccesses{}
		return nil
	}

	q := `
SELECT
    id,
    organization_id,
    compliance_portal_access_id,
    document_id,
    report_file_id,
    compliance_portal_file_id,
    status,
    requested_at,
    created_at,
    updated_at
FROM
    cp_document_accesses
WHERE
    %s
    AND compliance_portal_access_id = @compliance_portal_access_id
    AND compliance_portal_file_id = ANY(@compliance_portal_file_ids::text[]);
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"compliance_portal_access_id": compliancePortalAccessID,
		"compliance_portal_file_ids":  compliancePortalFileIDs,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query compliance portal document accesses: %w", err)
	}

	accesses, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CompliancePortalDocumentAccess])
	if err != nil {
		return fmt.Errorf("cannot collect compliance portal document accesses: %w", err)
	}

	*tcdas = accesses

	return nil
}

// RerequestByCompliancePortalFileIDs reactivates REJECTED/REVOKED file access
// rows for a visitor retry and stamps unstamped REQUESTED rows. See
// RerequestByDocumentIDs.
func RerequestByCompliancePortalFileIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalAccessID gid.GID,
	compliancePortalFileIDs []gid.GID,
	requestedAt time.Time,
) error {
	if len(compliancePortalFileIDs) == 0 {
		return nil
	}

	q := `
UPDATE cp_document_accesses
SET
    status = @requested_status::compliance_portal_document_access_status,
    requested_at = COALESCE(requested_at, @requested_at),
    updated_at = @requested_at
WHERE
    %s
    AND compliance_portal_access_id = @compliance_portal_access_id
    AND compliance_portal_file_id = ANY(@compliance_portal_file_ids)
    AND (
        status = ANY(@retryable_statuses::compliance_portal_document_access_status[])
        OR (
            status = @requested_status::compliance_portal_document_access_status
            AND requested_at IS NULL
        )
    )
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"compliance_portal_access_id": compliancePortalAccessID,
		"compliance_portal_file_ids":  compliancePortalFileIDs,
		"requested_at":                requestedAt,
		"requested_status":            CompliancePortalDocumentAccessStatusRequested,
		"retryable_statuses": []CompliancePortalDocumentAccessStatus{
			CompliancePortalDocumentAccessStatusRejected,
			CompliancePortalDocumentAccessStatusRevoked,
		},
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot rerequest compliance portal file accesses: %w", err)
	}

	return nil
}

func GrantByCompliancePortalFileIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalAccessID gid.GID,
	compliancePortalFileIDs []gid.GID,
	updatedAt time.Time,
) error {
	q := `
UPDATE cp_document_accesses
SET status = 'GRANTED'::compliance_portal_document_access_status, updated_at = @updated_at
WHERE
    %s
    AND compliance_portal_access_id = @compliance_portal_access_id
    AND compliance_portal_file_id = ANY(@compliance_portal_file_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"compliance_portal_access_id": compliancePortalAccessID,
		"compliance_portal_file_ids":  compliancePortalFileIDs,
		"updated_at":                  updatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot grant compliance portal document accesses by compliance portal file IDs: %w", err)
	}

	return nil
}

func RejectOrRevokeByCompliancePortalFileIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalAccessID gid.GID,
	compliancePortalFileIDs []gid.GID,
	updatedAt time.Time,
) error {
	q := `
UPDATE cp_document_accesses
SET
    status = CASE
        WHEN status = 'GRANTED'::compliance_portal_document_access_status THEN 'REVOKED'::compliance_portal_document_access_status
        ELSE 'REJECTED'::compliance_portal_document_access_status
    END,
    updated_at = @updated_at
WHERE
    %s
    AND compliance_portal_access_id = @compliance_portal_access_id
    AND compliance_portal_file_id = ANY(@compliance_portal_file_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"compliance_portal_access_id": compliancePortalAccessID,
		"compliance_portal_file_ids":  compliancePortalFileIDs,
		"updated_at":                  updatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot reject compliance portal document accesses by compliance portal file IDs: %w", err)
	}

	return nil
}

func (tcdas CompliancePortalDocumentAccesses) UpsertCompliancePortalFileAccesses(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	compliancePortalAccessID gid.GID,
	data []UpsertCompliancePortalDocumentAccessesData,
) error {
	if len(data) == 0 {
		return nil
	}

	q := `
WITH data AS (
    SELECT
        t.*
    FROM json_to_recordset(@data)
        AS t(
            id text,
            status compliance_portal_document_access_status
        )
)
INSERT INTO cp_document_accesses (
    id,
    tenant_id,
    organization_id,
    compliance_portal_access_id,
    document_id,
    report_file_id,
    compliance_portal_file_id,
    status,
    created_at,
    updated_at
)
SELECT
    generate_gid(decode_base64_unpadded(@tenant_id), @compliance_portal_document_access_entity_type),
    @tenant_id,
    @organization_id,
    @compliance_portal_access_id,
    NULL,
    NULL,
    data.id,
    data.status,
    @now::timestamptz,
    @now::timestamptz
FROM data
ON CONFLICT (compliance_portal_access_id, compliance_portal_file_id) DO UPDATE SET
    status = EXCLUDED.status,
    updated_at = EXCLUDED.updated_at
`

	args := pgx.StrictNamedArgs{
		"compliance_portal_document_access_entity_type": CompliancePortalDocumentAccessEntityType,
		"tenant_id":                   scope.GetTenantID(),
		"compliance_portal_access_id": compliancePortalAccessID,
		"organization_id":             organizationID,
		"now":                         time.Now(),
		"data":                        data,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot upsert compliance portal file accesses: %w", err)
	}

	return nil
}

func (tcdas CompliancePortalDocumentAccesses) BulkInsertCompliancePortalFileAccesses(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalAccessID gid.GID,
	organizationID gid.GID,
	compliancePortalFileIDs []gid.GID,
	status CompliancePortalDocumentAccessStatus,
	createdAt time.Time,
) error {
	q := `
WITH compliance_portal_file_access_data AS (
    SELECT
        generate_gid(decode_base64_unpadded(@tenant_id), @compliance_portal_document_access_entity_type) AS id,
        @tenant_id AS tenant_id,
        @organization_id AS organization_id,
        @compliance_portal_access_id AS compliance_portal_access_id,
        null::text AS document_id,
        null::text AS report_file_id,
        unnest(@compliance_portal_file_ids::text[]) AS compliance_portal_file_id,
        @status::compliance_portal_document_access_status AS status,
        @created_at::timestamptz AS requested_at,
        @created_at::timestamptz AS created_at,
        @updated_at::timestamptz AS updated_at
)
INSERT INTO cp_document_accesses (
    id,
    tenant_id,
    organization_id,
    compliance_portal_access_id,
    document_id,
    report_file_id,
    compliance_portal_file_id,
    status,
    requested_at,
    created_at,
    updated_at
)
SELECT * FROM compliance_portal_file_access_data
ON CONFLICT DO NOTHING
`

	args := pgx.StrictNamedArgs{
		"tenant_id":       scope.GetTenantID(),
		"organization_id": organizationID,
		"compliance_portal_document_access_entity_type": CompliancePortalDocumentAccessEntityType,
		"compliance_portal_access_id":                   compliancePortalAccessID,
		"compliance_portal_file_ids":                    compliancePortalFileIDs,
		"status":                                        status,
		"created_at":                                    createdAt,
		"updated_at":                                    createdAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot bulk insert compliance portal file accesses: %w", err)
	}

	return nil
}
