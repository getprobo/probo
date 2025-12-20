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
	PersonalAPIKeyMembership struct {
		ID               gid.GID   `db:"id"`
		PersonalAPIKeyID gid.GID   `db:"auth_personal_api_key_id"`
		MembershipID     gid.GID   `db:"membership_id"`
		Role             APIRole   `db:"role"`
		OrganizationID   gid.GID   `db:"organization_id"`
		OrganizationName string    `db:"organization_name"`
		CreatedAt        time.Time `db:"created_at"`
		UpdatedAt        time.Time `db:"updated_at"`
	}

	PersonalAPIKeyMemberships []*PersonalAPIKeyMembership
)

func (a *PersonalAPIKeyMembership) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    authz_api_keys_memberships (id, tenant_id, auth_personal_api_key_id, membership_id, role, organization_id, created_at, updated_at)
VALUES (
    @id,
    @tenant_id,
    @auth_personal_api_key_id,
    @membership_id,
    @role,
    @organization_id,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                       a.ID,
		"tenant_id":                scope.GetTenantID(),
		"auth_personal_api_key_id": a.PersonalAPIKeyID,
		"membership_id":            a.MembershipID,
		"role":                     a.Role,
		"organization_id":          a.OrganizationID,
		"created_at":               a.CreatedAt,
		"updated_at":               a.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert personal api key membership: %w", err)
	}

	return nil
}

func (a *PersonalAPIKeyMemberships) LoadByPersonalAPIKeyID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	personalAPIKeyID gid.GID,
) error {
	q := `
SELECT
    akm.id,
    akm.auth_personal_api_key_id,
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
    akm.auth_personal_api_key_id = @auth_personal_api_key_id
    AND m.%s
ORDER BY akm.created_at DESC
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"auth_personal_api_key_id": personalAPIKeyID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query personal api key memberships: %w", err)
	}

	memberships, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[PersonalAPIKeyMembership])
	if err != nil {
		return fmt.Errorf("cannot collect personal api key memberships: %w", err)
	}

	*a = memberships

	return nil
}

// LoadRoleByAPIKeyAndEntityID loads an API key's role by querying any entity to extract its organization_id
func (a *PersonalAPIKeyMembership) LoadRoleByAPIKeyAndEntityID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	apiKeyID gid.GID,
	entityID gid.GID,
) error {
	entityType := entityID.EntityType()

	// For organization, the entity ID is the organization ID
	if entityType == OrganizationEntityType {
		return a.LoadByAPIKeyIDAndOrganizationID(ctx, conn, scope, apiKeyID, entityID)
	}

	tableName, ok := EntityTable(entityType)
	if !ok {
		return fmt.Errorf("unsupported entity type for API key role lookup: %d", entityType)
	}

	query := fmt.Sprintf(`
SELECT
	akm.id,
	akm.auth_personal_api_key_id,
	akm.membership_id,
	akm.role,
	akm.created_at,
	akm.updated_at
FROM
	authz_api_keys_memberships akm
	INNER JOIN authz_memberships m ON m.id = akm.membership_id
	INNER JOIN %s e ON e.id = @entity_id
WHERE
	%s
	AND akm.auth_personal_api_key_id = @api_key_id
	AND m.organization_id = e.organization_id
LIMIT 1;
`, tableName, scope.SQLFragment())

	args := pgx.NamedArgs{
		"api_key_id": apiKeyID,
		"entity_id":  entityID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot query API key membership by entity: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return fmt.Errorf("API key membership not found for key %s and entity %s", apiKeyID, entityID)
	}

	var membership PersonalAPIKeyMembership
	err = rows.Scan(
		&membership.ID,
		&membership.PersonalAPIKeyID,
		&membership.MembershipID,
		&membership.Role,
		&membership.CreatedAt,
		&membership.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("cannot scan API key membership: %w", err)
	}

	*a = membership
	return nil
}

func (a *PersonalAPIKeyMembership) LoadByAPIKeyIDAndOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	apiKeyID gid.GID,
	organizationID gid.GID,
) error {
	q := `
SELECT
    akm.id,
    akm.auth_personal_api_key_id,
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
    akm.auth_personal_api_key_id = @api_key_id
    AND m.organization_id = @organization_id
    AND m.%s
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"api_key_id":      apiKeyID,
		"organization_id": organizationID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query personal api key membership: %w", err)
	}

	membership, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[PersonalAPIKeyMembership])
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("API key does not have access to organization")
		}
		return fmt.Errorf("cannot collect personal api key membership: %w", err)
	}

	*a = membership
	return nil
}

func (a *PersonalAPIKeyMembership) Delete(
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
		return fmt.Errorf("cannot delete personal api key membership: %w", err)
	}

	return nil
}

func (a *PersonalAPIKeyMemberships) LoadByMembershipID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	membershipID gid.GID,
) error {
	q := `
SELECT
    akm.id,
    akm.auth_personal_api_key_id,
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
    akm.membership_id = @membership_id
    AND m.%s
ORDER BY akm.created_at DESC
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"membership_id": membershipID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query personal api key memberships by membership id: %w", err)
	}

	memberships, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[PersonalAPIKeyMembership])
	if err != nil {
		return fmt.Errorf("cannot collect personal api key memberships: %w", err)
	}

	*a = memberships

	return nil
}

func DeleteAllPersonalAPIKeyMembershipsByPersonalAPIKeyID(
	ctx context.Context,
	conn pg.Conn,
	personalAPIKeyID gid.GID,
) error {
	q := `
DELETE FROM
    authz_api_keys_memberships
WHERE
    auth_personal_api_key_id = @auth_personal_api_key_id
`

	args := pgx.StrictNamedArgs{
		"auth_personal_api_key_id": personalAPIKeyID,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete personal api key memberships: %w", err)
	}

	return nil
}
