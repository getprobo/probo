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
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"go.probo.inc/probo/pkg/connector"
)

type ConnectorConfig struct {
	Provider    string                 `json:"provider"`
	Protocol    connector.ProtocolType `json:"protocol"`
	Config      connector.Connector    `json:"-"`
	RawConfig   any                    `json:"config,omitempty"`
	Settings    any                    `json:"-"`
	RawSettings any                    `json:"settings,omitempty"`
	// APIKey holds the Probo-supplied credential for an api_key-protocol
	// connector (ManagedAPIKey providers such as Crisp). It is resolved
	// from RawConfig by UnmarshalJSON and registered on the provider
	// Registry by probod. Empty for OAuth2 connectors.
	APIKey string `json:"-"`
	// ResourceID holds an optional Probo-supplied resource identifier for an
	// api_key-protocol connector, distinct from the credential (e.g. the
	// Crisp plugin ID required by the per-website plugin API). Resolved from
	// RawConfig by UnmarshalJSON and registered on the provider Registry by
	// probod. Empty for connectors that need no such identifier.
	ResourceID string `json:"-"`
}

// ConnectorEndpointsConfig overrides the endpoints compiled into a provider's
// registration, so a non-production deployment can point a connector at the
// provider's sandbox — DocuSign's account-d hosts, for instance — without a
// code change.
//
// It is keyed by provider on Config.ConnectorEndpoints rather than nested in a
// ConnectorConfig entry because a provider needs no credentials entry to have
// endpoints: an API-key provider is configured by the customer, yet its API
// base is still compiled in and still worth repointing on a staging
// deployment.
//
// Every field is optional; an omitted field keeps the compiled default. A
// field the provider builds per flow or per connection cannot be overridden —
// probod refuses to start rather than accept a value it would ignore.
type ConnectorEndpointsConfig struct {
	Auth  string `json:"auth,omitempty"`
	Token string `json:"token,omitempty"`
	Probe string `json:"probe,omitempty"`
	// Identity is the host a provider's driver resolves its real data host
	// from, for providers that split the two. It must move together with
	// Probe — probod refuses to start on a mismatch, because a moved probe
	// with a stale identity host reports healthy while every data call still
	// reaches the real provider.
	Identity string `json:"identity,omitempty"`
	APIBase  string `json:"api-base,omitempty"`
}

type ConnectorConfigOAuth2 struct {
	ClientID     string `json:"client-id"`
	ClientSecret string `json:"client-secret"`
	// IntegrationSlug is an operator-supplied value used by providers
	// whose authorization URL embeds it as a path segment (Vercel-style
	// integrations). It is propagated onto OAuth2Connector.IntegrationSlug
	// and resolved by (*provider.Registry).ApplyOAuth2Defaults.
	IntegrationSlug string `json:"integration-slug,omitempty"`
}

type ConnectorConfigGitHubApp struct {
	AppID        string `json:"app-id"`
	ClientID     string `json:"client-id"`
	ClientSecret string `json:"client-secret"`
	Slug         string `json:"slug"`
	PrivateKey   string `json:"private-key"`
}

// ConnectorConfigAPIKey carries the Probo-held API key for a
// ManagedAPIKey connector (e.g. Crisp's marketplace plugin token). The
// operator supplies it via bootstrap env; probod registers it on the
// provider Registry so the create-connector resolver can inject it.
// ResourceID is an optional companion identifier (e.g. the Crisp plugin
// ID) some managed connectors need beyond the credential.
type ConnectorConfigAPIKey struct {
	APIKey     string `json:"api-key"`
	ResourceID string `json:"resource-id,omitempty"`
}

func (c *Config) GetSlackSigningSecret() string {
	if c.Notifications.Slack.SigningSecret != "" {
		return c.Notifications.Slack.SigningSecret
	}

	for _, conn := range c.Connectors {
		if conn.Provider == "SLACK" {
			if settings, ok := conn.Settings.(map[string]any); ok {
				if signingSecret, ok := settings["signing-secret"].(string); ok {
					return signingSecret
				}
			}
		}
	}

	return ""
}

func (c *ConnectorConfig) UnmarshalJSON(data []byte) error {
	var tmp struct {
		Provider  string          `json:"provider"`
		Protocol  string          `json:"protocol"`
		RawConfig json.RawMessage `json:"config"`
		Settings  json.RawMessage `json:"settings"`
	}

	if err := json.NewDecoder(bytes.NewReader(data)).Decode(&tmp); err != nil {
		return fmt.Errorf("cannot unmarshal connector config: %w", err)
	}

	c.Provider = strings.ToUpper(tmp.Provider)
	c.Protocol = connector.ProtocolType(strings.ToUpper(tmp.Protocol))

	if len(tmp.Settings) > 0 {
		var settings map[string]any
		if err := json.NewDecoder(bytes.NewReader(tmp.Settings)).Decode(&settings); err != nil {
			return fmt.Errorf("cannot unmarshal settings: %w", err)
		}

		c.Settings = settings
	}

	switch c.Protocol {
	case connector.ProtocolOAuth2:
		var config ConnectorConfigOAuth2
		if err := json.NewDecoder(bytes.NewReader(tmp.RawConfig)).Decode(&config); err != nil {
			return fmt.Errorf("cannot unmarshal oauth2 connector config: %w", err)
		}

		oauth2Connector := connector.OAuth2Connector{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
		}

		oauth2Connector.IntegrationSlug = config.IntegrationSlug

		c.Config = &oauth2Connector
	case connector.ProtocolAPIKey:
		var config ConnectorConfigAPIKey
		if err := json.NewDecoder(bytes.NewReader(tmp.RawConfig)).Decode(&config); err != nil {
			return fmt.Errorf("cannot unmarshal api key connector config: %w", err)
		}

		c.APIKey = config.APIKey
		c.ResourceID = config.ResourceID
	case connector.ProtocolGitHubApp:
		if c.Provider != connector.GitHubProvider {
			return fmt.Errorf("github_app connector protocol requires GITHUB provider")
		}

		var config ConnectorConfigGitHubApp
		if err := json.NewDecoder(bytes.NewReader(tmp.RawConfig)).Decode(&config); err != nil {
			return fmt.Errorf("cannot unmarshal github app connector config: %w", err)
		}

		c.Config = &connector.GitHubAppConnector{
			AppID:        config.AppID,
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			Slug:         config.Slug,
			PrivateKey:   config.PrivateKey,
		}
	default:
		return fmt.Errorf("unknown connector protocol: %q", c.Protocol)
	}

	return nil
}
