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
	"go.probo.inc/probo/pkg/timespan"
)

type (
	InternalControl struct {
		ID                   gid.GID                             `db:"id"`
		OrganizationID       gid.GID                             `db:"organization_id"`
		Category             string                              `db:"category"`
		Name                 string                              `db:"name"`
		Description          *string                             `db:"description"`
		State                InternalControlState                `db:"state"`
		ReferenceID          string                              `db:"reference_id"`
		Code                 *string                             `db:"code"`
		ControlType          *InternalControlType                `db:"control_type"`
		Nature               *InternalControlNature              `db:"nature"`
		OperatingMode        *InternalControlOperatingMode       `db:"operating_mode"`
		OperatingInterval    *timespan.TimeSpan                  `db:"operating_frequency"`
		OperatingEvent       *string                             `db:"operating_event"`
		EvidenceCadence      *timespan.TimeSpan                  `db:"evidence_cadence"`
		TestingCadence       *timespan.TimeSpan                  `db:"testing_cadence"`
		NextEvidenceDue      *time.Time                          `db:"next_evidence_due"`
		NextTestDue          *time.Time                          `db:"next_test_due"`
		ImplementationStatus InternalControlImplementationStatus `db:"implementation_status"`
		OwnerID              *gid.GID                            `db:"owner_profile_id"`
		ReviewerID           *gid.GID                            `db:"reviewer_profile_id"`
		CreatedAt            time.Time                           `db:"created_at"`
		UpdatedAt            time.Time                           `db:"updated_at"`
	}

	InternalControls []*InternalControl
)

func (m InternalControl) CursorKey(orderBy InternalControlOrderField) page.CursorKey {
	switch orderBy {
	case InternalControlOrderFieldCreatedAt:
		return page.NewCursorKey(m.ID, m.CreatedAt)
	case InternalControlOrderFieldName:
		return page.NewCursorKey(m.ID, m.Name)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

// AuthorizationAttributes returns the authorization attributes for policy evaluation.
// Deleted internal controls fall back to the latest internal control event so as-of nested fields
// can still authorize against the organization they belonged to.
func (m *InternalControl) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `
SELECT DISTINCT ON (id)
	id,
	organization_id
FROM (
	SELECT
		id,
		organization_id,
		0 AS rank
	FROM
		internal_controls
	WHERE
		id = ANY(@resource_ids::text[])
	UNION ALL
	SELECT
		internal_control_id,
		organization_id,
		1 AS rank
	FROM (
		SELECT DISTINCT ON (internal_control_id)
			internal_control_id,
			organization_id
		FROM
			internal_control_events
		WHERE
			internal_control_id = ANY(@resource_ids::text[])
		ORDER BY
			internal_control_id,
			created_at DESC
	) events
) candidates
ORDER BY
	id,
	rank
`

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

func (m *InternalControls) CountByRiskID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	riskID gid.GID,
	filter *InternalControlFilter,
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
		internal_controls m
	INNER JOIN
		risks_internal_controls rm ON m.id = rm.internal_control_id
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

func (m *InternalControls) LoadByRiskID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	riskID gid.GID,
	cursor *page.Cursor[InternalControlOrderField],
	filter *InternalControlFilter,
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
		m.code,
		m.control_type,
		m.nature,
		m.operating_mode,
		m.operating_frequency,
		m.operating_event,
		m.evidence_cadence,
		m.testing_cadence,
		m.next_evidence_due,
		m.next_test_due,
		m.implementation_status,
		m.owner_profile_id,
		m.reviewer_profile_id,
		m.created_at,
		m.updated_at,
		m.search_vector
	FROM
		internal_controls m
	INNER JOIN
		risks_internal_controls rm ON m.id = rm.internal_control_id
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
	code,
	control_type,
	nature,
	operating_mode,
	operating_frequency,
	operating_event,
	evidence_cadence,
	testing_cadence,
	next_evidence_due,
	next_test_due,
	implementation_status,
	owner_profile_id,
	reviewer_profile_id,
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
		return fmt.Errorf("cannot query internalControls: %w", err)
	}

	internalControls, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[InternalControl])
	if err != nil {
		return fmt.Errorf("cannot collect internalControls: %w", err)
	}

	*m = internalControls

	return nil
}

func (m *InternalControls) CountByTreatmentPlanID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	treatmentPlanID gid.GID,
	filter *InternalControlFilter,
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
		internal_controls m
	INNER JOIN
		treatment_plans_internal_controls tpm ON m.id = tpm.internal_control_id
	WHERE
		tpm.treatment_plan_id = @treatment_plan_id
)
SELECT
	COUNT(id)
