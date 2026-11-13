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
		ID                          gid.GID                  `db:"id"`
		OrganizationID              gid.GID                  `db:"organization_id"`
		CMPDomainID                 *gid.GID                 `db:"cmp_domain_id"`
		Name                        string                   `db:"name"`
		Origin                      string                   `db:"origin"`
		State                       CookieBannerState        `db:"state"`
		PrivacyPolicyURL            *string                  `db:"privacy_policy_url"`
		CookiePolicyURL             string                   `db:"cookie_policy_url"`
		ConsentExpiryDays           int                      `db:"consent_expiry_days"`
		ShowBranding                bool                     `db:"show_branding"`
		Capabilities                CookieBannerCapabilities `db:"capabilities"`
		ForceOptIn                  bool                     `db:"force_opt_in"`
		DefaultLanguage             string                   `db:"default_language"`
		PublisherCountryCode        string                   `db:"publisher_country_code"`
		PatternAnalysisRequestedAt  *time.Time               `db:"pattern_analysis_requested_at"`
		PolicyDocumentID            *gid.GID                 `db:"policy_document_id"`
		PolicyGenerationRequestedAt *time.Time               `db:"policy_generation_requested_at"`
		SourcePromotionRequestedAt  *time.Time               `db:"source_promotion_requested_at"`
		SourcePromotionSince        *time.Time               `db:"source_promotion_since"`
		SourcePromotionAttempts     int                      `db:"source_promotion_attempts"`
		SourcePromotionAttemptedAt  *time.Time               `db:"source_promotion_attempted_at"`
		CreatedAt                   time.Time                `db:"created_at"`
		UpdatedAt                   time.Time                `db:"updated_at"`
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
	cmp_domain_id,
	name,
	origin,
	state,
	privacy_policy_url,
	cookie_policy_url,
	consent_expiry_days,
	show_branding,
	capabilities,
	force_opt_in,
	default_language,
	publisher_country_code,
	pattern_analysis_requested_at,
	policy_document_id,
	policy_generation_requested_at,
	source_promotion_requested_at,
	source_promotion_since,
	source_promotion_attempts,
	source_promotion_attempted_at,
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
	cmp_domain_id,
	name,
	origin,
	state,
	privacy_policy_url,
	cookie_policy_url,
	consent_expiry_days,
	show_branding,
	capabilities,
	force_opt_in,
	default_language,
	publisher_country_code,
	pattern_analysis_requested_at,
	policy_document_id,
	policy_generation_requested_at,
	source_promotion_requested_at,
	source_promotion_since,
	source_promotion_attempts,
	source_promotion_attempted_at,
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
	cmp_domain_id,
	name,
	origin,
	state,
	privacy_policy_url,
	cookie_policy_url,
	consent_expiry_days,
	show_branding,
	capabilities,
	force_opt_in,
	default_language,
	publisher_country_code,
	pattern_analysis_requested_at,
	policy_document_id,
	policy_generation_requested_at,
	source_promotion_requested_at,
	source_promotion_since,
	source_promotion_attempts,
	source_promotion_attempted_at,
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
	cmp_domain_id,
	name,
	origin,
	state,
	privacy_policy_url,
	cookie_policy_url,
	consent_expiry_days,
	show_branding,
	capabilities,
	force_opt_in,
	default_language,
	publisher_country_code,
	pattern_analysis_requested_at,
	policy_document_id,
	policy_generation_requested_at,
	source_promotion_requested_at,
	source_promotion_since,
	source_promotion_attempts,
	source_promotion_attempted_at,
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

	if len(banners) != len(gid.NewSet(bannerIDs...)) {
		return ErrResourceNotFound
	}

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
	cmp_domain_id,
	name,
	origin,
	state,
	privacy_policy_url,
	cookie_policy_url,
	consent_expiry_days,
	show_branding,
	capabilities,
	force_opt_in,
	default_language,
	publisher_country_code,
	pattern_analysis_requested_at,
	policy_document_id,
	policy_generation_requested_at,
	source_promotion_requested_at,
	source_promotion_since,
	source_promotion_attempts,
	source_promotion_attempted_at,
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

