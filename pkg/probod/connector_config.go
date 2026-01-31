// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"go.probo.inc/probo/pkg/connector"
)

type (
	connectorConfig struct {
		Provider string                 `json:"provider"`
		Protocol connector.ProtocolType `json:"protocol"`
		Config   connector.Connector    `json:"-"`
		Settings any                    `json:"-"`
	}

	connectorConfigOAuth2 struct {
		ClientID        string            `json:"client-id"`
		ClientSecret    string            `json:"client-secret"`
		RedirectURI     string            `json:"redirect-uri"`
		AuthURL         string            `json:"auth-url"`
		TokenURL        string            `json:"token-url"`
		Scopes          []string          `json:"scopes"`
		ExtraAuthParams map[string]string `json:"extra-auth-params,omitempty"`
	}
)

func (c *config) GetSlackSigningSecret() string {
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

func (c *connectorConfig) UnmarshalJSON(data []byte) error {
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
		var config connectorConfigOAuth2
		if err := json.NewDecoder(bytes.NewReader(tmp.RawConfig)).Decode(&config); err != nil {
			return fmt.Errorf("cannot unmarshal oauth2 connector config: %w", err)
		}

		oauth2Connector := connector.OAuth2Connector{
			ClientID:        config.ClientID,
			ClientSecret:    config.ClientSecret,
			RedirectURI:     config.RedirectURI,
			AuthURL:         config.AuthURL,
			TokenURL:        config.TokenURL,
			Scopes:          config.Scopes,
			ExtraAuthParams: config.ExtraAuthParams,
		}

		c.Config = &oauth2Connector
	default:
		return fmt.Errorf("unknown connector protocol: %q", c.Protocol)
	}

	return nil
}
