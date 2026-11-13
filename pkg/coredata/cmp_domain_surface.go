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
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	// CMPDomainSurface is an origin under a CMP domain that a scan found
	// a link to. Status is the user's decision. State is computed when
	// the row is read and is COVERED while a banner uses the origin.
	CMPDomainSurface struct {
		ID                         gid.GID                `db:"id"`
		OrganizationID             gid.GID                `db:"organization_id"`
		CMPDomainID                gid.GID                `db:"cmp_domain_id"`
		Origin                     string                 `db:"origin"`
		Status                     CMPDomainSurfaceStatus `db:"status"`
		State                      CMPDomainSurfaceState  `db:"state"`
		DiscoveredByCookieBannerID *gid.GID               `db:"discovered_by_cookie_banner_id"`
		LastCookieBannerScanID     *gid.GID               `db:"last_cookie_banner_scan_id"`
		LastSeenAt                 time.Time              `db:"last_seen_at"`
		CreatedAt                  time.Time              `db:"created_at"`
		UpdatedAt                  time.Time              `db:"updated_at"`
	}

	CMPDomainSurfaces []*CMPDomainSurface
)

func (s *CMPDomainSurface) CursorKey(field CMPDomainSurfaceOrderField) page.CursorKey {
	switch field {
	case CMPDomainSurfaceOrderFieldCreatedAt:
		return page.NewCursorKey(s.ID, s.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (s *CMPDomainSurface) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM cmp_domain_surfaces WHERE id = ANY(@resource_ids::text[])`

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

func (s *CMPDomainSurface) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	surfaceID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	cmp_domain_id,
	origin,
	status,
	CASE
		WHEN EXISTS (
			SELECT 1
			FROM cookie_banners
			WHERE cookie_banners.cmp_domain_id = cmp_domain_surfaces.cmp_domain_id
				AND lower(cookie_banners.origin) = lower(cmp_domain_surfaces.origin)
				AND cookie_banners.state = @active_banner_state
		) THEN @covered_state::text
		ELSE cmp_domain_surfaces.status
	END AS state,
	discovered_by_cookie_banner_id,
	last_cookie_banner_scan_id,
	last_seen_at,
	created_at,
	updated_at
FROM cmp_domain_surfaces
WHERE
	%s
	AND id = @surface_id
LIMIT 1
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"surface_id":          surfaceID,
		"covered_state":       CMPDomainSurfaceStateCovered,
		"active_banner_state": CookieBannerStateActive,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query cmp domain surface: %w", err)
	}

	surface, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CMPDomainSurface])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect cmp domain surface: %w", err)
	}

	*s = surface

	return nil
}

func (s *CMPDomainSurfaces) LoadByCMPDomainID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cmpDomainID gid.GID,
	cursor *page.Cursor[CMPDomainSurfaceOrderField],
	filter *CMPDomainSurfaceFilter,
) error {
	if filter == nil {
		filter = NewCMPDomainSurfaceFilter(nil)
	}

	q := `
SELECT
	id,
	organization_id,
	cmp_domain_id,
	origin,
	status,
	CASE
		WHEN EXISTS (
			SELECT 1
			FROM cookie_banners
			WHERE cookie_banners.cmp_domain_id = cmp_domain_surfaces.cmp_domain_id
				AND lower(cookie_banners.origin) = lower(cmp_domain_surfaces.origin)
				AND cookie_banners.state = @active_banner_state
		) THEN @covered_state::text
		ELSE cmp_domain_surfaces.status
	END AS state,
	discovered_by_cookie_banner_id,
	last_cookie_banner_scan_id,
	last_seen_at,
	created_at,
	updated_at
FROM cmp_domain_surfaces
WHERE
	%s
	AND cmp_domain_id = @cmp_domain_id
	AND %s
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"cmp_domain_id": cmpDomainID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	cursorArgs, err := cursor.SQLArguments()
	if err != nil {
		return fmt.Errorf("cannot validate cursor: %w", err)
	}

	maps.Copy(args, cursorArgs)

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query cmp domain surfaces: %w", err)
	}

	surfaces, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CMPDomainSurface])
	if err != nil {
		return fmt.Errorf("cannot collect cmp domain surfaces: %w", err)
	}

	*s = surfaces

	return nil
}

