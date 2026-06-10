// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	CommonThirdParty struct {
		ID                            gid.GID            `db:"id"`
		Name                          string             `db:"name"`
		Slug                          string             `db:"slug"`
		Category                      ThirdPartyCategory `db:"category"`
		HeadquarterAddress            *string            `db:"headquarter_address"`
		LegalName                     *string            `db:"legal_name"`
		WebsiteURL                    *string            `db:"website_url"`
		PrivacyPolicyURL              *string            `db:"privacy_policy_url"`
		ServiceLevelAgreementURL      *string            `db:"service_level_agreement_url"`
		ServiceSoftwareAgreementURL   *string            `db:"service_software_agreement_url"`
		DataProcessingAgreementURL    *string            `db:"data_processing_agreement_url"`
		BusinessAssociateAgreementURL *string            `db:"business_associate_agreement_url"`
		SubprocessorsListURL          *string            `db:"subprocessors_list_url"`
		Certifications                []string           `db:"certifications"`
		StatusPageURL                 *string            `db:"status_page_url"`
		TermsOfServiceURL             *string            `db:"terms_of_service_url"`
		SecurityPageURL               *string            `db:"security_page_url"`
		TrustPageURL                  *string            `db:"trust_page_url"`
		LogoFileID                    *gid.GID           `db:"logo_file_id"`
		EnrichmentRequestedAt         *time.Time         `db:"enrichment_requested_at"`
		Enrichment                    json.RawMessage    `db:"enrichment"`
		EnrichmentAttempts            int                `db:"enrichment_attempts"`
		CreatedAt                     time.Time          `db:"created_at"`
		UpdatedAt                     time.Time          `db:"updated_at"`
	}

	CommonThirdParties []*CommonThirdParty
)

// AuthorizationAttributes loads existence-only attributes for the global
// common third-party catalog: rows have no organization_id, and the
// identity-scoped policy that grants access has no condition. The
// authorizer still requires an entry per requested ID (missing entries are
// treated as ErrResourceNotFound), so this verifies existence and returns
// empty attribute maps for every row that exists.
func (t *CommonThirdParty) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id FROM common_third_parties WHERE id = ANY(@resource_ids::text[])`

	args := pgx.StrictNamedArgs{
		"resource_ids": resourceIDs,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query common third party authorization attributes: %w", err)
	}
	defer rows.Close()

	attrsByID := make(policy.AttributesByID)

	for rows.Next() {
		var id gid.GID

		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("cannot scan common third party authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate common third party authorization attributes: %w", err)
	}

	return attrsByID, nil
}

func (t *CommonThirdParty) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	id gid.GID,
) error {
	q := `
SELECT
    id,
    name,
    slug,
    category,
    headquarter_address,
    legal_name,
    website_url,
    privacy_policy_url,
    service_level_agreement_url,
    service_software_agreement_url,
    data_processing_agreement_url,
    business_associate_agreement_url,
    subprocessors_list_url,
    certifications,
    status_page_url,
    terms_of_service_url,
    security_page_url,
    trust_page_url,
    logo_file_id,
    enrichment_requested_at,
    enrichment,
    enrichment_attempts,
    created_at,
    updated_at
FROM
    common_third_parties
WHERE
    id = @id
LIMIT 1;
`

	args := pgx.StrictNamedArgs{"id": id}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query common third party: %w", err)
	}
	defer rows.Close()

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CommonThirdParty])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect common third party: %w", err)
	}

	*t = row

	return nil
}

func (t *CommonThirdParty) LoadByName(
	ctx context.Context,
	conn pg.Querier,
	name string,
) error {
	q := `
SELECT
    id,
    name,
    slug,
    category,
    headquarter_address,
    legal_name,
    website_url,
    privacy_policy_url,
    service_level_agreement_url,
    service_software_agreement_url,
    data_processing_agreement_url,
    business_associate_agreement_url,
    subprocessors_list_url,
    certifications,
    status_page_url,
    terms_of_service_url,
    security_page_url,
    trust_page_url,
    logo_file_id,
    enrichment_requested_at,
    enrichment,
    enrichment_attempts,
    created_at,
    updated_at
