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

	// LLMConfig holds model parameters for a single LLM consumer. Provider
	// references one of the keys in LLMSettings.Providers.
	LLMConfig struct {
		Provider    string   `json:"provider"` // key into LLMSettings.Providers
		ModelName   string   `json:"model-name"`
		Temperature *float64 `json:"temperature"`
		MaxTokens   *int     `json:"max-tokens"`
	}

	// LLMSettings groups LLM provider credentials and default model
	// settings. Defaults is used as a fallback when a consumer-specific
	// field is zero-valued.
	LLMSettings struct {
		Providers map[string]LLMProviderConfig `json:"providers"`
		Defaults  LLMConfig                    `json:"defaults"`
	}
)

// ResolveLLMConfig returns a fully populated LLMConfig by filling in
// zero-valued fields from the defaults.
func (s *LLMSettings) ResolveLLMConfig(cfg LLMConfig) LLMConfig {
	if cfg.Provider == "" {
		cfg.Provider = s.Defaults.Provider
	}
	if cfg.ModelName == "" {
		cfg.ModelName = s.Defaults.ModelName
	}
	if cfg.Temperature == nil {
		cfg.Temperature = s.Defaults.Temperature
	}
	if cfg.MaxTokens == nil {
		cfg.MaxTokens = s.Defaults.MaxTokens
	}
	return cfg
}