func (b *CookieBanners) LoadByCMPDomainID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cmpDomainID gid.GID,
	cursor *page.Cursor[CookieBannerOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	cmp_domain_id,
	name,
	origin,
	state,
	privacy_policy_url,
	cookie_policy_url,
	consent_expiry_days,
	show_branding,
	capabilities,
	force_opt_in,
	default_language,
	publisher_country_code,
	pattern_analysis_requested_at,
	policy_document_id,
	policy_generation_requested_at,
	source_promotion_requested_at,
	source_promotion_since,
	source_promotion_attempts,
	source_promotion_attempted_at,
	created_at,
	updated_at
FROM
	cookie_banners
WHERE
	%s
	AND cmp_domain_id = @cmp_domain_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"cmp_domain_id": cmpDomainID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query cookie banners by cmp domain: %w", err)
	}

	banners, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CookieBanner])
	if err != nil {
		return fmt.Errorf("cannot collect cookie banners by cmp domain: %w", err)
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
	cmp_domain_id,
	name,
	origin,
	state,
	privacy_policy_url,
	cookie_policy_url,
	consent_expiry_days,
	show_branding,
	capabilities,
	force_opt_in,
	default_language,
	publisher_country_code,
	pattern_analysis_requested_at,
	policy_document_id,
	policy_generation_requested_at,
	source_promotion_requested_at,
	source_promotion_since,
	source_promotion_attempts,
	source_promotion_attempted_at,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@cmp_domain_id,
	@name,
	@origin,
	@state,
	@privacy_policy_url,
	@cookie_policy_url,
	@consent_expiry_days,
	@show_branding,
	@capabilities,
	@force_opt_in,
	@default_language,
	@publisher_country_code,
	@pattern_analysis_requested_at,
	@policy_document_id,
	@policy_generation_requested_at,
	@source_promotion_requested_at,
	@source_promotion_since,
	@source_promotion_attempts,
	@source_promotion_attempted_at,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                             b.ID,
		"tenant_id":                      scope.GetTenantID(),
		"organization_id":                b.OrganizationID,
		"cmp_domain_id":                  b.CMPDomainID,
		"name":                           b.Name,
		"origin":                         b.Origin,
		"state":                          b.State,
		"privacy_policy_url":             b.PrivacyPolicyURL,
		"cookie_policy_url":              b.CookiePolicyURL,
		"consent_expiry_days":            b.ConsentExpiryDays,
		"show_branding":                  b.ShowBranding,
		"capabilities":                   b.Capabilities,
		"force_opt_in":                   b.ForceOptIn,
		"default_language":               b.DefaultLanguage,
		"publisher_country_code":         b.PublisherCountryCode,
		"pattern_analysis_requested_at":  b.PatternAnalysisRequestedAt,
		"policy_document_id":             b.PolicyDocumentID,
		"policy_generation_requested_at": b.PolicyGenerationRequestedAt,
		"source_promotion_requested_at":  b.SourcePromotionRequestedAt,
		"source_promotion_since":         b.SourcePromotionSince,
		"source_promotion_attempts":      b.SourcePromotionAttempts,
		"source_promotion_attempted_at":  b.SourcePromotionAttemptedAt,
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
	capabilities = @capabilities,
	force_opt_in = @force_opt_in,
	default_language = @default_language,
	publisher_country_code = @publisher_country_code,
	policy_document_id = @policy_document_id,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                     b.ID,
		"name":                   b.Name,
		"state":                  b.State,
		"privacy_policy_url":     b.PrivacyPolicyURL,
		"cookie_policy_url":      b.CookiePolicyURL,
		"consent_expiry_days":    b.ConsentExpiryDays,
		"show_branding":          b.ShowBranding,
		"capabilities":           b.Capabilities,
		"force_opt_in":           b.ForceOptIn,
		"default_language":       b.DefaultLanguage,
		"publisher_country_code": b.PublisherCountryCode,
		"policy_document_id":     b.PolicyDocumentID,
		"updated_at":             b.UpdatedAt,
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
	cmp_domain_id,
	name,
	origin,
	state,
	privacy_policy_url,
	cookie_policy_url,
	consent_expiry_days,
	show_branding,
	capabilities,
	force_opt_in,
	default_language,
	publisher_country_code,
	pattern_analysis_requested_at,
	policy_document_id,
	policy_generation_requested_at,
	source_promotion_requested_at,
	source_promotion_since,
	source_promotion_attempts,
	source_promotion_attempted_at,
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
	cmp_domain_id,
	name,
	origin,
	state,
	privacy_policy_url,
	cookie_policy_url,
	consent_expiry_days,
	show_branding,
	capabilities,
	force_opt_in,
	default_language,
	publisher_country_code,
	pattern_analysis_requested_at,
	policy_document_id,
	policy_generation_requested_at,
	source_promotion_requested_at,
	source_promotion_since,
	source_promotion_attempts,
	source_promotion_attempted_at,
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

// SetSourcePromotionRequested asks the banner source-promotion worker to
// compare this banner's recent cookie detections with its siblings.
// requested_at stays at the earliest report still waiting, so a later
// report does not hide rows written by an earlier one. A request that is
// already waiting keeps its attempt budget. A new request, including one
// after the budget was exhausted, starts the budget over.
func (b *CookieBanner) SetSourcePromotionRequested(
	ctx context.Context,
	tx pg.Tx,
	now time.Time,
) error {
	q := `
UPDATE cookie_banners
SET
	source_promotion_attempts = CASE
		WHEN source_promotion_requested_at IS NULL THEN 0
		ELSE source_promotion_attempts
	END,
	source_promotion_attempted_at = CASE
		WHEN source_promotion_requested_at IS NULL THEN NULL
		ELSE source_promotion_attempted_at
	END,
	source_promotion_requested_at = LEAST(COALESCE(source_promotion_requested_at, @now), @now),
	updated_at = NOW()
WHERE id = @id
`

	args := pgx.StrictNamedArgs{
		"id":  b.ID,
		"now": now,
	}

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot request cookie banner source promotion: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

// ClearSourcePromotionRequested removes the banner from the queue and
// counts the attempt. since keeps the window start the claim read, so a
// retry still sees those detections. The attempt is recorded before the
// work runs, so a crash between claim and completion still spends budget.
func (b *CookieBanner) ClearSourcePromotionRequested(
	ctx context.Context,
	tx pg.Tx,
) error {
	q := `
UPDATE cookie_banners
SET
	source_promotion_since = source_promotion_requested_at,
	source_promotion_requested_at = NULL,
	source_promotion_attempts = source_promotion_attempts + 1,
	source_promotion_attempted_at = NOW(),
	updated_at = NOW()
WHERE id = @id
RETURNING source_promotion_since, source_promotion_attempts, source_promotion_attempted_at
`

	args := pgx.StrictNamedArgs{"id": b.ID}

	var (
		since       *time.Time
		attempts    int
		attemptedAt *time.Time
	)

	err := tx.QueryRow(ctx, q, args).Scan(&since, &attempts, &attemptedAt)
	if err != nil {
		return fmt.Errorf("cannot clear cookie banner source promotion: %w", err)
	}

	b.SourcePromotionSince = since
	b.SourcePromotionRequestedAt = nil
	b.SourcePromotionAttempts = attempts
	b.SourcePromotionAttemptedAt = attemptedAt

	return nil
}

// RequeueSourcePromotionRequested puts a failed attempt back on the queue
// without resetting the budget. The window starts at the claimed since,
// unless a newer request is already earlier. The update matches the claim
// that failed, so an older task cannot queue the banner after a newer
// claim has taken it. A superseded claim is a no-op.
func (b *CookieBanner) RequeueSourcePromotionRequested(
	ctx context.Context,
	tx pg.Tx,
) error {
	q := `
UPDATE cookie_banners
SET
	source_promotion_requested_at = LEAST(
		COALESCE(source_promotion_requested_at, source_promotion_since),
		source_promotion_since
	),
	updated_at = NOW()
WHERE
	id = @id
	AND source_promotion_attempts = @source_promotion_attempts
	AND source_promotion_attempted_at IS NOT DISTINCT FROM @source_promotion_attempted_at
`

	args := pgx.StrictNamedArgs{
		"id":                            b.ID,
		"source_promotion_attempts":     b.SourcePromotionAttempts,
		"source_promotion_attempted_at": b.SourcePromotionAttemptedAt,
	}

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot requeue cookie banner source promotion: %w", err)
	}

	return nil
}

// ResetSourcePromotionAttempts clears the budget after a promotion finishes.
// It matches the claim that finished, so a late success cannot wipe the
// budget of a newer claim. A superseded claim is a no-op.
func (b *CookieBanner) ResetSourcePromotionAttempts(
	ctx context.Context,
	tx pg.Tx,
) error {
	q := `
UPDATE cookie_banners
SET
	source_promotion_attempts = 0,
	source_promotion_attempted_at = NULL,
	updated_at = NOW()
WHERE
	id = @id
	AND source_promotion_attempts = @source_promotion_attempts
	AND source_promotion_attempted_at IS NOT DISTINCT FROM @source_promotion_attempted_at
`

	args := pgx.StrictNamedArgs{
		"id":                            b.ID,
		"source_promotion_attempts":     b.SourcePromotionAttempts,
		"source_promotion_attempted_at": b.SourcePromotionAttemptedAt,
	}

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot reset cookie banner source promotion attempts: %w", err)
	}

	if result.RowsAffected() == 0 {
		return nil
	}

	b.SourcePromotionAttempts = 0
	b.SourcePromotionAttemptedAt = nil

	return nil
}

