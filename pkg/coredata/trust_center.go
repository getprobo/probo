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
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	TrustCenter struct {
		ID                           gid.GID      `db:"id"`
		OrganizationID               gid.GID      `db:"organization_id"`
		TenantID                     gid.TenantID `db:"tenant_id"`
		Active                       bool         `db:"active"`
		Slug                         string       `db:"slug"`
		LogoFileID                   *gid.GID     `db:"logo_file_id"`
		DarkLogoFileID               *gid.GID     `db:"dark_logo_file_id"`
		NonDisclosureAgreementFileID *gid.GID     `db:"non_disclosure_agreement_file_id"`
		CreatedAt                    time.Time    `db:"created_at"`
		UpdatedAt                    time.Time    `db:"updated_at"`
	}

	TrustCenters []*TrustCenter
)

func (tc *TrustCenter) CursorKey(orderBy TrustCenterOrderField) page.CursorKey {
	switch orderBy {
	case TrustCenterOrderFieldCreatedAt:
		return page.NewCursorKey(tc.ID, tc.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (tc *TrustCenter) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM trust_centers WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, tc.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query trust center authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (tc *TrustCenter) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	tenant_id,
	logo_file_id,
	dark_logo_file_id,
	active,
	slug,
	non_disclosure_agreement_file_id,
	created_at,
	updated_at
FROM
	trust_centers
WHERE
	%s
	AND id = @trust_center_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"trust_center_id": trustCenterID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust center: %w", err)
	}

	trustCenter, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TrustCenter])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect trust center: %w", err)
	}

	*tc = trustCenter

	return nil
}

func (tc *TrustCenter) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	tenant_id,
	logo_file_id,
	dark_logo_file_id,
	active,
	slug,
	non_disclosure_agreement_file_id,
	created_at,
	updated_at
FROM
	trust_centers
WHERE
	%s
	AND organization_id = @organization_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust center: %w", err)
	}

	trustCenter, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TrustCenter])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect trust center: %w", err)
	}

	*tc = trustCenter

	return nil
}

// Tenant id scope is not applied because we want to access trust centers by slug across all tenants for public access.
func (tc *TrustCenter) LoadBySlug(
	ctx context.Context,
	conn pg.Conn,
	slug string,
) error {
	q := `
SELECT
	id,
	organization_id,
	tenant_id,
	logo_file_id,
	dark_logo_file_id,
	active,
	slug,
	non_disclosure_agreement_file_id,
	created_at,
	updated_at
FROM
	trust_centers
WHERE
	slug = @slug
LIMIT 1;
`

	args := pgx.StrictNamedArgs{"slug": slug}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust center: %w", err)
	}

	trustCenter, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TrustCenter])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect trust center: %w", err)
	}

	*tc = trustCenter

	return nil
}

func (tc *TrustCenter) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO trust_centers (
	id,
	organization_id,
	tenant_id,
	logo_file_id,
	dark_logo_file_id,
	active,
	slug,
	non_disclosure_agreement_file_id,
	created_at,
	updated_at
) VALUES (
	@id,
	@organization_id,
	@tenant_id,
	@logo_file_id,
	@dark_logo_file_id,
	@active,
	@slug,
	@non_disclosure_agreement_file_id,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                               tc.ID,
		"organization_id":                  tc.OrganizationID,
		"tenant_id":                        tc.TenantID,
		"logo_file_id":                     tc.LogoFileID,
		"dark_logo_file_id":                tc.DarkLogoFileID,
		"active":                           tc.Active,
		"slug":                             tc.Slug,
		"non_disclosure_agreement_file_id": tc.NonDisclosureAgreementFileID,
		"created_at":                       tc.CreatedAt,
		"updated_at":                       tc.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "trust_centers_slug_key" {
				return ErrResourceAlreadyExists
			}
		}
		return fmt.Errorf("cannot insert trust center: %w", err)
	}

	return nil
}

func (tc *TrustCenter) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE trust_centers
SET
	active = @active,
	slug = @slug,
	logo_file_id = @logo_file_id,
	dark_logo_file_id = @dark_logo_file_id,
	non_disclosure_agreement_file_id = @non_disclosure_agreement_file_id,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                               tc.ID,
		"logo_file_id":                     tc.LogoFileID,
		"dark_logo_file_id":                tc.DarkLogoFileID,
		"active":                           tc.Active,
		"slug":                             tc.Slug,
		"non_disclosure_agreement_file_id": tc.NonDisclosureAgreementFileID,
		"updated_at":                       tc.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update trust center: %w", err)
	}

	return nil
}
