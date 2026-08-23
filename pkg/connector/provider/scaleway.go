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

func scalewayRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderScaleway,
		DisplayName:      "Scaleway",
		DocumentationURL: accessReviewDocsURL("scaleway"),
		// Scaleway authenticates with the secret key in the X-Auth-Token header
		// rather than Authorization: Bearer. APIKeyCustomHeader makes the
		// APIKeyConnection send that header and omit Authorization. The key is
		// bound to one Organization, but GET /iam/v1alpha1/users requires the
		// organization_id explicitly, so it is captured via APIKey.ExtraSettings
		// rather than discovered — hence no picker and a BuildProbeURL.
		Endpoints: Endpoints{
			// Every endpoint the driver calls lives under the same
			// /iam/v1alpha1 prefix, so the version segment stays in APIBase.
			APIBase: "https://api.scaleway.com/iam/v1alpha1",
		},
		// No name resolver: Scaleway exposes no read-only endpoint that maps
		// an Organization UUID to its display name, so the source keeps its
		// generic name.
		APIKey: &APIKeySpec{
			Presentation: APIKeyCustomHeader,
			Name:         "X-Auth-Token",
			ExtraSettings: []ExtraSetting{
				{Key: "organizationId", Label: "Organization ID", Required: true},
			},
		},
		BuildProbeURL: buildScalewayProbeURL,
	}
}
