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

type ConnectorConfigOAuth2 struct {
	ClientID     string `json:"client-id"`
	ClientSecret string `json:"client-secret"`
	// IntegrationSlug is an operator-supplied value used by providers
	// whose authorization URL embeds it as a path segment (Vercel-style
	// integrations). It is propagated onto OAuth2Connector.IntegrationSlug
	// and resolved by (*provider.Registry).ApplyOAuth2Defaults.
	IntegrationSlug string `json:"integration-slug,omitempty"`
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
	default:
		return fmt.Errorf("unknown connector protocol: %q", c.Protocol)
	}

	return nil
}
