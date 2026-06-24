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

package browser

import (
	"go.probo.inc/probo/pkg/agent/tools/internal/netcheck"
)

// validatePublicURL checks that a URL uses an http(s) scheme and that its
// host does not resolve to a private, loopback, or link-local IP address.
// This prevents SSRF attacks where the LLM could be tricked into requesting
// internal network endpoints.
func validatePublicURL(rawURL string) error {
	return netcheck.ValidatePublicURL(rawURL)
}

// validatePublicDomain checks that a domain does not resolve to a private,
// loopback, or link-local IP address.
func validatePublicDomain(domain string) error {
	return netcheck.ValidatePublicDomain(domain)
}
