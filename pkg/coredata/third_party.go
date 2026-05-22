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

func (v ThirdParty) GetGeneratedDocumentID(
	ctx context.Context,
	conn pg.Querier,
	organizationID gid.GID,
) (*gid.GID, error) {
	var documentID *gid.GID

	err := conn.QueryRow(
		ctx,
		`
SELECT
	third_parties_document_id
FROM
	generated_documents
WHERE
	organization_id = @organization_id
`,
		pgx.NamedArgs{"organization_id": organizationID},
	).Scan(&documentID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("cannot get thirdParty list document ID: %w", err)
	}

	return documentID, nil
}

func (v ThirdParty) UpsertGeneratedDocumentID(
	ctx context.Context,
	conn pg.Tx,
	organizationID gid.GID,
	tenantID gid.TenantID,
	documentID gid.GID,
) error {
	now := time.Now()

	_, err := conn.Exec(
		ctx,
		`
INSERT INTO generated_documents (
	organization_id,
	tenant_id,
	third_parties_document_id,
	created_at,
	updated_at
) VALUES (
	@organization_id,
	@tenant_id,
	@third_parties_document_id,
	@created_at,
	@updated_at
)
ON CONFLICT (organization_id) DO UPDATE
SET
	third_parties_document_id = @third_parties_document_id,
	updated_at = @updated_at
`,
		pgx.NamedArgs{
			"organization_id":           organizationID,
			"tenant_id":                 tenantID,
			"third_parties_document_id": documentID,
			"created_at":                now,
			"updated_at":                now,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot upsert thirdParty list document ID: %w", err)
	}

	return nil
}

func (v ThirdParty) ClearGeneratedDocumentID(
	ctx context.Context,
	conn pg.Tx,
	documentIDs []gid.GID,
) error {
	ids := make([]string, len(documentIDs))
	for i, id := range documentIDs {
		ids[i] = id.String()
	}

	_, err := conn.Exec(
		ctx,
		`
UPDATE
	generated_documents
SET
	third_parties_document_id = NULL,
	updated_at = @now
WHERE
	third_parties_document_id = ANY(@ids)
`,
		pgx.NamedArgs{
			"ids": ids,
			"now": time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("cannot clear thirdParty list document references: %w", err)
	}

	return nil
}

type (
	ThirdParty struct {
		ID                            gid.GID            `db:"id"`
		TenantID                      gid.TenantID       `db:"tenant_id"`
		OrganizationID                gid.GID            `db:"organization_id"`
		CommonThirdPartyID            *gid.GID           `db:"common_third_party_id"`
		Name                          string             `db:"name"`
		Description                   *string            `db:"description"`
		Category                      ThirdPartyCategory `db:"category"`
		HeadquarterAddress            *string            `db:"headquarter_address"`
		LegalName                     *string            `db:"legal_name"`
		WebsiteURL                    *string            `db:"website_url"`
		PrivacyPolicyURL              *string            `db:"privacy_policy_url"`
		ServiceLevelAgreementURL      *string            `db:"service_level_agreement_url"`
		DataProcessingAgreementURL    *string            `db:"data_processing_agreement_url"`
		BusinessAssociateAgreementURL *string            `db:"business_associate_agreement_url"`
		SubprocessorsListURL          *string            `db:"subprocessors_list_url"`
		Certifications                []string           `db:"certifications"`
		Countries                     CountryCodes       `db:"countries"`
		BusinessOwnerID               *gid.GID           `db:"business_owner_profile_id"`
		SecurityOwnerID               *gid.GID           `db:"security_owner_profile_id"`
		StatusPageURL                 *string            `db:"status_page_url"`
		TermsOfServiceURL             *string            `db:"terms_of_service_url"`
		SecurityPageURL               *string            `db:"security_page_url"`
		TrustPageURL                  *string            `db:"trust_page_url"`
		ShowOnTrustCenter             bool               `db:"show_on_trust_center"`
		FirstLevel                    bool               `db:"first_level"`
		CreatedAt                     time.Time          `db:"created_at"`
		UpdatedAt                     time.Time          `db:"updated_at"`
	}

	ThirdParties []*ThirdParty
)

func (v ThirdParty) CursorKey(orderBy ThirdPartyOrderField) page.CursorKey {
	switch orderBy {
	case ThirdPartyOrderFieldCreatedAt:
		return page.NewCursorKey(v.ID, v.CreatedAt)
	case ThirdPartyOrderFieldUpdatedAt:
		return page.NewCursorKey(v.ID, v.UpdatedAt)
	case ThirdPartyOrderFieldName:
		return page.NewCursorKey(v.ID, v.Name)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (v *ThirdParty) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM third_parties WHERE id = ANY(@resource_ids::text[])`

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

func (v *ThirdParty) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	thirdPartyID gid.GID,
) error {
	q := `
SELECT
    id,
    tenant_id,
    organization_id,
    common_third_party_id,
    name,
    description,
    category,
    headquarter_address,
    legal_name,
    website_url,
    privacy_policy_url,
    service_level_agreement_url,
    data_processing_agreement_url,
    business_associate_agreement_url,
    subprocessors_list_url,
    certifications,
    countries,
    business_owner_profile_id,
    security_owner_profile_id,
    status_page_url,
    terms_of_service_url,
    security_page_url,
    trust_page_url,
    show_on_trust_center,
    first_level,
    created_at,
    updated_at
FROM
    third_parties
WHERE
    %s
    AND id = @third_party_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"third_party_id": thirdPartyID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query thirdParty: %w", err)
	}
	defer rows.Close()

	thirdParty, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ThirdParty])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect thirdParty: %w", err)
	}

	*v = thirdParty

	return nil
}

