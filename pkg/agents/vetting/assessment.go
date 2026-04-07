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
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/agent/tools/browser"
	"go.probo.inc/probo/pkg/llm"
)

const (
	// DefaultMaxTokens is the fallback max-tokens budget used when the
	// vendor-assessor agent config does not specify a value. Sized to
	// leave headroom above the orchestrator's thinking budget on
	// Anthropic models.
	DefaultMaxTokens = 16384

	// AssessmentTimeout is the hard upper bound on a single assessment
	// run. This is also the timeout the CLI client should use.
	AssessmentTimeout = 20 * time.Minute
)

// vendorCategoryEnum is the canonical list of allowed values for
// VendorInfo.Category. It is duplicated into the jsonschema struct tag
// because Go struct tags must be compile-time string literals.
var vendorCategoryEnum = []string{
	"ANALYTICS", "ACCOUNTING", "CLOUD_MONITORING", "CLOUD_PROVIDER",
	"COLLABORATION", "CONSULTING", "CUSTOMER_SUPPORT",
	"DATA_STORAGE_AND_PROCESSING", "DOCUMENT_MANAGEMENT",
	"EMPLOYEE_MANAGEMENT", "ENGINEERING", "FINANCE", "IDENTITY_PROVIDER",
	"IT", "LEGAL", "MARKETING", "OFFICE_OPERATIONS", "OTHER",
	"PASSWORD_MANAGEMENT", "PRODUCT_AND_DESIGN", "PROFESSIONAL_SERVICES",
	"RECRUITING", "SALES", "SECURITY", "STAFFING", "VERSION_CONTROL",
}

// vendorTypeEnum is the canonical list of allowed values for
// VendorInfo.VendorType.
var vendorTypeEnum = []string{
	"SAAS", "INFRASTRUCTURE", "PROFESSIONAL_SERVICES", "STAFFING", "OTHER",
}

var (
	//go:embed prompts/extraction.txt
	extractionPrompt string
)

