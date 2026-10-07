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
)

type (
	InternalControlDocument struct {
		InternalControlID gid.GID      `db:"internal_control_id"`
		DocumentID        gid.GID      `db:"document_id"`
		OrganizationID    gid.GID      `db:"organization_id"`
		TenantID          gid.TenantID `db:"tenant_id"`
		CreatedAt         time.Time    `db:"created_at"`
	}

	InternalControlDocuments []*InternalControlDocument
)

func (md InternalControlDocument) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    internal_controls_documents (
        internal_control_id,
        document_id,
        organization_id,
        tenant_id,
        created_at
    )
VALUES (
    @internal_control_id,
    @document_id,
    @organization_id,
    @tenant_id,
    @created_at
);
`

	args := pgx.StrictNamedArgs{
		"internal_control_id": md.InternalControlID,
		"document_id":         md.DocumentID,
		"organization_id":     md.OrganizationID,
		"tenant_id":           scope.GetTenantID(),
		"created_at":          md.CreatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "internal_controls_documents_pkey" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert internal control document: %w", err)
	}

	return nil
}

func (md InternalControlDocument) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	internalControlID gid.GID,
	documentID gid.GID,
) error {
	q := `
DELETE
FROM
    internal_controls_documents
WHERE
    %s
    AND internal_control_id = @internal_control_id
    AND document_id = @document_id;
`

	args := pgx.StrictNamedArgs{
		"internal_control_id": internalControlID,
		"document_id":         documentID,
	}
	maps.Copy(args, scope.SQLArguments())

	q = fmt.Sprintf(q, scope.SQLFragment())

	_, err := conn.Exec(ctx, q, args)

	return err
}

func (md InternalControlDocument) DeleteByDocumentIDs(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	documentIDs []gid.GID,
) error {
	q := `
DELETE
FROM
    internal_controls_documents
WHERE
    %s
    AND document_id = ANY(@document_ids);
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_ids": documentIDs,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot delete internal control document mappings by document ids: %w", err)
	}

	return nil
}

func (md *InternalControlDocuments) DeleteByOrganizationID(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	organizationID gid.GID,
) error {
	q := `
DELETE FROM internal_controls_documents
WHERE
	%s
	AND organization_id = @organization_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete internal control document mappings: %w", err)
	}

	return nil
}
