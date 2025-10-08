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
	Membership struct {
		ID             gid.GID   `db:"id"`
		UserID         gid.GID   `db:"user_id"`
		OrganizationID gid.GID   `db:"organization_id"`
		Role           string    `db:"role"`
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

// Tenant id scope is not applied because memberships are managed at the organization level and don't require tenant isolation.
func (m *Membership) Create(ctx context.Context, conn pg.Conn) error {
	query := `
		INSERT INTO authz_memberships (id, user_id, organization_id, role, created_at, updated_at)
		SELECT
			generate_gid(decode_base64_unpadded(o.tenant_id), @entity_type),
			@user_id,
			@organization_id,
			@role,
			@created_at,
			@updated_at
		FROM organizations o
		WHERE o.id = @organization_id
	`

	args := pgx.StrictNamedArgs{
		"user_id":         m.UserID,
		"organization_id": m.OrganizationID,
		"role":            m.Role,
		"created_at":      m.CreatedAt,
		"updated_at":      m.UpdatedAt,
		"entity_type":     MembershipEntityType,
	}

	result, err := conn.Exec(ctx, query, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrMembershipAlreadyExists{UserID: m.UserID, OrgID: m.OrganizationID}
		}
		return fmt.Errorf("failed to create membership: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("failed to create membership: organization %s not found", m.OrganizationID)
	}

	return nil
}

// Tenant id scope is not applied because we want to access memberships across all tenants for authentication purposes.
func (m *Membership) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	membershipID gid.GID,
) error {
	query := `
		SELECT
			m.id,
			m.user_id,
			m.organization_id,
			m.role,
			u.fullname as full_name,
			u.email_address,
			m.created_at,
			m.updated_at
		FROM authz_memberships m
		JOIN users u ON m.user_id = u.id
		WHERE m.id = @membership_id
	`

	args := pgx.StrictNamedArgs{
		"membership_id": membershipID,
	}

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

// Tenant id scope is not applied because we want to access memberships across all tenants for authentication purposes.
func (m *Membership) LoadByUserAndOrg(
	ctx context.Context,
	conn pg.Conn,
	userID gid.GID,
	orgID gid.GID,
) error {
	query := `
		SELECT
			m.id,
			m.user_id,
			m.organization_id,
			m.role,
			u.fullname as full_name,
			u.email_address,
			m.created_at,
			m.updated_at
		FROM authz_memberships m
		JOIN users u ON m.user_id = u.id
		WHERE m.user_id = @user_id AND m.organization_id = @organization_id
	`

	args := pgx.StrictNamedArgs{
		"user_id":         userID,
		"organization_id": orgID,
	}

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

// Tenant id scope is not applied because memberships are managed at the organization level and don't require tenant isolation.
func (m *Membership) Update(ctx context.Context, conn pg.Conn) error {
	query := `
		UPDATE authz_memberships
		SET role = @role, updated_at = @updated_at
		WHERE id = @id
	`

	args := pgx.StrictNamedArgs{
		"id":         m.ID,
		"role":       m.Role,
		"updated_at": m.UpdatedAt,
	}

	result, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("failed to update membership: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrMembershipNotFound{UserID: m.UserID, OrgID: m.OrganizationID}
	}

	return nil
}

// Tenant id scope is not applied because memberships are managed at the organization level and don't require tenant isolation.
func (m *Membership) Delete(ctx context.Context, conn pg.Conn) error {
	query := `
		DELETE FROM authz_memberships
		WHERE id = @id
	`

	args := pgx.StrictNamedArgs{
		"id": m.ID,
	}

	result, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("failed to delete membership: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrMembershipNotFound{UserID: m.UserID, OrgID: m.OrganizationID}
	}

	return nil
}

// Tenant id scope is not applied because we want to access all user's memberships across tenants for authentication purposes.
func (m *Memberships) LoadByUserID(
	ctx context.Context,
	conn pg.Conn,
	userID gid.GID,
) error {
	query := `
SELECT
	m.id,
	m.user_id,
	m.organization_id,
	m.role,
	u.fullname as full_name,
	u.email_address,
	m.created_at,
	m.updated_at
FROM
	authz_memberships m
JOIN users u ON m.user_id = u.id
WHERE
	m.user_id = @user_id
ORDER BY
	m.created_at DESC
	`

	args := pgx.StrictNamedArgs{"user_id": userID}

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

// Tenant id scope is not applied because we want to access memberships across all tenants for authentication purposes.
func (m *Memberships) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	organizationID gid.GID,
	cursor *page.Cursor[MembershipOrderField],
) error {
	query := `
SELECT
	m.id,
	m.user_id,
	m.organization_id,
	m.role,
	u.fullname as full_name,
	u.email_address,
	m.created_at,
	m.updated_at
FROM
	authz_memberships m
JOIN users u ON m.user_id = u.id
WHERE
	m.organization_id = @organization_id
	AND %s
`

	query = fmt.Sprintf(query, cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
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
	organizationID gid.GID,
) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM authz_memberships
		WHERE organization_id = @organization_id
	`
	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	row := conn.QueryRow(ctx, query, args)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count memberships: %w", err)
	}
	return count, nil
}