type (
	Config struct {
		Client         *llm.Client
		Model          string
		MaxTokens      int
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
		Name                          string         `json:"name" jsonschema:"Vendor display name as shown on the website"`
		Description                   string         `json:"description" jsonschema:"One-sentence description of what the vendor does"`
		Category                      string         `json:"category" jsonschema:"Vendor category; one of vendorCategoryEnum"`
		VendorType                    string         `json:"vendor_type" jsonschema:"Vendor type; one of vendorTypeEnum"`
		HeadquarterAddress            string         `json:"headquarter_address" jsonschema:"Vendor headquarters address (city, country) if mentioned"`
		LegalName                     string         `json:"legal_name" jsonschema:"Legal entity name if different from display name (e.g. 'Datadog, Inc.')"`
		PrivacyPolicyURL              string         `json:"privacy_policy_url" jsonschema:"URL to the vendor's privacy policy page"`
		ServiceLevelAgreementURL      string         `json:"service_level_agreement_url" jsonschema:"URL to the SLA page"`
		DataProcessingAgreementURL    string         `json:"data_processing_agreement_url" jsonschema:"URL to the DPA page"`
		BusinessAssociateAgreementURL string         `json:"business_associate_agreement_url" jsonschema:"URL to the BAA page if HIPAA-eligible"`
		SubprocessorsListURL          string         `json:"subprocessors_list_url" jsonschema:"URL to the public subprocessors list"`
		SecurityPageURL               string         `json:"security_page_url" jsonschema:"URL to the vendor's security page"`
		TrustPageURL                  string         `json:"trust_page_url" jsonschema:"URL to the trust center"`
		TermsOfServiceURL             string         `json:"terms_of_service_url" jsonschema:"URL to the terms of service"`
		StatusPageURL                 string         `json:"status_page_url" jsonschema:"URL to the vendor's status / uptime page"`
		BugBountyURL                  string         `json:"bug_bounty_url" jsonschema:"URL to the bug bounty or responsible disclosure program"`
		IncidentResponseURL           string         `json:"incident_response_url" jsonschema:"URL to incident response or post-mortem documentation"`
		DataLocations                 []string       `json:"data_locations" jsonschema:"Countries or regions where data is processed or stored (e.g. 'United States', 'EU', 'Germany')"`
		Certifications                []string       `json:"certifications" jsonschema:"Compliance certifications found (e.g. 'SOC 2 Type II', 'ISO 27001')"`
		Subprocessors                 []Subprocessor `json:"subprocessors" jsonschema:"Sub-processors discovered with name, country, purpose"`

		// Privacy classification (ISO 27701).
		PrivacyRole         string `json:"privacy_role" jsonschema:"Privacy role under ISO 27701: CONTROLLER, PROCESSOR, SUBPROCESSOR, NONE"`
		ProcessesPII        bool   `json:"processes_pii" jsonschema:"Whether the vendor processes personal data"`
		CrossBorderTransfer bool   `json:"cross_border_transfer" jsonschema:"Whether cross-border data transfers occur"`

		// Privacy risk fields.
		DPAStatus         string `json:"dpa_status" jsonschema:"DPA accessibility: AVAILABLE, AVAILABLE_ON_REQUEST, NOT_FOUND, BEHIND_LOGIN"`
		DSARCapability    string `json:"dsar_capability" jsonschema:"Brief summary of how the vendor handles Data Subject Access Requests"`
		DataMinimization  string `json:"data_minimization" jsonschema:"Brief summary of data minimization practices"`
		PurposeLimitation string `json:"purpose_limitation" jsonschema:"Brief summary of purpose limitation commitments"`
		RetentionPolicy   string `json:"retention_policy" jsonschema:"Brief summary of data retention policy"`
		DeletionPolicy    string `json:"deletion_policy" jsonschema:"Brief summary of data deletion policy"`

		// AI classification (ISO 42001).
		InvolvesAI bool     `json:"involves_ai" jsonschema:"Whether the vendor uses AI/ML in their product or service"`
		AIUseCases []string `json:"ai_use_cases" jsonschema:"Array of AI use case descriptions (e.g. 'content generation', 'fraud detection')"`

		// AI risk fields.
		AIGovernanceDocURL     string `json:"ai_governance_doc_url" jsonschema:"URL to AI governance or responsible AI documentation"`
		AITransparency         string `json:"ai_transparency" jsonschema:"Brief summary of model transparency findings"`
		BiasControls           string `json:"bias_controls" jsonschema:"Brief summary of bias detection and fairness measures"`
		HumanOversight         string `json:"human_oversight" jsonschema:"Brief summary of human oversight mechanisms for AI decisions"`
		TrainingDataGovernance string `json:"training_data_governance" jsonschema:"Brief summary of training data governance"`

		// Contractual clause analysis.
		PrivacyClauses []string `json:"privacy_clauses" jsonschema:"Notable privacy contractual clauses found (e.g. '72-hour breach notification', 'SCCs included')"`
		AIClauses      []string `json:"ai_clauses" jsonschema:"Notable AI contractual clauses found (e.g. 'Customer data not used for training')"`

		// Minimum acceptance baseline.
		MinimumBaselineMet bool     `json:"minimum_baseline_met" jsonschema:"Whether all hard-reject baseline criteria are met"`
		BaselineFailures   []string `json:"baseline_failures" jsonschema:"List of failed baseline criteria descriptions"`

		// Risk scoring.
		OverallRiskRating    string      `json:"overall_risk_rating" jsonschema:"Overall risk rating: Low, Medium, High"`
		OverallRiskScore     int         `json:"overall_risk_score" jsonschema:"Overall risk score from the report (0-100)"`
		Recommendation       string      `json:"recommendation" jsonschema:"Recommendation: APPROVE, APPROVE_WITH_CONDITIONS, ESCALATE, REJECT"`
		RiskScores           []RiskScore `json:"risk_scores" jsonschema:"Per-category risk scores from the Risk Summary table"`
		SecurityRiskScore    int         `json:"security_risk_score" jsonschema:"Security pillar risk score (0-100)"`
		PrivacyRiskScore     int         `json:"privacy_risk_score" jsonschema:"Privacy pillar risk score (0-100)"`
		AIRiskScore          int         `json:"ai_risk_score" jsonschema:"AI pillar risk score (0-100), 0 if no AI"`
		InformationGaps      []string    `json:"information_gaps" jsonschema:"Concise descriptions of information gaps from the report"`
		ProfessionalLicenses []string    `json:"professional_licenses" jsonschema:"Professional license descriptions for services firms (e.g. 'New York State Bar')"`
		IndustryMemberships  []string    `json:"industry_memberships" jsonschema:"Industry body memberships (e.g. 'AICPA', 'American Bar Association')"`
		InsuranceCoverage    string      `json:"insurance_coverage" jsonschema:"Description of professional liability or E&O insurance"`
	}

	Result struct {
		Document string
		Info     VendorInfo
	}
)

