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
	_ "embed"
	"fmt"
	"strings"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/agent/tools/browser"
	"go.probo.inc/probo/pkg/agent/tools/search"
	"go.probo.inc/probo/pkg/agent/tools/security"
	"go.probo.inc/probo/pkg/llm"
)

var (
	//go:embed orchestrator_base_prompt.txt
	orchestratorBasePrompt string

	//go:embed default_procedure.txt
	defaultProcedure string
)

func newOrchestratorAgent(
	client *llm.Client,
	model string,
	maxTokens int,
	procedure string,
	logger *log.Logger,
	vendorBrowser *browser.Browser,
	researchBrowser *browser.Browser,
	searchEndpoint string,
	reporter agent.ProgressReporter,
) (*agent.Agent, error) {
	vendorToolset := browser.NewReadOnlyToolset(vendorBrowser)
	researchToolset := browser.NewInteractiveToolset(researchBrowser)
	securityToolset := security.NewToolset()

	readOnlyBrowserTools, err := vendorToolset.Tools()
	if err != nil {
		return nil, fmt.Errorf("cannot build read-only browser tools: %w", err)
	}

	// Build unrestricted browser tools for the subprocessor agent.
	// Subprocessor lists are frequently hosted on external platforms
	// (OneTrust, Transcend, Notion, etc.), so the domain-restricted
	// vendor browser cannot reach them.
	unrestrictedBrowserTools, err := researchToolset.Tools()
	if err != nil {
		return nil, fmt.Errorf("cannot build unrestricted browser tools: %w", err)
	}

	securityTools, err := securityToolset.Tools()
	if err != nil {
		return nil, fmt.Errorf("cannot build security tools: %w", err)
	}

	loggerOpt := agent.WithLogger(logger)

	maxTokensOpt := agent.WithMaxTokens(maxTokens)

	subAgentOpts := func(step string) []agent.Option {
		opts := []agent.Option{loggerOpt, maxTokensOpt}
		if reporter != nil {
			opts = append(opts, agent.WithHooks(newSubProgressHooks(reporter, step)))
		}
		return opts
	}

	crawler, err := newCrawlerAgent(client, model, readOnlyBrowserTools, subAgentOpts("crawl_vendor_website")...)
	if err != nil {
		return nil, fmt.Errorf("cannot create crawler agent: %w", err)
	}

	analyzer, err := newDocumentAnalyzerAgent(client, model, readOnlyBrowserTools, subAgentOpts("analyze_document")...)
	if err != nil {
		return nil, fmt.Errorf("cannot create document analyzer agent: %w", err)
	}

	securityAssessor, err := newSecurityAssessorAgent(client, model, securityTools, subAgentOpts("assess_security")...)
	if err != nil {
		return nil, fmt.Errorf("cannot create security assessor agent: %w", err)
	}

	compliance, err := newComplianceAssessorAgent(client, model, readOnlyBrowserTools, subAgentOpts("assess_compliance")...)
	if err != nil {
		return nil, fmt.Errorf("cannot create compliance assessor agent: %w", err)
	}

	market, err := newMarketPresenceAgent(client, model, readOnlyBrowserTools, subAgentOpts("assess_market_presence")...)
	if err != nil {
		return nil, fmt.Errorf("cannot create market presence agent: %w", err)
	}
	subprocessorTools := unrestrictedBrowserTools
	if searchEndpoint != "" {
		searchTool, err := search.WebSearchTool(searchEndpoint)
		if err != nil {
			return nil, fmt.Errorf("cannot build subprocessor search tool: %w", err)
		}
		subprocessorTools = append(subprocessorTools, searchTool)
	}
	subprocessor, err := newSubprocessorAgent(client, model, subprocessorTools, subAgentOpts("extract_subprocessors")...)
	if err != nil {
		return nil, fmt.Errorf("cannot create subprocessor agent: %w", err)
	}
	dataProcessing, err := newDataProcessingAgent(client, model, readOnlyBrowserTools, subAgentOpts("assess_data_processing")...)
	if err != nil {
		return nil, fmt.Errorf("cannot create data processing agent: %w", err)
	}
	aiRisk, err := newAIRiskAgent(client, model, readOnlyBrowserTools, subAgentOpts("assess_ai_risk")...)
	if err != nil {
		return nil, fmt.Errorf("cannot create ai risk agent: %w", err)
	}
	incidentResponse, err := newIncidentResponseAgent(client, model, readOnlyBrowserTools, subAgentOpts("assess_incident_response")...)
	if err != nil {
		return nil, fmt.Errorf("cannot create incident response agent: %w", err)
	}

	businessContinuity, err := newBusinessContinuityAgent(client, model, readOnlyBrowserTools, subAgentOpts("assess_business_continuity")...)
	if err != nil {
		return nil, fmt.Errorf("cannot create business continuity agent: %w", err)
	}
	professionalStanding, err := newProfessionalStandingAgent(client, model, readOnlyBrowserTools, subAgentOpts("assess_professional_standing")...)
	if err != nil {
		return nil, fmt.Errorf("cannot create professional standing agent: %w", err)
	}
	regulatoryCompliance, err := newRegulatoryComplianceAgent(client, model, readOnlyBrowserTools, subAgentOpts("assess_regulatory_compliance")...)
	if err != nil {
		return nil, fmt.Errorf("cannot create regulatory compliance agent: %w", err)
	}

	tools := []agent.Tool{
		crawler.AsTool(
			"crawl_vendor_website",
			"Crawl a vendor website to discover security, compliance, privacy, and legal pages. Returns structured JSON with categorized URLs (vendor_name, vendor_domain, discovered_urls, notes). Input: the vendor's main website URL.",
		),
		securityAssessor.AsTool(
			"assess_security",
			"Perform technical security checks on a domain. Returns structured JSON with per-check results (ssl, headers, dmarc, spf, breaches, dnssec, csp, cors, dns, whois) each with status (pass/warning/fail/error) and details. Input: the vendor's domain name (e.g. example.com).",
		),
		analyzer.AsTool(
			"analyze_document",
			"Analyze a specific document page (privacy policy, DPA, ToS) and extract key provisions. Returns structured JSON with document_type, retention, locations, GDPR/CCPA indicators, clauses, and summary. Input: the document URL.",
		),
		compliance.AsTool(
			"assess_compliance",
			"Identify certifications and compliance frameworks from a trust/compliance page. Returns structured JSON with certifications (name, status, details), audit reports, and frameworks. Input: the trust or compliance page URL.",
		),
		market.AsTool(
			"assess_market_presence",
			"Analyze a vendor's market presence. Returns structured JSON with notable_customers, case_studies, partnerships, company_size_signals, funding_info, and market_position. Input: the vendor's main website URL.",
		),
		subprocessor.AsTool(
			"extract_subprocessors",
			"Find and extract the list of sub-processors from a vendor's website. Returns structured JSON with subprocessors (name, country, purpose), total_count, and source. Input: the vendor's main website URL or a known subprocessors page URL.",
		),
		dataProcessing.AsTool(
			"assess_data_processing",
			"Assess data processing practices. Returns structured JSON with encryption, retention, deletion, data locations, transfer mechanisms, DPA status, DSAR handling, and rating. Input: a relevant page URL (privacy policy, DPA, security page, or trust center).",
		),
		incidentResponse.AsTool(
			"assess_incident_response",
			"Evaluate incident response capabilities. Returns structured JSON with ir_plan, notification_timeline, status_page, post_mortems, recent_incidents, security_contact, and rating. Input: a relevant page URL (security page, trust center, or status page).",
		),
		businessContinuity.AsTool(
			"assess_business_continuity",
			"Evaluate business continuity and disaster recovery. Returns structured JSON with dr_plan, rto, rpo, cloud_providers, uptime_sla, regions, backup_strategy, and rating. Input: a relevant page URL (SLA page, trust center, or infrastructure docs).",
		),
		professionalStanding.AsTool(
			"assess_professional_standing",
			"Evaluate professional standing for services firms. Returns structured JSON with licensing, memberships, insurance, team_credentials, coi_policy, and rating. Input: relevant page URL (team page, about page, credentials page).",
		),
		aiRisk.AsTool(
			"assess_ai_risk",
			"Evaluate AI governance (ISO 42001). Returns structured JSON with ai_involvement, use_cases, model_transparency, bias_controls, customer_data_training, human_oversight, and rating. Input: relevant page URL (AI policy, trust center, responsible AI page, or main website).",
		),
		regulatoryCompliance.AsTool(
			"assess_regulatory_compliance",
			"Deep regulatory compliance check. Returns structured JSON with per-framework assessment (gdpr, hipaa, pci_dss, sox) each with articles, status, and notes. Input: relevant page URL (DPA, compliance page, trust center).",
		),
	}

	if searchEndpoint != "" {
		researchBrowserTools, err := researchToolset.Tools()
		if err != nil {
			return nil, fmt.Errorf("cannot build research browser tools: %w", err)
		}

		searchTool, err := search.WebSearchTool(searchEndpoint)
		if err != nil {
			return nil, fmt.Errorf("cannot build web search tool: %w", err)
		}

		websearchTools := append([]agent.Tool{searchTool}, researchBrowserTools...)
		websearch, err := newWebSearchAgent(client, model, websearchTools, subAgentOpts("research_vendor_externally")...)
		if err != nil {
			return nil, fmt.Errorf("cannot create web search agent: %w", err)
		}

		tools = append(tools, websearch.AsTool(
			"research_vendor_externally",
			"Search the open web for external signals about the vendor. Returns structured JSON with security_incidents, regulatory_actions, customer_sentiment, recent_news, red_flags, and positive_signals. Input: the vendor's name and domain.",
		))

		// Build tools for sub-agents that need search + unrestricted browsing.
		govDBTool, err := search.CheckGovernmentDBTool(searchEndpoint)
		if err != nil {
			return nil, fmt.Errorf("cannot build government DB tool: %w", err)
		}

		waybackTool, err := search.CheckWaybackTool()
		if err != nil {
			return nil, fmt.Errorf("cannot build wayback tool: %w", err)
		}

		diffTool, err := search.DiffDocumentsTool()
		if err != nil {
			return nil, fmt.Errorf("cannot build diff tool: %w", err)
		}

		financialTools := append([]agent.Tool{searchTool, govDBTool, waybackTool}, researchBrowserTools...)
		financialStability, err := newFinancialStabilityAgent(client, model, financialTools, subAgentOpts("assess_financial_stability")...)
		if err != nil {
			return nil, fmt.Errorf("cannot create financial stability agent: %w", err)
		}

		codeSecurityTools := append([]agent.Tool{searchTool}, researchBrowserTools...)
		codeSecurity, err := newCodeSecurityAgent(client, model, codeSecurityTools, subAgentOpts("assess_code_security")...)
		if err != nil {
			return nil, fmt.Errorf("cannot create code security agent: %w", err)
		}

		comparisonTools := append([]agent.Tool{searchTool, diffTool}, researchBrowserTools...)
		vendorComparison, err := newVendorComparisonAgent(client, model, comparisonTools, subAgentOpts("compare_vendor")...)
		if err != nil {
			return nil, fmt.Errorf("cannot create vendor comparison agent: %w", err)
		}

		tools = append(
			tools,
			financialStability.AsTool(
				"assess_financial_stability",
				"Evaluate vendor financial stability. Returns structured JSON with company_age, funding, employee_count, legal_standing, ownership, risk_signals, overall_assessment, and confidence. Input: vendor name and website URL.",
			),
			codeSecurity.AsTool(
				"assess_code_security",
				"Evaluate open-source code security posture. Returns structured JSON with has_public_repos, security_advisories, dependency_management, release_cadence, security_policy, overall_assessment, and risk_signals. Input: vendor name and website URL.",
			),
			vendorComparison.AsTool(
				"compare_vendor",
				"Find and compare alternative vendors. Returns structured JSON with alternatives (name, certifications, security_score), comparison_summary, vendor_strengths, vendor_weaknesses, and overall_position. Input: vendor name, category, and website URL.",
			),
		)
	}

	if procedure == "" {
		procedure = defaultProcedure
	}
	systemPrompt := strings.Replace(orchestratorBasePrompt, "{procedure}", procedure, 1)

	opts := []agent.Option{
		agent.WithLogger(logger),
		agent.WithInstructions(systemPrompt),
		agent.WithModel(model),
		agent.WithMaxTokens(maxTokens),
		agent.WithTools(tools...),
		agent.WithMaxTurns(140),
		agent.WithParallelToolCalls(true),
		agent.WithThinking(40000),
	}

	if reporter != nil {
		opts = append(opts, agent.WithHooks(newProgressHooks(reporter)))
	}

	return agent.New(
		"vendor_assessment_orchestrator",
		client,
		opts...,
	), nil
}
