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

func clerkRegistration() *Registration {
	return &Registration{
		Provider:       coredata.ConnectorProviderClerk,
		DisplayName:    "Clerk",
		SupportsAPIKey: true,
		// Clerk's Backend API authenticates with a server-side secret key
		// (sk_...) presented as Authorization: Bearer, the default
		// APIKeyConnection scheme. There is no third-party OAuth2 flow for
		// account-listing: Clerk's OAuth is an end-user IdP (scoped consent
		// to a single user's profile), not a partner grant over the Backend
		// API. The secret key is bound to one Clerk instance, so there is
		// nothing to pick (Pattern 3): no settings struct, no picker, no
		// SetOrganizationSettings.
		ProbeURL: "https://api.clerk.com/v1/users?limit=1",
		// No NewNameResolver: the Backend API exposes no instance/application
		// name endpoint reachable with a secret key, so the source keeps its
		// generic name (the source-name worker degrades gracefully).
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			return drivers.NewClerkDriver(c), nil
		},
	}
}
