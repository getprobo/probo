// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

	SigNozConnectorSettings struct {
		BaseURL string `json:"base_url"`
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

	// DatadogConnectorSettings holds the per-customer Datadog site captured
	// during the OAuth callback. Region is the site key (e.g. "US3") used for
	// the AccessSource title; Domain is the API domain (e.g.
	// "us3.datadoghq.com") the driver and name resolver use to build hosts.
	DatadogConnectorSettings struct {
		Region string `json:"region"`
		Domain string `json:"domain"`
	}

	// OktaConnectorSettings holds the customer's Okta org domain (the bare
	// host, e.g. "acme.okta.com") supplied with the API token. It is the
	// single-tenant identifier the driver and name resolver use to build
	// the per-org API host (https://<domain>/api/v1/...). Okta has no
	// central API gateway — every org authenticates against its own
	// domain — so the host is operator-supplied and validated (see
	// connector.NormalizeOktaDomain) before it reaches URL construction.
	OktaConnectorSettings struct {
		Domain string `json:"domain"`
	}

	// ZendeskConnectorSettings holds the per-customer Zendesk subdomain
	// captured at connect time (the customer types it before the OAuth
	// redirect, and it rides the signed state token to the callback —
	// Zendesk does not echo it back). Subdomain is the <subdomain> part of
	// <subdomain>.zendesk.com, used by the driver to build the API host and
	// by the name resolver for the AccessSource title.
	ZendeskConnectorSettings struct {
		Subdomain string `json:"subdomain"`
	}

	BetterStackConnectorSettings struct {
		TeamName string `json:"team_name"`
	}

	QoveryConnectorSettings struct {
		OrganizationID string `json:"organization_id"`
	}

	// RenderConnectorSettings stores the Render workspace identifier. The
	// value is Render's owner ID (e.g. "tea-..." for a team workspace or
	// "usr-..." for a personal one), surfaced to operators as "Workspace ID"
	// and used as the {ownerId} path segment on /v1/owners/{ownerId}/...
	RenderConnectorSettings struct {
		OwnerID string `json:"owner_id"`
	}

	NeonConnectorSettings struct {
		OrganizationID string `json:"organization_id"`
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
