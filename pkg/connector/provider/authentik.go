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

// authentikRegistration wires the authentik access-review connector. authentik
// is self-hosted, so the API host comes from settings and Endpoints stays
// empty. The API key must be a Token whose intent is `api`, not the app
// password a service account is created with.
func authentikRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderAuthentik,
		DisplayName:      "authentik",
		DocumentationURL: accessReviewDocsURL("authentik"),
		APIKey: &APIKeyConfig{
			ExtraSettings: []ExtraSetting{
				{Key: "baseUrl", Label: "Base URL", Required: true},
			},
		},
		BuildProbeURL: buildAuthentikProbeURL,
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger, _ Endpoints) (drivers.Driver, error) {
			settings, err := coredata.ConnectorSettings[coredata.AuthentikConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read authentik connector settings: %w", err)
			}

			baseURL, err := normalizeSelfHostedBaseURL(settings.BaseURL)
			if err != nil {
				return nil, fmt.Errorf("cannot create authentik driver: %w", err)
			}

			return drivers.NewAuthentikDriver(c, baseURL), nil
		},
		NewNameResolver: func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger, _ Endpoints) drivers.NameResolver {
			settings, err := coredata.ConnectorSettings[coredata.AuthentikConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read authentik connector settings", log.Error(err))
				return nil
			}

			baseURL, err := normalizeSelfHostedBaseURL(settings.BaseURL)
			if err != nil {
				logger.ErrorCtx(ctx, "invalid authentik base url in connector settings", log.Error(err))
				return nil
			}

			return drivers.NewAuthentikNameResolver(c, baseURL)
		},
	}
}