FROM
    common_third_parties
WHERE
    lower(name) = lower(@name)
LIMIT 1;
`

	args := pgx.StrictNamedArgs{"name": name}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query common third party by name: %w", err)
	}
	defer rows.Close()

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CommonThirdParty])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect common third party by name: %w", err)
	}

	*t = row

	return nil
}

func (t *CommonThirdParty) LoadBySlug(
	ctx context.Context,
	conn pg.Querier,
	slug string,
) error {
	q := `
SELECT
    id,
    name,
    slug,
    category,
    headquarter_address,
    legal_name,
    website_url,
    privacy_policy_url,
    service_level_agreement_url,
    service_software_agreement_url,
    data_processing_agreement_url,
    business_associate_agreement_url,
    subprocessors_list_url,
    certifications,
    status_page_url,
    terms_of_service_url,
    security_page_url,
    trust_page_url,
    logo_file_id,
    enrichment_requested_at,
    enrichment,
    enrichment_attempts,
    created_at,
    updated_at
FROM
    common_third_parties
WHERE
    slug = @slug
LIMIT 1;
`

	args := pgx.StrictNamedArgs{"slug": slug}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query common third party by slug: %w", err)
	}
	defer rows.Close()

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CommonThirdParty])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect common third party by slug: %w", err)
	}

	*t = row

	return nil
}

func (t CommonThirdParty) Insert(
	ctx context.Context,
	conn pg.Tx,
) error {
	q := `
INSERT INTO common_third_parties (
    id,
    name,
    slug,
    category,
    headquarter_address,
    legal_name,
    website_url,
    privacy_policy_url,
    service_level_agreement_url,
    service_software_agreement_url,
    data_processing_agreement_url,
    business_associate_agreement_url,
    subprocessors_list_url,
    certifications,
    status_page_url,
    terms_of_service_url,
    security_page_url,
    trust_page_url,
    logo_file_id,
    enrichment_requested_at,
    enrichment,
    enrichment_attempts,
    created_at,
    updated_at
) VALUES (
    @id,
    @name,
    @slug,
    @category,
    @headquarter_address,
    @legal_name,
    @website_url,
    @privacy_policy_url,
    @service_level_agreement_url,
    @service_software_agreement_url,
    @data_processing_agreement_url,
    @business_associate_agreement_url,
    @subprocessors_list_url,
    @certifications,
    @status_page_url,
    @terms_of_service_url,
    @security_page_url,
    @trust_page_url,
    @logo_file_id,
    @enrichment_requested_at,
    @enrichment,
    @enrichment_attempts,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                               t.ID,
		"name":                             t.Name,
		"slug":                             t.Slug,
		"category":                         t.Category,
		"headquarter_address":              t.HeadquarterAddress,
		"legal_name":                       t.LegalName,
		"website_url":                      t.WebsiteURL,
		"privacy_policy_url":               t.PrivacyPolicyURL,
		"service_level_agreement_url":      t.ServiceLevelAgreementURL,
		"service_software_agreement_url":   t.ServiceSoftwareAgreementURL,
		"data_processing_agreement_url":    t.DataProcessingAgreementURL,
		"business_associate_agreement_url": t.BusinessAssociateAgreementURL,
		"subprocessors_list_url":           t.SubprocessorsListURL,
		"certifications":                   t.Certifications,
		"status_page_url":                  t.StatusPageURL,
		"terms_of_service_url":             t.TermsOfServiceURL,
		"security_page_url":                t.SecurityPageURL,
		"trust_page_url":                   t.TrustPageURL,
		"logo_file_id":                     t.LogoFileID,
		"enrichment_requested_at":          t.EnrichmentRequestedAt,
		"enrichment":                       t.Enrichment,
		"enrichment_attempts":              t.EnrichmentAttempts,
		"created_at":                       t.CreatedAt,
		"updated_at":                       t.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert common third party: %w", err)
	}

	return nil
}

