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

package coredata

import (
	"encoding/json"
	"fmt"

	"go.probo.inc/probo/pkg/connector"
)

type (
	SlackConnectorSettings struct {
		Channel   string `json:"channel,omitempty"`
		ChannelID string `json:"channel_id,omitempty"`
	}

	TallyConnectorSettings struct {
		OrganizationID string `json:"organization_id"`
	}

	OnePasswordConnectorSettings struct {
		SCIMBridgeURL string `json:"scim_bridge_url"`
	}

	SentryConnectorSettings struct {
		OrganizationSlug string `json:"organization_slug"`
	}

	GrafanaConnectorSettings struct {
		BaseURL string `json:"base_url"`
	}

	SupabaseConnectorSettings struct {
		OrganizationSlug string `json:"organization_slug"`
	}

	GitHubConnectorSettings struct {
		Organization string `json:"organization"`
	}

	OnePasswordUsersAPISettings struct {
		AccountID string `json:"account_id"`
		Region    string `json:"region"`
	}

	GitLabConnectorSettings struct {
		GroupID string `json:"group_id"`
	}

	BitbucketConnectorSettings struct {
		Workspace string `json:"workspace"`
	}

	HerokuConnectorSettings struct {
		TeamID string `json:"team_id"`
	}

	PagerDutyConnectorSettings struct {
		Subdomain string `json:"subdomain"`
	}

	AsanaConnectorSettings struct {
		WorkspaceGID string `json:"workspace_gid"`
	}

	NetlifyConnectorSettings struct {
		AccountSlug string `json:"account_slug"`
	}

	ClickUpConnectorSettings struct {
		TeamID string `json:"team_id"`
	}

	VercelConnectorSettings struct {
		TeamID string `json:"team_id"`
	}

	MetabaseConnectorSettings struct {
		InstanceURL string `json:"instance_url"`
	}

	// PostHogConnectorSettings carries the data-API base host for the
	// PostHog provider, which spans cloud and self-hosted deployments. For
	// API-key connections it is the region-pinned host
	// (https://us.posthog.com / https://eu.posthog.com) or a self-hosted
	// instance URL. It is empty for cloud OAuth connections — the driver
	// then discovers the region (us/eu) by probing, since the
	// region-agnostic oauth.posthog.com gateway does not serve the data API.
	PostHogConnectorSettings struct {
		BaseURL string `json:"base_url"`
	}
)

// GrantType returns the OAuth2 grant type recorded on the connector's
// Connection, or the empty string when the connector is not an OAuth2
// connector. Driver factories that dispatch on grant type (1Password)
// read this instead of inspecting the typed Connection directly.
func (c *Connector) GrantType() string {
	if oauth2Conn, ok := c.Connection.(*connector.OAuth2Connection); ok {
		return string(oauth2Conn.GrantType)
	}

	return ""
}

// SetSettings marshals a typed settings struct into the connector's RawSettings.
func (c *Connector) SetSettings(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("cannot marshal connector settings: %w", err)
	}

	c.RawSettings = data

	return nil
}

// ConnectorSettings unmarshals the connector's RawSettings into the
// requested settings struct. Empty or null RawSettings yields the zero
// value with no error. Use as:
//
//	settings, err := coredata.ConnectorSettings[coredata.GitHubConnectorSettings](dbConnector)
func ConnectorSettings[T any](c *Connector) (T, error) {
	var s T
	if len(c.RawSettings) == 0 || string(c.RawSettings) == "null" {
		return s, nil
	}

	if err := json.Unmarshal(c.RawSettings, &s); err != nil {
		return s, fmt.Errorf("cannot unmarshal connector settings: %w", err)
	}

	return s, nil
}