func (v *ThirdParties) LoadByIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	thirdPartyIDs []gid.GID,
) error {
	q := `
SELECT
    id,
    tenant_id,
    organization_id,
    common_third_party_id,
    name,
    description,
    category,
    headquarter_address,
    legal_name,
    website_url,
    privacy_policy_url,
    service_level_agreement_url,
    data_processing_agreement_url,
    business_associate_agreement_url,
    subprocessors_list_url,
    certifications,
    countries,
    business_owner_profile_id,
    security_owner_profile_id,
    status_page_url,
    terms_of_service_url,
    security_page_url,
    trust_page_url,
    show_on_trust_center,
    first_level,
    created_at,
    updated_at
FROM
    third_parties
WHERE
    %s
    AND id = ANY(@third_party_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"third_party_ids": thirdPartyIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query thirdParties: %w", err)
	}

	thirdParties, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ThirdParty])
	if err != nil {
		return fmt.Errorf("cannot collect thirdParties: %w", err)
	}

	*v = thirdParties

	return nil
}

func (v ThirdParty) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    third_parties (
        tenant_id,
        id,
        organization_id,
        common_third_party_id,
        name,
        description,
        category,
        headquarter_address,
        legal_name,
        website_url,
        privacy_policy_url,
        service_level_agreement_url,
        data_processing_agreement_url,
        business_associate_agreement_url,
        subprocessors_list_url,
        certifications,
        countries,
        business_owner_profile_id,
        security_owner_profile_id,
        status_page_url,
        terms_of_service_url,
        security_page_url,
        trust_page_url,
        show_on_trust_center,
        first_level,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @third_party_id,
    @organization_id,
    @common_third_party_id,
    @name,
    @description,
    @category,
    @headquarter_address,
    @legal_name,
    @website_url,
    @privacy_policy_url,
    @service_level_agreement_url,
    @data_processing_agreement_url,
    @business_associate_agreement_url,
    @subprocessors_list_url,
    @certifications,
    @countries,
    @business_owner_profile_id,
    @security_owner_profile_id,
    @status_page_url,
    @terms_of_service_url,
    @security_page_url,
    @trust_page_url,
    @show_on_trust_center,
    @first_level,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"tenant_id":                        scope.GetTenantID(),
		"third_party_id":                   v.ID,
		"organization_id":                  v.OrganizationID,
		"common_third_party_id":            v.CommonThirdPartyID,
		"name":                             v.Name,
		"description":                      v.Description,
		"category":                         v.Category,
		"headquarter_address":              v.HeadquarterAddress,
		"legal_name":                       v.LegalName,
		"website_url":                      v.WebsiteURL,
		"privacy_policy_url":               v.PrivacyPolicyURL,
		"service_level_agreement_url":      v.ServiceLevelAgreementURL,
		"data_processing_agreement_url":    v.DataProcessingAgreementURL,
		"business_associate_agreement_url": v.BusinessAssociateAgreementURL,
		"subprocessors_list_url":           v.SubprocessorsListURL,
		"certifications":                   v.Certifications,
		"countries":                        v.Countries,
		"business_owner_profile_id":        v.BusinessOwnerID,
		"security_owner_profile_id":        v.SecurityOwnerID,
		"status_page_url":                  v.StatusPageURL,
		"terms_of_service_url":             v.TermsOfServiceURL,
		"security_page_url":                v.SecurityPageURL,
		"trust_page_url":                   v.TrustPageURL,
		"show_on_trust_center":             v.ShowOnTrustCenter,
		"first_level":                      v.FirstLevel,
		"created_at":                       v.CreatedAt,
		"updated_at":                       v.UpdatedAt,
	}
	_, err := conn.Exec(ctx, q, args)

	return err
}

func (v ThirdParty) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM third_parties WHERE %s AND id = @third_party_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"third_party_id": v.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)

	return err
}

func (v *ThirdParties) CountByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	filter *ThirdPartyFilter,
) (int, error) {
	q := `
SELECT
    COUNT(id)
FROM
    third_parties
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

	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count thirdParties: %w", err)
	}

	return count, nil
}

