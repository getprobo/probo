// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
