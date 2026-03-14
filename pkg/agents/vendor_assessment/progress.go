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

	"go.probo.inc/probo/pkg/agent"
)

var toolMessages = map[string]string{
	// Orchestrator tools (top-level steps).
	"crawl_vendor_website": "Crawling vendor website to discover security and compliance pages",
	"assess_security":      "Running technical security checks",
	"analyze_document":     "Analyzing document provisions",
	"assess_compliance":    "Identifying certifications and compliance frameworks",

	// Security sub-agent tools.
	"check_ssl_certificate":  "Checking SSL certificate",
	"check_security_headers": "Checking security headers",
	"check_dmarc":            "Checking DMARC record",
	"check_breaches":         "Checking for known data breaches",
	"check_dnssec":           "Checking DNSSEC configuration",
	"analyze_csp":            "Analyzing Content Security Policy",
	"check_cors":             "Checking CORS configuration",

	// Browser tools used by crawler, analyzer, and compliance sub-agents.
	"navigate_to_url":     "Navigating to page",
	"extract_page_text":   "Extracting page content",
	"extract_links":       "Extracting links from page",
	"find_links_matching": "Searching for matching links",
}

// progressHooks translates orchestrator-level tool events into progress events.
type progressHooks struct {
	agent.NoOpHooks
	reporter agent.ProgressReporter
}

func (h *progressHooks) OnToolStart(ctx context.Context, _ *agent.Agent, tool agent.Tool, _ string) {
	msg, ok := toolMessages[tool.Name()]
	if !ok {
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
	msg, ok := toolMessages[tool.Name()]
	if !ok {
		return
	}

	eventType := agent.ProgressEventStepCompleted
	if err != nil {
		eventType = agent.ProgressEventStepFailed
	}

	h.reporter(
		ctx,
		agent.ProgressEvent{
			Type:    eventType,
			Step:    tool.Name(),
			Message: msg,
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

func (h *subProgressHooks) OnToolStart(ctx context.Context, _ *agent.Agent, tool agent.Tool, _ string) {
	msg, ok := toolMessages[tool.Name()]
	if !ok {
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
	msg, ok := toolMessages[tool.Name()]
	if !ok {
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
			Message:    msg,
		},
	)
}

var (
	_ agent.RunHooks = (*progressHooks)(nil)
	_ agent.RunHooks = (*subProgressHooks)(nil)
)
