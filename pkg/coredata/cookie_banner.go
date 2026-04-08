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
	CookieBanner struct {
		ID                   gid.GID           `db:"id"`
		OrganizationID       gid.GID           `db:"organization_id"`
		Name                 string            `db:"name"`
		Domain               string            `db:"domain"`
		State                CookieBannerState `db:"state"`
		Title                string            `db:"title"`
		Description          string            `db:"description"`
		AcceptAllLabel       string            `db:"accept_all_label"`
		RejectAllLabel       string            `db:"reject_all_label"`
		SavePreferencesLabel string            `db:"save_preferences_label"`
		PrivacyPolicyURL     string            `db:"privacy_policy_url"`
		ConsentExpiryDays    int               `db:"consent_expiry_days"`
		ConsentMode          ConsentMode       `db:"consent_mode"`
		Version              int               `db:"version"`
		Theme                json.RawMessage   `db:"theme"`
		CreatedAt            time.Time         `db:"created_at"`
		UpdatedAt            time.Time         `db:"updated_at"`
	}

	CookieBanners []*CookieBanner
)

func (b *CookieBanner) CursorKey(field CookieBannerOrderField) page.CursorKey {
	switch field {
	case CookieBannerOrderFieldCreatedAt:
		return page.NewCursorKey(b.ID, b.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (b *CookieBanner) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	q := `SELECT organization_id FROM cookie_banners WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, b.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}

		return nil, fmt.Errorf("cannot query cookie banner authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (b *CookieBanner) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	bannerID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	name,
	domain,
	state,
	title,
	description,
	accept_all_label,
	reject_all_label,
	save_preferences_label,
	privacy_policy_url,
	consent_expiry_days,
	consent_mode,
	version,
	theme,
	created_at,
	updated_at
FROM
	cookie_banners
WHERE
	%s
	AND id = @banner_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"banner_id": bannerID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query cookie banners: %w", err)
	}

	banner, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CookieBanner])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect cookie banner: %w", err)
	}

	*b = banner

	return nil
}

func (b *CookieBanner) LoadPublishedByID(
	ctx context.Context,
	conn pg.Querier,
	bannerID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	name,
	domain,
	state,
	title,
	description,
	accept_all_label,
	reject_all_label,
	save_preferences_label,
	privacy_policy_url,
	consent_expiry_days,
	consent_mode,
	version,
	theme,
	created_at,
	updated_at
FROM
	cookie_banners
WHERE
	id = @banner_id
	AND state = 'PUBLISHED'
LIMIT 1;
`

	args := pgx.StrictNamedArgs{"banner_id": bannerID}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query cookie banners: %w", err)
	}

	banner, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CookieBanner])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect cookie banner: %w", err)
	}

	*b = banner

	return nil
}

func (b *CookieBanners) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[CookieBannerOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	name,
	domain,
	state,
	title,
	description,
	accept_all_label,
	reject_all_label,
	save_preferences_label,
	privacy_policy_url,
	consent_expiry_days,
	consent_mode,
	version,
	theme,
	created_at,
	updated_at
FROM
	cookie_banners
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
		return fmt.Errorf("cannot query cookie banners: %w", err)
	}

	banners, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CookieBanner])
	if err != nil {
		return fmt.Errorf("cannot collect cookie banners: %w", err)
	}

	*b = banners

	return nil
}

func (b *CookieBanners) CountByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	cookie_banners
WHERE
	%s
	AND organization_id = @organization_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (b *CookieBanner) Insert(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
INSERT INTO cookie_banners (
	id,
	tenant_id,
	organization_id,
	name,
	domain,
	state,
	title,
	description,
	accept_all_label,
	reject_all_label,
	save_preferences_label,
	privacy_policy_url,
	consent_expiry_days,
	consent_mode,
	version,
	theme,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@name,
	@domain,
	@state,
	@title,
	@description,
	@accept_all_label,
	@reject_all_label,
	@save_preferences_label,
	@privacy_policy_url,
	@consent_expiry_days,
	@consent_mode,
	@version,
	@theme,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                     b.ID,
		"tenant_id":              scope.GetTenantID(),
		"organization_id":        b.OrganizationID,
		"name":                   b.Name,
		"domain":                 b.Domain,
		"state":                  b.State,
		"title":                  b.Title,
		"description":            b.Description,
		"accept_all_label":       b.AcceptAllLabel,
		"reject_all_label":       b.RejectAllLabel,
		"save_preferences_label": b.SavePreferencesLabel,
		"privacy_policy_url":     b.PrivacyPolicyURL,
		"consent_expiry_days":    b.ConsentExpiryDays,
		"consent_mode":           b.ConsentMode,
		"version":                b.Version,
		"theme":                  b.Theme,
		"created_at":             b.CreatedAt,
		"updated_at":             b.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert cookie banner: %w", err)
	}

	return nil
}

func (b *CookieBanner) Update(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
UPDATE cookie_banners
SET
	name = @name,
	domain = @domain,
	state = @state,
	title = @title,
	description = @description,
	accept_all_label = @accept_all_label,
	reject_all_label = @reject_all_label,
	save_preferences_label = @save_preferences_label,
	privacy_policy_url = @privacy_policy_url,
	consent_expiry_days = @consent_expiry_days,
	consent_mode = @consent_mode,
	version = @version,
	theme = @theme,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                     b.ID,
		"name":                   b.Name,
		"domain":                 b.Domain,
		"state":                  b.State,
		"title":                  b.Title,
		"description":            b.Description,
		"accept_all_label":       b.AcceptAllLabel,
		"reject_all_label":       b.RejectAllLabel,
		"save_preferences_label": b.SavePreferencesLabel,
		"privacy_policy_url":     b.PrivacyPolicyURL,
		"consent_expiry_days":    b.ConsentExpiryDays,
		"consent_mode":           b.ConsentMode,
		"version":                b.Version,
		"theme":                  b.Theme,
		"updated_at":             b.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update cookie banner: %w", err)
	}

	return nil
}

func (b *CookieBanner) Delete(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
DELETE FROM cookie_banners
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": b.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete cookie banner: %w", err)
	}

	return nil
}
