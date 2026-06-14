// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	"go.probo.inc/probo/pkg/iam/policy"
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

func (o *Organization) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id FROM organizations WHERE id = ANY(@resource_ids::text[])`

	args := pgx.StrictNamedArgs{
		"resource_ids": resourceIDs,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query authorization attributes: %w", err)
	}

	defer rows.Close()

	attrsByID := make(policy.AttributesByID)

	for rows.Next() {
		var id gid.GID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("cannot scan authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{
			"organization_id": id.String(),
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate authorization attributes: %w", err)
	}

	return attrsByID, nil
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
	conn pg.Querier,
	predicate Predicater,
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

	q = fmt.Sprintf(q, predicate.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, predicate.SQLArguments())

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

func (o *Organizations) LoadByIDs(
	ctx context.Context,
	conn pg.Querier,
	predicate Predicater,
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

	q = fmt.Sprintf(q, predicate.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_ids": organizationIDs}
	maps.Copy(args, predicate.SQLArguments())

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

func (o *Organizations) LoadByIdentityID(
	ctx context.Context,
	conn pg.Querier,
	predicate Predicater,
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

	q = fmt.Sprintf(q, predicate.SQLFragment(), cursor.SQLFragment())

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

func (o *Organizations) LoadAllByIdentityIDWithPendingInvitation(
	ctx context.Context,
	conn pg.Querier,
	predicate Predicater,
	identityID gid.GID,
) error {
	q := `
WITH invited_org AS (
	SELECT DISTINCT
		p.organization_id
	FROM
		iam_membership_profiles p
	INNER JOIN iam_invitations inv ON inv.user_id = p.id
	WHERE
		p.identity_id = @identity_id
		AND inv.accepted_at IS NULL
		AND inv.expires_at > NOW()
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
	invited_org ON organizations.id = invited_org.organization_id
WHERE
	%s
ORDER BY name ASC
`

	q = fmt.Sprintf(q, predicate.SQLFragment())

	args := pgx.StrictNamedArgs{"identity_id": identityID}
	maps.Copy(args, predicate.SQLArguments())

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
	conn pg.Tx,
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
	predicate Predicater,
	conn pg.Tx,
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

	q = fmt.Sprintf(q, predicate.SQLFragment())

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

	maps.Copy(args, predicate.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update organization: %w", err)
	}

	return nil
}

func (o *Organization) Delete(
	ctx context.Context,
	conn pg.Tx,
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
	conn pg.Querier,
	predicate Predicater,
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

	q = fmt.Sprintf(q, predicate.SQLFragment())

	args := pgx.StrictNamedArgs{"custom_domain_id": customDomainID}
	maps.Copy(args, predicate.SQLArguments())

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
