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

func squareRegistration() *Registration {
	return &Registration{
		Provider:    coredata.ConnectorProviderSquare,
		DisplayName: "Square",
		Endpoints: Endpoints{
			Auth:  "https://connect.squareup.com/oauth2/authorize",
			Token: "https://connect.squareup.com/oauth2/token",
			// Every data endpoint the driver calls shares the /v2 prefix, so
			// the version segment stays in APIBase.
			APIBase: "https://connect.squareup.com/v2",
		},
		// EMPLOYEES_READ lists team members; MERCHANT_PROFILE_READ is needed
		// for the merchant-name resolver and the /v2/merchants/me probe.
		// Square's confidential token endpoint accepts client credentials in
		// the form body (the default post-form scheme) and rejects HTTP Basic,
		// so no TokenEndpointAuth override is set.
		OAuth2: &OAuth2Config{
			Scopes: []string{"EMPLOYEES_READ", "MERCHANT_PROFILE_READ"},
		},
		Probe: probeSquare,
		// SupportsAPIKey enables the Personal Access Token fallback, which
		// authenticates with the same Authorization: Bearer scheme as the OAuth
		// token. A Square token — OAuth or PAT — is always scoped to one
		// merchant, so there is nothing to pick (Pattern 3): no settings
		// struct, no picker, no OAuth-callback capture.
		APIKey: &APIKeyConfig{},
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger, ep Endpoints) (drivers.Driver, error) {
			return drivers.NewSquareDriver(c, ep.APIBase), nil
		},
		NewNameResolver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger, ep Endpoints) drivers.NameResolver {
			return drivers.NewSquareNameResolver(c, ep.APIBase)
		},
	}
}
