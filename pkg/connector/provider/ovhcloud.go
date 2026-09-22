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
// struct.
//
//   - OAuth2 authorization code is the preferred path. OVHcloud's OAuth2 client
//     namespace is global — a client registered in Probo's account resolves
//     against any customer's session — so one Probo-side app serves every
//     customer. Verified live: an authorize request carrying a client owned by
//     a different account advances past client resolution to redirect_uri
//     validation, where an unregistered client stops at invalid_client.
//     Confirmed structurally too: AUTHORIZATION_CODE clients come back with
//     identity=null, while CLIENT_CREDENTIALS clients are bound to an identity
//     in the creating account.
//   - Client credentials covers customers who would rather hold their own
//     credential. OVHcloud exposes no way to list or revoke a third party's
//     OAuth2 grant, so this path — where revoking means deleting the service
//     account — is the answer for customers who need that.
//
// EU only for now: OAuth2 clients are region-local (an EU client is rejected by
// the CA and US token endpoints), and Endpoints.Token is a single pinned value
// shared by both connect paths. Serving ca.api.ovh.com / api.us.ovhcloud.com
// needs its own Probo account and its own client per region.
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
			// dependency. It answers 401 for a dead credential and 403 when
			// the IAM policy is missing the action — which is the
			// OperationRefused case doProbeRequest already reports.
			Probe: "https://eu.api.ovh.com/1.0/me/identity/user",
		},
		OAuth2: &OAuth2Config{
			// OVHcloud scopes are <product>/<permission>. "account/all" is the
			// narrowest one that reaches every route the driver reads: /me,
			// /me/identity/*, /me/logs/audit and /me/api/oauth2/client all
			// answer 200 under it, verified live. The bare "all" also works but
			// spans every product the customer owns, and the consent screen
			// says so — "Manage your whole account and all your services"
			// against "Manage your account" for this one.
			//
			// There is no read-only alternative: "account/read" is rejected as
			// invalid_request, so OVHcloud has no read permission level at the
			// scope layer. Read-only access comes from the IAM policy the
			// customer attaches, not from the grant.
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