func (v *ThirdParties) LoadAllByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) error {
	q := `
SELECT
	id,
	tenant_id,
	organization_id,
	common_third_party_id,
	name,
	description,
	category,
	headquarter_address,
	legal_name,
	website_url,
	privacy_policy_url,
	service_level_agreement_url,
	data_processing_agreement_url,
	business_associate_agreement_url,
	subprocessors_list_url,
	certifications,
	countries,
	business_owner_profile_id,
	security_owner_profile_id,
	status_page_url,
	terms_of_service_url,
	security_page_url,
	trust_page_url,
	show_on_trust_center,
	first_level,
	created_at,
	updated_at
FROM
	third_parties
WHERE
	%s
	AND organization_id = @organization_id
ORDER BY name ASC
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query thirdParties: %w", err)
	}

	thirdParties, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ThirdParty])
	if err != nil {
		return fmt.Errorf("cannot collect thirdParties: %w", err)
	}

	*v = thirdParties

	return nil
}

func (v *ThirdParties) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[ThirdPartyOrderField],
	filter *ThirdPartyFilter,
) error {
	q := `
SELECT
	id,
	tenant_id,
	organization_id,
	common_third_party_id,
	name,
	description,
	category,
	headquarter_address,
	legal_name,
	website_url,
	privacy_policy_url,
	service_level_agreement_url,
	data_processing_agreement_url,
	business_associate_agreement_url,
	subprocessors_list_url,
	certifications,
	countries,
	business_owner_profile_id,
	security_owner_profile_id,
	status_page_url,
	terms_of_service_url,
	security_page_url,
	trust_page_url,
	show_on_trust_center,
	first_level,
	created_at,
	updated_at
FROM
	third_parties
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
		return fmt.Errorf("cannot query thirdParties: %w", err)
	}

	thirdParties, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ThirdParty])
	if err != nil {
		return fmt.Errorf("cannot collect thirdParties: %w", err)
	}

	*v = thirdParties

	return nil
}

func (v *ThirdParty) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE third_parties
SET
	name = @name,
	description = @description,
	category = @category,
	headquarter_address = @headquarter_address,
	legal_name = @legal_name,
	website_url = @website_url,
	privacy_policy_url = @privacy_policy_url,
	service_level_agreement_url = @service_level_agreement_url,
	data_processing_agreement_url = @data_processing_agreement_url,
	business_associate_agreement_url = @business_associate_agreement_url,
	subprocessors_list_url = @subprocessors_list_url,
	certifications = @certifications,
	countries = @countries,
	status_page_url = @status_page_url,
	terms_of_service_url = @terms_of_service_url,
	security_page_url = @security_page_url,
	trust_page_url = @trust_page_url,
	business_owner_profile_id = @business_owner_profile_id,
	security_owner_profile_id = @security_owner_profile_id,
	show_on_trust_center = @show_on_trust_center,
	first_level = @first_level,
	updated_at = @updated_at
