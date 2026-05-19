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

package drivers

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCassettesUseSyntheticEmails enforces that VCR cassettes under
// testdata/ only contain emails from a controlled set of synthetic
// domains (RFC 2606 reserved + a small allowlist for OAuth bot
// identifiers in pre-existing cassettes). Real cassette recordings
// against test tenants must be scrubbed before commit; this test
// guards against accidental leakage of customer emails.
func TestCassettesUseSyntheticEmails(t *testing.T) {
	t.Parallel()

	emailRe := regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)

	// allowedDomainSuffixes lists trailing domain fragments that are safe
	// for synthetic test data. RFC 2606 reserves *.example.com / .org /
	// .net / .test / .invalid / .localhost. The allowlist is intentionally
	// narrow — adding to it should require maintainer review.
	allowedDomainSuffixes := []string{
		".example.com",
		".example.org",
		".example.net",
		".test",
		".invalid",
		".localhost",
		// Google Workspace test domain family (RFC-style synthetic).
		".test-google-a.com",
	}

	// allowedExactDomains lists individual domains that pre-date this
	// guard and were intentionally kept (synthetic OAuth bot identifiers
	// recorded against fixture tenants). Do not extend without review.
	allowedExactDomains := map[string]bool{
		"example.com":            true,
		"example.org":            true,
		"example.net":            true,
		"mail.com":               true,
		"contractor.example.com": true,
		"alias.example.com":      true,
		"oauthapp.linear.app":    true,
		"linear.linear.app":      true,
		"intercom.io":            true,
	}

	matches, err := filepath.Glob("testdata/*.yaml")
	require.NoError(t, err)
	require.NotEmpty(t, matches, "no cassettes found")

	for _, cassette := range matches {
		t.Run(filepath.Base(cassette), func(t *testing.T) {
			t.Parallel()

			data, err := os.ReadFile(cassette)
			require.NoError(t, err)

			seen := make(map[string]bool)
			for _, email := range emailRe.FindAllString(string(data), -1) {
				if seen[email] {
					continue
				}

				seen[email] = true

				domain := email[strings.IndexByte(email, '@')+1:]
				if allowedExactDomains[domain] {
					continue
				}

				ok := false

				for _, suffix := range allowedDomainSuffixes {
					if strings.HasSuffix("."+domain, suffix) || domain == strings.TrimPrefix(suffix, ".") {
						ok = true
						break
					}
				}

				// A failed assertion lands in CI logs. Keep both the
				// local-part AND the domain out of the message — together
				// they form a complete PII tuple, which is exactly what
				// this guard exists to prevent. Operators can grep the
				// cassette locally to identify the offending row.
				assert.Truef(
					t,
					ok,
					"cassette %s contains an email with a non-synthetic "+
						"domain; either replace with a synthetic "+
						"*.example.com address or add the domain to "+
						"allowedExactDomains in cassette_safety_test.go "+
						"with a justification",
					filepath.Base(cassette),
				)
			}
		})
	}
}
