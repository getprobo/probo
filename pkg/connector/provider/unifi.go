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

// unifiRegistration wires UniFi Network as an API-key access-review connector.
//
// Ubiquiti runs no partner OAuth program for the Network API, so the customer
// supplies a console API key (UniFi Site Manager > API) plus the ID of the
// console to review. The key authenticates in X-API-KEY rather than as a Bearer
// token, hence APIKeyHeader.
//
// APIBase is the Site Manager console-proxy root: its host is a compile-time
// constant and only the {consoleId} segment varies per connector, so the
// driver, the name resolver and the probe all compose from ep.APIBase (via
// drivers.UniFiConsoleBaseURL) and an override moves all three together.
//
// Reaching a console directly on its LAN address is deliberately not offered:
// that host resolves to a private IP, which the SSRF-protected connector client
// refuses to dial. The cloud proxy is the only route a hosted deployment can
// take.
func unifiRegistration() *Registration {
	return &Registration{
		Provider:    coredata.ConnectorProviderUniFi,
		DisplayName: "UniFi",
		// DocumentationURL is left empty until the probo.com docs page exists;
		// the console renders it as a link, so a slug that 404s is worse than
		// no link.
		SupportsAPIKey: true,
		APIKeyHeader:   "X-API-KEY",
		BuildProbeURL:  buildUniFiProbeURL,
		Endpoints: Endpoints{
			APIBase: "https://api.ui.com/v1/connector/consoles",
		},
		APIKeyExtraSettings: []ExtraSetting{
			{Key: "consoleId", Label: "Console ID", Required: true},
		},
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger, ep Endpoints) (drivers.Driver, error) {
			baseURL, err := unifiConsoleBaseURL(conn, ep)
			if err != nil {
				return nil, fmt.Errorf("cannot create unifi driver: %w", err)
			}

			return drivers.NewUniFiDriver(c, logger.Named("unifi"), baseURL), nil
		},
		// A single-site console is named after its site; a multi-site console
		// keeps the generic name (see drivers.NewUniFiNameResolver).
		NewNameResolver: func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger, ep Endpoints) drivers.NameResolver {
			baseURL, err := unifiConsoleBaseURL(conn, ep)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot resolve unifi console base url", log.Error(err))

				return nil
			}

			return drivers.NewUniFiNameResolver(c, logger.Named("unifi"), baseURL)
		},
	}
}

// unifiConsoleBaseURL reads the connector's console ID and composes the
// console's Network Integration API root under ep.APIBase.
func unifiConsoleBaseURL(conn *coredata.Connector, ep Endpoints) (string, error) {
	settings, err := coredata.ConnectorSettings[coredata.UniFiConnectorSettings](conn)
	if err != nil {
		return "", fmt.Errorf("cannot read unifi connector settings: %w", err)
	}

	return drivers.UniFiConsoleBaseURL(ep.APIBase, settings.ConsoleID)
}
