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
	Obligation struct {
		ID                     gid.GID          `db:"id"`
		OrganizationID         gid.GID          `db:"organization_id"`
		Area                   *string          `db:"area"`
		Source                 *string          `db:"source"`
		Requirement            *string          `db:"requirement"`
		ActionsToBeImplemented *string          `db:"actions_to_be_implemented"`
		Regulator              *string          `db:"regulator"`
		OwnerID                gid.GID          `db:"owner_profile_id"`
		LastReviewDate         *time.Time       `db:"last_review_date"`
		DueDate                *time.Time       `db:"due_date"`
		Status                 ObligationStatus `db:"status"`
		Type                   ObligationType   `db:"type"`
		SnapshotID             *gid.GID         `db:"snapshot_id"`
		SourceID               *gid.GID         `db:"source_id"`
		CreatedAt              time.Time        `db:"created_at"`
		UpdatedAt              time.Time        `db:"updated_at"`
	}

	Obligations []*Obligation
)

func (o *Obligation) CursorKey(field ObligationOrderField) page.CursorKey {
	switch field {
	case ObligationOrderFieldCreatedAt:
		return page.NewCursorKey(o.ID, o.CreatedAt)
	case ObligationOrderFieldLastReviewDate:
		return page.NewCursorKey(o.ID, o.LastReviewDate)
	case ObligationOrderFieldDueDate:
		return page.NewCursorKey(o.ID, o.DueDate)
	case ObligationOrderFieldStatus:
		return page.NewCursorKey(o.ID, o.Status)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (o *Obligation) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM obligations WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, o.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query obligation authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (o *Obligation) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	obligationID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	snapshot_id,
	source_id,
	area,
	source,
	requirement,
	actions_to_be_implemented,
	regulator,
	owner_profile_id,
	last_review_date,
	due_date,
	status,
	type,
	created_at,
	updated_at
FROM
	obligations
WHERE
	%s
	AND id = @obligation_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"obligation_id": obligationID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query obligation: %w", err)
	}

	obligation, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Obligation])
	if err != nil {
		return fmt.Errorf("cannot collect obligation: %w", err)
	}

	*o = obligation

	return nil
}

func (os *Obligations) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	filter *ObligationFilter,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	obligations
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
		return 0, fmt.Errorf("cannot count obligations: %w", err)
	}

	return count, nil
}

func (os *Obligations) CountByRiskID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	riskID gid.GID,
	filter *ObligationFilter,
) (int, error) {
	q := `
WITH obls AS (
	SELECT
		o.id,
		o.tenant_id,
		o.snapshot_id,
		o.search_vector
	FROM
		obligations o
	INNER JOIN
		risks_obligations ro ON o.id = ro.obligation_id
	WHERE
		ro.risk_id = @risk_id
)
SELECT
	COUNT(id)
FROM
	obls
WHERE %s
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"risk_id": riskID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count obligations: %w", err)
	}

	return count, nil
}

func (os *Obligations) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[ObligationOrderField],
	filter *ObligationFilter,
) error {
	q := `
SELECT
	id,
	organization_id,
	area,
	source,
	requirement,
	actions_to_be_implemented,
	regulator,
	owner_profile_id,
	last_review_date,
	due_date,
	status,
	type,
	snapshot_id,
	source_id,
	created_at,
	updated_at
FROM
	obligations
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
		return fmt.Errorf("cannot query obligations: %w", err)
	}

	obligations, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Obligation])
	if err != nil {
		return fmt.Errorf("cannot collect obligations: %w", err)
	}

	*os = obligations

	return nil
}

func (os *Obligations) LoadByRiskID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	riskID gid.GID,
	cursor *page.Cursor[ObligationOrderField],
	filter *ObligationFilter,
) error {
	q := `
WITH obls AS (
	SELECT
		o.id,
		o.organization_id,
		o.area,
		o.source,
		o.requirement,
		o.actions_to_be_implemented,
		o.regulator,
		o.owner_profile_id,
		o.last_review_date,
		o.due_date,
		o.status,
		o.type,
		o.snapshot_id,
		o.source_id,
		o.created_at,
		o.updated_at,
		o.tenant_id,
		o.search_vector
	FROM
		obligations o
	INNER JOIN
		risks_obligations ro ON o.id = ro.obligation_id
	WHERE
		ro.risk_id = @risk_id
)
SELECT
	id,
	organization_id,
	area,
	source,
	requirement,
	actions_to_be_implemented,
	regulator,
	owner_profile_id,
	last_review_date,
	due_date,
	status,
	type,
	snapshot_id,
	source_id,
	created_at,
	updated_at
FROM
	obls
WHERE %s
	AND %s
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"risk_id": riskID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query obligations: %w", err)
	}

	obligations, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Obligation])
	if err != nil {
		return fmt.Errorf("cannot collect obligations: %w", err)
	}

	*os = obligations

	return nil
}

func (os *Obligations) CountByControlID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	controlID gid.GID,
	filter *ObligationFilter,
) (int, error) {
	q := `
WITH obls AS (
	SELECT
		o.id,
		o.tenant_id,
		o.snapshot_id
	FROM
		obligations o
	INNER JOIN
		controls_obligations co ON o.id = co.obligation_id
	WHERE
		co.control_id = @control_id
)
SELECT
	COUNT(id)
FROM
	obls
WHERE %s
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"control_id": controlID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count obligations: %w", err)
	}

	return count, nil
}