// Upsert inserts a row, or on slug conflict updates every column except
// id and created_at. Returns true if a new row was inserted, false if an
// existing row was updated.
func (t *CommonThirdParty) Upsert(
	ctx context.Context,
	conn pg.Tx,
) (inserted bool, err error) {
	q := `
INSERT INTO common_third_parties (
    id,
    name,
    slug,
    category,
    headquarter_address,
    legal_name,
    website_url,
    privacy_policy_url,
    service_level_agreement_url,
    service_software_agreement_url,
    data_processing_agreement_url,
    business_associate_agreement_url,
    subprocessors_list_url,
    certifications,
    status_page_url,
    terms_of_service_url,
    security_page_url,
    trust_page_url,
    logo_file_id,
    enrichment_requested_at,
    enrichment,
    enrichment_attempts,
    created_at,
    updated_at
) VALUES (
    @id,
    @name,
    @slug,
    @category,
    @headquarter_address,
    @legal_name,
    @website_url,
    @privacy_policy_url,
    @service_level_agreement_url,
    @service_software_agreement_url,
    @data_processing_agreement_url,
    @business_associate_agreement_url,
    @subprocessors_list_url,
    @certifications,
    @status_page_url,
    @terms_of_service_url,
    @security_page_url,
    @trust_page_url,
    @logo_file_id,
    @enrichment_requested_at,
    @enrichment,
    @enrichment_attempts,
    @created_at,
    @updated_at
)
ON CONFLICT (slug) DO UPDATE
SET
    name                             = EXCLUDED.name,
    category                         = EXCLUDED.category,
    headquarter_address              = EXCLUDED.headquarter_address,
    legal_name                       = EXCLUDED.legal_name,
    website_url                      = EXCLUDED.website_url,
    privacy_policy_url               = EXCLUDED.privacy_policy_url,
    service_level_agreement_url      = EXCLUDED.service_level_agreement_url,
    service_software_agreement_url   = EXCLUDED.service_software_agreement_url,
    data_processing_agreement_url    = EXCLUDED.data_processing_agreement_url,
    business_associate_agreement_url = EXCLUDED.business_associate_agreement_url,
    subprocessors_list_url           = EXCLUDED.subprocessors_list_url,
    certifications                   = EXCLUDED.certifications,
    status_page_url                  = EXCLUDED.status_page_url,
    terms_of_service_url             = EXCLUDED.terms_of_service_url,
    security_page_url                = EXCLUDED.security_page_url,
    trust_page_url                   = EXCLUDED.trust_page_url,
    updated_at                       = EXCLUDED.updated_at
RETURNING
    id,
    name,
    slug,
    category,
    headquarter_address,
    legal_name,
    website_url,
    privacy_policy_url,
    service_level_agreement_url,
    service_software_agreement_url,
    data_processing_agreement_url,
    business_associate_agreement_url,
    subprocessors_list_url,
    certifications,
    status_page_url,
    terms_of_service_url,
    security_page_url,
    trust_page_url,
    logo_file_id,
    enrichment_requested_at,
    enrichment,
    enrichment_attempts,
    created_at,
    updated_at
`

	originalID := t.ID

	args := pgx.StrictNamedArgs{
		"id":                               t.ID,
		"name":                             t.Name,
		"slug":                             t.Slug,
		"category":                         t.Category,
		"headquarter_address":              t.HeadquarterAddress,
		"legal_name":                       t.LegalName,
		"website_url":                      t.WebsiteURL,
		"privacy_policy_url":               t.PrivacyPolicyURL,
		"service_level_agreement_url":      t.ServiceLevelAgreementURL,
		"service_software_agreement_url":   t.ServiceSoftwareAgreementURL,
		"data_processing_agreement_url":    t.DataProcessingAgreementURL,
		"business_associate_agreement_url": t.BusinessAssociateAgreementURL,
		"subprocessors_list_url":           t.SubprocessorsListURL,
		"certifications":                   t.Certifications,
		"status_page_url":                  t.StatusPageURL,
		"terms_of_service_url":             t.TermsOfServiceURL,
		"security_page_url":                t.SecurityPageURL,
		"trust_page_url":                   t.TrustPageURL,
		"logo_file_id":                     t.LogoFileID,
		"enrichment_requested_at":          t.EnrichmentRequestedAt,
		"enrichment":                       t.Enrichment,
		"enrichment_attempts":              t.EnrichmentAttempts,
		"created_at":                       t.CreatedAt,
		"updated_at":                       t.UpdatedAt,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot upsert common third party: %w", err)
	}
	defer rows.Close()

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CommonThirdParty])
	if err != nil {
		return false, fmt.Errorf("cannot collect upsert result: %w", err)
	}

	*t = row

	return originalID == t.ID, nil
}

