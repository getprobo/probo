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

package provider

import (
	"context"
	"fmt"
	"net/http"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
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

		// PublicClient: PostHog OAuth uses the CIMD flow — no client_secret,
		// authenticated by PKCE. probod auto-registers this connector with
		// the deployment's hosted CIMD client_id; no operator OAuth app or
		// credentials are required.
		PublicClient:      true,
		AuthURL:           "https://oauth.posthog.com/oauth/authorize/",
		TokenURL:          "https://oauth.posthog.com/oauth/token/",
		TokenEndpointAuth: "none",
		RequiresPKCE:      true,
		OAuth2Scopes:      []string{"organization:read", "organization_member:read"},
		// required_access_level=organization makes consent org-scoped so
		// organization_member:read applies org-wide and the org endpoints
		// resolve @current to the granted organization.
		ExtraAuthParams: map[string]string{"required_access_level": "organization"},
		Probe:           probePostHog,
		SupportsAPIKey:  true,
		// API-key connections are either PostHog Cloud (a region, us/eu) or
		// self-hosted (an instance URL). The two are mutually exclusive, so
		// neither is individually Required; apiKeyConnectorSettings enforces
		// that exactly one is supplied.
		ExtraSettings: []ExtraSetting{
			{Key: "region", Label: "Region"},
			{Key: "instanceUrl", Label: "Instance URL"},
		},

		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.PostHogConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read posthog connector settings: %w", err)
			}

			// BaseURL is empty for cloud OAuth connections; the driver then
			// discovers the region (us/eu) lazily by probing, since the
			// oauth.posthog.com gateway does not serve the data API.
			return drivers.NewPostHogDriver(c, s.BaseURL), nil
		},
		NewNameResolver: func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.PostHogConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read posthog connector settings", log.Error(err))
				return nil
			}

			return drivers.NewPostHogNameResolver(c, s.BaseURL)
		},
	}
}
