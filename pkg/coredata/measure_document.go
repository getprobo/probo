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
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
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
