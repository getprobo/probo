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
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	CookieBanner struct {
		ID                          gid.GID           `db:"id"`
		OrganizationID              gid.GID           `db:"organization_id"`
		Name                        string            `db:"name"`
		Origin                      string            `db:"origin"`
		State                       CookieBannerState `db:"state"`
		PrivacyPolicyURL            *string           `db:"privacy_policy_url"`
		CookiePolicyURL             string            `db:"cookie_policy_url"`
		ConsentExpiryDays           int               `db:"consent_expiry_days"`
		ShowBranding                bool              `db:"show_branding"`
		DefaultLanguage             string            `db:"default_language"`
		PatternAnalysisRequestedAt  *time.Time        `db:"pattern_analysis_requested_at"`
		PolicyDocumentID            *gid.GID          `db:"policy_document_id"`
		PolicyGenerationRequestedAt *time.Time        `db:"policy_generation_requested_at"`
		CreatedAt                   time.Time         `db:"created_at"`
		UpdatedAt                   time.Time         `db:"updated_at"`
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

func (b *CookieBanner) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM cookie_banners WHERE id = ANY(@resource_ids::text[])`

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
	origin,
	state,
	privacy_policy_url,
	cookie_policy_url,
	consent_expiry_days,
	show_branding,
	default_language,
	pattern_analysis_requested_at,
	policy_document_id,
	policy_generation_requested_at,
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

func (b *CookieBanner) LoadActiveByID(
	ctx context.Context,
	conn pg.Querier,
	bannerID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	name,
	origin,
	state,
	privacy_policy_url,
	cookie_policy_url,
	consent_expiry_days,
	show_branding,
	default_language,
	pattern_analysis_requested_at,
	policy_document_id,
	policy_generation_requested_at,
	created_at,
	updated_at
FROM
	cookie_banners
WHERE
	id = @banner_id
	AND state = @state
LIMIT 1;
`

	args := pgx.StrictNamedArgs{
		"banner_id": bannerID,
		"state":     CookieBannerStateActive,
	}

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

func (b *CookieBanner) LoadActiveByOrigin(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	origin string,
) error {
	q := `
SELECT
	id,
	organization_id,
	name,
	origin,
	state,
	privacy_policy_url,
	cookie_policy_url,
	consent_expiry_days,
	show_branding,
	default_language,
	pattern_analysis_requested_at,
	policy_document_id,
	policy_generation_requested_at,
	created_at,
	updated_at
FROM
	cookie_banners
WHERE
	%s
	AND origin = @origin
	AND state = @state
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"origin": origin,
		"state":  CookieBannerStateActive,
	}
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

func (b *CookieBanners) LoadByIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	bannerIDs []gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	name,
	origin,
	state,
	privacy_policy_url,
	cookie_policy_url,
	consent_expiry_days,
	show_branding,
	default_language,
	pattern_analysis_requested_at,
	policy_document_id,
	policy_generation_requested_at,
	created_at,
	updated_at
FROM
	cookie_banners
WHERE
	%s
	AND id = ANY(@banner_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"banner_ids": bannerIDs}
	maps.Copy(args, scope.SQLArguments())

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

func (b *CookieBanners) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[CookieBannerOrderField],
	filter *CookieBannerFilter,
) error {
	q := `
SELECT
	id,
	organization_id,
	name,
	origin,
	state,
	privacy_policy_url,
	cookie_policy_url,
	consent_expiry_days,
	show_branding,
	default_language,
	pattern_analysis_requested_at,
	policy_document_id,
	policy_generation_requested_at,
	created_at,
	updated_at
FROM
	cookie_banners
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
	filter *CookieBannerFilter,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	cookie_banners
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
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (b *CookieBanner) Insert(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO cookie_banners (
	id,
	tenant_id,
	organization_id,
	name,
	origin,
	state,
	privacy_policy_url,
	cookie_policy_url,
	consent_expiry_days,
	show_branding,
	default_language,
	pattern_analysis_requested_at,
	policy_document_id,
	policy_generation_requested_at,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@name,
	@origin,
	@state,
	@privacy_policy_url,
	@cookie_policy_url,
	@consent_expiry_days,
	@show_branding,
	@default_language,
	@pattern_analysis_requested_at,
	@policy_document_id,
	@policy_generation_requested_at,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                             b.ID,
		"tenant_id":                      scope.GetTenantID(),
		"organization_id":                b.OrganizationID,
		"name":                           b.Name,
		"origin":                         b.Origin,
		"state":                          b.State,
		"privacy_policy_url":             b.PrivacyPolicyURL,
		"cookie_policy_url":              b.CookiePolicyURL,
		"consent_expiry_days":            b.ConsentExpiryDays,
		"show_branding":                  b.ShowBranding,
		"default_language":               b.DefaultLanguage,
		"pattern_analysis_requested_at":  b.PatternAnalysisRequestedAt,
		"policy_document_id":             b.PolicyDocumentID,
		"policy_generation_requested_at": b.PolicyGenerationRequestedAt,
		"created_at":                     b.CreatedAt,
		"updated_at":                     b.UpdatedAt,
	}

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "idx_cookie_banners_unique_active_origin" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert cookie banner: %w", err)
	}

	return nil
}

func (b *CookieBanner) Update(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE cookie_banners
SET
	name = @name,
	state = @state,
	privacy_policy_url = @privacy_policy_url,
	cookie_policy_url = @cookie_policy_url,
	consent_expiry_days = @consent_expiry_days,
	show_branding = @show_branding,
	default_language = @default_language,
	policy_document_id = @policy_document_id,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                  b.ID,
		"name":                b.Name,
		"state":               b.State,
		"privacy_policy_url":  b.PrivacyPolicyURL,
		"cookie_policy_url":   b.CookiePolicyURL,
		"consent_expiry_days": b.ConsentExpiryDays,
		"show_branding":       b.ShowBranding,
		"default_language":    b.DefaultLanguage,
		"policy_document_id":  b.PolicyDocumentID,
		"updated_at":          b.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "idx_cookie_banners_unique_active_origin" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot update cookie banner: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (b *CookieBanner) UpdateShowBranding(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	show bool,
) error {
	q := `
UPDATE cookie_banners
SET
	show_branding = @show_branding,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":            b.ID,
		"show_branding": show,
		"updated_at":    time.Now(),
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update cookie banner show_branding: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	b.ShowBranding = show

	return nil
}

func (b *CookieBanner) Delete(
	ctx context.Context,
	tx pg.Tx,
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

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete cookie banner: %w", err)
	}

	return nil
}

func (b *CookieBanner) LoadNextForPatternAnalysisForUpdateSkipLocked(
	ctx context.Context,
	tx pg.Tx,
) error {
	q := `
SELECT
	id,
	organization_id,
	name,
	origin,
	state,
	privacy_policy_url,
	cookie_policy_url,
	consent_expiry_days,
	show_branding,
	default_language,
	pattern_analysis_requested_at,
	policy_document_id,
	policy_generation_requested_at,
	created_at,
	updated_at
FROM
	cookie_banners
WHERE
	pattern_analysis_requested_at IS NOT NULL
ORDER BY
	pattern_analysis_requested_at ASC
FOR UPDATE SKIP LOCKED
LIMIT 1;
`

	rows, err := tx.Query(ctx, q)
	if err != nil {
		return fmt.Errorf("cannot query cookie banners for pattern analysis: %w", err)
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

func (b *CookieBanner) ClearPatternAnalysisRequestedAt(
	ctx context.Context,
	tx pg.Tx,
) error {
	q := `
UPDATE cookie_banners
SET pattern_analysis_requested_at = NULL
WHERE id = @id
`

	args := pgx.StrictNamedArgs{"id": b.ID}

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot clear pattern analysis requested at: %w", err)
	}

	b.PatternAnalysisRequestedAt = nil

	return nil
}

func (b *CookieBanner) SetPatternAnalysisRequested(
	ctx context.Context,
	tx pg.Tx,
) error {
	q := `
UPDATE cookie_banners
SET pattern_analysis_requested_at = NOW()
WHERE id = @id
  AND pattern_analysis_requested_at IS NULL
`

	args := pgx.StrictNamedArgs{"id": b.ID}

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot set pattern analysis requested: %w", err)
	}

	return nil
}

func (b *CookieBanner) LoadNextForPolicyGenerationForUpdateSkipLocked(
	ctx context.Context,
	tx pg.Tx,
) error {
	q := `
SELECT
	id,
	organization_id,
	name,
	origin,
	state,
	privacy_policy_url,
	cookie_policy_url,
	consent_expiry_days,
	show_branding,
	default_language,
	pattern_analysis_requested_at,
	policy_document_id,
	policy_generation_requested_at,
	created_at,
	updated_at
FROM
	cookie_banners
WHERE
	policy_generation_requested_at IS NOT NULL
ORDER BY
	policy_generation_requested_at ASC
FOR UPDATE SKIP LOCKED
LIMIT 1;
`

	rows, err := tx.Query(ctx, q)
	if err != nil {
		return fmt.Errorf("cannot query cookie banners for policy generation: %w", err)
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

func (b *CookieBanner) ClearPolicyGenerationRequestedAt(
	ctx context.Context,
	tx pg.Tx,
) error {
	q := `
UPDATE cookie_banners
SET policy_generation_requested_at = NULL
WHERE id = @id
`

	args := pgx.StrictNamedArgs{"id": b.ID}

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot clear policy generation requested at: %w", err)
	}

	b.PolicyGenerationRequestedAt = nil

	return nil
}

func (b *CookieBanner) SetPolicyGenerationRequested(
	ctx context.Context,
	tx pg.Tx,
) error {
	q := `
UPDATE cookie_banners
SET policy_generation_requested_at = NOW()
WHERE id = @id
  AND policy_generation_requested_at IS NULL
`

	args := pgx.StrictNamedArgs{"id": b.ID}

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot set policy generation requested: %w", err)
	}

	return nil
}
