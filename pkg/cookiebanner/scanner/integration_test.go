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
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/chromedp/chromedp/remote"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/httpclient"
	"go.probo.inc/probo/pkg/coredata"
)

func TestRun_RecordsTrackersAndSkipsOtherOrigins(t *testing.T) {
	t.Parallel()

	alloc, cancel := testChrome(t)
	defer cancel()

	var posts atomic.Int32

	api := serve(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			posts.Add(1)
		}

		w.WriteHeader(http.StatusNoContent)
	}))

	third := serve(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/pixel.js" {
			w.Header().Set("Content-Type", "text/javascript")
			_, _ = w.Write([]byte(`document.cookie = "tp=1; path=/";`))

			return
		}

		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<!doctype html><script src="/pixel.js"></script>`))
	}))

	var site = serve(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/track.js":
			w.Header().Set("Content-Type", "text/javascript")
			_, _ = w.Write([]byte(`document.cookie = "js=1; path=/";`))
		case "/next":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<!doctype html><p>next</p>`))
		default:
			w.Header().Set("Set-Cookie", "sid=secret; HttpOnly; Path=/")
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<!doctype html>
<script src="/track.js"></script>
<iframe src="` + chromeURL(t, third) + `"></iframe>
<a href="/next">next</a>
<a href="https://example.com/elsewhere">away</a>
<script>
fetch("` + chromeURL(t, api) + `/banner/report", {method: "POST", body: "{}"});
fetch("` + chromeURL(t, api) + `/banner/config");
var hits = Object.getOwnPropertyNames(globalThis).filter(function (k) {
  return /^b[0-9a-f]{32}$/.test(k);
});
localStorage.setItem("binding_hits", hits.join(","));
localStorage.setItem("old_binding", typeof globalThis.__proboScanCookie);
</script>`))
		}
	}))

	var (
		cookies   []Cookie
		storage   []StorageItem
		resources []Resource
	)

	result, err := Run(context.Background(), alloc, Input{
		Origin:             chromeURL(t, site),
		BannerID:           "banner",
		APIBaseURL:         chromeURL(t, api),
		BrowserURL:         os.Getenv("CHROME_WS_URL"),
		Limits:             fastLimits(5),
		skipPublicURLCheck: true,
		httpClient: httpclient.DefaultClient(
			httpclient.WithSSRFProtection(),
			httpclient.WithSSRFAllowLoopback(),
		),
		OnPage: func(_ context.Context, page Page, _ Checkpoint) error {
			cookies = append(cookies, page.Cookies...)
			storage = append(storage, page.Storage...)
			resources = append(resources, page.Resources...)

			return nil
		},
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.PagesScanned, 1)
	assert.LessOrEqual(t, result.PagesScanned, 2)

	var httpOnly, scripted, thirdParty bool

	for _, cookie := range cookies {
		switch cookie.Name {
		case "sid":
			httpOnly = cookie.Source == coredata.CookieSourceHTTP && cookie.HTTPOnly != nil && *cookie.HTTPOnly
		case "js":
			if cookie.Source == coredata.CookieSourceScript && cookie.InitiatorURL != nil && strings.Contains(*cookie.InitiatorURL, "/track.js") {
				scripted = true
			}
		case "tp":
			thirdParty = cookie.InitiatorURL != nil && strings.Contains(*cookie.InitiatorURL, "/pixel.js")
		}
	}

	assert.True(t, httpOnly, "httpOnly cookie")
	assert.True(t, scripted, "script cookie with initiator")
	assert.True(t, thirdParty, "third-party cookie")

	var sawIframe, sawPixel bool

	for _, resource := range resources {
		if resource.Type == coredata.TrackerResourceTypeIframe {
			sawIframe = true
		}

		if resource.Path == "/pixel.js" {
			sawPixel = true
		}

		assert.NotContains(t, resource.Origin, "example.com")
	}

	assert.True(t, sawIframe)
	assert.True(t, sawPixel, "script loaded inside the cross-site frame")
	assert.Equal(t, int32(0), posts.Load())

	var bindingHits, oldBinding bool

	for _, item := range storage {
		switch item.Key {
		case "binding_hits":
			bindingHits = item.ValueSize == 0
		case "old_binding":
			oldBinding = item.ValueSize == len("undefined")
		}
	}

	assert.True(t, bindingHits, "cookie binding name leaked onto the page")
	assert.True(t, oldBinding, "fixed binding name is still defined")
}

func TestRun_StopsAtPageLimit(t *testing.T) {
	t.Parallel()

	alloc, cancel := testChrome(t)
	defer cancel()

	site := serve(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<!doctype html><a href="/a">a</a><a href="/b">b</a><a href="/c">c</a>`))
	}))

	result, err := Run(context.Background(), alloc, Input{
		Origin:             chromeURL(t, site),
		BannerID:           "banner",
		APIBaseURL:         "https://api.example.com",
		Limits:             fastLimits(2),
		skipPublicURLCheck: true,
	})
	require.NoError(t, err)
	assert.Equal(t, 2, result.PagesScanned)
	assert.NotEmpty(t, result.Checkpoint.Queue)
}

