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

func segmentRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderSegment,
		DisplayName:      "Segment",
		DocumentationURL: accessReviewDocsURL("segment"),
		APIKey: &APIKeySpec{
			ExtraSettings: []ExtraSetting{
				{Key: "region", Label: "Region", Required: true},
			},
		},
		// Segment authenticates with a Public API token as the default
		// Authorization: Bearer scheme, so no APIKeyHeader. The token is bound
		// to one workspace, but the workspace's region selects the API host
		// (US vs EU) and is not discoverable from the token, so it is captured
		// as an extra setting and resolved to a base URL (Pattern 3 + region);
		// there is nothing to pick.
		BuildProbeURL: buildSegmentProbeURL,
	}
}
