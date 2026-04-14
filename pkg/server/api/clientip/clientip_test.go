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

package clientip_test

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/server/api/clientip"
)

func TestExtract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		want       string
	}{
		{
			name:       "remote addr only",
			remoteAddr: "192.168.1.1:12345",
			want:       "192.168.1.1",
		},
		{
			name:       "remote addr without port",
			remoteAddr: "192.168.1.1",
			want:       "192.168.1.1",
		},
		{
			name:       "x-forwarded-for single ip",
			remoteAddr: "10.0.0.1:1234",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.50"},
			want:       "203.0.113.50",
		},
		{
			name:       "x-forwarded-for chain",
			remoteAddr: "10.0.0.1:1234",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.50, 70.41.3.18, 150.172.238.178"},
			want:       "203.0.113.50",
		},
		{
			name:       "x-forwarded-for with port",
			remoteAddr: "10.0.0.1:1234",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.50:8080"},
			want:       "203.0.113.50",
		},
		{
			name:       "forwarded header simple",
			remoteAddr: "10.0.0.1:1234",
			headers:    map[string]string{"Forwarded": "for=198.51.100.17"},
			want:       "198.51.100.17",
		},
		{
			name:       "forwarded header quoted",
			remoteAddr: "10.0.0.1:1234",
			headers:    map[string]string{"Forwarded": `for="198.51.100.17"`},
			want:       "198.51.100.17",
		},
		{
			name:       "forwarded header ipv6 bracketed",
			remoteAddr: "10.0.0.1:1234",
			headers:    map[string]string{"Forwarded": `for="[2001:db8::1]"`},
			want:       "2001:db8::1",
		},
		{
			name:       "forwarded header with port",
			remoteAddr: "10.0.0.1:1234",
			headers:    map[string]string{"Forwarded": `for="198.51.100.17:4711"`},
			want:       "198.51.100.17",
		},
		{
			name:       "forwarded header chain",
			remoteAddr: "10.0.0.1:1234",
			headers:    map[string]string{"Forwarded": "for=198.51.100.17, for=70.41.3.18"},
			want:       "198.51.100.17",
		},
		{
			name:       "forwarded takes precedence over x-forwarded-for",
			remoteAddr: "10.0.0.1:1234",
			headers: map[string]string{
				"Forwarded":       "for=198.51.100.17",
				"X-Forwarded-For": "203.0.113.50",
			},
			want: "198.51.100.17",
		},
		{
			name:       "forwarded with extra directives",
			remoteAddr: "10.0.0.1:1234",
			headers:    map[string]string{"Forwarded": "for=198.51.100.17;proto=https;by=203.0.113.60"},
			want:       "198.51.100.17",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				r := httptest.NewRequest("GET", "/", nil)
				r.RemoteAddr = tt.remoteAddr
				for k, v := range tt.headers {
					r.Header.Set(k, v)
				}

				assert.Equal(t, tt.want, clientip.Extract(r))
			},
		)
	}
}
