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

package github

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func resolveGitHubThirdParty(
	ctx context.Context,
	conn pg.Tx,
	scope coredata.Scoper,
	organizationID gid.GID,
) (*coredata.ThirdParty, error) {
	thirdParty, err := importGitHubThirdPartyFromCatalog(ctx, conn, scope, organizationID)
	if err != nil {
		return nil, err
	}

	if thirdParty != nil {
		return thirdParty, nil
	}

	thirdParty, err = loadLevel1GitHubThirdPartyByName(ctx, conn, scope, organizationID, thirdPartyName)
	if err != nil {
		return nil, err
	}

	if thirdParty != nil {
		return thirdParty, nil
	}

	return createMinimalGitHubThirdParty(ctx, conn, scope, organizationID)
}

func importGitHubThirdPartyFromCatalog(
	ctx context.Context,
	conn pg.Tx,
	scope coredata.Scoper,
	organizationID gid.GID,
) (*coredata.ThirdParty, error) {
	var commonParty coredata.CommonThirdParty

	err := commonParty.LoadByName(ctx, conn, thirdPartyName)
	if errors.Is(err, coredata.ErrResourceNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("cannot load github common third party: %w", err)
	}

	existing := &coredata.ThirdParty{}

	err = existing.LoadByOrganizationIDAndCommonThirdPartyID(
		ctx,
		conn,
		scope,
		organizationID,
		commonParty.ID,
	)
	if err == nil {
		if !isLevel1RootThirdParty(existing) {
			return nil, fmt.Errorf("github third party imported from catalog exists but is not level 1")
		}

		return existing, nil
	}

	if !errors.Is(err, coredata.ErrResourceNotFound) {
		return nil, fmt.Errorf("cannot load github third party by common id: %w", err)
	}

	return createOrgThirdPartyFromCommon(ctx, conn, scope, organizationID, &commonParty)
}

func loadLevel1GitHubThirdPartyByName(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	organizationID gid.GID,
	name string,
) (*coredata.ThirdParty, error) {
	q := `
SELECT
    id,
    organization_id,
    parent_third_party_id,
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
    level,
    vetting_status,
    vetting_website_url,
    vetting_procedure,
    vetting_processing_started_at,
    vetting_error_message,
    created_at,
    updated_at
FROM
    third_parties
WHERE
    %s
    AND organization_id = @organization_id
    AND lower(name) = lower(@name)
    AND level = 1
    AND parent_third_party_id IS NULL
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
		"name":            name,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query github third party by name: %w", err)
	}

	defer rows.Close()

	thirdParty, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[coredata.ThirdParty])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("cannot collect github third party by name: %w", err)
	}

	return &thirdParty, nil
}

func createOrgThirdPartyFromCommon(
	ctx context.Context,
	conn pg.Tx,
	scope coredata.Scoper,
	organizationID gid.GID,
	commonParty *coredata.CommonThirdParty,
) (*coredata.ThirdParty, error) {
	now := time.Now()
	commonID := commonParty.ID

	certifications := commonParty.Certifications
	if certifications == nil {
		certifications = []string{}
	}

	thirdParty := &coredata.ThirdParty{
		ID:                            gid.New(scope.GetTenantID(), coredata.ThirdPartyEntityType),
		OrganizationID:                organizationID,
		CommonThirdPartyID:            &commonID,
		Name:                          commonParty.Name,
		Category:                      commonParty.Category,
		HeadquarterAddress:            commonParty.HeadquarterAddress,
		LegalName:                     commonParty.LegalName,
		WebsiteURL:                    commonParty.WebsiteURL,
		PrivacyPolicyURL:              commonParty.PrivacyPolicyURL,
		ServiceLevelAgreementURL:      commonParty.ServiceLevelAgreementURL,
		DataProcessingAgreementURL:    commonParty.DataProcessingAgreementURL,
		BusinessAssociateAgreementURL: commonParty.BusinessAssociateAgreementURL,
		SubprocessorsListURL:          commonParty.SubprocessorsListURL,
		Certifications:                certifications,
		Countries:                     coredata.CountryCodes{},
		StatusPageURL:                 commonParty.StatusPageURL,
		TermsOfServiceURL:             commonParty.TermsOfServiceURL,
		SecurityPageURL:               commonParty.SecurityPageURL,
		TrustPageURL:                  commonParty.TrustPageURL,
		ShowOnTrustCenter:             false,
		Level:                         1,
		CreatedAt:                     now,
		UpdatedAt:                     now,
	}

	if err := thirdParty.Insert(ctx, conn, scope); err != nil {
		return nil, fmt.Errorf("cannot insert github third party from catalog: %w", err)
	}

	return thirdParty, nil
}

func createMinimalGitHubThirdParty(
	ctx context.Context,
	conn pg.Tx,
	scope coredata.Scoper,
	organizationID gid.GID,
) (*coredata.ThirdParty, error) {
	now := time.Now()

	thirdParty := &coredata.ThirdParty{
		ID:             gid.New(scope.GetTenantID(), coredata.ThirdPartyEntityType),
		OrganizationID: organizationID,
		Name:           thirdPartyName,
		Category:       coredata.ThirdPartyCategoryVersionControl,
		WebsiteURL:     new("https://github.com"),
		Level:          1,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := thirdParty.Insert(ctx, conn, scope); err != nil {
		return nil, fmt.Errorf("cannot insert github third party: %w", err)
	}

	return thirdParty, nil
}

func loadGitHubLinkedMeasures(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	thirdPartyID gid.GID,
) ([]ExistingMeasure, error) {
	var measures coredata.Measures

	if err := measures.LoadByThirdPartyID(
		ctx,
		conn,
		scope,
		thirdPartyID,
		nil,
		coredata.NewMeasureFilter(nil, nil, nil),
	); err != nil {
		return nil, fmt.Errorf("cannot load github-linked measures: %w", err)
	}

	out := make([]ExistingMeasure, 0, len(measures))

	for _, m := range measures {
		out = append(out, ExistingMeasure{
			ID:          m.ID,
			Name:        m.Name,
			Description: m.Description,
			Category:    m.Category,
			State:       m.State,
		})
	}

	return out, nil
}

func isLevel1RootThirdParty(thirdParty *coredata.ThirdParty) bool {
	return thirdParty.Level == 1 && thirdParty.ParentThirdPartyID == nil
}

func isNotFound(err error) bool {
	return errors.Is(err, coredata.ErrResourceNotFound)
}

// EnsureThirdParty resolves or creates the level-1 GitHub third party.
func EnsureThirdParty(
	ctx context.Context,
	pgClient *pg.Client,
	scope coredata.Scoper,
	organizationID gid.GID,
) (*coredata.ThirdParty, error) {
	var thirdParty *coredata.ThirdParty

	err := pgClient.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			tp, err := resolveGitHubThirdParty(ctx, tx, scope, organizationID)
			if err != nil {
				return err
			}

			thirdParty = tp

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot ensure github third party: %w", err)
	}

	return thirdParty, nil
}
