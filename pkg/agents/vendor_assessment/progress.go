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
	"context"
	"math/rand/v2"

	"go.probo.inc/probo/pkg/agent"
)

var toolMessages = map[string][]string{
	// Orchestrator tools (top-level steps).
	"crawl_vendor_website": {
		"Exploring vendor website for security and compliance pages",
		"Discovering key pages on the vendor website",
		"Mapping out the vendor's online presence",
	},
	"assess_security": {
		"Running technical security checks on the domain",
		"Evaluating the vendor's security posture",
		"Performing infrastructure security analysis",
	},
	"analyze_document": {
		"Reviewing document for key provisions",
		"Analyzing policy details and obligations",
		"Extracting important clauses from the document",
	},
	"assess_compliance": {
		"Identifying certifications and compliance frameworks",
		"Reviewing the vendor's compliance posture",
		"Checking for recognized security certifications",
	},

	// Security sub-agent tools.
	"check_ssl_certificate": {
		"Inspecting SSL/TLS certificate",
		"Verifying certificate validity and configuration",
		"Checking SSL certificate details",
	},
	"check_security_headers": {
		"Analyzing HTTP security headers",
		"Reviewing response headers for security best practices",
		"Checking for missing security headers",
	},
	"check_dmarc": {
		"Looking up DMARC email authentication record",
		"Checking DMARC policy configuration",
		"Verifying email spoofing protections",
	},
	"check_breaches": {
		"Searching for known data breaches",
		"Checking breach databases for past incidents",
		"Looking up the domain in breach records",
	},
	"check_dnssec": {
		"Verifying DNSSEC configuration",
		"Checking DNS security extensions",
		"Inspecting DNSSEC chain of trust",
	},
	"analyze_csp": {
		"Evaluating Content Security Policy",
		"Analyzing CSP directives for weaknesses",
		"Reviewing content security rules",
	},
	"check_cors": {
		"Checking CORS configuration",
		"Inspecting cross-origin resource sharing policy",
		"Reviewing CORS headers",
	},

	// Browser tools used by crawler, analyzer, and compliance sub-agents.
	"navigate_to_url": {
		"Opening page",
		"Loading page content",
		"Navigating to the page",
		"Visiting the page",
	},
	"extract_page_text": {
		"Reading page content",
		"Extracting text from the page",
		"Pulling content from the page",
		"Scanning page text",
	},
	"extract_links": {
		"Collecting links from the page",
		"Gathering all page links",
		"Discovering outgoing links",
	},
	"find_links_matching": {
		"Searching for relevant links",
		"Looking for links matching the pattern",
		"Filtering page links by keyword",
	},
}

func randomMessage(tool string) string {
	msgs, ok := toolMessages[tool]
	if !ok {
		return ""
	}

	return msgs[rand.IntN(len(msgs))]
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
