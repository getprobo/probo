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

	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
)

type (
	DocumentVersionSignature struct {
		ID                gid.GID                       `json:"id"`
		DocumentVersionID gid.GID                       `json:"document_version_id"`
		State             DocumentVersionSignatureState `json:"state"`
		SignedBy          gid.GID                       `json:"signed_by"`
		SignedAt          *time.Time                    `json:"signed_at"`
		RequestedAt       time.Time                     `json:"requested_at"`
		CreatedAt         time.Time                     `json:"created_at"`
		UpdatedAt         time.Time                     `json:"updated_at"`
	}

	DocumentVersionSignatures []*DocumentVersionSignature

	DocumentVersionSignatureWithPeople struct {
		DocumentVersionSignature
		SignedByFullName string `db:"signed_by_full_name"`
	}

	DocumentVersionSignaturesWithPeople []*DocumentVersionSignatureWithPeople

	ErrDocumentVersionSignatureNotFound struct {
		Identifier string
	}

	ErrDocumentVersionSignatureAlreadyExists struct {
		message string
	}
)

func (e ErrDocumentVersionSignatureNotFound) Error() string {
	return fmt.Sprintf("document version signature not found: %q", e.Identifier)
}

func (e ErrDocumentVersionSignatureAlreadyExists) Error() string {
	return e.message
}

func (pvs DocumentVersionSignature) CursorKey(orderBy DocumentVersionSignatureOrderField) page.CursorKey {
	switch orderBy {
	case DocumentVersionSignatureOrderFieldCreatedAt:
		return page.NewCursorKey(pvs.ID, pvs.CreatedAt)
	case DocumentVersionSignatureOrderFieldSignedAt:
		return page.NewCursorKey(pvs.ID, pvs.SignedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (pvs *DocumentVersionSignature) LoadByDocumentVersionIDAndSignatory(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	documentVersionID gid.GID,
	signatory gid.GID,
) error {
	q := `
SELECT
	id,
	document_version_id,
	state,
	signed_by,
	signed_at,
	requested_at,
	created_at,
	updated_at
FROM
	document_version_signatures
WHERE
	%s
	AND document_version_id = @document_version_id
	AND signed_by = @signatory
LIMIT 1
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"document_version_id": documentVersionID, "signatory": signatory}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document version signature: %w", err)
	}

	documentVersionSignature, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[DocumentVersionSignature])
	if err != nil {
		return fmt.Errorf("cannot collect document version signature: %w", err)
	}

	*pvs = documentVersionSignature

	return nil
}

func (pvs *DocumentVersionSignature) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	signatureID gid.GID,
) error {
	q := `
SELECT
	id,
	document_version_id,
	state,
	signed_by,
	signed_at,
	requested_at,
	created_at,
	updated_at
FROM
	document_version_signatures
WHERE
	id = @document_version_signature_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"document_version_signature_id": signatureID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document version signature: %w", err)
	}

	documentVersionSignature, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[DocumentVersionSignature])
	if err != nil {
		return fmt.Errorf("cannot collect document version signature: %w", err)
	}

	*pvs = documentVersionSignature

	return nil
}

func (pvs DocumentVersionSignature) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO document_version_signatures (
	id,
	tenant_id,
	document_version_id,
	state,
	signed_by,
	signed_at,
	requested_at,
	created_at,
	updated_at
) VALUES (
 	@id,
	@tenant_id,
	@document_version_id,
	@state,
	@signed_by,
	@signed_at,
	@requested_at,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                  pvs.ID,
		"tenant_id":           scope.GetTenantID(),
		"document_version_id": pvs.DocumentVersionID,
		"state":               pvs.State,
		"signed_by":           pvs.SignedBy,
		"signed_at":           pvs.SignedAt,
		"requested_at":        pvs.RequestedAt,
		"created_at":          pvs.CreatedAt,
		"updated_at":          pvs.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "policy_version_signatures_policy_version_id_signed_by_key" {
				return &ErrDocumentVersionSignatureAlreadyExists{
					message: fmt.Sprintf("document version signature with document_version_id %s and signed_by %s already exists", pvs.DocumentVersionID, pvs.SignedBy),
				}
			}
		}
		return fmt.Errorf("cannot insert document version signature: %w", err)
	}

	return nil
}

func (pvss *DocumentVersionSignatures) LoadByDocumentVersionID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	documentVersionID gid.GID,
	cursor *page.Cursor[DocumentVersionSignatureOrderField],
	filter *DocumentVersionSignatureFilter,
) error {
	q := `
SELECT
	id,
	document_version_id,
	state,
	signed_by,
	signed_at,
	requested_at,
	created_at,
	updated_at
FROM
	document_version_signatures
WHERE
	%s
	AND document_version_id = @document_version_id
	AND %s
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"document_version_id": documentVersionID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document version signatures: %w", err)
	}

	documentVersionSignatures, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[DocumentVersionSignature])
	if err != nil {
		return fmt.Errorf("cannot collect document version signatures: %w", err)
	}

	*pvss = documentVersionSignatures

	return nil
}

func (pvs *DocumentVersionSignature) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE document_version_signatures
SET
	state = @state,
	signed_by = @signed_by,
	signed_at = @signed_at,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":         pvs.ID,
		"state":      pvs.State,
		"signed_by":  pvs.SignedBy,
		"signed_at":  pvs.SignedAt,
		"updated_at": pvs.UpdatedAt,
	}

	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update document version signature: %w", err)
	}

	return nil
}

func (pvs *DocumentVersionSignature) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	documentVersionSignatureID gid.GID,
) error {
	q := `
DELETE FROM document_version_signatures
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": documentVersionSignatureID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete document version signature: %w", err)
	}

	return nil
}

func (pvss *DocumentVersionSignaturesWithPeople) LoadByDocumentVersionIDWithPeople(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	documentVersionID gid.GID,
	limit int,
) error {
	q := `
WITH sigs AS (
	SELECT
		dvs.id,
		dvs.tenant_id,
		dvs.document_version_id,
		dvs.state,
		dvs.signed_by,
		dvs.signed_at,
		dvs.requested_at,
		dvs.created_at,
		dvs.updated_at,
		p.full_name as signed_by_full_name
	FROM
		document_version_signatures dvs
	INNER JOIN
		peoples p ON dvs.signed_by = p.id
	WHERE
		dvs.document_version_id = @document_version_id
	ORDER BY
		p.full_name ASC
	LIMIT @limit
)
SELECT
	id,
	document_version_id,
	state,
	signed_by,
	signed_at,
	requested_at,
	created_at,
	updated_at,
	signed_by_full_name
FROM
	sigs
WHERE
	%s
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_version_id": documentVersionID,
		"limit":               limit,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document version signatures with people: %w", err)
	}

	signatures, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[DocumentVersionSignatureWithPeople])
	if err != nil {
		return fmt.Errorf("cannot collect document version signatures with people: %w", err)
	}

	*pvss = signatures

	return nil
}