func (t CommonThirdParty) Delete(
	ctx context.Context,
	conn pg.Tx,
	id gid.GID,
) error {
	q := `DELETE FROM common_third_parties WHERE id = @id`

	args := pgx.StrictNamedArgs{"id": id}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete common third party: %w", err)
	}

	return nil
}

func (t *CommonThirdParties) LoadByIDs(
	ctx context.Context,
	conn pg.Querier,
	ids []gid.GID,
) error {
	q := `
SELECT
    id,
    name,
    slug,
    category,
    headquarter_address,
    legal_name,
    website_url,
    privacy_policy_url,
    service_level_agreement_url,
    service_software_agreement_url,
    data_processing_agreement_url,
    business_associate_agreement_url,
    subprocessors_list_url,
    certifications,
    status_page_url,
    terms_of_service_url,
    security_page_url,
    trust_page_url,
    logo_file_id,
    enrichment_requested_at,
    enrichment,
    enrichment_attempts,
    created_at,
    updated_at
FROM
    common_third_parties
WHERE
    id = ANY(@ids)
`

	args := pgx.StrictNamedArgs{"ids": ids}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query common third parties: %w", err)
	}

	parties, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CommonThirdParty])
	if err != nil {
		return fmt.Errorf("cannot collect common third parties: %w", err)
	}

	*t = parties

	return nil
}

func (t *CommonThirdParties) LoadAll(
	ctx context.Context,
	conn pg.Querier,
	filter *CommonThirdPartyFilter,
) error {
	q := `
SELECT
    id,
    name,
    slug,
    category,
    headquarter_address,
    legal_name,
    website_url,
    privacy_policy_url,
    service_level_agreement_url,
    service_software_agreement_url,
    data_processing_agreement_url,
    business_associate_agreement_url,
    subprocessors_list_url,
    certifications,
    status_page_url,
    terms_of_service_url,
    security_page_url,
    trust_page_url,
    logo_file_id,
    enrichment_requested_at,
    enrichment,
    enrichment_attempts,
    created_at,
    updated_at
FROM
    common_third_parties
WHERE
    %s
ORDER BY name ASC
LIMIT 20
`

	q = fmt.Sprintf(q, filter.SQLFragment())

	args := pgx.StrictNamedArgs{}
	maps.Copy(args, filter.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query common third parties: %w", err)
	}

	parties, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CommonThirdParty])
	if err != nil {
		return fmt.Errorf("cannot collect common third parties: %w", err)
	}

	*t = parties

	return nil
}

