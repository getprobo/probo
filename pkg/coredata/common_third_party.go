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
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
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
		CreatedAt                     time.Time          `db:"created_at"`
		UpdatedAt                     time.Time          `db:"updated_at"`
	}

	CommonThirdParties []*CommonThirdParty
)

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
func (t CommonThirdParty) Upsert(
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
RETURNING (xmax = 0) AS inserted
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
		"created_at":                       t.CreatedAt,
		"updated_at":                       t.UpdatedAt,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot upsert common third party: %w", err)
	}
	defer rows.Close()

	inserted, err = pgx.CollectExactlyOneRow(rows, pgx.RowTo[bool])
	if err != nil {
		return false, fmt.Errorf("cannot collect upsert result: %w", err)
	}

	return inserted, nil
}

func (t CommonThirdParty) Delete(
	ctx context.Context,
	conn pg.Tx,
	id gid.GID,
) error {
	q := `DELETE FROM common_third_parties WHERE id = @id`

	args := pgx.StrictNamedArgs{"id": id}

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete common third party: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

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
