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

func bitbucketRegistration() *Registration {
	// Bitbucket scopes are pinned on the OAuth consumer at registration
	// time (`account` for workspace membership) rather than passed in the
	// authorize URL, so the spec requests none.
	return &Registration{
		Provider:    coredata.ConnectorProviderBitbucket,
		DisplayName: "Bitbucket",
		Endpoints: Endpoints{
			Auth:  "https://bitbucket.org/site/oauth2/authorize",
			Token: "https://bitbucket.org/site/oauth2/access_token",
			Probe: "https://api.bitbucket.org/2.0/user",
			// Every data endpoint the driver calls shares the /2.0 prefix,
			// so the version segment stays in APIBase.
			APIBase: "https://api.bitbucket.org/2.0",
		},
		OAuth2: &OAuth2Spec{},
		SetOrganizationSettings: func(c *coredata.Connector, workspace string) error {
			return c.SetSettings(&coredata.BitbucketConnectorSettings{Workspace: workspace})
		},
	}
}
