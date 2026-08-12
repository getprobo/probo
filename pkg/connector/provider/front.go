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
	"net/http"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
)

// frontRegistration wires Front (the shared-inbox platform) as an access
// source. Both credential paths are offered because either is enough on its
// own: an OAuth app grants the `teammates:read` scope the driver needs, and a
// company API token (Settings > API) carries the same read access for customers
// who would rather not install an app.
//
// Front authorizes and exchanges on app.frontapp.com while the Core API lives
// on api2.frontapp.com — a different host, so neither is derived from the other.
// The token exchange authenticates with HTTP Basic client credentials (Front
// documents it as required), which "basic-form" covers with the RFC 6749
// form-encoded body; Front's docs list the body parameters without naming a
// content type.
//
// Scopes are NOT requested per-authorization: Front resolves an OAuth token's
// scopes from the app's own configuration, so an authorize URL carrying a scope
// parameter would be neither honoured nor needed. The operator grants
// `teammates:read` when registering the app.
func frontRegistration() *Registration {
	return &Registration{
		Provider:    coredata.ConnectorProviderFront,
		DisplayName: "Front",
		// No probo.com docs page for Front yet; a 404-ing link is worse than
		// none, so this stays empty until the page ships.
		Endpoints: Endpoints{
			Auth:  "https://app.frontapp.com/oauth/authorize",
			Token: "https://app.frontapp.com/oauth/token",
			// GET /me is the cheapest call that proves the credential reaches
			// the company, and it is the same call the name resolver opens with.
			Probe:   "https://api2.frontapp.com/me",
			APIBase: "https://api2.frontapp.com",
		},
		TokenEndpointAuth: "basic-form",
		SupportsAPIKey:    true,
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger, ep Endpoints) (drivers.Driver, error) {
			return drivers.NewFrontDriver(c, ep.APIBase), nil
		},
		NewNameResolver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger, ep Endpoints) drivers.NameResolver {
			return drivers.NewFrontNameResolver(c, ep.APIBase)
		},
	}
}