FROM
	msrs
WHERE %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"treatment_plan_id": treatmentPlanID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (m *InternalControls) LoadByTreatmentPlanID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	treatmentPlanID gid.GID,
	cursor *page.Cursor[InternalControlOrderField],
	filter *InternalControlFilter,
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
		m.code,
		m.control_type,
		m.nature,
		m.operating_mode,
		m.operating_frequency,
		m.operating_event,
		m.evidence_cadence,
		m.testing_cadence,
		m.next_evidence_due,
		m.next_test_due,
		m.implementation_status,
		m.owner_profile_id,
		m.reviewer_profile_id,
		m.created_at,
		m.updated_at,
		m.search_vector
	FROM
		internal_controls m
	INNER JOIN
		treatment_plans_internal_controls tpm ON m.id = tpm.internal_control_id
	WHERE
		tpm.treatment_plan_id = @treatment_plan_id
)
SELECT
	id,
	organization_id,
	category,
	name,
	description,
	state,
	reference_id,
	code,
	control_type,
	nature,
	operating_mode,
	operating_frequency,
	operating_event,
	evidence_cadence,
	testing_cadence,
	next_evidence_due,
	next_test_due,
	implementation_status,
	owner_profile_id,
	reviewer_profile_id,
	created_at,
	updated_at
FROM
	msrs
WHERE %s
	AND %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"treatment_plan_id": treatmentPlanID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query internalControls: %w", err)
	}

	internalControls, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[InternalControl])
	if err != nil {
		return fmt.Errorf("cannot collect internalControls: %w", err)
	}

	*m = internalControls

	return nil
}

func (m *InternalControls) LoadByIDsAsOf(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	internalControlIDs []gid.GID,
	asOf time.Time,
	cursor *page.Cursor[InternalControlOrderField],
	filter *InternalControlFilter,
) error {
	if len(internalControlIDs) == 0 {
		*m = nil
		return nil
	}

	q := `
WITH latest AS (
	SELECT DISTINCT ON (internal_control_id)
		tenant_id,
		organization_id,
		internal_control_id,
		event_type,
		name,
		category,
		state,
		implementation_status,
		internal_control_created_at,
		created_at
	FROM
		internal_control_events
	WHERE
		%s
		AND internal_control_id = ANY(@internal_control_ids)
		AND created_at < @as_of
	ORDER BY
		internal_control_id,
		created_at DESC
),
msrs AS (
	SELECT
		latest.internal_control_id AS id,
		latest.tenant_id,
		latest.organization_id,
		latest.category,
		latest.name,
		CAST(NULL AS text) AS description,
		latest.state,
		'' AS reference_id,
		CAST(NULL AS text) AS code,
		CAST(NULL AS text) AS control_type,
		CAST(NULL AS text) AS nature,
		CAST(NULL AS text) AS operating_mode,
		CAST(NULL AS interval) AS operating_frequency,
		CAST(NULL AS text) AS operating_event,
		CAST(NULL AS interval) AS evidence_cadence,
		CAST(NULL AS interval) AS testing_cadence,
		CAST(NULL AS timestamptz) AS next_evidence_due,
		CAST(NULL AS timestamptz) AS next_test_due,
		COALESCE(
			latest.implementation_status,
			(
				CASE latest.state
					WHEN 'IN_PROGRESS' THEN 'IN_PROGRESS'
					WHEN 'IMPLEMENTED' THEN 'IMPLEMENTED'
					ELSE 'NOT_IMPLEMENTED'
				END
			)::internal_control_implementation_status
		) AS implementation_status,
		CAST(NULL AS text) AS owner_profile_id,
		CAST(NULL AS text) AS reviewer_profile_id,
		latest.internal_control_created_at AS created_at,
		latest.created_at AS updated_at
	FROM
		latest
	WHERE
		latest.event_type <> @deleted
)
SELECT
	id,
	organization_id,
	category,
	name,
	description,
	state,
	reference_id,
	code,
	control_type,
	nature,
	operating_mode,
	operating_frequency,
	operating_event,
	evidence_cadence,
	testing_cadence,
	next_evidence_due,
	next_test_due,
	implementation_status,
	owner_profile_id,
	reviewer_profile_id,
	created_at,
	updated_at
FROM
	msrs
WHERE
	%s
	AND %s
	AND %s
`
	q = fmt.Sprintf(
		q,
		scope.SQLFragment(),
		scope.SQLFragment(),
		filter.EventSQLFragment(),
		cursor.SQLFragment(),
	)

	args := pgx.StrictNamedArgs{
		"internal_control_ids": internalControlIDs,
		"as_of":                asOf,
		"deleted":              InternalControlEventTypeDeleted,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query internal controls as of: %w", err)
	}

	internalControls, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[InternalControl])
	if err != nil {
		return fmt.Errorf("cannot collect internal controls as of: %w", err)
	}

	*m = internalControls

	return nil
}

