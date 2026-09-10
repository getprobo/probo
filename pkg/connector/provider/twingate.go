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

// twingateRegistration wires the Twingate access-review connector.
//
// Twingate authenticates with an API token in its own X-API-KEY header and
// gives every tenant its own host, <network>.twingate.com, so the network name
// is the whole of what identifies the tenant. It is the customer's to supply
// and every host is built from it; Endpoints therefore carries no APIBase.
//
// A read-only token is enough. Twingate rejects a bad token with 401, so the
// probe adds only the per-network host, not a new verdict.
func twingateRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderTwingate,
		DisplayName:      "Twingate",
		DocumentationURL: accessReviewDocsURL("twingate"),
		APIKey: &APIKeyConfig{
			Auth: APIKeyAuth{Mode: APIKeyAuthHeader, Name: "X-API-KEY"},
			ExtraSettings: []ExtraSetting{
				{Key: "network", Label: "Network name", Required: true},
			},
			// No KeyFormat: a Twingate API token is an opaque string with no
			// documented prefix, so there is no shape to assert.
		},
		Probe: probeTwingate,
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger, _ Endpoints) (drivers.Driver, error) {
			endpoint, err := twingateEndpoint(conn)
			if err != nil {
				return nil, fmt.Errorf("cannot create twingate driver: %w", err)
			}

			return drivers.NewTwingateDriver(c, endpoint), nil
		},
		NewNameResolver: func(ctx context.Context, _ *http.Client, conn *coredata.Connector, logger *log.Logger, _ Endpoints) drivers.NameResolver {
			settings, err := coredata.ConnectorSettings[coredata.TwingateConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read twingate connector settings", log.Error(err))
				return nil
			}

			return drivers.NewTwingateNameResolver(settings.Network)
		},
	}
}

// twingateEndpoint resolves the GraphQL endpoint a connector's network names.
func twingateEndpoint(conn *coredata.Connector) (string, error) {
	settings, err := coredata.ConnectorSettings[coredata.TwingateConnectorSettings](conn)
	if err != nil {
		return "", fmt.Errorf("cannot read twingate connector settings: %w", err)
	}

	return drivers.TwingateEndpoint(settings.Network)
}