func (t CommonThirdParty) UpdateLogoFileID(
	ctx context.Context,
	conn pg.Tx,
) error {
	q := `
UPDATE common_third_parties
SET
    logo_file_id = @logo_file_id,
    updated_at   = @updated_at
WHERE
    id = @id
`

	args := pgx.StrictNamedArgs{
		"id":           t.ID,
		"logo_file_id": t.LogoFileID,
		"updated_at":   t.UpdatedAt,
	}

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update common third party logo: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (t *CommonThirdParty) CursorKey(field CommonThirdPartyOrderField) page.CursorKey {
	switch field {
	case CommonThirdPartyOrderFieldName:
		return page.NewCursorKey(t.ID, t.Name)
	case CommonThirdPartyOrderFieldCreatedAt:
		return page.NewCursorKey(t.ID, t.CreatedAt)
	case CommonThirdPartyOrderFieldUpdatedAt:
		return page.NewCursorKey(t.ID, t.UpdatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

// Load returns a cursor-paginated, filtered page of common third
// parties. The catalog is global (no tenant scope); the cursor supplies
// the limit and ordering. Unlike LoadAll (capped at 20, name only), this
// is the listing entry point a future API/CLI consumes.
func (t *CommonThirdParties) Load(
	ctx context.Context,
	conn pg.Querier,
	cursor *page.Cursor[CommonThirdPartyOrderField],
	filter *CommonThirdPartyFilter,
) error {
	q := `
SELECT
    id,
    name,
    slug,
    category,
    headquarter_address,
    legal_name,
    website_url,
    privacy_policy_url,
    service_level_agreement_url,
    service_software_agreement_url,
    data_processing_agreement_url,
    business_associate_agreement_url,
    subprocessors_list_url,
    certifications,
    status_page_url,
    terms_of_service_url,
    security_page_url,
    trust_page_url,
    logo_file_id,
    enrichment_requested_at,
    enrichment,
    enrichment_attempts,
    created_at,
    updated_at
FROM
    common_third_parties
WHERE
    %s
    AND %s
`

	q = fmt.Sprintf(q, filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{}
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query common third parties: %w", err)
	}

	parties, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CommonThirdParty])
	if err != nil {
		return fmt.Errorf("cannot collect common third parties: %w", err)
	}

	*t = parties

	return nil
}

// CountAll returns the number of common third parties matching the
// filter, ignoring pagination.
func (t *CommonThirdParties) CountAll(
	ctx context.Context,
	conn pg.Querier,
	filter *CommonThirdPartyFilter,
) (int, error) {
	q := `
SELECT
    COUNT(id)
FROM
    common_third_parties
WHERE
    %s
`

	q = fmt.Sprintf(q, filter.SQLFragment())

	args := pgx.StrictNamedArgs{}
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count common third parties: %w", err)
	}

	return count, nil
}

// LoadNextForEnrichmentForUpdateSkipLocked claims the oldest row queued
// for enrichment. The global catalog is not tenant-scoped, so the claim
// is intentionally cross-tenant: the enrichment worker is a system
// worker that drains the queue regardless of tenant.
func (t *CommonThirdParty) LoadNextForEnrichmentForUpdateSkipLocked(
	ctx context.Context,
	tx pg.Tx,
) error {
	q := `
SELECT
    id,
    name,
    slug,
    category,
    headquarter_address,
    legal_name,
    website_url,
    privacy_policy_url,
    service_level_agreement_url,
    service_software_agreement_url,
    data_processing_agreement_url,
    business_associate_agreement_url,
    subprocessors_list_url,
    certifications,
    status_page_url,
    terms_of_service_url,
    security_page_url,
    trust_page_url,
    logo_file_id,
    enrichment_requested_at,
    enrichment,
    enrichment_attempts,
    created_at,
    updated_at
FROM
    common_third_parties
WHERE
    enrichment_requested_at IS NOT NULL
ORDER BY
    enrichment_requested_at ASC
FOR UPDATE SKIP LOCKED
LIMIT 1;
`

	rows, err := tx.Query(ctx, q)
	if err != nil {
		return fmt.Errorf("cannot query common third party for enrichment: %w", err)
	}
	defer rows.Close()

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CommonThirdParty])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect common third party for enrichment: %w", err)
	}

	*t = row

	return nil
}

// ClearEnrichmentRequestedAt removes the row from the enrichment queue
// and bumps the attempt counter. It bumps updated_at so the
// stale-recovery clock starts at claim time, keeping
// ResetStaleCommonThirdPartyEnrichments from re-arming a row that is
// still being processed. The attempt counter is incremented up front so
// a crash between claim and persist still counts against the retry
// budget.
func (t *CommonThirdParty) ClearEnrichmentRequestedAt(
	ctx context.Context,
	tx pg.Tx,
) error {
	q := `
UPDATE common_third_parties
SET
    enrichment_requested_at = NULL,
    enrichment_attempts = enrichment_attempts + 1,
    updated_at = NOW()
WHERE id = @id
`

	args := pgx.StrictNamedArgs{"id": t.ID}

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot clear enrichment requested at: %w", err)
	}

	t.EnrichmentRequestedAt = nil
	t.EnrichmentAttempts++

	return nil
}