WHERE %s
    AND id = @third_party_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"third_party_id":                   v.ID,
		"updated_at":                       time.Now(),
		"name":                             v.Name,
		"description":                      v.Description,
		"category":                         v.Category,
		"headquarter_address":              v.HeadquarterAddress,
		"legal_name":                       v.LegalName,
		"website_url":                      v.WebsiteURL,
		"privacy_policy_url":               v.PrivacyPolicyURL,
		"service_level_agreement_url":      v.ServiceLevelAgreementURL,
		"data_processing_agreement_url":    v.DataProcessingAgreementURL,
		"business_associate_agreement_url": v.BusinessAssociateAgreementURL,
		"subprocessors_list_url":           v.SubprocessorsListURL,
		"certifications":                   v.Certifications,
		"countries":                        v.Countries,
		"status_page_url":                  v.StatusPageURL,
		"terms_of_service_url":             v.TermsOfServiceURL,
		"security_page_url":                v.SecurityPageURL,
		"trust_page_url":                   v.TrustPageURL,
		"business_owner_profile_id":        v.BusinessOwnerID,
		"security_owner_profile_id":        v.SecurityOwnerID,
		"show_on_trust_center":             v.ShowOnTrustCenter,
		"first_level":                      v.FirstLevel,
	}

	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)

	return err
}

func (v ThirdParty) ExpireNonExpiredRiskAssessments(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	now := time.Now()

	q := `
	UPDATE third_party_risk_assessments
	SET
		expires_at = @now,
		updated_at = @now
	WHERE
		%s
		AND third_party_id = @third_party_id
		AND expires_at > @now
	`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"third_party_id": v.ID,
		"now":            now,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot expire existing risk assessments: %w", err)
	}

	return nil
}

func (v *ThirdParties) CountByAssetID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	assetID gid.GID,
) (int, error) {
	q := `
WITH vend AS (
	SELECT
		v.id
	FROM
		third_parties v
	INNER JOIN
		asset_third_parties av ON v.id = av.third_party_id
	WHERE
		av.asset_id = @asset_id
)
SELECT
	COUNT(id)
FROM
	vend
WHERE %s
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"asset_id": assetID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int

	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count thirdParties: %w", err)
	}

	return count, nil
}

func (v *ThirdParties) LoadByAssetID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	assetID gid.GID,
	cursor *page.Cursor[ThirdPartyOrderField],
) error {
	q := `
WITH vend AS (
	SELECT
		v.id,
		v.tenant_id,
		v.organization_id,
		v.common_third_party_id,
		v.name,
		v.description,
		v.category,
		v.headquarter_address,
		v.legal_name,
		v.website_url,
		v.privacy_policy_url,
		v.service_level_agreement_url,
		v.data_processing_agreement_url,
		v.business_associate_agreement_url,
		v.subprocessors_list_url,
		v.certifications,
		v.countries,
		v.business_owner_profile_id,
		v.security_owner_profile_id,
		v.status_page_url,
		v.terms_of_service_url,
		v.security_page_url,
		v.trust_page_url,
		v.show_on_trust_center,
		v.first_level,
		v.created_at,
		v.updated_at
	FROM
		third_parties v
	INNER JOIN
		asset_third_parties av ON v.id = av.third_party_id
	WHERE
		av.asset_id = @asset_id
)
SELECT
	id,
	tenant_id,
	organization_id,
	common_third_party_id,
	name,
	description,
	category,
	headquarter_address,
	legal_name,
	website_url,
	privacy_policy_url,
	service_level_agreement_url,
	data_processing_agreement_url,
	business_associate_agreement_url,
	subprocessors_list_url,
	certifications,
	countries,
	business_owner_profile_id,
	security_owner_profile_id,
	status_page_url,
	terms_of_service_url,
	security_page_url,
	trust_page_url,
	show_on_trust_center,
	first_level,
	created_at,
	updated_at
FROM
	vend
WHERE %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"asset_id": assetID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query thirdParties: %w", err)
	}

	thirdParties, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ThirdParty])
	if err != nil {
		return fmt.Errorf("cannot collect thirdParties: %w", err)
	}

	*v = thirdParties

	return nil
}

