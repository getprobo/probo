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

package vendor_assessment

import (
	_ "embed"
	"fmt"

	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/agent/tools/browser"
	"go.probo.inc/probo/pkg/agent/tools/security"
	"go.probo.inc/probo/pkg/llm"
)

//go:embed orchestrator_prompt.txt
var orchestratorSystemPrompt string

func newOrchestratorAgent(
	client *llm.Client,
	model string,
	b *browser.Browser,
	reporter agent.ProgressReporter,
) (*agent.Agent, error) {
	browserTools, err := browser.BuildTools(b)
	if err != nil {
		return nil, fmt.Errorf("cannot build browser tools: %w", err)
	}

	securityTools, err := security.BuildTools()
	if err != nil {
		return nil, fmt.Errorf("cannot build security tools: %w", err)
	}

	var crawlerOpts, analyzerOpts, securityOpts, complianceOpts []agent.Option
	if reporter != nil {
		crawlerOpts = append(
			crawlerOpts,
			agent.WithHooks(newSubProgressHooks(reporter, "crawl_vendor_website")),
		)
		analyzerOpts = append(
			analyzerOpts,
			agent.WithHooks(newSubProgressHooks(reporter, "analyze_document")),
		)
		securityOpts = append(
			securityOpts,
			agent.WithHooks(newSubProgressHooks(reporter, "assess_security")),
		)
		complianceOpts = append(
			complianceOpts,
			agent.WithHooks(newSubProgressHooks(reporter, "assess_compliance")),
		)
	}

	crawler := newCrawlerAgent(client, model, browserTools, crawlerOpts...)
	analyzer := newDocumentAnalyzerAgent(client, model, browserTools, analyzerOpts...)
	securityAssessor := newSecurityAssessorAgent(client, model, securityTools, securityOpts...)
	compliance := newComplianceAssessorAgent(client, model, browserTools, complianceOpts...)

	opts := []agent.Option{
		agent.WithInstructions(orchestratorSystemPrompt),
		agent.WithModel(model),
		agent.WithTools(
			crawler.AsTool(
				"crawl_vendor_website",
				"Crawl a vendor website to discover security, compliance, privacy, and legal pages. Input: the vendor's main website URL.",
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
		),
		agent.WithMaxTurns(20),
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
