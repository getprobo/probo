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
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	Finding struct {
		ID                 gid.GID         `db:"id"`
		OrganizationID     gid.GID         `db:"organization_id"`
		SnapshotID         *gid.GID        `db:"snapshot_id"`
		SourceID           *gid.GID        `db:"source_id"`
		Kind               FindingKind     `db:"kind"`
		ReferenceID        string          `db:"reference_id"`
		Description        *string         `db:"description"`
		Source             *string         `db:"source"`
		IdentifiedOn     *time.Time      `db:"identified_on"`
		RootCause          *string         `db:"root_cause"`
		CorrectiveAction   *string         `db:"corrective_action"`
		OwnerID            *gid.GID        `db:"owner_id"`
		DueDate            *time.Time      `db:"due_date"`
		Status             FindingStatus   `db:"status"`
		Priority           FindingPriority `db:"priority"`
		RiskID             *gid.GID        `db:"risk_id"`
		EffectivenessCheck *string         `db:"effectiveness_check"`
		CreatedAt          time.Time       `db:"created_at"`
		UpdatedAt          time.Time       `db:"updated_at"`
	}

	Findings []*Finding
)

func (f *Finding) CursorKey(field FindingOrderField) page.CursorKey {
	switch field {
	case FindingOrderFieldCreatedAt:
		return page.NewCursorKey(f.ID, f.CreatedAt)
	case FindingOrderFieldIdentifiedOn:
		return page.NewCursorKey(f.ID, f.IdentifiedOn)
	case FindingOrderFieldDueDate:
		return page.NewCursorKey(f.ID, f.DueDate)
	case FindingOrderFieldStatus:
		return page.NewCursorKey(f.ID, f.Status)
	case FindingOrderFieldPriority:
		return page.NewCursorKey(f.ID, f.Priority)
	case FindingOrderFieldReferenceId:
		return page.NewCursorKey(f.ID, f.ReferenceID)
	case FindingOrderFieldKind:
		return page.NewCursorKey(f.ID, f.Kind)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (f *Finding) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM findings WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, f.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query finding authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (f *Finding) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	findingID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	snapshot_id,
	source_id,
	kind,
	reference_id,
	description,
	source,
	identified_on,
	root_cause,
	corrective_action,
	owner_id,
	due_date,
	status,
	priority,
	risk_id,
	effectiveness_check,
	created_at,
	updated_at
FROM
	findings
WHERE
	%s
	AND id = @finding_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"finding_id": findingID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query finding: %w", err)
	}

	finding, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Finding])
	if err != nil {
		return fmt.Errorf("cannot collect finding: %w", err)
	}

	*f = finding

	return nil
}

func (fs *Findings) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	filter *FindingFilter,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	findings
WHERE
	%s
	AND organization_id = @organization_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count findings: %w", err)
	}

	return count, nil
}

func (fs *Findings) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[FindingOrderField],
	filter *FindingFilter,
) error {
	q := `
SELECT
	id,
	organization_id,
	snapshot_id,
	source_id,
	kind,
	reference_id,
	description,
	source,
	identified_on,
	root_cause,
	corrective_action,
	owner_id,
	due_date,
	status,
	priority,
	risk_id,
	effectiveness_check,
	created_at,
	updated_at
FROM
	findings
WHERE
	%s
	AND organization_id = @organization_id
	AND %s
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query findings: %w", err)
	}

	findings, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Finding])
	if err != nil {
		return fmt.Errorf("cannot collect findings: %w", err)
	}

	*fs = findings

	return nil
}

func (f *Finding) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
WITH next_ref AS (
	SELECT pg_advisory_xact_lock(hashtext(@organization_id::text)),
		COALESCE(
			MAX(CAST(SUBSTRING(reference_id FROM 5) AS INTEGER)),
			0
		) + 1 AS next_num
	FROM findings
	WHERE organization_id = @organization_id AND snapshot_id IS NULL
)
INSERT INTO findings (
	id,
	tenant_id,
	organization_id,
	kind,
	reference_id,
	description,
	source,
	identified_on,
	root_cause,
	corrective_action,
	owner_id,
	due_date,
	status,
	priority,
	risk_id,
	effectiveness_check,
	created_at,
	updated_at
)
SELECT
	@id,
	@tenant_id,
	@organization_id,
	@kind,
	'FND-' || LPAD(next_ref.next_num::TEXT, 3, '0'),
	@description,
	@source,
	@identified_on,
	@root_cause,
	@corrective_action,
	@owner_id,
	@due_date,
	@status,
	@priority,
	@risk_id,
	@effectiveness_check,
	@created_at,
	@updated_at
FROM next_ref
RETURNING reference_id
`

	args := pgx.StrictNamedArgs{
		"id":                  f.ID,
		"tenant_id":           scope.GetTenantID(),
		"organization_id":     f.OrganizationID,
		"kind":                f.Kind,
		"description":         f.Description,
		"source":              f.Source,
		"identified_on":       f.IdentifiedOn,
		"root_cause":          f.RootCause,
		"corrective_action":   f.CorrectiveAction,
		"owner_id":    f.OwnerID,
		"due_date":            f.DueDate,
		"status":              f.Status,
		"priority":            f.Priority,
		"risk_id":             f.RiskID,
		"effectiveness_check": f.EffectivenessCheck,
		"created_at":          f.CreatedAt,
		"updated_at":          f.UpdatedAt,
	}

	err := conn.QueryRow(ctx, q, args).Scan(&f.ReferenceID)
	if err != nil {
		return fmt.Errorf("cannot insert finding: %w", err)
	}

	return nil
}

