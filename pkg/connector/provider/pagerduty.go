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

func pagerdutyRegistration() *Registration {
	// PagerDuty Scoped OAuth requires PKCE (RFC 7636). The customer
	// subdomain surfaces as a callback query parameter (or
	// occasionally in the token response body) and is persisted on
	// PagerDutyConnectorSettings by the OAuth callback handler.
	return &Registration{
		Provider:    coredata.ConnectorProviderPagerDuty,
		DisplayName: "PagerDuty",
		Endpoints: Endpoints{
			Auth:  "https://identity.pagerduty.com/oauth/authorize",
			Token: "https://identity.pagerduty.com/oauth/token",
			Probe: "https://api.pagerduty.com/users/me",
			// The REST API lives on api.pagerduty.com, a different host from
			// the identity.pagerduty.com OAuth endpoints above.
			APIBase: "https://api.pagerduty.com",
		},
		OAuth2: &OAuth2Spec{
			Scopes:       []string{"users.read"},
			RequiresPKCE: true,
		},
	}
}
