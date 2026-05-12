// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
	ComplianceFramework struct {
		ID             gid.GID   `db:"id"`
		OrganizationID gid.GID   `db:"organization_id"`
		TrustCenterID  gid.GID   `db:"trust_center_id"`
		FrameworkID    gid.GID   `db:"framework_id"`
		Rank           int       `db:"rank"`
		CreatedAt      time.Time `db:"created_at"`
		UpdatedAt      time.Time `db:"updated_at"`

		// Visibility is a non-db field used to return all frameworks for a trust center, including hidden ones.
		Visibility ComplianceFrameworkVisibility `db:"visibility"`
	}

	ComplianceFrameworks []*ComplianceFramework
)

func (c ComplianceFramework) CursorKey(orderBy ComplianceFrameworkOrderField) page.CursorKey {
	switch orderBy {
	case ComplianceFrameworkOrderFieldCreatedAt:
		return page.NewCursorKey(c.ID, c.CreatedAt)
	case ComplianceFrameworkOrderFieldRank:
		return page.NewCursorKey(c.ID, c.Rank)
	}
	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (c *ComplianceFramework) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	q := `SELECT organization_id FROM compliance_frameworks WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, c.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query compliance framework authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (c *ComplianceFramework) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	complianceFrameworkID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    trust_center_id,
    framework_id,
    rank,
    'PUBLIC' AS visibility,
    created_at,
    updated_at
FROM
    compliance_frameworks
WHERE
    %s
    AND id = @compliance_framework_id
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"compliance_framework_id": complianceFrameworkID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query compliance_frameworks: %w", err)
	}

	cf, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ComplianceFramework])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect compliance framework: %w", err)
	}

	*c = cf

	return nil
}

func (c *ComplianceFramework) LoadByTrustCenterIDAndFrameworkID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	trustCenterID gid.GID,
	frameworkID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    trust_center_id,
    framework_id,
    rank,
    'PUBLIC' AS visibility,
    created_at,
    updated_at
FROM
    compliance_frameworks
WHERE
    %s
    AND trust_center_id = @trust_center_id
    AND framework_id = @framework_id
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_id": trustCenterID,
		"framework_id":    frameworkID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query compliance_frameworks: %w", err)
	}

	cf, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ComplianceFramework])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect compliance framework: %w", err)
	}

	*c = cf

	return nil
}

func (c *ComplianceFramework) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    compliance_frameworks (
        id,
        tenant_id,
        organization_id,
        trust_center_id,
        framework_id,
        rank,
        created_at,
        updated_at
    )
VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @trust_center_id,
    @framework_id,
    (SELECT COALESCE(MAX(rank), 0) + 1 FROM compliance_frameworks WHERE trust_center_id = @trust_center_id),
    @created_at,
    @updated_at
)
RETURNING rank;
`

	args := pgx.StrictNamedArgs{
		"id":              c.ID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": c.OrganizationID,
		"trust_center_id": c.TrustCenterID,
		"framework_id":    c.FrameworkID,
		"created_at":      c.CreatedAt,
		"updated_at":      c.UpdatedAt,
	}

	err := conn.QueryRow(ctx, q, args).Scan(&c.Rank)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return ErrResourceAlreadyExists
			}
		}
		return fmt.Errorf("cannot insert compliance framework: %w", err)
	}

	return nil
}

func (c *ComplianceFramework) UpdateRank(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
WITH old AS (
  SELECT
    rank AS old_rank
  FROM compliance_frameworks
  WHERE %s AND id = @id AND trust_center_id = @trust_center_id
)

UPDATE compliance_frameworks
SET
    rank = CASE
        WHEN id = @id THEN @new_rank
        ELSE rank + CASE
            WHEN @new_rank < old.old_rank THEN 1
            WHEN @new_rank > old.old_rank THEN -1
        END
    END,
    updated_at = @updated_at
FROM old
WHERE %s
  AND (
    id = @id
    OR (rank BETWEEN LEAST(old.old_rank, @new_rank) AND GREATEST(old.old_rank, @new_rank))
  );
`

	scopeFragment := scope.SQLFragment()
	q = fmt.Sprintf(q, scopeFragment, scopeFragment)

	args := pgx.StrictNamedArgs{
		"id":              c.ID,
		"new_rank":        c.Rank,
		"trust_center_id": c.TrustCenterID,
		"updated_at":      c.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update compliance framework rank: %w", err)
	}

	return nil
}

func (c *ComplianceFramework) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM
    compliance_frameworks
WHERE
    %s
    AND id = @id;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": c.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete compliance framework: %w", err)
	}

	return nil
}

func (c *ComplianceFrameworks) LoadByTrustCenterID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	trustCenterID gid.GID,
	cursor *page.Cursor[ComplianceFrameworkOrderField],
) error {
	q := `
SELECT
    id,
    organization_id,
    trust_center_id,
    framework_id,
    rank,
    'PUBLIC' AS visibility,
    created_at,
    updated_at
FROM
    compliance_frameworks
WHERE
    %s
    AND trust_center_id = @trust_center_id
    AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"trust_center_id": trustCenterID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query compliance_frameworks: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ComplianceFramework])
	if err != nil {
		return fmt.Errorf("cannot collect compliance frameworks: %w", err)
	}

	*c = results

	return nil
}

func (c *ComplianceFrameworks) LoadWithHiddenByTrustCenterID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	trustCenterID gid.GID,
	cursor *page.Cursor[ComplianceFrameworkOrderField],
) error {
	q := `
WITH combined AS (
    SELECT
        COALESCE(cf.id, f.id) AS id,
        COALESCE(cf.organization_id, tc.organization_id) AS organization_id,
        COALESCE(cf.trust_center_id, tc.id) AS trust_center_id,
        f.id AS framework_id,
        CASE
            WHEN cf.id IS NOT NULL THEN cf.rank
            ELSE COALESCE(MAX(cf.rank) OVER (), 0) + ROW_NUMBER() OVER (PARTITION BY (cf.id IS NULL) ORDER BY f.created_at)
        END AS rank,
        CASE WHEN cf.id IS NULL THEN 'NONE' ELSE 'PUBLIC' END AS visibility,
        COALESCE(cf.created_at, f.created_at) AS created_at,
        COALESCE(cf.updated_at, f.updated_at) AS updated_at
    FROM trust_centers tc
    JOIN frameworks f
        ON f.organization_id = tc.organization_id
        AND f.tenant_id = @tenant_id
    LEFT JOIN compliance_frameworks cf
        ON cf.framework_id = f.id
        AND cf.trust_center_id = tc.id
        AND cf.tenant_id = @tenant_id
    WHERE tc.id = @trust_center_id
        AND tc.tenant_id = @tenant_id
)
SELECT id, organization_id, trust_center_id, framework_id, rank, visibility, created_at, updated_at
FROM combined
WHERE %s
`
	q = fmt.Sprintf(q, cursor.SQLFragment())

	args := pgx.NamedArgs{"trust_center_id": trustCenterID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query compliance frameworks with hidden: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ComplianceFramework])
	if err != nil {
		return fmt.Errorf("cannot collect compliance frameworks with hidden: %w", err)
	}

	*c = results

	return nil
}
