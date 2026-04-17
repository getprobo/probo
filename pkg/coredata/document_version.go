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
	"context"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	DocumentVersion struct {
		ID              gid.GID                    `db:"id"`
		OrganizationID  gid.GID                    `db:"organization_id"`
		DocumentID      gid.GID                    `db:"document_id"`
		Title           string                     `db:"title"`
		Major           int                        `db:"major"`
		Minor           int                        `db:"minor"`
		Classification  DocumentClassification     `db:"classification"`
		DocumentType    DocumentType               `db:"document_type"`
		Content         string                     `db:"content"`
		Changelog       string                     `db:"changelog"`
		Status          DocumentVersionStatus      `db:"status"`
		Orientation     DocumentVersionOrientation `db:"orientation"`
		FileID          *gid.GID                   `db:"file_id"`
		PdfAttemptCount int                        `db:"pdf_attempt_count"`
		PublishedAt     *time.Time                 `db:"published_at"`
		CreatedAt       time.Time                  `db:"created_at"`
		UpdatedAt       time.Time                  `db:"updated_at"`
	}

	DocumentVersions []*DocumentVersion
)

// AuthorizationAttributes returns the authorization attributes for policy evaluation.
func (dv *DocumentVersion) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	q := `
SELECT organization_id
FROM document_versions
WHERE id = $1
LIMIT 1;
`

	var organizationID gid.GID

	if err := conn.QueryRow(ctx, q, dv.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query document version authorization attributes: %w", err)
	}

	return map[string]string{
		"organization_id": organizationID.String(),
	}, nil
}

func (dv *DocumentVersions) LoadByDocumentID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentID gid.GID,
	cursor *page.Cursor[DocumentVersionOrderField],
	filter *DocumentVersionFilter,
) error {
	q := `
SELECT
	id,
	organization_id,
	document_id,
	title,
	major,
	minor,
	classification,
	document_type,
	content,
	changelog,
	status,
	orientation,
	file_id,
	pdf_attempt_count,
	published_at,
	created_at,
	updated_at
FROM
	document_versions
WHERE
	%s
	AND document_id = @document_id
	AND %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_id": documentID,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document versions: %w", err)
	}

	documentVersions, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[DocumentVersion])
	if err != nil {
		return fmt.Errorf("cannot collect document versions: %w", err)
	}

	*dv = documentVersions

	return nil
}

func (dv DocumentVersion) CursorKey(orderBy DocumentVersionOrderField) page.CursorKey {
	switch orderBy {
	case DocumentVersionOrderFieldCreatedAt:
		return page.NewCursorKey(dv.ID, dv.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (dv *DocumentVersion) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentVersionID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	document_id,
	title,
	major,
	minor,
	classification,
	document_type,
	content,
	changelog,
	status,
	orientation,
	file_id,
	pdf_attempt_count,
	published_at,
	created_at,
	updated_at
FROM
	document_versions
WHERE
	%s
	AND id = @document_version_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_version_id": documentVersionID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document versions: %w", err)
	}

	documentVersion, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[DocumentVersion])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect document version: %w", err)
	}

	*dv = documentVersion

	return nil
}

func (dv DocumentVersion) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO document_versions (
	tenant_id,
	id,
	organization_id,
	document_id,
	title,
	major,
	minor,
	classification,
	document_type,
	content,
	changelog,
	status,
	orientation,
	file_id,
	pdf_attempt_count,
	published_at,
	created_at,
	updated_at
)
VALUES (
	@tenant_id,
	@id,
	@organization_id,
	@document_id,
	@title,
	@major,
	@minor,
	@classification,
	@document_type,
	@content,
	@changelog,
	@status,
	@orientation,
	@file_id,
	@pdf_attempt_count,
	@published_at,
	@created_at,
	@updated_at
)
`
	args := pgx.StrictNamedArgs{
		"tenant_id":         scope.GetTenantID(),
		"id":                dv.ID,
		"organization_id":   dv.OrganizationID,
		"document_id":       dv.DocumentID,
		"title":             dv.Title,
		"major":             dv.Major,
		"minor":             dv.Minor,
		"classification":    dv.Classification,
		"document_type":     dv.DocumentType,
		"content":           dv.Content,
		"changelog":         dv.Changelog,
		"status":            dv.Status,
		"orientation":       dv.Orientation,
		"file_id":           dv.FileID,
		"pdf_attempt_count": dv.PdfAttemptCount,
		"published_at":      dv.PublishedAt,
		"created_at":        dv.CreatedAt,
		"updated_at":        dv.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				if pgErr.ConstraintName == "document_versions_document_id_major_minor_key" || pgErr.ConstraintName == "document_one_active_version_idx" {
					return ErrResourceAlreadyExists
				}
			}
		}
		return fmt.Errorf("error creating document version: %w", err)
	}

	return nil
}

