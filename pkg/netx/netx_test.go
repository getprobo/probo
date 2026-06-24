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

package netx_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/netx"
)

func TestIsLoopback(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		host string
		want bool
	}{
		// Localhost by name.
		{"localhost", "localhost", true},

		// IPv4 loopback addresses (127.0.0.0/8).
		{"ipv4 canonical loopback", "127.0.0.1", true},
		{"ipv4 loopback high octet", "127.255.255.255", true},
		{"ipv4 loopback alternate", "127.0.0.2", true},
		{"ipv4 loopback 127.1.2.3", "127.1.2.3", true},

		// IPv6 loopback.
		{"ipv6 loopback", "::1", true},

		// IPv4-mapped IPv6 loopback.
		{"ipv4-mapped ipv6 loopback", "::ffff:127.0.0.1", true},
		{"ipv4-mapped ipv6 loopback alternate", "::ffff:127.0.0.2", true},

		// Non-loopback addresses.
		{"ipv4 private 10.x", "10.0.0.1", false},
		{"ipv4 private 192.168.x", "192.168.1.1", false},
		{"ipv4 private 172.16.x", "172.16.0.1", false},
		{"ipv4 public", "8.8.8.8", false},
		{"ipv4 all interfaces", "0.0.0.0", false},
		{"ipv6 all interfaces", "::", false},
		{"ipv6 link-local", "fe80::1", false},
		{"ipv6 public", "2001:db8::1", false},
		{"ipv4-mapped ipv6 non-loopback", "::ffff:192.168.1.1", false},

		// Hostnames that are not localhost.
		{"example.com", "example.com", false},
		{"localhost.localdomain", "localhost.localdomain", false},
		{"myhost", "myhost", false},

		// Edge cases.
		{"empty string", "", false},
		{"whitespace", " ", false},
		{"localhost with trailing dot", "localhost.", false},
		{"uppercase LOCALHOST", "LOCALHOST", false},
		{"mixed case Localhost", "Localhost", false},
		{"128.0.0.1 not loopback", "128.0.0.1", false},
		{"126.255.255.255 not loopback", "126.255.255.255", false},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tt.want, netx.IsLoopback(tt.host))
			},
		)
	}
}
