// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	AccessSource struct {
		ID             gid.GID              `db:"id"`
		OrganizationID gid.GID              `db:"organization_id"`
		ConnectorID    *gid.GID             `db:"connector_id"`
		Name           string               `db:"name"`
		Category       AccessSourceCategory `db:"category"`
		CsvData        *string              `db:"csv_data"`
		NameSyncedAt   *time.Time           `db:"name_synced_at"`
		CreatedAt      time.Time            `db:"created_at"`
		UpdatedAt      time.Time            `db:"updated_at"`
	}

	AccessSources []*AccessSource
)

func (as AccessSource) CursorKey(orderBy AccessSourceOrderField) page.CursorKey {
	switch orderBy {
	case AccessSourceOrderFieldCreatedAt:
		return page.NewCursorKey(as.ID, as.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (as *AccessSource) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM access_sources WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, as.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query access source authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (as *AccessSource) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    connector_id,
    name,
    category,
    csv_data,
    name_synced_at,
    created_at,
    updated_at
FROM
    access_sources
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
		return fmt.Errorf("cannot query access_sources: %w", err)
	}

	source, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessSource])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect access source: %w", err)
	}

	*as = source

	return nil
}

func (as *AccessSource) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    access_sources (
        id,
        tenant_id,
        organization_id,
        connector_id,
        name,
        category,
        csv_data,
        name_synced_at,
        created_at,
        updated_at
    )
VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @connector_id,
    @name,
    @category,
    @csv_data,
    @name_synced_at,
    @created_at,
    @updated_at
);
`

	args := pgx.StrictNamedArgs{
		"id":              as.ID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": as.OrganizationID,
		"connector_id":    as.ConnectorID,
		"name":            as.Name,
		"category":        as.Category,
		"csv_data":        as.CsvData,
		"name_synced_at":  as.NameSyncedAt,
		"created_at":      as.CreatedAt,
		"updated_at":      as.UpdatedAt,
	}
	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert access_source: %w", err)
	}

	return nil
}

func (as *AccessSource) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE access_sources
SET
    name = @name,
    category = @category,
    connector_id = @connector_id,
    csv_data = @csv_data,
    name_synced_at = @name_synced_at,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":             as.ID,
		"name":           as.Name,
		"category":       as.Category,
		"connector_id":   as.ConnectorID,
		"csv_data":       as.CsvData,
		"name_synced_at": as.NameSyncedAt,
		"updated_at":     as.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update access_source: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (as *AccessSource) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM access_sources
WHERE %s AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": as.ID}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete access_source: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (sources *AccessSources) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[AccessSourceOrderField],
) error {
	q := `
SELECT
    id,
    organization_id,
    connector_id,
    name,
    category,
    csv_data,
    name_synced_at,
    created_at,
    updated_at
FROM
    access_sources
WHERE
    %s
    AND organization_id = @organization_id
    AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access_sources: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AccessSource])
	if err != nil {
		return fmt.Errorf("cannot collect access_sources: %w", err)
	}

	*sources = result

	return nil
}

func (sources *AccessSources) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT COUNT(id)
FROM access_sources
WHERE
    %s
    AND organization_id = @organization_id;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count access_sources: %w", err)
	}

	return count, nil
}

// LoadScopeSourcesByCampaignID loads the campaign scope sources in deterministic
// name order. Only explicitly scoped sources are returned.
func (sources *AccessSources) LoadScopeSourcesByCampaignID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	campaignID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    connector_id,
    name,
    category,
    csv_data,
    name_synced_at,
    created_at,
    updated_at
FROM
    access_sources
WHERE
    %s
    AND id IN (
        SELECT arcss.access_source_id
        FROM access_review_campaign_scope_systems arcss
        WHERE arcss.access_review_campaign_id = @campaign_id
    )
ORDER BY name ASC
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"campaign_id": campaignID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query scope access_sources: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AccessSource])
	if err != nil {
		return fmt.Errorf("cannot collect scope access_sources: %w", err)
	}

	*sources = result

	return nil
}

// ErrNoAccessSourceNameSyncAvailable is returned when no access source
// needs its name synced from its connector.
var ErrNoAccessSourceNameSyncAvailable = fmt.Errorf("no access source name sync available")

// LoadNextUnsyncedNameForUpdateSkipLocked claims the next access source that
// has a connector but has not yet had its name synced. The row is locked with
// FOR UPDATE SKIP LOCKED so concurrent workers do not pick the same row.
func (as *AccessSource) LoadNextUnsyncedNameForUpdateSkipLocked(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
SELECT
    id,
    organization_id,
    connector_id,
    name,
    category,
    csv_data,
    name_synced_at,
    created_at,
    updated_at
FROM
    access_sources
WHERE
    connector_id IS NOT NULL
    AND name_synced_at IS NULL
ORDER BY
    created_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED;
`

	rows, err := conn.Query(ctx, q)
	if err != nil {
		return fmt.Errorf("cannot query unsynced access_sources: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessSource])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoAccessSourceNameSyncAvailable
		}
		return fmt.Errorf("cannot collect unsynced access source: %w", err)
	}

	*as = row
	return nil
}
