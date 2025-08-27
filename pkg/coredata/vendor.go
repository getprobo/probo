// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
	"fmt"
	"maps"
	"time"

	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	Vendor struct {
		ID                            gid.GID        `db:"id"`
		TenantID                      gid.TenantID   `db:"tenant_id"`
		OrganizationID                gid.GID        `db:"organization_id"`
		Name                          string         `db:"name"`
		Description                   *string        `db:"description"`
		Category                      VendorCategory `db:"category"`
		HeadquarterAddress            *string        `db:"headquarter_address"`
		LegalName                     *string        `db:"legal_name"`
		WebsiteURL                    *string        `db:"website_url"`
		PrivacyPolicyURL              *string        `db:"privacy_policy_url"`
		ServiceLevelAgreementURL      *string        `db:"service_level_agreement_url"`
		DataProcessingAgreementURL    *string        `db:"data_processing_agreement_url"`
		BusinessAssociateAgreementURL *string        `db:"business_associate_agreement_url"`
		SubprocessorsListURL          *string        `db:"subprocessors_list_url"`
		Certifications                []string       `db:"certifications"`
		BusinessOwnerID               *gid.GID       `db:"business_owner_id"`
		SecurityOwnerID               *gid.GID       `db:"security_owner_id"`
		StatusPageURL                 *string        `db:"status_page_url"`
		TermsOfServiceURL             *string        `db:"terms_of_service_url"`
		SecurityPageURL               *string        `db:"security_page_url"`
		TrustPageURL                  *string        `db:"trust_page_url"`
		ShowOnTrustCenter             bool           `db:"show_on_trust_center"`
		SnapshotID                    *gid.GID       `db:"snapshot_id"`
		SourceID                      *gid.GID       `db:"source_id"`
		CreatedAt                     time.Time      `db:"created_at"`
		UpdatedAt                     time.Time      `db:"updated_at"`
	}

	Vendors []*Vendor
)

func (v Vendor) CursorKey(orderBy VendorOrderField) page.CursorKey {
	switch orderBy {
	case VendorOrderFieldCreatedAt:
		return page.NewCursorKey(v.ID, v.CreatedAt)
	case VendorOrderFieldUpdatedAt:
		return page.NewCursorKey(v.ID, v.UpdatedAt)
	case VendorOrderFieldName:
		return page.NewCursorKey(v.ID, v.Name)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (v *Vendor) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	vendorID gid.GID,
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
    business_owner_id,
    security_owner_id,
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
    vendors
WHERE
    %s
    AND id = @vendor_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"vendor_id": vendorID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query vendor: %w", err)
	}
	defer rows.Close()

	vendor, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Vendor])
	if err != nil {
		return fmt.Errorf("cannot collect vendor: %w", err)
	}

	*v = vendor

	return nil
}

func (v Vendor) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    vendors (
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
        business_owner_id,
        security_owner_id,
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
    @vendor_id,
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
    @business_owner_id,
    @security_owner_id,
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
		"vendor_id":                        v.ID,
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
		"business_owner_id":                v.BusinessOwnerID,
		"security_owner_id":                v.SecurityOwnerID,
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

func (v Vendor) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM vendors WHERE %s AND id = @vendor_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"vendor_id": v.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (v *Vendors) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
    COUNT(id)
FROM
    vendors
WHERE
    %s
    AND organization_id = @organization_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count vendors: %w", err)
	}

	return count, nil
}

func (v *Vendors) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[VendorOrderField],
	filter *VendorFilter,
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
	business_owner_id,
	security_owner_id,
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
	vendors
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
		return fmt.Errorf("cannot query vendors: %w", err)
	}

	vendors, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Vendor])
	if err != nil {
		return fmt.Errorf("cannot collect vendors: %w", err)
	}

	*v = vendors

	return nil
}

func (v *Vendor) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE vendors
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
	status_page_url = @status_page_url,
	terms_of_service_url = @terms_of_service_url,
	security_page_url = @security_page_url,
	trust_page_url = @trust_page_url,
	business_owner_id = @business_owner_id,
	security_owner_id = @security_owner_id,
	show_on_trust_center = @show_on_trust_center,
	updated_at = @updated_at
