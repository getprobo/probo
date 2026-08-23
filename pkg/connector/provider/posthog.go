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

package provider

import (
	"go.probo.inc/probo/pkg/coredata"
)

// posthogRegistration is PostHog — Cloud (US + EU) and self-hosted, under one
// provider. OAuth (CIMD public client) is the preferred path for Cloud: the
// region-agnostic oauth.posthog.com gateway handles the handshake for both
// regions, after which the driver resolves the data region (us/eu) itself,
// since that gateway does not serve the data API. An API-key fallback covers
// both deployments: Cloud personal API keys are region-pinned (the customer
// picks us/eu) and self-hosted connections carry an instance URL. Both store a
// single data-host BaseURL; cloud OAuth connections leave it empty for lazy
// region probing.
func posthogRegistration() *Registration {
	return &Registration{
		Provider:    coredata.ConnectorProviderPostHog,
		DisplayName: "PostHog",
		Endpoints: Endpoints{
			Auth:  "https://oauth.posthog.com/oauth/authorize/",
			Token: "https://oauth.posthog.com/oauth/token/",
		},
		// PublicClient: PostHog OAuth uses the CIMD flow — no client_secret,
		// authenticated by PKCE. probod auto-registers this connector with
		// the deployment's hosted CIMD client_id; no operator OAuth app or
		// credentials are required.
		// See Registration.EndpointOverrideUnsupported: cloud region discovery
		// (resolveBaseURL) probes us.posthog.com / eu.posthog.com directly, never
		// Endpoints, so no override can move the data host it resolves.
		EndpointOverrideUnsupported: "its Cloud region discovery is hardcoded to us.posthog.com and eu.posthog.com, never Endpoints",
		// API-key connections are either PostHog Cloud (a region, us/eu) or
		// self-hosted (an instance URL). The two are mutually exclusive, so
		// neither is individually Required; apiKeyConnectorSettings enforces
		// that exactly one is supplied.
		OAuth2: &OAuth2Spec{
			PublicClient:      true,
			TokenEndpointAuth: "none",
			RequiresPKCE:      true,
			Scopes:            []string{"organization:read", "organization_member:read"},
			ExtraAuthParams:   map[string]string{"required_access_level": "organization"},
		},
		APIKey: &APIKeySpec{
			ExtraSettings: []ExtraSetting{
				{Key: "region", Label: "Region"},
				{Key: "instanceUrl", Label: "Instance URL"},
			},
		},
		// required_access_level=organization makes consent org-scoped so
		// organization_member:read applies org-wide and the org endpoints
		// resolve @current to the granted organization.
		Probe: ProbeOver(probePostHog),
	}
}
