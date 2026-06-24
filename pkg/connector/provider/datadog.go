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
