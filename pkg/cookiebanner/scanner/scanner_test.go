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
	"testing"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

func TestTemplateKey_CollapsesVariableSegments(t *testing.T) {
	t.Parallel()

	first, ok := templateKey("https://Example.com/posts/123?utm=1#top")
	require.True(t, ok)
	second, ok := templateKey("https://example.com/posts/456")
	require.True(t, ok)
	assert.Equal(t, first, second)

	uuid, ok := templateKey("https://example.com/orders/550e8400-e29b-41d4-a716-446655440000")
	require.True(t, ok)
	assert.Equal(t, "https://example.com/orders/:id", uuid)

	hash, ok := templateKey("https://example.com/r/abcdef0123456789")
	require.True(t, ok)
	assert.Equal(t, "https://example.com/r/:id", hash)

	static, ok := templateKey("https://example.com/pricing")
	require.True(t, ok)
	assert.Equal(t, "https://example.com/pricing", static)
	assert.NotEqual(t, first, static)

	slug, ok := templateKey("https://example.com/privacy-policy-overview")
	require.True(t, ok)
	assert.Equal(t, "https://example.com/privacy-policy-overview", slug)

	encoded, ok := templateKey("https://example.com/a%2Fb")
	require.True(t, ok)
	plain, ok := templateKey("https://example.com/a/b")
	require.True(t, ok)
	assert.NotEqual(t, encoded, plain)

	assert.True(t, sameOrigin("https://example.com/app", "https://example.com:443/pricing"))
	assert.False(t, sameOrigin("https://example.com/app", "https://example.com:8443/pricing"))
}

func TestMapSetCookie_ReadsScannerAttributes(t *testing.T) {
	t.Parallel()

	cookie, ok := mapSetCookie(
		"sid=abc; HttpOnly; Secure; SameSite=Lax; Max-Age=3600; Domain=.Example.com; Partitioned",
		"https://example.com/login",
	)
	require.True(t, ok)
	assert.Equal(t, "sid", cookie.Name)
	assert.Equal(t, coredata.CookieSourceHTTP, cookie.Source)
	assert.Equal(t, 3, cookie.ValueSize)
	require.NotNil(t, cookie.HTTPOnly)
	assert.True(t, *cookie.HTTPOnly)
	require.NotNil(t, cookie.Secure)
	assert.True(t, *cookie.Secure)
	require.NotNil(t, cookie.SameSite)
	assert.Equal(t, coredata.CookieSameSiteLax, *cookie.SameSite)
	require.NotNil(t, cookie.MaxAgeSeconds)
	assert.Equal(t, 3600, *cookie.MaxAgeSeconds)
	require.NotNil(t, cookie.Domain)
	assert.Equal(t, "Example.com", *cookie.Domain)
	require.NotNil(t, cookie.HostOnly)
	assert.False(t, *cookie.HostOnly)
	require.NotNil(t, cookie.Partitioned)
	assert.True(t, *cookie.Partitioned)
	require.NotNil(t, cookie.SetByURL)
	assert.Equal(t, "https://example.com/login", *cookie.SetByURL)

	hostOnly, ok := mapSetCookie("id=1; Path=/", "https://example.com/")
	require.True(t, ok)
	assert.Nil(t, hostOnly.Domain)
	require.NotNil(t, hostOnly.HostOnly)
	assert.True(t, *hostOnly.HostOnly)

	_, ok = mapSetCookie("probo_consent=1; Path=/", "https://example.com/")
	assert.False(t, ok)

	deleted, ok := mapSetCookie("gone=1; Max-Age=0; Expires=Wed, 21 Oct 2030 07:28:00 GMT", "https://example.com/")
	require.True(t, ok)
	require.NotNil(t, deleted.MaxAgeSeconds)
	assert.Equal(t, 0, *deleted.MaxAgeSeconds)
}

func TestBlockSDKPost_AllowsConfigAndConsentRead(t *testing.T) {
	t.Parallel()

	banner := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	assert.True(t, blockSDKPost("POST", "https://api.example.com/v1/"+banner+"/report", banner))
	assert.True(t, blockSDKPost("POST", "https://api.example.com/v1/"+banner+"/detected-cookies", banner))
	assert.True(t, blockSDKPost("POST", "https://api.example.com/v1/"+banner+"/consents", banner))
	assert.False(t, blockSDKPost("GET", "https://api.example.com/v1/"+banner+"/config", banner))
	assert.False(t, blockSDKPost("GET", "https://api.example.com/v1/"+banner+"/consents/abc", banner))
	assert.False(t, blockSDKPost("POST", "https://api.example.com/v1/other/report", banner))
}

