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

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
)

type (
	Identity struct {
		ID                   gid.GID   `db:"id"`
		EmailAddress         mail.Addr `db:"email_address"`
		FullName             string    `db:"full_name"`
		HashedPassword       []byte    `db:"hashed_password"`
		EmailAddressVerified bool      `db:"email_address_verified"`
		SAMLSubject          *string   `db:"saml_subject"`
		CreatedAt            time.Time `db:"created_at"`
		UpdatedAt            time.Time `db:"updated_at"`
	}

	Identities []*Identity
)

func (i Identity) CursorKey(orderBy IdentityOrderField) page.CursorKey {
	switch orderBy {
	case IdentityOrderFieldCreatedAt:
		return page.NewCursorKey(i.ID, i.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (i *Identities) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	organizationID gid.GID,
	cursor *page.Cursor[IdentityOrderField],
) error {
	q := `
SELECT
    id,
    email_address,
    full_name,
    hashed_password,
    email_address_verified,
    saml_subject,
    created_at,
    updated_at
FROM
	identities
WHERE
	id IN (
		SELECT identity_id FROM iam_memberships WHERE organization_id = @organization_id
	)
	AND %s
`

	q = fmt.Sprintf(q, cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query identities: %w", err)
	}

	identities, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Identity])
	if err != nil {
		return fmt.Errorf("cannot collect identities: %w", err)
	}

	*i = identities

	return nil
}

func (i *Identities) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(*)
FROM
	identities
WHERE
	id IN (
		SELECT identity_id FROM iam_memberships WHERE organization_id = @organization_id AND %s
	)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count identities: %w", err)
	}

	return count, nil
}

// Tenant id scope is not applied because we want to access identities across all tenants for authentication purposes.
func (i *Identity) LoadByEmail(
	ctx context.Context,
	conn pg.Conn,
	email mail.Addr,
) error {
	q := `
SELECT
    id,
    email_address,
    full_name,
    hashed_password,
    email_address_verified,
    saml_subject,
    created_at,
    updated_at
FROM
    identities
WHERE
    email_address = @identity_email
LIMIT 1;
`

	args := pgx.StrictNamedArgs{"identity_email": email}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query identity: %w", err)
	}

	identity, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Identity])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect identity: %w", err)
	}

	*i = identity

	return nil
}

// Tenant id scope is not applied because we want to access identities across all tenants for authentication purposes.
func (i *Identity) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	identityID gid.GID,
) error {
	q := `
SELECT
    id,
    email_address,
    full_name,
    hashed_password,
    email_address_verified,
    saml_subject,
    created_at,
    updated_at
FROM
    identities
WHERE
    id = @identity_id
LIMIT 1;
`

	args := pgx.StrictNamedArgs{"identity_id": identityID}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query identity: %w", err)
	}

	identity, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Identity])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect identity: %w", err)
	}

	*i = identity

	return nil
}

// AuthorizationAttributes loads the minimal authorization attributes for policy condition evaluation.
// It is intentionally lightweight and does not populate the Identity struct.
func (i *Identity) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `
SELECT
    ident.id,
    ident.email_address,
    COALESCE(string_agg(inv.organization_id::text, ','), '') AS pending_invitation_organization_ids
FROM
    identities ident
LEFT JOIN
    iam_invitations inv ON inv.email = ident.email_address
        AND inv.accepted_at IS NULL
        AND inv.expires_at > NOW()
WHERE
    ident.id = $1
GROUP BY
    ident.id, ident.email_address
`

	var (
		id                      gid.GID
		emailAddress            string
		pendingInvitationOrgIDs string
	)
	if err := conn.QueryRow(ctx, q, i.ID).Scan(&id, &emailAddress, &pendingInvitationOrgIDs); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query identity iam attributes: %w", err)
	}

	return map[string]string{
		"identity_id":                         id.String(),
		"email":                               emailAddress,
		"pending_invitation_organization_ids": pendingInvitationOrgIDs,
	}, nil
}

func (i *Identity) Insert(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
INSERT INTO
    identities (id, email_address, full_name, hashed_password, email_address_verified, saml_subject, created_at, updated_at)
VALUES (
    @identity_id,
    @email_address,
    @full_name,
    @hashed_password,
    @email_address_verified,
    @saml_subject,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"identity_id":            i.ID,
		"email_address":          i.EmailAddress,
		"full_name":              i.FullName,
		"hashed_password":        i.HashedPassword,
		"saml_subject":           i.SAMLSubject,
		"created_at":             i.CreatedAt,
		"updated_at":             i.UpdatedAt,
		"email_address_verified": i.EmailAddressVerified,
	}

	_, err := conn.Exec(ctx, q, args)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && strings.Contains(pgErr.ConstraintName, "email_address") {
				return ErrResourceAlreadyExists
			}
		}

		return err
	}

	return nil
}

func (i *Identity) Update(ctx context.Context, conn pg.Conn) error {
	q := `
UPDATE
    identities
SET
    email_address = @email_address,
    full_name = @full_name,
    email_address_verified = @email_address_verified,
    saml_subject = @saml_subject,
    hashed_password = @hashed_password,
    updated_at = @updated_at
WHERE
    id = @identity_id
`

	args := pgx.StrictNamedArgs{
		"identity_id":            i.ID,
		"email_address":          i.EmailAddress,
		"full_name":              i.FullName,
		"email_address_verified": i.EmailAddressVerified,
		"saml_subject":           i.SAMLSubject,
		"updated_at":             i.UpdatedAt,
		"hashed_password":        i.HashedPassword,
	}

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update identity: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

// LoadBySAMLSubject loads an identity by their SAML subject (NameID)
func (i *Identity) LoadBySAMLSubject(
	ctx context.Context,
	conn pg.Conn,
	samlSubject string,
) error {
	q := `
SELECT
    id,
    email_address,
    full_name,
    hashed_password,
    email_address_verified,
    saml_subject,
    created_at,
    updated_at
FROM
    identities
WHERE
    saml_subject = @saml_subject
LIMIT 1;
`

	args := pgx.StrictNamedArgs{"saml_subject": samlSubject}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query identity by SAML subject: %w", err)
	}

	identity, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Identity])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect identity: %w", err)
	}

	*i = identity

	return nil
}

func (i *Identity) CountMemberships(
	ctx context.Context,
	conn pg.Conn,
) (int, error) {
	q := `
SELECT
    COUNT(*)
FROM
    iam_memberships
WHERE
    identity_id = @identity_id
`

	args := pgx.StrictNamedArgs{"identity_id": i.ID}

	var count int
	err := conn.QueryRow(ctx, q, args).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count identity memberships: %w", err)
	}

	return count, nil
}
