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
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	DocumentVersionApprovalQuorum struct {
		ID              gid.GID                             `db:"id"`
		OrganizationID  gid.GID                             `db:"organization_id"`
		VersionID       gid.GID                             `db:"version_id"`
		FileID          *gid.GID                            `db:"file_id"`
		PdfAttemptCount int                                 `db:"pdf_attempt_count"`
		PdfClaimedAt    *time.Time                          `db:"pdf_claimed_at"`
		Status          DocumentVersionApprovalQuorumStatus `db:"status"`
		CreatedAt       time.Time                           `db:"created_at"`
		UpdatedAt       time.Time                           `db:"updated_at"`
	}

	DocumentVersionApprovalQuorums []*DocumentVersionApprovalQuorum
)

func (q DocumentVersionApprovalQuorum) CursorKey(orderBy DocumentVersionApprovalQuorumOrderField) page.CursorKey {
	switch orderBy {
	case DocumentVersionApprovalQuorumOrderFieldCreatedAt:
		return page.NewCursorKey(q.ID, q.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (q *DocumentVersionApprovalQuorum) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	query := `SELECT id, organization_id FROM document_version_approval_quorums WHERE id = ANY(@resource_ids::text[])`

	args := pgx.StrictNamedArgs{
		"resource_ids": resourceIDs,
	}

	rows, err := conn.Query(ctx, query, args)
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

func (q *DocumentVersionApprovalQuorum) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	query := `
SELECT
	id,
	organization_id,
	version_id,
	file_id,
	pdf_attempt_count,
	pdf_claimed_at,
	status,
	created_at,
	updated_at
FROM
	document_version_approval_quorums
WHERE
	id = @id
	AND %s
`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot query approval quorum: %w", err)
	}

	quorum, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[DocumentVersionApprovalQuorum])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect approval quorum: %w", err)
	}

	*q = quorum

	return nil
}

func (q *DocumentVersionApprovalQuorums) LoadByIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	quorumIDs []gid.GID,
) error {
	query := `
SELECT
	id,
	organization_id,
	version_id,
	file_id,
	pdf_attempt_count,
	pdf_claimed_at,
	status,
	created_at,
	updated_at
FROM
	document_version_approval_quorums
WHERE
	%s
	AND id = ANY(@quorum_ids)
`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"quorum_ids": quorumIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot query approval quorums: %w", err)
	}

	quorums, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[DocumentVersionApprovalQuorum])
	if err != nil {
		return fmt.Errorf("cannot collect approval quorums: %w", err)
	}

	*q = quorums

	if len(quorums) != len(gid.NewSet(quorumIDs...)) {
		return ErrResourceNotFound
	}

	return nil
}

func (q *DocumentVersionApprovalQuorum) LoadLastByDocumentVersionID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentVersionID gid.GID,
) error {
	query := `
WITH source_version AS (
	SELECT document_id, major FROM document_versions WHERE id = @version_id
),
major_versions AS (
	SELECT dv.id FROM document_versions dv
	INNER JOIN source_version sv ON dv.document_id = sv.document_id AND dv.major = sv.major
)
SELECT
	document_version_approval_quorums.id,
	document_version_approval_quorums.organization_id,
	document_version_approval_quorums.version_id,
	document_version_approval_quorums.file_id,
	document_version_approval_quorums.pdf_attempt_count,
	document_version_approval_quorums.pdf_claimed_at,
	document_version_approval_quorums.status,
	document_version_approval_quorums.created_at,
	document_version_approval_quorums.updated_at
FROM
	document_version_approval_quorums
INNER JOIN major_versions mv ON document_version_approval_quorums.version_id = mv.id
WHERE
	%s
ORDER BY document_version_approval_quorums.created_at DESC
LIMIT 1
`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"version_id": documentVersionID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot query last approval quorum: %w", err)
	}

	quorum, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[DocumentVersionApprovalQuorum])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect last approval quorum: %w", err)
	}

	*q = quorum

	return nil
}

