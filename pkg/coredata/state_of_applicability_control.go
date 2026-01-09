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
)

type (
	StateOfApplicabilityControl struct {
		StateOfApplicabilityID gid.GID                          `db:"state_of_applicability_id"`
		ControlID              gid.GID                          `db:"control_id"`
		SnapshotID             *gid.GID                         `db:"snapshot_id"`
		State                  StateOfApplicabilityControlState `db:"state"`
		ExclusionJustification *string                          `db:"exclusion_justification"`
		CreatedAt              time.Time                        `db:"created_at"`
	}

	StateOfApplicabilityControls []*StateOfApplicabilityControl

	AvailableControlForStateOfApplicability struct {
		ControlID              gid.GID                           `db:"control_id"`
		SectionTitle           string                            `db:"section_title"`
		Name                   string                            `db:"name"`
		FrameworkID            gid.GID                           `db:"framework_id"`
		FrameworkName          string                            `db:"framework_name"`
		OrganizationID         gid.GID                           `db:"organization_id"`
		StateOfApplicabilityID *gid.GID                          `db:"state_of_applicability_id"`
		State                  *StateOfApplicabilityControlState `db:"state"`
		ExclusionJustification *string                           `db:"exclusion_justification"`
	}

	AvailableControlsForStateOfApplicability []*AvailableControlForStateOfApplicability

	ErrStateOfApplicabilityControlNotFound struct {
		StateOfApplicabilityID gid.GID
		ControlID              gid.GID
	}

	ErrStateOfApplicabilityControlAlreadyExists struct {
		StateOfApplicabilityID gid.GID
		ControlID              gid.GID
	}
)

func (e ErrStateOfApplicabilityControlNotFound) Error() string {
	return fmt.Sprintf("state of applicability control not found: state_of_applicability_id=%s, control_id=%s", e.StateOfApplicabilityID, e.ControlID)
}

func (e ErrStateOfApplicabilityControlAlreadyExists) Error() string {
	return fmt.Sprintf("state of applicability control already exists: state_of_applicability_id=%s, control_id=%s", e.StateOfApplicabilityID, e.ControlID)
}

func (sac *StateOfApplicabilityControl) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    states_of_applicability_controls (
        state_of_applicability_id,
        control_id,
        tenant_id,
        snapshot_id,
        state,
        exclusion_justification,
        created_at
    )
VALUES (
    @state_of_applicability_id,
    @control_id,
    @tenant_id,
    @snapshot_id,
    @state,
    @exclusion_justification,
    @created_at
);
`

	args := pgx.StrictNamedArgs{
		"state_of_applicability_id": sac.StateOfApplicabilityID,
		"control_id":                sac.ControlID,
		"tenant_id":                 scope.GetTenantID(),
		"snapshot_id":               sac.SnapshotID,
		"state":                     sac.State,
		"exclusion_justification":   sac.ExclusionJustification,
		"created_at":                sac.CreatedAt,
	}
	_, err := conn.Exec(ctx, q, args)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return &ErrStateOfApplicabilityControlAlreadyExists{
					StateOfApplicabilityID: sac.StateOfApplicabilityID,
					ControlID:              sac.ControlID,
				}
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
    state = @state,
    exclusion_justification = @exclusion_justification
WHERE %s
    AND state_of_applicability_id = @state_of_applicability_id
    AND control_id = @control_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"state_of_applicability_id": sac.StateOfApplicabilityID,
		"control_id":                sac.ControlID,
		"state":                     sac.State,
		"exclusion_justification":   sac.ExclusionJustification,
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
        state_of_applicability_id,
        control_id,
        tenant_id,
        snapshot_id,
        state,
        exclusion_justification,
        created_at
    )
VALUES (
    @state_of_applicability_id,
    @control_id,
    @tenant_id,
    @snapshot_id,
    @state,
    @exclusion_justification,
    @created_at
)
ON CONFLICT (state_of_applicability_id, control_id) DO UPDATE SET
    state = EXCLUDED.state,
    exclusion_justification = EXCLUDED.exclusion_justification
`

	args := pgx.StrictNamedArgs{
		"state_of_applicability_id": sac.StateOfApplicabilityID,
		"control_id":                sac.ControlID,
		"tenant_id":                 scope.GetTenantID(),
		"snapshot_id":               sac.SnapshotID,
		"state":                     sac.State,
		"exclusion_justification":   sac.ExclusionJustification,
		"created_at":                sac.CreatedAt,
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
DELETE
FROM
    states_of_applicability_controls
WHERE
    %s
    AND state_of_applicability_id = @state_of_applicability_id
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
    state_of_applicability_id,
    control_id,
    snapshot_id,
    state,
    exclusion_justification,
    created_at
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

func (acfs *AvailableControlsForStateOfApplicability) LoadAvailableByStateOfApplicabilityID(
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
        c.tenant_id AS control_tenant_id
    FROM controls c
    WHERE %s
),
all_controls AS (
    SELECT
        fc.control_id,
        fc.section_title,
        fc.name,
        fc.framework_id,
        fc.organization_id,
        f.name AS framework_name
    FROM filtered_controls fc
    INNER JOIN frameworks f ON fc.framework_id = f.id
    CROSS JOIN soa_info si
    WHERE fc.organization_id = si.organization_id
),
existing_links AS (
    SELECT
        soac.control_id,
        soac.state_of_applicability_id,
        soac.state,
        soac.exclusion_justification
    FROM states_of_applicability_controls soac
    CROSS JOIN soa_info si
    WHERE soac.tenant_id = si.soa_tenant_id
        AND soac.state_of_applicability_id = @state_of_applicability_id
)
SELECT
    ac.control_id,
    ac.section_title,
    ac.name,
    ac.framework_id,
    ac.organization_id,
    ac.framework_name,
    el.state_of_applicability_id,
    el.state,
    el.exclusion_justification
FROM all_controls ac
LEFT JOIN existing_links el ON ac.control_id = el.control_id
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

	controls, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AvailableControlForStateOfApplicability])
	if err != nil {
		return fmt.Errorf("cannot collect available controls: %w", err)
	}

	*acfs = controls
	return nil
}
