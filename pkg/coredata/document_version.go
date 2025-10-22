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
	"fmt"
	"maps"
	"time"

	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	DocumentVersion struct {
		ID             gid.GID                `db:"id"`
		DocumentID     gid.GID                `db:"document_id"`
		Title          string                 `db:"title"`
		OwnerID        gid.GID                `db:"owner_id"`
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

func (p *DocumentVersions) LoadByDocumentID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	documentID gid.GID,
	cursor *page.Cursor[DocumentVersionOrderField],
) error {
	q := `
SELECT
	id,
	document_id,
	title,
	owner_id,
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
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_id": documentID,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document versions: %w", err)
	}

	documentVersions, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[DocumentVersion])
	if err != nil {
		return fmt.Errorf("cannot collect document versions: %w", err)
	}

	*p = documentVersions

	return nil
}

func (p DocumentVersion) CursorKey(orderBy DocumentVersionOrderField) page.CursorKey {
	switch orderBy {
	case DocumentVersionOrderFieldCreatedAt:
		return page.NewCursorKey(p.ID, p.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (p *DocumentVersion) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	documentVersionID gid.GID,
) error {
	q := `
SELECT
	id,
	document_id,
	title,
	owner_id,
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

	*p = documentVersion

	return nil
}

func (p DocumentVersion) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO document_versions (
	tenant_id,
	id,
	document_id,
	title,
	owner_id,
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
	@document_id,
	@title,
	@owner_id,
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
		"tenant_id":      scope.GetTenantID(),
		"id":             p.ID,
		"document_id":    p.DocumentID,
		"title":          p.Title,
		"owner_id":       p.OwnerID,
		"version_number": p.VersionNumber,
		"classification": p.Classification,
		"content":        p.Content,
		"changelog":      p.Changelog,
		"status":         p.Status,
		"created_at":     p.CreatedAt,
		"updated_at":     p.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("error creating document version: %w", err)
	}

	return nil
}

func (p *DocumentVersion) LoadByDocumentIDAndVersionNumber(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	documentID gid.GID,
	versionNumber int,
) error {
	q := `
SELECT
	id,
	document_id,
	title,
	owner_id,
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

	*p = documentVersion

	return nil
}

func (p *DocumentVersion) LoadLatestVersion(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	documentID gid.GID,
) error {
	q := `
SELECT
	id,
	document_id,
	title,
	owner_id,
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

	*p = documentVersion

	return nil
}

func (p *DocumentVersion) LoadLatestPublishedVersion(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	documentID gid.GID,
) error {
	q := `
SELECT
	id,
	document_id,
	title,
	owner_id,
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

	*p = documentVersion

	return nil
}

func (p DocumentVersion) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE document_versions SET
	title = @title,
	owner_id = @owner_id,
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
		"document_version_id": p.ID,
		"title":               p.Title,
		"owner_id":            p.OwnerID,
		"changelog":           p.Changelog,
		"status":              p.Status,
		"content":             p.Content,
		"published_at":        p.PublishedAt,
		"classification":      p.Classification,
		"updated_at":          p.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update document version: %w", err)
	}

	return nil
}

func (p DocumentVersion) Delete(
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
		"document_version_id": p.ID,
	}

	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete document version: %w", err)
	}

	return nil
}
