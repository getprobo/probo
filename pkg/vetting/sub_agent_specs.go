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

import _ "embed"

// Specs for every vetting sub-agent. The orchestrator passes each spec
// to newSubAgent[T] together with the structured output type and the
// tool set the agent should use.
//
// Tuning notes:
//   - thinkingBudget=4000 is enabled on agents that need to reason over
//     multiple documents (analyzer, ai_risk, data_processing, business
//     continuity, incident response, regulatory compliance). The agent
//     runtime delays structured output enforcement until a dedicated
//     synthesis turn (run.go), so thinking no longer conflicts with the
//     JSON schema during tool exploration.
//   - parallelTools=true is enabled on agents that issue many independent
//     tool calls per turn (security_assessor, market, code_security,
//     financial_stability, web_search, regulatory_compliance).
//   - maxTurns is sized to give the agent enough room for tool calls plus
//     a few retries; subprocessor extraction needs the most because of
//     paginated subprocessor lists.

var (
	//go:embed prompts/crawler.txt
	crawlerPrompt string

	//go:embed prompts/analyzer.txt
	analyzerPrompt string

	//go:embed prompts/security.txt
	securityPrompt string

	//go:embed prompts/compliance.txt
	compliancePrompt string

	//go:embed prompts/market.txt
	marketPrompt string

	//go:embed prompts/subprocessor.txt
	subprocessorPrompt string

	//go:embed prompts/data_processing.txt
	dataProcessingPrompt string

	//go:embed prompts/ai_risk.txt
	aiRiskPrompt string

	//go:embed prompts/incident_response.txt
	incidentResponsePrompt string

	//go:embed prompts/business_continuity.txt
	businessContinuityPrompt string

	//go:embed prompts/professional_standing.txt
	professionalStandingPrompt string

	//go:embed prompts/regulatory_compliance.txt
	regulatoryCompliancePrompt string

	//go:embed prompts/websearch.txt
	websearchPrompt string

	//go:embed prompts/financial_stability.txt
	financialStabilityPrompt string

	//go:embed prompts/code_security.txt
	codeSecurityPrompt string

	//go:embed prompts/vendor_comparison.txt
	vendorComparisonPrompt string
)

