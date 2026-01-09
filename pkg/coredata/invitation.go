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
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
)

type (
	Invitation struct {
		ID             gid.GID          `db:"id"`
		OrganizationID gid.GID          `db:"organization_id"`
		Email          mail.Addr        `db:"email"`
		FullName       string           `db:"full_name"`
		Role           MembershipRole   `db:"role"`
		Status         InvitationStatus `db:"status"`
		ExpiresAt      time.Time        `db:"expires_at"`
		AcceptedAt     *time.Time       `db:"accepted_at"`
		CreatedAt      time.Time        `db:"created_at"`
	}

	Invitations []*Invitation
)

func (i Invitation) CursorKey(orderBy InvitationOrderField) page.CursorKey {
	switch orderBy {
	case InvitationOrderFieldFullName:
		return page.NewCursorKey(i.ID, i.FullName)
	case InvitationOrderFieldEmail:
		return page.NewCursorKey(i.ID, i.Email)
	case InvitationOrderFieldRole:
		return page.NewCursorKey(i.ID, i.Role)
	case InvitationOrderFieldCreatedAt:
		return page.NewCursorKey(i.ID, i.CreatedAt)
	case InvitationOrderFieldExpiresAt:
		return page.NewCursorKey(i.ID, i.ExpiresAt)
	case InvitationOrderFieldAcceptedAt:
		acceptedAt := time.Time{}
		if i.AcceptedAt != nil {
			acceptedAt = *i.AcceptedAt
		}
		return page.NewCursorKey(i.ID, acceptedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (i *Invitation) Insert(ctx context.Context, conn pg.Conn, scope Scoper) error {
	query := `
INSERT INTO
    iam_invitations (
        tenant_id,
        id,
        organization_id,
        email,
        full_name,
        role,
        expires_at,
        created_at
    )
VALUES (
    @tenant_id,
    @id,
    @organization_id,
    @email,
    @full_name,
    @role,
    @expires_at,
    @created_at
);
`

	args := pgx.StrictNamedArgs{
		"tenant_id":       scope.GetTenantID(),
		"id":              i.ID,
		"organization_id": i.OrganizationID,
		"email":           i.Email,
		"full_name":       i.FullName,
		"role":            i.Role,
		"expires_at":      i.ExpiresAt,
		"created_at":      i.CreatedAt,
	}

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot create invitation: %w", err)
	}

	return nil
}

func (i *Invitation) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	id gid.GID,
) error {
	query := `
SELECT
    id,
    organization_id,
    email,
    full_name,
    role,
    CASE
        WHEN accepted_at IS NOT NULL THEN 'ACCEPTED'
        WHEN expires_at < NOW() THEN 'EXPIRED'
        ELSE 'PENDING'
    END as status,
    expires_at,
    accepted_at,
    created_at
FROM
    iam_invitations
WHERE
    id = @id
    AND %s
`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id": id,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot query invitation: %w", err)
	}

	invitation, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Invitation])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect invitation: %w", err)
	}

	*i = invitation
	return nil
}

func (i *Invitations) ExpireByEmailAndOrganization(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	email mail.Addr,
	organizationID gid.GID,
	filter *InvitationFilter,
) error {
	q := `
UPDATE
    iam_invitations
SET
    expires_at = NOW()
WHERE
    email = @email
    AND organization_id = @organization_id
    AND %s
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{
		"email":           email,
		"organization_id": organizationID,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot expire invitations: %w", err)
	}

	return nil
}

// AuthorizationAttributes loads the minimal authorization attributes for policy condition evaluation.
// It is intentionally lightweight and does not populate the Invitation struct.
func (i *Invitation) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `
SELECT
    email
    , organization_id
FROM
    iam_invitations
WHERE
    id = $1
LIMIT 1;
`

	var email string
	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, i.ID).Scan(&email, &organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query invitation iam attributes: %w", err)
	}

	return map[string]string{
		"email":           email,
		"organization_id": organizationID.String(),
	}, nil
}

func (i *Invitation) Update(ctx context.Context, conn pg.Conn, scope Scoper) error {
	query := `
UPDATE
    iam_invitations
SET
    accepted_at = @accepted_at
WHERE
    id = @id
    AND %s
`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":          i.ID,
		"accepted_at": i.AcceptedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot update invitation: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (i *Invitation) Delete(ctx context.Context, conn pg.Conn, scope Scoper, invitationID gid.GID) error {
	query := `
DELETE FROM
    iam_invitations
WHERE
    %s
    AND id = @invitation_id
`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"invitation_id": invitationID,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot delete invitation: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (i *Invitations) LoadByIdentityID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	email mail.Addr,
	cursor *page.Cursor[InvitationOrderField],
	filter *InvitationFilter,
) error {
	query := `
SELECT
    id,
    organization_id,
    email,
    full_name,
    role,
    CASE
        WHEN accepted_at IS NOT NULL THEN 'ACCEPTED'
        WHEN expires_at < NOW() THEN 'EXPIRED'
        ELSE 'PENDING'
    END as status,
    expires_at,
    accepted_at,
    created_at
FROM
    iam_invitations
WHERE
    email = @email
    AND %s
    AND %s
`

	query = fmt.Sprintf(query, filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{
		"email": email,
	}
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot query invitations: %w", err)
	}

	invitations, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Invitation])
	if err != nil {
		return fmt.Errorf("cannot collect invitations: %w", err)
	}

	*i = invitations
	return nil
}

func (i *Invitations) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	orgID gid.GID,
	cursor *page.Cursor[InvitationOrderField],
	filter *InvitationFilter,
) error {
	query := `
SELECT
    id,
    organization_id,
    email,
    full_name,
    role,
    CASE
        WHEN accepted_at IS NOT NULL THEN 'ACCEPTED'
        WHEN expires_at < NOW() THEN 'EXPIRED'
        ELSE 'PENDING'
    END as status,
    expires_at,
    accepted_at,
    created_at
FROM
    iam_invitations
WHERE
    organization_id = @organization_id
    AND %s
    AND %s
    AND %s
`

	query = fmt.Sprintf(query, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": orgID,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot query invitations: %w", err)
	}

	invitations, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Invitation])
	if err != nil {
		return fmt.Errorf("cannot collect invitations: %w", err)
	}

	*i = invitations
	return nil
}

func (i *Invitations) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	orgID gid.GID,
	filter *InvitationFilter,
) (int, error) {
	q := `
SELECT
    COUNT(*)
FROM
    iam_invitations
WHERE
    organization_id = @organization_id AND %s AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": orgID,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count invitations: %w", err)
	}

	return count, nil
}

// Tenant scope is not applied because this is used to count invitations across all tenants
// for a user who doesn't have tenant access yet (before accepting an invitation).
func (i *Invitations) CountByEmail(
	ctx context.Context,
	conn pg.Conn,
	email mail.Addr,
	filter *InvitationFilter,
) (int, error) {
	q := `
SELECT
    COUNT(*)
FROM
    iam_invitations
WHERE
    email = @email
    AND %s
`

	q = fmt.Sprintf(q, filter.SQLFragment())

	args := pgx.StrictNamedArgs{
		"email": email,
	}
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count invitations: %w", err)
	}

	return count, nil
}
