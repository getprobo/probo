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
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
)

type (
	Document struct {
		ID                    gid.GID               `db:"id"`
		OrganizationID        gid.GID               `db:"organization_id"`
		CurrentPublishedMajor *int                  `db:"current_published_major"`
		CurrentPublishedMinor *int                  `db:"current_published_minor"`
		TrustCenterVisibility TrustCenterVisibility `db:"trust_center_visibility"`
		Status                DocumentStatus        `db:"status"`
		ArchivedAt            *time.Time            `db:"archived_at"`
		CreatedAt             time.Time             `db:"created_at"`
		UpdatedAt             time.Time             `db:"updated_at"`

		// ordering only
		Title        string       `db:"title"`
		DocumentType DocumentType `db:"document_type"`
	}

	Documents []*Document
)

func (p Document) CursorKey(orderBy DocumentOrderField) page.CursorKey {
	switch orderBy {
	case DocumentOrderFieldCreatedAt:
		return page.NewCursorKey(p.ID, p.CreatedAt)
	case DocumentOrderFieldUpdatedAt:
		return page.NewCursorKey(p.ID, p.UpdatedAt)
	case DocumentOrderFieldTitle:
		return page.NewCursorKey(p.ID, p.Title)
	case DocumentOrderFieldDocumentType:
		return page.NewCursorKey(p.ID, p.DocumentType)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

// AuthorizationAttributes returns the authorization attributes for policy evaluation.
func (d *Document) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	q := `
SELECT organization_id
FROM documents
WHERE id = $1
LIMIT 1;
`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, d.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query document authorization attributes: %w", err)
	}

	return map[string]string{
		"organization_id": organizationID.String(),
	}, nil
}

func (p *Document) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentID gid.GID,
) error {
	q := `
WITH latest_versions AS (
    SELECT DISTINCT ON (document_id) document_id, title, document_type
    FROM document_versions
    ORDER BY document_id, major DESC, minor DESC
)
SELECT
    documents.id,
    documents.organization_id,
    documents.current_published_major,
    documents.current_published_minor,
    documents.trust_center_visibility,
    documents.status,
    documents.archived_at,
    documents.created_at,
    documents.updated_at,
    COALESCE(lv.title, '') AS title,
    COALESCE(lv.document_type, 'OTHER') AS document_type
FROM
    documents
LEFT JOIN latest_versions lv ON lv.document_id = documents.id
WHERE
    %s
    AND documents.deleted_at IS NULL
    AND documents.id = @document_id
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
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect document: %w", err)
	}

	*p = document

	return nil
}

func (p *Document) LoadByIDWithFilter(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentID gid.GID,
	filter *DocumentFilter,
) error {
	q := `
WITH latest_versions AS (
    SELECT DISTINCT ON (document_id) document_id, title, document_type
    FROM document_versions
    ORDER BY document_id, major DESC, minor DESC
)
SELECT
    documents.id,
    documents.organization_id,
    documents.current_published_major,
    documents.current_published_minor,
    documents.trust_center_visibility,
    documents.status,
    documents.archived_at,
    documents.created_at,
    documents.updated_at,
    COALESCE(lv.title, '') AS title,
    COALESCE(lv.document_type, 'OTHER') AS document_type
FROM
    documents
LEFT JOIN latest_versions lv ON lv.document_id = documents.id
WHERE
    %s
    AND documents.deleted_at IS NULL
    AND documents.id = @document_id
    AND %s
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"document_id": documentID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query documents: %w", err)
	}

	document, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Document])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect document: %w", err)
	}

	*p = document

	return nil
}

