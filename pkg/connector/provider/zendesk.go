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
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
)

func zendeskRegistration() *Registration {
	// Zendesk is multi-tenant via per-customer subdomain
	// (<subdomain>.zendesk.com). The subdomain is collected at initiate (the
	// customer types it; it drives the authorize host) and rides the signed
	// OAuth state to the callback, where it builds the token host and is
	// persisted on the connector settings for the driver's API host. Unlike
	// Datadog, Zendesk does NOT echo a host back on the callback, so
	// BuildTokenURLForSite reads the subdomain from the state rather than a
	// query param. AuthURL and TokenURL are empty: the closures build the
	// per-customer hosts. BuildProbeURL targets the stored subdomain. The global confidential
	// client carries a client_secret, which both authenticates the token
	// exchange (default post-form) and signs the state.
	return &Registration{
		Provider:             coredata.ConnectorProviderZendesk,
		DisplayName:          "Zendesk",
		OAuth2Scopes:         []string{"users:read"},
		BuildAuthURLForSite:  connector.ZendeskAuthorizeURL,
		BuildTokenURLForSite: connector.ZendeskTokenURL,
		BuildProbeURL:        buildZendeskProbeURL,
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.ZendeskConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read zendesk connector settings: %w", err)
			}

			// Re-validate the stored subdomain at the construction site
			// (defense-in-depth). The OAuth callback validates on write, but
			// pinning the SSRF invariant here keeps the driver safe regardless
			// of how the connector row was populated. An empty subdomain also
			// fails this check.
			if !connector.IsValidZendeskSubdomain(s.Subdomain) {
				return nil, fmt.Errorf("cannot create zendesk driver: invalid or missing subdomain")
			}

			return drivers.NewZendeskDriver(c, s.Subdomain), nil
		},
		NewNameResolver: func(ctx context.Context, _ *http.Client, conn *coredata.Connector, logger *log.Logger) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.ZendeskConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read zendesk connector settings", log.Error(err))
				return nil
			}

			return drivers.NewZendeskNameResolver(s.Subdomain)
		},
	}
}
