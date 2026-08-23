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

func apolloRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderApollo,
		DisplayName:      "Apollo.io",
		DocumentationURL: accessReviewDocsURL("apollo"),
		// Apollo's REST API authenticates with a master API key in the
		// x-api-key header; it rejects Authorization: Bearer (and, since
		// Sept 2024, query/body key params). APIKeyCustomHeader makes the
		// APIKeyConnection send x-api-key instead of Bearer. There is no
		// OAuth2 flow needed: the customer supplies a master key, which is
		// bound to one Apollo account, so there is nothing to pick
		// (Pattern 3): no settings struct, no picker.
		Endpoints: Endpoints{
			APIBase: "https://api.apollo.io/api/v1",
			// ProbeURL lets the connection-status check confirm the key with a
			// lightweight GET; the transport attaches x-api-key, and a missing,
			// dead, or non-master key returns 401/403.
			Probe: "https://api.apollo.io/api/v1/users/search?page=1&per_page=1",
		},
		//
		// No name resolver: Apollo exposes no stable account-name
		// endpoint reachable with the master key, so the source keeps its
		// generic name.
		APIKey: &APIKeySpec{
			Presentation: APIKeyCustomHeader,
			Name:         "x-api-key",
		},
	}
}
