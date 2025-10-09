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
	"strings"
	"time"

	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
)

type (
	User struct {
		ID                   gid.GID   `db:"id"`
		EmailAddress         string    `db:"email_address"`
		HashedPassword       []byte    `db:"hashed_password"`
		FullName             string    `db:"fullname"`
		EmailAddressVerified bool      `db:"email_address_verified"`
		CreatedAt            time.Time `db:"created_at"`
		UpdatedAt            time.Time `db:"updated_at"`
	}

	Users []*User

	ErrUserNotFound struct {
		Identifier string
	}

	ErrUserAlreadyExists struct {
		message string
	}
)

func (e ErrUserNotFound) Error() string {
	return fmt.Sprintf("user not found: %q", e.Identifier)
}

func (e ErrUserAlreadyExists) Error() string {
	return e.message
}

func (u User) CursorKey(orderBy UserOrderField) page.CursorKey {
	switch orderBy {
	case UserOrderFieldCreatedAt:
		return page.NewCursorKey(u.ID, u.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (u *Users) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	organizationID gid.GID,
	cursor *page.Cursor[UserOrderField],
) error {
	q := `
SELECT
    id,
    email_address,
    hashed_password,
    email_address_verified,
    fullname,
	created_at,
	updated_at
FROM
	users
WHERE
	id IN (
		SELECT user_id FROM authz_memberships WHERE organization_id = @organization_id
	)
	AND %s
`

	q = fmt.Sprintf(q, cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query users: %w", err)
	}

	users, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[User])
	if err != nil {
		return fmt.Errorf("cannot collect users: %w", err)
	}

	*u = users

	return nil
}

func (u *Users) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(*)
FROM
	users
WHERE
	id IN (
		SELECT user_id FROM authz_memberships WHERE organization_id = @organization_id AND %s
	)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count users: %w", err)
	}

	return count, nil
}

// Tenant id scope is not applied because we want to access users across all tenants for authentication purposes.
func (u *User) LoadByEmail(
	ctx context.Context,
	conn pg.Conn,
	email string,
) error {
	q := `
SELECT
    id,
    email_address,
    hashed_password,
    email_address_verified,
    fullname,
    created_at,
    updated_at
FROM
    users
WHERE
    email_address = @user_email
LIMIT 1;
`

	args := pgx.StrictNamedArgs{"user_email": email}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query user: %w", err)
	}

	user, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &ErrUserNotFound{Identifier: email}
		}

		return fmt.Errorf("cannot collect user: %w", err)
	}

	*u = user

	return nil
}

// Tenant id scope is not applied because we want to access users across all tenants for authentication purposes.
func (u *User) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	userID gid.GID,
) error {
	q := `
SELECT
    id,
    email_address,
    hashed_password,
	email_address_verified,
    fullname,
    created_at,
    updated_at
FROM
    users
WHERE
    id = @user_id
LIMIT 1;
`

	args := pgx.StrictNamedArgs{"user_id": userID}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query user: %w", err)
	}

	user, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &ErrUserNotFound{Identifier: userID.String()}
		}

		return fmt.Errorf("cannot collect user: %w", err)
	}

	*u = user

	return nil
}

func (u *User) Insert(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
INSERT INTO
    users (id, email_address, hashed_password, email_address_verified, fullname, created_at, updated_at)
VALUES (
    @user_id,
    @email_address,
    @hashed_password,
    @email_address_verified,
    @fullname,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"user_id":                u.ID,
		"email_address":          u.EmailAddress,
		"hashed_password":        u.HashedPassword,
		"fullname":               u.FullName,
		"created_at":             u.CreatedAt,
		"updated_at":             u.UpdatedAt,
		"email_address_verified": u.EmailAddressVerified,
	}

	_, err := conn.Exec(ctx, q, args)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && strings.Contains(pgErr.ConstraintName, "email_address") {
				return &ErrUserAlreadyExists{
					message: fmt.Sprintf("user with email %s already exists", u.EmailAddress),
				}
			}
		}

		return err
	}

	return nil
}

func (u *User) UpdateEmailVerification(
	ctx context.Context,
	conn pg.Conn,
	verified bool,
) error {
	q := `
UPDATE
    users
SET
    email_address_verified = @email_address_verified,
    updated_at = @updated_at
WHERE
    id = @user_id
`

	args := pgx.StrictNamedArgs{
		"user_id":                u.ID,
		"email_address_verified": verified,
		"updated_at":             time.Now(),
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update user email verification: %w", err)
	}

	u.EmailAddressVerified = verified
	u.UpdatedAt = args["updated_at"].(time.Time)

	return nil
}

func (u *User) UpdatePassword(
	ctx context.Context,
	conn pg.Conn,
	hashedPassword []byte,
) error {
	q := `
UPDATE
    users
SET
    hashed_password = @hashed_password,
    updated_at = @updated_at
WHERE
    id = @user_id
`

	now := time.Now()
	args := pgx.StrictNamedArgs{
		"user_id":         u.ID,
		"hashed_password": hashedPassword,
		"updated_at":      now,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update user password: %w", err)
	}

	u.HashedPassword = hashedPassword
	u.UpdatedAt = now

	return nil
}
