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

package vetting

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type (
	saveThirdPartyInfoParams struct {
		Name                          string   `json:"name" jsonschema:"Third party display name"`
		Description                   string   `json:"description,omitempty" jsonschema:"One-sentence description"`
		Category                      string   `json:"category,omitempty" jsonschema:"Category: ANALYTICS, CLOUD_PROVIDER, SECURITY, etc."`
		HeadquarterAddress            string   `json:"headquarter_address,omitempty" jsonschema:"Headquarters city and country"`
		LegalName                     string   `json:"legal_name,omitempty" jsonschema:"Legal entity name"`
		PrivacyPolicyURL              string   `json:"privacy_policy_url,omitempty" jsonschema:"Privacy policy URL"`
		ServiceLevelAgreementURL      string   `json:"service_level_agreement_url,omitempty" jsonschema:"SLA URL"`
		DataProcessingAgreementURL    string   `json:"data_processing_agreement_url,omitempty" jsonschema:"DPA URL"`
		BusinessAssociateAgreementURL string   `json:"business_associate_agreement_url,omitempty" jsonschema:"BAA URL"`
		SubprocessorsListURL          string   `json:"subprocessors_list_url,omitempty" jsonschema:"Subprocessors list URL"`
		SecurityPageURL               string   `json:"security_page_url,omitempty" jsonschema:"Security page URL"`
		TrustPageURL                  string   `json:"trust_page_url,omitempty" jsonschema:"Trust center URL"`
		TermsOfServiceURL             string   `json:"terms_of_service_url,omitempty" jsonschema:"Terms of service URL"`
		StatusPageURL                 string   `json:"status_page_url,omitempty" jsonschema:"Status page URL"`
		Certifications                []string `json:"certifications,omitempty" jsonschema:"Compliance certifications found"`
	}

	linkSubThirdPartyParams struct {
		Name        string `json:"name" jsonschema:"Sub-third-party company name"`
		Description string `json:"description,omitempty" jsonschema:"One-sentence description of what this third party does"`
		Category    string `json:"category,omitempty" jsonschema:"Category: ANALYTICS, CLOUD_PROVIDER, SECURITY, etc."`
		WebsiteURL  string `json:"website_url,omitempty" jsonschema:"Website URL if known"`
		Country     string `json:"country,omitempty" jsonschema:"Country where the sub-third-party operates"`
		Purpose     string `json:"purpose,omitempty" jsonschema:"Purpose or role of this sub-third-party"`
	}

	// PersistenceContext holds the DB and entity references the tools need.
	PersistenceContext struct {
		PG             *pg.Client
		ThirdPartyID   gid.GID
		OrganizationID gid.GID
		WebsiteURL     string
	}
)

