// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package domaindns

import (
	"fmt"
	"strings"

	"golang.org/x/net/publicsuffix"
)

// ToFQDN normalizes a DNS name to lowercase FQDN form with a trailing dot.
func ToFQDN(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.TrimSuffix(name, ".")

	if name == "" {
		return "."
	}

	return name + "."
}

// EqualNames reports whether two DNS names refer to the same owner, ignoring
// case and an optional trailing dot.
func EqualNames(a, b string) bool {
	return ToFQDN(a) == ToFQDN(b)
}

// HostnamesForCAA returns the DNS names to evaluate for CAA, starting at the
// exact hostname being verified and walking up through each parent to the
// registrable apex (eTLD+1). The first entry is always the requested hostname
// itself, not its apex.
func HostnamesForCAA(hostname string) ([]string, error) {
	hostname = strings.ToLower(strings.TrimSpace(hostname))
	hostname = strings.TrimSuffix(hostname, ".")
	if hostname == "" {
		return nil, fmt.Errorf("cannot build CAA hostnames: empty hostname")
	}

	apex, err := publicsuffix.EffectiveTLDPlusOne(hostname)
	if err != nil {
		return nil, fmt.Errorf("cannot build CAA hostnames for %q: %w", hostname, err)
	}

	names := []string{hostname}
	current := hostname

	for !strings.EqualFold(current, apex) {
		dot := strings.Index(current, ".")
		if dot < 0 {
			break
		}

		current = current[dot+1:]
		names = append(names, current)
	}

	return names, nil
}
