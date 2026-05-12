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

package rfc5988_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/rfc5988"
)

func TestParse(t *testing.T) {
	t.Parallel()

	t.Run("empty header", func(t *testing.T) {
		t.Parallel()

		links := rfc5988.Parse("")
		assert.Nil(t, links)
	})

	t.Run("single link", func(t *testing.T) {
		t.Parallel()

		links := rfc5988.Parse(`<https://api.example.com/items?page=2>; rel="next"`)
		require.Len(t, links, 1)
		assert.Equal(t, "https://api.example.com/items?page=2", links[0].URL)
		assert.Equal(t, "next", links[0].Params["rel"])
	})

	t.Run("multiple links", func(t *testing.T) {
		t.Parallel()

		header := `<https://api.example.com/items?page=2>; rel="next", <https://api.example.com/items?page=5>; rel="last"`
		links := rfc5988.Parse(header)
		require.Len(t, links, 2)
		assert.Equal(t, "https://api.example.com/items?page=2", links[0].URL)
		assert.Equal(t, "next", links[0].Params["rel"])
		assert.Equal(t, "https://api.example.com/items?page=5", links[1].URL)
		assert.Equal(t, "last", links[1].Params["rel"])
	})

	t.Run("multiple params per link", func(t *testing.T) {
		t.Parallel()

		header := `<https://sentry.io/api/0/?cursor=abc>; rel="next"; results="true"; cursor="abc"`
		links := rfc5988.Parse(header)
		require.Len(t, links, 1)
		assert.Equal(t, "https://sentry.io/api/0/?cursor=abc", links[0].URL)
		assert.Equal(t, "next", links[0].Params["rel"])
		assert.Equal(t, "true", links[0].Params["results"])
		assert.Equal(t, "abc", links[0].Params["cursor"])
	})

	t.Run("github style link header", func(t *testing.T) {
		t.Parallel()

		header := `<https://api.github.com/orgs/foo/members?page=2>; rel="next", <https://api.github.com/orgs/foo/members?page=3>; rel="last"`
		links := rfc5988.Parse(header)
		require.Len(t, links, 2)
		assert.Equal(t, "next", links[0].Params["rel"])
		assert.Equal(t, "last", links[1].Params["rel"])
	})

	t.Run("sentry style link header", func(t *testing.T) {
		t.Parallel()

		header := `<https://sentry.io/api/0/orgs/slug/members/?cursor=prev>; rel="previous"; results="false"; cursor="prev", <https://sentry.io/api/0/orgs/slug/members/?cursor=next>; rel="next"; results="true"; cursor="next"`
		links := rfc5988.Parse(header)
		require.Len(t, links, 2)
		assert.Equal(t, "previous", links[0].Params["rel"])
		assert.Equal(t, "false", links[0].Params["results"])
		assert.Equal(t, "next", links[1].Params["rel"])
		assert.Equal(t, "true", links[1].Params["results"])
	})
}

func TestFindByRel(t *testing.T) {
	t.Parallel()

	t.Run("empty header", func(t *testing.T) {
		t.Parallel()

		url := rfc5988.FindByRel("", "next")
		assert.Empty(t, url)
	})

	t.Run("found", func(t *testing.T) {
		t.Parallel()

		header := `<https://api.example.com/items?page=2>; rel="next", <https://api.example.com/items?page=5>; rel="last"`
		url := rfc5988.FindByRel(header, "next")
		assert.Equal(t, "https://api.example.com/items?page=2", url)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		header := `<https://api.example.com/items?page=2>; rel="prev"`
		url := rfc5988.FindByRel(header, "next")
		assert.Empty(t, url)
	})
}
