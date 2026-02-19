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
	Report struct {
		ID                    gid.GID               `db:"id"`
		Name                  *string               `db:"name"`
		OrganizationID        gid.GID               `db:"organization_id"`
		FrameworkID           gid.GID               `db:"framework_id"`
		FileID                *gid.GID              `db:"file_id"`
		FrameworkType         *string               `db:"framework_type"`
		ValidFrom             *time.Time            `db:"valid_from"`
		ValidUntil            *time.Time            `db:"valid_until"`
		State                 ReportState           `db:"state"`
		TrustCenterVisibility TrustCenterVisibility `db:"trust_center_visibility"`
		CreatedAt             time.Time             `db:"created_at"`
		UpdatedAt             time.Time             `db:"updated_at"`
	}

	Reports []*Report
)

func (r *Report) CursorKey(field ReportOrderField) page.CursorKey {
	switch field {
	case ReportOrderFieldCreatedAt:
		return page.NewCursorKey(r.ID, r.CreatedAt)
	case ReportOrderFieldValidFrom:
		return page.NewCursorKey(r.ID, r.ValidFrom)
	case ReportOrderFieldValidUntil:
		return page.NewCursorKey(r.ID, r.ValidUntil)
	case ReportOrderFieldState:
		return page.NewCursorKey(r.ID, r.State)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (r *Report) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM reports WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, r.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query report authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (r *Report) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	reportID gid.GID,
) error {
	q := `
SELECT
	id,
	name,
	organization_id,
	framework_id,
	file_id,
	framework_type,
	valid_from,
	valid_until,
	state,
	trust_center_visibility,
	created_at,
	updated_at
FROM
	reports
WHERE
	%s
	AND id = @report_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"report_id": reportID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query report: %w", err)
	}

	report, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Report])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect report: %w", err)
	}

	*r = report

	return nil
}

func (r *Reports) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	reports
WHERE
	%s
	AND organization_id = @organization_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count reports: %w", err)
	}

	return count, nil
}

func (r *Reports) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[ReportOrderField],
	filter *ReportFilter,
) error {
	q := `
SELECT
	id,
	name,
	organization_id,
	framework_id,
	file_id,
	framework_type,
	valid_from,
	valid_until,
	state,
	trust_center_visibility,
	created_at,
	updated_at
FROM
	reports
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
		return fmt.Errorf("cannot query reports: %w", err)
	}

	reports, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Report])
	if err != nil {
		return fmt.Errorf("cannot collect reports: %w", err)
	}

	*r = reports

	return nil
}

func (r *Reports) LoadAllByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	filter *ReportFilter,
) error {
	q := `
SELECT
	id,
	name,
	organization_id,
	framework_id,
	file_id,
	framework_type,
	valid_from,
	valid_until,
	state,
	trust_center_visibility,
	created_at,
	updated_at
FROM
	reports
WHERE
	%s
	AND organization_id = @organization_id
	AND %s
ORDER BY valid_from DESC
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query reports: %w", err)
	}

	reports, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Report])
	if err != nil {
		return fmt.Errorf("cannot collect reports: %w", err)
	}

	*r = reports

	return nil
}

func (r *Report) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO reports (
	id,
	name,
	tenant_id,
	organization_id,
	framework_id,
	file_id,
	framework_type,
	valid_from,
	valid_until,
	state,
	trust_center_visibility,
	created_at,
	updated_at
) VALUES (
	@id,
	@name,
	@tenant_id,
	@organization_id,
	@framework_id,
	@file_id,
	@framework_type,
	@valid_from,
	@valid_until,
	@state,
	@trust_center_visibility,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                      r.ID,
		"name":                    r.Name,
		"tenant_id":               scope.GetTenantID(),
		"organization_id":         r.OrganizationID,
		"framework_id":            r.FrameworkID,
		"file_id":                 r.FileID,
		"framework_type":          r.FrameworkType,
		"valid_from":              r.ValidFrom,
		"valid_until":             r.ValidUntil,
		"state":                   r.State,
		"trust_center_visibility": r.TrustCenterVisibility,
		"created_at":              r.CreatedAt,
		"updated_at":              r.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert report: %w", err)
	}

	return nil
}

func (r *Report) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE reports
SET
	name = @name,
	file_id = @file_id,
	framework_type = @framework_type,
	valid_from = @valid_from,
	valid_until = @valid_until,
	state = @state,
	trust_center_visibility = @trust_center_visibility,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                      r.ID,
		"name":                    r.Name,
		"file_id":                 r.FileID,
		"framework_type":          r.FrameworkType,
		"valid_from":              r.ValidFrom,
		"valid_until":             r.ValidUntil,
		"state":                   r.State,
		"trust_center_visibility": r.TrustCenterVisibility,
		"updated_at":              r.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update report: %w", err)
	}

	return nil
}

func (r *Report) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM reports
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": r.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete report: %w", err)
	}

	return nil
}

func (r *Reports) CountByControlID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	controlID gid.GID,
) (int, error) {
	q := `
WITH reports_by_control AS (
		SELECT
			a.id,
			a.tenant_id
		FROM
			reports a
		INNER JOIN
			controls_reports ca ON a.id = ca.report_id
		WHERE
			ca.control_id = @control_id
	)
	SELECT
		COUNT(id)
	FROM
		reports_by_control
	WHERE %s
	`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"control_id": controlID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (r *Reports) LoadByControlID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	controlID gid.GID,
	cursor *page.Cursor[ReportOrderField],
) error {
	q := `
WITH reports_by_control AS (
	SELECT
		a.id,
		a.tenant_id,
		a.name,
		a.organization_id,
		a.framework_id,
		a.file_id,
		a.framework_type,
		a.valid_from,
		a.valid_until,
		a.state,
		a.trust_center_visibility,
		a.created_at,
		a.updated_at
	FROM
		reports a
	INNER JOIN
		controls_reports ca ON a.id = ca.report_id
	WHERE
		ca.control_id = @control_id
)
SELECT
	id,
	name,
	organization_id,
	framework_id,
	file_id,
	framework_type,
	valid_from,
	valid_until,
	state,
	trust_center_visibility,
	created_at,
	updated_at
FROM
	reports_by_control
WHERE %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"control_id": controlID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query reports: %w", err)
	}

	reports, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Report])
	if err != nil {
		return fmt.Errorf("cannot collect reports: %w", err)
	}

	*r = reports

	return nil
}

func (r *Report) LoadByFileID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	fileID gid.GID,
) error {
	q := `
SELECT
	id,
	name,
	organization_id,
	framework_id,
	file_id,
	framework_type,
	valid_from,
	valid_until,
	state,
	trust_center_visibility,
	created_at,
	updated_at
FROM
	reports
WHERE %s
	AND file_id = @file_id
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"file_id": fileID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query report: %w", err)
	}

	report, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Report])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect report: %w", err)
	}

	*r = report

	return nil
}
