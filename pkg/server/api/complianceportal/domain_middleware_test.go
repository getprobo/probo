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
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompliancePortalRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                    string
		host                    string
		serverName              string
		forwardedHost           []string
		forwardedProto          []string
		externallyTerminatedTLS bool
		expected                string
		expectedProto           string
		expectedOK              bool
	}{
		{
			name:          "direct TLS uses SNI",
			serverName:    "Portal.Example.com",
			expected:      "portal.example.com",
			expectedProto: "https",
			expectedOK:    true,
		},
		{
			name:                    "direct mode rejects plain HTTP",
			host:                    "portal.example.com",
			forwardedProto:          []string{"https"},
			externallyTerminatedTLS: false,
		},
		{
			name:                    "external mode uses forwarded host",
			host:                    "internal.railway.app",
			forwardedHost:           []string{"Portal.Example.com"},
			forwardedProto:          []string{"https"},
			externallyTerminatedTLS: true,
			expected:                "portal.example.com",
			expectedProto:           "https",
			expectedOK:              true,
		},
		{
			name:                    "external mode falls back to request host",
			host:                    "Portal.Example.com:8080",
			forwardedProto:          []string{"https"},
			externallyTerminatedTLS: true,
			expected:                "portal.example.com",
			expectedProto:           "https",
			expectedOK:              true,
		},
		{
			name:                    "external mode accepts HTTP for HTTPS redirect",
			host:                    "portal.example.com",
			forwardedProto:          []string{"http"},
			externallyTerminatedTLS: true,
			expected:                "portal.example.com",
			expectedProto:           "http",
			expectedOK:              true,
		},
		{
			name:                    "external mode rejects unsupported forwarding protocol",
			host:                    "portal.example.com",
			forwardedProto:          []string{"ftp"},
			externallyTerminatedTLS: true,
		},
		{
			name:                    "external mode rejects forwarded protocol chains",
			host:                    "portal.example.com",
			forwardedProto:          []string{"https, http"},
			externallyTerminatedTLS: true,
		},
		{
			name:                    "external mode rejects multiple forwarded hosts",
			host:                    "portal.example.com",
			forwardedHost:           []string{"portal.example.com", "evil.example.com"},
			forwardedProto:          []string{"https"},
			externallyTerminatedTLS: true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				req := httptest.NewRequest("GET", "http://internal/", nil)
				req.Host = tt.host
				if tt.serverName != "" {
					req.TLS = &tls.ConnectionState{ServerName: tt.serverName}
				}

				for _, value := range tt.forwardedHost {
					req.Header.Add("X-Forwarded-Host", value)
				}

				for _, value := range tt.forwardedProto {
					req.Header.Add("X-Forwarded-Proto", value)
				}

				actual, actualProto, ok := compliancePortalRequest(req, tt.externallyTerminatedTLS)

				assert.Equal(t, tt.expectedOK, ok)
				assert.Equal(t, tt.expected, actual)
				assert.Equal(t, tt.expectedProto, actualProto)
			},
		)
	}
}
