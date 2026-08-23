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
	"fmt"
	"net/url"

	"go.probo.inc/probo/pkg/coredata"
)

func vercelRegistration() *Registration {
	// Vercel's authorization URL embeds the operator's integration slug
	// as a path segment; the operator supplies it via the
	// `integration-slug` config field. BuildAuthURL constructs the URL
	// with net/url so the slug is escaped. Vercel does not use OAuth
	// scopes — capabilities are pinned on the integration registration
	// in the Vercel dashboard.
	return &Registration{
		Provider:    coredata.ConnectorProviderVercel,
		DisplayName: "Vercel",
		Endpoints: Endpoints{
			Token: "https://api.vercel.com/v2/oauth/access_token",
			Probe: "https://api.vercel.com/v2/user",
			// No version segment: the driver reads members from v3 while the
			// name resolver reads the team from v2, so each call site joins
			// its own version onto this origin.
			APIBase: "https://api.vercel.com",
		},
		// See Registration.EndpointOverrideUnsupported: BuildAuthURL builds the
		// authorize URL from an operator slug and FetchVercelUserID resolves the
		// OAuth callback's user from a pinned host — neither reads Endpoints.
		EndpointOverrideUnsupported: "its authorize URL is built from an operator slug and its OAuth callback resolves the user from a pinned host",
		OAuth2: &OAuth2Spec{
			BuildAuthURL: func(slug string) (string, error) {
				u, err := url.JoinPath("https://vercel.com/integrations", url.PathEscape(slug), "new")
				if err != nil {
					return "", fmt.Errorf("cannot build vercel auth URL: %w", err)
				}

				return u, nil
			},
		},
	}
}
