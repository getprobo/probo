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

// Package netcheck provides shared network validation functions to prevent
// SSRF attacks and DNS rebinding across agent tool packages.
package netcheck

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
)

// IsPublicIP reports whether ip is a publicly routable address. It returns
// false for loopback, private, link-local, multicast (any range), and
// unspecified addresses.
func IsPublicIP(ip net.IP) bool {
	if ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsMulticast() ||
		ip.IsUnspecified() {
		return false
	}

	return true
}

// ValidatePublicURL checks that rawURL uses an http or https scheme and that
// its host does not resolve to a private, loopback, or link-local IP address.
// This prevents SSRF attacks where the LLM could be tricked into requesting
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
			return fmt.Errorf("host %q resolves to non-public IP %s", host, ip)
		}
	}

	return nil
}

// ValidatePublicDomain checks that a domain does not resolve to a private,
// loopback, or link-local IP address.
func ValidatePublicDomain(domain string) error {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return fmt.Errorf("cannot resolve host %q: %w", domain, err)
	}

	for _, ip := range ips {
		if !IsPublicIP(ip) {
			return fmt.Errorf("host %q resolves to non-public IP %s", domain, ip)
		}
	}

	return nil
}

// NewPinnedTransport returns an *http.Transport with a custom DialContext that
// resolves the target host once, validates all resolved IPs with IsPublicIP,
// and dials the validated IP directly. This prevents DNS rebinding attacks
// where the first lookup returns a public IP but a subsequent lookup (at
// connection time) returns a private IP.
func NewPinnedTransport() *http.Transport {
	return &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, fmt.Errorf("cannot parse address: %w", err)
			}

			ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			if err != nil {
				return nil, fmt.Errorf("cannot resolve host: %w", err)
			}

			if len(ips) == 0 {
				return nil, fmt.Errorf("cannot resolve host: no addresses found")
			}

			for _, ip := range ips {
				if !IsPublicIP(ip.IP) {
					return nil, fmt.Errorf("cannot connect to non-public IP %s", ip.IP)
				}
			}

			// Dial the first validated IP directly to prevent DNS rebinding.
			pinnedAddr := net.JoinHostPort(ips[0].IP.String(), port)
			var d net.Dialer
			return d.DialContext(ctx, network, pinnedAddr)
		},
	}
}
