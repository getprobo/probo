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

// frontRegistration wires Front. Every credential is bound to one company, so
// there is nothing to pick and nothing to capture (Pattern 3).
//
// The probe is the teammates endpoint, not /me: /me needs no scope and answers
// 200 for a credential that cannot list teammates.
func frontRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderFront,
		DisplayName:      "Front",
		DocumentationURL: accessReviewDocsURL("front"),
		Endpoints: Endpoints{
			Auth:    "https://app.frontapp.com/oauth/authorize",
			Token:   "https://app.frontapp.com/oauth/token",
			Probe:   "https://api2.frontapp.com/teammates",
			APIBase: "https://api2.frontapp.com",
		},
		OAuth2: &OAuth2Config{
			// Front takes the grant (teammates:read) from the app's
			// configuration; the authorize request has no scope parameter.
			TokenEndpointAuth: "basic-form",
		},
		// A company API token presented as a bearer token. No KeyFormat:
		// Front documents no prefix.
		APIKey: &APIKeyConfig{
			CredentialPage: &CredentialPage{URL: "https://app.frontapp.com/settings/developers/tokens/new"},
		},
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger, ep Endpoints) (drivers.Driver, error) {
			return drivers.NewFrontDriver(c, ep.APIBase), nil
		},
		NewNameResolver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger, ep Endpoints) drivers.NameResolver {
			return drivers.NewFrontNameResolver(c, ep.APIBase)
		},
	}
}