// ResetStaleSourcePromotions requeues a claim that was taken but never
// finished and has been idle longer than staleAfter. The window stays at
// the claimed since. A finished promotion has a zero attempt count, and a
// claim at the cap stays clear.
func ResetStaleSourcePromotions(
	ctx context.Context,
	conn pg.Querier,
	staleAfter time.Duration,
	maxAttempts int,
) error {
	q := `
UPDATE cookie_banners
SET
	source_promotion_requested_at = LEAST(
		COALESCE(source_promotion_requested_at, source_promotion_since),
		source_promotion_since
	),
	updated_at = NOW()
WHERE
	source_promotion_requested_at IS NULL
	AND source_promotion_attempts > 0
	AND source_promotion_attempts < @max_attempts
	AND source_promotion_attempted_at < @stale_before
`

	args := pgx.StrictNamedArgs{
		"max_attempts": maxAttempts,
		"stale_before": time.Now().Add(-staleAfter),
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot reset stale cookie banner source promotions: %w", err)
	}

	return nil
}

func (b *CookieBanner) LoadNextForSourcePromotionForUpdateSkipLocked(
	ctx context.Context,
	tx pg.Tx,
) error {
	q := `
SELECT
	id,
	organization_id,
	cmp_domain_id,
	name,
	origin,
	state,
	privacy_policy_url,
	cookie_policy_url,
	consent_expiry_days,
	show_branding,
	capabilities,
	force_opt_in,
	default_language,
	publisher_country_code,
	pattern_analysis_requested_at,
	policy_document_id,
	policy_generation_requested_at,
	source_promotion_requested_at,
	source_promotion_since,
	source_promotion_attempts,
	source_promotion_attempted_at,
	created_at,
	updated_at
FROM
	cookie_banners
WHERE source_promotion_requested_at IS NOT NULL
ORDER BY source_promotion_requested_at ASC, id ASC
FOR UPDATE SKIP LOCKED
LIMIT 1
`

	rows, err := tx.Query(ctx, q)
	if err != nil {
		return fmt.Errorf("cannot query cookie banners for source promotion: %w", err)
	}

	banner, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CookieBanner])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect cookie banner for source promotion: %w", err)
	}

	*b = banner

	return nil
}