func (p *Documents) LoadByIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentIDs []gid.GID,
) error {
	q := `
WITH latest_versions AS (
    SELECT DISTINCT ON (document_id) document_id, title, document_type
    FROM document_versions
    ORDER BY document_id, major DESC, minor DESC
)
SELECT
    documents.id,
    documents.organization_id,
    documents.current_published_major,
    documents.current_published_minor,
    documents.trust_center_visibility,
    documents.status,
    documents.archived_at,
    documents.created_at,
    documents.updated_at,
    COALESCE(lv.title, '') AS title,
    COALESCE(lv.document_type, 'OTHER') AS document_type
FROM
    documents
LEFT JOIN latest_versions lv ON lv.document_id = documents.id
WHERE
    %s
    AND documents.deleted_at IS NULL
    AND documents.id = ANY(@document_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"document_ids": documentIDs}
	maps.Copy(args, scope.SQLArguments())

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

func (p *Documents) CountByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
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
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[DocumentOrderField],
	filter *DocumentFilter,
) error {
	q := `
WITH latest_versions AS (
    SELECT DISTINCT ON (document_id) document_id, title, document_type
    FROM document_versions
    ORDER BY document_id, major DESC, minor DESC
),
base AS (
    SELECT
        documents.id,
        documents.organization_id,
        documents.current_published_major,
        documents.current_published_minor,
        documents.trust_center_visibility,
        documents.status,
        documents.archived_at,
        documents.created_at,
        documents.updated_at,
        COALESCE(lv.title, '') AS title,
        COALESCE(lv.document_type, 'OTHER') AS document_type
    FROM
        documents
    LEFT JOIN latest_versions lv ON lv.document_id = documents.id
    WHERE
        %s
        AND documents.deleted_at IS NULL
        AND documents.organization_id = @organization_id
        AND %s
)
SELECT * FROM base WHERE %s
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

func (p *Documents) LoadAllByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	filter *DocumentFilter,
) error {
	q := `
WITH latest_versions AS (
    SELECT DISTINCT ON (document_id) document_id, title, document_type
    FROM document_versions
    ORDER BY document_id, major DESC, minor DESC
)
SELECT
	documents.id,
    documents.organization_id,
    documents.current_published_major,
    documents.current_published_minor,
    documents.trust_center_visibility,
    documents.status,
    documents.archived_at,
    documents.created_at,
    documents.updated_at,
    COALESCE(lv.title, '') AS title,
    COALESCE(lv.document_type, 'OTHER') AS document_type
FROM
    documents
LEFT JOIN latest_versions lv ON lv.document_id = documents.id
WHERE
    %s
    AND documents.deleted_at IS NULL
    AND documents.organization_id = @organization_id
    AND %s
ORDER BY title ASC
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.NamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

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

func (p *Documents) LoadPublishedByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[DocumentOrderField],
	filter *DocumentFilter,
) error {
	q := `
WITH latest_versions AS (
	SELECT DISTINCT ON (document_id) document_id, title, document_type
	FROM document_versions
	ORDER BY document_id, major DESC, minor DESC
),
published_versions AS (
	SELECT
		dv.document_id,
		dv.title AS published_title
	FROM
		document_versions dv
		INNER JOIN documents d
			ON dv.document_id = d.id
			AND dv.major = d.current_published_major
			AND dv.minor = d.current_published_minor
	WHERE
		d.deleted_at IS NULL
		AND d.organization_id = @organization_id
),
base AS (
	SELECT
		documents.id,
		documents.organization_id,
		documents.current_published_major,
		documents.current_published_minor,
		documents.trust_center_visibility,
		documents.status,
		documents.archived_at,
		documents.created_at,
		documents.updated_at,
		COALESCE(pv.published_title, lv.title, '') AS title,
		COALESCE(lv.document_type, 'OTHER') AS document_type
	FROM
		documents
	LEFT JOIN latest_versions lv ON lv.document_id = documents.id
	LEFT JOIN published_versions pv ON pv.document_id = documents.id
	WHERE
		%s
		AND documents.deleted_at IS NULL
		AND documents.organization_id = @organization_id
		AND %s
)
SELECT * FROM base WHERE %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query published documents: %w", err)
	}

	documents, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Document])
	if err != nil {
		return fmt.Errorf("cannot collect published documents: %w", err)
	}

	*p = documents

	return nil
}

func (p Document) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    documents (
        tenant_id,
		id,
		organization_id,
		current_published_major,
		current_published_minor,
		trust_center_visibility,
		status,
		archived_at,
		created_at,
		updated_at
    )
VALUES (
    @tenant_id,
    @document_id,
    @organization_id,
    @current_published_major,
    @current_published_minor,
    @trust_center_visibility,
    @status,
    @archived_at,
    @created_at,
    @updated_at
);
`

	args := pgx.StrictNamedArgs{
		"tenant_id":               scope.GetTenantID(),
		"document_id":             p.ID,
		"organization_id":         p.OrganizationID,
		"current_published_major": p.CurrentPublishedMajor,
		"current_published_minor": p.CurrentPublishedMinor,
		"trust_center_visibility": p.TrustCenterVisibility,
		"status":                  p.Status,
		"archived_at":             p.ArchivedAt,
		"created_at":              p.CreatedAt,
		"updated_at":              p.UpdatedAt,
	}
	_, err := conn.Exec(ctx, q, args)
	return err
}

func (p Document) SoftDelete(
	ctx context.Context,
	conn pg.Tx,
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
	conn pg.Tx,
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
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE
	documents
SET
	current_published_major = @current_published_major,
	current_published_minor = @current_published_minor,
	trust_center_visibility = @trust_center_visibility,
	status = @status,
	archived_at = @archived_at,
	updated_at = @updated_at
WHERE
	%s
	AND id = @document_id
	AND deleted_at IS NULL
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_id":             p.ID,
		"updated_at":              time.Now(),
		"current_published_major": p.CurrentPublishedMajor,
		"current_published_minor": p.CurrentPublishedMinor,
		"trust_center_visibility": p.TrustCenterVisibility,
		"status":                  p.Status,
		"archived_at":             p.ArchivedAt,
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
	conn pg.Querier,
	scope Scoper,
	controlID gid.GID,
	filter *DocumentFilter,
) (int, error) {
	q := `
