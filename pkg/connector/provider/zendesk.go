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