var (
	crawlerAgentSpec = subAgentSpec{
		name:       "website_crawler",
		outputName: "crawler_output",
		prompt:     crawlerPrompt,
		maxTurns:   40,
	}

	analyzerAgentSpec = subAgentSpec{
		name:           "document_analyzer",
		outputName:     "document_analysis_output",
		prompt:         analyzerPrompt,
		maxTurns:       20,
		thinkingBudget: 4000,
	}

	securityAgentSpec = subAgentSpec{
		name:          "security_assessor",
		outputName:    "security_output",
		prompt:        securityPrompt,
		maxTurns:      32,
		parallelTools: true,
	}

	complianceAgentSpec = subAgentSpec{
		name:       "compliance_assessor",
		outputName: "compliance_output",
		prompt:     compliancePrompt,
		maxTurns:   20,
	}

	marketAgentSpec = subAgentSpec{
		name:          "market_presence_analyst",
		outputName:    "market_output",
		prompt:        marketPrompt,
		maxTurns:      40,
		parallelTools: true,
	}

	subprocessorAgentSpec = subAgentSpec{
		name:       "subprocessor_extractor",
		outputName: "subprocessor_output",
		prompt:     subprocessorPrompt,
		maxTurns:   100,
	}

	dataProcessingAgentSpec = subAgentSpec{
		name:           "data_processing_assessor",
		outputName:     "data_processing_output",
		prompt:         dataProcessingPrompt,
		maxTurns:       28,
		thinkingBudget: 4000,
	}

	aiRiskAgentSpec = subAgentSpec{
		name:           "ai_risk_assessor",
		outputName:     "ai_risk_output",
		prompt:         aiRiskPrompt,
		maxTurns:       28,
		thinkingBudget: 4000,
	}

	incidentResponseAgentSpec = subAgentSpec{
		name:           "incident_response_assessor",
		outputName:     "incident_response_output",
		prompt:         incidentResponsePrompt,
		maxTurns:       28,
		thinkingBudget: 4000,
	}

	businessContinuityAgentSpec = subAgentSpec{
		name:           "business_continuity_assessor",
		outputName:     "business_continuity_output",
		prompt:         businessContinuityPrompt,
		maxTurns:       28,
		thinkingBudget: 4000,
	}

	professionalStandingAgentSpec = subAgentSpec{
		name:       "professional_standing_assessor",
		outputName: "professional_standing_output",
		prompt:     professionalStandingPrompt,
		maxTurns:   28,
	}

	regulatoryComplianceAgentSpec = subAgentSpec{
		name:           "regulatory_compliance_assessor",
		outputName:     "regulatory_compliance_output",
		prompt:         regulatoryCompliancePrompt,
		maxTurns:       40,
		thinkingBudget: 4000,
		parallelTools:  true,
	}

	websearchAgentSpec = subAgentSpec{
		name:          "web_search_analyst",
		outputName:    "web_search_output",
		prompt:        websearchPrompt,
		maxTurns:      40,
		parallelTools: true,
	}

	financialStabilityAgentSpec = subAgentSpec{
		name:          "financial_stability_assessor",
		outputName:    "financial_stability_output",
		prompt:        financialStabilityPrompt,
		maxTurns:      40,
		parallelTools: true,
	}

	codeSecurityAgentSpec = subAgentSpec{
		name:          "code_security_assessor",
		outputName:    "code_security_output",
		prompt:        codeSecurityPrompt,
		maxTurns:      40,
		parallelTools: true,
	}

	vendorComparisonAgentSpec = subAgentSpec{
		name:       "vendor_comparison_assessor",
		outputName: "vendor_comparison_output",
		prompt:     vendorComparisonPrompt,
		maxTurns:   40,
	}
)

// Per-output-type builders. Defining them here lets the orchestrator hold
// a slice of (toolName, description, tools, builder) entries instead of
// embedding a closure with an explicit type parameter at every call site.
var (
	buildCrawlerAgent              = buildFor[CrawlerOutput](crawlerAgentSpec)
	buildAnalyzerAgent             = buildFor[DocumentAnalysisOutput](analyzerAgentSpec)
	buildSecurityAgent             = buildFor[SecurityOutput](securityAgentSpec)
	buildComplianceAgent           = buildFor[ComplianceOutput](complianceAgentSpec)
	buildMarketAgent               = buildFor[MarketOutput](marketAgentSpec)
	buildSubprocessorAgent         = buildFor[SubprocessorOutput](subprocessorAgentSpec)
	buildDataProcessingAgent       = buildFor[DataProcessingOutput](dataProcessingAgentSpec)
	buildAIRiskAgent               = buildFor[AIRiskOutput](aiRiskAgentSpec)
	buildIncidentResponseAgent     = buildFor[IncidentResponseOutput](incidentResponseAgentSpec)
	buildBusinessContinuityAgent   = buildFor[BusinessContinuityOutput](businessContinuityAgentSpec)
	buildProfessionalStandingAgent = buildFor[ProfessionalStandingOutput](professionalStandingAgentSpec)
	buildRegulatoryComplianceAgent = buildFor[RegulatoryComplianceOutput](regulatoryComplianceAgentSpec)
	buildWebsearchAgent            = buildFor[WebSearchOutput](websearchAgentSpec)
	buildFinancialStabilityAgent   = buildFor[FinancialStabilityOutput](financialStabilityAgentSpec)
	buildCodeSecurityAgent         = buildFor[CodeSecurityOutput](codeSecurityAgentSpec)
	buildVendorComparisonAgent     = buildFor[VendorComparisonOutput](vendorComparisonAgentSpec)
)
