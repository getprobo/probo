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
	"time"

	"github.com/getprobo/probo/pkg/gid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
)

type (
	UserOrganization struct {
		UserID         gid.GID   `db:"user_id"`
		OrganizationID gid.GID   `db:"organization_id"`
		CreatedAt      time.Time `db:"created_at"`
	}

	UserOrganizations []*UserOrganization
)

func (uo UserOrganization) Insert(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
INSERT INTO users_organizations (user_id, organization_id, created_at)
VALUES (@user_id, @organization_id, @created_at)
`

	_, err := conn.Exec(ctx, q, pgx.StrictNamedArgs{"user_id": uo.UserID, "organization_id": uo.OrganizationID, "created_at": uo.CreatedAt})

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "usrmgr_user_organizations_pkey" {
				return &ErrResourceAlreadyExists{
					Resource: "user organization membership",
					Message:  "user is already a member of this organization",
				}
			}
		}

		return err
	}

	return nil
}

func (uo UserOrganization) Delete(ctx context.Context, conn pg.Conn) error {
	q := `
DELETE FROM users_organizations WHERE user_id = @user_id AND organization_id = @organization_id
`

	_, err := conn.Exec(ctx, q, pgx.StrictNamedArgs{"user_id": uo.UserID, "organization_id": uo.OrganizationID})
	return err
}

func (uo *UserOrganizations) ForUserID(
	ctx context.Context,
	conn pg.Conn,
	userID gid.GID,
) error {
	q := `
SELECT user_id, organization_id, created_at FROM users_organizations WHERE user_id = @user_id
`

	rows, err := conn.Query(ctx, q, pgx.StrictNamedArgs{"user_id": userID})
	if err != nil {
		return err
	}

	userOrganizations, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[UserOrganization])
	if err != nil {
		return err
	}

	*uo = userOrganizations

	return nil
}
