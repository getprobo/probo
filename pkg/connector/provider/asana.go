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
	"go.probo.inc/probo/pkg/coredata"
)

func asanaRegistration() *Registration {
	return &Registration{
		Provider:    coredata.ConnectorProviderAsana,
		DisplayName: "Asana",
		Endpoints: Endpoints{
			Auth:  "https://app.asana.com/-/oauth_authorize",
			Token: "https://app.asana.com/-/oauth_token",
			Probe: "https://app.asana.com/api/1.0/users/me",
			// Every data endpoint the driver calls shares the /api/1.0
			// prefix, so the version segment stays in APIBase.
			APIBase: "https://app.asana.com/api/1.0",
		},
		OAuth2: &OAuth2Spec{
			Scopes:          []string{"default"},
			ExclusiveScopes: true,
		},
		// Asana publishes no granular scope for workspace_memberships, the
		// only endpoint exposing admin/guest/view-only status, and answers a
		// granular token with 403 "Full permissions are required to use this
		// endpoint". Requesting "default" is the documented opt-out, and
		// naming it keeps SourceNeedsReconnect able to spot the connectors
		// still holding a narrower grant.
		// Asana rejects the whole authorize request when it carries a scope
		// the app no longer offers, so a reconnect must not replay the
		// granular grant these connectors were created with.
		SetOrganizationSettings: func(c *coredata.Connector, workspaceGID string) error {
			return c.SetSettings(&coredata.AsanaConnectorSettings{WorkspaceGID: workspaceGID})
		},
	}
}