func (dv *DocumentVersion) LoadByDocumentIDAndVersion(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentID gid.GID,
	major int,
	minor int,
) error {
	q := `
SELECT
	id,
	organization_id,
	document_id,
	title,
	major,
	minor,
	classification,
	document_type,
	content,
	changelog,
	status,
	orientation,
	file_id,
	pdf_attempt_count,
	published_at,
	created_at,
	updated_at
FROM
	document_versions
WHERE
	%s
	AND document_id = @document_id
	AND major = @major
	AND minor = @minor
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_id": documentID,
		"major":       major,
		"minor":       minor,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document versions: %w", err)
	}

	documentVersion, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[DocumentVersion])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect document version: %w", err)
	}

	*dv = documentVersion

	return nil
}

func (dv *DocumentVersion) LoadLatestVersion(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	document_id,
	title,
	major,
	minor,
	classification,
	document_type,
	content,
	changelog,
	status,
	orientation,
	file_id,
	pdf_attempt_count,
	published_at,
	created_at,
	updated_at
FROM
	document_versions
WHERE
	%s
	AND document_id = @document_id
ORDER BY created_at DESC
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_id": documentID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document versions: %w", err)
	}

	documentVersion, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[DocumentVersion])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect document version: %w", err)
	}

	*dv = documentVersion

	return nil
}

func (dv *DocumentVersion) LoadLatestPublishedVersion(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	document_id,
	title,
	major,
	minor,
	classification,
	document_type,
	content,
	changelog,
	status,
	orientation,
	file_id,
	pdf_attempt_count,
	published_at,
	created_at,
	updated_at
FROM
	document_versions
WHERE
	%s
	AND document_id = @document_id
	AND status = @status
ORDER BY published_at DESC
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_id": documentID,
		"status":      DocumentVersionStatusPublished,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document versions: %w", err)
	}

	documentVersion, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[DocumentVersion])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect document version: %w", err)
	}

	*dv = documentVersion

	return nil
}

func (dv DocumentVersion) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE document_versions SET
	title = @title,
	major = @major,
	minor = @minor,
	changelog = @changelog,
	status = @status,
	content = @content,
	published_at = @published_at,
	classification = @classification,
	document_type = @document_type,
	orientation = @orientation,
	file_id = @file_id,
	pdf_attempt_count = @pdf_attempt_count,
	updated_at = @updated_at
WHERE %s
	AND id = @document_version_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_version_id": dv.ID,
		"title":               dv.Title,
		"major":               dv.Major,
		"minor":               dv.Minor,
		"changelog":           dv.Changelog,
		"status":              dv.Status,
		"content":             dv.Content,
		"published_at":        dv.PublishedAt,
		"classification":      dv.Classification,
		"document_type":       dv.DocumentType,
		"orientation":         dv.Orientation,
		"file_id":             dv.FileID,
		"pdf_attempt_count":   dv.PdfAttemptCount,
		"updated_at":          dv.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update document version: %w", err)
	}

	return nil
}

func (dv DocumentVersion) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM document_versions
WHERE %s
	AND id = @document_version_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{
		"document_version_id": dv.ID,
	}

	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete document version: %w", err)
	}

	return nil
}

func (dv *DocumentVersion) ClaimNextPublishedWithoutFileForUpdate(
	ctx context.Context,
	conn pg.Tx,
	maxAttempts int,
) error {
	q := `
SELECT
	dv.id,
	dv.organization_id,
	dv.document_id,
	dv.title,
	dv.major,
	dv.minor,
	dv.classification,
	dv.document_type,
	dv.content,
	dv.changelog,
	dv.status,
	dv.orientation,
	dv.file_id,
	dv.pdf_attempt_count,
	dv.published_at,
	dv.created_at,
	dv.updated_at
FROM
	document_versions dv
INNER JOIN
	documents d ON d.id = dv.document_id AND d.tenant_id = dv.tenant_id
WHERE
	dv.status = 'PUBLISHED'
	AND dv.file_id IS NULL
	AND dv.pdf_attempt_count < @max_pdf_attempts
	AND d.deleted_at IS NULL
ORDER BY dv.created_at ASC
LIMIT 1
FOR UPDATE OF dv SKIP LOCKED;
`

	rows, err := conn.Query(ctx, q, pgx.StrictNamedArgs{
		"max_pdf_attempts": maxAttempts,
	})
	if err != nil {
		return fmt.Errorf("cannot query document versions: %w", err)
	}

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[DocumentVersion])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoDocumentPDFJobAvailable
		}
		return fmt.Errorf("cannot collect document version: %w", err)
	}

	now := time.Now()
	result.PdfAttemptCount++
	result.UpdatedAt = now

	uq := `
UPDATE document_versions SET
	pdf_attempt_count = @pdf_attempt_count,
	updated_at = @updated_at
WHERE
	tenant_id = @tenant_id
	AND id = @id
`
	uargs := pgx.StrictNamedArgs{
		"id":                result.ID,
		"tenant_id":         result.ID.TenantID(),
		"pdf_attempt_count": result.PdfAttemptCount,
		"updated_at":        result.UpdatedAt,
	}

	if _, err := conn.Exec(ctx, uq, uargs); err != nil {
		return fmt.Errorf("cannot mark document version as generating PDF: %w", err)
	}

	*dv = result

	return nil
}

func (dv *DocumentVersions) CountByDocumentID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentID gid.GID,
	filter *DocumentVersionFilter,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	document_versions
WHERE
	%s
	AND document_id = @document_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.NamedArgs{"document_id": documentID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}
