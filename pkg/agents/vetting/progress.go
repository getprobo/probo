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
	"math/rand/v2"

	"go.probo.inc/probo/pkg/agent"
)

var (
	toolMessages = map[string][]string{
		// Orchestrator tools (top-level steps).
		"crawl_vendor_website": {
			"Exploring vendor website for security and compliance pages",
			"Discovering key pages on the vendor website",
			"Mapping out the vendor's online presence",
			"Scanning the website structure for relevant sections",
			"Browsing the vendor site to locate important resources",
		},
		"assess_security": {
			"Running technical security checks on the domain",
			"Evaluating the vendor's security posture",
			"Performing infrastructure security analysis",
			"Auditing the domain's technical defenses",
			"Probing the vendor's security configuration",
		},
		"analyze_document": {
			"Reviewing document for key provisions",
			"Analyzing policy details and obligations",
			"Extracting important clauses from the document",
			"Parsing the document for notable terms",
			"Breaking down the document's main points",
		},
		"assess_compliance": {
			"Identifying certifications and compliance frameworks",
			"Reviewing the vendor's compliance posture",
			"Checking for recognized security certifications",
			"Surveying the vendor's regulatory standing",
			"Evaluating adherence to industry standards",
		},
		"assess_market_presence": {
			"Investigating the vendor's market presence",
			"Looking for notable customers and case studies",
			"Checking who uses this vendor",
			"Assessing the vendor's market credibility",
			"Identifying the vendor's customer base",
		},
		"extract_subprocessors": {
			"Extracting sub-processor information",
			"Reading the vendor's sub-processor list",
			"Identifying third-party sub-processors",
			"Parsing sub-processor details",
			"Cataloging the vendor's sub-processors",
		},
		"assess_data_processing": {
			"Analyzing data processing practices",
			"Reviewing encryption and data handling",
			"Evaluating data retention and transfer policies",
			"Checking data processing documentation",
			"Assessing cross-border data transfer mechanisms",
		},
		"assess_incident_response": {
			"Evaluating incident response capabilities",
			"Reviewing breach notification procedures",
			"Checking incident history and transparency",
			"Assessing security incident readiness",
			"Examining post-incident review processes",
		},
		"assess_business_continuity": {
			"Assessing business continuity planning",
			"Reviewing disaster recovery capabilities",
			"Checking SLA and uptime commitments",
			"Evaluating infrastructure redundancy",
			"Examining geographic distribution and failover",
		},
		"assess_professional_standing": {
			"Evaluating professional standing and credentials",
			"Reviewing licensing and industry memberships",
			"Checking professional qualifications and accreditation",
			"Assessing team credentials and experience",
			"Examining professional liability and insurance coverage",
		},
		"assess_ai_risk": {
			"Evaluating AI governance and responsible AI practices",
			"Reviewing AI transparency and bias controls",
			"Checking AI risk management documentation",
			"Assessing automated decision-making safeguards",
			"Examining AI training data governance",
		},
		"research_vendor_externally": {
			"Researching the vendor across the web",
			"Searching for external signals about the vendor",
			"Looking for news and breach reports",
			"Investigating the vendor's external reputation",
			"Scanning public sources for vendor intelligence",
		},
		"assess_regulatory_compliance": {
			"Performing deep regulatory compliance analysis",
			"Checking GDPR article-level compliance",
			"Analyzing regulatory framework adherence",
			"Reviewing compliance against specific regulations",
			"Evaluating regulatory requirements coverage",
		},
		"assess_financial_stability": {
			"Assessing vendor financial stability",
			"Investigating company funding and financial health",
			"Checking business registration and SEC filings",
			"Evaluating vendor viability and longevity",
			"Researching company financial standing",
		},
		"assess_code_security": {
			"Evaluating open-source code security posture",
			"Checking for security advisories and CVEs",
			"Reviewing dependency management practices",
			"Analyzing release cadence and maintenance",
			"Inspecting code security practices",
		},
		"compare_vendor": {
			"Comparing vendor against alternatives",
			"Finding competing vendors in the same category",
			"Benchmarking security and compliance posture",
			"Evaluating vendor relative to market alternatives",
			"Assessing competitive landscape",
		},
		"extract_vendor_info": {
			"Extracting vendor information from assessment",
			"Parsing assessment into structured data",
			"Building vendor profile from findings",
			"Distilling key vendor details from report",
			"Organizing vendor metadata from assessment",
		},

		// Web search sub-agent tools.
		"web_search": {
			"Searching the web",
			"Running a web search query",
			"Looking up information online",
			"Querying search results",
			"Fetching search results",
		},

		// Security sub-agent tools.
		"check_ssl_certificate": {
			"Inspecting SSL/TLS certificate",
			"Verifying certificate validity and configuration",
			"Checking SSL certificate details",
			"Reviewing the certificate chain",
			"Examining TLS setup and expiration",
		},
		"check_security_headers": {
			"Analyzing HTTP security headers",
			"Reviewing response headers for security best practices",
			"Checking for missing security headers",
			"Scanning HTTP headers for protective directives",
			"Evaluating header-based security controls",
		},
		"check_dmarc": {
			"Looking up DMARC email authentication record",
			"Checking DMARC policy configuration",
			"Verifying email spoofing protections",
			"Querying DNS for DMARC policy",
			"Reviewing email authentication settings",
		},
		"check_spf": {
			"Looking up SPF email authentication record",
			"Checking SPF policy configuration",
			"Verifying sender policy framework",
			"Querying DNS for SPF record",
			"Reviewing SPF authorization settings",
		},
		"check_breaches": {
			"Searching for known data breaches",
			"Checking breach databases for past incidents",
			"Looking up the domain in breach records",
			"Scanning public breach disclosures",
			"Querying breach intelligence sources",
		},
		"check_dnssec": {
			"Verifying DNSSEC configuration",
			"Checking DNS security extensions",
			"Inspecting DNSSEC chain of trust",
			"Validating DNS signing status",
			"Reviewing DNSSEC deployment",
		},
		"analyze_csp": {
			"Evaluating Content Security Policy",
			"Analyzing CSP directives for weaknesses",
			"Reviewing content security rules",
			"Checking CSP for unsafe directives",
			"Parsing Content Security Policy header",
		},
		"check_cors": {
			"Checking CORS configuration",
			"Inspecting cross-origin resource sharing policy",
			"Reviewing CORS headers",
			"Evaluating cross-origin access rules",
			"Analyzing CORS allow-origin settings",
		},

		// Browser tools used by crawler, analyzer, and compliance sub-agents.
		"navigate_to_url": {
			"Opening page",
			"Loading page content",
			"Navigating to the page",
			"Visiting the page",
			"Heading to the page",
		},
		"extract_page_text": {
			"Reading page content",
			"Extracting text from the page",
			"Pulling content from the page",
			"Scanning page text",
			"Capturing the page body",
		},
		"extract_links": {
			"Collecting links from the page",
			"Gathering all page links",
			"Discovering outgoing links",
			"Harvesting links on the page",
			"Listing page hyperlinks",
		},
		"find_links_matching": {
			"Searching for relevant links",
			"Looking for links matching the pattern",
			"Filtering page links by keyword",
			"Hunting for specific links on the page",
			"Sifting through links for a match",
		},
		"click_element": {
			"Clicking on the page",
			"Interacting with the page",
			"Pressing a button on the page",
			"Navigating within the page",
			"Triggering a page action",
		},
		"select_option": {
			"Selecting an option on the page",
			"Changing a dropdown selection",
			"Adjusting page settings",
			"Picking a value from a dropdown",
			"Updating a page filter",
		},

		// New security tools.
		"check_whois": {
			"Looking up domain registration details",
			"Checking WHOIS records",
			"Querying domain registrar information",
			"Inspecting domain ownership data",
			"Retrieving domain age and registrant info",
		},
		"check_dns_records": {
			"Querying DNS records",
			"Looking up A, MX, and NS records",
			"Checking DNS configuration",
			"Resolving domain DNS entries",
			"Inspecting hosting and email providers",
		},

		// New browser tools.
		"fetch_robots_txt": {
			"Fetching robots.txt",
			"Checking robots.txt for hidden pages",
			"Reading site crawl directives",
			"Discovering sitemap URLs from robots.txt",
			"Parsing robots.txt disallow rules",
		},
		"fetch_sitemap": {
			"Fetching sitemap",
			"Parsing sitemap for page URLs",
			"Discovering pages from sitemap",
			"Reading sitemap index",
			"Extracting URLs from sitemap XML",
		},
		"download_pdf": {
			"Downloading and extracting PDF",
			"Reading PDF document content",
			"Extracting text from PDF",
			"Processing PDF document",
			"Parsing PDF for analysis",
		},

		// New search tools.
		"check_wayback": {
			"Checking Wayback Machine archives",
			"Looking for historical page snapshots",
			"Querying Internet Archive",
			"Searching for archived versions",
			"Checking page history in Wayback Machine",
		},
		"check_government_databases": {
			"Searching government regulatory databases",
			"Checking SEC and FTC records",
			"Looking for GDPR enforcement actions",
			"Querying regulatory databases",
			"Searching for enforcement history",
		},
		"diff_documents": {
			"Comparing document versions",
			"Diffing document texts",
			"Analyzing document changes",
			"Checking for document modifications",
			"Computing document differences",
		},
	}
)

