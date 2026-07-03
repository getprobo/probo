// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	Measure struct {
		ID             gid.GID      `db:"id"`
		OrganizationID gid.GID      `db:"organization_id"`
		Category       string       `db:"category"`
		Name           string       `db:"name"`
		Description    *string      `db:"description"`
		State          MeasureState `db:"state"`
		ReferenceID    string       `db:"reference_id"`
		CreatedAt      time.Time    `db:"created_at"`
		UpdatedAt      time.Time    `db:"updated_at"`
	}

	Measures []*Measure
)

func (m Measure) CursorKey(orderBy MeasureOrderField) page.CursorKey {
	switch orderBy {
	case MeasureOrderFieldCreatedAt:
		return page.NewCursorKey(m.ID, m.CreatedAt)
	case MeasureOrderFieldName:
		return page.NewCursorKey(m.ID, m.Name)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

// AuthorizationAttributes returns the authorization attributes for policy evaluation.
func (m *Measure) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM measures WHERE id = ANY(@resource_ids::text[])`

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

func (m *Measures) CountByRiskID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	riskID gid.GID,
	filter *MeasureFilter,
) (int, error) {
	q := `
WITH msrs AS (
	SELECT
		m.id,
		m.tenant_id,
		m.search_vector,
		m.state,
		m.category
	FROM
		measures m
	INNER JOIN
		risks_measures rm ON m.id = rm.measure_id
	WHERE
		rm.risk_id = @risk_id
)
SELECT
	COUNT(id)
FROM
	msrs
WHERE %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.NamedArgs{"risk_id": riskID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (m *Measures) LoadByRiskID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	riskID gid.GID,
	cursor *page.Cursor[MeasureOrderField],
	filter *MeasureFilter,
) error {
	q := `
WITH msrs AS (
	SELECT
		m.id,
		m.tenant_id,
		m.organization_id,
		m.category,
		m.name,
		m.description,
		m.state,
		m.reference_id,
		m.created_at,
		m.updated_at,
		m.search_vector
	FROM
		measures m
	INNER JOIN
		risks_measures rm ON m.id = rm.measure_id
	WHERE
		rm.risk_id = @risk_id
)
SELECT
	id,
	organization_id,
	category,
	name,
	description,
	state,
	reference_id,
	created_at,
	updated_at
FROM
	msrs
WHERE %s
	AND %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"risk_id": riskID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query measures: %w", err)
	}

	measures, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Measure])
	if err != nil {
		return fmt.Errorf("cannot collect measures: %w", err)
	}

	*m = measures

	return nil
}

func (m *Measures) CountByControlID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	controlID gid.GID,
	filter *MeasureFilter,
) (int, error) {
	q := `
WITH mtgtns AS (
		SELECT
			m.id,
			m.tenant_id,
			m.search_vector,
			m.state,
			m.category
		FROM
			measures m
		INNER JOIN
			controls_measures cm ON m.id = cm.measure_id
		WHERE
			cm.control_id = @control_id
	)
	SELECT
		COUNT(id)
	FROM
		mtgtns
	WHERE %s
		AND %s
	`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.NamedArgs{"control_id": controlID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (m *Measures) LoadByControlID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	controlID gid.GID,
	cursor *page.Cursor[MeasureOrderField],
	filter *MeasureFilter,
) error {
	q := `
WITH mtgtns AS (
	SELECT
		m.id,
		m.tenant_id,
		m.organization_id,
		m.category,
		m.name,
		m.description,
		m.state,
		m.reference_id,
		m.search_vector,
		m.created_at,
		m.updated_at
	FROM
		measures m
	INNER JOIN
		controls_measures cm ON m.id = cm.measure_id
	WHERE
		cm.control_id = @control_id
)
SELECT
	id,
	organization_id,
	category,
	name,
	description,
	state,
	reference_id,
	created_at,
	updated_at
FROM
	mtgtns
WHERE %s
	AND %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"control_id": controlID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query measures: %w", err)
	}

	measures, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Measure])
	if err != nil {
		return fmt.Errorf("cannot collect measures: %w", err)
	}

	*m = measures

	return nil
}

func (m *Measures) CountByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	filter *MeasureFilter,
) (int, error) {
	q := `
SELECT
    COUNT(id)
FROM
    measures
WHERE
    %s
    AND organization_id = @organization_id
    AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.NamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (m *Measures) LoadDistinctCategoriesByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) ([]string, error) {
	q := `
SELECT DISTINCT
    category
FROM
    measures
WHERE
    %s
    AND organization_id = @organization_id
ORDER BY
    category ASC
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query measure categories: %w", err)
	}

	categories, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("cannot collect measure categories: %w", err)
	}

	return categories, nil
}

