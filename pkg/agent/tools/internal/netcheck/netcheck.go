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
	"fmt"
	"net"
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