func TestRun_BlocksOffSiteDocument(t *testing.T) {
	t.Parallel()

	alloc, cancel := testChrome(t)
	defer cancel()

	other := serve(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Set-Cookie", "offsite=1; Path=/")
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<!doctype html><p>elsewhere</p>`))
	}))

	site := serve(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/bounce" {
			http.Redirect(w, r, chromeURL(t, other), http.StatusFound)

			return
		}

		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<!doctype html><a href="/bounce">bounce</a>`))
	}))

	var cookies []Cookie

	result, err := Run(context.Background(), alloc, Input{
		Origin:             chromeURL(t, site),
		BannerID:           "banner",
		APIBaseURL:         "https://api.example.com",
		Limits:             fastLimits(3),
		skipPublicURLCheck: true,
		OnPage: func(_ context.Context, page Page, _ Checkpoint) error {
			cookies = append(cookies, page.Cookies...)

			return nil
		},
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.PagesScanned, 1)

	for _, cookie := range cookies {
		assert.NotEqual(t, "offsite", cookie.Name)
	}
}

func serve(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()

	ln, err := net.Listen("tcp", "0.0.0.0:0")
	require.NoError(t, err)

	server := httptest.NewUnstartedServer(handler)
	server.Listener = ln
	server.Start()
	t.Cleanup(server.Close)

	return server
}

// chromeURL rewrites a test server URL when Chrome runs in a container.
// Set CHROME_TEST_HOST=host.docker.internal so the container can reach it.
func chromeURL(t *testing.T, server *httptest.Server) string {
	t.Helper()

	host := os.Getenv("CHROME_TEST_HOST")
	if host == "" {
		return server.URL
	}

	parsed, err := url.Parse(server.URL)
	require.NoError(t, err)

	parsed.Host = net.JoinHostPort(host, parsed.Port())

	return parsed.String()
}

func testChrome(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()

	addr := os.Getenv("CHROME_WS_URL")
	if addr == "" {
		t.Skip("Skipping test: CHROME_WS_URL environment variable not set")
	}

	return remote.NewAllocator(context.Background(), addr)
}

func fastLimits(pages int) Limits {
	return Limits{
		MaxPages:          pages,
		MaxDepth:          1,
		NavigationTimeout: 10 * time.Second,
		HomepageDwell:     time.Millisecond,
		IdleWindow:        200 * time.Millisecond,
		LoadIdleTimeout:   time.Second,
		ScrollIdleTimeout: 300 * time.Millisecond,
		MaxSitemapIndexes: 1,
		MaxSitemapBytes:   1 << 20,
		MaxSitemapURLs:    10,
		MaxSitemapFetches: 2,
		MaxObservations:   100,
		MaxQueue:          50,
		ScanTimeout:       30 * time.Second,
	}
}
