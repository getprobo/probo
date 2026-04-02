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
	_ "embed"
	"fmt"
	"net/url"
	"time"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/agent/tools/browser"
	"go.probo.inc/probo/pkg/llm"
)

var (
	//go:embed extraction_prompt.txt
	extractionPrompt string
)

type (
	Config struct {
		Client         *llm.Client
		Model          string
		ChromeAddr     string
		SearchEndpoint string
		Logger         *log.Logger
	}

	Assessor struct {
		cfg Config
	}

	Subprocessor struct {
		Name    string `json:"name"`
		Country string `json:"country"`
		Purpose string `json:"purpose"`
	}

	RiskScore struct {
		Category string `json:"category"`
		Rating   string `json:"rating"`
		Notes    string `json:"notes"`
	}

	VendorInfo struct {
		Name                          string         `json:"name"`
		Description                   string         `json:"description"`
		Category                      string         `json:"category"`
		VendorType                    string         `json:"vendor_type"`
		HeadquarterAddress            string         `json:"headquarter_address"`
		LegalName                     string         `json:"legal_name"`
		PrivacyPolicyURL              string         `json:"privacy_policy_url"`
		ServiceLevelAgreementURL      string         `json:"service_level_agreement_url"`
		DataProcessingAgreementURL    string         `json:"data_processing_agreement_url"`
		BusinessAssociateAgreementURL string         `json:"business_associate_agreement_url"`
		SubprocessorsListURL          string         `json:"subprocessors_list_url"`
		SecurityPageURL               string         `json:"security_page_url"`
		TrustPageURL                  string         `json:"trust_page_url"`
		TermsOfServiceURL             string         `json:"terms_of_service_url"`
		StatusPageURL                 string         `json:"status_page_url"`
		BugBountyURL                  string         `json:"bug_bounty_url"`
		IncidentResponseURL           string         `json:"incident_response_url"`
		DataLocations                 []string       `json:"data_locations"`
		Certifications                []string       `json:"certifications"`
		Subprocessors                 []Subprocessor `json:"subprocessors"`

		// Privacy classification (ISO 27701).
		PrivacyRole         string `json:"privacy_role"`
		ProcessesPII        bool   `json:"processes_pii"`
		CrossBorderTransfer bool   `json:"cross_border_transfer"`

		// Privacy risk fields.
		DPAStatus         string `json:"dpa_status"`
		DSARCapability    string `json:"dsar_capability"`
		DataMinimization  string `json:"data_minimization"`
		PurposeLimitation string `json:"purpose_limitation"`
		RetentionPolicy   string `json:"retention_policy"`
		DeletionPolicy    string `json:"deletion_policy"`

		// AI classification (ISO 42001).
		InvolvesAI bool     `json:"involves_ai"`
		AIUseCases []string `json:"ai_use_cases"`

		// AI risk fields.
		AIGovernanceDocURL     string `json:"ai_governance_doc_url"`
		AITransparency         string `json:"ai_transparency"`
		BiasControls           string `json:"bias_controls"`
		HumanOversight         string `json:"human_oversight"`
		TrainingDataGovernance string `json:"training_data_governance"`

		// Contractual clause analysis.
		PrivacyClauses []string `json:"privacy_clauses"`
		AIClauses      []string `json:"ai_clauses"`

		// Minimum acceptance baseline.
		MinimumBaselineMet bool     `json:"minimum_baseline_met"`
		BaselineFailures   []string `json:"baseline_failures"`

		// Risk scoring.
		OverallRiskRating    string      `json:"overall_risk_rating"`
		OverallRiskScore     int         `json:"overall_risk_score"`
		Recommendation       string      `json:"recommendation"`
		RiskScores           []RiskScore `json:"risk_scores"`
		SecurityRiskScore    int         `json:"security_risk_score"`
		PrivacyRiskScore     int         `json:"privacy_risk_score"`
		AIRiskScore          int         `json:"ai_risk_score"`
		InformationGaps      []string    `json:"information_gaps"`
		ProfessionalLicenses []string    `json:"professional_licenses"`
		IndustryMemberships  []string    `json:"industry_memberships"`
		InsuranceCoverage    string      `json:"insurance_coverage"`
	}

	Result struct {
		Document string
		Info     VendorInfo
	}

	// CrawlResult is the structured output from the crawler agent.
	CrawlResult struct {
		VendorName     string            `json:"vendor_name"`
		VendorDomain   string            `json:"vendor_domain"`
		DiscoveredURLs map[string]string `json:"discovered_urls"`
		Notes          string            `json:"notes"`
	}
)

func NewAssessor(cfg Config) *Assessor {
	return &Assessor{cfg: cfg}
}

func (a *Assessor) Assess(ctx context.Context, websiteURL string, procedure string, reporter agent.ProgressReporter) (*Result, error) {
	// Detach from the caller's context (typically the HTTP request) so
	// that the assessment is not cancelled when the client disconnects.
	// A dedicated timeout prevents the assessment from running forever.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Minute)
	defer cancel()

	vendorBrowser := browser.NewBrowser(ctx, a.cfg.ChromeAddr)
	defer vendorBrowser.Close()

	if u, err := url.Parse(websiteURL); err == nil {
		vendorBrowser.SetAllowedDomain(u.Hostname())
	}

	// Create an unrestricted browser for web search agents that need to
	// follow links to external sites (news, reviews, etc.).
	researchBrowser := browser.NewBrowser(ctx, a.cfg.ChromeAddr)
	defer researchBrowser.Close()

	orchestrator, err := newOrchestratorAgent(
		a.cfg.Client,
		a.cfg.Model,
		procedure,
		a.cfg.Logger,
		vendorBrowser,
		researchBrowser,
		a.cfg.SearchEndpoint,
		reporter,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create orchestrator agent: %w", err)
	}

	result, err := orchestrator.Run(
		ctx,
		[]llm.Message{
			{
				Role:  llm.RoleUser,
				Parts: []llm.Part{llm.TextPart{Text: websiteURL}},
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot assess vendor: %w", err)
	}

	document := result.FinalMessage().Text()

	reportProgress(ctx, reporter, "extract_vendor_info", agent.ProgressEventStepStarted)

	info, err := a.extractVendorInfo(ctx, document)
	if err != nil {
		reportProgress(ctx, reporter, "extract_vendor_info", agent.ProgressEventStepFailed)
		return nil, fmt.Errorf("cannot extract vendor info: %w", err)
	}

	reportProgress(ctx, reporter, "extract_vendor_info", agent.ProgressEventStepCompleted)

	return &Result{
		Document: document,
		Info:     *info,
	}, nil
}

func (a *Assessor) extractVendorInfo(ctx context.Context, document string) (*VendorInfo, error) {
	extractor := agent.New(
		"vendor_info_extractor",
		a.cfg.Client,
		agent.WithInstructions(extractionPrompt),
		agent.WithModel(a.cfg.Model),
		agent.WithLogger(a.cfg.Logger),
	)

	typedResult, err := agent.RunTyped[VendorInfo](
		ctx,
		extractor,
		[]llm.Message{
			{
				Role:  llm.RoleUser,
				Parts: []llm.Part{llm.TextPart{Text: document}},
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot extract vendor info: %w", err)
	}

	return &typedResult.Output, nil
}
