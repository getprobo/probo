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
	"fmt"
	"net/url"

	"go.probo.inc/probo/pkg/coredata"
)

func metabaseRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderMetabase,
		DisplayName:      "Metabase",
		DocumentationURL: accessReviewDocsURL("metabase"),
		APIKey: &APIKeySpec{
			Presentation: APIKeyCustomHeader,
			Name:         "x-api-key",
			ExtraSettings: []ExtraSetting{
				{Key: "instanceUrl", Label: "Instance URL", Required: true},
			},
		},
		BuildProbeURL: buildMetabaseProbeURL,
	}
}

// ValidateMetabaseInstanceURL rejects a stored instance URL that no request
// could target. Metabase is self-hosted, so the host comes from connector
// settings rather than from Endpoints.
func ValidateMetabaseInstanceURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("cannot create metabase driver: instance_url is invalid: %w", err)
	}

	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return fmt.Errorf("cannot create metabase driver: instance_url must be an http(s) URL")
	}

	return nil
}
