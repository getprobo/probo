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
	Document struct {
		ID                      gid.GID      `db:"id"`
		OrganizationID          gid.GID      `db:"organization_id"`
		OwnerID                 gid.GID      `db:"owner_id"`
		Title                   string       `db:"title"`
		DocumentType            DocumentType `db:"document_type"`
		CurrentPublishedVersion *int         `db:"current_published_version"`
		ShowOnTrustCenter       bool         `db:"show_on_trust_center"`
		CreatedAt               time.Time    `db:"created_at"`
		UpdatedAt               time.Time    `db:"updated_at"`
	}

	Documents []*Document
)

func (p Document) CursorKey(orderBy DocumentOrderField) page.CursorKey {
	switch orderBy {
	case DocumentOrderFieldCreatedAt:
		return page.NewCursorKey(p.ID, p.CreatedAt)
	case DocumentOrderFieldTitle:
		return page.NewCursorKey(p.ID, p.Title)
	case DocumentOrderFieldDocumentType:
		return page.NewCursorKey(p.ID, p.DocumentType)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (p *Document) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	documentID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    owner_id,
    title,
    document_type,
    current_published_version,
    show_on_trust_center,
    created_at,
    updated_at
FROM
    documents
WHERE
    %s
    AND deleted_at IS NULL
    AND id = @document_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"document_id": documentID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query documents: %w", err)
	}

	document, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Document])
	if err != nil {
		return fmt.Errorf("cannot collect document: %w", err)
	}

	*p = document

	return nil
}

func (p *Documents) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	filter *DocumentFilter,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
    documents
WHERE
    %s
    AND deleted_at IS NULL
    AND organization_id = @organization_id
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.NamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (p *Documents) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[DocumentOrderField],
	filter *DocumentFilter,
) error {
	q := `
SELECT
	id,
    organization_id,
    owner_id,
    title,
    document_type,
    current_published_version,
    show_on_trust_center,
    created_at,
    updated_at
FROM
    documents
WHERE
    %s
    AND deleted_at IS NULL
    AND organization_id = @organization_id
    AND %s
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query documents: %w", err)
	}

	documents, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Document])
	if err != nil {
		return fmt.Errorf("cannot collect documents: %w", err)
	}

	*p = documents

	return nil
}

func (p Document) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    documents (
        tenant_id,
		id,
		organization_id,
		owner_id,
		title,
		document_type,
		current_published_version,
		show_on_trust_center,
		created_at,
		updated_at
    )
VALUES (
    @tenant_id,
    @document_id,
    @organization_id,
    @owner_id,
    @title,
    @document_type,
    @current_published_version,
    @show_on_trust_center,
    @created_at,
    @updated_at
);
`

	args := pgx.StrictNamedArgs{
		"tenant_id":                 scope.GetTenantID(),
		"document_id":               p.ID,
		"organization_id":           p.OrganizationID,
		"owner_id":                  p.OwnerID,
		"title":                     p.Title,
		"document_type":             p.DocumentType,
		"current_published_version": p.CurrentPublishedVersion,
		"show_on_trust_center":      p.ShowOnTrustCenter,
		"created_at":                p.CreatedAt,
		"updated_at":                p.UpdatedAt,
	}
	_, err := conn.Exec(ctx, q, args)
	return err
}

func (p Document) SoftDelete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE documents SET deleted_at = @deleted_at WHERE %s AND id = @document_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"document_id": p.ID, "deleted_at": time.Now()}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (p Document) DeleteByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) error {
	q := `
DELETE FROM documents WHERE %s AND organization_id = @organization_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (p *Document) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE
	documents
SET
	title = @title,
	current_published_version = @current_published_version,
	owner_id = @owner_id,
	document_type = @document_type,
	show_on_trust_center = @show_on_trust_center,
	updated_at = @updated_at
WHERE
	%s
	AND id = @document_id
	AND deleted_at IS NULL
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_id":               p.ID,
		"updated_at":                time.Now(),
		"title":                     p.Title,
		"current_published_version": p.CurrentPublishedVersion,
		"owner_id":                  p.OwnerID,
		"document_type":             p.DocumentType,
		"show_on_trust_center":      p.ShowOnTrustCenter,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update document: %w", err)
	}

	return nil
}

func (p *Documents) CountByControlID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	controlID gid.GID,
	filter *DocumentFilter,
) (int, error) {
	q := `
WITH plcs AS (
	SELECT
		p.id,
		p.tenant_id,
		p.search_vector,
		p.show_on_trust_center,
		p.deleted_at
	FROM
		documents p
	INNER JOIN
		controls_documents cp ON p.id = cp.document_id
	WHERE
		cp.control_id = @control_id
)
SELECT
	COUNT(id)
FROM
	plcs
WHERE
	%s
	AND deleted_at IS NULL
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.NamedArgs{"control_id": controlID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (p *Documents) LoadByControlID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	controlID gid.GID,
	cursor *page.Cursor[DocumentOrderField],
	filter *DocumentFilter,
) error {
	q := `
WITH plcs AS (
	SELECT
		p.id,
		p.tenant_id,
		p.search_vector,
		p.organization_id,
		p.owner_id,
		p.title,
		p.document_type,
		p.current_published_version,
		p.show_on_trust_center,
		p.created_at,
		p.updated_at,
		p.deleted_at
	FROM
		documents p
	INNER JOIN
		controls_documents cp ON p.id = cp.document_id
	WHERE
		cp.control_id = @control_id
)
SELECT
	id,
	organization_id,
	owner_id,
	title,
	document_type,
	current_published_version,
	show_on_trust_center,
	created_at,
	updated_at
FROM
	plcs
WHERE
	%s
	AND deleted_at IS NULL
	AND %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"control_id": controlID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query documents: %w", err)
	}

	documents, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Document])
	if err != nil {
		return fmt.Errorf("cannot collect documents: %w", err)
	}

	*p = documents

	return nil
}

func (p *Documents) CountByRiskID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	riskID gid.GID,
	filter *DocumentFilter,
) (int, error) {
	q := `
WITH plcs AS (
	SELECT
		p.id,
		p.tenant_id,
		p.search_vector,
		p.show_on_trust_center,
		p.deleted_at
	FROM
		documents p
	INNER JOIN
		risks_documents rp ON p.id = rp.document_id
	WHERE
		rp.risk_id = @risk_id
)
SELECT
	COUNT(id)
FROM
	plcs
WHERE
	%s
	AND deleted_at IS NULL
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.NamedArgs{"risk_id": riskID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (p *Documents) LoadByRiskID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	riskID gid.GID,
	cursor *page.Cursor[DocumentOrderField],
	filter *DocumentFilter,
) error {
	q := `
WITH plcs AS (
	SELECT
		p.id,
		p.tenant_id,
		p.organization_id,
		p.owner_id,
		p.title,
		p.document_type,
		p.current_published_version,
		p.show_on_trust_center,
		p.created_at,
		p.updated_at,
		p.search_vector,
		p.deleted_at
	FROM
		documents p
	INNER JOIN
		risks_documents rp ON p.id = rp.document_id
	WHERE
		rp.risk_id = @risk_id
)
SELECT
	id,
	organization_id,
	owner_id,
	title,
	document_type,
	current_published_version,
	show_on_trust_center,
	created_at,
	updated_at
FROM
	plcs
WHERE
	%s
	AND deleted_at IS NULL
	AND %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"risk_id": riskID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query documents: %w", err)
	}

	documents, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Document])
	if err != nil {
		return fmt.Errorf("cannot collect documents: %w", err)
	}

	*p = documents

	return nil
}