WITH scoped_documents AS (
	SELECT *
	FROM documents
	WHERE %s
		AND deleted_at IS NULL
		AND %s
)
SELECT COUNT(scoped_documents.id)
FROM scoped_documents
INNER JOIN controls_documents cp ON scoped_documents.id = cp.document_id
WHERE cp.control_id = @control_id
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
	conn pg.Querier,
	scope Scoper,
	controlID gid.GID,
	cursor *page.Cursor[DocumentOrderField],
	filter *DocumentFilter,
) error {
	q := `
WITH latest_versions AS (
	SELECT DISTINCT ON (document_id) document_id, title, document_type
	FROM document_versions
	ORDER BY document_id, major DESC, minor DESC
),
scoped_documents AS (
	SELECT *
	FROM documents
	WHERE %s
		AND deleted_at IS NULL
		AND %s
),
base AS (
	SELECT
		sd.id,
		sd.organization_id,
		sd.current_published_major,
		sd.current_published_minor,
		sd.trust_center_visibility,
		sd.status,
		sd.archived_at,
		sd.created_at,
		sd.updated_at,
		COALESCE(lv.title, '') AS title,
		COALESCE(lv.document_type, 'OTHER') AS document_type
	FROM scoped_documents sd
	INNER JOIN controls_documents cp ON sd.id = cp.document_id
	LEFT JOIN latest_versions lv ON lv.document_id = sd.id
	WHERE cp.control_id = @control_id
)
SELECT * FROM base WHERE %s
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
	conn pg.Querier,
	scope Scoper,
	riskID gid.GID,
	filter *DocumentFilter,
) (int, error) {
	q := `
WITH scoped_documents AS (
	SELECT *
	FROM documents
	WHERE %s
		AND deleted_at IS NULL
		AND %s
)
SELECT COUNT(scoped_documents.id)
FROM scoped_documents
INNER JOIN risks_documents rp ON scoped_documents.id = rp.document_id
WHERE rp.risk_id = @risk_id
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
	conn pg.Querier,
	scope Scoper,
	riskID gid.GID,
	cursor *page.Cursor[DocumentOrderField],
	filter *DocumentFilter,
) error {
	q := `
WITH latest_versions AS (
	SELECT DISTINCT ON (document_id) document_id, title, document_type
	FROM document_versions
	ORDER BY document_id, major DESC, minor DESC
),
scoped_documents AS (
	SELECT *
	FROM documents
	WHERE %s
		AND deleted_at IS NULL
		AND %s
),
base AS (
	SELECT
		sd.id,
		sd.organization_id,
		sd.current_published_major,
		sd.current_published_minor,
		sd.trust_center_visibility,
		sd.status,
		sd.archived_at,
		sd.created_at,
		sd.updated_at,
		COALESCE(lv.title, '') AS title,
		COALESCE(lv.document_type, 'OTHER') AS document_type
	FROM scoped_documents sd
	INNER JOIN risks_documents rp ON sd.id = rp.document_id
	LEFT JOIN latest_versions lv ON lv.document_id = sd.id
	WHERE rp.risk_id = @risk_id
)
SELECT * FROM base WHERE %s
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

func (p *Documents) CountByMeasureID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	measureID gid.GID,
	filter *DocumentFilter,
) (int, error) {
	q := `
WITH scoped_documents AS (
	SELECT *
	FROM documents
	WHERE %s
		AND deleted_at IS NULL
		AND %s
)
SELECT COUNT(scoped_documents.id)
FROM scoped_documents
INNER JOIN measures_documents md ON scoped_documents.id = md.document_id
WHERE md.measure_id = @measure_id
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.NamedArgs{"measure_id": measureID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (p *Documents) LoadByMeasureID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	measureID gid.GID,
	cursor *page.Cursor[DocumentOrderField],
	filter *DocumentFilter,
) error {
	q := `
WITH latest_versions AS (
	SELECT DISTINCT ON (document_id) document_id, title, document_type
	FROM document_versions
	ORDER BY document_id, major DESC, minor DESC
),
scoped_documents AS (
	SELECT *
	FROM documents
	WHERE %s
		AND deleted_at IS NULL
		AND %s
),
base AS (
	SELECT
		sd.id,
		sd.organization_id,
		sd.current_published_major,
		sd.current_published_minor,
		sd.trust_center_visibility,
		sd.status,
		sd.archived_at,
		sd.created_at,
		sd.updated_at,
		COALESCE(lv.title, '') AS title,
		COALESCE(lv.document_type, 'OTHER') AS document_type
	FROM scoped_documents sd
	INNER JOIN measures_documents md ON sd.id = md.document_id
	LEFT JOIN latest_versions lv ON lv.document_id = sd.id
	WHERE md.measure_id = @measure_id
)
SELECT * FROM base WHERE %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"measure_id": measureID}
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

func (p *Documents) BulkSoftDelete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE documents SET deleted_at = @deleted_at WHERE %s AND id = ANY(@document_ids)
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	ids := make([]gid.GID, len(*p))
	for i, doc := range *p {
		ids[i] = doc.ID
	}

	args := pgx.StrictNamedArgs{
		"document_ids": ids,
		"deleted_at":   time.Now()}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (p *Documents) BulkArchive(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
UPDATE documents SET status = 'ARCHIVED', archived_at = @archived_at, trust_center_visibility = 'NONE' WHERE %s AND id = ANY(@document_ids)
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	ids := make([]gid.GID, len(*p))
	for i, doc := range *p {
		ids[i] = doc.ID
	}

	args := pgx.StrictNamedArgs{
		"document_ids": ids,
		"archived_at":  time.Now(),
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot bulk archive documents: %w", err)
	}
	return nil
}

func (p *Documents) BulkUnarchive(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
UPDATE documents SET status = 'ACTIVE', archived_at = NULL WHERE %s AND id = ANY(@document_ids)
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	ids := make([]gid.GID, len(*p))
	for i, doc := range *p {
		ids[i] = doc.ID
	}

	args := pgx.StrictNamedArgs{
		"document_ids": ids,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot bulk unarchive documents: %w", err)
	}
	return nil
}

func (p *Document) IsLastSignableVersionSignedByUserEmail(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentID gid.GID,
	userEmail mail.Addr,
) (bool, error) {
	q := `
WITH max_signable_major AS (
	SELECT MAX(dv.major) AS major
	FROM document_versions dv
	INNER JOIN document_version_signatures dvs ON dvs.document_version_id = dv.id
	INNER JOIN iam_membership_profiles p ON dvs.signed_by_profile_id = p.id
	INNER JOIN identities i ON p.identity_id = i.id
	WHERE dv.document_id = @document_id
		AND i.email_address = @user_email::CITEXT
),
last_signable_version AS (
	SELECT
		d.id AS document_id,
		d.tenant_id,
		dv.major,
		dvs.state
	FROM documents d
	INNER JOIN document_versions dv ON dv.document_id = d.id
	INNER JOIN max_signable_major msm ON dv.major = msm.major
	INNER JOIN document_version_signatures dvs ON dvs.document_version_id = dv.id
	INNER JOIN iam_membership_profiles p ON dvs.signed_by_profile_id = p.id
	INNER JOIN identities i ON p.identity_id = i.id
	WHERE d.id = @document_id
		AND i.email_address = @user_email::CITEXT
)
SELECT EXISTS (
	SELECT 1
	FROM last_signable_version
	WHERE %s
		AND state = 'SIGNED'
) AS signed
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_id": documentID,
		"user_email":  userEmail,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot query document signed status: %w", err)
	}

	signed, err := pgx.CollectOneRow(rows, pgx.RowTo[bool])
	if err != nil {
		return false, fmt.Errorf("cannot collect signed status: %w", err)
	}

	return signed, nil
}

func (p *Document) GetViewerApprovalStateForLastVersion(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentID gid.GID,
	identityID gid.GID,
) (DocumentVersionApprovalDecisionState, error) {
	q := `
WITH viewer_decision AS (
	SELECT
		dvad.tenant_id,
		dvad.state,
		dv.major,
		dvaq.created_at AS quorum_created_at
	FROM documents d
	INNER JOIN document_versions dv ON dv.document_id = d.id
	INNER JOIN document_version_approval_quorums dvaq ON dvaq.version_id = dv.id
	INNER JOIN document_version_approval_decisions dvad ON dvad.quorum_id = dvaq.id
	INNER JOIN iam_membership_profiles p ON dvad.approver_id = p.id
	WHERE d.id = @document_id
		AND p.identity_id = @identity_id
)
SELECT state
FROM viewer_decision
WHERE %s
ORDER BY major DESC, quorum_created_at DESC
LIMIT 1
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_id": documentID,
		"identity_id": identityID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return "", fmt.Errorf("cannot query document approval state: %w", err)
	}

	state, err := pgx.CollectOneRow(rows, pgx.RowTo[DocumentVersionApprovalDecisionState])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("cannot collect approval state: %w", err)
	}

	return state, nil
}
