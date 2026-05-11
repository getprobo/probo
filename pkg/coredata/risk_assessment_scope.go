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
	RiskAssessmentScope struct {
		ID               gid.GID   `db:"id"`
		OrganizationID   gid.GID   `db:"organization_id"`
		RiskAssessmentID gid.GID   `db:"risk_assessment_id"`
		Name             string    `db:"name"`
		CreatedAt        time.Time `db:"created_at"`
		UpdatedAt        time.Time `db:"updated_at"`
	}

	RiskAssessmentScopes []*RiskAssessmentScope
)

func (s *RiskAssessmentScope) CursorKey(orderBy RiskAssessmentScopeOrderField) page.CursorKey {
	switch orderBy {
	case RiskAssessmentScopeOrderFieldCreatedAt:
		return page.CursorKey{ID: s.ID, Value: s.CreatedAt}
	case RiskAssessmentScopeOrderFieldName:
		return page.CursorKey{ID: s.ID, Value: s.Name}
	}
	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (s *RiskAssessmentScope) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	q := `SELECT organization_id FROM risk_assessment_scopes WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, s.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query risk assessment scope authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (ss *RiskAssessmentScopes) LoadByRiskAssessmentID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	riskAssessmentID gid.GID,
	cursor *page.Cursor[RiskAssessmentScopeOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	risk_assessment_id,
	name,
	created_at,
	updated_at
FROM
	risk_assessment_scopes
WHERE
	%s
	AND risk_assessment_id = @risk_assessment_id
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())
	args := pgx.NamedArgs{"risk_assessment_id": riskAssessmentID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query risk assessment scopes: %w", err)
	}
	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[RiskAssessmentScope])
	if err != nil {
		return fmt.Errorf("cannot collect risk assessment scopes: %w", err)
	}
	*ss = results
	return nil
}

func (ss *RiskAssessmentScopes) CountByRiskAssessmentID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	riskAssessmentID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	risk_assessment_scopes
WHERE
	%s
	AND risk_assessment_id = @risk_assessment_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.NamedArgs{"risk_assessment_id": riskAssessmentID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count risk assessment scopes: %w", err)
	}
	return count, nil
}

func (s *RiskAssessmentScope) LoadByID(ctx context.Context, conn pg.Querier, scope Scoper, id gid.GID) error {
	q := `
SELECT
	id,
	organization_id,
	risk_assessment_id,
	name,
	created_at,
	updated_at
FROM
	risk_assessment_scopes
WHERE
	%s
	AND id = @id
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())
	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query risk assessment scope: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[RiskAssessmentScope])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect risk assessment scope: %w", err)
	}
	*s = result
	return nil
}

func (s *RiskAssessmentScope) Insert(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
INSERT INTO risk_assessment_scopes (
	id,
	tenant_id,
	organization_id,
	risk_assessment_id,
	name,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@risk_assessment_id,
	@name,
	@created_at,
	@updated_at
)
`
	args := pgx.StrictNamedArgs{
		"id":                 s.ID,
		"tenant_id":          scope.GetTenantID(),
		"organization_id":    s.OrganizationID,
		"risk_assessment_id": s.RiskAssessmentID,
		"name":               s.Name,
		"created_at":         s.CreatedAt,
		"updated_at":         s.UpdatedAt,
	}
	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			return ErrResourceAlreadyExists
		}
		return fmt.Errorf("cannot insert risk assessment scope: %w", err)
	}
	return nil
}

func (s *RiskAssessmentScope) Update(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
UPDATE risk_assessment_scopes
SET
	name = @name,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{"id": s.ID, "name": s.Name, "updated_at": s.UpdatedAt}
	maps.Copy(args, scope.SQLArguments())
	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update risk assessment scope: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}
	return nil
}

func (s *RiskAssessmentScope) Delete(ctx context.Context, conn pg.Tx, scope Scoper, id gid.GID) error {
	q := `
DELETE FROM risk_assessment_scopes
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
