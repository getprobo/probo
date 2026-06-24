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
// to package defaults. ChromeAddr enables the read-only browser toolset
// (so the agent can open cookie-database and cookie-policy pages to read
// the true setter); when empty the agent relies on web search alone.
type TrackerMappingAgentConfig struct {
	LLMClient       *llm.Client
	Model           string
	FirecrawlAPIKey string
	ChromeAddr      string
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
// to package defaults. ChromeAddr enables the read-only browser toolset
// (so the agent can open authoritative vendor and cookie-database pages
// to ground a description); when empty the agent relies on web search
// alone.
type TrackerEnrichmentAgentConfig struct {
	LLMClient       *llm.Client
	Model           string
	FirecrawlAPIKey string
	ChromeAddr      string
	MaxTokens       *int
	Temperature     *float64
	Timeout         time.Duration
	MaxTurns        int
}
