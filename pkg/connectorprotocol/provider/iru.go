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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/itam/mdm"
)

// iruRegistration wires the Iru (formerly Kandji) device-sync connector.
//
// Iru gives every tenant its own API host, shown in Iru as "Your
// organization's API URL" (acme.api.kandji.io, or acme.api.eu.kandji.io in
// the EU), so that host is the whole of what identifies the tenant and
// Endpoints carries no APIBase. The token is a bearer token whose
// permissions are set per endpoint; the sync needs only "Device list".
//
// The connector syncs devices and registers no access-review driver: Iru's
// API lists the directory users devices are assigned to, but not the admins
// who can sign in to Iru, so it has no roster of access to review.
func iruRegistration() *Registration {
	return &Registration{
		Provider: coredata.ConnectorProviderIru,
		InitialAccountFunc: initialAccount(
			func(s coredata.IruConnectorSettings) string {
				return s.APIHost
			},
		),
		DisplayName: "Iru",
		APIKey: &APIKeyConfig{
			Auth: APIKeyAuth{Mode: APIKeyAuthBearer},
			ExtraSettings: []ExtraSetting{
				{
					Key:         iruAPIURLSetting,
					Label:       "API URL",
					Required:    true,
					Description: "In Iru, open Settings > Access and copy \"Your organization's API URL\", such as acme.api.kandji.io.",
				},
			},
			// No KeyFormat: Iru documents no prefix for its API tokens.
		},
		BuildProbeURL: func(conn *coredata.Connector, _ Endpoints) (string, error) {
			settings, err := coredata.ConnectorSettings[coredata.IruConnectorSettings](conn)
			if err != nil {
				return "", fmt.Errorf("cannot read iru connector settings: %w", err)
			}

			return mdm.IruDevicesProbeURL(settings.APIHost)
		},
		NewDeviceDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger, _ Endpoints) (mdm.Driver, error) {
			settings, err := coredata.ConnectorSettings[coredata.IruConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read iru connector settings: %w", err)
			}

			baseURL, err := mdm.IruBaseURL(settings.APIHost)
			if err != nil {
				return nil, fmt.Errorf("cannot create iru driver: %w", err)
			}

			return mdm.NewIruDriver(c, baseURL), nil
		},
	}
}

// iruAPIURLSetting is the ExtraSetting that carries the tenant's API host.
const iruAPIURLSetting = "apiUrl"
