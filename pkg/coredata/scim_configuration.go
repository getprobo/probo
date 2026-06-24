// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
	"go.probo.inc/probo/pkg/iam/policy"
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

func (s *SCIMConfiguration) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM iam_scim_configurations WHERE id = ANY(@resource_ids::text[])`

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
		var id, organizationID gid.GID

		if err := rows.Scan(&id, &organizationID); err != nil {
			return nil, fmt.Errorf("cannot scan authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{
			"organization_id": organizationID.String(),
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate authorization attributes: %w", err)
	}

	return attrsByID, nil
}

func (s *SCIMConfiguration) LoadByID(
	ctx context.Context,
	conn pg.Querier,
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
	conn pg.Querier,
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
	conn pg.Querier,
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
	conn pg.Tx,
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
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
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
	conn pg.Tx,
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
	conn pg.Tx,
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
