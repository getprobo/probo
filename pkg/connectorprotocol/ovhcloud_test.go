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

package connectorprotocol_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/connectorprotocol"
)

func TestOVHcloudRegionURLs(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		region    string
		wantAuth  string
		wantToken string
		wantAPI   string
	}{
		{"eu", "https://www.ovh.com/auth/oauth2/authorize", "https://www.ovh.com/auth/oauth2/token", "https://eu.api.ovh.com/1.0"},
		{"US", "https://us.ovhcloud.com/auth/oauth2/authorize", "https://us.ovhcloud.com/auth/oauth2/token", "https://api.us.ovhcloud.com/1.0"},
	} {
		t.Run(tc.region, func(t *testing.T) {
			t.Parallel()

			authURL, err := connectorprotocol.OVHcloudAuthorizeURL(tc.region)
			require.NoError(t, err)
			assert.Equal(t, tc.wantAuth, authURL)

			tokenURL, err := connectorprotocol.OVHcloudTokenURL(tc.region)
			require.NoError(t, err)
			assert.Equal(t, tc.wantToken, tokenURL)

			apiBase, err := connectorprotocol.OVHcloudAPIBase(tc.region)
			require.NoError(t, err)
			assert.Equal(t, tc.wantAPI, apiBase)
		})
	}

	for _, region := range []string{"", "  ", "AP", "eu.evil.example", "US/"} {
		_, err := connectorprotocol.OVHcloudAPIBase(region)
		assert.Error(t, err, region)
	}
}