func SaveThirdPartyInfoTool(pc *PersistenceContext) agent.Tool {
	return agent.FunctionTool(
		"save_third_party_info",
		"Persist the discovered third party metadata to the database. Call this once after completing the analysis with all fields you were able to discover. Only include fields that have actual values — omit fields with no data.",
		func(ctx context.Context, p saveThirdPartyInfoParams) (agent.ToolResult, error) {
			scope := coredata.NewScopeFromObjectID(pc.ThirdPartyID)

			err := pc.PG.WithTx(
				ctx,
				func(ctx context.Context, conn pg.Tx) error {
					thirdParty := &coredata.ThirdParty{}

					if err := thirdParty.LoadByID(ctx, conn, scope, pc.ThirdPartyID); err != nil {
						return fmt.Errorf("cannot load third party: %w", err)
					}

					if p.Name != "" {
						thirdParty.Name = p.Name
					}

					thirdParty.WebsiteURL = &pc.WebsiteURL

					if p.Category != "" {
						thirdParty.Category = coredata.ThirdPartyCategory(p.Category)
					}

					if p.Description != "" {
						thirdParty.Description = &p.Description
					}

					if p.HeadquarterAddress != "" {
						thirdParty.HeadquarterAddress = &p.HeadquarterAddress
					}

					if p.LegalName != "" {
						thirdParty.LegalName = &p.LegalName
					}

					if p.PrivacyPolicyURL != "" {
						thirdParty.PrivacyPolicyURL = &p.PrivacyPolicyURL
					}

					if p.ServiceLevelAgreementURL != "" {
						thirdParty.ServiceLevelAgreementURL = &p.ServiceLevelAgreementURL
					}

					if p.DataProcessingAgreementURL != "" {
						thirdParty.DataProcessingAgreementURL = &p.DataProcessingAgreementURL
					}

					if p.BusinessAssociateAgreementURL != "" {
						thirdParty.BusinessAssociateAgreementURL = &p.BusinessAssociateAgreementURL
					}

					if p.SubprocessorsListURL != "" {
						thirdParty.SubprocessorsListURL = &p.SubprocessorsListURL
					}

					if p.SecurityPageURL != "" {
						thirdParty.SecurityPageURL = &p.SecurityPageURL
					}

					if p.TrustPageURL != "" {
						thirdParty.TrustPageURL = &p.TrustPageURL
					}

					if p.TermsOfServiceURL != "" {
						thirdParty.TermsOfServiceURL = &p.TermsOfServiceURL
					}

					if p.StatusPageURL != "" {
						thirdParty.StatusPageURL = &p.StatusPageURL
					}

					if len(p.Certifications) > 0 {
						thirdParty.Certifications = p.Certifications
					}

					thirdParty.UpdatedAt = time.Now()

					if err := thirdParty.Update(ctx, conn, scope); err != nil {
						return fmt.Errorf("cannot update third party: %w", err)
					}

					return nil
				},
			)
			if err != nil {
				return agent.ToolResult{}, fmt.Errorf("cannot save third party info: %w", err)
			}

			return agent.ToolResult{Content: "Third party info saved successfully."}, nil
		},
	)
}

func LinkSubThirdPartyTool(pc *PersistenceContext) agent.Tool {
	return agent.FunctionTool(
		"link_sub_third_party",
		"Link a discovered sub-third-party (sub-processor, vendor dependency) to the parent. If a third party with the same name already exists in the organization it is linked as-is; otherwise a new one is created with the provided info. Call once per sub-third-party discovered.",
		func(ctx context.Context, p linkSubThirdPartyParams) (agent.ToolResult, error) {
			if p.Name == "" {
				return agent.ToolResult{Content: "Skipped: empty name."}, nil
			}

			scope := coredata.NewScopeFromObjectID(pc.ThirdPartyID)

			err := pc.PG.WithTx(
				ctx,
				func(ctx context.Context, conn pg.Tx) error {
					child := &coredata.ThirdParty{}

					err := child.LoadByNameAndOrganizationID(ctx, conn, scope, p.Name, pc.OrganizationID)
					if err != nil {
						if !errors.Is(err, coredata.ErrResourceNotFound) {
							return fmt.Errorf("cannot find child third party %q: %w", p.Name, err)
						}

						now := time.Now()
						child = &coredata.ThirdParty{
							ID:             gid.New(scope.GetTenantID(), coredata.ThirdPartyEntityType),
							OrganizationID: pc.OrganizationID,
							Name:           p.Name,
							Category:       coredata.ThirdPartyCategoryOther,
							FirstLevel:     false,
							CreatedAt:      now,
							UpdatedAt:      now,
						}

						if p.Description != "" {
							child.Description = &p.Description
						}

						if p.Category != "" {
							child.Category = coredata.ThirdPartyCategory(p.Category)
						}

						if p.WebsiteURL != "" {
							child.WebsiteURL = &p.WebsiteURL
						}

						if err := child.Insert(ctx, conn, scope); err != nil {
							return fmt.Errorf("cannot create child third party %q: %w", p.Name, err)
						}
					}

					relation := &coredata.ThirdPartyThirdParty{
						ParentThirdPartyID: pc.ThirdPartyID,
						ChildThirdPartyID:  child.ID,
						CreatedAt:          time.Now(),
					}

					if err := relation.Insert(ctx, conn, scope); err != nil {
						return fmt.Errorf("cannot link child third party %q: %w", p.Name, err)
					}

					return nil
				},
			)
			if err != nil {
				return agent.ToolResult{}, fmt.Errorf("cannot link sub third party: %w", err)
			}

			return agent.ToolResult{Content: fmt.Sprintf("Linked %q as sub third party.", p.Name)}, nil
		},
	)
}
