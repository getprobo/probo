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

package oauth2

import (
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/iam/oauth2scope"
)

func advertisedWriteScopes(registry *oauth2scope.Registry) coredata.OAuth2Scopes {
	if registry == nil {
		return nil
	}

	return authorizationServerScopes(registry.AllWriteScopes())
}

func resolveCIMDAuthorizationScopes(
	client *coredata.OAuth2Client,
	requested coredata.OAuth2Scopes,
	registry *oauth2scope.Registry,
) (coredata.OAuth2Scopes, coredata.OAuth2Scopes) {
	if client == nil {
		return nil, requested
	}

	allowed := client.Scopes
	resolved := requested.OrDefault(allowed)

	if client.ExternalClientID == "" {
		return allowed, resolved
	}

	advertised := advertisedWriteScopes(registry)
	allowed = allowed.Union(advertised)
	resolved = resolved.Union(advertised)

	return allowed, resolved
}