func (f *Finding) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE findings
SET
	description = @description,
	source = @source,
	identified_on = @identified_on,
	root_cause = @root_cause,
	corrective_action = @corrective_action,
	owner_id = @owner_id,
	due_date = @due_date,
	status = @status,
	priority = @priority,
	risk_id = @risk_id,
	effectiveness_check = @effectiveness_check,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
	AND snapshot_id IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                  f.ID,
		"description":         f.Description,
		"source":              f.Source,
		"identified_on":       f.IdentifiedOn,
		"root_cause":          f.RootCause,
		"corrective_action":   f.CorrectiveAction,
		"owner_id":    f.OwnerID,
		"due_date":            f.DueDate,
		"status":              f.Status,
		"priority":            f.Priority,
		"risk_id":             f.RiskID,
		"effectiveness_check": f.EffectivenessCheck,
		"updated_at":          f.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update finding: %w", err)
	}

	return nil
}

func (f *Finding) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM findings
WHERE
	%s
	AND id = @id AND snapshot_id IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": f.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete finding: %w", err)
	}

	return nil
}

func (fs Findings) Snapshot(ctx context.Context, conn pg.Conn, scope Scoper, organizationID, snapshotID gid.GID) error {
	query := `
INSERT INTO findings (
	id,
	tenant_id,
	snapshot_id,
	source_id,
	organization_id,
	kind,
	reference_id,
	description,
	source,
	identified_on,
	root_cause,
	corrective_action,
	owner_id,
	due_date,
	status,
	priority,
	risk_id,
	effectiveness_check,
	created_at,
	updated_at
)
SELECT
	generate_gid(decode_base64_unpadded(@tenant_id), @finding_entity_type),
	@tenant_id,
	@snapshot_id,
	f.id,
	f.organization_id,
	f.kind,
	f.reference_id,
	f.description,
	f.source,
	f.identified_on,
	f.root_cause,
	f.corrective_action,
	f.owner_id,
	f.due_date,
	f.status,
	f.priority,
	f.risk_id,
	f.effectiveness_check,
	f.created_at,
	f.updated_at
FROM findings f
WHERE %s AND f.organization_id = @organization_id AND f.snapshot_id IS NULL
	`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"tenant_id":           scope.GetTenantID(),
		"snapshot_id":         snapshotID,
		"organization_id":     organizationID,
		"finding_entity_type": FindingEntityType,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert finding snapshots: %w", err)
	}

	auditQuery := `
INSERT INTO findings_audits (finding_id, audit_id, reference_id, organization_id, tenant_id, created_at)
SELECT
	snap.id,
	fa.audit_id,
	fa.reference_id,
	fa.organization_id,
	fa.tenant_id,
	fa.created_at
FROM findings_audits fa
JOIN findings live ON fa.finding_id = live.id AND live.snapshot_id IS NULL
JOIN findings snap ON snap.source_id = live.id AND snap.snapshot_id = @snapshot_id
WHERE %s AND live.organization_id = @organization_id
	`

	auditQuery = fmt.Sprintf(auditQuery, scope.SQLFragment())

	_, err = conn.Exec(ctx, auditQuery, args)
	if err != nil {
		return fmt.Errorf("cannot insert finding audit snapshots: %w", err)
	}

	return nil
}

func (fs *Findings) LoadByAuditID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	auditID gid.GID,
	cursor *page.Cursor[FindingOrderField],
	filter *FindingFilter,
) error {
	q := `
WITH f AS (
	SELECT
		fi.id,
		fi.tenant_id,
		fi.organization_id,
		fi.snapshot_id,
		fi.source_id,
		fi.kind,
		fi.reference_id,
		fi.description,
		fi.source,
		fi.identified_on,
		fi.root_cause,
		fi.corrective_action,
		fi.owner_id,
		fi.due_date,
		fi.status,
		fi.priority,
		fi.risk_id,
		fi.effectiveness_check,
		fi.created_at,
		fi.updated_at
	FROM
		findings fi
	INNER JOIN
		findings_audits fa ON fi.id = fa.finding_id
	WHERE
		fa.audit_id = @audit_id
)
SELECT
	id,
	organization_id,
	snapshot_id,
	source_id,
	kind,
	reference_id,
	description,
	source,
	identified_on,
	root_cause,
	corrective_action,
	owner_id,
	due_date,
	status,
	priority,
	risk_id,
	effectiveness_check,
	created_at,
	updated_at
FROM
	f
WHERE %s
	AND %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"audit_id": auditID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query findings: %w", err)
	}

	findings, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Finding])
	if err != nil {
		return fmt.Errorf("cannot collect findings: %w", err)
	}

	*fs = findings

	return nil
}

func (fs *Findings) CountByAuditID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	auditID gid.GID,
	filter *FindingFilter,
) (int, error) {
	q := `
WITH f AS (
	SELECT
		fi.id,
		fi.tenant_id
	FROM
		findings fi
	INNER JOIN
		findings_audits fa ON fi.id = fa.finding_id
	WHERE
		fa.audit_id = @audit_id
)
SELECT
	COUNT(id)
FROM
	f
WHERE
	%s
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"audit_id": auditID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count findings: %w", err)
	}

	return count, nil
}
