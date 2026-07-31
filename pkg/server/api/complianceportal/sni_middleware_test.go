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

package complianceportal

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPortalHostFromRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		host                 string
		serverName           string
		tls                  bool
		tlsTerminatedByProxy bool
		want                 portalHost
		wantOK               bool
	}{
		{
			name:       "tls request resolves the page by sni",
			host:       "trust.example.com:8443",
			serverName: "trust.example.com",
			tls:        true,
			want:       portalHost{serverName: "trust.example.com", current: "trust.example.com:8443"},
			wantOK:     true,
		},
		{
			name:   "plain http request carries no portal host",
			host:   "trust.example.com",
			wantOK: false,
		},
		{
			name:                 "plain http request behind a tls terminating proxy resolves the page by host",
			host:                 "trust.example.com",
			tlsTerminatedByProxy: true,
			want:                 portalHost{serverName: "trust.example.com", current: "trust.example.com"},
			wantOK:               true,
		},
		{
			name:                 "port is stripped from the host header",
			host:                 "trust.example.com:8080",
			tlsTerminatedByProxy: true,
			want:                 portalHost{serverName: "trust.example.com", current: "trust.example.com"},
			wantOK:               true,
		},
		{
			name:                 "tls still wins over the host header",
			host:                 "trust.example.com",
			serverName:           "sni.example.com",
			tls:                  true,
			tlsTerminatedByProxy: true,
			want:                 portalHost{serverName: "sni.example.com", current: "trust.example.com"},
			wantOK:               true,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				t.Parallel()

				r := httptest.NewRequest(http.MethodGet, "/", nil)
				r.Host = test.host

				if test.tls {
					r.TLS = &tls.ConnectionState{ServerName: test.serverName}
				}

				host, ok := portalHostFromRequest(r, test.tlsTerminatedByProxy)

				assert.Equal(t, test.wantOK, ok)
				assert.Equal(t, test.want, host)
			},
		)
	}
}
