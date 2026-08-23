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

func langfuseRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderLangfuse,
		DisplayName:      "Langfuse",
		DocumentationURL: accessReviewDocsURL("langfuse"),
		//
		// No name resolver: the memberships endpoint carries no
		// organization name, so the source keeps its generic name.
		APIKey: &APIKeySpec{
			Presentation: APIKeyBasicUserPass,
			ExtraSettings: []ExtraSetting{
				{Key: "baseUrl", Label: "Base URL", Required: true},
			},
		},
		// Langfuse's organization-scoped public API authenticates with HTTP
		// Basic auth where the credential is publicKey:secretKey.
		// APIKeyBasicUserPass base64s the verbatim "publicKey:secretKey" the
		// operator pastes (the empty-password APIKeyBasic cannot carry
		// the secret). The org API key is bound to one organization, so
		// there is nothing to pick; only the regional/self-hosted base URL
		// is per-tenant and is surfaced as an extra setting.
		// BuildProbeURL derives the probe endpoint from the per-connection
		// base URL (the host is regional/self-hosted, so a static ProbeURL
		// cannot express it); the transport attaches the Basic credential
		// and a dead key returns 401/403.
		BuildProbeURL: buildLangfuseProbeURL,
	}
}