// UpdateEnrichment persists the enrichment result: the resolved metadata
// fields plus the per-field enrichment provenance JSON. It is a targeted
// partial update that never touches id, slug, category, name, logo, the
// queue column, or the attempt counter (logo is owned by
// UpdateLogoFileID; the queue column and counter are managed by
// ClearEnrichmentRequestedAt). The caller decides which scalar fields to
// write versus leave untouched, then passes the merged receiver here.
func (t CommonThirdParty) UpdateEnrichment(
	ctx context.Context,
	conn pg.Tx,
) error {
	q := `
UPDATE common_third_parties
SET
    headquarter_address              = @headquarter_address,
    legal_name                       = @legal_name,
    website_url                      = @website_url,
    privacy_policy_url               = @privacy_policy_url,
    service_level_agreement_url      = @service_level_agreement_url,
    service_software_agreement_url   = @service_software_agreement_url,
    data_processing_agreement_url    = @data_processing_agreement_url,
    business_associate_agreement_url = @business_associate_agreement_url,
    subprocessors_list_url           = @subprocessors_list_url,
    certifications                   = @certifications,
    status_page_url                  = @status_page_url,
    terms_of_service_url             = @terms_of_service_url,
    security_page_url                = @security_page_url,
    trust_page_url                   = @trust_page_url,
    enrichment                       = @enrichment,
    updated_at                       = @updated_at
WHERE
    id = @id
`

	args := pgx.StrictNamedArgs{
		"id":                               t.ID,
		"headquarter_address":              t.HeadquarterAddress,
		"legal_name":                       t.LegalName,
		"website_url":                      t.WebsiteURL,
		"privacy_policy_url":               t.PrivacyPolicyURL,
		"service_level_agreement_url":      t.ServiceLevelAgreementURL,
		"service_software_agreement_url":   t.ServiceSoftwareAgreementURL,
		"data_processing_agreement_url":    t.DataProcessingAgreementURL,
		"business_associate_agreement_url": t.BusinessAssociateAgreementURL,
		"subprocessors_list_url":           t.SubprocessorsListURL,
		"certifications":                   t.Certifications,
		"status_page_url":                  t.StatusPageURL,
		"terms_of_service_url":             t.TermsOfServiceURL,
		"security_page_url":                t.SecurityPageURL,
		"trust_page_url":                   t.TrustPageURL,
		"enrichment":                       t.Enrichment,
		"updated_at":                       t.UpdatedAt,
	}

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update common third party enrichment: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

// ResetStaleCommonThirdPartyEnrichments re-arms enrichment_requested_at
// on rows whose enrichment was claimed but never completed and have been
// idle longer than staleAfter, so a crashed or timed-out run is retried.
//
// A claimed row has enrichment_attempts > 0 (Claim increments it) and a
// completed row has a non-null enrichment payload (Process always writes
// it, even on a no-result run), so the sweep targets rows that were
// claimed but carry no enrichment yet. Curated rows that were never
// enqueued keep enrichment_attempts = 0 and are left untouched. The
// max-attempts ceiling stops permanently failing rows from looping
// forever.
//
// Like the claim query, this sweep is intentionally cross-tenant: the
// enrichment worker is a system worker that drains the queue regardless
// of tenant.
func ResetStaleCommonThirdPartyEnrichments(
	ctx context.Context,
	conn pg.Querier,
	staleAfter time.Duration,
	maxAttempts int,
) error {
	q := `
UPDATE common_third_parties
SET
    enrichment_requested_at = NOW(),
    updated_at = NOW()
WHERE
    enrichment_requested_at IS NULL
    AND enrichment IS NULL
    AND enrichment_attempts > 0
    AND enrichment_attempts < @max_attempts
    AND updated_at < @stale_before
`

	args := pgx.StrictNamedArgs{
		"max_attempts": maxAttempts,
		"stale_before": time.Now().Add(-staleAfter),
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot reset stale common third party enrichments: %w", err)
	}

	return nil
}
