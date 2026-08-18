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
	RiskAnalysisDiagram struct {
		ID             gid.GID   `db:"id"`
		OrganizationID gid.GID   `db:"organization_id"`
		RiskAnalysisID gid.GID   `db:"risk_analysis_id"`
		Name           string    `db:"name"`
		CreatedAt      time.Time `db:"created_at"`
		UpdatedAt      time.Time `db:"updated_at"`
	}

	RiskAnalysisDiagrams []*RiskAnalysisDiagram
)

func (s *RiskAnalysisDiagram) CursorKey(orderBy RiskAnalysisDiagramOrderField) page.CursorKey {
	switch orderBy {
	case RiskAnalysisDiagramOrderFieldCreatedAt:
		return page.CursorKey{ID: s.ID, Value: s.CreatedAt}
	case RiskAnalysisDiagramOrderFieldName:
		return page.CursorKey{ID: s.ID, Value: s.Name}
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (s *RiskAnalysisDiagram) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM risk_analysis_diagrams WHERE id = ANY(@resource_ids)`

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

func (ss *RiskAnalysisDiagrams) LoadByRiskAnalysisID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	riskAnalysisID gid.GID,
	cursor *page.Cursor[RiskAnalysisDiagramOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	risk_analysis_id,
	name,
	created_at,
	updated_at
FROM
	risk_analysis_diagrams
WHERE
	%s
	AND risk_analysis_id = @risk_analysis_id
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())
	args := pgx.NamedArgs{"risk_analysis_id": riskAnalysisID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query risk analysis diagrams: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[RiskAnalysisDiagram])
	if err != nil {
		return fmt.Errorf("cannot collect risk analysis diagrams: %w", err)
	}

	*ss = results

	return nil
}

func (ss *RiskAnalysisDiagrams) CountByRiskAnalysisID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	riskAnalysisID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	risk_analysis_diagrams
WHERE
	%s
	AND risk_analysis_id = @risk_analysis_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.NamedArgs{"risk_analysis_id": riskAnalysisID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count risk analysis diagrams: %w", err)
	}

	return count, nil
}

func (s *RiskAnalysisDiagram) LoadByID(ctx context.Context, conn pg.Querier, scope Scoper, id gid.GID) error {
	q := `
SELECT
	id,
	organization_id,
	risk_analysis_id,
	name,
	created_at,
	updated_at
FROM
	risk_analysis_diagrams
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
		return fmt.Errorf("cannot query risk analysis diagram: %w", err)
	}

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[RiskAnalysisDiagram])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect risk analysis diagram: %w", err)
	}

	*s = result

	return nil
}

func (ss *RiskAnalysisDiagrams) LoadByIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	diagramIDs []gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	risk_analysis_id,
	name,
	created_at,
	updated_at
FROM
	risk_analysis_diagrams
WHERE
	%s
	AND id = ANY(@diagram_ids)
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"diagram_ids": diagramIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query risk analysis diagrams: %w", err)
	}

	diagrams, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[RiskAnalysisDiagram])
	if err != nil {
		return fmt.Errorf("cannot collect risk analysis diagrams: %w", err)
	}

	*ss = diagrams

	if len(diagrams) != len(gid.NewSet(diagramIDs...)) {
		return ErrResourceNotFound
	}

	return nil
}

func (s *RiskAnalysisDiagram) Insert(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
INSERT INTO risk_analysis_diagrams (
	id,
	tenant_id,
	organization_id,
	risk_analysis_id,
	name,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@risk_analysis_id,
	@name,
	@created_at,
	@updated_at
)
`
	args := pgx.StrictNamedArgs{
		"id":               s.ID,
		"tenant_id":        scope.GetTenantID(),
		"organization_id":  s.OrganizationID,
		"risk_analysis_id": s.RiskAnalysisID,
		"name":             s.Name,
		"created_at":       s.CreatedAt,
		"updated_at":       s.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert risk analysis diagram: %w", err)
	}

	return nil
}

func (s *RiskAnalysisDiagram) Update(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
UPDATE risk_analysis_diagrams
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
		return fmt.Errorf("cannot update risk analysis diagram: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (s *RiskAnalysisDiagram) Delete(ctx context.Context, conn pg.Tx, scope Scoper, id gid.GID) error {
	q := `
DELETE FROM risk_analysis_diagrams
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
