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

	subAgentOpts := func(step string) []agent.Option {
		opts := []agent.Option{loggerOpt}
		if reporter != nil {
			opts = append(opts, agent.WithHooks(newSubProgressHooks(reporter, step)))
		}
		return opts
	}

	crawler := newCrawlerAgent(client, model, readOnlyBrowserTools, subAgentOpts("crawl_vendor_website")...)
	analyzer := newDocumentAnalyzerAgent(client, model, readOnlyBrowserTools, subAgentOpts("analyze_document")...)
	securityAssessor := newSecurityAssessorAgent(client, model, securityTools, subAgentOpts("assess_security")...)
	compliance := newComplianceAssessorAgent(client, model, readOnlyBrowserTools, subAgentOpts("assess_compliance")...)
	market := newMarketPresenceAgent(client, model, readOnlyBrowserTools, subAgentOpts("assess_market_presence")...)
	subprocessorTools := unrestrictedBrowserTools
	if searchEndpoint != "" {
		searchTool, err := search.WebSearchTool(searchEndpoint)
		if err != nil {
			return nil, fmt.Errorf("cannot build subprocessor search tool: %w", err)
		}
		subprocessorTools = append(subprocessorTools, searchTool)
	}
	subprocessor := newSubprocessorAgent(client, model, subprocessorTools, subAgentOpts("extract_subprocessors")...)
	dataProcessing := newDataProcessingAgent(client, model, readOnlyBrowserTools, subAgentOpts("assess_data_processing")...)
	aiRisk := newAIRiskAgent(client, model, readOnlyBrowserTools, subAgentOpts("assess_ai_risk")...)
	incidentResponse := newIncidentResponseAgent(client, model, readOnlyBrowserTools, subAgentOpts("assess_incident_response")...)
	businessContinuity := newBusinessContinuityAgent(client, model, readOnlyBrowserTools, subAgentOpts("assess_business_continuity")...)
	professionalStanding := newProfessionalStandingAgent(client, model, readOnlyBrowserTools, subAgentOpts("assess_professional_standing")...)
	regulatoryCompliance := newRegulatoryComplianceAgent(client, model, readOnlyBrowserTools, subAgentOpts("assess_regulatory_compliance")...)

	tools := []agent.Tool{
		crawler.AsTool(
			"crawl_vendor_website",
			"Crawl a vendor website to discover security, compliance, privacy, and legal pages. Returns structured JSON with categorized URLs. Input: the vendor's main website URL.",
		),
		securityAssessor.AsTool(
			"assess_security",
			"Perform technical security checks on a domain (SSL, headers, DMARC, breaches, DNSSEC). Input: the vendor's domain name (e.g. example.com).",
		),
		analyzer.AsTool(
			"analyze_document",
			"Analyze a specific document page (privacy policy, DPA, ToS) and extract key provisions. Input: the document URL.",
		),
		compliance.AsTool(
			"assess_compliance",
			"Identify certifications and compliance frameworks from a trust/compliance page. Input: the trust or compliance page URL.",
		),
		market.AsTool(
			"assess_market_presence",
			"Analyze a vendor's market presence by identifying notable customers, case studies, and company size signals. Input: the vendor's main website URL.",
		),
		subprocessor.AsTool(
			"extract_subprocessors",
			"Find and extract the list of sub-processors from a vendor's website. Input: the vendor's main website URL or a known subprocessors page URL.",
		),
		dataProcessing.AsTool(
			"assess_data_processing",
			"Assess data processing practices including encryption, retention, cross-border transfers, and backup procedures. Input: a relevant page URL (privacy policy, DPA, security page, or trust center).",
		),
		incidentResponse.AsTool(
			"assess_incident_response",
			"Evaluate incident response capabilities, breach notification procedures, and incident history. Input: a relevant page URL (security page, trust center, or status page).",
		),
		businessContinuity.AsTool(
			"assess_business_continuity",
			"Evaluate business continuity and disaster recovery capabilities including SLA, uptime, and infrastructure redundancy. Input: a relevant page URL (SLA page, trust center, or infrastructure docs).",
		),
		professionalStanding.AsTool(
			"assess_professional_standing",
			"Evaluate professional standing for services firms: licensing, credentials, insurance, industry memberships. Input: relevant page URL (team page, about page, credentials page).",
		),
		aiRisk.AsTool(
			"assess_ai_risk",
			"Evaluate AI governance, model transparency, bias controls, human oversight, and training data governance (ISO 42001). Input: relevant page URL (AI policy, trust center, responsible AI page, or main website).",
		),
		regulatoryCompliance.AsTool(
			"assess_regulatory_compliance",
			"Deep regulatory compliance check against specific frameworks (GDPR articles, HIPAA, PCI DSS, SOX). Downloads and analyzes PDFs. Input: relevant page URL (DPA, compliance page, trust center).",
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
		websearch := newWebSearchAgent(client, model, websearchTools, subAgentOpts("research_vendor_externally")...)

		tools = append(tools, websearch.AsTool(
			"research_vendor_externally",
			"Search the open web for external signals about the vendor: news, breaches, reviews, regulatory actions. Input: the vendor's name and domain.",
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
		financialStability := newFinancialStabilityAgent(client, model, financialTools, subAgentOpts("assess_financial_stability")...)

		codeSecurityTools := append([]agent.Tool{searchTool}, researchBrowserTools...)
		codeSecurity := newCodeSecurityAgent(client, model, codeSecurityTools, subAgentOpts("assess_code_security")...)

		comparisonTools := append([]agent.Tool{searchTool, diffTool}, researchBrowserTools...)
		vendorComparison := newVendorComparisonAgent(client, model, comparisonTools, subAgentOpts("compare_vendor")...)

		tools = append(
			tools,
			financialStability.AsTool(
				"assess_financial_stability",
				"Evaluate vendor financial stability: funding, company age, employee count, SEC filings, bankruptcy signals, ownership changes. Input: vendor name and website URL.",
			),
			codeSecurity.AsTool(
				"assess_code_security",
				"Evaluate open-source code security posture: GitHub advisories, CVEs, dependency management, release cadence, security policy. Input: vendor name and website URL.",
			),
			vendorComparison.AsTool(
				"compare_vendor",
				"Find and compare alternative vendors in the same category on security, compliance, and market presence. Input: vendor name, category, and website URL.",
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
		agent.WithTools(tools...),
		agent.WithMaxTurns(35),
		agent.WithParallelToolCalls(true),
		agent.WithThinking(10000),
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
