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

// unifiConsoleIDSetting is the ExtraSetting the connection check reports a
// console the key cannot reach against.
const unifiConsoleIDSetting = "consoleId"

// unifiRegistration wires the UniFi access-review connector. A Site Manager API
// key reaches a console through the api.ui.com cloud connector, the only route
// Probo can take: a console's own address is on the customer's LAN. No
// KeyFormat: Ubiquiti documents no prefix.
func unifiRegistration() *Registration {
	return &Registration{
		Provider: coredata.ConnectorProviderUniFi,
		InitialAccountFunc: initialAccount(
			func(s coredata.UniFiConnectorSettings) string {
				return s.ConsoleID
			},
		),
		DisplayName:      "UniFi",
		DocumentationURL: accessReviewDocsURL("unifi"),
		APIKey: &APIKeyConfig{
			Auth: APIKeyAuth{Mode: APIKeyAuthHeader, Name: "X-API-KEY"},
			ExtraSettings: []ExtraSetting{
				{
					Key:         unifiConsoleIDSetting,
					Label:       "Console ID",
					Required:    true,
					Description: "In UniFi Site Manager, open the console and copy the part of the address between /consoles/ and the next /.",
				},
			},
			CheckSettings:  true,
			CredentialPage: &CredentialPage{URL: "https://unifi.ui.com/settings/api-keys"},
		},
		Endpoints: Endpoints{
			APIBase: "https://api.ui.com",
		},
		Probe: probeUniFi,
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger, ep Endpoints) (drivers.Driver, error) {
			settings, err := coredata.ConnectorSettings[coredata.UniFiConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read unifi connector settings: %w", err)
			}

			networkURL, err := drivers.UniFiNetworkURL(ep.APIBase, settings.ConsoleID)
			if err != nil {
				return nil, fmt.Errorf("cannot create unifi driver: %w", err)
			}

			return drivers.NewUniFiDriver(c, networkURL), nil
		},
		NewNameResolver: func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger, ep Endpoints) drivers.NameResolver {
			settings, err := coredata.ConnectorSettings[coredata.UniFiConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read unifi connector settings", log.Error(err))
				return nil
			}

			hostURL, err := drivers.UniFiHostURL(ep.APIBase, settings.ConsoleID)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot build unifi host URL", log.Error(err))
				return nil
			}

			return drivers.NewUniFiNameResolver(c, hostURL)
		},
	}
}