func (v *ThirdParties) CountByDatumID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	datumID gid.GID,
) (int, error) {
	q := `
WITH vend AS (
	SELECT
		v.id
	FROM
		third_parties v
	INNER JOIN
		data_third_parties dv ON v.id = dv.third_party_id
	WHERE
		dv.datum_id = @datum_id
)
SELECT
	COUNT(id)
FROM
	vend
WHERE %s
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"datum_id": datumID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int

	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count thirdParties: %w", err)
	}

	return count, nil
}

func (vs *ThirdParties) LoadAllByDatumID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	datumID gid.GID,
) error {
	q := `
WITH vend AS (
	SELECT
		v.id,
		v.tenant_id,
		v.organization_id,
		v.common_third_party_id,
		v.name,
		v.description,
		v.category,
		v.headquarter_address,
		v.legal_name,
		v.website_url,
		v.privacy_policy_url,
		v.service_level_agreement_url,
		v.data_processing_agreement_url,
		v.business_associate_agreement_url,
		v.subprocessors_list_url,
		v.certifications,
		v.countries,
		v.business_owner_profile_id,
		v.security_owner_profile_id,
		v.status_page_url,
		v.terms_of_service_url,
		v.security_page_url,
		v.trust_page_url,
		v.show_on_trust_center,
		v.first_level,
		v.created_at,
		v.updated_at
	FROM
		third_parties v
	INNER JOIN
		data_third_parties dv ON v.id = dv.third_party_id
	WHERE
		dv.datum_id = @datum_id
)
SELECT
	id,
	tenant_id,
	organization_id,
	common_third_party_id,
	name,
	description,
	category,
	headquarter_address,
	legal_name,
	website_url,
	privacy_policy_url,
	service_level_agreement_url,
	data_processing_agreement_url,
	business_associate_agreement_url,
	subprocessors_list_url,
	certifications,
	countries,
	business_owner_profile_id,
	security_owner_profile_id,
	status_page_url,
	terms_of_service_url,
	security_page_url,
	trust_page_url,
	show_on_trust_center,
	first_level,
	created_at,
	updated_at
FROM
	vend
WHERE %s
ORDER BY name ASC
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"datum_id": datumID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query thirdParties: %w", err)
	}

	thirdParties, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ThirdParty])
	if err != nil {
		return fmt.Errorf("cannot collect thirdParties: %w", err)
	}

	*vs = thirdParties

	return nil
}

func (vs *ThirdParties) LoadByDatumID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	datumID gid.GID,
	cursor *page.Cursor[ThirdPartyOrderField],
) error {
	q := `
WITH vend AS (
	SELECT
		v.id,
		v.tenant_id,
		v.organization_id,
		v.common_third_party_id,
		v.name,
		v.description,
		v.category,
		v.headquarter_address,
		v.legal_name,
		v.website_url,
		v.privacy_policy_url,
		v.service_level_agreement_url,
		v.data_processing_agreement_url,
		v.business_associate_agreement_url,
		v.subprocessors_list_url,
		v.certifications,
		v.countries,
		v.business_owner_profile_id,
		v.security_owner_profile_id,
		v.status_page_url,
		v.terms_of_service_url,
		v.security_page_url,
		v.trust_page_url,
		v.show_on_trust_center,
		v.first_level,
		v.created_at,
		v.updated_at
	FROM
		third_parties v
	INNER JOIN
		data_third_parties dv ON v.id = dv.third_party_id
	WHERE
		dv.datum_id = @datum_id
)
SELECT
	id,
	tenant_id,
	organization_id,
	common_third_party_id,
	name,
	description,
	category,
	headquarter_address,
	legal_name,
	website_url,
	privacy_policy_url,
	service_level_agreement_url,
	data_processing_agreement_url,
	business_associate_agreement_url,
	subprocessors_list_url,
	certifications,
	countries,
	business_owner_profile_id,
	security_owner_profile_id,
	status_page_url,
	terms_of_service_url,
	security_page_url,
	trust_page_url,
	show_on_trust_center,
	first_level,
	created_at,
	updated_at
FROM
	vend
WHERE %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"datum_id": datumID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query thirdParties: %w", err)
	}

	thirdParties, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ThirdParty])
	if err != nil {
		return fmt.Errorf("cannot collect thirdParties: %w", err)
	}

	*vs = thirdParties

	return nil
}

func (v *ThirdParties) LoadByProcessingActivityID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	processingActivityID gid.GID,
	cursor *page.Cursor[ThirdPartyOrderField],
) error {
	q := `
WITH vend AS (
	SELECT
		v.id,
		v.tenant_id,
		v.organization_id,
		v.common_third_party_id,
		v.name,
		v.description,
		v.category,
		v.headquarter_address,
		v.legal_name,
		v.website_url,
		v.privacy_policy_url,
		v.service_level_agreement_url,
		v.data_processing_agreement_url,
		v.business_associate_agreement_url,
		v.subprocessors_list_url,
		v.certifications,
		v.countries,
		v.business_owner_profile_id,
		v.security_owner_profile_id,
		v.status_page_url,
		v.terms_of_service_url,
		v.security_page_url,
		v.trust_page_url,
		v.show_on_trust_center,
		v.first_level,
		v.created_at,
		v.updated_at
	FROM
		third_parties v
	INNER JOIN
		processing_activity_third_parties pav ON v.id = pav.third_party_id
	WHERE
		pav.processing_activity_id = @processing_activity_id
)
SELECT
	id,
	tenant_id,
	organization_id,
	common_third_party_id,
	name,
	description,
	category,
	headquarter_address,
	legal_name,
	website_url,
	privacy_policy_url,
	service_level_agreement_url,
	data_processing_agreement_url,
	business_associate_agreement_url,
	subprocessors_list_url,
	certifications,
	countries,
	business_owner_profile_id,
	security_owner_profile_id,
	status_page_url,
	terms_of_service_url,
	security_page_url,
	trust_page_url,
	show_on_trust_center,
	first_level,
	created_at,
	updated_at
FROM
	vend
WHERE %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"processing_activity_id": processingActivityID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query thirdParties: %w", err)
	}

	thirdParties, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ThirdParty])
	if err != nil {
		return fmt.Errorf("cannot collect thirdParties: %w", err)
	}

	*v = thirdParties

	return nil
}