WHERE %s
    AND id = @vendor_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"vendor_id":                        v.ID,
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
		"status_page_url":                  v.StatusPageURL,
		"terms_of_service_url":             v.TermsOfServiceURL,
		"security_page_url":                v.SecurityPageURL,
		"trust_page_url":                   v.TrustPageURL,
		"business_owner_id":                v.BusinessOwnerID,
		"security_owner_id":                v.SecurityOwnerID,
		"show_on_trust_center":             v.ShowOnTrustCenter,
	}

	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (v Vendor) ExpireNonExpiredRiskAssessments(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	now := time.Now()

	q := `
	UPDATE vendor_risk_assessments
	SET
		expires_at = @now,
		updated_at = @now
	WHERE
		%s
		AND vendor_id = @vendor_id
		AND expires_at > @now
	`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"vendor_id": v.ID,
		"now":       now,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot expire existing risk assessments: %w", err)
	}

	return nil
}

func (v *Vendors) CountByAssetID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	assetID gid.GID,
) (int, error) {
	q := `
WITH vend AS (
	SELECT
		v.id
	FROM
		vendors v
	INNER JOIN
		asset_vendors av ON v.id = av.vendor_id
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
		return 0, fmt.Errorf("cannot count vendors: %w", err)
	}

	return count, nil
}

func (v *Vendors) LoadByAssetID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	assetID gid.GID,
	cursor *page.Cursor[VendorOrderField],
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
		v.business_owner_id,
		v.security_owner_id,
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
		vendors v
	INNER JOIN
		asset_vendors av ON v.id = av.vendor_id
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
	business_owner_id,
	security_owner_id,
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
		return fmt.Errorf("cannot query vendors: %w", err)
	}

	vendors, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Vendor])
	if err != nil {
		return fmt.Errorf("cannot collect vendors: %w", err)
	}

	*v = vendors

	return nil
}

func (v *Vendors) CountByDatumID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	datumID gid.GID,
) (int, error) {
	q := `
WITH vend AS (
	SELECT
		v.id
	FROM
		vendors v
	INNER JOIN
		data_vendors dv ON v.id = dv.vendor_id
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
		return 0, fmt.Errorf("cannot count vendors: %w", err)
	}

	return count, nil
}

func (vs *Vendors) LoadByDatumID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	datumID gid.GID,
	cursor *page.Cursor[VendorOrderField],
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
		v.business_owner_id,
		v.security_owner_id,
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
		vendors v
	INNER JOIN
		data_vendors dv ON v.id = dv.vendor_id
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
	business_owner_id,
	security_owner_id,
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
		return fmt.Errorf("cannot query vendors: %w", err)
	}

	vendors, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Vendor])
	if err != nil {
		return fmt.Errorf("cannot collect vendors: %w", err)
	}

	*vs = vendors

	return nil
}

func (d Vendors) InsertDataSnapshots(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	snapshotID gid.GID,
) error {
	query := `
WITH
	source_data AS (
		SELECT id
		FROM data
		WHERE organization_id = @organization_id AND snapshot_id IS NULL
	),
	source_data_vendors AS (
		SELECT datum_id, vendor_id, snapshot_id, created_at
		FROM data_vendors
		WHERE datum_id = ANY(SELECT id FROM source_data)
	),
	source_vendors AS (
		SELECT *
		FROM vendors
		WHERE %s AND id = ANY(SELECT vendor_id FROM source_data_vendors)
	)
INSERT INTO vendors (
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
	business_owner_id,
	security_owner_id,
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
	generate_gid(decode_base64_unpadded(@tenant_id), @vendor_entity_type),
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
	v.business_owner_id,
	v.security_owner_id,
	v.status_page_url,
	v.terms_of_service_url,
	v.security_page_url,
	v.trust_page_url,
	v.show_on_trust_center,
	v.created_at,
	v.updated_at
FROM source_vendors v
	`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"tenant_id":          scope.GetTenantID(),
		"snapshot_id":        snapshotID,
		"organization_id":    organizationID,
		"vendor_entity_type": VendorEntityType,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert vendor snapshots: %w", err)
	}

	return nil
}
