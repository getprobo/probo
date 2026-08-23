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

func sentryRegistration() *Registration {
	return &Registration{
		Provider:    coredata.ConnectorProviderSentry,
		DisplayName: "Sentry",
		Endpoints: Endpoints{
			Auth:  "https://sentry.io/oauth/authorize/",
			Token: "https://sentry.io/oauth/token/",
			Probe: "https://sentry.io/api/0/organizations/",
			// Every data endpoint the driver calls shares the /api/0
			// prefix, so the version segment stays in APIBase. The trailing
			// slash Sentry requires belongs to the path elements the driver
			// joins on, never to APIBase.
			APIBase: "https://sentry.io/api/0",
		},
		OAuth2: &OAuth2Spec{
			Scopes: []string{"org:read", "member:read"},
		},
		APIKey: &APIKeySpec{
			ExtraSettings: []ExtraSetting{
				{Key: "organizationSlug", Label: "Organization Slug", Required: true},
			},
		},
		SetOrganizationSettings: func(c *coredata.Connector, slug string) error {
			return c.SetSettings(&coredata.SentryConnectorSettings{OrganizationSlug: slug})
		},
	}
}
