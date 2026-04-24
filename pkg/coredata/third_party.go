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
	"go.probo.inc/probo/pkg/page"
)

type (
	ThirdParty struct {
		ID                            gid.GID            `db:"id"`
		TenantID                      gid.TenantID       `db:"tenant_id"`
		OrganizationID                gid.GID            `db:"organization_id"`
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
		SnapshotID                    *gid.GID           `db:"snapshot_id"`
		SourceID                      *gid.GID           `db:"source_id"`
		CreatedAt                     time.Time          `db:"created_at"`
		UpdatedAt                     time.Time          `db:"updated_at"`
	}

	ThirdParties []*ThirdParty

	ThirdPartySnapshotter interface {
		InsertThirdPartySnapshots(ctx context.Context, conn pg.Tx, scope Scoper, organizationID, snapshotID gid.GID) error
	}
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

func (v *ThirdParty) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	q := `SELECT organization_id FROM third_parties WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, v.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query thirdParty authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
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
    snapshot_id,
    source_id,
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
    snapshot_id,
    source_id,
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
		return fmt.Errorf("cannot query third_parties: %w", err)
	}

	thirdParties, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ThirdParty])
	if err != nil {
		return fmt.Errorf("cannot collect third_parties: %w", err)
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
        snapshot_id,
        source_id,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @third_party_id,
    @organization_id,
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
    @snapshot_id,
    @source_id,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"tenant_id":                        scope.GetTenantID(),
		"third_party_id":                   v.ID,
		"organization_id":                  v.OrganizationID,
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
		"snapshot_id":                      v.SnapshotID,
		"source_id":                        v.SourceID,
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
		return 0, fmt.Errorf("cannot count third_parties: %w", err)
	}

	return count, nil
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
	snapshot_id,
	source_id,
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
		return fmt.Errorf("cannot query third_parties: %w", err)
	}

	thirdParties, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ThirdParty])
	if err != nil {
		return fmt.Errorf("cannot collect third_parties: %w", err)
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
		return 0, fmt.Errorf("cannot count third_parties: %w", err)
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
		v.snapshot_id,
		v.source_id,
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
	snapshot_id,
	source_id,
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
		return fmt.Errorf("cannot query third_parties: %w", err)
	}

	thirdParties, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ThirdParty])
	if err != nil {
		return fmt.Errorf("cannot collect third_parties: %w", err)
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
		return 0, fmt.Errorf("cannot count third_parties: %w", err)
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
		v.snapshot_id,
		v.source_id,
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
	snapshot_id,
	source_id,
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
		return fmt.Errorf("cannot query third_parties: %w", err)
	}

	thirdParties, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ThirdParty])
	if err != nil {
		return fmt.Errorf("cannot collect third_parties: %w", err)
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
		v.snapshot_id,
		v.source_id,
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
	snapshot_id,
	source_id,
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
		return fmt.Errorf("cannot query third_parties: %w", err)
	}

	thirdParties, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ThirdParty])
	if err != nil {
		return fmt.Errorf("cannot collect third_parties: %w", err)
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
		v.snapshot_id,
		v.source_id,
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
	snapshot_id,
	source_id,
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
		return fmt.Errorf("cannot query third_parties: %w", err)
	}

	thirdParties, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ThirdParty])
	if err != nil {
		return fmt.Errorf("cannot collect third_parties: %w", err)
	}

	*v = thirdParties

	return nil
}

func (v *ThirdParties) LoadAllByProcessingActivities(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	filter *ProcessingActivityFilter,
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
		AND %s
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
	q = fmt.Sprintf(q, filter.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query third_parties: %w", err)
	}
	defer rows.Close()

	thirdPartyMap := make(map[gid.GID][]string)
	for rows.Next() {
		var processingActivityID gid.GID
		var thirdPartyName string
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
		v.snapshot_id,
		v.source_id,
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
	snapshot_id,
	source_id,
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
		return fmt.Errorf("cannot query third_parties: %w", err)
	}

	thirdParties, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ThirdParty])
	if err != nil {
		return fmt.Errorf("cannot collect third_parties: %w", err)
	}

	*vs = thirdParties

	return nil
}

func (vs ThirdParties) InsertProcessingActivitySnapshots(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	organizationID gid.GID,
	snapshotID gid.GID,
) error {
	query := `
WITH
	source_processing_activities AS (
		SELECT id
		FROM processing_activities
		WHERE organization_id = @organization_id AND snapshot_id IS NULL
	),
	source_processing_activity_third_parties AS (
		SELECT processing_activity_id, third_party_id, snapshot_id, created_at
		FROM processing_activity_third_parties
		WHERE processing_activity_id = ANY(SELECT id FROM source_processing_activities)
	),
	source_third_parties AS (
		SELECT *
		FROM third_parties
		WHERE %s AND id = ANY(SELECT third_party_id FROM source_processing_activity_third_parties)
	)
INSERT INTO third_parties (
	tenant_id,
	id,
	snapshot_id,
	source_id,
	organization_id,
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
	created_at,
	updated_at
)
SELECT
	@tenant_id,
	generate_gid(decode_base64_unpadded(@tenant_id), @third_party_entity_type),
	@snapshot_id,
	v.id,
	v.organization_id,
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
	v.created_at,
	v.updated_at
FROM source_third_parties v
	`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"tenant_id":               scope.GetTenantID(),
		"snapshot_id":             snapshotID,
		"organization_id":         organizationID,
		"third_party_entity_type": ThirdPartyEntityType,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert thirdParty snapshots for processing activities: %w", err)
	}

	return nil
}

func (v ThirdParties) Snapshot(ctx context.Context, conn pg.Tx, scope Scoper, organizationID, snapshotID gid.GID) error {
	for _, snapshotter := range []ThirdPartySnapshotter{
		ThirdParties{},
		ThirdPartyServices{},
		ThirdPartyContacts{},
		ThirdPartyRiskAssessments{},
		ThirdPartyComplianceReports{},
		ThirdPartyBusinessAssociateAgreements{},
		ThirdPartyDataPrivacyAgreements{},
	} {
		if err := snapshotter.InsertThirdPartySnapshots(ctx, conn, scope, organizationID, snapshotID); err != nil {
			return fmt.Errorf("cannot create thirdParty snapshots: (%T) %w", snapshotter, err)
		}
	}

	return nil
}

func (v ThirdParties) InsertThirdPartySnapshots(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	organizationID gid.GID,
	snapshotID gid.GID,
) error {
	query := `
INSERT INTO third_parties (
	tenant_id,
	id,
	snapshot_id,
	source_id,
	organization_id,
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
	created_at,
	updated_at
)
SELECT
	@tenant_id,
	generate_gid(decode_base64_unpadded(@tenant_id), @third_party_entity_type),
	@snapshot_id,
	v.id,
	v.organization_id,
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
	v.created_at,
	v.updated_at
FROM third_parties v
WHERE %s AND organization_id = @organization_id AND snapshot_id IS NULL
	`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"tenant_id":               scope.GetTenantID(),
		"snapshot_id":             snapshotID,
		"organization_id":         organizationID,
		"third_party_entity_type": ThirdPartyEntityType,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert thirdParty snapshots: %w", err)
	}

	return nil
}
