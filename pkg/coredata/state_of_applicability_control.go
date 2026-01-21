// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
	StateOfApplicabilityControl struct {
		ID                     gid.GID   `db:"id"`
		StateOfApplicabilityID gid.GID   `db:"state_of_applicability_id"`
		ControlID              gid.GID   `db:"control_id"`
		OrganizationID         gid.GID   `db:"organization_id"`
		SnapshotID             *gid.GID  `db:"snapshot_id"`
		Applicability          bool      `db:"applicability"`
		Justification          *string   `db:"justification"`
		CreatedAt              time.Time `db:"created_at"`
		UpdatedAt              time.Time `db:"updated_at"`
	}

	StateOfApplicabilityControls []*StateOfApplicabilityControl

	AvailableStateOfApplicabilityControl struct {
		ControlID              gid.GID  `db:"control_id"`
		SectionTitle           string   `db:"section_title"`
		Name                   string   `db:"name"`
		FrameworkID            gid.GID  `db:"framework_id"`
		FrameworkName          string   `db:"framework_name"`
		OrganizationID         gid.GID  `db:"organization_id"`
		StateOfApplicabilityID *gid.GID `db:"state_of_applicability_id"`
		Applicability          *bool    `db:"applicability"`
		Justification          *string  `db:"justification"`
		BestPractice           bool     `db:"best_practice"`
		Regulatory             bool     `db:"regulatory"`
		Contractual            bool     `db:"contractual"`
		RiskAssessment         bool     `db:"risk_assessment"`
	}

	AvailableStateOfApplicabilityControls []*AvailableStateOfApplicabilityControl
)

func (s StateOfApplicabilityControl) CursorKey(orderBy StateOfApplicabilityOrderField) page.CursorKey {
	switch orderBy {
	case StateOfApplicabilityOrderFieldName:
		return page.NewCursorKey(s.ID, s.StateOfApplicabilityID)
	case StateOfApplicabilityOrderFieldCreatedAt:
		return page.NewCursorKey(s.ID, s.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (s *StateOfApplicabilityControl) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM states_of_applicability_controls WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, s.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query state of applicability control authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (sac *StateOfApplicabilityControl) LoadByStateOfApplicabilityIDAndControlID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	stateOfApplicabilityID gid.GID,
	controlID gid.GID,
) error {
	q := `
WITH current_soa AS (
    SELECT id
    FROM states_of_applicability
    WHERE
        %s
        AND id = @state_of_applicability_id
        AND snapshot_id IS NULL
)
SELECT
    soac.id,
    soac.state_of_applicability_id,
    soac.control_id,
    soac.organization_id,
    soac.snapshot_id,
    soac.applicability,
    soac.justification,
    soac.created_at,
    soac.updated_at
FROM
    states_of_applicability_controls soac
INNER JOIN
    current_soa ON soac.state_of_applicability_id = current_soa.id
WHERE
    soac.control_id = @control_id
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"state_of_applicability_id": stateOfApplicabilityID,
		"control_id":                controlID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query states_of_applicability_controls: %w", err)
	}

	control, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[StateOfApplicabilityControl])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect state of applicability control: %w", err)
	}

	*sac = control
	return nil
}

func (sac *StateOfApplicabilityControl) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    states_of_applicability_controls (
        id,
        state_of_applicability_id,
        control_id,
        organization_id,
        tenant_id,
        snapshot_id,
        applicability,
        justification,
        created_at,
        updated_at
    )
VALUES (
    @id,
    @state_of_applicability_id,
    @control_id,
    @organization_id,
    @tenant_id,
    @snapshot_id,
    @applicability,
    @justification,
    @created_at,
    @updated_at
);
`

	args := pgx.StrictNamedArgs{
		"id":                        sac.ID,
		"state_of_applicability_id": sac.StateOfApplicabilityID,
		"control_id":                sac.ControlID,
		"organization_id":           sac.OrganizationID,
		"tenant_id":                 scope.GetTenantID(),
		"snapshot_id":               sac.SnapshotID,
		"applicability":             sac.Applicability,
		"justification":             sac.Justification,
		"created_at":                sac.CreatedAt,
		"updated_at":                sac.UpdatedAt,
	}
	_, err := conn.Exec(ctx, q, args)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert state_of_applicability_control: %w", err)
	}

	return nil
}

func (sac *StateOfApplicabilityControl) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE states_of_applicability_controls
SET
    applicability = @applicability,
    justification = @justification,
    updated_at = @updated_at
WHERE
    %s
    AND state_of_applicability_id = @state_of_applicability_id
    AND control_id = @control_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"state_of_applicability_id": sac.StateOfApplicabilityID,
		"control_id":                sac.ControlID,
		"applicability":             sac.Applicability,
		"justification":             sac.Justification,
		"updated_at":                sac.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update state_of_applicability_control: %w", err)
	}

	return nil
}