func (m *InternalControls) CountByIDsAsOf(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	internalControlIDs []gid.GID,
	asOf time.Time,
	filter *InternalControlFilter,
) (int, error) {
	if len(internalControlIDs) == 0 {
		return 0, nil
	}

	q := `
WITH latest AS (
	SELECT DISTINCT ON (internal_control_id)
		tenant_id,
		internal_control_id,
		event_type,
		name,
		category,
		state
	FROM
		internal_control_events
	WHERE
		%s
		AND internal_control_id = ANY(@internal_control_ids)
		AND created_at < @as_of
	ORDER BY
		internal_control_id,
		created_at DESC
),
msrs AS (
	SELECT
		latest.internal_control_id AS id,
		latest.tenant_id,
		latest.name,
		latest.category,
		latest.state
	FROM
		latest
	WHERE
		latest.event_type <> @deleted
)
SELECT
	COUNT(id)
FROM
	msrs
WHERE
	%s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), scope.SQLFragment(), filter.EventSQLFragment())

	args := pgx.StrictNamedArgs{
		"internal_control_ids": internalControlIDs,
		"as_of":                asOf,
		"deleted":              InternalControlEventTypeDeleted,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan internal controls as of count: %w", err)
	}

	return count, nil
}

func (m *InternalControls) CountByControlID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	controlID gid.GID,
	filter *InternalControlFilter,
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
			internal_controls m
		INNER JOIN
			controls_internal_controls cm ON m.id = cm.internal_control_id
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

func (m *InternalControls) LoadByControlID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	controlID gid.GID,
	cursor *page.Cursor[InternalControlOrderField],
	filter *InternalControlFilter,
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
		m.code,
		m.control_type,
		m.nature,
		m.operating_mode,
		m.operating_frequency,
		m.operating_event,
		m.evidence_cadence,
		m.testing_cadence,
		m.next_evidence_due,
		m.next_test_due,
		m.implementation_status,
		m.owner_profile_id,
		m.reviewer_profile_id,
		m.search_vector,
		m.created_at,
		m.updated_at
	FROM
		internal_controls m
	INNER JOIN
		controls_internal_controls cm ON m.id = cm.internal_control_id
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
	code,
	control_type,
	nature,
	operating_mode,
	operating_frequency,
	operating_event,
	evidence_cadence,
	testing_cadence,
	next_evidence_due,
	next_test_due,
	implementation_status,
	owner_profile_id,
	reviewer_profile_id,
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
		return fmt.Errorf("cannot query internalControls: %w", err)
	}

	internalControls, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[InternalControl])
	if err != nil {
		return fmt.Errorf("cannot collect internalControls: %w", err)
	}

	*m = internalControls

	return nil
}

func (m *InternalControls) CountByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	filter *InternalControlFilter,
) (int, error) {
	q := `
SELECT
    COUNT(id)
FROM
    internal_controls
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

func (m *InternalControls) LoadDistinctCategoriesByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) ([]string, error) {
	q := `
SELECT DISTINCT
    category
FROM
    internal_controls
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
		return nil, fmt.Errorf("cannot query internal control categories: %w", err)
	}

	categories, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("cannot collect internal control categories: %w", err)
	}

	return categories, nil
}

func (m *InternalControls) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[InternalControlOrderField],
	filter *InternalControlFilter,
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
    code,
    control_type,
    nature,
    operating_mode,
    operating_frequency,
    operating_event,
    evidence_cadence,
    testing_cadence,
    next_evidence_due,
    next_test_due,
    implementation_status,
    owner_profile_id,
    reviewer_profile_id,
    created_at,
    updated_at
FROM
    internal_controls
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
		return fmt.Errorf("cannot query internalControls: %w", err)
	}

	internalControls, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[InternalControl])
	if err != nil {
		return fmt.Errorf("cannot collect internalControls: %w", err)
	}

	*m = internalControls

	return nil
}

func (m *InternalControl) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	internalControlID gid.GID,
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
    code,
    control_type,
    nature,
    operating_mode,
    operating_frequency,
    operating_event,
    evidence_cadence,
    testing_cadence,
    next_evidence_due,
    next_test_due,
    implementation_status,
    owner_profile_id,
    reviewer_profile_id,
    created_at,
    updated_at
FROM
    internal_controls
WHERE
    %s
    AND id = @internal_control_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"internal_control_id": internalControlID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query internalControls: %w", err)
	}

	internalControl, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[InternalControl])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect internalControls: %w", err)
	}

	*m = internalControl

	return nil
}

func (m *InternalControls) LoadByIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	internalControlIDs []gid.GID,
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
    code,
    control_type,
    nature,
    operating_mode,
    operating_frequency,
    operating_event,
    evidence_cadence,
    testing_cadence,
    next_evidence_due,
    next_test_due,
    implementation_status,
    owner_profile_id,
    reviewer_profile_id,
    created_at,
    updated_at
FROM
    internal_controls
WHERE
    %s
    AND id = ANY(@internal_control_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"internal_control_ids": internalControlIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query internalControls: %w", err)
	}

	internalControls, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[InternalControl])
	if err != nil {
		return fmt.Errorf("cannot collect internalControls: %w", err)
	}

	*m = internalControls

	if len(internalControls) != len(gid.NewSet(internalControlIDs...)) {
		return ErrResourceNotFound
	}

	return nil
}

func (m *InternalControl) Upsert(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
INSERT INTO
    internal_controls (
        tenant_id,
        id,
        organization_id,
		category,
        name,
		state,
        description,
        reference_id,
        code,
        control_type,
        nature,
        operating_mode,
        operating_frequency,
        operating_event,
        evidence_cadence,
        testing_cadence,
        next_evidence_due,
        next_test_due,
        implementation_status,
        owner_profile_id,
        reviewer_profile_id,
        created_at,
        updated_at
	)
VALUES (
    @tenant_id,
    @internal_control_id,
    @organization_id,
	@category,
    @name,
	@state,
    @description,
    @reference_id,
    @code,
    @control_type,
    @nature,
    @operating_mode,
    @operating_frequency,
    @operating_event,
    @evidence_cadence,
    @testing_cadence,
    @next_evidence_due,
    @next_test_due,
    @implementation_status,
    @owner_profile_id,
    @reviewer_profile_id,
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
    code,
    control_type,
    nature,
    operating_mode,
    operating_frequency,
    operating_event,
    evidence_cadence,
    testing_cadence,
    next_evidence_due,
    next_test_due,
    implementation_status,
    owner_profile_id,
    reviewer_profile_id,
    created_at,
    updated_at
`

	args := pgx.StrictNamedArgs{
		"tenant_id":             scope.GetTenantID(),
		"internal_control_id":   m.ID,
		"organization_id":       m.OrganizationID,
		"category":              m.Category,
		"name":                  m.Name,
		"state":                 m.State,
		"description":           m.Description,
		"reference_id":          m.ReferenceID,
		"code":                  m.Code,
		"control_type":          m.ControlType,
		"nature":                m.Nature,
		"operating_mode":        m.OperatingMode,
		"operating_frequency":   m.OperatingInterval,
		"operating_event":       m.OperatingEvent,
		"evidence_cadence":      m.EvidenceCadence,
		"testing_cadence":       m.TestingCadence,
		"next_evidence_due":     m.NextEvidenceDue,
		"next_test_due":         m.NextTestDue,
		"implementation_status": m.ImplementationStatus,
		"owner_profile_id":      m.OwnerID,
		"reviewer_profile_id":   m.ReviewerID,
		"created_at":            m.CreatedAt,
		"updated_at":            m.UpdatedAt,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" &&
				(pgErr.ConstraintName == "internal_controls_organization_id_reference_id_key" ||
					pgErr.ConstraintName == "internal_controls_organization_id_code_key") {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot query internalControls: %w", err)
	}

	internalControl, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[InternalControl])
	if err != nil {
		return fmt.Errorf("cannot collect internalControls: %w", err)
	}

	*m = internalControl

	return nil
}

