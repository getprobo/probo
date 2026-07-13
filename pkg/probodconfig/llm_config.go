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

package probodconfig

type (
	// LLMProviderConfig holds authentication and connection settings for an
	// LLM provider (e.g. OpenAI, Anthropic).
	LLMProviderConfig struct {
		Type   string `json:"type"`              // "openai", "anthropic", "bedrock"
		APIKey string `json:"api-key,omitempty"` // for OpenAI and Anthropic
	}

	// LLMAgentConfig holds model parameters for a single agent. Provider
	// references one of the keys in AgentsConfig.Providers.
	LLMAgentConfig struct {
		Provider    string   `json:"provider,omitempty"` // key into AgentsConfig.Providers
		ModelName   string   `json:"model-name,omitempty"`
		Temperature *float64 `json:"temperature,omitempty"`
		MaxTokens   *int     `json:"max-tokens,omitempty"`
	}

	// EvidenceDescriberConfig holds worker-side tuning for the evidence
	// description background worker. LLM parameters for the same worker
	// live under AgentsConfig.EvidenceDescriber.
	EvidenceDescriberConfig struct {
		Interval       int `json:"interval"`    // seconds between polls
		StaleAfter     int `json:"stale-after"` // seconds before a claim is recycled
		MaxConcurrency int `json:"max-concurrency"`
	}

	// ThirdPartyVettingWorkerConfig holds worker-side tuning for the
	// third-party vetting background worker. LLM parameters for the
	// vetter live under AgentsConfig.ThirdPartyVetter.
	ThirdPartyVettingWorkerConfig struct {
		Interval       int `json:"interval"`    // seconds between polls
		StaleAfter     int `json:"stale-after"` // seconds before a claim is recycled
		MaxConcurrency int `json:"max-concurrency"`
	}

	// GitHubDiscoveryWorkerConfig tunes the shared agent-run worker that
	// schedules github-discovery and other agent_runs jobs. LLM parameters
	// for synthesis live under AgentsConfig.GitHubDiscovery.
	GitHubDiscoveryWorkerConfig struct {
		Interval       int `json:"interval"` // seconds between polls
		MaxConcurrency int `json:"max-concurrency"`
	}

	// TrackerMappingWorkerConfig holds worker-side tuning for the
	// tracker-mapping background worker. LLM parameters for the mapping
	// agent it runs live under AgentsConfig.TrackerMapping. AgentTimeout
	// and AgentMaxTurns bound a single mapping agent run.
	// DisambiguationAgentTimeout caps a single third-party
	// disambiguation agent run; that agent runs inside this worker but
	// uses its own LLM parameters from AgentsConfig.ThirdPartyDisambiguation.
	TrackerMappingWorkerConfig struct {
		Interval                   int `json:"interval"` // seconds between polls
		MaxConcurrency             int `json:"max-concurrency"`
		StaleAfter                 int `json:"stale-after"`   // seconds before a claim is recycled
		AgentTimeout               int `json:"agent-timeout"` // seconds, single agent run
		AgentMaxTurns              int `json:"agent-max-turns"`
		DisambiguationAgentTimeout int `json:"disambiguation-agent-timeout"` // seconds, single disambiguation run
	}

	// CommonPatternEnrichmentWorkerConfig holds worker-side tuning for
	// the common-pattern enrichment background worker. LLM parameters
	// for the enrichment agent live under AgentsConfig.TrackerEnrichment.
	CommonPatternEnrichmentWorkerConfig struct {
		Interval       int `json:"interval"` // seconds between polls
		MaxConcurrency int `json:"max-concurrency"`
		StaleAfter     int `json:"stale-after"`   // seconds before a claim is recycled
		AgentTimeout   int `json:"agent-timeout"` // seconds, single agent run
		AgentMaxTurns  int `json:"agent-max-turns"`
	}

	// CommonThirdPartyEnrichmentWorkerConfig holds worker-side tuning for
	// the common-third-party enrichment background worker. LLM parameters
	// for its agents live under AgentsConfig.CommonThirdPartyEnrichment.
	// ConfidenceThreshold is the floor a resolved value must clear before
	// it is written to its column; MaxAttempts caps stale-recovery
	// retries.
	CommonThirdPartyEnrichmentWorkerConfig struct {
		Interval            int     `json:"interval"` // seconds between polls
		MaxConcurrency      int     `json:"max-concurrency"`
		StaleAfter          int     `json:"stale-after"`   // seconds before a claim is recycled
		AgentTimeout        int     `json:"agent-timeout"` // seconds, single agent run
		AgentMaxTurns       int     `json:"agent-max-turns"`
		ConfidenceThreshold float64 `json:"confidence-threshold"`
		MaxAttempts         int     `json:"max-attempts"`
	}

	// AgentToolsConfig holds API keys and settings for external tools
	// that agents can use (web search, scraping, etc.).
	AgentToolsConfig struct {
		FirecrawlAPIKey string `json:"firecrawl-api-key,omitempty"`
	}

	// AgentsConfig groups LLM provider credentials and per-agent model
	// settings. Default is used as a fallback when an agent-specific field
	// is zero-valued.
	AgentsConfig struct {
		Providers                  map[string]LLMProviderConfig `json:"providers,omitempty"`
		Default                    LLMAgentConfig               `json:"defaults"`
		Probo                      LLMAgentConfig               `json:"probo,omitzero"`
		EvidenceDescriber          LLMAgentConfig               `json:"evidence-describer,omitzero"`
		ThirdPartyVetter           LLMAgentConfig               `json:"third-party-vetter,omitzero"`
		GitHubDiscovery            LLMAgentConfig               `json:"github-discovery,omitzero"`
		ThirdPartyDisambiguation   LLMAgentConfig               `json:"third-party-disambiguation,omitzero"`
		TrackerMapping             LLMAgentConfig               `json:"tracker-mapping,omitzero"`
		TrackerEnrichment          LLMAgentConfig               `json:"tracker-enrichment,omitzero"`
		CommonThirdPartyEnrichment LLMAgentConfig               `json:"common-third-party-enrichment,omitzero"`
		Tools                      AgentToolsConfig             `json:"tools,omitzero"`
	}
)

func (c LLMProviderConfig) IsZero() bool {
	return c.APIKey == ""
}

func (c LLMAgentConfig) IsZero() bool {
	return c.Provider == "" && c.ModelName == ""
}

func (c AgentToolsConfig) IsZero() bool {
	return c.FirecrawlAPIKey == ""
}

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
