// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package thirdparty

import (
	_ "embed"
	"fmt"
	"strings"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/agent/tools/search"
)

//go:embed prompts/common_third_party_compliance_docs.txt.tmpl
var complianceDocsPrompt string

// ComplianceDocsResult is the structured output of the compliance-docs
// agent (Agent B): the legal-document URLs, trust/security/status pages,
// and certifications. These all live in the same source ecosystem (the
// vendor footer and trust portal), so one agent resolves them together.
type ComplianceDocsResult struct {
	PrivacyPolicyURL              EnrichedField       `json:"privacy_policy_url" jsonschema:"URL of the vendor's privacy policy."`
	TermsOfServiceURL             EnrichedField       `json:"terms_of_service_url" jsonschema:"URL of the vendor's terms of service / terms of use."`
	ServiceLevelAgreementURL      EnrichedField       `json:"service_level_agreement_url" jsonschema:"URL of the vendor's public service level agreement (SLA). Often gated behind sales; return empty when not public."`
	ServiceSoftwareAgreementURL   EnrichedField       `json:"service_software_agreement_url" jsonschema:"URL of the vendor's master software/subscription agreement (MSA). Often gated or identical to the terms of service; return empty when not public."`
	DataProcessingAgreementURL    EnrichedField       `json:"data_processing_agreement_url" jsonschema:"URL of the vendor's data processing agreement (DPA). Often a PDF; return empty when only available on request."`
	BusinessAssociateAgreementURL EnrichedField       `json:"business_associate_agreement_url" jsonschema:"URL of the vendor's HIPAA business associate agreement (BAA). Almost always gated behind sales; return empty when not public."`
	SubprocessorsListURL          EnrichedField       `json:"subprocessors_list_url" jsonschema:"URL of the vendor's sub-processors list page."`
	StatusPageURL                 EnrichedField       `json:"status_page_url" jsonschema:"URL of the vendor's uptime/status page (e.g. status.vendor.com)."`
	SecurityPageURL               EnrichedField       `json:"security_page_url" jsonschema:"URL of the vendor's security page or security overview."`
	TrustPageURL                  EnrichedField       `json:"trust_page_url" jsonschema:"URL of the vendor's trust center / trust portal (e.g. Vanta, SafeBase, Drata hosted)."`
	Certifications                CertificationsField `json:"certifications" jsonschema:"Certifications and compliance frameworks the vendor publicly claims."`
}

// buildComplianceDocsAgent builds Agent B. extraTools carries the browser
// read-only toolset when a headless Chrome endpoint is configured; it is
// empty otherwise, in which case the agent relies on web_search alone.
func buildComplianceDocsAgent(
	cfg EnrichmentConfig,
	logger *log.Logger,
	extraTools []agent.Tool,
) *agent.Agent {
	tools := append([]agent.Tool{}, extraTools...)

	if cfg.FirecrawlAPIKey != "" {
		tools = append(tools, search.FirecrawlSearchTool(cfg.FirecrawlAPIKey))
	}

	outputType, err := agent.NewOutputType[ComplianceDocsResult]("common_third_party_compliance_docs")
	if err != nil {
		panic(fmt.Sprintf("thirdparty: cannot build compliance docs output type: %s", err))
	}

	opts := []agent.Option{
		agent.WithInstructions(complianceDocsPrompt),
		agent.WithModel(cfg.Model),
		agent.WithOutputType(outputType),
		agent.WithMaxTurns(resolveEnrichmentMaxTurns(cfg.MaxTurns)),
		agent.WithMaxTokens(resolveEnrichmentMaxTokens(cfg.MaxTokens)),
		agent.WithLogger(logger),
	}

	if len(tools) > 0 {
		opts = append(opts, agent.WithTools(tools...))
	}

	if cfg.Temperature != nil {
		opts = append(opts, agent.WithTemperature(*cfg.Temperature))
	}

	return agent.New("common-third-party-compliance-docs", cfg.LLMClient, opts...)
}

// buildComplianceDocsPrompt renders the per-row input for Agent B,
// seeding it with the vendor name and the website/legal name resolved by
// Agent A so it can scope its search to the vendor's own domain.
func buildComplianceDocsPrompt(name, websiteURL, legalName string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Find the compliance documents and trust pages for this vendor.\n\n")
	fmt.Fprintf(&b, "<name> %s </name>\n", name)

	if w := strings.TrimSpace(websiteURL); w != "" {
		fmt.Fprintf(&b, "<website> %s </website>\n", w)
	}

	if l := strings.TrimSpace(legalName); l != "" {
		fmt.Fprintf(&b, "<legal_name> %s </legal_name>\n", l)
	}

	return b.String()
}