func (v *ThirdParties) LoadAllByProcessingActivities(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) (map[gid.GID][]string, error) {
	q := `
WITH filtered_processing_activities AS (
	SELECT
		pa.id
	FROM
		processing_activities pa
	WHERE
		pa.tenant_id = @tenant_id
		AND pa.organization_id = @organization_id
),
filtered_third_parties AS (
	SELECT
		v.id,
		v.name
	FROM
		third_parties v
	WHERE
		v.tenant_id = @tenant_id
)
SELECT
	pav.processing_activity_id,
	fv.name
FROM
	processing_activity_third_parties pav
INNER JOIN
	filtered_third_parties fv ON fv.id = pav.third_party_id
INNER JOIN
	filtered_processing_activities fpa ON fpa.id = pav.processing_activity_id
WHERE
	pav.tenant_id = @tenant_id
ORDER BY
	pav.processing_activity_id, fv.name
	`

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query thirdParties: %w", err)
	}
	defer rows.Close()

	thirdPartyMap := make(map[gid.GID][]string)

	for rows.Next() {
		var (
			processingActivityID gid.GID
			thirdPartyName       string
		)

		if err := rows.Scan(&processingActivityID, &thirdPartyName); err != nil {
			return nil, fmt.Errorf("cannot scan thirdParty: %w", err)
		}

		thirdPartyMap[processingActivityID] = append(thirdPartyMap[processingActivityID], thirdPartyName)
	}

	return thirdPartyMap, nil
}

