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
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
)

type (
	Membership struct {
		ID               gid.GID          `db:"id"`
		IdentityID       gid.GID          `db:"identity_id"`
		OrganizationID   gid.GID          `db:"organization_id"`
		Role             MembershipRole   `db:"role"`
		Source           MembershipSource `db:"source"`
		State            MembershipState  `db:"state"`
		FullName         string           `db:"full_name"`
		EmailAddress     mail.Addr        `db:"email_address"`
		OrganizationName string           `db:"organization_name"`
		CreatedAt        time.Time        `db:"created_at"`
		UpdatedAt        time.Time        `db:"updated_at"`
	}

	Memberships []*Membership
)

func (m Membership) CursorKey(orderBy MembershipOrderField) page.CursorKey {
	switch orderBy {
	case MembershipOrderFieldOrganizationName:
		return page.NewCursorKey(m.ID, m.OrganizationName)
	case MembershipOrderFieldFullName:
		return page.NewCursorKey(m.ID, m.FullName)
	case MembershipOrderFieldEmailAddress:
		return page.NewCursorKey(m.ID, m.EmailAddress)
	case MembershipOrderFieldRole:
		return page.NewCursorKey(m.ID, m.Role)
	case MembershipOrderFieldCreatedAt:
		return page.NewCursorKey(m.ID, m.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (m *Membership) LoadByIdentityInOrganization(ctx context.Context, conn pg.Conn, identityID gid.GID, organizationID gid.GID) error {
	q := `
WITH mbr AS (
    SELECT
        id,
        identity_id,
        organization_id,
        role,
        source,
        state,
        created_at,
        updated_at
    FROM
        iam_memberships
    WHERE
        identity_id = @identity_id
        AND organization_id = @organization_id
)
SELECT
    mbr.id,
    mbr.identity_id,
    mbr.organization_id,
    mbr.role,
    mbr.source,
    mbr.state,
    COALESCE(mp.full_name, i.full_name, '') as full_name,
    i.email_address,
    o.name AS organization_name,
    mbr.created_at,
    mbr.updated_at
FROM
    mbr
JOIN identities i ON mbr.identity_id = i.id
JOIN organizations o ON mbr.organization_id = o.id
LEFT JOIN iam_membership_profiles mp ON mp.membership_id = mbr.id
`

	args := pgx.StrictNamedArgs{
		"identity_id":     identityID,
		"organization_id": organizationID,
	}

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

func (m *Membership) Insert(ctx context.Context, conn pg.Conn, scope Scoper) error {
	query := `
INSERT INTO
    iam_memberships (
        tenant_id,
        id,
        identity_id,
        organization_id,
        role,
        source,
        state,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @id,
    @identity_id,
    @organization_id,
    @role,
    @source,
    @state,
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
		"source":          m.Source,
		"state":           m.State,
		"created_at":      m.CreatedAt,
		"updated_at":      m.UpdatedAt,
	}

	result, err := conn.Exec(ctx, query, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
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
	conn pg.Conn,
	scope Scoper,
	membershipID gid.GID,
) error {
	query := `
WITH mbr AS (
    SELECT
        id,
        identity_id,
        organization_id,
        role,
        source,
        state,
        created_at,
        updated_at
    FROM
        iam_memberships
    WHERE
        id = @membership_id
        AND %s
)
SELECT
    mbr.id,
    mbr.identity_id,
    mbr.organization_id,
    mbr.role,
    mbr.source,
    mbr.state,
    COALESCE(mp.full_name, i.full_name, '') as full_name,
    i.email_address,
    o.name as organization_name,
    mbr.created_at,
    mbr.updated_at
FROM
    mbr
JOIN
    identities i ON mbr.identity_id = i.id
JOIN
    organizations o ON mbr.organization_id = o.id
LEFT JOIN
    iam_membership_profiles mp ON mp.membership_id = mbr.id
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

func (m *Membership) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `
SELECT
    identity_id,
    organization_id,
    role,
    source
FROM
    iam_memberships
WHERE
    id = $1
LIMIT 1;
`

	var identityID gid.GID
	var organizationID gid.GID
	var role MembershipRole
	var source MembershipSource
	if err := conn.QueryRow(ctx, q, m.ID).Scan(
		&identityID,
		&organizationID,
		&role,
		&source,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query membership iam attributes: %w", err)
	}

	return map[string]string{
		"identity_id":     identityID.String(),
		"organization_id": organizationID.String(),
		"role":            role.String(),
		"source":          source.String(),
	}, nil
}

func (m *Membership) LoadByIdentityAndOrg(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	identityID gid.GID,
	organizationID gid.GID,
) error {
	q := `
WITH mbr AS (
    SELECT
        am.id,
        am.identity_id,
        am.organization_id,
        am.role,
        am.source,
        am.state,
        am.created_at,
        am.updated_at
    FROM
        iam_memberships am
    WHERE
        am.identity_id = @identity_id
        AND am.organization_id = @organization_id
        AND %s
)
SELECT
    mbr.id,
    mbr.identity_id,
    mbr.organization_id,
    mbr.role,
    mbr.source,
    mbr.state,
    COALESCE(mp.full_name, i.full_name, '') as full_name,
    i.email_address,
    o.name as organization_name,
    mbr.created_at,
    mbr.updated_at
FROM
    mbr
JOIN
    identities i ON mbr.identity_id = i.id
JOIN
    organizations o ON mbr.organization_id = o.id
LEFT JOIN
    iam_membership_profiles mp ON mp.membership_id = mbr.id
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

func (m *Membership) Update(ctx context.Context, conn pg.Conn, scope Scoper) error {
	query := `
UPDATE
    iam_memberships
SET
    role = @role,
    source = @source,
    state = @state,
    updated_at = @updated_at
WHERE
    id = @id
    AND %s
`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":         m.ID,
		"role":       m.Role,
		"source":     m.Source,
		"state":      m.State,
		"updated_at": m.UpdatedAt,
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

func (m *Membership) Delete(ctx context.Context, conn pg.Conn, scope Scoper, membershipID gid.GID) error {
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

	result, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot delete membership: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (m *Memberships) LoadByIdentityID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	identityID gid.GID,
	cursor *page.Cursor[MembershipOrderField],
) error {
	query := `
WITH mbr AS (
    SELECT
        id,
        identity_id,
        organization_id,
        role,
        source,
        state,
        created_at,
        updated_at
    FROM
        iam_memberships
    WHERE
        identity_id = @identity_id
        AND state = 'ACTIVE'
        AND %s
),
r AS (
    SELECT
        mbr.id,
        mbr.identity_id,
        mbr.organization_id,
        mbr.role,
        mbr.source,
        mbr.state,
        COALESCE(mp.full_name, i.full_name, '') as full_name,
        i.email_address,
        o.name as organization_name,
        mbr.created_at,
        mbr.updated_at
    FROM
        mbr
    JOIN
        identities i ON mbr.identity_id = i.id
    JOIN
        organizations o ON mbr.organization_id = o.id
    LEFT JOIN
        iam_membership_profiles mp ON mp.membership_id = mbr.id
)
SELECT
    id,
    identity_id,
    organization_id,
    role,
    source,
    state,
    full_name,
    email_address,
    organization_name,
    created_at,
    updated_at
FROM
    r
WHERE
    %s
`

	query = fmt.Sprintf(query, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{
		"identity_id": identityID,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot query memberships: %w", err)
	}

	memberships, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Membership])
	if err != nil {
		return fmt.Errorf("cannot collect memberships: %w", err)
	}

	*m = memberships
	return nil
}

func (m *Memberships) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[MembershipOrderField],
	filter *MembershipFilter,
) error {
	query := `
WITH m AS (
    SELECT
        id,
        identity_id,
        organization_id,
        role,
        source,
        state,
        created_at,
        updated_at
    FROM
        iam_memberships
    WHERE
        organization_id = @organization_id
        AND %s
),
r AS (
    SELECT
        m.id,
        m.identity_id,
        m.organization_id,
        m.role,
        m.source,
        m.state,
        COALESCE(mp.full_name, i.full_name, '') as full_name,
        o.name as organization_name,
        i.email_address,
        m.created_at,
        m.updated_at
    FROM
        m
    JOIN
        identities i ON m.identity_id = i.id
    JOIN
        organizations o ON m.organization_id = o.id
    LEFT JOIN
        iam_membership_profiles mp ON mp.membership_id = m.id
    WHERE
        %s
)
SELECT
    id,
    identity_id,
    organization_id,
    role,
    source,
    state,
    full_name,
    organization_name,
    email_address,
    created_at,
    updated_at
FROM
    r
WHERE
    %s
`

	query = fmt.Sprintf(query, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot query memberships: %w", err)
	}

	memberships, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Membership])
	if err != nil {
		return fmt.Errorf("cannot collect memberships: %w", err)
	}

	*m = memberships
	return nil
}

func (m *Memberships) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	filter *MembershipFilter,
) (int, error) {
	query := `
SELECT
    COUNT(*)
FROM
    iam_memberships m
JOIN
    identities i ON m.identity_id = i.id
WHERE
    m.organization_id = @organization_id
    AND m.%s
    AND %s
`
	query = fmt.Sprintf(query, scope.SQLFragment(), filter.SQLFragment())
	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	row := conn.QueryRow(ctx, query, args)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count memberships: %w", err)
	}
	return count, nil
}

func (m *Memberships) CountByIdentityID(
	ctx context.Context,
	conn pg.Conn,
	identityID gid.GID,
) (int, error) {
	query := `
SELECT
    COUNT(*)
FROM
    iam_memberships
WHERE
    identity_id = @identity_id
    AND state = 'ACTIVE'
`
	args := pgx.StrictNamedArgs{
		"identity_id": identityID,
	}

	row := conn.QueryRow(ctx, query, args)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count memberships: %w", err)
	}

	return count, nil
}

func (m *Memberships) LoadAllByIdentityID(
	ctx context.Context,
	conn pg.Conn,
	identityID gid.GID,
) error {
	q := `
SELECT
    id,
    identity_id,
    organization_id,
    role,
    source,
    state,
    '' as full_name,
    NULL as email_address,
    '' as organization_name,
    created_at,
    updated_at
FROM
    iam_memberships
WHERE
    identity_id = $1
;
`

	rows, err := conn.Query(ctx, q, identityID)
	if err != nil {
		return fmt.Errorf("cannot query memberships: %w", err)
	}

	memberships, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Membership])
	if err != nil {
		return fmt.Errorf("cannot collect memberships: %w", err)
	}

	*m = memberships
	return nil
}

func (m *Memberships) ResetSCIMSources(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) error {
	q := `
UPDATE iam_memberships
SET
    source = 'MANUAL',
    updated_at = @updated_at
WHERE
    %s
    AND organization_id = @organization_id
    AND source = 'SCIM'
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{
		"organization_id": organizationID,
		"updated_at":      time.Now(),
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot reset SCIM membership sources: %w", err)
	}

	return nil
}
