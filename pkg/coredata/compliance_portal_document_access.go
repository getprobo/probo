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
		CompliancePortalAccessID gid.GID                              `db:"trust_center_access_id"`
		DocumentID               *gid.GID                             `db:"document_id"`
		ReportFileID             *gid.GID                             `db:"report_file_id"`
		CompliancePortalFileID   *gid.GID                             `db:"trust_center_file_id"`
		Status                   CompliancePortalDocumentAccessStatus `db:"status"`
		CreatedAt                time.Time                            `db:"created_at"`
		UpdatedAt                time.Time                            `db:"updated_at"`
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
	q := `SELECT id, organization_id FROM trust_center_document_accesses WHERE id = ANY(@resource_ids::text[])`

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
    trust_center_access_id,
    document_id,
    report_file_id,
    trust_center_file_id,
    status,
    created_at,
    updated_at
FROM
    trust_center_document_accesses
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
    trust_center_access_id,
    document_id,
    report_file_id,
    trust_center_file_id,
    status,
    created_at,
    updated_at
FROM
    trust_center_document_accesses
WHERE
    %s
    AND trust_center_access_id = @trust_center_access_id
    AND document_id = @document_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_access_id": compliancePortalAccessID,
		"document_id":            documentID,
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
    trust_center_access_id,
    document_id,
    report_file_id,
    trust_center_file_id,
    status,
    created_at,
    updated_at
FROM
    trust_center_document_accesses