func (sac *StateOfApplicabilityControl) Upsert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    states_of_applicability_controls (
        id,
        state_of_applicability_id,
        control_id,
        organization_id,
        tenant_id,
        snapshot_id,
        applicability,
        justification,
        created_at,
        updated_at
    )
VALUES (
    @id,
    @state_of_applicability_id,
    @control_id,
    @organization_id,
    @tenant_id,
    @snapshot_id,
    @applicability,
    @justification,
    @created_at,
    @updated_at
)
ON CONFLICT (state_of_applicability_id, control_id) DO UPDATE SET
    applicability = EXCLUDED.applicability,
    justification = EXCLUDED.justification,
    updated_at = EXCLUDED.updated_at
`

	args := pgx.StrictNamedArgs{
		"id":                        sac.ID,
		"state_of_applicability_id": sac.StateOfApplicabilityID,
		"control_id":                sac.ControlID,
		"organization_id":           sac.OrganizationID,
		"tenant_id":                 scope.GetTenantID(),
		"snapshot_id":               sac.SnapshotID,
		"applicability":             sac.Applicability,
		"justification":             sac.Justification,
		"created_at":                sac.CreatedAt,
		"updated_at":                sac.UpdatedAt,
	}
	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot upsert state_of_applicability_control: %w", err)
	}

	return nil
}

func (sac *StateOfApplicabilityControl) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
WITH current_soa AS (
    SELECT id
    FROM states_of_applicability
    WHERE
        %s
        AND id = @state_of_applicability_id
        AND snapshot_id IS NULL
)
DELETE FROM states_of_applicability_controls
WHERE state_of_applicability_id IN (SELECT id FROM current_soa)
    AND control_id = @control_id;
`

	args := pgx.StrictNamedArgs{
		"state_of_applicability_id": sac.StateOfApplicabilityID,
		"control_id":                sac.ControlID,
	}
	maps.Copy(args, scope.SQLArguments())
	q = fmt.Sprintf(q, scope.SQLFragment())

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (sacs *StateOfApplicabilityControls) LoadByStateOfApplicabilityID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	stateOfApplicabilityID gid.GID,
) error {
	q := `
SELECT
    id,
    state_of_applicability_id,
    control_id,
    organization_id,
    snapshot_id,
    applicability,
    justification,
    created_at,
    updated_at
FROM
    states_of_applicability_controls
WHERE
    %s
    AND state_of_applicability_id = @state_of_applicability_id
ORDER BY created_at ASC
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"state_of_applicability_id": stateOfApplicabilityID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query states_of_applicability_controls: %w", err)
	}

	controls, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[StateOfApplicabilityControl])
	if err != nil {
		return fmt.Errorf("cannot collect states_of_applicability_controls: %w", err)
	}

	*sacs = controls
	return nil
}

func (sacs *StateOfApplicabilityControls) LoadByControlID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	controlID gid.GID,
	cursor *page.Cursor[StateOfApplicabilityOrderField],
) error {
	q := `
WITH soac_ctrl AS (
    SELECT
        soac.id,
        soac.state_of_applicability_id,
        soac.control_id,
        soac.organization_id,
        soac.snapshot_id,
        soac.applicability,
        soac.justification,
        soac.created_at,
        soac.updated_at,
        soac.tenant_id
    FROM
        states_of_applicability_controls soac
    INNER JOIN
        states_of_applicability soa ON soac.state_of_applicability_id = soa.id
    WHERE
        soac.%[1]s
        AND soac.control_id = @control_id
        AND soa.snapshot_id IS NULL
)
SELECT
    id,
    state_of_applicability_id,
    control_id,
    organization_id,
    snapshot_id,
    applicability,
    justification,
    created_at,
    updated_at
FROM
    soac_ctrl
WHERE
    %[1]s
    AND %[2]s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"control_id": controlID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query state_of_applicability_controls: %w", err)
	}

	controls, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[StateOfApplicabilityControl])
	if err != nil {
		return fmt.Errorf("cannot collect state_of_applicability_controls: %w", err)
	}

	*sacs = controls
	return nil
}

