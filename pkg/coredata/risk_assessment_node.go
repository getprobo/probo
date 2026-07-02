// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	RiskAssessmentNode struct {
		ID                    gid.GID                `db:"id"`
		OrganizationID        gid.GID                `db:"organization_id"`
		RiskAssessmentScopeID gid.GID                `db:"risk_assessment_scope_id"`
		BoundaryID            *gid.GID               `db:"boundary_id"`
		NodeType              RiskAssessmentNodeType `db:"node_type"`
		Name                  string                 `db:"name"`
		CreatedAt             time.Time              `db:"created_at"`
		UpdatedAt             time.Time              `db:"updated_at"`
	}

	RiskAssessmentNodes []*RiskAssessmentNode
)

func (n *RiskAssessmentNode) CursorKey(orderBy RiskAssessmentNodeOrderField) page.CursorKey {
	switch orderBy {
	case RiskAssessmentNodeOrderFieldCreatedAt:
		return page.CursorKey{ID: n.ID, Value: n.CreatedAt}
	case RiskAssessmentNodeOrderFieldName:
		return page.CursorKey{ID: n.ID, Value: n.Name}
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (n *RiskAssessmentNode) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM risk_assessment_nodes WHERE id = ANY(@resource_ids::text[])`

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

func (ns *RiskAssessmentNodes) LoadByRiskAssessmentScopeID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	riskAssessmentScopeID gid.GID,
	cursor *page.Cursor[RiskAssessmentNodeOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	risk_assessment_scope_id,
	boundary_id,
	node_type,
	name,
	created_at,
	updated_at
FROM
	risk_assessment_nodes
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
		return fmt.Errorf("cannot query risk assessment nodes: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[RiskAssessmentNode])
	if err != nil {
		return fmt.Errorf("cannot collect risk assessment nodes: %w", err)
	}

	*ns = results

	return nil
}

func (ns *RiskAssessmentNodes) CountByRiskAssessmentScopeID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	riskAssessmentScopeID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	risk_assessment_nodes
WHERE
	%s
	AND risk_assessment_scope_id = @risk_assessment_scope_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.NamedArgs{"risk_assessment_scope_id": riskAssessmentScopeID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count risk assessment nodes: %w", err)
	}

	return count, nil
}

func (n *RiskAssessmentNode) LoadByID(ctx context.Context, conn pg.Querier, scope Scoper, id gid.GID) error {
	q := `
SELECT
	id,
	organization_id,
	risk_assessment_scope_id,
	boundary_id,
	node_type,
	name,
	created_at,
	updated_at
FROM
	risk_assessment_nodes
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
		return fmt.Errorf("cannot query risk assessment node: %w", err)
	}

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[RiskAssessmentNode])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect risk assessment node: %w", err)
	}

	*n = result

	return nil
}

func (n *RiskAssessmentNode) Insert(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
INSERT INTO risk_assessment_nodes (
	id,
	tenant_id,
	organization_id,
	risk_assessment_scope_id,
	boundary_id,
	node_type,
	name,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@risk_assessment_scope_id,
	@boundary_id,
	@node_type,
	@name,
	@created_at,
	@updated_at
)
`
	args := pgx.StrictNamedArgs{
		"id":                       n.ID,
		"tenant_id":                scope.GetTenantID(),
		"organization_id":          n.OrganizationID,
		"risk_assessment_scope_id": n.RiskAssessmentScopeID,
		"boundary_id":              n.BoundaryID,
		"node_type":                n.NodeType,
		"name":                     n.Name,
		"created_at":               n.CreatedAt,
		"updated_at":               n.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" && pgErr.ConstraintName == "risk_assessment_nodes_unique_name" {
			return ErrResourceAlreadyExists
		}

		return fmt.Errorf("cannot insert risk assessment node: %w", err)
	}

	return nil
}

func (n *RiskAssessmentNode) Update(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
UPDATE risk_assessment_nodes
SET
	boundary_id = @boundary_id,
	node_type = @node_type,
	name = @name,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"id":          n.ID,
		"boundary_id": n.BoundaryID,
		"node_type":   n.NodeType,
		"name":        n.Name,
		"updated_at":  n.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update risk assessment node: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (n *RiskAssessmentNode) Delete(ctx context.Context, conn pg.Tx, scope Scoper, id gid.GID) error {
	q := `
DELETE FROM risk_assessment_nodes
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
