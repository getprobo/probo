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
	AccessInventoryResource struct {
		ID                   gid.GID   `db:"id"`
		OrganizationID       gid.GID   `db:"organization_id"`
		AccessReviewSourceID gid.GID   `db:"access_review_source_id"`
		ResourceKey          string    `db:"resource_key"`
		DisplayName          string    `db:"display_name"`
		Kind                 string    `db:"kind"`
		FirstSeenAt          time.Time `db:"first_seen_at"`
		LastSeenAt           time.Time `db:"last_seen_at"`
		CreatedAt            time.Time `db:"created_at"`
		UpdatedAt            time.Time `db:"updated_at"`
	}

	AccessInventoryResources []*AccessInventoryResource
)

func (r AccessInventoryResource) CursorKey(orderBy AccessInventoryResourceOrderField) page.CursorKey {
	switch orderBy {
	case AccessInventoryResourceOrderFieldCreatedAt:
		return page.NewCursorKey(r.ID, r.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (r *AccessInventoryResource) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM access_inventory_resources WHERE id = ANY(@resource_ids::text[])`

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

func (r *AccessInventoryResource) LoadByID(
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
    resource_key,
    display_name,
    kind,
    first_seen_at,
    last_seen_at,
    created_at,
    updated_at
FROM
    access_inventory_resources
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
		return fmt.Errorf("cannot query access_inventory_resources: %w", err)
	}

	resource, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessInventoryResource])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect access inventory resource: %w", err)
	}

	*r = resource

	return nil
}

func (r *AccessInventoryResource) LoadBySourceIDAndResourceKey(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	sourceID gid.GID,
	resourceKey string,
) error {
	q := `
SELECT
    id,
    organization_id,
    access_review_source_id,
    resource_key,
    display_name,
    kind,
    first_seen_at,
    last_seen_at,
    created_at,
    updated_at
FROM
    access_inventory_resources
WHERE
    %s
    AND access_review_source_id = @access_review_source_id
    AND resource_key = @resource_key
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"access_review_source_id": sourceID,
		"resource_key":            resourceKey,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access_inventory_resources: %w", err)
	}

	resource, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessInventoryResource])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect access inventory resource: %w", err)
	}

	*r = resource

	return nil
}

func (rs *AccessInventoryResources) LoadBySourceID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	sourceID gid.GID,
	cursor *page.Cursor[AccessInventoryResourceOrderField],
) error {
	q := `
SELECT
    id,
    organization_id,
    access_review_source_id,
    resource_key,
    display_name,
    kind,
    first_seen_at,
    last_seen_at,
    created_at,
    updated_at
FROM
    access_inventory_resources
WHERE
    %s
    AND access_review_source_id = @access_review_source_id
    AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"access_review_source_id": sourceID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access_inventory_resources: %w", err)
	}

	resources, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AccessInventoryResource])
	if err != nil {
		return fmt.Errorf("cannot collect access_inventory_resources: %w", err)
	}

	*rs = resources

	return nil
}

func (r *AccessInventoryResource) Upsert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) (inserted bool, err error) {
	q := `
INSERT INTO access_inventory_resources (
    id,
    tenant_id,
    organization_id,
    access_review_source_id,
    resource_key,
    display_name,
    kind,
    first_seen_at,
    last_seen_at,
    created_at,
    updated_at
) VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @access_review_source_id,
    @resource_key,
    @display_name,
    @kind,
    @first_seen_at,
    @last_seen_at,
    @created_at,
    @updated_at
)
ON CONFLICT (access_review_source_id, resource_key) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    kind         = EXCLUDED.kind,
    last_seen_at = EXCLUDED.last_seen_at,
    updated_at   = EXCLUDED.updated_at
RETURNING
    id,
    organization_id,
    access_review_source_id,
    resource_key,
    display_name,
    kind,
    first_seen_at,
    last_seen_at,
    created_at,
    updated_at
`
	originalID := r.ID

	args := pgx.StrictNamedArgs{
		"id":                      r.ID,
		"tenant_id":               scope.GetTenantID(),
		"organization_id":         r.OrganizationID,
		"access_review_source_id": r.AccessReviewSourceID,
		"resource_key":            r.ResourceKey,
		"display_name":            r.DisplayName,
		"kind":                    r.Kind,
		"first_seen_at":           r.FirstSeenAt,
		"last_seen_at":            r.LastSeenAt,
		"created_at":              r.CreatedAt,
		"updated_at":              r.UpdatedAt,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot upsert access inventory resource: %w", err)
	}

	resource, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessInventoryResource])
	if err != nil {
		return false, fmt.Errorf("cannot collect access inventory resource upsert result: %w", err)
	}

	*r = resource

	return originalID == r.ID, nil
}

func (r *AccessInventoryResource) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM access_inventory_resources
WHERE
    %s
    AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": r.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete access inventory resource: %w", err)
	}

	return nil
}