func (os *Obligations) LoadByControlID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	controlID gid.GID,
	cursor *page.Cursor[ObligationOrderField],
	filter *ObligationFilter,
) error {
	q := `
WITH obls AS (
	SELECT
		o.id,
		o.organization_id,
		o.area,
		o.source,
		o.requirement,
		o.actions_to_be_implemented,
		o.regulator,
		o.owner_profile_id,
		o.last_review_date,
		o.due_date,
		o.status,
		o.type,
		o.snapshot_id,
		o.source_id,
		o.created_at,
		o.updated_at,
		o.tenant_id
	FROM
		obligations o
	INNER JOIN
		controls_obligations co ON o.id = co.obligation_id
	WHERE
		co.control_id = @control_id
)
SELECT
	id,
	organization_id,
	area,
	source,
	requirement,
	actions_to_be_implemented,
	regulator,
	owner_profile_id,
	last_review_date,
	due_date,
	status,
	type,
	snapshot_id,
	source_id,
	created_at,
	updated_at
FROM
	obls
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
		return fmt.Errorf("cannot query obligations: %w", err)
	}

	obligations, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Obligation])
	if err != nil {
		return fmt.Errorf("cannot collect obligations: %w", err)
	}

	*os = obligations

	return nil
}

func (o *Obligation) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO obligations (
	id,
	tenant_id,
	organization_id,
	area,
	source,
	requirement,
	actions_to_be_implemented,
	regulator,
	owner_profile_id,
	last_review_date,
	due_date,
	status,
	type,
	snapshot_id,
	source_id,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@area,
	@source,
	@requirement,
	@actions_to_be_implemented,
	@regulator,
	@owner_profile_id,
	@last_review_date,
	@due_date,
	@status,
	@type,
	@snapshot_id,
	@source_id,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                        o.ID,
		"tenant_id":                 scope.GetTenantID(),
		"organization_id":           o.OrganizationID,
		"area":                      o.Area,
		"source":                    o.Source,
		"requirement":               o.Requirement,
		"actions_to_be_implemented": o.ActionsToBeImplemented,
		"regulator":                 o.Regulator,
		"owner_profile_id":          o.OwnerID,
		"last_review_date":          o.LastReviewDate,
		"due_date":                  o.DueDate,
		"status":                    o.Status,
		"type":                      o.Type,
		"snapshot_id":               o.SnapshotID,
		"source_id":                 o.SourceID,
		"created_at":                o.CreatedAt,
		"updated_at":                o.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert obligation: %w", err)
	}

	return nil
}

func (o *Obligation) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE obligations SET
	area = @area,
	source = @source,
	requirement = @requirement,
	actions_to_be_implemented = @actions_to_be_implemented,
	regulator = @regulator,
	owner_profile_id = @owner_profile_id,
	last_review_date = @last_review_date,
	due_date = @due_date,
	status = @status,
	type = @type,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
	AND snapshot_id IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                        o.ID,
		"area":                      o.Area,
		"source":                    o.Source,
		"requirement":               o.Requirement,
		"actions_to_be_implemented": o.ActionsToBeImplemented,
		"regulator":                 o.Regulator,
		"owner_profile_id":          o.OwnerID,
		"last_review_date":          o.LastReviewDate,
		"due_date":                  o.DueDate,
		"status":                    o.Status,
		"type":                      o.Type,
		"updated_at":                o.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update obligation: %w", err)
	}

	return nil
}

func (o *Obligation) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM obligations
WHERE
	%s
	AND id = @id
	AND snapshot_id IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": o.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete obligation: %w", err)
	}

	return nil
}

func (os Obligations) Snapshot(ctx context.Context, conn pg.Conn, scope Scoper, organizationID, snapshotID gid.GID) error {
	query := `
INSERT INTO obligations (
	id,
	tenant_id,
	snapshot_id,
	source_id,
	organization_id,
	area,
	source,
	requirement,
	actions_to_be_implemented,
	regulator,
	owner_profile_id,
	last_review_date,
	due_date,
	status,
	type,
	created_at,
	updated_at
)
SELECT
	generate_gid(decode_base64_unpadded(@tenant_id), @obligation_entity_type),
	@tenant_id,
	@snapshot_id,
	o.id,
	o.organization_id,
	o.area,
	o.source,
	o.requirement,
	o.actions_to_be_implemented,
	o.regulator,
	o.owner_profile_id,
	o.last_review_date,
	o.due_date,
	o.status,
	o.type,
	o.created_at,
	o.updated_at
FROM obligations o
WHERE %s AND o.organization_id = @organization_id AND o.snapshot_id IS NULL
	`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"tenant_id":              scope.GetTenantID(),
		"snapshot_id":            snapshotID,
		"organization_id":        organizationID,
		"obligation_entity_type": ObligationEntityType,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert obligation snapshots: %w", err)
	}

	return nil
}
