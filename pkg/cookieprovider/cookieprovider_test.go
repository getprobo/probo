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

package cookieprovider

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAll(t *testing.T) {
	t.Parallel()

	providers := All()
	require.NotEmpty(t, providers)
	assert.Len(t, providers, 20)

	for _, p := range providers {
		assert.NotEmpty(t, p.Key, "provider key should not be empty")
		assert.NotEmpty(t, p.Name, "provider name should not be empty")
		assert.NotEmpty(t, p.Description, "provider description should not be empty")
		assert.NotEmpty(t, p.Category, "provider category should not be empty")
		assert.NotEmpty(t, p.WebsiteURL, "provider website URL should not be empty")
		assert.NotEmpty(t, p.Cookies, "provider %s should have at least one cookie", p.Key)
	}
}

func TestByKey(t *testing.T) {
	t.Parallel()

	t.Run("existing provider", func(t *testing.T) {
		t.Parallel()

		p, ok := ByKey("google-analytics")
		require.True(t, ok)
		assert.Equal(t, "Google Analytics", p.Name)
		assert.Equal(t, CategoryAnalytics, p.Category)
		assert.NotEmpty(t, p.Cookies)
	})

	t.Run("non-existing provider", func(t *testing.T) {
		t.Parallel()

		_, ok := ByKey("nonexistent")
		assert.False(t, ok)
	})
}

func TestByCategory(t *testing.T) {
	t.Parallel()

	analytics := ByCategory(CategoryAnalytics)
	require.NotEmpty(t, analytics)

	for _, p := range analytics {
		assert.Equal(t, CategoryAnalytics, p.Category)
	}

	marketing := ByCategory(CategoryMarketing)
	require.NotEmpty(t, marketing)

	for _, p := range marketing {
		assert.Equal(t, CategoryMarketing, p.Category)
	}
}

func TestCookieItems(t *testing.T) {
	t.Parallel()

	p, ok := ByKey("google-analytics")
	require.True(t, ok)

	items := p.CookieItems()
	require.Len(t, items, len(p.Cookies))

	for i, item := range items {
		assert.Equal(t, p.Cookies[i].Name, item.Name)
		assert.Equal(t, p.Cookies[i].Duration, item.Duration)
		assert.Equal(t, p.Cookies[i].Description, item.Description)
	}
}

func TestAllReturnsACopy(t *testing.T) {
	t.Parallel()

	p1 := All()
	p2 := All()

	p1[0].Name = "modified"
	assert.NotEqual(t, p1[0].Name, p2[0].Name)
}

func TestUniqueKeys(t *testing.T) {
	t.Parallel()

	seen := make(map[string]bool)
	for _, p := range All() {
		assert.False(t, seen[p.Key], "duplicate provider key: %s", p.Key)
		seen[p.Key] = true
	}
}
