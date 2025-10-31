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

	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
)

type (
	Membership struct {
		ID             gid.GID   `db:"id"`
		UserID         gid.GID   `db:"user_id"`
		OrganizationID gid.GID   `db:"organization_id"`
		Role           Role      `db:"role"`
		FullName       string    `db:"full_name"`
		EmailAddress   string    `db:"email_address"`
		CreatedAt      time.Time `db:"created_at"`
		UpdatedAt      time.Time `db:"updated_at"`
	}

	Memberships []*Membership

	ErrMembershipNotFound struct {
		UserID gid.GID
		OrgID  gid.GID
	}

	ErrMembershipAlreadyExists struct {
		UserID gid.GID
		OrgID  gid.GID
	}
)

func (e ErrMembershipNotFound) Error() string {
	return fmt.Sprintf("membership not found for user %s in organization %s", e.UserID, e.OrgID)
}

func (e ErrMembershipAlreadyExists) Error() string {
	return fmt.Sprintf("membership already exists for user %s in organization %s", e.UserID, e.OrgID)
}

func (m Membership) CursorKey(orderBy MembershipOrderField) page.CursorKey {
	switch orderBy {
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

func (m *Membership) Create(ctx context.Context, conn pg.Conn, scope Scoper) error {
	query := `
INSERT INTO
    authz_memberships (
        tenant_id,
        id,
        user_id,
        organization_id,
        role,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @id,
    @user_id,
    @organization_id,
    @role,
    @created_at,
    @updated_at
);
`

	args := pgx.StrictNamedArgs{
		"tenant_id":       scope.GetTenantID(),
		"id":              m.ID,
		"user_id":         m.UserID,
		"organization_id": m.OrganizationID,
		"role":            m.Role,
		"created_at":      m.CreatedAt,
		"updated_at":      m.UpdatedAt,
	}

	result, err := conn.Exec(ctx, query, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrMembershipAlreadyExists{UserID: m.UserID, OrgID: m.OrganizationID}
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
		user_id,
		organization_id,
		role,
		created_at,
		updated_at
	FROM
		authz_memberships
	WHERE
		id = @membership_id
		AND %s
)
SELECT
    mbr.id,
    mbr.user_id,
    mbr.organization_id,
    mbr.role,
    u.fullname as full_name,
    u.email_address,
    mbr.created_at,
    mbr.updated_at
FROM
    mbr
JOIN
    users u ON mbr.user_id = u.id
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
			return ErrMembershipNotFound{UserID: gid.GID{}, OrgID: gid.GID{}}
		}
		return fmt.Errorf("cannot collect membership: %w", err)
	}

	*m = membership
	return nil
}

func (m *Membership) LoadByUserAndOrg(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	userID gid.GID,
	orgID gid.GID,
) error {
	query := `
WITH mbr AS (
	SELECT
		id,
		user_id,
		organization_id,
		role,
		created_at,
		updated_at
	FROM
		authz_memberships
	WHERE
		user_id = @user_id
		AND organization_id = @organization_id
		AND %s
)
SELECT
    mbr.id,
    mbr.user_id,
    mbr.organization_id,
    mbr.role,
    u.fullname as full_name,
    u.email_address,
    mbr.created_at,
    mbr.updated_at
FROM
    mbr
JOIN
    users u ON mbr.user_id = u.id
`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"user_id":         userID,
		"organization_id": orgID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot query membership: %w", err)
	}

	membership, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Membership])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrMembershipNotFound{UserID: userID, OrgID: orgID}
		}
		return fmt.Errorf("cannot collect membership: %w", err)
	}

	*m = membership
	return nil
}

func (m *Membership) Update(ctx context.Context, conn pg.Conn, scope Scoper) error {
	query := `
UPDATE
    authz_memberships
SET
    role = @role,
    updated_at = @updated_at
WHERE
    id = @id
    AND %s
`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":         m.ID,
		"role":       m.Role,
		"updated_at": m.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot update membership: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrMembershipNotFound{UserID: m.UserID, OrgID: m.OrganizationID}
	}

	return nil
}

func (m *Membership) Delete(ctx context.Context, conn pg.Conn, scope Scoper) error {
	query := `
DELETE FROM
    authz_memberships
WHERE
    id = @id
    AND %s
`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id": m.ID,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot delete membership: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrMembershipNotFound{UserID: m.UserID, OrgID: m.OrganizationID}
	}

	return nil
}

func (m *Memberships) LoadByUserID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	userID gid.GID,
) error {
	query := `
WITH mbr AS (
	SELECT
		id,
		user_id,
		organization_id,
		role,
		created_at,
		updated_at
	FROM
		authz_memberships
	WHERE
		user_id = @user_id
		AND %s
	ORDER BY
		created_at DESC
)
SELECT
    mbr.id,
    mbr.user_id,
    mbr.organization_id,
    mbr.role,
    u.fullname as full_name,
    u.email_address,
    mbr.created_at,
    mbr.updated_at
FROM
    mbr
JOIN
    users u ON mbr.user_id = u.id
ORDER BY
    mbr.created_at DESC
`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"user_id": userID,
	}
	maps.Copy(args, scope.SQLArguments())

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
) error {
	query := `
WITH mbr AS (
	SELECT
		id,
		user_id,
		organization_id,
		role,
		created_at,
		updated_at
	FROM
		authz_memberships
	WHERE
		organization_id = @organization_id
		AND %s
)
SELECT
    id,
    user_id,
    organization_id,
    role,
    full_name,
    email_address,
    created_at,
    updated_at
FROM (
	SELECT
		mbr.id,
		mbr.user_id,
		mbr.organization_id,
		mbr.role,
		u.fullname as full_name,
		u.email_address,
		mbr.created_at,
		mbr.updated_at
	FROM
		mbr
	JOIN
		users u ON mbr.user_id = u.id
) AS membership_with_user
WHERE %s
`

	query = fmt.Sprintf(query, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
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

func (m *Memberships) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	query := `
SELECT
    COUNT(*)
FROM
    authz_memberships
WHERE
    organization_id = @organization_id
    AND %s
`
	query = fmt.Sprintf(query, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
	}
	maps.Copy(args, scope.SQLArguments())
	row := conn.QueryRow(ctx, query, args)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count memberships: %w", err)
	}
	return count, nil
}

func LoadUserIDsByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) ([]gid.GID, error) {
	query := `
SELECT user_id
FROM authz_memberships
WHERE organization_id = @organization_id AND %s
`
	query = fmt.Sprintf(query, scope.SQLFragment())
	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query memberships: %w", err)
	}

	var userIDs []gid.GID
	for rows.Next() {
		var userID gid.GID
		if err := rows.Scan(&userID); err != nil {
			rows.Close()
			return nil, fmt.Errorf("cannot scan user_id: %w", err)
		}
		userIDs = append(userIDs, userID)
	}
	rows.Close()

	return userIDs, nil
}

func UpdateMembershipUserID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	oldUserID gid.GID,
	newUserID gid.GID,
	organizationID gid.GID,
) error {
	query := `
UPDATE authz_memberships
SET user_id = @new_user_id, updated_at = @updated_at
WHERE user_id = @old_user_id AND organization_id = @organization_id AND %s
`
	query = fmt.Sprintf(query, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"new_user_id":     newUserID,
		"old_user_id":     oldUserID,
		"organization_id": organizationID,
		"updated_at":      time.Now(),
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot update membership: %w", err)
	}

	return nil
}