WHERE
    %s
    AND trust_center_access_id = @trust_center_access_id
    AND report_file_id = @report_file_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_access_id": compliancePortalAccessID,
		"report_file_id":         reportFileID,
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
INSERT INTO trust_center_document_accesses (
    id,
    tenant_id,
    organization_id,
    trust_center_access_id,
    document_id,
    report_file_id,
    trust_center_file_id,
    status,
    created_at,
    updated_at
) VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @trust_center_access_id,
    @document_id,
    @report_file_id,
    @trust_center_file_id,
    @status::trust_center_document_access_status,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                     tcda.ID,
		"tenant_id":              scope.GetTenantID(),
		"organization_id":        tcda.OrganizationID,
		"trust_center_access_id": tcda.CompliancePortalAccessID,
		"document_id":            tcda.DocumentID,
		"report_file_id":         tcda.ReportFileID,
		"trust_center_file_id":   tcda.CompliancePortalFileID,
		"status":                 tcda.Status,
		"created_at":             tcda.CreatedAt,
		"updated_at":             tcda.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" {
				switch pgErr.ConstraintName {
				case "trust_center_document_accesse_trust_center_access_id_docume_key",
					"trust_center_document_accesses_trust_center_access_id_report_file_key",
					"trust_center_document_accesses_trust_center_file_id_key":
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
UPDATE trust_center_document_accesses SET
    status = @status::trust_center_document_access_status,
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
DELETE FROM trust_center_document_accesses
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
    trust_center_document_accesses
WHERE
    %s
    AND trust_center_access_id = @trust_center_access_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_access_id": compliancePortalAccessID,
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
    trust_center_document_accesses
WHERE
    %s
    AND trust_center_access_id = @trust_center_access_id
    AND status = 'REQUESTED'::trust_center_document_access_status
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_access_id": compliancePortalAccessID,
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
    trust_center_document_accesses
WHERE
    %s
    AND trust_center_access_id = @trust_center_access_id
    AND status = 'GRANTED'::trust_center_document_access_status
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_access_id": compliancePortalAccessID,
	}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (tcdas *CompliancePortalDocumentAccesses) LoadAvailableByCompliancePortalAccessID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalAccessID gid.GID,
	cursor *page.Cursor[CompliancePortalDocumentAccessOrderField],
) error {
	q := `
WITH organization AS (
    SELECT tc.organization_id
    FROM trust_center_accesses tca
    INNER JOIN trust_centers tc ON tca.trust_center_id = tc.id
    WHERE tca.tenant_id = @tenant_id
        AND tca.id = @trust_center_access_id
),
tenant_organization AS (
    SELECT o.id AS organization_id
    FROM organizations o
    WHERE %s
),
all_items AS (
    SELECT
        d.id AS item_id,
        d.id AS document_id,
        NULL::text AS report_file_id,
        NULL::text AS trust_center_file_id,
        d.created_at AS item_created_at,
        d.updated_at AS item_updated_at
    FROM documents d, tenant_organization o
    WHERE d.organization_id = o.organization_id
        AND d.deleted_at IS NULL
        AND d.trust_center_visibility = 'PRIVATE'::trust_center_visibility
        AND d.current_published_major IS NOT NULL

    UNION ALL

    SELECT
        r.report_file_id AS item_id,
        NULL::text AS document_id,
        r.report_file_id AS report_file_id,
        NULL::text AS trust_center_file_id,
        r.created_at AS item_created_at,
        r.updated_at AS item_updated_at
    FROM audits r, tenant_organization o
    WHERE r.organization_id = o.organization_id
        AND r.trust_center_visibility = 'PRIVATE'::trust_center_visibility
        AND r.report_file_id IS NOT NULL

    UNION ALL

    SELECT
        tcf.id AS item_id,
        NULL::text AS document_id,
        NULL::text AS report_file_id,
        tcf.id AS trust_center_file_id,
        tcf.created_at AS item_created_at,
        tcf.updated_at AS item_updated_at
    FROM trust_center_files tcf, tenant_organization o
    WHERE tcf.organization_id = o.organization_id
        AND (
            tcf.trust_center_visibility = 'PRIVATE'::trust_center_visibility
            OR tcf.trust_center_visibility = 'NONE'::trust_center_visibility
        )
),
final_items AS (
  SELECT
      COALESCE(tcda.id, ai.item_id) AS id,
      tcda.tenant_id,
      (SELECT organization_id FROM organization) AS organization_id,
      @trust_center_access_id AS trust_center_access_id,
      ai.document_id,
      ai.report_file_id,
      ai.trust_center_file_id,
      COALESCE(tcda.status, 'REQUESTED'::trust_center_document_access_status) AS status,
      COALESCE(tcda.created_at, ai.item_created_at) AS created_at,
      COALESCE(tcda.updated_at, ai.item_updated_at) AS updated_at
  FROM all_items ai
  LEFT JOIN trust_center_document_accesses tcda ON (
      tcda.trust_center_access_id = @trust_center_access_id
      AND (
          (tcda.document_id = ai.document_id AND ai.document_id IS NOT NULL)
          OR (tcda.report_file_id = ai.report_file_id AND ai.report_file_id IS NOT NULL)
          OR (tcda.trust_center_file_id = ai.trust_center_file_id AND ai.trust_center_file_id IS NOT NULL)
      )
  )
)
SELECT
    id,
    organization_id,
    trust_center_access_id,
    document_id,
    report_file_id,
    trust_center_file_id,
    status,
    created_at,
    updated_at
FROM final_items
WHERE %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_access_id": compliancePortalAccessID,
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
    trust_center_access_id,
    document_id,
    report_file_id,
    trust_center_file_id,
    status,
    created_at,
    updated_at
FROM
    trust_center_document_accesses
WHERE
    %s
    AND trust_center_access_id = @trust_center_access_id
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_access_id": compliancePortalAccessID,
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

func GrantByDocumentIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalAccessID gid.GID,
	documentIDs []gid.GID,
	updatedAt time.Time,
) error {
	q := `
UPDATE trust_center_document_accesses
SET status = 'GRANTED'::trust_center_document_access_status, updated_at = @updated_at
WHERE
    %s
    AND trust_center_access_id = @trust_center_access_id
    AND document_id = ANY(@document_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_access_id": compliancePortalAccessID,
		"document_ids":           documentIDs,
		"updated_at":             updatedAt,
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
UPDATE trust_center_document_accesses
SET
    status = CASE
        WHEN status = 'GRANTED'::trust_center_document_access_status THEN 'REVOKED'::trust_center_document_access_status
        ELSE 'REJECTED'::trust_center_document_access_status
    END,
    updated_at = @updated_at
WHERE
    %s
    AND trust_center_access_id = @trust_center_access_id
    AND document_id = ANY(@document_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_access_id": compliancePortalAccessID,
		"document_ids":           documentIDs,
		"updated_at":             updatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot reject compliance portal document accesses by document IDs: %w", err)
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
UPDATE trust_center_document_accesses
SET status = 'GRANTED'::trust_center_document_access_status, updated_at = @updated_at
WHERE
    %s
    AND trust_center_access_id = @trust_center_access_id
    AND report_file_id = ANY(@report_file_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_access_id": compliancePortalAccessID,
		"report_file_ids":        reportFileIDs,
		"updated_at":             updatedAt,
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
UPDATE trust_center_document_accesses
SET
    status = CASE
        WHEN status = 'GRANTED'::trust_center_document_access_status THEN 'REVOKED'::trust_center_document_access_status
        ELSE 'REJECTED'::trust_center_document_access_status
    END,
    updated_at = @updated_at
WHERE
    %s
    AND trust_center_access_id = @trust_center_access_id
    AND report_file_id = ANY(@report_file_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_access_id": compliancePortalAccessID,
		"report_file_ids":        reportFileIDs,
		"updated_at":             updatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot reject compliance portal document accesses by report file IDs: %w", err)
	}

	return nil
}

type MergeCompliancePortalDocumentAccessesData struct {
	ID     gid.GID                              `json:"id"`
	Status CompliancePortalDocumentAccessStatus `json:"status"`
}

func (tcdas CompliancePortalDocumentAccesses) MergeDocumentAccesses(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	compliancePortalAccessID gid.GID,
	data []MergeCompliancePortalDocumentAccessesData,
) error {
	q := `
WITH data AS (
    SELECT
        t.*
    FROM json_to_recordset(@data)
        AS t(
            id text,
            status trust_center_document_access_status
        )
)
MERGE INTO trust_center_document_accesses AS tcda
USING data
    ON data.id = tcda.document_id
    AND tcda.tenant_id = @tenant_id
    AND tcda.trust_center_access_id = @trust_center_access_id
WHEN MATCHED
    THEN UPDATE SET status = data.status, updated_at = @now::timestamptz
WHEN NOT MATCHED BY SOURCE
    AND tcda.tenant_id = @tenant_id
    AND tcda.trust_center_access_id = @trust_center_access_id
    AND tcda.document_id IS NOT NULL
    THEN DELETE
WHEN NOT MATCHED
    THEN INSERT (
        id,
        tenant_id,
        organization_id,
        trust_center_access_id,
        document_id,
        report_file_id,
        trust_center_file_id,
        status,
        created_at,
        updated_at
    )
    VALUES (
        generate_gid(decode_base64_unpadded(@tenant_id), @trust_center_document_access_entity_type),
        @tenant_id,
        @organization_id,
        @trust_center_access_id,
        data.id,
        NULL,
        NULL,
        data.status,
        @now::timestamptz,
        @now::timestamptz
    )
`

	args := pgx.StrictNamedArgs{
		"trust_center_document_access_entity_type": CompliancePortalDocumentAccessEntityType,
		"tenant_id":              scope.GetTenantID(),
		"trust_center_access_id": compliancePortalAccessID,
		"organization_id":        organizationID,
		"now":                    time.Now(),
		"data":                   data,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return err
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
        generate_gid(decode_base64_unpadded(@tenant_id), @trust_center_document_access_entity_type) AS id,
        @tenant_id AS tenant_id,
        @organization_id AS organization_id,
        @trust_center_access_id AS trust_center_access_id,
        unnest(@document_ids::text[]) AS document_id,
        null::text AS report_file_id,
        null::text AS trust_center_file_id,
        @status::trust_center_document_access_status AS status,
        @created_at::timestamptz AS created_at,
        @updated_at::timestamptz AS updated_at
)
INSERT INTO trust_center_document_accesses (
    id,
    tenant_id,
    organization_id,
    trust_center_access_id,
    document_id,
    report_file_id,
    trust_center_file_id,
    status,
    created_at,
    updated_at
)
SELECT * FROM document_access_data
ON CONFLICT DO NOTHING
`

	args := pgx.StrictNamedArgs{
		"tenant_id":              scope.GetTenantID(),
		"organization_id":        organizationID,
		"trust_center_access_id": compliancePortalAccessID,
		"document_ids":           documentIDs,
		"trust_center_document_access_entity_type": CompliancePortalDocumentAccessEntityType,
		"status":     status,
		"created_at": createdAt,
		"updated_at": createdAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot bulk insert compliance portal document accesses: %w", err)
	}

	return nil
}

func (tcdas CompliancePortalDocumentAccesses) MergeReportFileAccesses(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	compliancePortalAccessID gid.GID,
	data []MergeCompliancePortalDocumentAccessesData,
) error {
	q := `
WITH data AS (
    SELECT
        t.*
    FROM json_to_recordset(@data)
        AS t(
            id text,
            status trust_center_document_access_status
        )
)
MERGE INTO trust_center_document_accesses AS tcda
USING data
    ON data.id = tcda.report_file_id
    AND tcda.tenant_id = @tenant_id
    AND tcda.trust_center_access_id = @trust_center_access_id
WHEN MATCHED
    THEN UPDATE SET status = data.status, updated_at = @now::timestamptz
WHEN NOT MATCHED BY SOURCE
    AND tcda.tenant_id = @tenant_id
    AND tcda.trust_center_access_id = @trust_center_access_id
    AND tcda.report_file_id IS NOT NULL
    THEN DELETE
WHEN NOT MATCHED
    THEN INSERT (
        id,
        tenant_id,
        organization_id,
        trust_center_access_id,
        document_id,
        report_file_id,
        trust_center_file_id,
        status,
        created_at,
        updated_at
    )
    VALUES (
        generate_gid(decode_base64_unpadded(@tenant_id), @trust_center_document_access_entity_type),
        @tenant_id,
        @organization_id,
        @trust_center_access_id,
        NULL,
        data.id,
        NULL,
        data.status,
        @now::timestamptz,
        @now::timestamptz
    )
`

	args := pgx.StrictNamedArgs{
		"trust_center_document_access_entity_type": CompliancePortalDocumentAccessEntityType,
		"tenant_id":              scope.GetTenantID(),
		"trust_center_access_id": compliancePortalAccessID,
		"organization_id":        organizationID,
		"now":                    time.Now(),
		"data":                   data,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return err
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
        generate_gid(decode_base64_unpadded(@tenant_id), @trust_center_document_access_entity_type) AS id,
        @tenant_id AS tenant_id,
        @organization_id AS organization_id,
        @trust_center_access_id AS trust_center_access_id,
        null::text AS document_id,
        unnest(@report_file_ids::text[]) AS report_file_id,
        null::text AS trust_center_file_id,
        @status::trust_center_document_access_status AS status,
        @created_at::timestamptz AS created_at,
        @updated_at::timestamptz AS updated_at
)
INSERT INTO trust_center_document_accesses (
    id,
    tenant_id,
    organization_id,
    trust_center_access_id,
    document_id,
    report_file_id,
    trust_center_file_id,
    status,
    created_at,
    updated_at
)
SELECT * FROM report_file_access_data
ON CONFLICT DO NOTHING
`

	args := pgx.StrictNamedArgs{
		"tenant_id":       scope.GetTenantID(),
		"organization_id": organizationID,
		"trust_center_document_access_entity_type": CompliancePortalDocumentAccessEntityType,
		"trust_center_access_id":                   compliancePortalAccessID,
		"report_file_ids":                          reportFileIDs,
		"status":                                   status,
		"created_at":                               createdAt,
		"updated_at":                               createdAt,
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
    trust_center_access_id,
    document_id,
    report_file_id,
    trust_center_file_id,
    status,
    created_at,
    updated_at
FROM
    trust_center_document_accesses
WHERE
    %s
    AND trust_center_access_id = @trust_center_access_id
    AND trust_center_file_id = @trust_center_file_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_access_id": compliancePortalAccessID,
		"trust_center_file_id":   compliancePortalFileID,
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

func GrantByCompliancePortalFileIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalAccessID gid.GID,
	compliancePortalFileIDs []gid.GID,
	updatedAt time.Time,
) error {
	q := `
UPDATE trust_center_document_accesses
SET status = 'GRANTED'::trust_center_document_access_status, updated_at = @updated_at
WHERE
    %s
    AND trust_center_access_id = @trust_center_access_id
    AND trust_center_file_id = ANY(@trust_center_file_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_access_id": compliancePortalAccessID,
		"trust_center_file_ids":  compliancePortalFileIDs,
		"updated_at":             updatedAt,
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
UPDATE trust_center_document_accesses
SET
    status = CASE
        WHEN status = 'GRANTED'::trust_center_document_access_status THEN 'REVOKED'::trust_center_document_access_status
        ELSE 'REJECTED'::trust_center_document_access_status
    END,
    updated_at = @updated_at
WHERE
    %s
    AND trust_center_access_id = @trust_center_access_id
    AND trust_center_file_id = ANY(@trust_center_file_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_access_id": compliancePortalAccessID,
		"trust_center_file_ids":  compliancePortalFileIDs,
		"updated_at":             updatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot reject compliance portal document accesses by compliance portal file IDs: %w", err)
	}

	return nil
}

func (tcdas CompliancePortalDocumentAccesses) MergeCompliancePortalFileAccesses(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	compliancePortalAccessID gid.GID,
	data []MergeCompliancePortalDocumentAccessesData,
) error {
	q := `
WITH data AS (
    SELECT
        t.*
    FROM json_to_recordset(@data)
        AS t(
            id text,
            status trust_center_document_access_status
        )
)
MERGE INTO trust_center_document_accesses AS tcda
USING data
    ON data.id = tcda.trust_center_file_id
    AND tcda.tenant_id = @tenant_id
    AND tcda.trust_center_access_id = @trust_center_access_id
WHEN MATCHED
    THEN UPDATE SET status = data.status, updated_at = @now::timestamptz
WHEN NOT MATCHED BY SOURCE
    AND tcda.tenant_id = @tenant_id
    AND tcda.trust_center_access_id = @trust_center_access_id
    AND tcda.trust_center_file_id IS NOT NULL
    THEN DELETE
WHEN NOT MATCHED
    THEN INSERT (
        id,
        tenant_id,
        organization_id,
        trust_center_access_id,
        document_id,
        report_file_id,
        trust_center_file_id,
        status,
        created_at,
        updated_at
    )
    VALUES (
        generate_gid(decode_base64_unpadded(@tenant_id), @trust_center_document_access_entity_type),
        @tenant_id,
        @organization_id,
        @trust_center_access_id,
        NULL,
        NULL,
        data.id,
        data.status,
        @now::timestamptz,
        @now::timestamptz
    )
`

	args := pgx.StrictNamedArgs{
		"trust_center_document_access_entity_type": CompliancePortalDocumentAccessEntityType,
		"tenant_id":              scope.GetTenantID(),
		"trust_center_access_id": compliancePortalAccessID,
		"organization_id":        organizationID,
		"now":                    time.Now(),
		"data":                   data,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return err
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
WITH trust_center_file_access_data AS (
    SELECT
        generate_gid(decode_base64_unpadded(@tenant_id), @trust_center_document_access_entity_type) AS id,
        @tenant_id AS tenant_id,
        @organization_id AS organization_id,
        @trust_center_access_id AS trust_center_access_id,
        null::text AS document_id,
        null::text AS report_file_id,
        unnest(@trust_center_file_ids::text[]) AS trust_center_file_id,
        @status::trust_center_document_access_status AS status,
        @created_at::timestamptz AS created_at,
        @updated_at::timestamptz AS updated_at
)
INSERT INTO trust_center_document_accesses (
    id,
    tenant_id,
    organization_id,
    trust_center_access_id,
    document_id,
    report_file_id,
    trust_center_file_id,
    status,
    created_at,
    updated_at
)
SELECT * FROM trust_center_file_access_data
ON CONFLICT DO NOTHING
`

	args := pgx.StrictNamedArgs{
		"tenant_id":       scope.GetTenantID(),
		"organization_id": organizationID,
		"trust_center_document_access_entity_type": CompliancePortalDocumentAccessEntityType,
		"trust_center_access_id":                   compliancePortalAccessID,
		"trust_center_file_ids":                    compliancePortalFileIDs,
		"status":                                   status,
		"created_at":                               createdAt,
		"updated_at":                               createdAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot bulk insert compliance portal file accesses: %w", err)
	}

	return nil
}
