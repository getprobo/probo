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
)

type (
	ControlDocument struct {
		ControlID      gid.GID      `db:"control_id"`
		DocumentID     gid.GID      `db:"document_id"`
		OrganizationID gid.GID      `db:"organization_id"`
		TenantID       gid.TenantID `db:"tenant_id"`
		CreatedAt      time.Time    `db:"created_at"`
	}

	ControlDocuments []*ControlDocument
)

func (cp ControlDocument) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    controls_documents (
        control_id,
        document_id,
        organization_id,
        tenant_id,
        created_at
    )
VALUES (
    @control_id,
    @document_id,
    @organization_id,
    @tenant_id,
    @created_at
);
`

	args := pgx.StrictNamedArgs{
		"control_id":      cp.ControlID,
		"document_id":     cp.DocumentID,
		"organization_id": cp.OrganizationID,
		"tenant_id":       scope.GetTenantID(),
		"created_at":      cp.CreatedAt,
	}
	_, err := conn.Exec(ctx, q, args)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "controls_policies_pkey" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert control document: %w", err)
	}

	return nil
}

func (cp ControlDocument) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	controlID gid.GID,
	documentID gid.GID,
) error {
	q := `
DELETE
FROM
    controls_documents
WHERE
    %s
    AND control_id = @control_id
    AND document_id = @document_id;
`

	args := pgx.StrictNamedArgs{
		"control_id":  controlID,
		"document_id": documentID,
	}
	maps.Copy(args, scope.SQLArguments())

	q = fmt.Sprintf(q, scope.SQLFragment())

	_, err := conn.Exec(ctx, q, args)
	return err
}
