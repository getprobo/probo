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

package cookiebanner

import (
	"time"

	"go.probo.inc/probo/pkg/llm"
)

// TrackerAgentsConfig configures the tracker agents that share one LLM
// client, model, and tool surface: the tracker-mapping agent (catalog
// identification) and the common-pattern enrichment agent (description
// research). Both use DB-backed search tools and may also use Firecrawl
// for web search when an API key is supplied.
//
// MaxTokens and Temperature bound and steer each LLM call (both
// outputs are tiny structured JSON). AgentTimeout caps a single agent
// run, and the per-worker max-turns bound the agent reasoning loop.
// Zero-valued tuning fields fall back to package defaults.
type TrackerAgentsConfig struct {
	LLMClient          *llm.Client
	Model              string
	FirecrawlAPIKey    string
	MaxTokens          *int
	Temperature        *float64
	AgentTimeout       time.Duration
	MappingMaxTurns    int
	EnrichmentMaxTurns int
}