func (sacs *StateOfApplicabilityControls) CountByControlID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	controlID gid.GID,
) (int, error) {
	q := `
WITH soac_ctrl AS (
    SELECT
        soac.id,
        soac.organization_id,
        soac.tenant_id
    FROM
        states_of_applicability_controls soac
    INNER JOIN
        states_of_applicability soa ON soac.state_of_applicability_id = soa.id
    WHERE
        soac.%[1]s
        AND soac.control_id = @control_id
        AND soa.snapshot_id IS NULL
)
SELECT
    COUNT(id)
FROM
    soac_ctrl
WHERE
    %[1]s;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"control_id": controlID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (acfs *AvailableStateOfApplicabilityControls) LoadAvailableByStateOfApplicabilityID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	stateOfApplicabilityID gid.GID,
) error {
	q := `
WITH soa_info AS (
    SELECT
        soa.organization_id,
        soa.tenant_id AS soa_tenant_id
    FROM states_of_applicability soa
    WHERE soa.tenant_id = @tenant_id
        AND soa.id = @state_of_applicability_id
),
filtered_controls AS (
    SELECT
        c.id AS control_id,
        c.section_title,
        c.name,
        c.framework_id,
        c.organization_id,
        c.tenant_id,
        c.best_practice
    FROM controls c
    WHERE
        %s
),
all_controls AS (
    SELECT
        fc.control_id,
        fc.section_title,
        fc.name,
        fc.framework_id,
        fc.organization_id,
        fc.tenant_id,
        f.name AS framework_name,
        fc.best_practice
    FROM filtered_controls fc
    INNER JOIN frameworks f ON fc.framework_id = f.id
    CROSS JOIN soa_info si
    WHERE fc.organization_id = si.organization_id
),
existing_links AS (
    SELECT
        soac.control_id,
        soac.state_of_applicability_id,
        soac.applicability,
        soac.justification
    FROM states_of_applicability_controls soac
    CROSS JOIN soa_info si
    WHERE soac.tenant_id = si.soa_tenant_id
        AND soac.state_of_applicability_id = @state_of_applicability_id
),
regulatory_controls AS (
    SELECT DISTINCT co.control_id
    FROM controls_obligations co
    INNER JOIN obligations o ON o.id = co.obligation_id
    CROSS JOIN soa_info si
    WHERE co.tenant_id = si.soa_tenant_id
        AND o.tenant_id = si.soa_tenant_id
        AND o.type = 'LEGAL'
),
contractual_controls AS (
    SELECT DISTINCT co.control_id
    FROM controls_obligations co
    INNER JOIN obligations o ON o.id = co.obligation_id
    CROSS JOIN soa_info si
    WHERE co.tenant_id = si.soa_tenant_id
        AND o.tenant_id = si.soa_tenant_id
        AND o.type = 'CONTRACTUAL'
),
risk_controls AS (
    SELECT DISTINCT cm.control_id
    FROM controls_measures cm
    INNER JOIN risks_measures rm ON rm.measure_id = cm.measure_id
    CROSS JOIN soa_info si
    WHERE cm.tenant_id = si.soa_tenant_id
        AND rm.tenant_id = si.soa_tenant_id
)
SELECT
    ac.control_id,
    ac.section_title,
    ac.name,
    ac.framework_id,
    ac.organization_id,
    ac.framework_name,
    el.state_of_applicability_id,
    el.applicability,
    el.justification,
    ac.best_practice,
    CASE WHEN reg.control_id IS NOT NULL THEN TRUE ELSE FALSE END AS regulatory,
    CASE WHEN cont.control_id IS NOT NULL THEN TRUE ELSE FALSE END AS contractual,
    CASE WHEN risk.control_id IS NOT NULL THEN TRUE ELSE FALSE END AS risk_assessment
FROM all_controls ac
LEFT JOIN existing_links el ON ac.control_id = el.control_id
LEFT JOIN regulatory_controls reg ON reg.control_id = ac.control_id
LEFT JOIN contractual_controls cont ON cont.control_id = ac.control_id
LEFT JOIN risk_controls risk ON risk.control_id = ac.control_id
ORDER BY ac.framework_name, ac.section_title, ac.name
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"state_of_applicability_id": stateOfApplicabilityID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query available controls: %w", err)
	}

	controls, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AvailableStateOfApplicabilityControl])
	if err != nil {
		return fmt.Errorf("cannot collect available controls: %w", err)
	}

	*acfs = controls
	return nil
}
