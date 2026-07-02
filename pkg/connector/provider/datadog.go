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

func datadogRegistration() *Registration {
	// Datadog is multi-site: the customer's region drives the authorize
	// host (built at initiate from the region pick) and the token + API
	// host (built at callback from Datadog's `domain` param). AuthURL and
	// TokenURL are empty — the closures build the per-customer hosts.
	// BuildProbeURL targets the stored API domain. Confidential client + PKCE map
	// to the default post-form token-endpoint auth.
	return &Registration{
		Provider:               coredata.ConnectorProviderDatadog,
		DisplayName:            "Datadog",
		OAuth2Scopes:           []string{"user_access_read"},
		RequiresPKCE:           true,
		BuildAuthURLForSite:    connector.DatadogAuthorizeURL,
		BuildTokenURLForDomain: connector.DatadogTokenURL,
		BuildProbeURL:          buildDatadogProbeURL,
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.DatadogConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read datadog connector settings: %w", err)
			}

			// Re-validate the stored domain against the fixed allow-list at
			// the construction site (defense-in-depth). The OAuth callback
			// validates on write, but pinning the SSRF invariant here keeps
			// the driver safe regardless of how the connector row was
			// populated. An empty domain also fails this check.
			if !connector.IsValidDatadogDomain(s.Domain) {
				return nil, fmt.Errorf("cannot create datadog driver: invalid or missing domain")
			}

			return drivers.NewDatadogDriver(c, s.Domain), nil
		},
		NewNameResolver: func(ctx context.Context, _ *http.Client, conn *coredata.Connector, logger *log.Logger) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.DatadogConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read datadog connector settings", log.Error(err))
				return nil
			}

			return drivers.NewDatadogNameResolver(s.Region)
		},
	}
}
