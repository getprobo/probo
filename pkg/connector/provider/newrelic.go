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

// newRelicRegistration wires the New Relic access-review connector.
//
// New Relic authenticates NerdGraph with a user key in its own API-Key header.
// The key belongs to one organization, so there is nothing to pick (Pattern
// 3) — but it also belongs to one of three data REGIONS, and the others answer
// it with 403 "not authorized for account region". The region
// is not discoverable from the credential, so the customer names it and every
// host is resolved from it; Endpoints therefore carries no APIBase.
func newRelicRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderNewRelic,
		DisplayName:      "New Relic",
		DocumentationURL: accessReviewDocsURL("new-relic"),
		APIKey: &APIKeyConfig{
			Auth: APIKeyAuth{Mode: APIKeyAuthHeader, Name: "API-Key"},
			ExtraSettings: []ExtraSetting{
				{Key: "region", Label: "Region (us, eu or jp)", Required: true},
			},
			// No KeyFormat: New Relic's current user keys carry an NRAK-
			// prefix, but the REST keys it migrated from did not and the
			// docs still tell holders of those they need no update. A shape
			// asserted here would lock those customers out.
		},
		Probe: probeNewRelic,
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger, _ Endpoints) (drivers.Driver, error) {
			endpoint, err := newRelicEndpoint(conn)
			if err != nil {
				return nil, fmt.Errorf("cannot create new relic driver: %w", err)
			}

			return drivers.NewNewRelicDriver(c, endpoint), nil
		},
		NewNameResolver: func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger, _ Endpoints) drivers.NameResolver {
			endpoint, err := newRelicEndpoint(conn)
			if err != nil {
				logger.ErrorCtx(ctx, "invalid new relic region in connector settings", log.Error(err))
				return nil
			}

			return drivers.NewNewRelicNameResolver(c, endpoint)
		},
	}
}

// newRelicEndpoint resolves the NerdGraph endpoint a connector's region names.
func newRelicEndpoint(conn *coredata.Connector) (string, error) {
	settings, err := coredata.ConnectorSettings[coredata.NewRelicConnectorSettings](conn)
	if err != nil {
		return "", fmt.Errorf("cannot read new relic connector settings: %w", err)
	}

	return drivers.NewRelicEndpoint(settings.Region)
}
