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

func grafanaRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderGrafana,
		DisplayName:      "Grafana",
		DocumentationURL: accessReviewDocsURL("grafana"),
		SupportsAPIKey:   true,
		BuildProbeURL:    buildGrafanaProbeURL,
		APIKeyExtraSettings: []ExtraSetting{
			{Key: "baseUrl", Label: "Base URL", Required: true},
		},
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger, _ Endpoints) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.GrafanaConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read grafana connector settings: %w", err)
			}

			baseURL, err := normalizeSelfHostedBaseURL(s.BaseURL)
			if err != nil {
				return nil, fmt.Errorf("cannot create grafana driver: %w", err)
			}

			return drivers.NewGrafanaDriver(c, baseURL), nil
		},
		NewNameResolver: func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger, _ Endpoints) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.GrafanaConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read grafana connector settings", log.Error(err))
				return nil
			}

			baseURL, err := normalizeSelfHostedBaseURL(s.BaseURL)
			if err != nil {
				logger.ErrorCtx(ctx, "invalid grafana base url in connector settings", log.Error(err))
				return nil
			}

			return drivers.NewGrafanaNameResolver(c, baseURL)
		},
	}
}
