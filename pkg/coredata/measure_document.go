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
)

type (
	MeasureDocument struct {
		MeasureID      gid.GID      `db:"measure_id"`
		DocumentID     gid.GID      `db:"document_id"`
		OrganizationID gid.GID      `db:"organization_id"`
		TenantID       gid.TenantID `db:"tenant_id"`
		CreatedAt      time.Time    `db:"created_at"`
	}

	MeasureDocuments []*MeasureDocument
)

func (md MeasureDocument) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    measures_documents (
        measure_id,
        document_id,
        organization_id,
        tenant_id,
        created_at
    )
VALUES (
    @measure_id,
    @document_id,
    @organization_id,
    @tenant_id,
    @created_at
);
`

	args := pgx.StrictNamedArgs{
		"measure_id":      md.MeasureID,
		"document_id":     md.DocumentID,
		"organization_id": md.OrganizationID,
		"tenant_id":       scope.GetTenantID(),
		"created_at":      md.CreatedAt,
	}
	_, err := conn.Exec(ctx, q, args)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "measures_documents_pkey" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert measure document: %w", err)
	}

	return nil
}

func (md MeasureDocument) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	measureID gid.GID,
	documentID gid.GID,
) error {
	q := `
DELETE
FROM
    measures_documents
WHERE
    %s
    AND measure_id = @measure_id
    AND document_id = @document_id;
`

	args := pgx.StrictNamedArgs{
		"measure_id":  measureID,
		"document_id": documentID,
	}
	maps.Copy(args, scope.SQLArguments())

	q = fmt.Sprintf(q, scope.SQLFragment())

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (md MeasureDocument) DeleteByDocumentIDs(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	documentIDs []gid.GID,
) error {
	q := `
DELETE
FROM
    measures_documents
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
		return fmt.Errorf("cannot delete measure document mappings by document ids: %w", err)
	}

	return nil
}