func (s *CMPDomainSurfaces) CountByCMPDomainID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cmpDomainID gid.GID,
	filter *CMPDomainSurfaceFilter,
) (int, error) {
	if filter == nil {
		filter = NewCMPDomainSurfaceFilter(nil)
	}

	q := `
SELECT COUNT(id)
FROM cmp_domain_surfaces
WHERE
	%s
	AND cmp_domain_id = @cmp_domain_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"cmp_domain_id": cmpDomainID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count cmp domain surfaces: %w", err)
	}

	return count, nil
}

func (s *CMPDomainSurface) Insert(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	if s.Status == "" {
		s.Status = CMPDomainSurfaceStatusUnconfirmed
	}

	q := `
INSERT INTO cmp_domain_surfaces (
	id,
	tenant_id,
	organization_id,
	cmp_domain_id,
	origin,
	status,
	discovered_by_cookie_banner_id,
	last_cookie_banner_scan_id,
	last_seen_at,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@cmp_domain_id,
	@origin,
	@status,
	@discovered_by_cookie_banner_id,
	@last_cookie_banner_scan_id,
	@last_seen_at,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                             s.ID,
		"tenant_id":                      scope.GetTenantID(),
		"organization_id":                s.OrganizationID,
		"cmp_domain_id":                  s.CMPDomainID,
		"origin":                         s.Origin,
		"status":                         s.Status,
		"discovered_by_cookie_banner_id": s.DiscoveredByCookieBannerID,
		"last_cookie_banner_scan_id":     s.LastCookieBannerScanID,
		"last_seen_at":                   s.LastSeenAt,
		"created_at":                     s.CreatedAt,
		"updated_at":                     s.UpdatedAt,
	}

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "cmp_domain_surfaces_cmp_domain_id_origin_key" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert cmp domain surface: %w", err)
	}

	return nil
}

// UpsertSeen inserts the surface or, when the origin was already
// discovered, refreshes when it was last seen and which scan saw it.
// The stored status and the banner that discovered it first stay.
// The caller sets ID; this method does not allocate one.
func (s *CMPDomainSurface) UpsertSeen(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	now time.Time,
) error {
	if s.ID == gid.Nil {
		return fmt.Errorf("cmp domain surface id is required")
	}

	if s.Status == "" {
		s.Status = CMPDomainSurfaceStatusUnconfirmed
	}

	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}

	s.LastSeenAt = now
	s.UpdatedAt = now

	q := `
INSERT INTO cmp_domain_surfaces (
	id,
	tenant_id,
	organization_id,
	cmp_domain_id,
	origin,
	status,
	discovered_by_cookie_banner_id,
	last_cookie_banner_scan_id,
	last_seen_at,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@cmp_domain_id,
	@origin,
	@status,
	@discovered_by_cookie_banner_id,
	@last_cookie_banner_scan_id,
	@last_seen_at,
	@created_at,
	@updated_at
)
ON CONFLICT ON CONSTRAINT cmp_domain_surfaces_cmp_domain_id_origin_key
DO UPDATE SET
	last_seen_at = EXCLUDED.last_seen_at,
	last_cookie_banner_scan_id = EXCLUDED.last_cookie_banner_scan_id,
	updated_at = EXCLUDED.updated_at
RETURNING id
`

	args := pgx.StrictNamedArgs{
		"id":                             s.ID,
		"tenant_id":                      scope.GetTenantID(),
		"organization_id":                s.OrganizationID,
		"cmp_domain_id":                  s.CMPDomainID,
		"origin":                         s.Origin,
		"status":                         s.Status,
		"discovered_by_cookie_banner_id": s.DiscoveredByCookieBannerID,
		"last_cookie_banner_scan_id":     s.LastCookieBannerScanID,
		"last_seen_at":                   s.LastSeenAt,
		"created_at":                     s.CreatedAt,
		"updated_at":                     s.UpdatedAt,
	}

	row := tx.QueryRow(ctx, q, args)

	if err := row.Scan(&s.ID); err != nil {
		return fmt.Errorf("cannot upsert cmp domain surface: %w", err)
	}

	return nil
}

func (s *CMPDomainSurface) UpdateStatus(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	status CMPDomainSurfaceStatus,
) error {
	if !status.IsValid() {
		return fmt.Errorf("cannot update cmp domain surface: invalid status %q", status)
	}

	q := `
UPDATE cmp_domain_surfaces
SET
	status = @status,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	now := time.Now()
	args := pgx.StrictNamedArgs{
		"id":         s.ID,
		"status":     status,
		"updated_at": now,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update cmp domain surface: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	s.Status = status
	s.UpdatedAt = now

	return nil
}

func (s *CMPDomainSurface) Delete(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM cmp_domain_surfaces
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": s.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete cmp domain surface: %w", err)
	}

	return nil
}
