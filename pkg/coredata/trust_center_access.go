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
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	TrustCenterAccess struct {
		ID                    gid.GID      `db:"id"`
		OrganizationID        gid.GID      `db:"organization_id"`
		TenantID              gid.TenantID `db:"tenant_id"`
		IdentityID            gid.GID      `db:"identity_id"`
		TrustCenterID         gid.GID      `db:"trust_center_id"`
		ElectronicSignatureID *gid.GID     `db:"electronic_signature_id"`
		CreatedAt             time.Time    `db:"created_at"`
		UpdatedAt             time.Time    `db:"updated_at"`
	}

	TrustCenterAccesses []*TrustCenterAccess
)

func (tca *TrustCenterAccess) CursorKey(orderBy TrustCenterAccessOrderField) page.CursorKey {
	switch orderBy {
	case TrustCenterAccessOrderFieldCreatedAt:
		return page.NewCursorKey(tca.ID, tca.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (tca *TrustCenterAccess) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM trust_center_accesses WHERE id = ANY(@resource_ids::text[])`

	args := pgx.StrictNamedArgs{
		"resource_ids": resourceIDs,
	}

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

func (tca *TrustCenterAccess) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	accessID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	tenant_id,
	identity_id,
	trust_center_id,
	electronic_signature_id,
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
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect trust center access: %w", err)
	}

	*tca = access

	return nil
}

func (tca *TrustCenterAccess) LoadByTrustCenterIDAndIdentityID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	trustCenterID gid.GID,
	identityID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	tenant_id,
	identity_id,
	trust_center_id,
	electronic_signature_id,
	created_at,
	updated_at
FROM
	trust_center_accesses
WHERE
	%s
	AND trust_center_id = @trust_center_id
	AND identity_id = @identity_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_id": trustCenterID,
		"identity_id":     identityID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust center access: %w", err)
	}

	access, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TrustCenterAccess])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect trust center access: %w", err)
	}

	*tca = access

	return nil
}

func (tca *TrustCenterAccess) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO trust_center_accesses (
	id,
	tenant_id,
	organization_id,
	identity_id,
	trust_center_id,
	electronic_signature_id,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@identity_id,
	@trust_center_id,
	@electronic_signature_id,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                      tca.ID,
		"tenant_id":               tca.TenantID,
		"organization_id":         tca.OrganizationID,
		"identity_id":             tca.IdentityID,
		"trust_center_id":         tca.TrustCenterID,
		"electronic_signature_id": tca.ElectronicSignatureID,
		"created_at":              tca.CreatedAt,
		"updated_at":              tca.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "trust_center_accesses_trust_center_id_email_key" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert trust center access: %w", err)
	}

	return nil
}

func (tca *TrustCenterAccess) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE trust_center_accesses SET
	updated_at = @updated_at,
	electronic_signature_id = @electronic_signature_id
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                      tca.ID,
		"updated_at":              tca.UpdatedAt,
		"electronic_signature_id": tca.ElectronicSignatureID,
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
	conn pg.Tx,
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
	conn pg.Querier,
	scope Scoper,
	trustCenterID gid.GID,
	cursor *page.Cursor[TrustCenterAccessOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	tenant_id,
	identity_id,
	trust_center_id,
	electronic_signature_id,
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