func (q *DocumentVersionApprovalQuorums) LoadAllByDocumentVersionID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentVersionID gid.GID,
	cursor *page.Cursor[DocumentVersionApprovalQuorumOrderField],
) error {
	query := `
WITH source_version AS (
	SELECT document_id, major FROM document_versions WHERE id = @version_id
),
major_versions AS (
	SELECT dv.id FROM document_versions dv
	INNER JOIN source_version sv ON dv.document_id = sv.document_id AND dv.major = sv.major
)
SELECT
	document_version_approval_quorums.id,
	document_version_approval_quorums.organization_id,
	document_version_approval_quorums.version_id,
	document_version_approval_quorums.file_id,
	document_version_approval_quorums.pdf_attempt_count,
	document_version_approval_quorums.pdf_claimed_at,
	document_version_approval_quorums.status,
	document_version_approval_quorums.created_at,
	document_version_approval_quorums.updated_at
FROM
	document_version_approval_quorums
INNER JOIN major_versions mv ON document_version_approval_quorums.version_id = mv.id
WHERE
	%s
	AND %s
`

	query = fmt.Sprintf(query, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"version_id": documentVersionID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot query approval quorums: %w", err)
	}

	quorums, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[DocumentVersionApprovalQuorum])
	if err != nil {
		return fmt.Errorf("cannot collect approval quorums: %w", err)
	}

	*q = quorums

	return nil
}

func (q *DocumentVersionApprovalQuorums) CountByDocumentVersionID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentVersionID gid.GID,
) (int, error) {
	query := `
WITH source_version AS (
	SELECT document_id, major FROM document_versions WHERE id = @version_id
),
major_versions AS (
	SELECT dv.id FROM document_versions dv
	INNER JOIN source_version sv ON dv.document_id = sv.document_id AND dv.major = sv.major
)
SELECT
	COUNT(document_version_approval_quorums.id)
FROM
	document_version_approval_quorums
INNER JOIN major_versions mv ON document_version_approval_quorums.version_id = mv.id
WHERE
	%s
`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"version_id": documentVersionID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, query, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (q *DocumentVersionApprovalQuorum) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	query := `
INSERT INTO document_version_approval_quorums (
	id,
	tenant_id,
	organization_id,
	version_id,
	file_id,
	pdf_attempt_count,
	pdf_claimed_at,
	status,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@version_id,
	@file_id,
	@pdf_attempt_count,
	@pdf_claimed_at,
	@status,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                q.ID,
		"tenant_id":         scope.GetTenantID(),
		"organization_id":   q.OrganizationID,
		"version_id":        q.VersionID,
		"file_id":           q.FileID,
		"pdf_attempt_count": q.PdfAttemptCount,
		"pdf_claimed_at":    q.PdfClaimedAt,
		"status":            q.Status,
		"created_at":        q.CreatedAt,
		"updated_at":        q.UpdatedAt,
	}

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "document_one_pending_quorum_idx" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert approval quorum: %w", err)
	}

	return nil
}

func (q *DocumentVersionApprovalQuorum) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	query := `
DELETE FROM document_version_approval_quorums
WHERE
	%s
	AND id = @id
`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": q.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot delete approval quorum: %w", err)
	}

	return nil
}

func (q *DocumentVersionApprovalQuorum) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	query := `
UPDATE document_version_approval_quorums
SET
	status = @status,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":         q.ID,
		"status":     q.Status,
		"updated_at": q.UpdatedAt,
	}

	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot update approval quorum: %w", err)
	}

	return nil
}

