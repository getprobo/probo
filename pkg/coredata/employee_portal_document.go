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
	"go.probo.inc/probo/pkg/page"
)

type (
	EmployeePortalDocument struct {
		ID               gid.GID   `db:"id"`
		OrganizationID   gid.GID   `db:"organization_id"`
		EmployeePortalID gid.GID   `db:"employee_portal_id"`
		DocumentID       gid.GID   `db:"document_id"`
		CreatedAt        time.Time `db:"created_at"`
		UpdatedAt        time.Time `db:"updated_at"`
	}

	EmployeePortalDocuments []*EmployeePortalDocument
)

func (d EmployeePortalDocument) CursorKey(orderBy EmployeePortalDocumentOrderField) page.CursorKey {
	switch orderBy {
	case EmployeePortalDocumentOrderFieldCreatedAt:
		return page.NewCursorKey(d.ID, d.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (d *EmployeePortalDocument) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `
SELECT
	id,
	organization_id
FROM
	employee_portal_documents
WHERE
	id = ANY(@resource_ids::text[])
`

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

func (d *EmployeePortalDocument) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	linkID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	employee_portal_id,
	document_id,
	created_at,
	updated_at
FROM
	employee_portal_documents
WHERE
	%s
	AND id = @id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": linkID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query employee portal document: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[EmployeePortalDocument])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect employee portal document: %w", err)
	}

	*d = row

	return nil
}

func (d *EmployeePortalDocument) LoadByEmployeePortalIDAndDocumentID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	employeePortalID gid.GID,
	documentID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	employee_portal_id,
	document_id,
	created_at,
	updated_at
FROM
	employee_portal_documents
WHERE
	%s
	AND employee_portal_id = @employee_portal_id
	AND document_id = @document_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"employee_portal_id": employeePortalID,
		"document_id":        documentID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query employee portal document: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[EmployeePortalDocument])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect employee portal document: %w", err)
	}

	*d = row

	return nil
}

func (d *EmployeePortalDocument) LoadByOrganizationIDAndDocumentID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	documentID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	employee_portal_id,
	document_id,
	created_at,
	updated_at
FROM
	employee_portal_documents
WHERE
	%s
	AND organization_id = @organization_id
	AND document_id = @document_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
		"document_id":     documentID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query employee portal document: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[EmployeePortalDocument])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect employee portal document: %w", err)
	}

	*d = row

	return nil
}

func (ds *EmployeePortalDocuments) LoadByEmployeePortalIDAndDocumentIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	employeePortalID gid.GID,
	documentIDs []gid.GID,
) error {
	if len(documentIDs) == 0 {
		*ds = EmployeePortalDocuments{}
		return nil
	}

	q := `
SELECT
	id,
	organization_id,
	employee_portal_id,
	document_id,
	created_at,
	updated_at
FROM
	employee_portal_documents
WHERE
	%s
	AND employee_portal_id = @employee_portal_id
	AND document_id = ANY(@document_ids::text[]);
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"employee_portal_id": employeePortalID,
		"document_ids":       documentIDs,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query employee portal documents: %w", err)
	}

	links, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[EmployeePortalDocument])
	if err != nil {
		return fmt.Errorf("cannot collect employee portal documents: %w", err)
	}

	*ds = links

	return nil
}

func (ds *EmployeePortalDocuments) LoadByEmployeePortalID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	employeePortalID gid.GID,
	documentIDs []gid.GID,
	cursor *page.Cursor[EmployeePortalDocumentOrderField],
) error {
	if documentIDs != nil && len(documentIDs) == 0 {
		*ds = EmployeePortalDocuments{}
		return nil
	}

	q := `
SELECT
	id,
	organization_id,
	employee_portal_id,
	document_id,
	created_at,
	updated_at
FROM
	employee_portal_documents
WHERE
	%s
	AND employee_portal_id = @employee_portal_id
	AND (
		NOT @has_document_ids
		OR document_id = ANY(@document_ids::text[])
	)
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	if documentIDs == nil {
		documentIDs = []gid.GID{}
	}

	args := pgx.StrictNamedArgs{
		"employee_portal_id": employeePortalID,
		"has_document_ids":   len(documentIDs) > 0,
		"document_ids":       documentIDs,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query employee portal documents: %w", err)
	}

	links, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[EmployeePortalDocument])
	if err != nil {
		return fmt.Errorf("cannot collect employee portal documents: %w", err)
	}

	*ds = links

	return nil
}

func (d *EmployeePortalDocument) Upsert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO employee_portal_documents (
	id,
	tenant_id,
	organization_id,
	employee_portal_id,
	document_id,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@employee_portal_id,
	@document_id,
	@created_at,
	@updated_at
)
ON CONFLICT (employee_portal_id, document_id) DO UPDATE
SET
	updated_at = EXCLUDED.updated_at
RETURNING
	id,
	organization_id,
	employee_portal_id,
	document_id,
	created_at,
	updated_at
`

	args := pgx.StrictNamedArgs{
		"id":                 d.ID,
		"tenant_id":          scope.GetTenantID(),
		"organization_id":    d.OrganizationID,
		"employee_portal_id": d.EmployeePortalID,
		"document_id":        d.DocumentID,
		"created_at":         d.CreatedAt,
		"updated_at":         d.UpdatedAt,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot upsert employee portal document: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[EmployeePortalDocument])
	if err != nil {
		return fmt.Errorf("cannot collect employee portal document: %w", err)
	}

	*d = row

	return nil
}

func (EmployeePortalDocument) DeleteByDocumentIDs(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	documentIDs []gid.GID,
) error {
	q := `
DELETE FROM employee_portal_documents
WHERE
	%s
	AND document_id = ANY(@document_ids::text[])
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"document_ids": documentIDs}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot delete employee portal documents by document ids: %w", err)
	}

	return nil
}

func DeleteEmployeePortalDocumentByID(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	linkID gid.GID,
) error {
	q := `
DELETE FROM employee_portal_documents
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": linkID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete employee portal document: %w", err)
	}

	return nil
}
