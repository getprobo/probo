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
	"encoding/json"
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
	CookieBannerVersionSnapshot struct {
		PrivacyPolicyURL  string                                            `json:"privacy_policy_url"`
		ConsentExpiryDays int                                               `json:"consent_expiry_days"`
		ConsentMode       string                                            `json:"consent_mode"`
		DefaultLanguage   string                                            `json:"default_language"`
		Categories        []CookieBannerVersionSnapshotCategory             `json:"categories"`
		Translations      map[string]CookieBannerVersionSnapshotTranslation `json:"translations,omitempty"`
	}

	CookieBannerVersionSnapshotTranslation struct {
		UI         map[string]string                                `json:"ui"`
		Categories []CookieBannerVersionSnapshotCategoryTranslation `json:"categories"`
	}

	CookieBannerVersionSnapshotCategoryTranslation struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	CookieBannerVersionSnapshotCategory struct {
		Name        string             `json:"name"`
		Description string             `json:"description"`
		Kind        CookieCategoryKind `json:"kind"`
		Cookies     CookieItems        `json:"cookies"`
	}

	CookieBannerVersion struct {
		ID             gid.GID                  `db:"id"`
		OrganizationID gid.GID                  `db:"organization_id"`
		CookieBannerID gid.GID                  `db:"cookie_banner_id"`
		Version        int                      `db:"version"`
		State          CookieBannerVersionState `db:"state"`
		Snapshot       json.RawMessage          `db:"snapshot"`
		CreatedAt      time.Time                `db:"created_at"`
		UpdatedAt      time.Time                `db:"updated_at"`
	}

	CookieBannerVersions []*CookieBannerVersion
)

func (v *CookieBannerVersion) CursorKey(field CookieBannerVersionOrderField) page.CursorKey {
	switch field {
	case CookieBannerVersionOrderFieldCreatedAt:
		return page.NewCursorKey(v.ID, v.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (v *CookieBannerVersion) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	q := `SELECT organization_id FROM cookie_banner_versions WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, v.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}

		return nil, fmt.Errorf("cannot query cookie banner version authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (v *CookieBannerVersion) GetSnapshot() (CookieBannerVersionSnapshot, error) {
	var snapshot CookieBannerVersionSnapshot
	if err := json.Unmarshal(v.Snapshot, &snapshot); err != nil {
		return snapshot, fmt.Errorf("cannot unmarshal cookie banner version snapshot: %w", err)
	}
	return snapshot, nil
}

func (v *CookieBannerVersion) SetSnapshot(snapshot CookieBannerVersionSnapshot) error {
	data, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("cannot marshal cookie banner version snapshot: %w", err)
	}
	v.Snapshot = data
	return nil
}

func (v *CookieBannerVersion) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	versionID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	version,
	state,
	snapshot,
	created_at,
	updated_at
FROM
	cookie_banner_versions
WHERE
	%s
	AND id = @version_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"version_id": versionID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query cookie banner versions: %w", err)
	}

	version, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CookieBannerVersion])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect cookie banner version: %w", err)
	}

	*v = version

	return nil
}

func (v *CookieBannerVersions) LoadByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
	cursor *page.Cursor[CookieBannerVersionOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	version,
	state,
	snapshot,
	created_at,
	updated_at
FROM
	cookie_banner_versions
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_banner_id": cookieBannerID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query cookie banner versions: %w", err)
	}

	versions, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CookieBannerVersion])
	if err != nil {
		return fmt.Errorf("cannot collect cookie banner versions: %w", err)
	}

	*v = versions

	return nil
}

func (v *CookieBannerVersions) CountByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	cookie_banner_versions
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_banner_id": cookieBannerID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (v *CookieBannerVersion) LoadByCookieBannerIDAndVersion(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
	version int,
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	version,
	state,
	snapshot,
	created_at,
	updated_at
FROM
	cookie_banner_versions
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
	AND version = @version
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"cookie_banner_id": cookieBannerID,
		"version":          version,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query cookie banner versions: %w", err)
	}

	ver, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CookieBannerVersion])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect cookie banner version: %w", err)
	}

	*v = ver

	return nil
}

func (v *CookieBannerVersion) LoadLatestByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	version,
	state,
	snapshot,
	created_at,
	updated_at
FROM
	cookie_banner_versions
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
ORDER BY version DESC
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_banner_id": cookieBannerID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query cookie banner versions: %w", err)
	}

	ver, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CookieBannerVersion])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect cookie banner version: %w", err)
	}

	*v = ver

	return nil
}

func (v *CookieBannerVersion) LoadNextVersion(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
) (int, error) {
	q := `
SELECT
	COALESCE(MAX(version), 0) + 1
FROM
	cookie_banner_versions
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_banner_id": cookieBannerID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var nextVersion int
	if err := row.Scan(&nextVersion); err != nil {
		return 0, fmt.Errorf("cannot scan next version: %w", err)
	}

	return nextVersion, nil
}

func (v *CookieBannerVersion) Insert(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO cookie_banner_versions (
	id,
	tenant_id,
	organization_id,
	cookie_banner_id,
	version,
	state,
	snapshot,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@cookie_banner_id,
	@version,
	@state,
	@snapshot,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":               v.ID,
		"tenant_id":        scope.GetTenantID(),
		"organization_id":  v.OrganizationID,
		"cookie_banner_id": v.CookieBannerID,
		"version":          v.Version,
		"state":            v.State,
		"snapshot":         v.Snapshot,
		"created_at":       v.CreatedAt,
		"updated_at":       v.UpdatedAt,
	}

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert cookie banner version: %w", err)
	}

	return nil
}

func (v *CookieBannerVersion) Update(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE cookie_banner_versions
SET
	state = @state,
	snapshot = @snapshot,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":         v.ID,
		"state":      v.State,
		"snapshot":   v.Snapshot,
		"updated_at": v.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update cookie banner version: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (v *CookieBannerVersion) LoadLatestPublishedByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	version,
	state,
	snapshot,
	created_at,
	updated_at
FROM
	cookie_banner_versions
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
	AND state = @state
ORDER BY version DESC
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"cookie_banner_id": cookieBannerID,
		"state":            CookieBannerVersionStatePublished,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query cookie banner versions: %w", err)
	}

	ver, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CookieBannerVersion])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect cookie banner version: %w", err)
	}

	*v = ver

	return nil
}
