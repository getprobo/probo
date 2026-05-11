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
	RiskAssessmentThreat struct {
		ID                    gid.GID   `db:"id"`
		OrganizationID        gid.GID   `db:"organization_id"`
		RiskAssessmentScopeID gid.GID   `db:"risk_assessment_scope_id"`
		ProcessID             gid.GID   `db:"process_id"`
		Name                  string    `db:"name"`
		Category              string    `db:"category"`
		CreatedAt             time.Time `db:"created_at"`
		UpdatedAt             time.Time `db:"updated_at"`
	}

	RiskAssessmentThreats []*RiskAssessmentThreat
)

func (t *RiskAssessmentThreat) CursorKey(orderBy RiskAssessmentThreatOrderField) page.CursorKey {
	switch orderBy {
	case RiskAssessmentThreatOrderFieldCreatedAt:
		return page.CursorKey{ID: t.ID, Value: t.CreatedAt}
	case RiskAssessmentThreatOrderFieldName:
		return page.CursorKey{ID: t.ID, Value: t.Name}
	}
	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (t *RiskAssessmentThreat) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	q := `SELECT organization_id FROM risk_assessment_threats WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, t.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query risk assessment threat authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (ts *RiskAssessmentThreats) LoadByRiskAssessmentScopeID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	riskAssessmentScopeID gid.GID,
	cursor *page.Cursor[RiskAssessmentThreatOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	risk_assessment_scope_id,
	process_id,
	name,
	category,
	created_at,
	updated_at
FROM
	risk_assessment_threats
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
		return fmt.Errorf("cannot query risk threats: %w", err)
	}
	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[RiskAssessmentThreat])
	if err != nil {
		return fmt.Errorf("cannot collect risk threats: %w", err)
	}
	*ts = results
	return nil
}

func (ts *RiskAssessmentThreats) CountByRiskAssessmentScopeID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	riskAssessmentScopeID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	risk_assessment_threats
WHERE
	%s
	AND risk_assessment_scope_id = @risk_assessment_scope_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.NamedArgs{"risk_assessment_scope_id": riskAssessmentScopeID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count risk threats: %w", err)
	}
	return count, nil
}

func (t *RiskAssessmentThreat) LoadByID(ctx context.Context, conn pg.Querier, scope Scoper, id gid.GID) error {
	q := `
SELECT
	id,
	organization_id,
	risk_assessment_scope_id,
	process_id,
	name,
	category,
	created_at,
	updated_at
FROM
	risk_assessment_threats
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
		return fmt.Errorf("cannot query risk threat: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[RiskAssessmentThreat])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect risk threat: %w", err)
	}
	*t = result
	return nil
}

func (t *RiskAssessmentThreat) Insert(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
INSERT INTO risk_assessment_threats (
	id,
	tenant_id,
	organization_id,
	risk_assessment_scope_id,
	process_id,
	name,
	category,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@risk_assessment_scope_id,
	@process_id,
	@name,
	@category,
	@created_at,
	@updated_at
)
`
	args := pgx.StrictNamedArgs{
		"id":                       t.ID,
		"tenant_id":                scope.GetTenantID(),
		"organization_id":          t.OrganizationID,
		"risk_assessment_scope_id": t.RiskAssessmentScopeID,
		"process_id":               t.ProcessID,
		"name":                     t.Name,
		"category":                 t.Category,
		"created_at":               t.CreatedAt,
		"updated_at":               t.UpdatedAt,
	}
	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			return ErrResourceAlreadyExists
		}
		return fmt.Errorf("cannot insert risk threat: %w", err)
	}
	return nil
}

func (t *RiskAssessmentThreat) Update(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
UPDATE risk_assessment_threats
SET
	process_id = @process_id,
	name = @name,
	category = @category,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"id":         t.ID,
		"process_id": t.ProcessID,
		"name":       t.Name,
		"category":   t.Category,
		"updated_at": t.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())
	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update risk threat: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}
	return nil
}

func (t *RiskAssessmentThreat) Delete(ctx context.Context, conn pg.Tx, scope Scoper, id gid.GID) error {
	q := `
DELETE FROM risk_assessment_threats
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
