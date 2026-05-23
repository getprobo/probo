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
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	TrackerResource struct {
		ID               gid.GID             `db:"id"`
		OrganizationID   gid.GID             `db:"organization_id"`
		CookieBannerID   gid.GID             `db:"cookie_banner_id"`
		CookieCategoryID gid.GID             `db:"cookie_category_id"`
		ResourceType     TrackerResourceType `db:"resource_type"`
		Origin           string              `db:"origin"`
		Path             string              `db:"path"`
		DisplayName      string              `db:"display_name"`
		Description      string              `db:"description"`
		Excluded         bool                `db:"excluded"`
		LastDetectedAt   *time.Time          `db:"last_detected_at"`
		CreatedAt        time.Time           `db:"created_at"`
		UpdatedAt        time.Time           `db:"updated_at"`
	}

	TrackerResources []*TrackerResource
)

func (tr *TrackerResource) CursorKey(field TrackerResourceOrderField) page.CursorKey {
	switch field {
	case TrackerResourceOrderFieldCreatedAt:
		return page.NewCursorKey(tr.ID, tr.CreatedAt)
	case TrackerResourceOrderFieldLastDetectedAt:
		if tr.LastDetectedAt == nil {
			return page.NewCursorKey(tr.ID, time.Time{})
		}

		return page.NewCursorKey(tr.ID, *tr.LastDetectedAt)
	case TrackerResourceOrderFieldOrigin:
		return page.NewCursorKey(tr.ID, tr.Origin)
	case TrackerResourceOrderFieldUpdatedAt:
		return page.NewCursorKey(tr.ID, tr.UpdatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (tr *TrackerResource) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM tracker_resources WHERE id = ANY(@resource_ids::text[])`

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

func (tr *TrackerResource) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	trackerResourceID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	cookie_category_id,
	resource_type,
	origin,
	path,
	display_name,
	description,
	excluded,
	last_detected_at,
	created_at,
	updated_at
FROM
	tracker_resources
WHERE
	%s
	AND id = @tracker_resource_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"tracker_resource_id": trackerResourceID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query tracker resources: %w", err)
	}

	resource, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TrackerResource])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect tracker resource: %w", err)
	}

	*tr = resource

	return nil
}

func (tr *TrackerResource) LoadByBannerTypeOriginPath(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
	resourceType TrackerResourceType,
	origin string,
	path string,
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	cookie_category_id,
	resource_type,
	origin,
	path,
	display_name,
	description,
	excluded,
	last_detected_at,
	created_at,
	updated_at
FROM
	tracker_resources
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
	AND resource_type = @resource_type
	AND origin = @origin
	AND path = @path
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"cookie_banner_id": cookieBannerID,
		"resource_type":    resourceType,
		"origin":           origin,
		"path":             path,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query tracker resources: %w", err)
	}

	resource, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TrackerResource])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect tracker resource: %w", err)
	}

	*tr = resource

	return nil
}

func (tr *TrackerResource) Insert(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO tracker_resources (
	id,
	tenant_id,
	organization_id,
	cookie_banner_id,
	cookie_category_id,
	resource_type,
	origin,
	path,
	display_name,
	description,
	excluded,
	last_detected_at,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@cookie_banner_id,
	@cookie_category_id,
	@resource_type,
	@origin,
	@path,
	@display_name,
	@description,
	@excluded,
	@last_detected_at,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                 tr.ID,
		"tenant_id":          scope.GetTenantID(),
		"organization_id":    tr.OrganizationID,
		"cookie_banner_id":   tr.CookieBannerID,
		"cookie_category_id": tr.CookieCategoryID,
		"resource_type":      tr.ResourceType,
		"origin":             tr.Origin,
		"path":               tr.Path,
		"display_name":       tr.DisplayName,
		"description":        tr.Description,
		"excluded":           tr.Excluded,
		"last_detected_at":   tr.LastDetectedAt,
		"created_at":         tr.CreatedAt,
		"updated_at":         tr.UpdatedAt,
	}

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "idx_tracker_resources_unique_resource_per_banner" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert tracker resource: %w", err)
	}

	return nil
}

// Upsert inserts a new tracker resource or bumps last_detected_at on the
// existing row matching (cookie_banner_id, resource_type, origin, path).
// Returns true when a new row was inserted.
func (tr *TrackerResource) Upsert(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) (bool, error) {
	q := `
INSERT INTO tracker_resources (
	id,
	tenant_id,
	organization_id,
	cookie_banner_id,
	cookie_category_id,
	resource_type,
	origin,
	path,
	display_name,
	description,
	excluded,
	last_detected_at,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@cookie_banner_id,
	@cookie_category_id,
	@resource_type,
	@origin,
	@path,
	@display_name,
	@description,
	@excluded,
	@last_detected_at,
	@created_at,
	@updated_at
)
ON CONFLICT (cookie_banner_id, resource_type, origin, path) DO UPDATE SET
	last_detected_at = GREATEST(tracker_resources.last_detected_at, EXCLUDED.last_detected_at),
	updated_at = EXCLUDED.updated_at
RETURNING
	id,
	organization_id,
	cookie_banner_id,
	cookie_category_id,
	resource_type,
	origin,
	path,
	display_name,
	description,
	excluded,
	last_detected_at,
	created_at,
	updated_at
`

	originalID := tr.ID

	args := pgx.StrictNamedArgs{
		"id":                 tr.ID,
		"tenant_id":          scope.GetTenantID(),
		"organization_id":    tr.OrganizationID,
		"cookie_banner_id":   tr.CookieBannerID,
		"cookie_category_id": tr.CookieCategoryID,
		"resource_type":      tr.ResourceType,
		"origin":             tr.Origin,
		"path":               tr.Path,
		"display_name":       tr.DisplayName,
		"description":        tr.Description,
		"excluded":           tr.Excluded,
		"last_detected_at":   tr.LastDetectedAt,
		"created_at":         tr.CreatedAt,
		"updated_at":         tr.UpdatedAt,
	}

	rows, err := tx.Query(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot upsert tracker resource: %w", err)
	}
	defer rows.Close()

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TrackerResource])
	if err != nil {
		return false, fmt.Errorf("cannot collect upsert result: %w", err)
	}

	*tr = row

	return originalID == tr.ID, nil
}

func (tr *TrackerResource) Update(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE tracker_resources
SET
	cookie_category_id = @cookie_category_id,
	display_name = @display_name,
	description = @description,
	excluded = @excluded,
	last_detected_at = @last_detected_at,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                 tr.ID,
		"cookie_category_id": tr.CookieCategoryID,
		"display_name":       tr.DisplayName,
		"description":        tr.Description,
		"excluded":           tr.Excluded,
		"last_detected_at":   tr.LastDetectedAt,
		"updated_at":         tr.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update tracker resource: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (tr *TrackerResource) Delete(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM tracker_resources
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": tr.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete tracker resource: %w", err)
	}

	return nil
}

func (trs *TrackerResources) LoadAllByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
	filter *TrackerResourceFilter,
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	cookie_category_id,
	resource_type,
	origin,
	path,
	display_name,
	description,
	excluded,
	last_detected_at,
	created_at,
	updated_at
FROM
	tracker_resources
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
	AND %s
ORDER BY
	created_at ASC, id ASC;
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_banner_id": cookieBannerID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query tracker resources: %w", err)
	}

	resources, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TrackerResource])
	if err != nil {
		return fmt.Errorf("cannot collect tracker resources: %w", err)
	}

	*trs = resources

	return nil
}

func (trs *TrackerResources) LoadUncategorisedByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
	cursor *page.Cursor[TrackerResourceOrderField],
	filter *TrackerResourceFilter,
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	cookie_category_id,
	resource_type,
	origin,
	path,
	display_name,
	description,
	excluded,
	last_detected_at,
	created_at,
	updated_at
FROM
	tracker_resources
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
	AND cookie_category_id = (
		SELECT id FROM cookie_categories
		WHERE cookie_banner_id = @cookie_banner_id
			AND kind = @category_kind
			AND %s
		LIMIT 1
	)
	AND %s
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{
		"cookie_banner_id": cookieBannerID,
		"category_kind":    CookieCategoryKindUncategorised,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query uncategorised tracker resources: %w", err)
	}

	resources, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TrackerResource])
	if err != nil {
		return fmt.Errorf("cannot collect uncategorised tracker resources: %w", err)
	}

	*trs = resources

	return nil
}

func (trs *TrackerResources) CountUncategorisedByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
	filter *TrackerResourceFilter,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	tracker_resources
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
	AND cookie_category_id = (
		SELECT id FROM cookie_categories
		WHERE cookie_banner_id = @cookie_banner_id
			AND kind = @category_kind
			AND %s
		LIMIT 1
	)
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{
		"cookie_banner_id": cookieBannerID,
		"category_kind":    CookieCategoryKindUncategorised,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (trs *TrackerResources) LoadByCookieCategoryID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieCategoryID gid.GID,
	cursor *page.Cursor[TrackerResourceOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	cookie_category_id,
	resource_type,
	origin,
	path,
	display_name,
	description,
	excluded,
	last_detected_at,
	created_at,
	updated_at
FROM
	tracker_resources
WHERE
	%s
	AND cookie_category_id = @cookie_category_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_category_id": cookieCategoryID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query tracker resources: %w", err)
	}

	resources, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TrackerResource])
	if err != nil {
		return fmt.Errorf("cannot collect tracker resources: %w", err)
	}

	*trs = resources

	return nil
}

func (trs *TrackerResources) CountByCookieCategoryID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieCategoryID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	tracker_resources
WHERE
	%s
	AND cookie_category_id = @cookie_category_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_category_id": cookieCategoryID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (trs *TrackerResources) MoveToCategoryByCookieCategoryID(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	sourceCategoryID gid.GID,
	targetCategoryID gid.GID,
) error {
	q := `
UPDATE tracker_resources
SET
	cookie_category_id = @target_category_id,
	updated_at = @updated_at
WHERE
	%s
	AND cookie_category_id = @source_category_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"source_category_id": sourceCategoryID,
		"target_category_id": targetCategoryID,
		"updated_at":         time.Now(),
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot move tracker resources to category: %w", err)
	}

	return nil
}
