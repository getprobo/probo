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

package console_v1

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/server/api/console/v1/types"
)

// The Segment region is the only API-key setting resolved to a derived value
// rather than stored verbatim, so the region -> host mapping is pinned here:
// a typo in either host would otherwise only surface as a live 404.
func TestApiKeyConnectorSettings_SegmentRegion(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		region  string
		baseURL string
	}{
		{name: "US", region: "US", baseURL: "https://api.segmentapis.com"},
		{name: "EU", region: "EU", baseURL: "https://eu1.api.segmentapis.com"},
		{name: "lowercase", region: "eu", baseURL: "https://eu1.api.segmentapis.com"},
		{name: "padded", region: "  us  ", baseURL: "https://api.segmentapis.com"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			raw, err := apiKeyConnectorSettings(types.CreateAPIKeyConnectorInput{
				Provider:      coredata.ConnectorProviderSegment,
				SegmentRegion: &tc.region,
			})
			require.NoError(t, err)

			var settings coredata.SegmentConnectorSettings
			require.NoError(t, json.Unmarshal(raw, &settings))
			assert.Equal(t, tc.baseURL, settings.BaseURL)
		})
	}
}

func TestApiKeyConnectorSettings_SegmentRejectsUnknownRegion(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"empty", "whitespace", "apac", "eu1", "host"} {
		region := map[string]string{
			"empty":      "",
			"whitespace": "   ",
			"apac":       "APAC",
			// The region Segment's own UI shows for the EU workspace; it must
			// not silently fall through to a wrong host.
			"eu1":  "EU1",
			"host": "https://eu1.api.segmentapis.com",
		}[name]

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := apiKeyConnectorSettings(types.CreateAPIKeyConnectorInput{
				Provider:      coredata.ConnectorProviderSegment,
				SegmentRegion: &region,
			})
			require.Error(t, err)
		})
	}
}

func TestApiKeyConnectorSettings_SegmentRequiresRegion(t *testing.T) {
	t.Parallel()

	_, err := apiKeyConnectorSettings(types.CreateAPIKeyConnectorInput{
		Provider: coredata.ConnectorProviderSegment,
	})
	require.Error(t, err)
}
