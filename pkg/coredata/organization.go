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
	"go.probo.inc/probo/pkg/page"
)

type (
	Organization struct {
		ID                   gid.GID      `db:"id"`
		TenantID             gid.TenantID `db:"tenant_id"`
		Name                 string       `db:"name"`
		LogoFileID           *gid.GID     `db:"logo_file_id"`
		HorizontalLogoFileID *gid.GID     `db:"horizontal_logo_file_id"`
		Description          *string      `db:"description"`
		WebsiteURL           *string      `db:"website_url"`
		Email                *string      `db:"email"`
		HeadquarterAddress   *string      `db:"headquarter_address"`
		CustomDomainID       *gid.GID     `db:"custom_domain_id"`
		CreatedAt            time.Time    `db:"created_at"`
		UpdatedAt            time.Time    `db:"updated_at"`
	}

	Organizations []*Organization
)

func (o *Organization) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT id FROM organizations WHERE id = $1 LIMIT 1;`

	var id gid.GID
	if err := conn.QueryRow(ctx, q, o.ID).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query organization authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": o.ID.String()}, nil
}

func (o Organization) CursorKey(orderBy OrganizationOrderField) page.CursorKey {
	switch orderBy {
	case OrganizationOrderFieldName:
		return page.NewCursorKey(o.ID, o.Name)
	case OrganizationOrderFieldCreatedAt:
		return page.NewCursorKey(o.ID, o.CreatedAt)
	case OrganizationOrderFieldUpdatedAt:
		return page.NewCursorKey(o.ID, o.UpdatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (o *Organization) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) error {
	q := `
SELECT
    tenant_id,
    id,
    name,
    logo_file_id,
    horizontal_logo_file_id,
    description,
    website_url,
    email,
    headquarter_address,
    custom_domain_id,
    created_at,
    updated_at
FROM
    organizations
WHERE
    %s
    AND id = @organization_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query organizations: %w", err)
	}

	organization, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Organization])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect organization: %w", err)
	}

	*o = organization

	return nil
}

func (o *Organizations) LoadByIdentityID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	identityID gid.GID,
	cursor *page.Cursor[OrganizationOrderField],
) error {
	q := `
WITH identity_org AS (
	SELECT
		organization_id
	FROM
		iam_memberships
	WHERE
		identity_id = @identity_id
)
SELECT
	tenant_id,
    id,
    name,
    logo_file_id,
    horizontal_logo_file_id,
    description,
    website_url,
    email,
    headquarter_address,
    custom_domain_id,
    created_at,
    updated_at
FROM
	organizations
INNER JOIN
	identity_org ON organizations.id = identity_org.organization_id
WHERE
	%s
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"identity_id": identityID}
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query organizations: %w", err)
	}

	organizations, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Organization])
	if err != nil {
		return fmt.Errorf("cannot collect organizations: %w", err)
	}

	*o = organizations

	return nil
}

func (o *Organizations) LoadAllByIdentityID(
	ctx context.Context,
	conn pg.Conn,
	identityID gid.GID,
) error {
	q := `
WITH identity_org AS (
	SELECT
		organization_id
	FROM
		iam_memberships
	WHERE
		identity_id = @identity_id
)
SELECT
	tenant_id,
    id,
    name,
    description,
    website_url,
    email,
    headquarter_address,
    custom_domain_id,
    logo_file_id,
    horizontal_logo_file_id,
    created_at,
    updated_at
FROM
	organizations
INNER JOIN
	identity_org ON organizations.id = identity_org.organization_id
ORDER BY
	name ASC
`

	args := pgx.StrictNamedArgs{"identity_id": identityID}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query organizations: %w", err)
	}

	organizations, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Organization])
	if err != nil {
		return fmt.Errorf("cannot collect organizations: %w", err)
	}

	*o = organizations

	return nil
}

func (o *Organizations) LoadAllByIdentityIDWithRole(
	ctx context.Context,
	conn pg.Conn,
	identityID gid.GID,
	role MembershipRole,
) error {
	q := `
WITH identity_org AS (
	SELECT
		organization_id
	FROM
		iam_memberships
	WHERE
		identity_id = @identity_id
		AND role = @role
)
SELECT
	tenant_id,
    id,
    name,
    description,
    website_url,
    email,
    headquarter_address,
    custom_domain_id,
    logo_file_id,
    horizontal_logo_file_id,
    created_at,
    updated_at
FROM
	organizations
INNER JOIN
	identity_org ON organizations.id = identity_org.organization_id
ORDER BY
	name ASC
`

	args := pgx.StrictNamedArgs{
		"identity_id": identityID,
		"role":        role,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query organizations: %w", err)
	}

	organizations, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Organization])
	if err != nil {
		return fmt.Errorf("cannot collect organizations: %w", err)
	}

	*o = organizations

	return nil
}

func (o *Organizations) LoadAllByPersonalAPIKeyID(
	ctx context.Context,
	conn pg.Conn,
	personalAPIKeyID gid.GID,
) error {
	q := `
WITH personal_api_key_org AS (
	SELECT
		am.organization_id
	FROM
		iam_personal_api_key_memberships akm
	INNER JOIN
		iam_memberships am ON akm.membership_id = am.id
	WHERE
		akm.personal_api_key_id = @personal_api_key_id
)
SELECT
	tenant_id,
    id,
    name,
    description,
    website_url,
    email,
    headquarter_address,
    custom_domain_id,
    logo_file_id,
    horizontal_logo_file_id,
    created_at,
    updated_at
FROM
	organizations
INNER JOIN
	personal_api_key_org ON organizations.id = personal_api_key_org.organization_id
ORDER BY
	name ASC
`

	args := pgx.StrictNamedArgs{"personal_api_key_id": personalAPIKeyID}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query organizations: %w", err)
	}

	organizations, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Organization])
	if err != nil {
		return fmt.Errorf("cannot collect organizations: %w", err)
	}

	*o = organizations

	return nil
}

func (o *Organization) Insert(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
INSERT INTO organizations (
    tenant_id,
    id,
    name,
    logo_file_id,
    horizontal_logo_file_id,
    description,
    website_url,
    email,
    headquarter_address,
    custom_domain_id,
    created_at,
    updated_at
) VALUES (@tenant_id, @id, @name, @logo_file_id, @horizontal_logo_file_id, @description, @website_url, @email, @headquarter_address, @custom_domain_id, @created_at, @updated_at)
`

	args := pgx.StrictNamedArgs{
		"tenant_id":               o.TenantID,
		"id":                      o.ID,
		"name":                    o.Name,
		"logo_file_id":            o.LogoFileID,
		"horizontal_logo_file_id": o.HorizontalLogoFileID,
		"description":             o.Description,
		"website_url":             o.WebsiteURL,
		"email":                   o.Email,
		"headquarter_address":     o.HeadquarterAddress,
		"custom_domain_id":        o.CustomDomainID,
		"created_at":              o.CreatedAt,
		"updated_at":              o.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return err
	}

	return nil
}

