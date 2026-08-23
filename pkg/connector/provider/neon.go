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

func neonRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderNeon,
		DisplayName:      "Neon",
		DocumentationURL: accessReviewDocsURL("neon"),
		Endpoints: Endpoints{
			// Every endpoint the driver calls lives under the same /api/v2
			// prefix, so the version segment stays in APIBase.
			APIBase: "https://console.neon.tech/api/v2",
		},
		APIKey: &APIKeySpec{
			ExtraSettings: []ExtraSetting{
				{Key: "organizationId", Label: "Organization ID", Required: true},
			},
		},
		// Neon's API authenticates with an API key (napi_...) presented
		// as Authorization: Bearer, the default APIKeyConnection scheme.
		// Neon's OAuth is partner-gated (manual application), so the
		// connector is API-key only. A personal or organization API key
		// can belong to several organizations; the operator supplies the
		// org ID (org-...) of the one to review.
		//
		BuildProbeURL: buildNeonProbeURL,
	}
}
