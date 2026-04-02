// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package probod

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

	// AgentsConfig groups LLM provider credentials and per-agent model
	// settings. Default is used as a fallback when an agent-specific field
	// is zero-valued.
	AgentsConfig struct {
		Providers         map[string]LLMProviderConfig `json:"providers"`
		Default           LLMAgentConfig               `json:"default"`
		Probo             LLMAgentConfig               `json:"probo"`
		EvidenceDescriber LLMAgentConfig               `json:"evidence-describer"`
		VendorAssessor    LLMAgentConfig               `json:"vendor-assessor"`
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
	if agent.Temperature == nil {
		agent.Temperature = c.Default.Temperature
	}
	if agent.MaxTokens == nil {
		agent.MaxTokens = c.Default.MaxTokens
	}
	return agent
}
