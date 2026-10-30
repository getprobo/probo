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

package scanner

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadStaticPage_CollectsEmbedHostsAndLinks(t *testing.T) {
	t.Parallel()

	site, ok := parseOrigin("https://example.com")
	require.True(t, ok)
	api, ok := parseOrigin("https://api.example.com/v1")
	require.True(t, ok)

	base, err := url.Parse("https://example.com/stories/a")
	require.NoError(t, err)

	body := `<!doctype html>
<probo-cookie-banner banner-id="banner"></probo-cookie-banner>
<script src="//cdn.example.net/a.js"></script>
<script src="/local.js"></script>
<script src="https://www.example.com/www.js"></script>
<script src="https://api.example.com/v1/banner.js"></script>
<iframe data-src="https://www.youtube.com/embed/x" data-cookie-consent="functional"></iframe>
<iframe data-tally-src="https://tally.so/embed/abc"></iframe>
<img src="https://images.example.org/a.png">
<object data="https://player.example.tv/x"></object>
<video><source src="https://media.example.io/v.mp4"></video>
<a href="/next">next</a>
<a href="https://www.example.com/www">www</a>
<a href="mailto:hello@example.com">mail</a>`

	page := readStaticPage(base, strings.NewReader(body), site, api)

	assert.Equal(
		t,
		[]string{"cdn.example.net", "media.example.io", "player.example.tv", "tally.so", "www.youtube.com"},
		page.hosts,
	)
	assert.Equal(t, []string{"https://example.com/next", "https://www.example.com/www"}, page.links)
	assert.Equal(t, "banner", page.bannerID)
}

func TestReadStaticPage_ResolvesAgainstBaseHref(t *testing.T) {
	t.Parallel()

	site, ok := parseOrigin("https://example.com")
	require.True(t, ok)
	api, ok := parseOrigin("https://api.example.com/v1")
	require.True(t, ok)

	base, err := url.Parse("https://example.com/stories/a")
	require.NoError(t, err)

	body := `<!doctype html>
<a href="/early">early</a>
<base href="https://cdn.example.net/app/">
<a href="next">next</a>
<iframe src="player"></iframe>
<base href="https://ignored.example/late/">
<a href="after">after</a>`

	page := readStaticPage(base, strings.NewReader(body), site, api)

	assert.Equal(t, []string{"cdn.example.net"}, page.hosts)
	assert.Equal(
		t,
		[]string{
			"https://example.com/early",
			"https://cdn.example.net/app/next",
			"https://cdn.example.net/app/after",
		},
		page.links,
	)
}

func TestProbeEmbeds_CancelIsNotAMissingBanner(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	t.Cleanup(server.Close)

	s := probeScan(t, server)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	confirmed, err := s.probeEmbeds(ctx, nil)
	require.NoError(t, err)
	assert.False(t, confirmed)
}

func TestProbeEmbeds_TagsQueuedPages(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html")

				switch r.URL.Path {
				case "/":
					_, _ = w.Write([]byte(`<probo-cookie-banner banner-id="banner"></probo-cookie-banner><a href="/stories/a">a</a><a href="/pricing">p</a><script src="https://cdn.example.net/app.js"></script>`))
				case "/stories/a":
					_, _ = w.Write([]byte(`<iframe data-src="https://www.youtube.com/embed/x"></iframe>`))
				case "/docs/x":
					_, _ = w.Write([]byte(`<iframe data-tally-src="https://tally.so/embed/abc"></iframe>`))
				case "/feed":
					w.Header().Set("Content-Type", "application/rss+xml")
					_, _ = w.Write([]byte(`<iframe src="https://feed.example.org/x"></iframe>`))
				default:
					_, _ = w.Write([]byte(`<p>nothing</p>`))
				}
			},
		),
	)
	t.Cleanup(server.Close)

	s := probeScan(t, server)

	sitemapURLs := []string{server.URL + "/docs/x", server.URL + "/feed", server.URL + "/stories/a"}
	for _, raw := range sitemapURLs {
		s.front.push(QueueItem{URL: raw, FromSitemap: true})
	}

	_, err := s.probeEmbeds(context.Background(), sitemapURLs)
	require.NoError(t, err)

	queue := map[string]QueueItem{}
	for _, item := range s.front.checkpoint(0, 0, false).Queue {
		queue[strings.TrimPrefix(item.URL, server.URL)] = item
	}

	assert.Equal(t, []string{"cdn.example.net"}, queue["/"].EmbedHosts)
	assert.Equal(t, []string{"tally.so"}, queue["/docs/x"].EmbedHosts)
	assert.True(t, queue["/docs/x"].FromSitemap)
	assert.Empty(t, queue["/feed"].EmbedHosts)
	assert.Equal(t, []string{"www.youtube.com"}, queue["/stories/a"].EmbedHosts)
	assert.False(t, queue["/stories/a"].FromSitemap)
	assert.Equal(t, 1, queue["/stories/a"].Depth)
	assert.NotContains(t, queue, "/pricing")
}

func TestProbeEmbeds_StopsWhenTheSiteRefuses(t *testing.T) {
	t.Parallel()

	var others atomic.Int32

	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/" {
					w.WriteHeader(http.StatusTooManyRequests)

					return
				}

				others.Add(1)

				_, _ = w.Write([]byte(`<iframe src="https://www.youtube.com/embed/x"></iframe>`))
			},
		),
	)
	t.Cleanup(server.Close)

	s := probeScan(t, server)
	s.front.push(QueueItem{URL: server.URL + "/docs/x", FromSitemap: true})

	_, err := s.probeEmbeds(context.Background(), []string{server.URL + "/docs/x"})
	require.ErrorIs(t, err, ErrBannerNotInstalled)

	assert.Equal(t, int32(0), others.Load())

	for _, item := range s.front.checkpoint(0, 0, false).Queue {
		assert.Empty(t, item.EmbedHosts)
	}
}

func TestProbeEmbeds_RequiresThisBanner(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		html string
		err  error
	}{
		{name: "matching id", html: `<probo-cookie-banner banner-id="banner"></probo-cookie-banner>`},
		{name: "different id", html: `<probo-cookie-banner banner-id="other"></probo-cookie-banner>`, err: ErrDifferentBanner},
		{name: "no element", html: `<p>home</p>`, err: ErrBannerNotInstalled},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var others atomic.Int32

			server := httptest.NewServer(
				http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Content-Type", "text/html")

						if r.URL.Path != "/" {
							others.Add(1)
						}

						_, _ = w.Write([]byte(tc.html))
					},
				),
			)
			t.Cleanup(server.Close)

			s := probeScan(t, server)
			s.front.push(QueueItem{URL: server.URL + "/docs", FromSitemap: true})

			_, err := s.probeEmbeds(context.Background(), []string{server.URL + "/docs"})
			if tc.err == nil {
				require.NoError(t, err)

				return
			}

			require.ErrorIs(t, err, tc.err)
			assert.Equal(t, int32(0), others.Load())
		})
	}
}

func probeScan(t *testing.T, server *httptest.Server) *scan {
	t.Helper()

	site, ok := parseOrigin(server.URL)
	require.True(t, ok)

	s := &scan{
		in:       Input{BannerID: "banner", Limits: Limits{}.withDefaults()},
		site:     site,
		startURL: site.root(),
		policy:   newRequestPolicy(site, origin{}, true),
		front:    newFrontier(Checkpoint{}, 50),
		http:     server.Client(),
	}
	s.front.push(QueueItem{URL: s.startURL, Depth: 0})

	return s
}
