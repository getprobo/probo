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

package cookiebanner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/coredata"
)

func TestBrowserFamily(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		userAgent string
		want      coredata.DiscoveryBrowserFamily
	}{
		{
			name:      "chrome",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			want:      coredata.DiscoveryBrowserFamilyChrome,
		},
		{
			name:      "chrome ios",
			userAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/120.0.6099.119 Mobile/15E148 Safari/604.1",
			want:      coredata.DiscoveryBrowserFamilyChrome,
		},
		{
			name:      "edge",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
			want:      coredata.DiscoveryBrowserFamilyEdge,
		},
		{
			name:      "edge ios",
			userAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) EdgiOS/120.0.2210.105 Version/17.0 Mobile/15E148 Safari/604.1",
			want:      coredata.DiscoveryBrowserFamilyEdge,
		},
		{
			name:      "edge android",
			userAgent: "Mozilla/5.0 (Linux; Android 10; HD1913) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.6099.144 Mobile Safari/537.36 EdgA/120.0.2210.141",
			want:      coredata.DiscoveryBrowserFamilyEdge,
		},
		{
			name:      "firefox",
			userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:120.0) Gecko/20100101 Firefox/120.0",
			want:      coredata.DiscoveryBrowserFamilyFirefox,
		},
		{
			name:      "firefox ios",
			userAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) FxiOS/120.0 Mobile/15E148 Safari/605.1.15",
			want:      coredata.DiscoveryBrowserFamilyFirefox,
		},
		{
			name:      "safari",
			userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
			want:      coredata.DiscoveryBrowserFamilySafari,
		},
		{
			name:      "empty",
			userAgent: "",
			want:      coredata.DiscoveryBrowserFamilyOther,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, BrowserFamily(tt.userAgent))
		})
	}
}
