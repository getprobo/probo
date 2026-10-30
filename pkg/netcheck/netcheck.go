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

// Package netcheck provides shared network validation functions to prevent
// SSRF attacks and DNS rebinding across agent tool packages.
package netcheck

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"syscall"
)

// ErrBlockedAddress is returned when an address is refused because it belongs
// to a blocked range.
var ErrBlockedAddress = errors.New("netcheck: address blocked by SSRF protection")

// blockedPrefixes are reserved or otherwise unsafe ranges that netip.Addr's
// helper predicates do not cover. RFC 1918 private, ULA, loopback, link-local,
// multicast, and unspecified are handled by the stdlib helpers in isPublicAddr.
//
// The IPv6 transition prefixes (NAT64, 6to4, Teredo) are blocked wholesale
// rather than decoded: they wrap an arbitrary IPv4 destination — including
// RFC 1918 hosts and cloud metadata endpoints — inside a globally routable
// IPv6 envelope, and agent tools never legitimately need to reach a
// transition-addressed host.
var blockedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),       // RFC 1122 "this network"
	netip.MustParsePrefix("100.64.0.0/10"),   // RFC 6598 CGNAT
	netip.MustParsePrefix("192.0.0.0/24"),    // RFC 6890 IETF protocol assignments
	netip.MustParsePrefix("192.0.2.0/24"),    // RFC 5737 TEST-NET-1
	netip.MustParsePrefix("198.18.0.0/15"),   // RFC 2544 benchmarking
	netip.MustParsePrefix("198.51.100.0/24"), // RFC 5737 TEST-NET-2
	netip.MustParsePrefix("203.0.113.0/24"),  // RFC 5737 TEST-NET-3
	netip.MustParsePrefix("240.0.0.0/4"),     // RFC 1112 reserved (incl. 255.255.255.255)
	netip.MustParsePrefix("::/96"),           // RFC 4291 deprecated IPv4-compatible
	netip.MustParsePrefix("64:ff9b::/96"),    // RFC 6052 NAT64 well-known prefix
	netip.MustParsePrefix("64:ff9b:1::/48"),  // RFC 8215 NAT64 local-use prefix
	netip.MustParsePrefix("100::/64"),        // RFC 6666 discard prefix
	netip.MustParsePrefix("2001::/23"),       // RFC 2928 IETF protocol assignments (incl. 2001::/32 Teredo)
	netip.MustParsePrefix("2001:db8::/32"),   // RFC 3849 documentation
	netip.MustParsePrefix("2002::/16"),       // RFC 3056 6to4
}

// IsPublicIP reports whether ip is a publicly routable address. It returns
// false for loopback, private, link-local, multicast (any range), unspecified,
// IPv6 transition, and other reserved addresses.
func IsPublicIP(ip net.IP) bool {
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return false
	}

	return isPublicAddr(addr)
}

func isPublicAddr(addr netip.Addr) bool {
	if !addr.IsValid() {
		return false
	}

	// Unwrap IPv4-in-IPv6 (::ffff:a.b.c.d) so a single check covers both
	// representations.
	if addr.Is4In6() {
		addr = addr.Unmap()
	}

	if addr.IsUnspecified() ||
		addr.IsLoopback() ||
		addr.IsPrivate() ||
		addr.IsLinkLocalUnicast() ||
		addr.IsLinkLocalMulticast() ||
		addr.IsMulticast() ||
		addr.IsInterfaceLocalMulticast() {
		return false
	}

	for _, prefix := range blockedPrefixes {
		if prefix.Contains(addr) {
			return false
		}
	}

	return true
}

// DialControl is a net.Dialer.Control function that rejects connections to
// non-public peer addresses. The check runs after DNS resolution on the actual
// peer address, which defeats DNS rebinding between ValidatePublicURL and the
// dial. Use it on the raw dialers of tools that do not go through
// go.gearno.de/kit/httpclient.
func DialControl(network, address string, _ syscall.RawConn) error {
	switch network {
	case "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6":
	default:
		return fmt.Errorf("%w: refusing non-IP network %q", ErrBlockedAddress, network)
	}

	addrPort, err := netip.ParseAddrPort(address)
	if err != nil {
		return fmt.Errorf("%w: cannot parse peer address %q: %v", ErrBlockedAddress, address, err)
	}

	if !isPublicAddr(addrPort.Addr()) {
		return fmt.Errorf("%w: %s", ErrBlockedAddress, addrPort.Addr())
	}

	return nil
}

// ValidatePublicURL checks that rawURL uses an http or https scheme and that
// its host resolves exclusively to publicly routable IP addresses. This
// prevents SSRF attacks where the LLM could be tricked into requesting
// internal network endpoints.
func ValidatePublicURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("cannot parse URL: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("unsupported URL scheme %q: only http and https are allowed", u.Scheme)
	}

	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("URL has no host")
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("cannot resolve host %q: %w", host, err)
	}

	for _, ip := range ips {
		if !IsPublicIP(ip) {
			return fmt.Errorf("%w: host %q resolves to non-public IP %s", ErrBlockedAddress, host, ip)
		}
	}

	return nil
}

// ValidatePublicDomain checks that a domain resolves exclusively to publicly
// routable IP addresses.
func ValidatePublicDomain(domain string) error {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return fmt.Errorf("cannot resolve host %q: %w", domain, err)
	}

	for _, ip := range ips {
		if !IsPublicIP(ip) {
			return fmt.Errorf("%w: host %q resolves to non-public IP %s", ErrBlockedAddress, domain, ip)
		}
	}

	return nil
}
