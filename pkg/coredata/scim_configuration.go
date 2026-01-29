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
	SCIMConfiguration struct {
		ID             gid.GID   `db:"id"`
		OrganizationID gid.GID   `db:"organization_id"`
		BridgeID       *gid.GID  `db:"bridge_id"`
		HashedToken    []byte    `db:"hashed_token"`
		CreatedAt      time.Time `db:"created_at"`
		UpdatedAt      time.Time `db:"updated_at"`
	}

	SCIMConfigurations []*SCIMConfiguration
)

func (s *SCIMConfiguration) CursorKey(orderBy SCIMConfigurationOrderField) page.CursorKey {
	switch orderBy {
	case SCIMConfigurationOrderFieldCreatedAt:
		return page.NewCursorKey(s.ID, s.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (s *SCIMConfiguration) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM iam_scim_configurations WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, s.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query scim configuration authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (s *SCIMConfiguration) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	configID gid.GID,
) error {
	q := `
WITH scim_config AS (
    SELECT
        id,
        organization_id,
        hashed_token,
        created_at,
        updated_at
    FROM
        iam_scim_configurations
    WHERE
        %s
        AND id = @id
    LIMIT 1
)
SELECT
    sc.id,
    sc.organization_id,
    b.id AS bridge_id,
    sc.hashed_token,
    sc.created_at,
    sc.updated_at
FROM
    scim_config sc
LEFT JOIN
    iam_scim_bridges b ON b.scim_configuration_id = sc.id;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": configID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query iam_scim_configurations: %w", err)
	}

	config, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SCIMConfiguration])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect scim_configuration: %w", err)
	}

	*s = config

	return nil
}

func (s *SCIMConfiguration) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) error {
	q := `
WITH scim_config AS (
    SELECT
        id,
        organization_id,
        hashed_token,
        created_at,
        updated_at
    FROM
        iam_scim_configurations
    WHERE
        %s
        AND organization_id = @organization_id
    LIMIT 1
)
SELECT
    sc.id,
    sc.organization_id,
    b.id AS bridge_id,
    sc.hashed_token,
    sc.created_at,
    sc.updated_at
FROM
    scim_config sc
LEFT JOIN
    iam_scim_bridges b ON b.scim_configuration_id = sc.id;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query iam_scim_configurations: %w", err)
	}

	config, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SCIMConfiguration])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect scim_configuration: %w", err)
	}

	*s = config

	return nil
}

func (s *SCIMConfiguration) LoadByHashedToken(
	ctx context.Context,
	conn pg.Conn,
	hashedToken []byte,
) error {
	q := `
WITH scim_config AS (
    SELECT
        id,
        organization_id,
        hashed_token,
        created_at,
        updated_at
    FROM
        iam_scim_configurations
    WHERE
        hashed_token = @hashed_token
    LIMIT 1
)
SELECT
    sc.id,
    sc.organization_id,
    b.id AS bridge_id,
    sc.hashed_token,
    sc.created_at,
    sc.updated_at
FROM
    scim_config sc
LEFT JOIN
    iam_scim_bridges b ON b.scim_configuration_id = sc.id;
`

	args := pgx.StrictNamedArgs{"hashed_token": hashedToken}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query iam_scim_configurations: %w", err)
	}

	config, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SCIMConfiguration])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect scim_configuration: %w", err)
	}

	*s = config

	return nil
}

func (s *SCIMConfiguration) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO iam_scim_configurations (
    id,
    tenant_id,
    organization_id,
    hashed_token,
    created_at,
    updated_at
) VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @hashed_token,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":              s.ID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": s.OrganizationID,
		"hashed_token":    s.HashedToken,
		"created_at":      s.CreatedAt,
		"updated_at":      s.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "iam_scim_configurations_organization_unique" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert scim_configuration: %w", err)
	}

	return nil
}

func (s *SCIMConfiguration) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE iam_scim_configurations
SET
    hashed_token = @hashed_token,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":           s.ID,
		"hashed_token": s.HashedToken,
		"updated_at":   s.UpdatedAt,
	}

	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update scim_configuration: %w", err)
	}

	return nil
}

func (s *SCIMConfiguration) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM iam_scim_configurations
WHERE
    %s
    AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": s.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete scim_configuration: %w", err)
	}

	return nil
}
