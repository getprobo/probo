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

package console_v1

import (
	"context"
	"net/http"

	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
)

// providerOrgConfig binds a connector provider to its picker-UI behavior.
//
// ListOrgs returns the orgs/workspaces/teams the authenticated user can
// scope the connector to (nil for Pattern 2-auto providers like
// PagerDuty and Vercel where the value is captured during OAuth).
//
// SelectedSlug returns the currently-configured org identifier for the
// connector (empty string if none).
//
// NeedsPicker reports whether the picker mutation should surface in the
// UI; false for 2-auto providers.
type providerOrgConfig struct {
	ListOrgs     func(ctx context.Context, httpClient *http.Client) ([]drivers.Organization, error)
	SelectedSlug func(c *coredata.Connector) string
	NeedsPicker  bool
}

// providerOrgConfigs is the single source of truth that the three
// AccessReviewSource picker resolvers (ProviderOrganizations,
// SelectedOrganization, NeedsConfiguration) dispatch through. Adding a
// provider takes one entry here, not three switch arms.
var providerOrgConfigs = map[coredata.ConnectorProvider]providerOrgConfig{
	coredata.ConnectorProviderGitHub: {
		ListOrgs: drivers.ListGitHubOrganizations,
		SelectedSlug: func(c *coredata.Connector) string {
			s, _ := coredata.ConnectorSettings[coredata.GitHubConnectorSettings](c)
			return s.Organization
		},
		NeedsPicker: true,
	},
	coredata.ConnectorProviderSentry: {
		ListOrgs: drivers.ListSentryOrganizations,
		SelectedSlug: func(c *coredata.Connector) string {
			s, _ := coredata.ConnectorSettings[coredata.SentryConnectorSettings](c)
			return s.OrganizationSlug
		},
		NeedsPicker: true,
	},
	coredata.ConnectorProviderGitLab: {
		ListOrgs: drivers.ListGitLabOrganizations,
		SelectedSlug: func(c *coredata.Connector) string {
			s, _ := coredata.ConnectorSettings[coredata.GitLabConnectorSettings](c)
			return s.GroupID
		},
		NeedsPicker: true,
	},
	coredata.ConnectorProviderBitbucket: {
		ListOrgs: drivers.ListBitbucketOrganizations,
		SelectedSlug: func(c *coredata.Connector) string {
			s, _ := coredata.ConnectorSettings[coredata.BitbucketConnectorSettings](c)
			return s.Workspace
		},
		NeedsPicker: true,
	},
	coredata.ConnectorProviderHeroku: {
		ListOrgs: drivers.ListHerokuOrganizations,
		SelectedSlug: func(c *coredata.Connector) string {
			s, _ := coredata.ConnectorSettings[coredata.HerokuConnectorSettings](c)
			return s.TeamID
		},
		NeedsPicker: true,
	},
	coredata.ConnectorProviderAsana: {
		ListOrgs: drivers.ListAsanaOrganizations,
		SelectedSlug: func(c *coredata.Connector) string {
			s, _ := coredata.ConnectorSettings[coredata.AsanaConnectorSettings](c)
			return s.WorkspaceGID
		},
		NeedsPicker: true,
	},
	coredata.ConnectorProviderNetlify: {
		ListOrgs: drivers.ListNetlifyOrganizations,
		SelectedSlug: func(c *coredata.Connector) string {
			s, _ := coredata.ConnectorSettings[coredata.NetlifyConnectorSettings](c)
			return s.AccountSlug
		},
		NeedsPicker: true,
	},
	coredata.ConnectorProviderClickUp: {
		ListOrgs: drivers.ListClickUpOrganizations,
		SelectedSlug: func(c *coredata.Connector) string {
			s, _ := coredata.ConnectorSettings[coredata.ClickUpConnectorSettings](c)
			return s.TeamID
		},
		NeedsPicker: true,
	},
	coredata.ConnectorProviderDocuSign: {
		ListOrgs: drivers.ListDocuSignOrganizations,
		SelectedSlug: func(c *coredata.Connector) string {
			s, _ := coredata.ConnectorSettings[coredata.DocuSignConnectorSettings](c)
			return s.AccountID
		},
		NeedsPicker: true,
	},
	// Pattern 2-auto: identifier is captured during the OAuth callback
	// (subdomain for PagerDuty, team_id or fallback /v2/user.id for
	// Vercel). No picker UI; NeedsPicker = false.
	coredata.ConnectorProviderPagerDuty: {
		SelectedSlug: func(c *coredata.Connector) string {
			s, _ := coredata.ConnectorSettings[coredata.PagerDutyConnectorSettings](c)
			return s.Subdomain
		},
	},
	coredata.ConnectorProviderVercel: {
		SelectedSlug: func(c *coredata.Connector) string {
			s, _ := coredata.ConnectorSettings[coredata.VercelConnectorSettings](c)
			return s.TeamID
		},
	},
	// Pattern 2-auto: the API domain is captured during the OAuth
	// callback from Datadog's `domain` parameter; no picker UI.
	coredata.ConnectorProviderDatadog: {
		SelectedSlug: func(c *coredata.Connector) string {
			s, _ := coredata.ConnectorSettings[coredata.DatadogConnectorSettings](c)
			return s.Domain
		},
	},
	// Pattern 2-auto: the subdomain is collected at initiate and persisted
	// from the signed OAuth state on the callback; no picker UI.
	coredata.ConnectorProviderZendesk: {
		SelectedSlug: func(c *coredata.Connector) string {
			s, _ := coredata.ConnectorSettings[coredata.ZendeskConnectorSettings](c)
			return s.Subdomain
		},
	},
}
