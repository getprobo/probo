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

// attioRegistration wires Attio. Every Attio credential is bound to exactly one
// workspace, so there is nothing to pick and nothing to capture (Pattern 3).
//
// The probe is the members endpoint, not the /v2/self introspection endpoint:
// /v2/self answers 200 with {"active": false} for a revoked token and 400 for a
// malformed one, both of which read as connected.
func attioRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderAttio,
		DisplayName:      "Attio",
		DocumentationURL: accessReviewDocsURL("attio"),
		Endpoints: Endpoints{
			Auth:    "https://app.attio.com/authorize",
			Token:   "https://app.attio.com/oauth/token",
			Probe:   "https://api.attio.com/v2/workspace_members",
			APIBase: "https://api.attio.com/v2",
		},
		OAuth2: &OAuth2Config{
			// The scope Attio's OpenAPI requires for the members endpoint.
			// Attio takes an app's real grant from its configuration at
			// build.attio.com rather than from the authorize request, which
			// accepts this parameter without acting on it, so the value
			// records what the app must be configured with.
			Scopes: []string{"user_management:read"},
		},
		// A workspace-scoped key presented as a bearer token: the default auth
		// mode applies and no settings are needed. No KeyFormat, because Attio
		// documents no prefix for it.
		APIKey: &APIKeyConfig{},
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger, ep Endpoints) (drivers.Driver, error) {
			return drivers.NewAttioDriver(c, ep.APIBase), nil
		},
		NewNameResolver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger, ep Endpoints) drivers.NameResolver {
			return drivers.NewAttioNameResolver(c, ep.APIBase)
		},
	}
}
