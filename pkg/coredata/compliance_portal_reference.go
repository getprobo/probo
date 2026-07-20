// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	CompliancePortalReference struct {
		ID                 gid.GID   `db:"id"`
		OrganizationID     gid.GID   `db:"organization_id"`
		CompliancePortalID gid.GID   `db:"trust_center_id"`
		Name               string    `db:"name"`
		Description        *string   `db:"description"`
		WebsiteURL         string    `db:"website_url"`
		LogoFileID         gid.GID   `db:"logo_file_id"`
		Rank               int       `db:"rank"`
		CreatedAt          time.Time `db:"created_at"`
		UpdatedAt          time.Time `db:"updated_at"`
	}

	CompliancePortalReferences []*CompliancePortalReference
)

func (t CompliancePortalReference) CursorKey(orderBy CompliancePortalReferenceOrderField) page.CursorKey {
	switch orderBy {
	case CompliancePortalReferenceOrderFieldRank:
		return page.NewCursorKey(t.ID, t.Rank)
	case CompliancePortalReferenceOrderFieldName:
		return page.NewCursorKey(t.ID, t.Name)
	case CompliancePortalReferenceOrderFieldCreatedAt:
		return page.NewCursorKey(t.ID, t.CreatedAt)
	case CompliancePortalReferenceOrderFieldUpdatedAt:
		return page.NewCursorKey(t.ID, t.UpdatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (t *CompliancePortalReference) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM trust_center_references WHERE id = ANY(@resource_ids::text[])`

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

func (t *CompliancePortalReference) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalReferenceID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    trust_center_id,
    name,
    description,
    website_url,
    logo_file_id,
    rank,
    created_at,
    updated_at
FROM
    trust_center_references
WHERE
    %s
    AND id = @trust_center_reference_id
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"trust_center_reference_id": compliancePortalReferenceID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust_center_references: %w", err)
	}

	reference, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CompliancePortalReference])
	if err != nil {
		return fmt.Errorf("cannot collect compliance portal reference: %w", err)
	}

	*t = reference

	return nil
}

func (t *CompliancePortalReference) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    trust_center_references (
        tenant_id,
        id,
        organization_id,
        trust_center_id,
        name,
        description,
        website_url,
        logo_file_id,
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
    @description,
    @website_url,
    @logo_file_id,
    (SELECT COALESCE(MAX(rank), 0) + 1 FROM trust_center_references WHERE trust_center_id = @trust_center_id),
    @created_at,
    @updated_at
)
RETURNING rank;
`

	args := pgx.StrictNamedArgs{
		"tenant_id":       scope.GetTenantID(),
		"id":              t.ID,
		"organization_id": t.OrganizationID,
		"trust_center_id": t.CompliancePortalID,
		"name":            t.Name,
		"description":     t.Description,
		"website_url":     t.WebsiteURL,
		"logo_file_id":    t.LogoFileID,
		"created_at":      t.CreatedAt,
		"updated_at":      t.UpdatedAt,
	}

	err := conn.QueryRow(ctx, q, args).Scan(&t.Rank)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "trust_center_references_trust_center_id_rank_key" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert compliance portal reference: %w", err)
	}

	return nil
}

func (t *CompliancePortalReference) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE trust_center_references
SET
    name = @name,
    description = @description,
    website_url = @website_url,
    logo_file_id = @logo_file_id,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":           t.ID,
		"name":         t.Name,
		"description":  t.Description,
		"website_url":  t.WebsiteURL,
		"logo_file_id": t.LogoFileID,
		"updated_at":   t.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update compliance portal reference: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (t *CompliancePortalReference) UpdateRank(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
WITH old AS (
  SELECT
	rank AS old_rank
  FROM trust_center_references
  WHERE %s AND id = @id AND trust_center_id = @trust_center_id
)

UPDATE trust_center_references
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
		"id":              t.ID,
		"new_rank":        t.Rank,
		"trust_center_id": t.CompliancePortalID,
		"updated_at":      t.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update compliance portal reference rank: %w", err)
	}

	return nil
}

func (t *CompliancePortalReference) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM
    trust_center_references
WHERE
    %s
    AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": t.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete compliance portal reference: %w", err)
	}

	return nil
}

func (t *CompliancePortalReferences) LoadByCompliancePortalID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalID gid.GID,
	cursor *page.Cursor[CompliancePortalReferenceOrderField],
) error {
	q := `
SELECT
    id,
    organization_id,
    trust_center_id,
    name,
    description,
    website_url,
    logo_file_id,
    rank,
    created_at,
    updated_at
FROM
    trust_center_references
WHERE
    %s
    AND trust_center_id = @trust_center_id
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"trust_center_id": compliancePortalID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust_center_references: %w", err)
	}

	references, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CompliancePortalReference])
	if err != nil {
		return fmt.Errorf("cannot collect compliance portal references: %w", err)
	}

	*t = references

	return nil
}

func (t *CompliancePortalReferences) CountByCompliancePortalID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalID gid.GID,
) (int, error) {
	q := `
SELECT
    COUNT(*)
FROM
    trust_center_references
WHERE
    %s
    AND trust_center_id = @trust_center_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"trust_center_id": compliancePortalID}
	maps.Copy(args, scope.SQLArguments())

	var count int

	err := conn.QueryRow(ctx, q, args).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count compliance portal references: %w", err)
	}

	return count, nil
}
