// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package cookiebanner

import (
	"time"

	"go.probo.inc/probo/pkg/llm"
)

// TrackerMappingAgentConfig configures the tracker-mapping agent
// (catalog identification). It uses DB-backed search tools and may also
// use Firecrawl for web search when an API key is supplied.
//
// MaxTokens and Temperature bound and steer the LLM call (the output is
// tiny structured JSON). Timeout caps a single agent run and MaxTurns
// bounds the agent reasoning loop. Zero-valued tuning fields fall back
// to package defaults.
type TrackerMappingAgentConfig struct {
	LLMClient       *llm.Client
	Model           string
	FirecrawlAPIKey string
	MaxTokens       *int
	Temperature     *float64
	Timeout         time.Duration
	MaxTurns        int
}

// TrackerEnrichmentAgentConfig configures the common-pattern enrichment
// agent (description research). It uses DB-backed search tools and may
// also use Firecrawl for web search when an API key is supplied.
//
// MaxTokens and Temperature bound and steer the LLM call (the output is
// tiny structured JSON). Timeout caps a single agent run and MaxTurns
// bounds the agent reasoning loop. Zero-valued tuning fields fall back
// to package defaults.
type TrackerEnrichmentAgentConfig struct {
	LLMClient       *llm.Client
	Model           string
	FirecrawlAPIKey string
	MaxTokens       *int
	Temperature     *float64
	Timeout         time.Duration
	MaxTurns        int
}
