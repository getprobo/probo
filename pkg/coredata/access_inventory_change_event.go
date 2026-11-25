// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	"encoding/json"
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
	AccessInventoryChangeEvent struct {
		ID                   gid.GID                   `db:"id"`
		OrganizationID       gid.GID                   `db:"organization_id"`
		AccessReviewSourceID gid.GID                   `db:"access_review_source_id"`
		GrantID              *gid.GID                  `db:"grant_id"`
		PrincipalKey         string                    `db:"principal_key"`
		ResourceKey          string                    `db:"resource_key"`
		EntitlementKey       string                    `db:"entitlement_key"`
		ChangeType           AccessInventoryChangeType `db:"change_type"`
		ObservedAt           time.Time                 `db:"observed_at"`
		ProviderOccurredAt   *time.Time                `db:"provider_occurred_at"`
		SyncRunID            *gid.GID                  `db:"sync_run_id"`
		Before               *json.RawMessage          `db:"before"`
		After                *json.RawMessage          `db:"after"`
		IdempotencyKey       string                    `db:"idempotency_key"`
		CreatedAt            time.Time                 `db:"created_at"`
	}

	AccessInventoryChangeEvents []*AccessInventoryChangeEvent
)

func (e AccessInventoryChangeEvent) CursorKey(orderBy AccessInventoryChangeEventOrderField) page.CursorKey {
	switch orderBy {
	case AccessInventoryChangeEventOrderFieldCreatedAt:
		return page.NewCursorKey(e.ID, e.CreatedAt)
	case AccessInventoryChangeEventOrderFieldObservedAt:
		return page.NewCursorKey(e.ID, e.ObservedAt)
	case AccessInventoryChangeEventOrderFieldEffectiveAt:
		return page.NewCursorKey(e.ID, e.EffectiveAt())
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (e AccessInventoryChangeEvent) EffectiveAt() time.Time {
	if e.ProviderOccurredAt != nil {
		return *e.ProviderOccurredAt
	}

	return e.ObservedAt
}

func (e *AccessInventoryChangeEvent) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM access_inventory_change_events WHERE id = ANY(@resource_ids::text[])`

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

func (e *AccessInventoryChangeEvent) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    access_review_source_id,
    grant_id,
    principal_key,
    resource_key,
    entitlement_key,
    change_type,
    observed_at,
    provider_occurred_at,
    sync_run_id,
    before,
    after,
    idempotency_key,
    created_at
FROM
    access_inventory_change_events
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
		return fmt.Errorf("cannot query access_inventory_change_events: %w", err)
	}

	event, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessInventoryChangeEvent])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect access inventory change event: %w", err)
	}

	*e = event

	return nil
}

func (es *AccessInventoryChangeEvents) LoadByGrantID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	grantID gid.GID,
	cursor *page.Cursor[AccessInventoryChangeEventOrderField],
	filter *AccessInventoryChangeEventFilter,
) error {
	q := `
SELECT
    id,
    organization_id,
    access_review_source_id,
    grant_id,
    principal_key,
    resource_key,
    entitlement_key,
    change_type,
    observed_at,
    provider_occurred_at,
    sync_run_id,
    before,
    after,
    idempotency_key,
    created_at
FROM
    access_inventory_change_events
WHERE
    %s
    AND grant_id = @grant_id
    AND %s
    AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"grant_id": grantID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access_inventory_change_events: %w", err)
	}

	events, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AccessInventoryChangeEvent])
	if err != nil {
		return fmt.Errorf("cannot collect access_inventory_change_events: %w", err)
	}

	*es = events

	return nil
}

