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

package cookiebanner_v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeCookieDomain(t *testing.T) {
	t.Parallel()

	t.Run("normalizes a Domain attribute", func(t *testing.T) {
		t.Parallel()

		raw := ".Example.COM"
		got := sanitizeCookieDomain(&raw)
		require.NotNil(t, got)
		assert.Equal(t, "example.com", *got)
	})

	t.Run("drops an invalid hostname", func(t *testing.T) {
		t.Parallel()

		raw := "not a domain"
		assert.Nil(t, sanitizeCookieDomain(&raw))
	})

	t.Run("omits a missing domain", func(t *testing.T) {
		t.Parallel()

		assert.Nil(t, sanitizeCookieDomain(nil))
	})
}

func TestSanitizeCookieDomainFields(t *testing.T) {
	t.Parallel()

	t.Run("Domain attribute implies not host-only", func(t *testing.T) {
		t.Parallel()

		raw := "example.com"
		domain, hostOnly := sanitizeCookieDomainFields(&raw, nil)
		require.NotNil(t, domain)
		assert.Equal(t, "example.com", *domain)
		require.NotNil(t, hostOnly)
		assert.False(t, *hostOnly)
	})

	t.Run("drops host_only with a bad domain", func(t *testing.T) {
		t.Parallel()

		raw := "https://evil.example/path"
		notHostOnly := false
		domain, hostOnly := sanitizeCookieDomainFields(&raw, &notHostOnly)
		assert.Nil(t, domain)
		assert.Nil(t, hostOnly)
	})

	t.Run("keeps a confirmed host-only flag", func(t *testing.T) {
		t.Parallel()

		hostOnlyTrue := true
		domain, hostOnly := sanitizeCookieDomainFields(nil, &hostOnlyTrue)
		assert.Nil(t, domain)
		require.NotNil(t, hostOnly)
		assert.True(t, *hostOnly)
	})

	t.Run("drops host_only false without a domain", func(t *testing.T) {
		t.Parallel()

		notHostOnly := false
		domain, hostOnly := sanitizeCookieDomainFields(nil, &notHostOnly)
		assert.Nil(t, domain)
		assert.Nil(t, hostOnly)
	})

	t.Run("drops a trailing-dot Domain", func(t *testing.T) {
		t.Parallel()

		raw := "example.com."
		assert.Nil(t, sanitizeCookieDomain(&raw))
	})
}

func TestSanitizeInt4(t *testing.T) {
	t.Parallel()

	assert.Nil(t, sanitizeInt4(nil))
	assert.Nil(t, sanitizeInt4(new(0)))
	assert.Nil(t, sanitizeInt4(new(251610986978)))

	got := sanitizeInt4(new(3600))
	require.NotNil(t, got)
	assert.Equal(t, 3600, *got)
}