func (o *Organization) Update(
	ctx context.Context,
	scope Scoper,
	conn pg.Conn,
) error {
	q := `
UPDATE organizations
SET
    name = @name,
    logo_file_id = @logo_file_id,
    horizontal_logo_file_id = @horizontal_logo_file_id,
    description = @description,
    website_url = @website_url,
    email = @email,
    headquarter_address = @headquarter_address,
    custom_domain_id = @custom_domain_id,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                      o.ID,
		"name":                    o.Name,
		"logo_file_id":            o.LogoFileID,
		"horizontal_logo_file_id": o.HorizontalLogoFileID,
		"description":             o.Description,
		"website_url":             o.WebsiteURL,
		"email":                   o.Email,
		"headquarter_address":     o.HeadquarterAddress,
		"custom_domain_id":        o.CustomDomainID,
		"updated_at":              o.UpdatedAt,
	}

	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update organization: %w", err)
	}

	return nil
}

func (o *Organization) Delete(
	ctx context.Context,
	conn pg.Conn,
	organizationID gid.GID,
) error {
	q := `
DELETE FROM organizations
WHERE id = @id
`

	args := pgx.StrictNamedArgs{"id": o.ID}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete organization: %w", err)
	}

	return nil
}

func (o *Organization) LoadByCustomDomainID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	customDomainID gid.GID,
) error {
	q := `
SELECT
    tenant_id,
    id,
    name,
    logo_file_id,
    horizontal_logo_file_id,
    description,
    website_url,
    email,
    headquarter_address,
    custom_domain_id,
    created_at,
    updated_at
FROM
    organizations
WHERE
    %s
    AND custom_domain_id = @custom_domain_id
LIMIT 1
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"custom_domain_id": customDomainID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query organization by custom domain: %w", err)
	}

	organization, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Organization])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect organization: %w", err)
	}

	*o = organization

	return nil
}

func (o *Organizations) BatchLoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationIDs []gid.GID,
) error {
	q := `
SELECT
    tenant_id,
    id,
    name,
    logo_file_id,
    horizontal_logo_file_id,
    description,
    website_url,
    email,
    headquarter_address,
    custom_domain_id,
    created_at,
    updated_at
FROM
    organizations
WHERE
	%s
    AND id = ANY(@organization_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_ids": organizationIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query organizations: %w", err)
	}

	organizations, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Organization])
	if err != nil {
		return fmt.Errorf("cannot collect organizations: %w", err)
	}

	*o = organizations

	return nil
}