func (m InternalControl) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    internal_controls (
        tenant_id,
        id,
        organization_id,
		category,
        name,
		state,
        description,
        reference_id,
        code,
        control_type,
        nature,
        operating_mode,
        operating_frequency,
        operating_event,
        evidence_cadence,
        testing_cadence,
        next_evidence_due,
        next_test_due,
        implementation_status,
        owner_profile_id,
        reviewer_profile_id,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @internal_control_id,
    @organization_id,
	@category,
    @name,
	@state,
    @description,
    @reference_id,
    @code,
    @control_type,
    @nature,
    @operating_mode,
    @operating_frequency,
    @operating_event,
    @evidence_cadence,
    @testing_cadence,
    @next_evidence_due,
    @next_test_due,
    @implementation_status,
    @owner_profile_id,
    @reviewer_profile_id,
    @created_at,
    @updated_at
);
`

	args := pgx.StrictNamedArgs{
		"tenant_id":             scope.GetTenantID(),
		"internal_control_id":   m.ID,
		"organization_id":       m.OrganizationID,
		"category":              m.Category,
		"name":                  m.Name,
		"description":           m.Description,
		"reference_id":          m.ReferenceID,
		"code":                  m.Code,
		"control_type":          m.ControlType,
		"nature":                m.Nature,
		"operating_mode":        m.OperatingMode,
		"operating_frequency":   m.OperatingInterval,
		"operating_event":       m.OperatingEvent,
		"evidence_cadence":      m.EvidenceCadence,
		"testing_cadence":       m.TestingCadence,
		"next_evidence_due":     m.NextEvidenceDue,
		"next_test_due":         m.NextTestDue,
		"implementation_status": m.ImplementationStatus,
		"owner_profile_id":      m.OwnerID,
		"reviewer_profile_id":   m.ReviewerID,
		"created_at":            m.CreatedAt,
		"updated_at":            m.UpdatedAt,
		"state":                 m.State,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" &&
				(pgErr.ConstraintName == "internal_controls_organization_id_reference_id_key" ||
					pgErr.ConstraintName == "internal_controls_organization_id_code_key") {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert internalControl: %w", err)
	}

	return nil
}

func (m *InternalControl) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE internal_controls
SET
  name = @name,
  description = @description,
  category = @category,
  state = @state,
  code = @code,
  control_type = @control_type,
  nature = @nature,
  operating_mode = @operating_mode,
  operating_frequency = @operating_frequency,
  operating_event = @operating_event,
  evidence_cadence = @evidence_cadence,
  testing_cadence = @testing_cadence,
  next_evidence_due = @next_evidence_due,
  next_test_due = @next_test_due,
  implementation_status = @implementation_status,
  owner_profile_id = @owner_profile_id,
  reviewer_profile_id = @reviewer_profile_id,
  updated_at = @updated_at
WHERE %s
    AND id = @internal_control_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{
		"internal_control_id":   m.ID,
		"name":                  m.Name,
		"description":           m.Description,
		"category":              m.Category,
		"state":                 m.State,
		"code":                  m.Code,
		"control_type":          m.ControlType,
		"nature":                m.Nature,
		"operating_mode":        m.OperatingMode,
		"operating_frequency":   m.OperatingInterval,
		"operating_event":       m.OperatingEvent,
		"evidence_cadence":      m.EvidenceCadence,
		"testing_cadence":       m.TestingCadence,
		"next_evidence_due":     m.NextEvidenceDue,
		"next_test_due":         m.NextTestDue,
		"implementation_status": m.ImplementationStatus,
		"owner_profile_id":      m.OwnerID,
		"reviewer_profile_id":   m.ReviewerID,
		"updated_at":            m.UpdatedAt,
	}

	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "internal_controls_organization_id_code_key" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot update internalControl: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (m *InternalControl) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	internalControlID gid.GID,
) error {
	q := `