func (es *AccessInventoryChangeEvents) CountByGrantID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	grantID gid.GID,
	filter *AccessInventoryChangeEventFilter,
) (int, error) {
	q := `
SELECT COUNT(id)
FROM access_inventory_change_events
WHERE
    %s
    AND grant_id = @grant_id
    AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"grant_id": grantID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count access_inventory_change_events: %w", err)
	}

	return count, nil
}

func (es *AccessInventoryChangeEvents) LoadBySourceID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	sourceID gid.GID,
	cursor *page.Cursor[AccessInventoryChangeEventOrderField],
	filter *AccessInventoryChangeEventFilter,
) error {
	q := `
SELECT
    id,
    organization_id,
    access_review_source_id,
    grant_id,
    principal_key,
    resource_key,
    entitlement_key,
    change_type,
    observed_at,
    provider_occurred_at,
    sync_run_id,
    before,
    after,
    idempotency_key,
    created_at
FROM
    access_inventory_change_events
WHERE
    %s
    AND access_review_source_id = @access_review_source_id
    AND %s
    AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"access_review_source_id": sourceID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access_inventory_change_events: %w", err)
	}

	events, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AccessInventoryChangeEvent])
	if err != nil {
		return fmt.Errorf("cannot collect access_inventory_change_events: %w", err)
	}

	*es = events

	return nil
}

func (es *AccessInventoryChangeEvents) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[AccessInventoryChangeEventOrderField],
	filter *AccessInventoryChangeEventFilter,
) error {
	q := `
SELECT
    id,
    organization_id,
    access_review_source_id,
    grant_id,
    principal_key,
    resource_key,
    entitlement_key,
    change_type,
    observed_at,
    provider_occurred_at,
    sync_run_id,
    before,
    after,
    idempotency_key,
    created_at
FROM
    access_inventory_change_events
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
		return fmt.Errorf("cannot query access_inventory_change_events: %w", err)
	}

	events, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AccessInventoryChangeEvent])
	if err != nil {
		return fmt.Errorf("cannot collect access_inventory_change_events: %w", err)
	}

	*es = events

	return nil
}

func (es *AccessInventoryChangeEvents) CountByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	filter *AccessInventoryChangeEventFilter,
) (int, error) {
	q := `
SELECT COUNT(id)
FROM access_inventory_change_events
WHERE
    %s
    AND organization_id = @organization_id
    AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count access_inventory_change_events: %w", err)
	}

	return count, nil
}

// Insert appends the change event. A false result means the idempotency key
// already exists for this source (ON CONFLICT DO NOTHING) and the caller
// should treat the write as a no-op.
func (e *AccessInventoryChangeEvent) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) (inserted bool, err error) {
	q := `
INSERT INTO access_inventory_change_events (
    id,
    tenant_id,
    organization_id,
    access_review_source_id,
    grant_id,
    principal_key,
    resource_key,
    entitlement_key,
    change_type,
    observed_at,
    provider_occurred_at,
    sync_run_id,
    before,
    after,
    idempotency_key,
    created_at
) VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @access_review_source_id,
    @grant_id,
    @principal_key,
    @resource_key,
    @entitlement_key,
    @change_type,
    @observed_at,
    @provider_occurred_at,
    @sync_run_id,
    @before,
    @after,
    @idempotency_key,
    @created_at
)
ON CONFLICT (access_review_source_id, idempotency_key) DO NOTHING
`

	args := pgx.StrictNamedArgs{
		"id":                      e.ID,
		"tenant_id":               scope.GetTenantID(),
		"organization_id":         e.OrganizationID,
		"access_review_source_id": e.AccessReviewSourceID,
		"grant_id":                e.GrantID,
		"principal_key":           e.PrincipalKey,
		"resource_key":            e.ResourceKey,
		"entitlement_key":         e.EntitlementKey,
		"change_type":             e.ChangeType,
		"observed_at":             e.ObservedAt,
		"provider_occurred_at":    e.ProviderOccurredAt,
		"sync_run_id":             e.SyncRunID,
		"before":                  e.Before,
		"after":                   e.After,
		"idempotency_key":         e.IdempotencyKey,
		"created_at":              e.CreatedAt,
	}

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot insert access inventory change event: %w", err)
	}

	return result.RowsAffected() > 0, nil
}
