// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package trustedproxy_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/server/trustedproxy"
)

func runMiddleware(t *testing.T, trusted []string, remoteAddr string, headers map[string]string) *http.Request {
	t.Helper()

	middleware, err := trustedproxy.NewMiddleware(trusted)
	require.NoError(t, err)

	var captured *http.Request
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = r
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = remoteAddr
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	handler.ServeHTTP(httptest.NewRecorder(), req)

	require.NotNil(t, captured)
	return captured
}

func TestNewMiddleware_HeaderHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		trusted          []string
		remoteAddr       string
		expectPreserved  bool
		forwardedHeaders map[string]string
	}{
		{
			name:            "untrusted proxy strips forwarded headers",
			trusted:         []string{"10.0.0.1"},
			remoteAddr:      "192.168.1.1:1234",
			expectPreserved: false,
		},
		{
			name:            "trusted proxy preserves forwarded headers",
			trusted:         []string{"10.0.0.1"},
			remoteAddr:      "10.0.0.1:1234",
			expectPreserved: true,
		},
		{
			name:            "empty trusted list strips all forwarded headers",
			trusted:         nil,
			remoteAddr:      "10.0.0.1:1234",
			expectPreserved: false,
		},
		{
			name:            "multiple trusted IPs",
			trusted:         []string{"10.0.0.1", "10.0.0.2"},
			remoteAddr:      "10.0.0.2:5678",
			expectPreserved: true,
		},
		{
			name:            "CIDR range trusts addresses within the range",
			trusted:         []string{"10.0.0.0/24"},
			remoteAddr:      "10.0.0.50:1234",
			expectPreserved: true,
		},
		{
			name:            "CIDR range strips addresses outside the range",
			trusted:         []string{"10.0.0.0/24"},
			remoteAddr:      "10.0.1.50:1234",
			expectPreserved: false,
		},
		{
			name:            "mixed IP and CIDR list trusts plain IP",
			trusted:         []string{"192.168.1.1", "10.0.0.0/24"},
			remoteAddr:      "192.168.1.1:1234",
			expectPreserved: true,
		},
		{
			name:            "mixed IP and CIDR list trusts address in CIDR",
			trusted:         []string{"192.168.1.1", "10.0.0.0/24"},
			remoteAddr:      "10.0.0.99:1234",
			expectPreserved: true,
		},
		{
			name:            "IPv6 CIDR range trusts addresses within the range",
			trusted:         []string{"fd00::/8"},
			remoteAddr:      "[fd12:3456::1]:1234",
			expectPreserved: true,
		},
		{
			name:            "IPv6 CIDR range strips addresses outside the range",
			trusted:         []string{"fd00::/8"},
			remoteAddr:      "[2001:db8::1]:1234",
			expectPreserved: false,
		},
	}

	const (
		xff = "203.0.113.50"
		fwd = "for=198.51.100.17"
	)

	for _, tc := range tests {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				captured := runMiddleware(
					t,
					tc.trusted,
					tc.remoteAddr,
					map[string]string{
						"X-Forwarded-For": xff,
						"Forwarded":       fwd,
					},
				)

				if tc.expectPreserved {
					assert.Equal(t, xff, captured.Header.Get("X-Forwarded-For"))
					assert.Equal(t, fwd, captured.Header.Get("Forwarded"))
				} else {
					assert.Empty(t, captured.Header.Get("X-Forwarded-For"))
					assert.Empty(t, captured.Header.Get("Forwarded"))
				}
			},
		)
	}
}

func TestNewMiddleware_InvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		trusted []string
	}{
		{
			name:    "invalid IP",
			trusted: []string{"not-an-ip"},
		},
		{
			name:    "invalid CIDR mask",
			trusted: []string{"10.0.0.0/99"},
		},
	}

	for _, tc := range tests {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				_, err := trustedproxy.NewMiddleware(tc.trusted)
				require.Error(t, err)
			},
		)
	}
}
