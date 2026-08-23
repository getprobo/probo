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

func openrouterRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderOpenRouter,
		DisplayName:      "OpenRouter",
		DocumentationURL: accessReviewDocsURL("openrouter"),
		Endpoints: Endpoints{
			// Every endpoint the driver calls shares the /api/v1 prefix, so
			// the version segment stays in APIBase.
			APIBase: "https://openrouter.ai/api/v1",
		},
		//
		// No name resolver: the members endpoint carries no organization
		// name, so the source keeps its generic name.
		APIKey: &APIKeySpec{},
		// OpenRouter authenticates with an organization management
		// (provisioning) API key presented as Authorization: Bearer, the
		// default APIKeyConnection scheme. The key is bound to one
		// organization, so there is nothing to pick (Pattern 3): no settings
		// struct, no picker.
		//
		// Probe confirms the key with a lightweight GET against the members
		// endpoint; the transport attaches the Bearer token. probeOpenRouter
		// rejects 401/403 (revoked/invalid key) and also 404, which a valid
		// but personal (non-organization) key returns — so a key that cannot
		// list members shows as not-connected rather than failing later.
		Probe: ProbeOver(probeOpenRouter),
	}
}
