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
	CompliancePortalCommitment struct {
		ID                 gid.GID                        `db:"id"`
		OrganizationID     gid.GID                        `db:"organization_id"`
		CompliancePortalID gid.GID                        `db:"trust_center_id"`
		GroupID            gid.GID                        `db:"group_id"`
		Icon               CompliancePortalCommitmentIcon `db:"icon"`
		Eyebrow            string                         `db:"eyebrow"`
		Title              string                         `db:"title"`
		Description        string                         `db:"description"`
		Rank               int                            `db:"rank"`
		CreatedAt          time.Time                      `db:"created_at"`
		UpdatedAt          time.Time                      `db:"updated_at"`
	}

	CompliancePortalCommitments []*CompliancePortalCommitment
)

func (t CompliancePortalCommitment) CursorKey(orderBy CompliancePortalCommitmentOrderField) page.CursorKey {
	switch orderBy {
	case CompliancePortalCommitmentOrderFieldRank:
		return page.NewCursorKey(t.ID, t.Rank)
	case CompliancePortalCommitmentOrderFieldCreatedAt:
		return page.NewCursorKey(t.ID, t.CreatedAt)
	case CompliancePortalCommitmentOrderFieldUpdatedAt:
		return page.NewCursorKey(t.ID, t.UpdatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (t *CompliancePortalCommitment) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM compliance_portal_commitments WHERE id = ANY(@resource_ids::text[])`

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

func (t *CompliancePortalCommitment) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	commitmentID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    trust_center_id,
    group_id,
    icon,
    eyebrow,
    title,
    description,
    rank,
    created_at,
    updated_at
FROM
    compliance_portal_commitments
WHERE
    %s
    AND id = @commitment_id
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"commitment_id": commitmentID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query compliance_portal_commitments: %w", err)
	}

	commitment, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CompliancePortalCommitment])
	if err != nil {
		return fmt.Errorf("cannot collect compliance portal commitment: %w", err)
	}

	*t = commitment

	return nil
}

func (t *CompliancePortalCommitment) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    compliance_portal_commitments (
        tenant_id,
        id,
        organization_id,
        trust_center_id,
        group_id,
        icon,
        eyebrow,
        title,
        description,
        rank,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @id,
    @organization_id,
    @trust_center_id,
    @group_id,
    @icon,
    @eyebrow,
    @title,
    @description,
    (SELECT COALESCE(MAX(rank), 0) + 1 FROM compliance_portal_commitments WHERE group_id = @group_id),
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
		"group_id":        t.GroupID,
		"icon":            t.Icon,
		"eyebrow":         t.Eyebrow,
		"title":           t.Title,
		"description":     t.Description,
		"created_at":      t.CreatedAt,
		"updated_at":      t.UpdatedAt,
	}

	err := conn.QueryRow(ctx, q, args).Scan(&t.Rank)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "compliance_portal_commitments_group_id_rank_key" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert compliance portal commitment: %w", err)
	}

	return nil
}

func (t *CompliancePortalCommitment) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE compliance_portal_commitments
SET
    icon = @icon,
    eyebrow = @eyebrow,
    title = @title,
    description = @description,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":          t.ID,
		"icon":        t.Icon,
		"eyebrow":     t.Eyebrow,
		"title":       t.Title,
		"description": t.Description,
		"updated_at":  t.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update compliance portal commitment: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (t *CompliancePortalCommitment) UpdateRank(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
WITH old AS (
  SELECT
	rank AS old_rank
  FROM compliance_portal_commitments
  WHERE %s AND id = @id AND group_id = @group_id
)
UPDATE compliance_portal_commitments
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
  AND group_id = @group_id
  AND (
    id = @id
    OR (rank BETWEEN LEAST(old.old_rank, @new_rank) AND GREATEST(old.old_rank, @new_rank))
  );
`

	scopeFragment := scope.SQLFragment()
	q = fmt.Sprintf(q, scopeFragment, scopeFragment)

	args := pgx.StrictNamedArgs{
		"id":         t.ID,
		"new_rank":   t.Rank,
		"group_id":   t.GroupID,
		"updated_at": t.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update compliance portal commitment rank: %w", err)
	}

	return nil
}

func (t *CompliancePortalCommitment) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM
    compliance_portal_commitments
WHERE
    %s
    AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": t.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete compliance portal commitment: %w", err)
	}

	return nil
}

func (t *CompliancePortalCommitments) LoadByGroupID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	groupID gid.GID,
	cursor *page.Cursor[CompliancePortalCommitmentOrderField],
) error {
	q := `
SELECT
    id,
    organization_id,
    trust_center_id,
    group_id,
    icon,
    eyebrow,
    title,
    description,
    rank,
    created_at,
    updated_at
FROM
    compliance_portal_commitments
WHERE
    %s
    AND group_id = @group_id
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"group_id": groupID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query compliance_portal_commitments: %w", err)
	}

	commitments, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CompliancePortalCommitment])
	if err != nil {
		return fmt.Errorf("cannot collect compliance portal commitments: %w", err)
	}

	*t = commitments

	return nil
}

func (t *CompliancePortalCommitments) CountByGroupID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	groupID gid.GID,
) (int, error) {
	q := `
SELECT
    COUNT(*)
FROM
    compliance_portal_commitments
WHERE
    %s
    AND group_id = @group_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"group_id": groupID}
	maps.Copy(args, scope.SQLArguments())

	var count int

	err := conn.QueryRow(ctx, q, args).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count compliance portal commitments: %w", err)
	}

	return count, nil
}
