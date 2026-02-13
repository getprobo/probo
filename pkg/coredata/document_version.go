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
		ID             gid.GID                `db:"id"`
		OrganizationID gid.GID                `db:"organization_id"`
		DocumentID     gid.GID                `db:"document_id"`
		Title          string                 `db:"title"`
		VersionNumber  int                    `db:"version_number"`
		Classification DocumentClassification `db:"classification"`
		Content        string                 `db:"content"`
		Changelog      string                 `db:"changelog"`
		Status         DocumentStatus         `db:"status"`
		PublishedAt    *time.Time             `db:"published_at"`
		CreatedAt      time.Time              `db:"created_at"`
		UpdatedAt      time.Time              `db:"updated_at"`
	}

	DocumentVersions []*DocumentVersion
)

// AuthorizationAttributes returns the authorization attributes for policy evaluation.
func (dv *DocumentVersion) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM document_versions WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, dv.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query document version authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (dv *DocumentVersions) LoadByDocumentID(
	ctx context.Context,
	conn pg.Conn,
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
	version_number,
	classification,
	content,
	changelog,
	status,
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
	conn pg.Conn,
	scope Scoper,
	documentVersionID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	document_id,
	title,
	version_number,
	classification,
	content,
	changelog,
	status,
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
		return fmt.Errorf("cannot collect document version: %w", err)
	}

	*dv = documentVersion

	return nil
}

func (dv DocumentVersion) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO document_versions (
	tenant_id,
	id,
	organization_id,
	document_id,
	title,
	version_number,
	classification,
	content,
	changelog,
	status,
	created_at,
	updated_at
)
VALUES (
	@tenant_id,
	@id,
	@organization_id,
	@document_id,
	@title,
	@version_number,
	@classification,
	@content,
	@changelog,
	@status,
	@created_at,
	@updated_at
)
`
	args := pgx.StrictNamedArgs{
		"tenant_id":       scope.GetTenantID(),
		"id":              dv.ID,
		"organization_id": dv.OrganizationID,
		"document_id":     dv.DocumentID,
		"title":           dv.Title,
		"version_number":  dv.VersionNumber,
		"classification":  dv.Classification,
		"content":         dv.Content,
		"changelog":       dv.Changelog,
		"status":          dv.Status,
		"created_at":      dv.CreatedAt,
		"updated_at":      dv.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				if pgErr.ConstraintName == "document_versions_document_id_version_number_key" || pgErr.ConstraintName == "document_one_draft_version_idx" {
					return ErrResourceAlreadyExists
				}
			}
		}
		return fmt.Errorf("error creating document version: %w", err)
	}

	return nil
}

func (dv *DocumentVersion) LoadByDocumentIDAndVersionNumber(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	documentID gid.GID,
	versionNumber int,
) error {
	q := `
SELECT
	id,
	organization_id,
	document_id,
	title,
	version_number,
	classification,
	content,
	changelog,
	status,
	published_at,
	created_at,
	updated_at
FROM
	document_versions
WHERE
	%s
	AND document_id = @document_id
	AND version_number = @version_number
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_id":    documentID,
		"version_number": versionNumber,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document versions: %w", err)
	}

	documentVersion, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[DocumentVersion])
	if err != nil {
		return fmt.Errorf("cannot collect document version: %w", err)
	}

	*dv = documentVersion

	return nil
}

func (dv *DocumentVersion) LoadLatestVersion(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	documentID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	document_id,
	title,
	version_number,
	classification,
	content,
	changelog,
	status,
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
		return fmt.Errorf("cannot collect document version: %w", err)
	}

	*dv = documentVersion

	return nil
}

func (dv *DocumentVersion) LoadLatestPublishedVersion(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	documentID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	document_id,
	title,
	version_number,
	classification,
	content,
	changelog,
	status,
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
		"status":      DocumentStatusPublished,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document versions: %w", err)
	}

	documentVersion, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[DocumentVersion])
	if err != nil {
		return fmt.Errorf("cannot collect document version: %w", err)
	}

	*dv = documentVersion

	return nil
}

func (dv DocumentVersion) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE document_versions SET
	title = @title,
	changelog = @changelog,
	status = @status,
	content = @content,
	published_at = @published_at,
	classification = @classification,
	updated_at = @updated_at
WHERE %s
	AND id = @document_version_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_version_id": dv.ID,
		"title":               dv.Title,
		"changelog":           dv.Changelog,
		"status":              dv.Status,
		"content":             dv.Content,
		"published_at":        dv.PublishedAt,
		"classification":      dv.Classification,
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
	conn pg.Conn,
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

func (dv *DocumentVersions) CountByDocumentID(
	ctx context.Context,
	conn pg.Conn,
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
