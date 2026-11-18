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

package probodconfig

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

type (
	// LLMProviderConfig holds authentication and connection settings for an
	// LLM provider (e.g. OpenAI, Anthropic).
	LLMProviderConfig struct {
		Type    string `json:"type"`               // "openai", "anthropic", "gemini", "openai_compatible", "bedrock"
		APIKey  string `json:"api-key,omitempty"`  // optional for openai_compatible: a locally hosted server often has no auth
		BaseURL string `json:"base-url,omitempty"` // required for openai_compatible; empty uses the provider default elsewhere

		// Vertex mode for the gemini provider. When Project is set the
		// provider authenticates with Application Default Credentials
		// instead of APIKey, so a workload on GCP needs no credential of
		// its own.
		Project  string `json:"project,omitempty"`
		Location string `json:"location,omitempty"`
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

	// TrackerMappingWorkerConfig holds worker-side tuning for the
	// tracker-mapping background worker. LLM parameters for the mapping
	// agent it runs live under AgentsConfig.TrackerMapping. AgentTimeout
	// and AgentMaxTurns bound a single mapping agent run.
	TrackerMappingWorkerConfig struct {
		Interval       int `json:"interval"` // seconds between polls
		MaxConcurrency int `json:"max-concurrency"`
		StaleAfter     int `json:"stale-after"`   // seconds before a claim is recycled
		AgentTimeout   int `json:"agent-timeout"` // seconds, single agent run
		AgentMaxTurns  int `json:"agent-max-turns"`
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
		TrackerMapping             LLMAgentConfig               `json:"tracker-mapping,omitzero"`
		TrackerEnrichment          LLMAgentConfig               `json:"tracker-enrichment,omitzero"`
		CommonThirdPartyEnrichment LLMAgentConfig               `json:"common-third-party-enrichment,omitzero"`
		Slackbot                   LLMAgentConfig               `json:"slackbot,omitzero"`
		Tools                      AgentToolsConfig             `json:"tools,omitzero"`
	}
)

// Validate reports a BaseURL that is not an absolute URL, so a typo fails at
// startup instead of on every LLM call. An empty BaseURL is valid for every
// provider except openai_compatible and selects the provider default.
//
// https is required except when the host is local. A self-hosted model served
// by llama.cpp, vLLM or Ollama listens on plain http -- on loopback, on a
// private address, or on a single-label service name inside a compose project
// or cluster -- and requiring TLS there would make the openai_compatible
// provider unusable for the case it exists to serve. Nothing outside those
// hosts may go over plaintext, because that is where a credential would
// actually cross a network.
func (c LLMProviderConfig) Validate() error {
	if c.BaseURL == "" {
		if c.Type == "openai_compatible" {
			return errors.New("openai_compatible provider requires a base URL")
		}

		return nil
	}

	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return fmt.Errorf("cannot parse %s base URL: %w", c.Type, err)
	}

	if u.Host == "" {
		return fmt.Errorf("%s base URL must be an absolute URL, got %q", c.Type, c.BaseURL)
	}

	switch u.Scheme {
	case "https":
		return nil
	case "http":
		if isLocalHost(u.Hostname()) {
			return nil
		}

		return fmt.Errorf(
			"%s base URL must use https unless the host is local, got %q",
			c.Type,
			c.BaseURL,
		)
	}

	return fmt.Errorf("%s base URL must be http or https, got %q", c.Type, c.BaseURL)
}

// isLocalHost reports whether plaintext to this host stays inside the
// deployment: loopback, a private or link-local address, or a name with no dot
// in it (a compose service or a Kubernetes service in the same namespace).
func isLocalHost(host string) bool {
	if host == "" {
		return false
	}

	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast()
	}

	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return true
	}

	return !strings.Contains(host, ".")
}

// IsZero reports whether the provider is unconfigured. A base URL alone is
// enough to count as configured: openai_compatible pointed at a local server
// legitimately has no API key, and keying this on APIKey alone silently
// dropped such a provider from the registry, which surfaced only as
// "unknown LLM provider" at startup.
func (c LLMProviderConfig) IsZero() bool {
	return c.APIKey == "" && c.BaseURL == "" && c.Project == ""
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
