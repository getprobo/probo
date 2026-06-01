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

package probodconfig

type (
	// LLMProviderConfig holds authentication and connection settings for an
	// LLM provider (e.g. OpenAI, Anthropic).
	LLMProviderConfig struct {
		Type   string `json:"type"`    // "openai", "anthropic", "bedrock"
		APIKey string `json:"api-key"` // for OpenAI and Anthropic
	}

	// LLMAgentConfig holds model parameters for a single agent. Provider
	// references one of the keys in AgentsConfig.Providers.
	LLMAgentConfig struct {
		Provider    string   `json:"provider"` // key into AgentsConfig.Providers
		ModelName   string   `json:"model-name"`
		Temperature *float64 `json:"temperature"`
		MaxTokens   *int     `json:"max-tokens"`
	}

	// EvidenceDescriberConfig holds worker-side tuning for the evidence
	// description background worker. LLM parameters for the same worker
	// live under AgentsConfig.EvidenceDescriber.
	EvidenceDescriberConfig struct {
		Interval       int `json:"interval"`    // seconds between polls
		StaleAfter     int `json:"stale-after"` // seconds before a claim is recycled
		MaxConcurrency int `json:"max-concurrency"`
	}

	// TrackerMappingWorkerConfig holds worker-side tuning for the
	// tracker-mapping background worker. LLM parameters for the agents
	// it runs live under AgentsConfig.TrackerMapping. AgentTimeout and
	// AgentMaxTurns bound a single agent run (the identification and
	// disambiguation agents).
	TrackerMappingWorkerConfig struct {
		Interval       int `json:"interval"` // seconds between polls
		MaxConcurrency int `json:"max-concurrency"`
		StaleAfter     int `json:"stale-after"`   // seconds before a claim is recycled
		AgentTimeout   int `json:"agent-timeout"` // seconds, single agent run
		AgentMaxTurns  int `json:"agent-max-turns"`
	}

	// CommonPatternEnrichmentWorkerConfig holds worker-side tuning for
	// the common-pattern enrichment background worker. LLM parameters
	// for the enrichment agent live under AgentsConfig.TrackerMapping
	// (the agents share one config slot).
	CommonPatternEnrichmentWorkerConfig struct {
		Interval       int `json:"interval"` // seconds between polls
		MaxConcurrency int `json:"max-concurrency"`
		StaleAfter     int `json:"stale-after"`   // seconds before a claim is recycled
		AgentTimeout   int `json:"agent-timeout"` // seconds, single agent run
		AgentMaxTurns  int `json:"agent-max-turns"`
	}

	// AgentToolsConfig holds API keys and settings for external tools
	// that agents can use (web search, scraping, etc.).
	AgentToolsConfig struct {
		FirecrawlAPIKey string `json:"firecrawl-api-key"`
	}

	// AgentsConfig groups LLM provider credentials and per-agent model
	// settings. Default is used as a fallback when an agent-specific field
	// is zero-valued.
	AgentsConfig struct {
		Providers          map[string]LLMProviderConfig `json:"providers"`
		Default            LLMAgentConfig               `json:"defaults"`
		Probo              LLMAgentConfig               `json:"probo"`
		EvidenceDescriber  LLMAgentConfig               `json:"evidence-describer"`
		ThirdPartyAssessor LLMAgentConfig               `json:"third-party-assessor"`
		TrackerMapping     LLMAgentConfig               `json:"tracker-mapping"`
		Tools              AgentToolsConfig             `json:"tools"`
	}
)

// ResolveAgent returns a fully populated LLMAgentConfig by filling in
// zero-valued fields from the default config.
func (c *AgentsConfig) ResolveAgent(agent LLMAgentConfig) LLMAgentConfig {
	if agent.Provider == "" {
		agent.Provider = c.Default.Provider
	}

	if agent.ModelName == "" {
		agent.ModelName = c.Default.ModelName
	}

	if agent.MaxTokens == nil && c.Default.MaxTokens != nil {
		agent.MaxTokens = new(*c.Default.MaxTokens)
	}

	return agent
}
