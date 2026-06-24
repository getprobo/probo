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

package thirdparty

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveOwnedDomains(t *testing.T) {
	t.Parallel()

	const threshold = 0.85

	domainsOf := func(owned []ownedDomain) []string {
		out := make([]string, len(owned))
		for i, d := range owned {
			out[i] = d.Domain
		}

		return out
	}

	t.Run(
		"keeps website and brand-related domains, drops unrelated and generic",
		func(t *testing.T) {
			t.Parallel()

			got := resolveOwnedDomains(
				"Dark Reader",
				"https://darkreader.org",
				DomainsResult{
					Domains: []DomainCandidate{
						{Domain: "darkreader.org", Confidence: 0.95},
						{Domain: "darkreader.ltd", Confidence: 0.9},
						{Domain: "https://api.darkreader.org/v1", Confidence: 0.9},
						{Domain: "cloudflare.com", Confidence: 0.9},
						{Domain: "googleapis.com", Confidence: 0.95},
					},
				},
				threshold,
			)

			assert.Equal(t, []string{"darkreader.org", "darkreader.ltd"}, domainsOf(got))
		},
	)

	t.Run(
		"always includes the website domain even when absent from candidates",
		func(t *testing.T) {
			t.Parallel()

			got := resolveOwnedDomains(
				"Dark Reader",
				"https://darkreader.org",
				DomainsResult{},
				threshold,
			)

			assert.Equal(t, []string{"darkreader.org"}, domainsOf(got))
			assert.Equal(t, 1.0, got[0].Confidence)
			assert.Equal(t, "https://darkreader.org", got[0].SourceURL)
		},
	)

	t.Run(
		"drops candidates below the confidence threshold",
		func(t *testing.T) {
			t.Parallel()

			got := resolveOwnedDomains(
				"Dark Reader",
				"https://darkreader.org",
				DomainsResult{
					Domains: []DomainCandidate{
						{Domain: "darkreader.io", Confidence: 0.5},
					},
				},
				threshold,
			)

			assert.Equal(t, []string{"darkreader.org"}, domainsOf(got))
		},
	)

	t.Run(
		"keeps a separate brand CDN domain",
		func(t *testing.T) {
			t.Parallel()

			got := resolveOwnedDomains(
				"Intercom",
				"https://intercom.com",
				DomainsResult{
					Domains: []DomainCandidate{
						{Domain: "intercomcdn.com", Confidence: 0.9},
					},
				},
				threshold,
			)

			assert.Equal(t, []string{"intercom.com", "intercomcdn.com"}, domainsOf(got))
		},
	)

	t.Run(
		"keeps the vendor's own shared-infrastructure domain when the vendor is that provider",
		func(t *testing.T) {
			t.Parallel()

			got := resolveOwnedDomains(
				"Fastly",
				"https://fastly.com",
				DomainsResult{
					Domains: []DomainCandidate{
						{Domain: "fastly.net", Confidence: 0.95},
					},
				},
				threshold,
			)

			assert.Equal(t, []string{"fastly.com", "fastly.net"}, domainsOf(got))
		},
	)

	t.Run(
		"drops shared-infrastructure domains for an unrelated vendor even at high confidence",
		func(t *testing.T) {
			t.Parallel()

			got := resolveOwnedDomains(
				"Dark Reader",
				"https://darkreader.org",
				DomainsResult{
					Domains: []DomainCandidate{
						{Domain: "amazonaws.com", Confidence: 0.99},
						{Domain: "fastly.net", Confidence: 0.99},
					},
				},
				threshold,
			)

			assert.Equal(t, []string{"darkreader.org"}, domainsOf(got))
		},
	)
}
