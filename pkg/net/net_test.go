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

package net_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/net"
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

				assert.Equal(t, tt.want, net.IsLoopback(tt.host))
			},
		)
	}
}
