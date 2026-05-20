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
	RiskAssessmentProcess struct {
		ID                    gid.GID   `db:"id"`
		OrganizationID        gid.GID   `db:"organization_id"`
		RiskAssessmentScopeID gid.GID   `db:"risk_assessment_scope_id"`
		SourceNodeID          gid.GID   `db:"source_node_id"`
		TargetNodeID          gid.GID   `db:"target_node_id"`
		Name                  string    `db:"name"`
		CreatedAt             time.Time `db:"created_at"`
		UpdatedAt             time.Time `db:"updated_at"`
	}

	RiskAssessmentProcesses []*RiskAssessmentProcess
)

func (p *RiskAssessmentProcess) CursorKey(orderBy RiskAssessmentProcessOrderField) page.CursorKey {
	switch orderBy {
	case RiskAssessmentProcessOrderFieldCreatedAt:
		return page.CursorKey{ID: p.ID, Value: p.CreatedAt}
	case RiskAssessmentProcessOrderFieldName:
		return page.CursorKey{ID: p.ID, Value: p.Name}
	}
	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (p *RiskAssessmentProcess) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	q := `SELECT organization_id FROM risk_assessment_processes WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, p.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query risk assessment process authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (ps *RiskAssessmentProcesses) LoadByRiskAssessmentScopeID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	riskAssessmentScopeID gid.GID,
	cursor *page.Cursor[RiskAssessmentProcessOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	risk_assessment_scope_id,
	source_node_id,
	target_node_id,
	name,
	created_at,
	updated_at
FROM
	risk_assessment_processes
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
		return fmt.Errorf("cannot query risk assessment processes: %w", err)
	}
	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[RiskAssessmentProcess])
	if err != nil {
		return fmt.Errorf("cannot collect risk assessment processes: %w", err)
	}
	*ps = results
	return nil
}

func (ps *RiskAssessmentProcesses) LoadAllByRiskAssessmentScopeID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	riskAssessmentScopeID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	risk_assessment_scope_id,
	source_node_id,
	target_node_id,
	name,
	created_at,
	updated_at
FROM
	risk_assessment_processes
WHERE
	%s
	AND risk_assessment_scope_id = @risk_assessment_scope_id
ORDER BY
	created_at ASC, id ASC
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.NamedArgs{"risk_assessment_scope_id": riskAssessmentScopeID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query risk assessment processes: %w", err)
	}
	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[RiskAssessmentProcess])
	if err != nil {
		return fmt.Errorf("cannot collect risk assessment processes: %w", err)
	}
	*ps = results
	return nil
}

func (ps *RiskAssessmentProcesses) CountByRiskAssessmentScopeID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	riskAssessmentScopeID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	risk_assessment_processes
WHERE
	%s
	AND risk_assessment_scope_id = @risk_assessment_scope_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.NamedArgs{"risk_assessment_scope_id": riskAssessmentScopeID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count risk assessment processes: %w", err)
	}
	return count, nil
}

func (p *RiskAssessmentProcess) LoadByID(ctx context.Context, conn pg.Querier, scope Scoper, id gid.GID) error {
	q := `
SELECT
	id,
	organization_id,
	risk_assessment_scope_id,
	source_node_id,
	target_node_id,
	name,
	created_at,
	updated_at
FROM
	risk_assessment_processes
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
		return fmt.Errorf("cannot query risk assessment process: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[RiskAssessmentProcess])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect risk assessment process: %w", err)
	}
	*p = result
	return nil
}

func (p *RiskAssessmentProcess) Insert(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
INSERT INTO risk_assessment_processes (
	id,
	tenant_id,
	organization_id,
	risk_assessment_scope_id,
	source_node_id,
	target_node_id,
	name,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@risk_assessment_scope_id,
	@source_node_id,
	@target_node_id,
	@name,
	@created_at,
	@updated_at
)
`
	args := pgx.StrictNamedArgs{
		"id":                       p.ID,
		"tenant_id":                scope.GetTenantID(),
		"organization_id":          p.OrganizationID,
		"risk_assessment_scope_id": p.RiskAssessmentScopeID,
		"source_node_id":           p.SourceNodeID,
		"target_node_id":           p.TargetNodeID,
		"name":                     p.Name,
		"created_at":               p.CreatedAt,
		"updated_at":               p.UpdatedAt,
	}
	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" && pgErr.ConstraintName == "risk_assessment_processes_unique_name" {
			return ErrResourceAlreadyExists
		}
		return fmt.Errorf("cannot insert risk assessment process: %w", err)
	}
	return nil
}

func (p *RiskAssessmentProcess) Update(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
UPDATE risk_assessment_processes
SET
	source_node_id = @source_node_id,
	target_node_id = @target_node_id,
	name = @name,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"id":             p.ID,
		"source_node_id": p.SourceNodeID,
		"target_node_id": p.TargetNodeID,
		"name":           p.Name,
		"updated_at":     p.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())
	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update risk assessment process: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}
	return nil
}

func (p *RiskAssessmentProcess) Delete(ctx context.Context, conn pg.Tx, scope Scoper, id gid.GID) error {
	q := `
DELETE FROM risk_assessment_processes
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
