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
	// OrganizationMembership represents a user's membership in an organization
	OrganizationMembership struct {
		UserID         gid.GID   `db:"user_id"`
		OrganizationID gid.GID   `db:"organization_id"`
		Role           string    `db:"role"`
		CreatedAt      time.Time `db:"created_at"`
		UpdatedAt      time.Time `db:"updated_at"`
	}

	// OrganizationMemberships is a slice of OrganizationMembership
	OrganizationMemberships []*OrganizationMembership

	// ErrMembershipNotFound is returned when a membership is not found
	ErrMembershipNotFound struct {
		UserID gid.GID
		OrgID  gid.GID
	}

	// ErrMembershipAlreadyExists is returned when trying to create a duplicate membership
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

// CursorKey returns the cursor key for pagination
func (m OrganizationMembership) CursorKey(orderBy MembershipOrderField) page.CursorKey {
	switch orderBy {
	case MembershipOrderFieldCreatedAt:
		return page.NewCursorKey(m.UserID, m.CreatedAt)
	case MembershipOrderFieldUpdatedAt:
		return page.NewCursorKey(m.UserID, m.UpdatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

// Create creates a new organization membership
func (m *OrganizationMembership) Create(ctx context.Context, conn pg.Conn) error {
	query := `
		INSERT INTO authz_memberships (user_id, organization_id, role, created_at, updated_at)
		VALUES (@user_id, @organization_id, @role, @created_at, @updated_at)
	`

	args := pgx.StrictNamedArgs{
		"user_id":         m.UserID,
		"organization_id": m.OrganizationID,
		"role":            m.Role,
		"created_at":      m.CreatedAt,
		"updated_at":      m.UpdatedAt,
	}

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return ErrMembershipAlreadyExists{UserID: m.UserID, OrgID: m.OrganizationID}
		}
		return fmt.Errorf("failed to create membership: %w", err)
	}

	return nil
}

// LoadByUserAndOrg loads a membership by user and organization ID
func (m *OrganizationMembership) LoadByUserAndOrg(
	ctx context.Context,
	conn pg.Conn,
	userID gid.GID,
	orgID gid.GID,
) error {
	query := `
		SELECT user_id, organization_id, role, created_at, updated_at
		FROM authz_memberships
		WHERE user_id = @user_id AND organization_id = @organization_id
	`

	args := pgx.StrictNamedArgs{
		"user_id":         userID,
		"organization_id": orgID,
	}

	rows, err := conn.Query(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot query membership: %w", err)
	}

	membership, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[OrganizationMembership])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrMembershipNotFound{UserID: userID, OrgID: orgID}
		}
		return fmt.Errorf("cannot collect membership: %w", err)
	}

	*m = membership
	return nil
}

// Update updates an existing membership
func (m *OrganizationMembership) Update(ctx context.Context, conn pg.Conn) error {
	query := `
		UPDATE authz_memberships
		SET role = @role, updated_at = @updated_at
		WHERE user_id = @user_id AND organization_id = @organization_id
	`

	args := pgx.StrictNamedArgs{
		"user_id":         m.UserID,
		"organization_id": m.OrganizationID,
		"role":            m.Role,
		"updated_at":      m.UpdatedAt,
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

// Delete deletes a membership
func (m *OrganizationMembership) Delete(ctx context.Context, conn pg.Conn) error {
	query := `
		DELETE FROM authz_memberships
		WHERE user_id = @user_id AND organization_id = @organization_id
	`

	args := pgx.StrictNamedArgs{
		"user_id":         m.UserID,
		"organization_id": m.OrganizationID,
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

// LoadByUserID loads all memberships for a specific user
func (m *OrganizationMemberships) LoadByUserID(
	ctx context.Context,
	conn pg.Conn,
	userID gid.GID,
) error {
	query := `
SELECT
	user_id,
	organization_id,
	role,
	created_at,
	updated_at
FROM
	authz_memberships
WHERE
	user_id = @user_id
ORDER BY
	created_at DESC
	`

	args := pgx.StrictNamedArgs{"user_id": userID}

	rows, err := conn.Query(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot query memberships: %w", err)
	}

	memberships, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[OrganizationMembership])
	if err != nil {
		return fmt.Errorf("cannot collect memberships: %w", err)
	}

	*m = memberships
	return nil
}

// LoadByOrganizationID loads all memberships for a specific organization with pagination
func (m *OrganizationMemberships) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	orgID gid.GID,
	cursor *page.Cursor[MembershipOrderField],
) error {
	query := `
SELECT
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
`

	query = fmt.Sprintf(query, cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": orgID}
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot query memberships: %w", err)
	}

	memberships, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[OrganizationMembership])
	if err != nil {
		return fmt.Errorf("cannot collect memberships: %w", err)
	}

	*m = memberships
	return nil
}
