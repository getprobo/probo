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

package vetting

import (
	"context"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type (
	saveThirdPartyInfoToolParams struct {
		Name                          string   `json:"name" jsonschema:"Third party display name"`
		Description                   string   `json:"description" jsonschema:"One-sentence description"`
		Category                      string   `json:"category" jsonschema:"Category: ANALYTICS, CLOUD_PROVIDER, SECURITY, etc."`
		HeadquarterAddress            string   `json:"headquarter_address" jsonschema:"Headquarters city and country"`
		LegalName                     string   `json:"legal_name" jsonschema:"Legal entity name"`
		PrivacyPolicyURL              string   `json:"privacy_policy_url" jsonschema:"Privacy policy URL"`
		ServiceLevelAgreementURL      string   `json:"service_level_agreement_url" jsonschema:"SLA URL"`
		DataProcessingAgreementURL    string   `json:"data_processing_agreement_url" jsonschema:"DPA URL"`
		BusinessAssociateAgreementURL string   `json:"business_associate_agreement_url" jsonschema:"BAA URL"`
		SubprocessorsListURL          string   `json:"subprocessors_list_url" jsonschema:"Subprocessors list URL"`
		SecurityPageURL               string   `json:"security_page_url" jsonschema:"Security page URL"`
		TrustPageURL                  string   `json:"trust_page_url" jsonschema:"Trust center URL"`
		TermsOfServiceURL             string   `json:"terms_of_service_url" jsonschema:"Terms of service URL"`
		StatusPageURL                 string   `json:"status_page_url" jsonschema:"Status page URL"`
		Certifications                []string `json:"certifications" jsonschema:"Compliance certifications found"`
	}

	saveThirdPartyInfoParams struct {
		saveThirdPartyInfoToolParams
		Countries coredata.CountryCodes
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
	return vettingFunctionTool(
		"save_third_party_info",
		"Persist the discovered third party metadata to the database. Call this once after completing the analysis. Use an empty string for any field you could not discover.",
		func(ctx context.Context, p saveThirdPartyInfoToolParams) (agent.ToolResult, error) {
			scope := coredata.NewScopeFromObjectID(pc.ThirdPartyID)

			err := pc.PG.WithTx(
				ctx,
				func(ctx context.Context, conn pg.Tx) error {
					thirdParty := &coredata.ThirdParty{}

					if err := thirdParty.LoadByID(ctx, conn, scope, pc.ThirdPartyID); err != nil {
						return fmt.Errorf("cannot load third party: %w", err)
					}

					if p.Category != "" {
						if _, err := parseThirdPartyCategory(p.Category); err != nil {
							return err
						}
					}

					applySaveParams(thirdParty, pc.WebsiteURL, saveThirdPartyInfoParams{
						saveThirdPartyInfoToolParams: p,
					})
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
	return vettingFunctionTool(
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
					return linkSubThirdParty(ctx, conn, scope, pc, p)
				},
			)
			if err != nil {
				return agent.ToolResult{}, fmt.Errorf("cannot link sub third party: %w", err)
			}

			return agent.ToolResult{Content: fmt.Sprintf("Linked %q as sub third party.", p.Name)}, nil
		},
	)
}

func parseThirdPartyCategory(raw string) (coredata.ThirdPartyCategory, error) {
	category := coredata.ThirdPartyCategory(raw)
	if !category.IsValid() {
		return "", fmt.Errorf("invalid third party category %q", raw)
	}

	return category, nil
}