func (vs *ThirdParties) LoadAllByAssetID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	assetID gid.GID,
) error {
	q := `
WITH vend AS (
	SELECT
		v.id,
		v.tenant_id,
		v.organization_id,
		v.common_third_party_id,
		v.name,
		v.description,
		v.category,
		v.headquarter_address,
		v.legal_name,
		v.website_url,
		v.privacy_policy_url,
		v.service_level_agreement_url,
		v.data_processing_agreement_url,
		v.business_associate_agreement_url,
		v.subprocessors_list_url,
		v.certifications,
		v.countries,
		v.business_owner_profile_id,
		v.security_owner_profile_id,
		v.status_page_url,
		v.terms_of_service_url,
		v.security_page_url,
		v.trust_page_url,
		v.show_on_trust_center,
		v.first_level,
		v.created_at,
		v.updated_at
	FROM
		third_parties v
	INNER JOIN
		asset_third_parties av ON v.id = av.third_party_id
	WHERE
		av.asset_id = @asset_id
)
SELECT
	id,
	tenant_id,
	organization_id,
	common_third_party_id,
	name,
	description,
	category,
	headquarter_address,
	legal_name,
	website_url,
	privacy_policy_url,
	service_level_agreement_url,
	data_processing_agreement_url,
	business_associate_agreement_url,
	subprocessors_list_url,
	certifications,
	countries,
	business_owner_profile_id,
	security_owner_profile_id,
	status_page_url,
	terms_of_service_url,
	security_page_url,
	trust_page_url,
	show_on_trust_center,
	first_level,
	created_at,
	updated_at
FROM
	vend
WHERE %s
ORDER BY name ASC
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"asset_id": assetID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query thirdParties: %w", err)
	}

	thirdParties, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ThirdParty])
	if err != nil {
		return fmt.Errorf("cannot collect thirdParties: %w", err)
	}

	*vs = thirdParties

	return nil
}

func (v *ThirdParty) LoadByOrganizationIDAndCommonThirdPartyID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	commonThirdPartyID gid.GID,
) error {
	q := `
SELECT
	id,
	tenant_id,
	organization_id,
	common_third_party_id,
	name,
	description,
	category,
	headquarter_address,
	legal_name,
	website_url,
	privacy_policy_url,
	service_level_agreement_url,
	data_processing_agreement_url,
	business_associate_agreement_url,
	subprocessors_list_url,
	certifications,
	countries,
	business_owner_profile_id,
	security_owner_profile_id,
	status_page_url,
	terms_of_service_url,
	security_page_url,
	trust_page_url,
	show_on_trust_center,
	first_level,
	created_at,
	updated_at
FROM
	third_parties
WHERE
	%s
	AND organization_id = @organization_id
	AND common_third_party_id = @common_third_party_id
	AND snapshot_id IS NULL
ORDER BY id ASC
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id":       organizationID,
		"common_third_party_id": commonThirdPartyID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query vendor by common third party: %w", err)
	}
	defer rows.Close()

	vendor, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ThirdParty])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect vendor by common third party: %w", err)
	}

	*v = vendor

	return nil
}

func (v *ThirdParties) CountByMeasureID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	measureID gid.GID,
) (int, error) {
	q := `
WITH tps AS (
	SELECT
		v.id,
		v.tenant_id
	FROM
		third_parties v
	INNER JOIN
		measures_third_parties mtp ON v.id = mtp.third_party_id
	WHERE
		mtp.measure_id = @measure_id
)
SELECT
	COUNT(id)
FROM
	tps
WHERE %s
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"measure_id": measureID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int

	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count thirdParties: %w", err)
	}

	return count, nil
}

func (v *ThirdParties) LoadByMeasureID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	measureID gid.GID,
	cursor *page.Cursor[ThirdPartyOrderField],
) error {
	q := `
WITH tps AS (
	SELECT
		v.id,
		v.tenant_id,
		v.organization_id,
		v.common_third_party_id,
		v.name,
		v.description,
		v.category,
		v.headquarter_address,
		v.legal_name,
		v.website_url,
		v.privacy_policy_url,
		v.service_level_agreement_url,
		v.data_processing_agreement_url,
		v.business_associate_agreement_url,
		v.subprocessors_list_url,
		v.certifications,
		v.countries,
		v.business_owner_profile_id,
		v.security_owner_profile_id,
		v.status_page_url,
		v.terms_of_service_url,
		v.security_page_url,
		v.trust_page_url,
		v.show_on_trust_center,
		v.first_level,
		v.created_at,
		v.updated_at
	FROM
		third_parties v
	INNER JOIN
		measures_third_parties mtp ON v.id = mtp.third_party_id
	WHERE
		mtp.measure_id = @measure_id
)
SELECT
	id,
	tenant_id,
	organization_id,
	common_third_party_id,
	name,
	description,
	category,
	headquarter_address,
	legal_name,
	website_url,
	privacy_policy_url,
	service_level_agreement_url,
	data_processing_agreement_url,
	business_associate_agreement_url,
	subprocessors_list_url,
	certifications,
	countries,
	business_owner_profile_id,
	security_owner_profile_id,
	status_page_url,
	terms_of_service_url,
	security_page_url,
	trust_page_url,
	show_on_trust_center,
	first_level,
	created_at,
	updated_at
FROM
	tps
WHERE %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"measure_id": measureID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query thirdParties: %w", err)
	}

	thirdParties, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ThirdParty])
	if err != nil {
		return fmt.Errorf("cannot collect thirdParties: %w", err)
	}

	*v = thirdParties

	return nil
}
