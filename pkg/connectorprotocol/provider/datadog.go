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
	"regexp"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/connectorprotocol"
	"go.probo.inc/probo/pkg/coredata"
)

// datadogKeyPattern accepts the two token classes that authenticate alone as a
// Bearer: service access tokens and personal access tokens. It turns away an
// application key, which works only paired with an API key.
var datadogKeyPattern = regexp.MustCompile(`^dd[ps]at_[\s\S]`)

func datadogRegistration() *Registration {
	// Datadog is multi-site: the customer's region drives the authorize
	// host (built at initiate from the region pick) and the token + API
	// host (built at callback from Datadog's `domain` param). AuthURL and
	// TokenURL are empty — the closures build the per-customer hosts.
	// BuildProbeURL targets the stored API domain. Confidential client + PKCE map
	// to the default post-form token-endpoint auth.
	return &Registration{
		Provider: coredata.ConnectorProviderDatadog,
		InitialAccountFunc: initialAccount(
			func(s coredata.DatadogConnectorSettings) string {
				return s.Domain
			},
		),
		DisplayName:      "Datadog",
		DocumentationURL: accessReviewDocsURL("datadog"),
		OAuth2: &OAuth2Config{
			Scopes:                 []string{"user_access_read"},
			RequiresPKCE:           true,
			BuildAuthURLForSite:    connectorprotocol.DatadogAuthorizeURL,
			BuildTokenURLForDomain: connectorprotocol.DatadogTokenURL,
		},
		// A token does not reveal its site, so the customer picks it; the
		// settings resolver maps it to the same domain OAuth stores.
		APIKey: &APIKeyConfig{
			KeyFormat: &KeyFormat{
				Pattern: datadogKeyPattern,
				Example: "ddsat_…",
			},
			ExtraSettings: []ExtraSetting{
				{
					Key:         "site",
					Label:       "Site",
					Required:    true,
					Description: "The site key, such as US1 or EU1, matching the address you sign in at: app.datadoghq.com is US1, app.datadoghq.eu is EU1.",
				},
			},
		},
		BuildProbeURL: buildDatadogProbeURL,
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger, _ Endpoints) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.DatadogConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read datadog connector settings: %w", err)
			}

			// Re-validate the stored domain against the fixed allow-list at
			// the construction site (defense-in-depth). Both writers
			// validate on write, but pinning the SSRF invariant here keeps
			// the driver safe regardless of how the connector row was
			// populated. An empty domain also fails this check.
			if !connectorprotocol.IsValidDatadogDomain(s.Domain) {
				return nil, fmt.Errorf("cannot create datadog driver: invalid or missing domain")
			}

			return drivers.NewDatadogDriver(c, s.Domain), nil
		},
		NewNameResolver: func(ctx context.Context, _ *http.Client, conn *coredata.Connector, logger *log.Logger, _ Endpoints) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.DatadogConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read datadog connector settings", log.Error(err))
				return nil
			}

			return drivers.NewDatadogNameResolver(s.Region)
		},
	}
}
