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

func linearSyncRegistration() *Registration {
	return &Registration{
		Provider:    coredata.ConnectorProviderLinearSync,
		DisplayName: "Linear Sync",
		Endpoints: Endpoints{
			Auth:  "https://linear.app/oauth/authorize",
			Token: "https://api.linear.app/oauth/token",
			// Same GraphQL host as LINEAR. Task sync follows this
			// registration's APIBase so a LINEAR_SYNC override moves
			// issue write calls with the OAuth app that mints the token.
			APIBase: "https://api.linear.app/graphql",
		},
		Probe: probeLinear,
		OAuth2: &OAuth2Config{
			Scopes: []string{"read", "write", "issues:create"},
			// Linear's authorize endpoint requires a comma-separated scope
			// list. A space-separated string is treated as a single unknown
			// scope and Linear falls back to read.
			ScopeSeparator: ",",
			ExtraAuthParams: map[string]string{
				"actor": "app",
			},
			SupportsIncrementalAuth: true,
		},
	}
}
