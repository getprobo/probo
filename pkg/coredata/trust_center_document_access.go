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
	"go.gearno.de/kit/pg"
)

type (
	TrustCenterDocumentAccess struct {
		ID                  gid.GID   `db:"id"`
		TrustCenterAccessID gid.GID   `db:"trust_center_access_id"`
		DocumentID          *gid.GID  `db:"document_id"`
		ReportID            *gid.GID  `db:"report_id"`
		Active              bool      `db:"active"`
		CreatedAt           time.Time `db:"created_at"`
		UpdatedAt           time.Time `db:"updated_at"`
	}

	TrustCenterDocumentAccesses []*TrustCenterDocumentAccess

	ErrTrustCenterDocumentAccessNotFound struct {
		Identifier string
	}
)

func (e ErrTrustCenterDocumentAccessNotFound) Error() string {
	return fmt.Sprintf("trust center document access not found: %s", e.Identifier)
}

func (tcda *TrustCenterDocumentAccess) CursorKey(orderBy TrustCenterDocumentAccessOrderField) page.CursorKey {
	switch orderBy {
	case TrustCenterDocumentAccessOrderFieldCreatedAt:
		return page.NewCursorKey(tcda.ID, tcda.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (tcda *TrustCenterDocumentAccess) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	accessID gid.GID,
) error {
	q := `
SELECT
	id,
	trust_center_access_id,
	document_id,
	report_id,
	active,
	created_at,
	updated_at
FROM
	trust_center_document_accesses
WHERE
	%s
	AND id = @access_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"access_id": accessID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust center document access: %w", err)
	}

	access, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TrustCenterDocumentAccess])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &ErrTrustCenterDocumentAccessNotFound{Identifier: accessID.String()}
		}

		return fmt.Errorf("cannot collect trust center document access: %w", err)
	}

	*tcda = access

	return nil
}

func (tcda *TrustCenterDocumentAccess) LoadByTrustCenterAccessIDAndDocumentID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterAccessID gid.GID,
	documentID gid.GID,
) error {
	q := `
SELECT
	id,
	trust_center_access_id,
	document_id,
	report_id,
	active,
	created_at,
	updated_at
FROM
	trust_center_document_accesses
WHERE
	%s
	AND trust_center_access_id = @trust_center_access_id
	AND document_id = @document_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_access_id": trustCenterAccessID,
		"document_id":            documentID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust center document access: %w", err)
	}

	access, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TrustCenterDocumentAccess])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &ErrTrustCenterDocumentAccessNotFound{Identifier: fmt.Sprintf("access_id=%s,document_id=%s", trustCenterAccessID, documentID)}
		}

		return fmt.Errorf("cannot collect trust center document access: %w", err)
	}

	*tcda = access

	return nil
}

func (tcda *TrustCenterDocumentAccess) LoadByTrustCenterAccessIDAndReportID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterAccessID gid.GID,
	reportID gid.GID,
) error {
	q := `
SELECT
	id,
	trust_center_access_id,
	document_id,
	report_id,
	active,
	created_at,
	updated_at
FROM
	trust_center_document_accesses
WHERE
	%s
	AND trust_center_access_id = @trust_center_access_id
	AND report_id = @report_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_access_id": trustCenterAccessID,
		"report_id":              reportID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust center document access: %w", err)
	}

	access, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TrustCenterDocumentAccess])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &ErrTrustCenterDocumentAccessNotFound{Identifier: fmt.Sprintf("access_id=%s,report_id=%s", trustCenterAccessID, reportID)}
		}

		return fmt.Errorf("cannot collect trust center document access: %w", err)
	}

	*tcda = access

	return nil
}

func (tcda *TrustCenterDocumentAccess) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO trust_center_document_accesses (
	id,
	tenant_id,
	trust_center_access_id,
	document_id,
	report_id,
	active,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@trust_center_access_id,
	@document_id,
	@report_id,
	@active,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                     tcda.ID,
		"tenant_id":              scope.GetTenantID(),
		"trust_center_access_id": tcda.TrustCenterAccessID,
		"document_id":            tcda.DocumentID,
		"report_id":              tcda.ReportID,
		"active":                 tcda.Active,
		"created_at":             tcda.CreatedAt,
		"updated_at":             tcda.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert trust center document access: %w", err)
	}

	return nil
}

func (tcda *TrustCenterDocumentAccess) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE trust_center_document_accesses SET
	active = @active,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":         tcda.ID,
		"active":     tcda.Active,
		"updated_at": tcda.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update trust center document access: %w", err)
	}

	return nil
}

func (tcda *TrustCenterDocumentAccess) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM trust_center_document_accesses
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id": tcda.ID,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete trust center document access: %w", err)
	}

	return nil
}

func (tcdas *TrustCenterDocumentAccesses) LoadByTrustCenterAccessID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterAccessID gid.GID,
	cursor *page.Cursor[TrustCenterDocumentAccessOrderField],
) error {
	q := `
SELECT
	id,
	trust_center_access_id,
	document_id,
	report_id,
	active,
	created_at,
	updated_at
FROM
	trust_center_document_accesses
WHERE
	%s
	AND trust_center_access_id = @trust_center_access_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_access_id": trustCenterAccessID,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust center document accesses: %w", err)
	}

	accesses, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TrustCenterDocumentAccess])
	if err != nil {
		return fmt.Errorf("cannot collect trust center document accesses: %w", err)
	}

	*tcdas = accesses

	return nil
}

func DeleteByTrustCenterAccessID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterAccessID gid.GID,
) error {
	q := `
DELETE FROM trust_center_document_accesses
WHERE
	%s
	AND trust_center_access_id = @trust_center_access_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_access_id": trustCenterAccessID,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete trust center document accesses: %w", err)
	}

	return nil
}
