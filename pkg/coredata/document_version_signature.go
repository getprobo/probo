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
)

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
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"document_version_id": documentVersionID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

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
