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
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	ContinualImprovement struct {
		ID             gid.GID                      `db:"id"`
		OrganizationID gid.GID                      `db:"organization_id"`
		ReferenceID    string                       `db:"reference_id"`
		Description    *string                      `db:"description"`
		Source         *string                      `db:"source"`
		OwnerID        gid.GID                      `db:"owner_profile_id"`
		TargetDate     *time.Time                   `db:"target_date"`
		Status         ContinualImprovementStatus   `db:"status"`
		Priority       ContinualImprovementPriority `db:"priority"`
		SnapshotID     *gid.GID                     `db:"snapshot_id"`
		SourceID       *gid.GID                     `db:"source_id"`
		CreatedAt      time.Time                    `db:"created_at"`
		UpdatedAt      time.Time                    `db:"updated_at"`
	}

	ContinualImprovements []*ContinualImprovement
)

func (ci *ContinualImprovement) CursorKey(field ContinualImprovementOrderField) page.CursorKey {
	switch field {
	case ContinualImprovementOrderFieldCreatedAt:
		return page.NewCursorKey(ci.ID, ci.CreatedAt)
	case ContinualImprovementOrderFieldTargetDate:
		return page.NewCursorKey(ci.ID, ci.TargetDate)
	case ContinualImprovementOrderFieldStatus:
		return page.NewCursorKey(ci.ID, ci.Status)
	case ContinualImprovementOrderFieldPriority:
		return page.NewCursorKey(ci.ID, ci.Priority)
	case ContinualImprovementOrderFieldReferenceId:
		return page.NewCursorKey(ci.ID, ci.ReferenceID)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

// AuthorizationAttributes returns the authorization attributes for policy evaluation.
func (ci *ContinualImprovement) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM continual_improvements WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, ci.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query continual improvement authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (ci *ContinualImprovement) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	continualImprovementID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	reference_id,
	description,
	source,
	owner_profile_id,
	target_date,
	status,
	priority,
	snapshot_id,
	source_id,
	created_at,
	updated_at
FROM
	continual_improvements
WHERE
	%s
	AND id = @continual_improvement_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"continual_improvement_id": continualImprovementID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query continual improvement: %w", err)
	}

	improvement, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ContinualImprovement])
	if err != nil {
		return fmt.Errorf("cannot collect continual improvement: %w", err)
	}

	*ci = improvement

	return nil
}

func (cis *ContinualImprovements) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	filter *ContinualImprovementFilter,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	continual_improvements
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
		return 0, fmt.Errorf("cannot count continual improvements: %w", err)
	}

	return count, nil
}

func (cis *ContinualImprovements) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[ContinualImprovementOrderField],
	filter *ContinualImprovementFilter,
) error {
	q := `
SELECT
	id,
	organization_id,
	reference_id,
	description,
	source,
	owner_profile_id,
	target_date,
	status,
	priority,
	snapshot_id,
	source_id,
	created_at,
	updated_at
FROM
	continual_improvements
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
		return fmt.Errorf("cannot query continual improvements: %w", err)
	}

	improvements, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ContinualImprovement])
	if err != nil {
		return fmt.Errorf("cannot collect continual improvements: %w", err)
	}

	*cis = improvements

	return nil
}

func (ci *ContinualImprovement) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO continual_improvements (
	id,
	tenant_id,
	organization_id,
	reference_id,
	description,
	source,
	owner_profile_id,
	target_date,
	status,
	priority,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@reference_id,
	@description,
	@source,
	@owner_profile_id,
	@target_date,
	@status,
	@priority,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":               ci.ID,
		"tenant_id":        scope.GetTenantID(),
		"organization_id":  ci.OrganizationID,
		"reference_id":     ci.ReferenceID,
		"description":      ci.Description,
		"source":           ci.Source,
		"owner_profile_id": ci.OwnerID,
		"target_date":      ci.TargetDate,
		"status":           ci.Status,
		"priority":         ci.Priority,
		"created_at":       ci.CreatedAt,
		"updated_at":       ci.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert continual improvement: %w", err)
	}

	return nil
}

func (ci *ContinualImprovement) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE continual_improvements SET
	reference_id = @reference_id,
	description = @description,
	source = @source,
	owner_profile_id = @owner_profile_id,
	target_date = @target_date,
	status = @status,
	priority = @priority,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":               ci.ID,
		"reference_id":     ci.ReferenceID,
		"description":      ci.Description,
		"source":           ci.Source,
		"owner_profile_id": ci.OwnerID,
		"target_date":      ci.TargetDate,
		"status":           ci.Status,
		"priority":         ci.Priority,
		"updated_at":       ci.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update continual improvement: %w", err)
	}

	return nil
}

func (ci *ContinualImprovement) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM continual_improvements
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": ci.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete continual improvement: %w", err)
	}

	return nil
}

func (cis ContinualImprovements) Snapshot(ctx context.Context, conn pg.Conn, scope Scoper, organizationID, snapshotID gid.GID) error {
	query := `
INSERT INTO continual_improvements (
	id,
	tenant_id,
	snapshot_id,
	source_id,
	organization_id,
	reference_id,
	description,
	source,
	owner_profile_id,
	target_date,
	status,
	priority,
	created_at,
	updated_at
)
SELECT
	generate_gid(decode_base64_unpadded(@tenant_id), @continual_improvement_entity_type),
	@tenant_id,
	@snapshot_id,
	r.id,
	r.organization_id,
	r.reference_id,
	r.description,
	r.source,
	r.owner_profile_id,
	r.target_date,
	r.status,
	r.priority,
	r.created_at,
	r.updated_at
FROM continual_improvements r
WHERE %s AND r.organization_id = @organization_id AND r.snapshot_id IS NULL
	`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"tenant_id":                         scope.GetTenantID(),
		"snapshot_id":                       snapshotID,
		"organization_id":                   organizationID,
		"continual_improvement_entity_type": ContinualImprovementEntityType,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert continual improvement snapshots: %w", err)
	}

	return nil
}