DELETE FROM internal_controls
WHERE %s
    AND id = @internal_control_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"internal_control_id": internalControlID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)

	return err
}

func (m *InternalControls) CountByThirdPartyID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	thirdPartyID gid.GID,
	filter *InternalControlFilter,
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
			internal_controls m
		INNER JOIN
			internal_controls_third_parties mtp ON m.id = mtp.internal_control_id
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

func (m *InternalControls) LoadByThirdPartyID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	thirdPartyID gid.GID,
	cursor *page.Cursor[InternalControlOrderField],
	filter *InternalControlFilter,
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
		m.code,
		m.control_type,
		m.nature,
		m.operating_mode,
		m.operating_frequency,
		m.operating_event,
		m.evidence_cadence,
		m.testing_cadence,
		m.next_evidence_due,
		m.next_test_due,
		m.implementation_status,
		m.owner_profile_id,
		m.reviewer_profile_id,
		m.search_vector,
		m.created_at,
		m.updated_at
	FROM
		internal_controls m
	INNER JOIN
		internal_controls_third_parties mtp ON m.id = mtp.internal_control_id
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
	code,
	control_type,
	nature,
	operating_mode,
	operating_frequency,
	operating_event,
	evidence_cadence,
	testing_cadence,
	next_evidence_due,
	next_test_due,
	implementation_status,
	owner_profile_id,
	reviewer_profile_id,
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
		return fmt.Errorf("cannot query internalControls: %w", err)
	}

	internalControls, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[InternalControl])
	if err != nil {
		return fmt.Errorf("cannot collect internalControls: %w", err)
	}

	*m = internalControls

	return nil
}
