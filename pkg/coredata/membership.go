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
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	Membership struct {
		ID             gid.GID        `db:"id"`
		IdentityID     gid.GID        `db:"identity_id"`
		OrganizationID gid.GID        `db:"organization_id"`
		Role           MembershipRole `db:"role"`
		CreatedAt      time.Time      `db:"created_at"`
		UpdatedAt      time.Time      `db:"updated_at"`
	}

	Memberships []*Membership
)

func (m Membership) CursorKey(orderBy MembershipOrderField) page.CursorKey {
	switch orderBy {
	case MembershipOrderFieldRole:
		return page.NewCursorKey(m.ID, m.Role)
	case MembershipOrderFieldCreatedAt:
		return page.NewCursorKey(m.ID, m.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (m *Membership) LoadByIdentityIDAndOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	identityID gid.GID,
	organizationID gid.GID,
) error {
	q := `
SELECT
    id,
    identity_id,
    organization_id,
    role,
    created_at,
    updated_at
FROM
    iam_memberships
WHERE
    identity_id = @identity_id
    AND organization_id = @organization_id
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"identity_id":     identityID,
		"organization_id": organizationID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows,
		err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query membership: %w", err)
	}

	membership, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Membership])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect membership: %w", err)
	}

	*m = membership

	return nil
}

func (m *Membership) Insert(ctx context.Context, conn pg.Tx, scope Scoper) error {
	query := `
INSERT INTO
    iam_memberships (
        tenant_id,
        id,
        identity_id,
        organization_id,
        role,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @id,
    @identity_id,
    @organization_id,
    @role,
    @created_at,
    @updated_at
);
`

	args := pgx.StrictNamedArgs{
		"tenant_id":       scope.GetTenantID(),
		"id":              m.ID,
		"identity_id":     m.IdentityID,
		"organization_id": m.OrganizationID,
		"role":            m.Role,
		"created_at":      m.CreatedAt,
		"updated_at":      m.UpdatedAt,
	}

	result, err := conn.Exec(ctx, query, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" && pgErr.ConstraintName == "authz_memberships_user_id_organization_id_key" {
			return ErrResourceAlreadyExists
		}

		return fmt.Errorf("cannot create membership: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("cannot create membership: organization %s not found", m.OrganizationID)
	}

	return nil
}

func (m *Membership) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	membershipID gid.GID,
) error {
	query := `
SELECT
    id,
    identity_id,
    organization_id,
    role,
    created_at,
    updated_at
FROM
    iam_memberships
WHERE
    id = @membership_id
    AND %s
`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"membership_id": membershipID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot query membership: %w", err)
	}

	membership, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Membership])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect membership: %w", err)
	}

	*m = membership

	return nil
}

func (m *Membership) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `
SELECT
    id,
    identity_id,
    organization_id,
    role
FROM
    iam_memberships
WHERE
    id = ANY(@resource_ids::text[])
`

	args := pgx.StrictNamedArgs{
		"resource_ids": resourceIDs,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query membership authorization attributes: %w", err)
	}
	defer rows.Close()

	attrsByID := make(policy.AttributesByID, len(resourceIDs))
	for rows.Next() {
		var (
			id             gid.GID
			identityID     gid.GID
			organizationID gid.GID
			role           MembershipRole
		)

		err = rows.Scan(
			&id,
			&identityID,
			&organizationID,
			&role,
		)
		if err != nil {
			return nil, fmt.Errorf("cannot scan membership authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{
			"identity_id":     identityID.String(),
			"organization_id": organizationID.String(),
			"role":            role.String(),
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate membership authorization attributes: %w", err)
	}

	return attrsByID, nil
}

func (m *Membership) LoadByIdentityAndOrg(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	identityID gid.GID,
	organizationID gid.GID,
) error {
	q := `
SELECT
    id,
    identity_id,
    organization_id,
    role,
    created_at,
    updated_at
FROM
    iam_memberships
WHERE
    identity_id = @identity_id
    AND organization_id = @organization_id
    AND %s
)
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"identity_id":     identityID,
		"organization_id": organizationID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query membership: %w", err)
	}

	membership, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Membership])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect membership: %w", err)
	}

	*m = membership

	return nil
}

func (m *Membership) Update(ctx context.Context, conn pg.Tx, scope Scoper) error {
	query := `
UPDATE
    iam_memberships
SET
    identity_id = @identity_id,
    role = @role,
    updated_at = @updated_at
WHERE
    id = @id
    AND %s
`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":          m.ID,
		"identity_id": m.IdentityID,
		"role":        m.Role,
		"updated_at":  m.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot update membership: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (m *Membership) Delete(ctx context.Context, conn pg.Tx, scope Scoper, membershipID gid.GID) error {
	query := `
DELETE FROM
    iam_memberships
WHERE
    %s
    AND id = @membership_id
`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"membership_id": membershipID,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot delete membership: %w", err)
	}

	return nil
}

func (m *Membership) LoadActiveByIdentityIDAndOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	identityID gid.GID,
	organizationID gid.GID,
) error {
	q := `
SELECT
    m.id,
    m.identity_id,
    m.organization_id,
    m.role,
    m.created_at,
    m.updated_at
FROM
    iam_memberships m
INNER JOIN iam_membership_profiles p
    ON p.identity_id = m.identity_id AND p.organization_id = m.organization_id
WHERE
    p.state = @state
    AND m.identity_id = @identity_id
    AND m.organization_id = @organization_id
LIMIT 1
`

	args := pgx.StrictNamedArgs{
		"state":           ProfileStateActive,
		"identity_id":     identityID,
		"organization_id": organizationID,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query memberships: %w", err)
	}

	membership, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[Membership])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect membership: %w", err)
	}

	*m = *membership

	return nil
}