func TestRecorder_DeduplicatesAndDropsAPIOrigin(t *testing.T) {
	t.Parallel()

	rec := newRecorder("https://api.example.com/v1", "banner", 10)
	request := &network.EventRequestWillBeSent{
		RequestID: "1",
		Request:   &network.Request{URL: "https://cdn.example.com/gtm.js"},
		Type:      network.ResourceTypeScript,
		FrameID:   "main",
		Initiator: &network.Initiator{URL: "https://example.com/"},
	}
	rec.append(request)
	rec.append(request)

	api := &network.EventRequestWillBeSent{
		RequestID: "2",
		Request:   &network.Request{URL: "https://api.example.com/v1/banner/config"},
		Type:      network.ResourceTypeFetch,
		FrameID:   "main",
	}
	rec.append(api)

	header := &network.EventResponseReceivedExtraInfo{
		RequestID: "1",
		Headers:   network.Headers{"set-cookie": "sid=abc; HttpOnly; Secure; Domain=example.com"},
	}
	rec.append(header)

	rec.append(&runtime.EventBindingCalled{
		Name:    cookieBinding,
		Payload: `{"name":"sid","initiator":"https://example.com/gtm.js"}`,
	})
	rec.append(&runtime.EventBindingCalled{
		Name:    cookieBinding,
		Payload: `{"name":"_ga","initiator":"https://cdn.example.com/gtm.js"}`,
	})

	page := rec.collect(0, []*network.Cookie{{
		Name:   "sid",
		Value:  "abc",
		Domain: "example.com",
	}, {
		Name:   "extra",
		Value:  "1",
		Domain: "other.example",
	}, {
		Name:   "api_sid",
		Value:  "1",
		Domain: "api.example.com",
	}}, nil)

	require.Len(t, page.Resources, 1)
	assert.Equal(t, coredata.TrackerResourceTypeScript, page.Resources[0].Type)
	assert.Equal(t, "https://cdn.example.com", page.Resources[0].Origin)
	assert.Equal(t, "/gtm.js", page.Resources[0].Path)

	require.Len(t, page.Cookies, 3)
	assert.Equal(t, "sid", page.Cookies[0].Name)
	assert.Equal(t, coredata.CookieSourceHTTP, page.Cookies[0].Source)
	assert.Equal(t, "_ga", page.Cookies[1].Name)
	assert.Equal(t, coredata.CookieSourceScript, page.Cookies[1].Source)
	assert.Equal(t, "extra", page.Cookies[2].Name)

	for _, cookie := range page.Cookies {
		assert.NotEqual(t, "api_sid", cookie.Name)
	}

	again := rec.collect(len(rec.events), nil, nil)
	assert.Empty(t, again.Resources)
	assert.Empty(t, again.Cookies)
}

func TestRecorder_RedirectDoesNotDoubleCountInflight(t *testing.T) {
	t.Parallel()

	rec := newRecorder("", "banner", 10)
	rec.append(&network.EventRequestWillBeSent{RequestID: "1", Request: &network.Request{URL: "https://example.com/a"}})
	rec.append(&network.EventRequestWillBeSent{RequestID: "1", Request: &network.Request{URL: "https://example.com/b"}})
	rec.append(&network.EventRequestServedFromCache{RequestID: "1"})
	assert.False(t, rec.quietFor(0))

	rec.append(&network.EventLoadingFinished{RequestID: "1"})
	assert.True(t, rec.quietFor(0))
}

func TestReadSitemap_DoesNotEnqueueIndexAtDepthLimit(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<sitemapindex><sitemap><loc>https://example.com/nested.xml</loc></sitemap></sitemapindex>`))
	}))
	t.Cleanup(server.Close)

	in := Input{
		Limits: Limits{
			MaxSitemapIndexes: 1,
			MaxSitemapURLs:    10,
			MaxSitemapBytes:   1 << 20,
		},
		skipPublicURLCheck: true,
		httpClient:         server.Client(),
	}

	var (
		urls    []string
		fetches int
	)
	in.readSitemap(context.Background(), server.URL, 1, map[string]struct{}{}, &urls, &fetches)
	assert.Empty(t, urls)
}
