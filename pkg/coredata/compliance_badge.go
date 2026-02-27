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
	ComplianceBadge struct {
		ID             gid.GID   `db:"id"`
		OrganizationID gid.GID   `db:"organization_id"`
		TrustCenterID  gid.GID   `db:"trust_center_id"`
		Name           string    `db:"name"`
		IconFileID     gid.GID   `db:"icon_file_id"`
		Rank           int       `db:"rank"`
		CreatedAt      time.Time `db:"created_at"`
		UpdatedAt      time.Time `db:"updated_at"`
	}

	ComplianceBadges []*ComplianceBadge
)

func (t ComplianceBadge) CursorKey(orderBy ComplianceBadgeOrderField) page.CursorKey {
	switch orderBy {
	case ComplianceBadgeOrderFieldRank:
		return page.NewCursorKey(t.ID, t.Rank)
	case ComplianceBadgeOrderFieldName:
		return page.NewCursorKey(t.ID, t.Name)
	case ComplianceBadgeOrderFieldCreatedAt:
		return page.NewCursorKey(t.ID, t.CreatedAt)
	case ComplianceBadgeOrderFieldUpdatedAt:
		return page.NewCursorKey(t.ID, t.UpdatedAt)
	}
	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (t *ComplianceBadge) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM compliance_badges WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, t.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query compliance badge authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (t *ComplianceBadge) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	complianceBadgeID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    trust_center_id,
    name,
    icon_file_id,
    rank,
    created_at,
    updated_at
FROM
    compliance_badges
WHERE
    %s
    AND id = @compliance_badge_id
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"compliance_badge_id": complianceBadgeID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query compliance_badges: %w", err)
	}

	badge, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ComplianceBadge])
	if err != nil {
		return fmt.Errorf("cannot collect compliance badge: %w", err)
	}

	*t = badge

	return nil
}

func (t *ComplianceBadge) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    compliance_badges (
        tenant_id,
        id,
        organization_id,
        trust_center_id,
        name,
        icon_file_id,
        rank,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @id,
    @organization_id,
    @trust_center_id,
    @name,
    @icon_file_id,
    (SELECT COALESCE(MAX(rank), 0) + 1 FROM compliance_badges WHERE trust_center_id = @trust_center_id),
    @created_at,
    @updated_at
)
RETURNING rank;
`

	args := pgx.StrictNamedArgs{
		"tenant_id":       scope.GetTenantID(),
		"id":              t.ID,
		"organization_id": t.OrganizationID,
		"trust_center_id": t.TrustCenterID,
		"name":            t.Name,
		"icon_file_id":    t.IconFileID,
		"created_at":      t.CreatedAt,
		"updated_at":      t.UpdatedAt,
	}

	err := conn.QueryRow(ctx, q, args).Scan(&t.Rank)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "compliance_badges_trust_center_id_rank_key" {
				return ErrResourceAlreadyExists
			}
		}
		return fmt.Errorf("cannot insert compliance badge: %w", err)
	}

	return nil
}

func (t *ComplianceBadge) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE compliance_badges
SET
    name = @name,
    icon_file_id = @icon_file_id,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":           t.ID,
		"name":         t.Name,
		"icon_file_id": t.IconFileID,
		"updated_at":   t.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update compliance badge: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (t *ComplianceBadge) UpdateRank(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
WITH old AS (
  SELECT
	rank AS old_rank
  FROM compliance_badges
  WHERE %s AND id = @id AND trust_center_id = @trust_center_id
)

UPDATE compliance_badges
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
  AND trust_center_id = @trust_center_id
  AND (
    id = @id
    OR (rank BETWEEN LEAST(old.old_rank, @new_rank) AND GREATEST(old.old_rank, @new_rank))
  );
`

	scopeFragment := scope.SQLFragment()
	q = fmt.Sprintf(q, scopeFragment, scopeFragment)

	args := pgx.StrictNamedArgs{
		"id":              t.ID,
		"new_rank":        t.Rank,
		"trust_center_id": t.TrustCenterID,
		"updated_at":      t.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update compliance badge rank: %w", err)
	}

	return nil
}

func (t *ComplianceBadge) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM
    compliance_badges
WHERE
    %s
    AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": t.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete compliance badge: %w", err)
	}

	return nil
}

func (t *ComplianceBadges) LoadByTrustCenterID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterID gid.GID,
	cursor *page.Cursor[ComplianceBadgeOrderField],
) error {
	q := `
SELECT
    id,
    organization_id,
    trust_center_id,
    name,
    icon_file_id,
    rank,
    created_at,
    updated_at
FROM
    compliance_badges
WHERE
    %s
    AND trust_center_id = @trust_center_id
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"trust_center_id": trustCenterID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query compliance_badges: %w", err)
	}

	badges, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ComplianceBadge])
	if err != nil {
		return fmt.Errorf("cannot collect compliance badges: %w", err)
	}

	*t = badges

	return nil
}

func (t *ComplianceBadges) CountByTrustCenterID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterID gid.GID,
) (int, error) {
	q := `
SELECT
    COUNT(*)
FROM
    compliance_badges
WHERE
    %s
    AND trust_center_id = @trust_center_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"trust_center_id": trustCenterID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	err := conn.QueryRow(ctx, q, args).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count compliance badges: %w", err)
	}

	return count, nil
}
