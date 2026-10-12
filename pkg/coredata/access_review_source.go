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
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	AccessReviewSource struct {
		ID                 gid.GID   `db:"id"`
		OrganizationID     gid.GID   `db:"organization_id"`
		ConnectorAccountID *gid.GID  `db:"connector_account_id"`
		Name               string    `db:"name"`
		CsvData            *string   `db:"csv_data"`
		CreatedAt          time.Time `db:"created_at"`
		UpdatedAt          time.Time `db:"updated_at"`
	}

	AccessReviewSources []*AccessReviewSource
)

func (as AccessReviewSource) CursorKey(orderBy AccessReviewSourceOrderField) page.CursorKey {
	switch orderBy {
	case AccessReviewSourceOrderFieldCreatedAt:
		return page.NewCursorKey(as.ID, as.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (as *AccessReviewSource) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM access_review_sources WHERE id = ANY(@resource_ids::text[])`

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

func (as *AccessReviewSource) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    connector_account_id,
    name,
    csv_data,
    created_at,
    updated_at
FROM
    access_review_sources
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
		return fmt.Errorf("cannot query access_review_sources: %w", err)
	}

	source, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessReviewSource])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect access source: %w", err)
	}

	*as = source

	return nil
}

// LoadByIDForUpdate is LoadByID under FOR UPDATE so concurrent updates
// of the same source serialize.
func (as *AccessReviewSource) LoadByIDForUpdate(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    connector_account_id,
    name,
    csv_data,
    created_at,
    updated_at
FROM
    access_review_sources
WHERE
    %s
    AND id = @id
LIMIT 1
FOR UPDATE;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access_review_sources: %w", err)
	}

	source, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessReviewSource])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect access source: %w", err)
	}

	*as = source

	return nil
}

func (sources *AccessReviewSources) LoadByIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	ids []gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    connector_account_id,
    name,
    csv_data,
    created_at,
    updated_at
FROM
    access_review_sources
WHERE
    %s
    AND id = ANY(@ids)
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"ids": ids}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access_review_sources: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AccessReviewSource])
	if err != nil {
		return fmt.Errorf("cannot collect access_review_sources: %w", err)
	}

	*sources = result

	if len(result) != len(gid.NewSet(ids...)) {
		return ErrResourceNotFound
	}

	return nil
}

func (sources *AccessReviewSources) CountByConnectorID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	connectorID gid.GID,
) (int, error) {
	q := `
SELECT COUNT(id)
FROM access_review_sources
WHERE
    %s
    AND connector_account_id IN (
        SELECT id
        FROM connector_accounts
        WHERE
            %s
            AND connector_id = @connector_id
    );
`
	q = fmt.Sprintf(q, scope.SQLFragment(), scope.SQLFragment())

	args := pgx.StrictNamedArgs{"connector_id": connectorID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count access_review_sources by connector ID: %w", err)
	}

	return count, nil
}

// Insert reports whether a row was inserted. The partial unique index
// on connector_account_id arbitrates: a source referencing an
// already-taken connector account is skipped, making creation
// idempotent per account. CSV sources (nil account) always insert.
func (as *AccessReviewSource) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) (bool, error) {
	q := `
INSERT INTO
    access_review_sources (
        id,
        tenant_id,
        organization_id,
        connector_account_id,
        name,
        csv_data,
        created_at,
        updated_at
    )
VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @connector_account_id,
    @name,
    @csv_data,
    @created_at,
    @updated_at
)
ON CONFLICT (connector_account_id) WHERE connector_account_id IS NOT NULL DO NOTHING
RETURNING id;
`

	args := pgx.StrictNamedArgs{
		"id":                   as.ID,
		"tenant_id":            scope.GetTenantID(),
		"organization_id":      as.OrganizationID,
		"connector_account_id": as.ConnectorAccountID,
		"name":                 as.Name,
		"csv_data":             as.CsvData,
		"created_at":           as.CreatedAt,
		"updated_at":           as.UpdatedAt,
	}

	var insertedID gid.GID

	err := conn.QueryRow(ctx, q, args).Scan(&insertedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("cannot insert access_source: %w", err)
	}

	return true, nil
}

func (as *AccessReviewSource) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	// connector_id is a leftover foreign key. Clear it so a relink does not
	// keep the previous connector referenced.
	q := `
UPDATE access_review_sources
SET
    name = @name,
    connector_account_id = @connector_account_id,
    connector_id = NULL,
    csv_data = @csv_data,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                   as.ID,
		"name":                 as.Name,
		"connector_account_id": as.ConnectorAccountID,
		"csv_data":             as.CsvData,
		"updated_at":           as.UpdatedAt,
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

// Delete removes the source. ErrResourceNotFound when the source does
// not exist.
func (as *AccessReviewSource) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	if err := as.LoadByIDForUpdate(ctx, conn, scope, as.ID); err != nil {
		return err
	}

	q := `
DELETE FROM access_review_sources
WHERE %s AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": as.ID}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot delete access_source: %w", err)
	}

	return nil
}

// LoadByConnectorAccountID loads the access source referencing the
// connector account. Returns ErrResourceNotFound when none references it.
func (as *AccessReviewSource) LoadByConnectorAccountID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	connectorAccountID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    connector_account_id,
    name,
    csv_data,
    created_at,
    updated_at
FROM
    access_review_sources
WHERE
    %s
    AND connector_account_id = @connector_account_id
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"connector_account_id": connectorAccountID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access_review_sources: %w", err)
	}

	source, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessReviewSource])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect access source: %w", err)
	}

	*as = source

	return nil
}

func (sources *AccessReviewSources) CountByConnectorAccountID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	connectorAccountID gid.GID,
) (int, error) {
	q := `
SELECT COUNT(id)
FROM access_review_sources
WHERE
    %s
    AND connector_account_id = @connector_account_id;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"connector_account_id": connectorAccountID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count access_review_sources by connector account ID: %w", err)
	}

	return count, nil
}

func (sources *AccessReviewSources) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[AccessReviewSourceOrderField],
) error {
	q := `
SELECT
    id,
    organization_id,
    connector_account_id,
    name,
    csv_data,
    created_at,
    updated_at
FROM
    access_review_sources
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
		return fmt.Errorf("cannot query access_review_sources: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AccessReviewSource])
	if err != nil {
		return fmt.Errorf("cannot collect access_review_sources: %w", err)
	}

	*sources = result

	return nil
}

func (sources *AccessReviewSources) CountByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT COUNT(id)
FROM access_review_sources
WHERE
    %s
    AND organization_id = @organization_id;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count access_review_sources: %w", err)
	}

	return count, nil
}