func (m *Measures) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[MeasureOrderField],
	filter *MeasureFilter,
) error {
	q := `
SELECT
    id,
    organization_id,
	category,
    name,
    description,
    state,
    reference_id,
    created_at,
    updated_at
FROM
    measures
WHERE
    %s
    AND organization_id = @organization_id
    AND %s
    AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query measures: %w", err)
	}

	measures, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Measure])
	if err != nil {
		return fmt.Errorf("cannot collect measures: %w", err)
	}

	*m = measures

	return nil
}

func (m *Measure) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	measureID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    category,
    name,
    description,
    state,
    reference_id,
    created_at,
    updated_at
FROM
    measures
WHERE
    %s
    AND id = @measure_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"measure_id": measureID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query measures: %w", err)
	}

	measure, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Measure])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect measures: %w", err)
	}

	*m = measure

	return nil
}

func (m *Measures) LoadByIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	measureIDs []gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    category,
    name,
    description,
    state,
    reference_id,
    created_at,
    updated_at
FROM
    measures
WHERE
    %s
    AND id = ANY(@measure_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"measure_ids": measureIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query measures: %w", err)
	}

	measures, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Measure])
	if err != nil {
		return fmt.Errorf("cannot collect measures: %w", err)
	}

	*m = measures

	if len(measures) != len(gid.NewSet(measureIDs...)) {
		return ErrResourceNotFound
	}

	return nil
}

func (m *Measure) Upsert(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
INSERT INTO
    measures (
        tenant_id,
        id,
        organization_id,
		category,
        name,
		state,
        description,
        reference_id,
        created_at,
        updated_at
	)
VALUES (
    @tenant_id,
    @measure_id,
    @organization_id,
	@category,
    @name,
	@state,
    @description,
    @reference_id,
    @created_at,
    @updated_at
)
ON CONFLICT (organization_id, reference_id) DO UPDATE SET
    name = @name,
    description = @description,
    category = @category,
    updated_at = @updated_at
RETURNING
    id,
    organization_id,
	category,
    name,
	state,
    description,
	reference_id,
    created_at,
    updated_at
`

	args := pgx.StrictNamedArgs{
		"tenant_id":       scope.GetTenantID(),
		"measure_id":      m.ID,
		"organization_id": m.OrganizationID,
		"category":        m.Category,
		"name":            m.Name,
		"state":           m.State,
		"description":     m.Description,
		"reference_id":    m.ReferenceID,
		"created_at":      m.CreatedAt,
		"updated_at":      m.UpdatedAt,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query measures: %w", err)
	}

	measure, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Measure])
	if err != nil {
		return fmt.Errorf("cannot collect measures: %w", err)
	}

	*m = measure

	return nil
}

func (m Measure) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    measures (
        tenant_id,
        id,
        organization_id,
		category,
        name,
		state,
        description,
        reference_id,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @measure_id,
    @organization_id,
	@category,
    @name,
	@state,
    @description,
    @reference_id,
    @created_at,
    @updated_at
);
`

	args := pgx.StrictNamedArgs{
		"tenant_id":       scope.GetTenantID(),
		"measure_id":      m.ID,
		"organization_id": m.OrganizationID,
		"category":        m.Category,
		"name":            m.Name,
		"description":     m.Description,
		"reference_id":    m.ReferenceID,
		"created_at":      m.CreatedAt,
		"updated_at":      m.UpdatedAt,
		"state":           m.State,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "mitigations_org_ref_unique" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert measure: %w", err)
	}

	return nil
}

func (m *Measure) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE measures
SET
  name = @name,
  description = @description,
  category = @category,
  state = @state,
  updated_at = @updated_at
WHERE %s
    AND id = @measure_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{
		"measure_id":  m.ID,
		"name":        m.Name,
		"description": m.Description,
		"category":    m.Category,
		"state":       m.State,
		"updated_at":  m.UpdatedAt,
	}

	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)

	return err
}

func (m *Measure) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	measureID gid.GID,
) error {
	q := `
DELETE FROM measures
WHERE %s
    AND id = @measure_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"measure_id": measureID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)

	return err
}

func (m *Measures) CountByThirdPartyID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	thirdPartyID gid.GID,
	filter *MeasureFilter,
) (int, error) {
	q := `
WITH mtgtns AS (
		SELECT
			m.id,
			m.tenant_id,
			m.search_vector,
			m.state,
			m.category
		FROM
			measures m
		INNER JOIN
			measures_third_parties mtp ON m.id = mtp.measure_id
		WHERE
			mtp.third_party_id = @third_party_id
	)
	SELECT
		COUNT(id)
	FROM
		mtgtns
	WHERE %s
		AND %s
	`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.NamedArgs{"third_party_id": thirdPartyID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (m *Measures) LoadByThirdPartyID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	thirdPartyID gid.GID,
	cursor *page.Cursor[MeasureOrderField],
	filter *MeasureFilter,
) error {
	q := `
WITH mtgtns AS (
	SELECT
		m.id,
		m.tenant_id,
		m.organization_id,
		m.category,
		m.name,
		m.description,
		m.state,
		m.reference_id,
		m.search_vector,
		m.created_at,
		m.updated_at
	FROM
		measures m
	INNER JOIN
		measures_third_parties mtp ON m.id = mtp.measure_id
	WHERE
		mtp.third_party_id = @third_party_id
)
SELECT
	id,
	organization_id,
	category,
	name,
	description,
	state,
	reference_id,
	created_at,
	updated_at
FROM
	mtgtns
WHERE %s
	AND %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"third_party_id": thirdPartyID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query measures: %w", err)
	}

	measures, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Measure])
	if err != nil {
		return fmt.Errorf("cannot collect measures: %w", err)
	}

	*m = measures

	return nil
}
