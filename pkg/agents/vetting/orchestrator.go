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
	//go:embed prompts/orchestrator_base.txt
	orchestratorBasePrompt string

	//go:embed prompts/default_procedure.txt
	defaultProcedure string
)

const (
	// orchestratorMaxTurns bounds the orchestrator loop. Each turn typically
	// dispatches one sub-agent in parallel; with 16 sub-agents and a few
	// retries we need ~140 turns of headroom before timing out.
	orchestratorMaxTurns = 140

	// orchestratorThinkingBudget is the extended-thinking budget for the
	// orchestrator. It is high because the orchestrator must reason over
	// the outputs of all 16 sub-agents to produce the final report.
	orchestratorThinkingBudget = 40000
)

// subAgentEntry binds a sub-agent's LLM-facing name and description to
// the tools it needs and a typed builder. The orchestrator iterates over
// a slice of these and turns each into an agent + AsTool wrapper.
type subAgentEntry struct {
	toolName    string
	description string
	tools       []agent.Tool
	build       subAgentBuilder
}

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
	readOnlyBrowserTools := browser.NewReadOnlyToolset(vendorBrowser).Tools()

	// Unrestricted browser tools for sub-agents that need to follow links
	// to external sites (subprocessor lists hosted on OneTrust/Transcend,
	// research, vendor comparison).
	unrestrictedBrowserTools := browser.NewInteractiveToolset(researchBrowser).Tools()

	securityTools := security.NewToolset().Tools()

	maxTokensOpt := agent.WithMaxTokens(maxTokens)
	loggerOpt := agent.WithLogger(logger)

	subAgentOpts := func(step string) []agent.Option {
		opts := []agent.Option{loggerOpt, maxTokensOpt}
		if reporter != nil {
			opts = append(opts, agent.WithHooks(newSubProgressHooks(reporter, step)))
		}
		return opts
	}

	// Subprocessor agent benefits from web search when available so it can
	// find subprocessor pages hosted on third-party platforms.
	subprocessorTools := unrestrictedBrowserTools
	if searchEndpoint != "" {
		subprocessorTools = append(subprocessorTools, search.WebSearchTool(searchEndpoint))
	}

	// Core sub-agents that always run.
	entries := []subAgentEntry{
		{
			toolName:    "crawl_vendor_website",
			description: "Crawl a vendor website to discover security, compliance, privacy, and legal pages. Returns structured JSON with categorized URLs (vendor_name, vendor_domain, discovered_urls, notes). Input: the vendor's main website URL.",
			tools:       readOnlyBrowserTools,
			build:       buildCrawlerAgent,
		},
		{
			toolName:    "assess_security",
			description: "Perform technical security checks on a domain. Returns structured JSON with per-check results (ssl, headers, dmarc, spf, breaches, dnssec, csp, cors, dns, whois) each with status (pass/warning/fail/error) and details. Input: the vendor's domain name (e.g. example.com).",
			tools:       securityTools,
			build:       buildSecurityAgent,
		},
		{
			toolName:    "analyze_document",
			description: "Analyze a specific document page (privacy policy, DPA, ToS) and extract key provisions. Returns structured JSON with document_type, retention, locations, GDPR/CCPA indicators, clauses, and summary. Input: the document URL.",
			tools:       readOnlyBrowserTools,
			build:       buildAnalyzerAgent,
		},
		{
			toolName:    "assess_compliance",
			description: "Identify certifications and compliance frameworks from a trust/compliance page. Returns structured JSON with certifications (name, status, details), audit reports, and frameworks. Input: the trust or compliance page URL.",
			tools:       readOnlyBrowserTools,
			build:       buildComplianceAgent,
		},
		{
			toolName:    "assess_market_presence",
			description: "Analyze a vendor's market presence. Returns structured JSON with notable_customers, case_studies, partnerships, company_size_signals, funding_info, and market_position. Input: the vendor's main website URL.",
			tools:       readOnlyBrowserTools,
			build:       buildMarketAgent,
		},
		{
			toolName:    "extract_subprocessors",
			description: "Find and extract the list of sub-processors from a vendor's website. Returns structured JSON with subprocessors (name, country, purpose), total_count, and source. Input: the vendor's main website URL or a known subprocessors page URL.",
			tools:       subprocessorTools,
			build:       buildSubprocessorAgent,
		},
		{
			toolName:    "assess_data_processing",
			description: "Assess data processing practices. Returns structured JSON with encryption, retention, deletion, data locations, transfer mechanisms, DPA status, DSAR handling, and rating. Input: a relevant page URL (privacy policy, DPA, security page, or trust center).",
			tools:       readOnlyBrowserTools,
			build:       buildDataProcessingAgent,
		},
		{
			toolName:    "assess_incident_response",
			description: "Evaluate incident response capabilities. Returns structured JSON with ir_plan, notification_timeline, status_page, post_mortems, recent_incidents, security_contact, and rating. Input: a relevant page URL (security page, trust center, or status page).",
			tools:       readOnlyBrowserTools,
			build:       buildIncidentResponseAgent,
		},
		{
			toolName:    "assess_business_continuity",
			description: "Evaluate business continuity and disaster recovery. Returns structured JSON with dr_plan, rto, rpo, cloud_providers, uptime_sla, regions, backup_strategy, and rating. Input: a relevant page URL (SLA page, trust center, or infrastructure docs).",
			tools:       readOnlyBrowserTools,
			build:       buildBusinessContinuityAgent,
		},
		{
			toolName:    "assess_professional_standing",
			description: "Evaluate professional standing for services firms. Returns structured JSON with licensing, memberships, insurance, team_credentials, coi_policy, and rating. Input: relevant page URL (team page, about page, credentials page).",
			tools:       readOnlyBrowserTools,
			build:       buildProfessionalStandingAgent,
		},
		{
			toolName:    "assess_ai_risk",
			description: "Evaluate AI governance (ISO 42001). Returns structured JSON with ai_involvement, use_cases, model_transparency, bias_controls, customer_data_training, human_oversight, and rating. Input: relevant page URL (AI policy, trust center, responsible AI page, or main website).",
			tools:       readOnlyBrowserTools,
			build:       buildAIRiskAgent,
		},
		{
			toolName:    "assess_regulatory_compliance",
			description: "Deep regulatory compliance check. Returns structured JSON with per-framework assessment (gdpr, hipaa, pci_dss, sox) each with articles, status, and notes. Input: relevant page URL (DPA, compliance page, trust center).",
			tools:       readOnlyBrowserTools,
			build:       buildRegulatoryComplianceAgent,
		},
	}

	// Optional sub-agents: only added when a search endpoint is configured.
	if searchEndpoint != "" {
		researchBrowserTools := browser.NewInteractiveToolset(researchBrowser).Tools()

		searchTool := search.WebSearchTool(searchEndpoint)
		govDBTool := search.CheckGovernmentDBTool(searchEndpoint)
		waybackTool := search.CheckWaybackTool()
		diffTool := search.DiffDocumentsTool()

		// withResearchTools returns a fresh slice combining the supplied
		// extra tools with the research browser tools. The fresh
		// allocation is required so the four sub-agent tool slices do
		// not share a backing array.
		withResearchTools := func(extra ...agent.Tool) []agent.Tool {
			out := make([]agent.Tool, 0, len(extra)+len(researchBrowserTools))
			out = append(out, extra...)
			out = append(out, researchBrowserTools...)
			return out
		}

		websearchTools := withResearchTools(searchTool)
		financialTools := withResearchTools(searchTool, govDBTool, waybackTool)
		codeSecurityTools := withResearchTools(searchTool)
		comparisonTools := withResearchTools(searchTool, diffTool)

		entries = append(entries,
			subAgentEntry{
				toolName:    "research_vendor_externally",
				description: "Search the open web for external signals about the vendor. Returns structured JSON with security_incidents, regulatory_actions, customer_sentiment, recent_news, red_flags, and positive_signals. Input: the vendor's name and domain.",
				tools:       websearchTools,
				build:       buildWebsearchAgent,
			},
			subAgentEntry{
				toolName:    "assess_financial_stability",
				description: "Evaluate vendor financial stability. Returns structured JSON with company_age, funding, employee_count, legal_standing, ownership, risk_signals, overall_assessment, and confidence. Input: vendor name and website URL.",
				tools:       financialTools,
				build:       buildFinancialStabilityAgent,
			},
			subAgentEntry{
				toolName:    "assess_code_security",
				description: "Evaluate open-source code security posture. Returns structured JSON with has_public_repos, security_advisories, dependency_management, release_cadence, security_policy, overall_assessment, and risk_signals. Input: vendor name and website URL.",
				tools:       codeSecurityTools,
				build:       buildCodeSecurityAgent,
			},
			subAgentEntry{
				toolName:    "compare_vendor",
				description: "Find and compare alternative vendors. Returns structured JSON with alternatives (name, certifications, security_score), comparison_summary, vendor_strengths, vendor_weaknesses, and overall_position. Input: vendor name, category, and website URL.",
				tools:       comparisonTools,
				build:       buildVendorComparisonAgent,
			},
		)
	}

	tools := make([]agent.Tool, 0, len(entries))
	for _, e := range entries {
		ag, err := e.build(client, model, e.tools, subAgentOpts(e.toolName)...)
		if err != nil {
			return nil, fmt.Errorf("cannot create %s sub-agent: %w", e.toolName, err)
		}
		tools = append(tools, ag.AsTool(e.toolName, e.description))
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
		agent.WithMaxTurns(orchestratorMaxTurns),
		agent.WithParallelToolCalls(true),
		agent.WithThinking(orchestratorThinkingBudget),
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