func randomMessage(step string) string {
	msgs, ok := toolMessages[step]
	if !ok {
		return ""
	}

	return msgs[rand.IntN(len(msgs))]
}

// reportProgress emits a progress event to the reporter if non-nil.
func reportProgress(
	ctx context.Context,
	reporter agent.ProgressReporter,
	step string,
	eventType agent.ProgressEventType,
) {
	if reporter == nil {
		return
	}

	event := agent.ProgressEvent{
		Type: eventType,
		Step: step,
	}

	if eventType == agent.ProgressEventStepStarted {
		event.Message = randomMessage(step)
	}

	reporter(ctx, event)
}

// progressHooks translates orchestrator-level tool events into progress events.
type progressHooks struct {
	agent.NoOpHooks
	reporter agent.ProgressReporter
}

func newProgressHooks(reporter agent.ProgressReporter) *progressHooks {
	return &progressHooks{reporter: reporter}
}

func (h *progressHooks) OnToolStart(ctx context.Context, _ *agent.Agent, tool agent.Tool, _ string) {
	msg := randomMessage(tool.Name())
	if msg == "" {
		return
	}

	h.reporter(
		ctx,
		agent.ProgressEvent{
			Type:    agent.ProgressEventStepStarted,
			Step:    tool.Name(),
			Message: msg,
		},
	)
}