func (q *DocumentVersionApprovalQuorum) ClaimNextWithoutFileForUpdate(
	ctx context.Context,
	conn pg.Tx,
	maxAttempts int,
	now time.Time,
	lease time.Duration,
) error {
	query := `
SELECT
	q.id,
	q.organization_id,
	q.version_id,
	q.file_id,
	q.pdf_attempt_count,
	q.pdf_claimed_at,
	q.status,
	q.created_at,
	q.updated_at
FROM
	document_version_approval_quorums q
INNER JOIN
	document_versions dv ON dv.id = q.version_id AND dv.tenant_id = q.tenant_id
INNER JOIN
	documents d ON d.id = dv.document_id AND d.tenant_id = q.tenant_id
WHERE
	q.file_id IS NULL
	AND q.pdf_attempt_count < @max_pdf_attempts
	AND q.status = @status
	AND d.deleted_at IS NULL
	AND (
		q.pdf_claimed_at IS NULL
		OR q.pdf_claimed_at < @claimed_before
	)
ORDER BY q.created_at ASC
LIMIT 1
FOR UPDATE OF q SKIP LOCKED
`

	rows, err := conn.Query(
		ctx,
		query,
		pgx.StrictNamedArgs{
			"max_pdf_attempts": maxAttempts,
			"status":           DocumentVersionApprovalQuorumStatusPending,
			"claimed_before":   now.Add(-lease),
		},
	)
	if err != nil {
		return fmt.Errorf("cannot query approval quorums: %w", err)
	}

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[DocumentVersionApprovalQuorum])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoDocumentPDFJobAvailable
		}

		return fmt.Errorf("cannot collect approval quorum: %w", err)
	}

	result.PdfAttemptCount++
	result.PdfClaimedAt = new(now)
	result.UpdatedAt = now

	updateQuery := `
UPDATE document_version_approval_quorums SET
	pdf_attempt_count = @pdf_attempt_count,
	pdf_claimed_at = @pdf_claimed_at,
	updated_at = @updated_at
WHERE
	tenant_id = @tenant_id
	AND id = @id
`

	_, err = conn.Exec(
		ctx,
		updateQuery,
		pgx.StrictNamedArgs{
			"id":                result.ID,
			"tenant_id":         result.ID.TenantID(),
			"pdf_attempt_count": result.PdfAttemptCount,
			"pdf_claimed_at":    result.PdfClaimedAt,
			"updated_at":        result.UpdatedAt,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot mark approval quorum as generating PDF: %w", err)
	}

	*q = result

	return nil
}

func (q *DocumentVersionApprovalQuorum) HasPDFClaim(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) (bool, error) {
	if q.PdfClaimedAt == nil {
		return false, nil
	}

	query := `
SELECT EXISTS (
	SELECT 1
	FROM document_version_approval_quorums
	WHERE
		%s
		AND id = @id
		AND file_id IS NULL
		AND status = @status
		AND pdf_claimed_at = @pdf_claimed_at
)
`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":             q.ID,
		"status":         DocumentVersionApprovalQuorumStatusPending,
		"pdf_claimed_at": q.PdfClaimedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	var exists bool
	if err := conn.QueryRow(ctx, query, args).Scan(&exists); err != nil {
		return false, fmt.Errorf("cannot check approval quorum PDF claim: %w", err)
	}

	return exists, nil
}

func (q *DocumentVersionApprovalQuorum) AttachPDFFile(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	fileID gid.GID,
	now time.Time,
) (bool, error) {
	if q.PdfClaimedAt == nil {
		return false, nil
	}

	query := `
UPDATE document_version_approval_quorums
SET
	file_id = @file_id,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
	AND file_id IS NULL
	AND status = @status
	AND pdf_claimed_at = @pdf_claimed_at
`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":             q.ID,
		"file_id":        fileID,
		"status":         DocumentVersionApprovalQuorumStatusPending,
		"pdf_claimed_at": q.PdfClaimedAt,
		"updated_at":     now,
	}
	maps.Copy(args, scope.SQLArguments())

	commandTag, err := conn.Exec(ctx, query, args)
	if err != nil {
		return false, fmt.Errorf("cannot attach approval quorum PDF: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return false, nil
	}

	q.FileID = new(fileID)
	q.UpdatedAt = now

	return true, nil
}

func (q *DocumentVersionApprovalQuorum) ReleasePDFClaim(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	now time.Time,
) error {
	if q.PdfClaimedAt == nil {
		return nil
	}

	query := `
UPDATE document_version_approval_quorums
SET
	pdf_claimed_at = NULL,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
	AND file_id IS NULL
	AND pdf_claimed_at = @pdf_claimed_at
`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":             q.ID,
		"pdf_claimed_at": q.PdfClaimedAt,
		"updated_at":     now,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot release approval quorum PDF claim: %w", err)
	}

	q.PdfClaimedAt = nil
	q.UpdatedAt = now

	return nil
}