func NewAssessor(cfg Config) *Assessor {
	return &Assessor{cfg: cfg}
}

func (a *Assessor) Assess(ctx context.Context, websiteURL string, procedure string, reporter agent.ProgressReporter) (*Result, error) {
	u, err := url.Parse(websiteURL)
	if err != nil {
		return nil, fmt.Errorf("cannot parse website URL %q: %w", websiteURL, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("website URL must use http or https, got %q", u.Scheme)
	}
	if u.Hostname() == "" {
		return nil, fmt.Errorf("website URL %q has no host", websiteURL)
	}

	// Detach from the caller's context (typically the HTTP request) so
	// that the assessment is not cancelled when the client disconnects.
	// A dedicated timeout prevents the assessment from running forever.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), AssessmentTimeout)
	defer cancel()

	vendorBrowser := browser.NewBrowser(ctx, a.cfg.ChromeAddr)
	defer vendorBrowser.Close()

	vendorBrowser.SetAllowedDomain(u.Hostname())

	// Create an unrestricted browser for web search agents that need to
	// follow links to external sites (news, reviews, etc.).
	researchBrowser := browser.NewBrowser(ctx, a.cfg.ChromeAddr)
	defer researchBrowser.Close()

	orchestrator, err := newOrchestratorAgent(
		a.cfg.Client,
		a.cfg.Model,
		a.cfg.MaxTokens,
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
	outputType, err := vendorInfoOutputType()
	if err != nil {
		return nil, fmt.Errorf("cannot build vendor info output type: %w", err)
	}

	extractor := agent.New(
		"vendor_info_extractor",
		a.cfg.Client,
		agent.WithInstructions(extractionPrompt),
		agent.WithModel(a.cfg.Model),
		agent.WithMaxTokens(a.cfg.MaxTokens),
		agent.WithLogger(a.cfg.Logger),
		agent.WithOutputType(outputType),
	)

	result, err := extractor.Run(
		ctx,
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

	var info VendorInfo
	if err := json.Unmarshal([]byte(result.FinalMessage().Text()), &info); err != nil {
		return nil, fmt.Errorf("cannot parse vendor info output: %w", err)
	}

	return &info, nil
}

// vendorInfoOutputType builds the VendorInfo structured output type and
// decorates its JSON Schema with explicit enum constraints on fields
// whose allowed values live in package-level slices. jsonschema-go only
// reads struct tags as free-form descriptions, so the enum list cannot
// be encoded in the tag itself.
func vendorInfoOutputType() (*agent.OutputType, error) {
	outputType, err := agent.NewOutputType[VendorInfo]("vendor_info")
	if err != nil {
		return nil, fmt.Errorf("cannot create vendor info output type: %w", err)
	}

	var schema map[string]any
	if err := json.Unmarshal(outputType.Schema, &schema); err != nil {
		return nil, fmt.Errorf("cannot unmarshal vendor info schema: %w", err)
	}

	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("vendor info schema has no properties")
	}

	enums := map[string][]string{
		"category":    vendorCategoryEnum,
		"vendor_type": vendorTypeEnum,
	}
	for field, values := range enums {
		prop, ok := properties[field].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("vendor info schema has no %q property", field)
		}
		prop["enum"] = values
	}

	decorated, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal decorated vendor info schema: %w", err)
	}
	outputType.Schema = decorated

	return outputType, nil
}
