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

	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	People struct {
		ID                       gid.GID    `db:"id"`
		OrganizationID           gid.GID    `db:"organization_id"`
		Kind                     PeopleKind `db:"kind"`
		UserID                   *gid.GID   `db:"user_id"`
		FullName                 string     `db:"full_name"`
		PrimaryEmailAddress      string     `db:"primary_email_address"`
		AdditionalEmailAddresses []string   `db:"additional_email_addresses"`
		CreatedAt                time.Time  `db:"created_at"`
		UpdatedAt                time.Time  `db:"updated_at"`
	}

	Peoples []*People
)

func (p People) CursorKey(orderBy PeopleOrderField) page.CursorKey {
	switch orderBy {
	case PeopleOrderFieldCreatedAt:
		return page.NewCursorKey(p.ID, p.CreatedAt)
	case PeopleOrderFieldFullName:
		return page.NewCursorKey(p.ID, p.FullName)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (p *People) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	peopleID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    kind,
    user_id,
    full_name,
    primary_email_address,
    additional_email_addresses,
    created_at,
    updated_at
FROM
    peoples
WHERE
    %s
    AND id = @people_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"people_id": peopleID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query people: %w", err)
	}

	people, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[People])
	if err != nil {
		return fmt.Errorf("cannot collect people: %w", err)
	}

	*p = people

	return nil
}

func (p *People) LoadByUserID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	userID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    kind,
    user_id,
    full_name,
    primary_email_address,
    additional_email_addresses,
    created_at,
    updated_at
FROM
    peoples
WHERE
    %s
    AND user_id = @user_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"user_id": userID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query people: %w", err)
	}

	people, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[People])
	if err != nil {
		return fmt.Errorf("cannot collect people: %w", err)
	}

	*p = people

	return nil
}

func (p People) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    peoples (
        tenant_id,
        id,
        organization_id,
        user_id,
        kind,
        full_name,
        primary_email_address,
        additional_email_addresses,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @people_id,
    @organization_id,
    @user_id,
    @kind,
    @full_name,
    @primary_email_address,
    @additional_email_addresses,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"tenant_id":                  scope.GetTenantID(),
		"people_id":                  p.ID,
		"organization_id":            p.OrganizationID,
		"user_id":                    p.UserID,
		"kind":                       p.Kind,
		"full_name":                  p.FullName,
		"primary_email_address":      p.PrimaryEmailAddress,
		"additional_email_addresses": p.AdditionalEmailAddresses,
		"created_at":                 p.CreatedAt,
		"updated_at":                 p.UpdatedAt,
	}
	_, err := conn.Exec(ctx, q, args)
	return err
}

func (p People) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM peoples WHERE %s AND id = @people_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"people_id": p.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (p *Peoples) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[PeopleOrderField],
) error {
	q := `
SELECT
    id,
    organization_id,
    kind,
    user_id,
    full_name,
    primary_email_address,
    additional_email_addresses,
    created_at,
    updated_at
FROM
    peoples
WHERE
    %s
    AND organization_id = @organization_id
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, cursor.SQLArguments())
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query people: %w", err)
	}

	peoples, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[People])
	if err != nil {
		return fmt.Errorf("cannot collect people: %w", err)
	}

	*p = peoples

	return nil
}

func (p *People) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE peoples SET
	user_id = @user_id,
	full_name = @full_name,
	primary_email_address = @primary_email_address,
	additional_email_addresses = @additional_email_addresses,
	kind = @kind,
	updated_at = @updated_at
WHERE %s
    AND id = @people_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"people_id":                  p.ID,
		"user_id":                    p.UserID,
		"full_name":                  p.FullName,
		"primary_email_address":      p.PrimaryEmailAddress,
		"additional_email_addresses": p.AdditionalEmailAddresses,
		"kind":                       p.Kind,
		"updated_at":                 p.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update people: %w", err)
	}

	return nil
}

func (p *Peoples) LoadAwaitingSigning(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
WITH signatories AS (
	SELECT
		signed_by
	FROM
		policy_version_signatures
	WHERE
	    %s
		AND state = 'REQUESTED'
	GROUP BY
		signed_by
)
SELECT
	id,
	organization_id,
	kind,
	user_id,
	full_name,
	primary_email_address,
	additional_email_addresses,
	created_at,
	updated_at
FROM
	peoples
INNER JOIN signatories ON peoples.id = signatories.signed_by
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	rows, err := conn.Query(ctx, q, scope.SQLArguments())
	if err != nil {
		return fmt.Errorf("cannot query people: %w", err)
	}

	peoples, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[People])
	if err != nil {
		return fmt.Errorf("cannot collect people: %w", err)
	}

	*p = peoples

	return nil
}
