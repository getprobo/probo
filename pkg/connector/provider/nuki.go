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

func nukiRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderNuki,
		DisplayName:      "Nuki",
		DocumentationURL: accessReviewDocsURL("nuki"),
		SupportsAPIKey:   true,
		// A Nuki Web API token (Nuki Web > Menu > API, scopes `account` and
		// `smartlock.auth`) authenticates as a standard Bearer token, so the
		// default APIKeyConnection mode applies; no Header/Scheme/BasicAuth
		// override is needed. Nuki also offers OAuth2, but only with an
		// implicit flow, which yields no refreshable credential for a
		// background fetch — the API token is the supported path.
		//
		// The token is bound to one Nuki Web account and /account/user returns
		// every account user of it, so there is nothing to pick (Pattern 3): no
		// settings struct, no picker.
		Endpoints: Endpoints{
			// Both the driver and the name resolver join their paths onto this
			// root, so an override moves the whole connector.
			APIBase: "https://api.nuki.io",
			// The connection check confirms the token with the same lightweight
			// GET the driver uses; a token that is dead or missing a scope
			// returns 401. Same host as APIBase, as Register requires.
			Probe: "https://api.nuki.io/account/user?limit=1",
		},
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger, ep Endpoints) (drivers.Driver, error) {
			return drivers.NewNukiDriver(c, ep.APIBase), nil
		},
		// GET /account names the source after the Nuki Web account itself.
		NewNameResolver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger, ep Endpoints) drivers.NameResolver {
			return drivers.NewNukiNameResolver(c, ep.APIBase)
		},
	}
}
