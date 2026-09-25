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

package netcheck_test

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/agent/tools/internal/netcheck"
)

func TestIsPublicIP(t *testing.T) {
	t.Parallel()

	cases := []struct {
		ip     string
		public bool
	}{
		// IPv4: not public.
		{"0.0.0.0", false},
		{"0.1.2.3", false},
		{"10.0.0.1", false},
		{"10.255.255.255", false},
		{"100.64.0.1", false},
		{"100.127.255.255", false},
		{"127.0.0.1", false},
		{"127.255.255.254", false},
		{"169.254.169.254", false},
		{"172.16.0.1", false},
		{"172.31.255.254", false},
		{"192.0.0.1", false},
		{"192.0.2.1", false},
		{"192.168.1.1", false},
		{"198.18.0.1", false},
		{"198.19.255.254", false},
		{"198.51.100.1", false},
		{"203.0.113.1", false},
		{"224.0.0.1", false},
		{"239.255.255.255", false},
		{"240.0.0.1", false},
		{"255.255.255.255", false},

		// IPv4: public.
		{"1.1.1.1", true},
		{"8.8.8.8", true},
		{"9.255.255.255", true},
		{"100.63.255.255", true},
		{"100.128.0.0", true},
		{"172.15.255.255", true},
		{"172.32.0.0", true},
		{"192.167.255.255", true},
		{"192.169.0.0", true},
		{"198.17.255.255", true},
		{"198.20.0.0", true},

		// IPv6: not public.
		{"::", false},
		{"::1", false},
		{"fc00::1", false},
		{"fd00::1", false},
		{"fe80::1", false},
		{"ff00::1", false},
		{"ff01::1", false},
		{"ff02::1", false},
		{"2001:db8::1", false},
		{"100::1", false},

		// IPv4-mapped IPv6.
		{"::ffff:127.0.0.1", false},
		{"::ffff:10.0.0.1", false},
		{"::ffff:169.254.169.254", false},

		// Deprecated IPv4-compatible IPv6.
		{"::10.0.0.1", false},
		{"::169.254.169.254", false},

		// NAT64 (RFC 6052 / RFC 8215): wraps an arbitrary IPv4 target.
		{"64:ff9b::169.254.169.254", false},
		{"64:ff9b::10.0.0.1", false},
		{"64:ff9b::1.1.1.1", false},
		{"64:ff9b:1::10.0.0.1", false},

		// 6to4 (RFC 3056): 2002:<ipv4>::/48.
		{"2002:c0a8:0101::1", false}, // 192.168.1.1
		{"2002:a9fe:a9fe::1", false}, // 169.254.169.254
		{"2002:0a00:0001::", false},  // 10.0.0.1

		// Teredo (RFC 4380).
		{"2001:0:4136:e378:8000:63bf:5765:5765", false},
		{"2001::1", false},

		// IPv6: public.
		{"2606:4700:4700::1111", true},
		{"2001:4860:4860::8888", true},
	}

	for _, c := range cases {
		t.Run(c.ip, func(t *testing.T) {
			t.Parallel()

			ip := net.ParseIP(c.ip)
			require.NotNil(t, ip)
			assert.Equal(t, c.public, netcheck.IsPublicIP(ip))
		})
	}
}

func TestIsPublicIP_InvalidAddress(t *testing.T) {
	t.Parallel()

	assert.False(t, netcheck.IsPublicIP(nil))
	assert.False(t, netcheck.IsPublicIP(net.IP{1, 2, 3}))
}

func TestDialControl(t *testing.T) {
	t.Parallel()

	cases := []struct {
		network string
		address string
		blocked bool
	}{
		{"tcp", "1.1.1.1:443", false},
		{"tcp", "[2606:4700:4700::1111]:443", false},
		{"tcp", "127.0.0.1:443", true},
		{"tcp", "169.254.169.254:80", true},
		{"tcp", "[::ffff:10.0.0.1]:443", true},
		{"tcp", "[64:ff9b::169.254.169.254]:80", true},
		{"tcp", "[2002:c0a8:0101::1]:80", true},
		{"tcp", "not-an-address", true},
		{"unix", "/var/run/docker.sock", true},
	}

	for _, c := range cases {
		t.Run(c.network+" "+c.address, func(t *testing.T) {
			t.Parallel()

			err := netcheck.DialControl(c.network, c.address, nil)
			if !c.blocked {
				assert.NoError(t, err)

				return
			}

			require.Error(t, err)
			assert.ErrorIs(t, err, netcheck.ErrBlockedAddress)
		})
	}
}

func TestValidatePublicURL_Scheme(t *testing.T) {
	t.Parallel()

	for _, scheme := range []string{"file", "gopher", "ftp", "data", "javascript"} {
		t.Run(scheme, func(t *testing.T) {
			t.Parallel()

			err := netcheck.ValidatePublicURL(scheme + "://example.com/x")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "unsupported URL scheme")
		})
	}
}

func TestValidatePublicURL_LiteralIPHosts(t *testing.T) {
	t.Parallel()

	// Literal IP hosts do not need a resolver, so these assert the guard
	// itself rather than DNS behaviour.
	cases := []string{
		"http://127.0.0.1/latest/meta-data/",
		"http://169.254.169.254/latest/meta-data/",
		"http://[64:ff9b::169.254.169.254]/latest/meta-data/",
		"http://[2002:c0a8:0101::1]/admin",
		"http://[2001:0:4136:e378:8000:63bf:5765:5765]/",
		"http://[::ffff:10.0.0.1]/",
	}

	for _, rawURL := range cases {
		t.Run(rawURL, func(t *testing.T) {
			t.Parallel()

			err := netcheck.ValidatePublicURL(rawURL)
			require.Error(t, err)
			assert.ErrorIs(t, err, netcheck.ErrBlockedAddress)
		})
	}
}

func TestValidatePublicURL_NoHost(t *testing.T) {
	t.Parallel()

	err := netcheck.ValidatePublicURL("http:///path")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "URL has no host")
}
