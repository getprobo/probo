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

package connector

// CallbackPath is the HTTP path for the OAuth2 callback endpoint.
const CallbackPath = "/api/console/v1/connectors/complete"

// CIMDMetadataPath is the HTTP path serving the public OAuth Client ID
// Metadata Document (CIMD). For public clients, the deployment's
// (baseURL + CIMDMetadataPath) URL IS the OAuth client_id: the provider
// (e.g. PostHog) fetches this document server-to-server during the
// authorization flow to learn the client's name and redirect URIs, so no
// app pre-registration is required. The endpoint must be reachable
// unauthenticated from the public internet.
const CIMDMetadataPath = "/api/console/v1/connectors/oauth-client-metadata"
