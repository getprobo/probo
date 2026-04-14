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
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/server/trustedproxy"
)

func newRequest(remoteAddr string, headers map[string]string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = remoteAddr
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	return r
}

func TestNewMiddleware(t *testing.T) {
	t.Parallel()

	t.Run(
		"strips forwarded headers from untrusted proxy",
		func(t *testing.T) {
			t.Parallel()

			trusted := []net.IP{net.ParseIP("10.0.0.1")}
			middleware := trustedproxy.NewMiddleware(trusted)

			var captured *http.Request
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				captured = r
			}))

			r := newRequest("192.168.1.1:1234", map[string]string{
				"X-Forwarded-For": "203.0.113.50",
				"Forwarded":       "for=198.51.100.17",
			})
			handler.ServeHTTP(httptest.NewRecorder(), r)

			require.NotNil(t, captured)
			assert.Empty(t, captured.Header.Get("X-Forwarded-For"))
			assert.Empty(t, captured.Header.Get("Forwarded"))
		},
	)

	t.Run(
		"preserves forwarded headers from trusted proxy",
		func(t *testing.T) {
			t.Parallel()

			trusted := []net.IP{net.ParseIP("10.0.0.1")}
			middleware := trustedproxy.NewMiddleware(trusted)

			var captured *http.Request
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				captured = r
			}))

			r := newRequest("10.0.0.1:1234", map[string]string{
				"X-Forwarded-For": "203.0.113.50",
				"Forwarded":       "for=198.51.100.17",
			})
			handler.ServeHTTP(httptest.NewRecorder(), r)

			require.NotNil(t, captured)
			assert.Equal(t, "203.0.113.50", captured.Header.Get("X-Forwarded-For"))
			assert.Equal(t, "for=198.51.100.17", captured.Header.Get("Forwarded"))
		},
	)

	t.Run(
		"empty trusted list strips all forwarded headers",
		func(t *testing.T) {
			t.Parallel()

			middleware := trustedproxy.NewMiddleware(nil)

			var captured *http.Request
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				captured = r
			}))

			r := newRequest("10.0.0.1:1234", map[string]string{
				"X-Forwarded-For": "203.0.113.50",
			})
			handler.ServeHTTP(httptest.NewRecorder(), r)

			require.NotNil(t, captured)
			assert.Empty(t, captured.Header.Get("X-Forwarded-For"))
		},
	)

	t.Run(
		"multiple trusted proxies",
		func(t *testing.T) {
			t.Parallel()

			trusted := []net.IP{
				net.ParseIP("10.0.0.1"),
				net.ParseIP("10.0.0.2"),
			}
			middleware := trustedproxy.NewMiddleware(trusted)

			var captured *http.Request
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				captured = r
			}))

			r := newRequest("10.0.0.2:5678", map[string]string{
				"X-Forwarded-For": "203.0.113.50",
			})
			handler.ServeHTTP(httptest.NewRecorder(), r)

			require.NotNil(t, captured)
			assert.Equal(t, "203.0.113.50", captured.Header.Get("X-Forwarded-For"))
		},
	)
}
