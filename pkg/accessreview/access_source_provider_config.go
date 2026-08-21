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

package accessreview

import (
	"go.probo.inc/probo/pkg/coredata"
)

// providerOrgConfig binds a connector provider to its picker-UI behavior.
//
// Listing the orgs themselves is not here: that needs the connector's
// credential, so it lives on the provider's Registration as
// ListOrganizations and is reached through an opened Handle.
//
// SelectedSlug returns the currently-configured org identifier for the
// connector (empty string if none).
//
// NeedsPicker reports whether the picker mutation should surface in the
// UI; false for 2-auto providers.
type providerOrgConfig struct {
	SelectedSlug func(c *coredata.Connector) string
	NeedsPicker  bool
}

// providerOrgConfigs is the single source of truth that the access-source
// picker paths dispatch through: the console/MCP picker resolvers
// (ProviderOrganizations, SelectedOrganization, NeedsConfiguration) and the
// AutoSelectDefaultOrganization defaulting run on create/update. Adding a
// provider takes one entry here.
var providerOrgConfigs = map[coredata.ConnectorProvider]providerOrgConfig{
	coredata.ConnectorProviderGitHub: {
		SelectedSlug: func(c *coredata.Connector) string {
			s, _ := coredata.ConnectorSettings[coredata.GitHubConnectorSettings](c)
			return s.Organization
		},
		NeedsPicker: true,
	},
	coredata.ConnectorProviderSentry: {
		SelectedSlug: func(c *coredata.Connector) string {
			s, _ := coredata.ConnectorSettings[coredata.SentryConnectorSettings](c)
			return s.OrganizationSlug
		},
		NeedsPicker: true,
	},
	coredata.ConnectorProviderGoogleAnalytics: {
		SelectedSlug: func(c *coredata.Connector) string {
			s, _ := coredata.ConnectorSettings[coredata.GoogleAnalyticsConnectorSettings](c)
			return s.AccountID
		},
		NeedsPicker: true,
	},
	coredata.ConnectorProviderGitLab: {
		SelectedSlug: func(c *coredata.Connector) string {
			s, _ := coredata.ConnectorSettings[coredata.GitLabConnectorSettings](c)
			return s.GroupID
		},
		NeedsPicker: true,
	},
	coredata.ConnectorProviderBitbucket: {
		SelectedSlug: func(c *coredata.Connector) string {
			s, _ := coredata.ConnectorSettings[coredata.BitbucketConnectorSettings](c)
			return s.Workspace
		},
		NeedsPicker: true,
	},
	coredata.ConnectorProviderHeroku: {
		SelectedSlug: func(c *coredata.Connector) string {
			s, _ := coredata.ConnectorSettings[coredata.HerokuConnectorSettings](c)
			return s.TeamID
		},
		NeedsPicker: true,
	},
	coredata.ConnectorProviderAsana: {
		SelectedSlug: func(c *coredata.Connector) string {
			s, _ := coredata.ConnectorSettings[coredata.AsanaConnectorSettings](c)
			return s.WorkspaceGID
		},
		NeedsPicker: true,
	},
	coredata.ConnectorProviderNetlify: {
		SelectedSlug: func(c *coredata.Connector) string {
			s, _ := coredata.ConnectorSettings[coredata.NetlifyConnectorSettings](c)
			return s.AccountSlug
		},
		NeedsPicker: true,
	},
	coredata.ConnectorProviderClickUp: {
		SelectedSlug: func(c *coredata.Connector) string {
			s, _ := coredata.ConnectorSettings[coredata.ClickUpConnectorSettings](c)
			return s.TeamID
		},
		NeedsPicker: true,
	},
	coredata.ConnectorProviderCloudflare: {
		SelectedSlug: func(c *coredata.Connector) string {
			s, _ := coredata.ConnectorSettings[coredata.CloudflareConnectorSettings](c)
			return s.AccountID
		},
		NeedsPicker: true,
	},
	coredata.ConnectorProviderDocuSign: {
		SelectedSlug: func(c *coredata.Connector) string {
			s, _ := coredata.ConnectorSettings[coredata.DocuSignConnectorSettings](c)
			return s.AccountID
		},
		NeedsPicker: true,
	},
	// Pattern 2-auto: identifier is captured during the OAuth callback
	// (subdomain for PagerDuty, teamId or fallback /v2/user.id for
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

// ProviderSupportsOrganizationPicker reports whether the provider surfaces an
// organization picker at all.
//
// It distinguishes two cases the org listing itself cannot: a provider that
// offers a picker and returned nothing, versus a provider that has no picker
// because its identifier is captured during the OAuth callback (PagerDuty's
// subdomain, Vercel's team, Datadog's domain, Zendesk's subdomain). Both come
// back as an empty list, but only the first means something is wrong — telling
// a healthy 2-auto source that its organization "may not have approved Probo"
// would be a lie.
func (s *Service) ProviderSupportsOrganizationPicker(p coredata.ConnectorProvider) bool {
	reg, ok := s.providerRegistry.Get(p)

	return ok && reg.ListOrganizations != nil
}
