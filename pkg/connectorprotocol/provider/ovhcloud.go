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

// ovhcloudRegistration is OVHcloud, EU region.
//
// Two connect paths, both single-tenant: an OVHcloud credential is bound to
// exactly one account, so there is no organisation to pick and no settings
// struct. OVHcloud's OAuth2 client namespace is global, so one Probo-side app
// serves every customer on the authorization-code path.
//
// EU only: OAuth2 clients are region-local and Endpoints.Token is one pinned
// value. Another region needs its own Probo account and client. The commit body
// carries the rest of the reasoning.
func ovhcloudRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderOVHcloud,
		DisplayName:      "OVHcloud",
		DocumentationURL: accessReviewDocsURL("ovhcloud"),
		Endpoints: Endpoints{
			Auth:  "https://www.ovh.com/auth/oauth2/authorize",
			Token: "https://www.ovh.com/auth/oauth2/token",
			// Every route the driver and resolver touch hangs off the
			// versioned root, so the version segment stays in APIBase.
			APIBase: "https://eu.api.ovh.com/1.0",
			// The roster endpoint itself, so the probe exercises the exact
			// dependency: 401 on a dead credential, 403 on a missing IAM action.
			Probe: "https://eu.api.ovh.com/1.0/me/identity/user",
		},
		OAuth2: &OAuth2Config{
			// "account/all" is the narrowest scope reaching every route the
			// driver reads; the bare "all" spans every product the customer
			// owns. There is no read-only level: "account/read" is rejected as
			// invalid_request.
			Scopes:       []string{"account/all"},
			RequiresPKCE: true,
		},
		// No ExtraSettings: the credential already identifies the account.
		ClientCredentials: &ClientCredentialsConfig{},
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger, ep Endpoints) (drivers.Driver, error) {
			return drivers.NewOVHcloudDriver(c, ep.APIBase), nil
		},
		NewNameResolver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger, ep Endpoints) drivers.NameResolver {
			return drivers.NewOVHcloudNameResolver(c, ep.APIBase)
		},
	}
}
