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

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	UserAPIKeyMembership struct {
		ID               gid.GID   `db:"id"`
		UserAPIKeyID     gid.GID   `db:"auth_user_api_key_id"`
		MembershipID     gid.GID   `db:"membership_id"`
		Role             APIRole   `db:"role"`
		OrganizationID   gid.GID   `db:"organization_id"`
		OrganizationName string    `db:"organization_name"`
		CreatedAt        time.Time `db:"created_at"`
		UpdatedAt        time.Time `db:"updated_at"`
	}

	UserAPIKeyMemberships []*UserAPIKeyMembership
)

func (a *UserAPIKeyMembership) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    authz_api_keys_memberships (id, tenant_id, auth_user_api_key_id, membership_id, role, created_at, updated_at)
VALUES (
    @id,
    @tenant_id,
    @auth_user_api_key_id,
    @membership_id,
    @role,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                   a.ID,
		"tenant_id":            scope.GetTenantID(),
		"auth_user_api_key_id": a.UserAPIKeyID,
		"membership_id":        a.MembershipID,
		"role":                 a.Role,
		"created_at":           a.CreatedAt,
		"updated_at":           a.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert user api key membership: %w", err)
	}

	return nil
}

func (a *UserAPIKeyMemberships) LoadByUserAPIKeyID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	userAPIKeyID gid.GID,
) error {
	q := `
SELECT
    akm.id,
    akm.auth_user_api_key_id,
    akm.membership_id,
    akm.role,
    akm.created_at,
    akm.updated_at,
    m.organization_id,
    o.name as organization_name
FROM
    authz_api_keys_memberships akm
JOIN
    authz_memberships m ON akm.membership_id = m.id
JOIN
    organizations o ON m.organization_id = o.id
WHERE
    akm.auth_user_api_key_id = @auth_user_api_key_id
    AND m.%s
ORDER BY akm.created_at DESC
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"auth_user_api_key_id": userAPIKeyID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query user api key memberships: %w", err)
	}

	memberships, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[UserAPIKeyMembership])
	if err != nil {
		return fmt.Errorf("cannot collect user api key memberships: %w", err)
	}

	*a = memberships

	return nil
}

func (a *UserAPIKeyMembership) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM
    authz_api_keys_memberships
WHERE
    id = @id
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id": a.ID,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete user api key membership: %w", err)
	}

	return nil
}

func DeleteAllUserAPIKeyMembershipsByUserAPIKeyID(
	ctx context.Context,
	conn pg.Conn,
	userAPIKeyID gid.GID,
) error {
	q := `
DELETE FROM
    authz_api_keys_memberships
WHERE
    auth_user_api_key_id = @auth_user_api_key_id
`

	args := pgx.StrictNamedArgs{
		"auth_user_api_key_id": userAPIKeyID,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete user api key memberships: %w", err)
	}

	return nil
}
