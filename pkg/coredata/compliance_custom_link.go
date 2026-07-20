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
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	ComplianceCustomLink struct {
		ID                 gid.GID   `db:"id"`
		OrganizationID     gid.GID   `db:"organization_id"`
		CompliancePortalID gid.GID   `db:"trust_center_id"`
		Name               string    `db:"name"`
		URL                string    `db:"url"`
		Rank               int       `db:"rank"`
		CreatedAt          time.Time `db:"created_at"`
		UpdatedAt          time.Time `db:"updated_at"`
	}

	ComplianceCustomLinks []*ComplianceCustomLink
)

func (c ComplianceCustomLink) CursorKey(orderBy ComplianceCustomLinkOrderField) page.CursorKey {
	switch orderBy {
	case ComplianceCustomLinkOrderFieldCreatedAt:
		return page.NewCursorKey(c.ID, c.CreatedAt)
	case ComplianceCustomLinkOrderFieldRank:
		return page.NewCursorKey(c.ID, c.Rank)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (c *ComplianceCustomLink) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM compliance_custom_links WHERE id = ANY(@resource_ids::text[])`

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

func (c *ComplianceCustomLink) LoadByID(
	ctx context.Context,
	conn pg.Querier,
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
    compliance_custom_links
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
		return fmt.Errorf("cannot query compliance_custom_links: %w", err)
	}

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ComplianceCustomLink])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect compliance custom link: %w", err)
	}

	*c = result

	return nil
}

func (c *ComplianceCustomLink) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    compliance_custom_links (
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
    (SELECT COALESCE(MAX(rank), 0) + 1 FROM compliance_custom_links WHERE trust_center_id = @trust_center_id),
    @created_at,
    @updated_at
)
RETURNING rank;
`

	args := pgx.StrictNamedArgs{
		"id":              c.ID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": c.OrganizationID,
		"trust_center_id": c.CompliancePortalID,
		"name":            c.Name,
		"url":             c.URL,
		"created_at":      c.CreatedAt,
		"updated_at":      c.UpdatedAt,
	}

	if err := conn.QueryRow(ctx, q, args).Scan(&c.Rank); err != nil {
		return fmt.Errorf("cannot insert compliance custom link: %w", err)
	}

	return nil
}

func (c *ComplianceCustomLink) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE compliance_custom_links
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
		return fmt.Errorf("cannot update compliance custom link: %w", err)
	}

	return nil
}

func (c *ComplianceCustomLink) UpdateRank(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
WITH old AS (
  SELECT
    rank AS old_rank
  FROM compliance_custom_links
  WHERE %s AND id = @id AND trust_center_id = @trust_center_id
)

UPDATE compliance_custom_links
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
		"trust_center_id": c.CompliancePortalID,
		"updated_at":      c.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update compliance custom link rank: %w", err)
	}

	return nil
}

func (c *ComplianceCustomLink) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM
    compliance_custom_links
WHERE
    %s
    AND id = @id;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": c.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete compliance custom link: %w", err)
	}

	return nil
}

func (c *ComplianceCustomLinks) LoadByCompliancePortalID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalID gid.GID,
	cursor *page.Cursor[ComplianceCustomLinkOrderField],
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
    compliance_custom_links
WHERE
    %s
    AND trust_center_id = @trust_center_id
    AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"trust_center_id": compliancePortalID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query compliance_custom_links: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ComplianceCustomLink])
	if err != nil {
		return fmt.Errorf("cannot collect compliance custom links: %w", err)
	}

	*c = results

	return nil
}
