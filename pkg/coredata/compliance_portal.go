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
	CompliancePortal struct {
		ID                           gid.GID              `db:"id"`
		OrganizationID               gid.GID              `db:"organization_id"`
		TenantID                     gid.TenantID         `db:"tenant_id"`
		Active                       bool                 `db:"active"`
		Slug                         string               `db:"slug"`
		SearchEngineIndexing         SearchEngineIndexing `db:"search_engine_indexing"`
		MailingListID                *gid.GID             `db:"mailing_list_id"`
		LogoFileID                   *gid.GID             `db:"logo_file_id"`
		DarkLogoFileID               *gid.GID             `db:"dark_logo_file_id"`
		NonDisclosureAgreementFileID *gid.GID             `db:"non_disclosure_agreement_file_id"`
		DefaultDomainID              *gid.GID             `db:"default_domain_id"`
		CustomDomainID               *gid.GID             `db:"custom_domain_id"`
		Title                        string               `db:"title"`
		Description                  *string              `db:"description"`
		WebsiteURL                   *string              `db:"website_url"`
		Email                        *string              `db:"email"`
		HeadquarterAddress           *string              `db:"headquarter_address"`
		CreatedAt                    time.Time            `db:"created_at"`
		UpdatedAt                    time.Time            `db:"updated_at"`
	}

	CompliancePortals []*CompliancePortal
)

func (tc *CompliancePortal) CursorKey(orderBy CompliancePortalOrderField) page.CursorKey {
	switch orderBy {
	case CompliancePortalOrderFieldCreatedAt:
		return page.NewCursorKey(tc.ID, tc.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (tc *CompliancePortal) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM trust_centers WHERE id = ANY(@resource_ids::text[])`

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

func (tc *CompliancePortal) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	tenant_id,
	mailing_list_id,
	logo_file_id,
	dark_logo_file_id,
	active,
	slug,
	search_engine_indexing,
	non_disclosure_agreement_file_id,
	default_domain_id,
	custom_domain_id,
	title,
	description,
	website_url,
	email,
	headquarter_address,
	created_at,
	updated_at
FROM
	trust_centers
WHERE
	%s
	AND id = @trust_center_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"trust_center_id": compliancePortalID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query compliance portal: %w", err)
	}

	compliancePortal, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CompliancePortal])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect compliance portal: %w", err)
	}

	*tc = compliancePortal

	return nil
}

func (tc *CompliancePortal) LoadByMailingListID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	mailingListID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	tenant_id,
	mailing_list_id,
	logo_file_id,
	dark_logo_file_id,
	active,
	slug,
	search_engine_indexing,
	non_disclosure_agreement_file_id,
	default_domain_id,
	custom_domain_id,
	title,
	description,
	website_url,
	email,
	headquarter_address,
	created_at,
	updated_at
FROM
	trust_centers
WHERE
	%s
	AND mailing_list_id = @mailing_list_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"mailing_list_id": mailingListID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query compliance portal by mailing list id: %w", err)
	}

	compliancePortal, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CompliancePortal])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect compliance portal: %w", err)
	}

	*tc = compliancePortal

	return nil
}

func (tc *CompliancePortal) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	tenant_id,
	mailing_list_id,
	logo_file_id,
	dark_logo_file_id,
	active,
	slug,
	search_engine_indexing,
	non_disclosure_agreement_file_id,
	default_domain_id,
	custom_domain_id,
	title,
	description,
	website_url,
	email,
	headquarter_address,
	created_at,
	updated_at
FROM
	trust_centers
WHERE
	%s
	AND organization_id = @organization_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query compliance portal: %w", err)
	}

	compliancePortal, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CompliancePortal])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect compliance portal: %w", err)
	}

	*tc = compliancePortal

	return nil
}

// Tenant id scope is not applied because we want to access compliance portals by slug across all tenants for public access.
func (tc *CompliancePortal) LoadBySlug(
	ctx context.Context,
	conn pg.Querier,
	slug string,
) error {
	q := `
SELECT
	id,
	organization_id,
	tenant_id,
	mailing_list_id,
	logo_file_id,
	dark_logo_file_id,
	active,
	slug,
	search_engine_indexing,
	non_disclosure_agreement_file_id,
	default_domain_id,
	custom_domain_id,
	title,
	description,
	website_url,
	email,
	headquarter_address,
	created_at,
	updated_at
FROM
	trust_centers
WHERE
	slug = @slug
LIMIT 1;
`

	args := pgx.StrictNamedArgs{"slug": slug}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query compliance portal: %w", err)
	}

	compliancePortal, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CompliancePortal])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect compliance portal: %w", err)
	}

	*tc = compliancePortal

	return nil
}

// LoadByDomainID loads the compliance page that references the given custom
// domain in either of its two slots (default or custom). It powers the
// reverse SNI lookup: a served host resolves to a custom domain, which resolves
// back to its page. Tenant scope is not applied because SNI resolution happens
// across all tenants for public access.
func (tc *CompliancePortal) LoadByDomainID(
	ctx context.Context,
	conn pg.Querier,
	domainID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	tenant_id,
	mailing_list_id,
	logo_file_id,
	dark_logo_file_id,
	active,
	slug,
	search_engine_indexing,
	non_disclosure_agreement_file_id,
	default_domain_id,
	custom_domain_id,
	title,
	description,
	website_url,
	email,
	headquarter_address,
	created_at,
	updated_at
FROM
	trust_centers
WHERE
	default_domain_id = @domain_id
	OR custom_domain_id = @domain_id
LIMIT 1;
`

	args := pgx.StrictNamedArgs{"domain_id": domainID}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query compliance portal by domain id: %w", err)
	}

	compliancePortal, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CompliancePortal])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect compliance portal: %w", err)
	}

	*tc = compliancePortal

	return nil
}

