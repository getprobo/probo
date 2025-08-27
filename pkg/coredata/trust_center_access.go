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
	TrustCenterAccess struct {
		ID            gid.GID      `db:"id"`
		TenantID      gid.TenantID `db:"tenant_id"`
		TrustCenterID gid.GID      `db:"trust_center_id"`
		Email         string       `db:"email"`
		Name          string       `db:"name"`
		Active        bool         `db:"active"`
		CreatedAt     time.Time    `db:"created_at"`
		UpdatedAt     time.Time    `db:"updated_at"`
	}

	TrustCenterAccesses []*TrustCenterAccess

	ErrTrustCenterAccessNotFound struct {
		Identifier string
	}
)

func (e ErrTrustCenterAccessNotFound) Error() string {
	return fmt.Sprintf("trust center access not found: %s", e.Identifier)
}

func (tca *TrustCenterAccess) CursorKey(orderBy TrustCenterAccessOrderField) page.CursorKey {
	switch orderBy {
	case TrustCenterAccessOrderFieldCreatedAt:
		return page.NewCursorKey(tca.ID, tca.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (tca *TrustCenterAccess) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	accessID gid.GID,
) error {
	q := `
SELECT
	id,
	tenant_id,
	trust_center_id,
	email,
	name,
	active,
	created_at,
	updated_at
FROM
	trust_center_accesses
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
		return fmt.Errorf("cannot query trust center access: %w", err)
	}

	access, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TrustCenterAccess])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &ErrTrustCenterAccessNotFound{Identifier: accessID.String()}
		}

		return fmt.Errorf("cannot collect trust center access: %w", err)
	}

	*tca = access

	return nil
}

func (tca *TrustCenterAccess) LoadByTrustCenterIDAndEmail(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterID gid.GID,
	email string,
) error {
	q := `
SELECT
	id,
	tenant_id,
	trust_center_id,
	email,
	name,
	active,
	created_at,
	updated_at
FROM
	trust_center_accesses
WHERE
	%s
	AND trust_center_id = @trust_center_id
	AND email = @email
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_id": trustCenterID,
		"email":           email,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust center access: %w", err)
	}

	access, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TrustCenterAccess])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &ErrTrustCenterAccessNotFound{Identifier: fmt.Sprintf("trust_center_id=%s, email=%s", trustCenterID, email)}
		}

		return fmt.Errorf("cannot collect trust center access: %w", err)
	}

	*tca = access

	return nil
}

func (tca *TrustCenterAccess) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO trust_center_accesses (
	id,
	tenant_id,
	trust_center_id,
	email,
	name,
	active,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@trust_center_id,
	@email,
	@name,
	@active,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":              tca.ID,
		"tenant_id":       tca.TenantID,
		"trust_center_id": tca.TrustCenterID,
		"email":           tca.Email,
		"name":            tca.Name,
		"active":          tca.Active,
		"created_at":      tca.CreatedAt,
		"updated_at":      tca.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert trust center access: %w", err)
	}

	return nil
}

func (tca *TrustCenterAccess) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE trust_center_accesses SET
	name = @name,
	active = @active,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":         tca.ID,
		"name":       tca.Name,
		"active":     tca.Active,
		"updated_at": tca.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update trust center access: %w", err)
	}

	return nil
}

func (tca *TrustCenterAccess) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM trust_center_accesses
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id": tca.ID,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete trust center access: %w", err)
	}

	return nil
}

func (tcas *TrustCenterAccesses) LoadByTrustCenterID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterID gid.GID,
	cursor *page.Cursor[TrustCenterAccessOrderField],
) error {
	q := `
SELECT
	id,
	tenant_id,
	trust_center_id,
	email,
	name,
	active,
	created_at,
	updated_at
FROM
	trust_center_accesses
WHERE
	%s
	AND trust_center_id = @trust_center_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_id": trustCenterID,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust center accesses: %w", err)
	}

	accesses, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TrustCenterAccess])
	if err != nil {
		return fmt.Errorf("cannot collect trust center accesses: %w", err)
	}

	*tcas = accesses

	return nil
}
