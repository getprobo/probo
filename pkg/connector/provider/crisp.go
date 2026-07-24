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
	"go.probo.inc/probo/pkg/coredata"
)

func crispRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderCrisp,
		DisplayName:      "Crisp",
		DocumentationURL: accessReviewDocsURL("crisp"),
		// Model B: the plugin token is Probo's own Crisp Marketplace plugin
		// credential, held server-side in bootstrap config, not pasted by
		// the customer. ManagedAPIKey injects it at connect time; the
		// customer supplies only the Website ID. SupportsAPIKey stays false
		// so the provider is hidden from the driver catalog until the
		// operator configures PROBOD_CONNECTOR_CRISP_PLUGIN_TOKEN — it ships
		// deactivated until Crisp validates the production plugin and
		// activates with no code change once the token is set.
		ManagedAPIKey: true,
		// The per-website plugin API also needs the plugin ID (a distinct value
		// from the token identifier), supplied via bootstrap alongside the
		// token. Require it so Crisp stays hidden until both are configured
		// rather than surfacing as connectable and failing at connect time.
		RequiresManagedResourceID: true,
		// Crisp authenticates with the plugin token presented as HTTP Basic,
		// the credential being the verbatim "identifier:key" pair.
		// APIKeyBasicAuthUserPass base64-encodes it (the empty-password
		// APIKeyBasicAuth cannot carry the key). A plugin token can serve
		// several websites, so the reviewed website is captured via
		// ExtraSettings. Every request also needs the non-auth X-Crisp-Tier
		// header (set by the driver/probe/name resolver), so the probe is a
		// custom closure.
		APIKeyBasicAuthUserPass: true,
		ExtraSettings: []ExtraSetting{
			{Key: "websiteId", Label: "Website ID", Required: true},
		},
		Probe: probeCrisp,
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.CrispConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read crisp connector settings: %w", err)
			}

			if s.WebsiteID == "" {
				return nil, fmt.Errorf("cannot create crisp driver: website_id is required")
			}

			return drivers.NewCrispDriver(c, s.WebsiteID), nil
		},
		NewNameResolver: func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.CrispConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read crisp connector settings", log.Error(err))

				return nil
			}

			return drivers.NewCrispNameResolver(c, s.WebsiteID)
		},
	}
}
