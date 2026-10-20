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
	"net/url"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/connectorprotocol"
	"go.probo.inc/probo/pkg/coredata"
)

// ovhcloudRegistration is OVHcloud in the EU region.
//
// Two connect paths, both single-tenant: an OVHcloud credential is bound to
// exactly one account, so there is no organisation to pick. US is a
// separate connector; each region runs its own identity service.
func ovhcloudRegistration() *Registration {
	return ovhcloudRegistrationFor(
		coredata.ConnectorProviderOVHcloud,
		"OVHcloud",
		connectorprotocol.OVHcloudRegionEU,
	)
}

func ovhcloudRegistrationFor(provider coredata.ConnectorProvider, displayName, region string) *Registration {
	return &Registration{
		Provider:         provider,
		DisplayName:      displayName,
		DocumentationURL: accessReviewDocsURL("ovhcloud"),
		Endpoints:        ovhcloudEndpoints(region),
		OAuth2: &OAuth2Config{
			// "account/all" is the narrowest scope reaching every route the
			// driver reads; the bare "all" spans every product the customer
			// owns. There is no read-only level: "account/read" is rejected as
			// invalid_request.
			Scopes:       []string{"account/all"},
			RequiresPKCE: true,
		},
		ClientCredentials: &ClientCredentialsConfig{},
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger, ep Endpoints) (drivers.Driver, error) {
			return drivers.NewOVHcloudDriver(c, ep.APIBase), nil
		},
		NewNameResolver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger, ep Endpoints) drivers.NameResolver {
			return drivers.NewOVHcloudNameResolver(c, ep.APIBase)
		},
	}
}

func ovhcloudEndpoints(region string) Endpoints {
	authURL, err := connectorprotocol.OVHcloudAuthorizeURL(region)
	if err != nil {
		panic(err)
	}

	tokenURL, err := connectorprotocol.OVHcloudTokenURL(region)
	if err != nil {
		panic(err)
	}

	apiBase, err := connectorprotocol.OVHcloudAPIBase(region)
	if err != nil {
		panic(err)
	}

	probe, err := url.JoinPath(apiBase, "me", "identity", "user")
	if err != nil {
		panic(err)
	}

	return Endpoints{
		Auth:    authURL,
		Token:   tokenURL,
		Probe:   probe,
		APIBase: apiBase,
	}
}
