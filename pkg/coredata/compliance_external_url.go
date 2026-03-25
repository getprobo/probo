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
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	ComplianceExternalURL struct {
		ID             gid.GID   `db:"id"`
		OrganizationID gid.GID   `db:"organization_id"`
		TrustCenterID  gid.GID   `db:"trust_center_id"`
		Name           string    `db:"name"`
		URL            string    `db:"url"`
		Rank           int       `db:"rank"`
		CreatedAt      time.Time `db:"created_at"`
		UpdatedAt      time.Time `db:"updated_at"`
	}

	ComplianceExternalURLs []*ComplianceExternalURL
)

func (c ComplianceExternalURL) CursorKey(orderBy ComplianceExternalURLOrderField) page.CursorKey {
	switch orderBy {
	case ComplianceExternalURLOrderFieldCreatedAt:
		return page.NewCursorKey(c.ID, c.CreatedAt)
	case ComplianceExternalURLOrderFieldRank:
		return page.NewCursorKey(c.ID, c.Rank)
	}
	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (c *ComplianceExternalURL) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM compliance_external_urls WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, c.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query compliance external URL authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (c *ComplianceExternalURL) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    trust_center_id,
    name,
    url,
    rank,
    created_at,
    updated_at
FROM
    compliance_external_urls
WHERE
    %s
    AND id = @id
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query compliance_external_urls: %w", err)
	}

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ComplianceExternalURL])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect compliance external URL: %w", err)
	}

	*c = result

	return nil
}

func (c *ComplianceExternalURL) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    compliance_external_urls (
        id,
        tenant_id,
        organization_id,
        trust_center_id,
        name,
        url,
        rank,
        created_at,
        updated_at
    )
VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @trust_center_id,
    @name,
    @url,
    (SELECT COALESCE(MAX(rank), 0) + 1 FROM compliance_external_urls WHERE trust_center_id = @trust_center_id),
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
		"name":            c.Name,
		"url":             c.URL,
		"created_at":      c.CreatedAt,
		"updated_at":      c.UpdatedAt,
	}

	if err := conn.QueryRow(ctx, q, args).Scan(&c.Rank); err != nil {
		return fmt.Errorf("cannot insert compliance external URL: %w", err)
	}

	return nil
}

func (c *ComplianceExternalURL) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE compliance_external_urls
SET
    name = @name,
    url = @url,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":         c.ID,
		"name":       c.Name,
		"url":        c.URL,
		"updated_at": c.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update compliance external URL: %w", err)
	}

	return nil
}

func (c *ComplianceExternalURL) UpdateRank(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
WITH old AS (
  SELECT
    rank AS old_rank
  FROM compliance_external_urls
  WHERE %s AND id = @id AND trust_center_id = @trust_center_id
)

UPDATE compliance_external_urls
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
		"id":              c.ID,
		"new_rank":        c.Rank,
		"trust_center_id": c.TrustCenterID,
		"updated_at":      c.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update compliance external URL rank: %w", err)
	}

	return nil
}

func (c *ComplianceExternalURL) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM
    compliance_external_urls
WHERE
    %s
    AND id = @id;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": c.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete compliance external URL: %w", err)
	}

	return nil
}

func (c *ComplianceExternalURLs) LoadByTrustCenterID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterID gid.GID,
	cursor *page.Cursor[ComplianceExternalURLOrderField],
) error {
	q := `
SELECT
    id,
    organization_id,
    trust_center_id,
    name,
    url,
    rank,
    created_at,
    updated_at
FROM
    compliance_external_urls
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
		return fmt.Errorf("cannot query compliance_external_urls: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ComplianceExternalURL])
	if err != nil {
		return fmt.Errorf("cannot collect compliance external URLs: %w", err)
	}

	*c = results

	return nil
}
