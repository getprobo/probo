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

// oktaRegistration wires the Okta access-review connector. Okta is a
// per-tenant IdP with no central API gateway, so a one-click OAuth flow is not
// possible — it authenticates with a read-only API token presented under the
// `SSWS` Authorization scheme (APIKeyAuthScheme), plus the customer's org
// domain. The token + domain identify exactly one org, so there is no picker
// and no OAuth metadata. BuildProbeURL targets the org's own API host.
func oktaRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderOkta,
		DisplayName:      "Okta",
		DocumentationURL: accessReviewDocsURL("okta"),
		APIKey: &APIKeySpec{
			Presentation: APIKeyCustomScheme,
			Name:         "SSWS",
			ExtraSettings: []ExtraSetting{
				{Key: "domain", Label: "Okta Domain", Required: true},
			},
		},
		BuildProbeURL: buildOktaProbeURL,
	}
}
