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
	RiskAssessmentBoundary struct {
		ID                    gid.GID   `db:"id"`
		OrganizationID        gid.GID   `db:"organization_id"`
		RiskAssessmentScopeID gid.GID   `db:"risk_assessment_scope_id"`
		ParentBoundaryID      *gid.GID  `db:"parent_boundary_id"`
		Name                  string    `db:"name"`
		CreatedAt             time.Time `db:"created_at"`
		UpdatedAt             time.Time `db:"updated_at"`
	}

	RiskAssessmentBoundaries []*RiskAssessmentBoundary
)

func (b *RiskAssessmentBoundary) CursorKey(orderBy RiskAssessmentBoundaryOrderField) page.CursorKey {
	switch orderBy {
	case RiskAssessmentBoundaryOrderFieldCreatedAt:
		return page.CursorKey{ID: b.ID, Value: b.CreatedAt}
	case RiskAssessmentBoundaryOrderFieldName:
		return page.CursorKey{ID: b.ID, Value: b.Name}
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (b *RiskAssessmentBoundary) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM risk_assessment_boundaries WHERE id = ANY(@resource_ids::text[])`

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

func (bs *RiskAssessmentBoundaries) LoadByRiskAssessmentScopeID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	riskAssessmentScopeID gid.GID,
	cursor *page.Cursor[RiskAssessmentBoundaryOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	risk_assessment_scope_id,
	parent_boundary_id,
	name,
	created_at,
	updated_at
FROM
	risk_assessment_boundaries
WHERE
	%s
	AND risk_assessment_scope_id = @risk_assessment_scope_id
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())
	args := pgx.NamedArgs{"risk_assessment_scope_id": riskAssessmentScopeID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query risk assessment boundaries: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[RiskAssessmentBoundary])
	if err != nil {
		return fmt.Errorf("cannot collect risk assessment boundaries: %w", err)
	}

	*bs = results

	return nil
}

func (bs *RiskAssessmentBoundaries) CountByRiskAssessmentScopeID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	riskAssessmentScopeID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	risk_assessment_boundaries
WHERE
	%s
	AND risk_assessment_scope_id = @risk_assessment_scope_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.NamedArgs{"risk_assessment_scope_id": riskAssessmentScopeID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count risk assessment boundaries: %w", err)
	}

	return count, nil
}

func (b *RiskAssessmentBoundary) LoadByID(ctx context.Context, conn pg.Querier, scope Scoper, id gid.GID) error {
	q := `
SELECT
	id,
	organization_id,
	risk_assessment_scope_id,
	parent_boundary_id,
	name,
	created_at,
	updated_at
FROM
	risk_assessment_boundaries
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
		return fmt.Errorf("cannot query risk assessment boundary: %w", err)
	}

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[RiskAssessmentBoundary])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect risk assessment boundary: %w", err)
	}

	*b = result

	return nil
}

func (b *RiskAssessmentBoundary) Insert(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
INSERT INTO risk_assessment_boundaries (
	id,
	tenant_id,
	organization_id,
	risk_assessment_scope_id,
	parent_boundary_id,
	name,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@risk_assessment_scope_id,
	@parent_boundary_id,
	@name,
	@created_at,
	@updated_at
)
`
	args := pgx.StrictNamedArgs{
		"id":                       b.ID,
		"tenant_id":                scope.GetTenantID(),
		"organization_id":          b.OrganizationID,
		"risk_assessment_scope_id": b.RiskAssessmentScopeID,
		"parent_boundary_id":       b.ParentBoundaryID,
		"name":                     b.Name,
		"created_at":               b.CreatedAt,
		"updated_at":               b.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" && pgErr.ConstraintName == "risk_assessment_boundaries_unique_name" {
			return ErrResourceAlreadyExists
		}

		return fmt.Errorf("cannot insert risk assessment boundary: %w", err)
	}

	return nil
}

func (b *RiskAssessmentBoundary) Update(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
UPDATE risk_assessment_boundaries
SET
	parent_boundary_id = @parent_boundary_id,
	name = @name,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"id":                 b.ID,
		"parent_boundary_id": b.ParentBoundaryID,
		"name":               b.Name,
		"updated_at":         b.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" && pgErr.ConstraintName == "risk_assessment_boundaries_unique_name" {
			return ErrResourceAlreadyExists
		}

		return fmt.Errorf("cannot update risk assessment boundary: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (b *RiskAssessmentBoundary) Delete(ctx context.Context, conn pg.Tx, scope Scoper, id gid.GID) error {
	q := `
DELETE FROM risk_assessment_boundaries
WHERE
	%s
	AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)

	return err
}
