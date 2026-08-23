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

// renderRegistration wires Render as an API-key access-review connector.
// Render exposes no partner OAuth program, so the customer supplies a
// read-scoped API key plus their Workspace ID (Render's owner ID). The key
// authenticates with the default Authorization: Bearer scheme, so no
// APIKeyAuthScheme override is set. There is no picker — the workspace is
// captured up front via APIKey.ExtraSettings — so SetOrganizationSettings is
// omitted.
func renderRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderRender,
		DisplayName:      "Render",
		DocumentationURL: accessReviewDocsURL("render"),
		Endpoints: Endpoints{
			// Every endpoint the driver calls lives under the same /v1 prefix,
			// so the version segment stays in APIBase.
			APIBase: "https://api.render.com/v1",
		},
		APIKey: &APIKeySpec{
			ExtraSettings: []ExtraSetting{
				{Key: "workspaceId", Label: "Workspace ID", Required: true},
			},
		},
		BuildProbeURL: buildRenderProbeURL,
	}
}