func (tc *CompliancePortal) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO trust_centers (
	id,
	organization_id,
	tenant_id,
	mailing_list_id,
	logo_file_id,
	dark_logo_file_id,
	active,
	slug,
	search_engine_indexing,
	non_disclosure_agreement_file_id,
	default_domain_id,
	custom_domain_id,
	title,
	description,
	website_url,
	email,
	headquarter_address,
	created_at,
	updated_at
) VALUES (
	@id,
	@organization_id,
	@tenant_id,
	@mailing_list_id,
	@logo_file_id,
	@dark_logo_file_id,
	@active,
	@slug,
	@search_engine_indexing,
	@non_disclosure_agreement_file_id,
	@default_domain_id,
	@custom_domain_id,
	@title,
	@description,
	@website_url,
	@email,
	@headquarter_address,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                               tc.ID,
		"organization_id":                  tc.OrganizationID,
		"tenant_id":                        tc.TenantID,
		"mailing_list_id":                  tc.MailingListID,
		"logo_file_id":                     tc.LogoFileID,
		"dark_logo_file_id":                tc.DarkLogoFileID,
		"active":                           tc.Active,
		"slug":                             tc.Slug,
		"search_engine_indexing":           tc.SearchEngineIndexing,
		"non_disclosure_agreement_file_id": tc.NonDisclosureAgreementFileID,
		"default_domain_id":                tc.DefaultDomainID,
		"custom_domain_id":                 tc.CustomDomainID,
		"title":                            tc.Title,
		"description":                      tc.Description,
		"website_url":                      tc.WebsiteURL,
		"email":                            tc.Email,
		"headquarter_address":              tc.HeadquarterAddress,
		"created_at":                       tc.CreatedAt,
		"updated_at":                       tc.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "trust_centers_slug_key" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert compliance portal: %w", err)
	}

	return nil
}

func (tc *CompliancePortal) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE trust_centers
SET
	active = @active,
	slug = @slug,
	search_engine_indexing = @search_engine_indexing,
	logo_file_id = @logo_file_id,
	dark_logo_file_id = @dark_logo_file_id,
	non_disclosure_agreement_file_id = @non_disclosure_agreement_file_id,
	default_domain_id = @default_domain_id,
	custom_domain_id = @custom_domain_id,
	title = @title,
	description = @description,
	website_url = @website_url,
	email = @email,
	headquarter_address = @headquarter_address,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                               tc.ID,
		"logo_file_id":                     tc.LogoFileID,
		"dark_logo_file_id":                tc.DarkLogoFileID,
		"active":                           tc.Active,
		"slug":                             tc.Slug,
		"search_engine_indexing":           tc.SearchEngineIndexing,
		"non_disclosure_agreement_file_id": tc.NonDisclosureAgreementFileID,
		"default_domain_id":                tc.DefaultDomainID,
		"custom_domain_id":                 tc.CustomDomainID,
		"title":                            tc.Title,
		"description":                      tc.Description,
		"website_url":                      tc.WebsiteURL,
		"email":                            tc.Email,
		"headquarter_address":              tc.HeadquarterAddress,
		"updated_at":                       tc.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update compliance portal: %w", err)
	}

	return nil
}