func (h *progressHooks) OnToolEnd(ctx context.Context, _ *agent.Agent, tool agent.Tool, _ agent.ToolResult, err error) {
	if _, ok := toolMessages[tool.Name()]; !ok {
		return
	}

	eventType := agent.ProgressEventStepCompleted
	if err != nil {
		eventType = agent.ProgressEventStepFailed
	}

	h.reporter(
		ctx,
		agent.ProgressEvent{
			Type: eventType,
			Step: tool.Name(),
		},
	)
}

// subProgressHooks translates sub-agent tool events into progress events
// scoped under a parent step.
type subProgressHooks struct {
	agent.NoOpHooks
	reporter   agent.ProgressReporter
	parentStep string
}

func newSubProgressHooks(reporter agent.ProgressReporter, parentStep string) *subProgressHooks {
	return &subProgressHooks{
		reporter:   reporter,
		parentStep: parentStep,
	}
}

func (h *subProgressHooks) OnToolStart(ctx context.Context, _ *agent.Agent, tool agent.Tool, _ string) {
	msg := randomMessage(tool.Name())
	if msg == "" {
		return
	}

	h.reporter(
		ctx,
		agent.ProgressEvent{
			Type:       agent.ProgressEventStepStarted,
			Step:       tool.Name(),
			ParentStep: h.parentStep,
			Message:    msg,
		},
	)
}

func (h *subProgressHooks) OnToolEnd(ctx context.Context, _ *agent.Agent, tool agent.Tool, _ agent.ToolResult, err error) {
	if _, ok := toolMessages[tool.Name()]; !ok {
		return
	}

	eventType := agent.ProgressEventStepCompleted
	if err != nil {
		eventType = agent.ProgressEventStepFailed
	}

	h.reporter(
		ctx,
		agent.ProgressEvent{
			Type:       eventType,
			Step:       tool.Name(),
			ParentStep: h.parentStep,
		},
	)
}

var (
	_ agent.RunHooks = (*progressHooks)(nil)
	_ agent.RunHooks = (*subProgressHooks)(nil)
)
