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
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/netcheck"
)

func TestTemplateKey_CollapsesVariableSegments(t *testing.T) {
	t.Parallel()

	t.Run("numeric segments", func(t *testing.T) {
		t.Parallel()

		first, ok := templateKey("https://Example.com/posts/123?utm=1#top")
		require.True(t, ok)
		second, ok := templateKey("https://example.com/posts/456")
		require.True(t, ok)
		assert.Equal(t, first, second)
	})

	t.Run("uuid and hex segments", func(t *testing.T) {
		t.Parallel()

		uuid, ok := templateKey("https://example.com/orders/550e8400-e29b-41d4-a716-446655440000")
		require.True(t, ok)
		assert.Equal(t, "https://example.com/orders/:id", uuid)

		hash, ok := templateKey("https://example.com/r/abcdef0123456789")
		require.True(t, ok)
		assert.Equal(t, "https://example.com/r/:id", hash)
	})

	t.Run("static paths stay distinct", func(t *testing.T) {
		t.Parallel()

		static, ok := templateKey("https://example.com/pricing")
		require.True(t, ok)
		assert.Equal(t, "https://example.com/pricing", static)

		slug, ok := templateKey("https://example.com/privacy-policy-overview")
		require.True(t, ok)
		assert.Equal(t, "https://example.com/privacy-policy-overview", slug)
		assert.NotEqual(t, static, slug)
	})

	t.Run("encoded slash is a different path", func(t *testing.T) {
		t.Parallel()

		encoded, ok := templateKey("https://example.com/a%2Fb")
		require.True(t, ok)
		plain, ok := templateKey("https://example.com/a/b")
		require.True(t, ok)
		assert.NotEqual(t, encoded, plain)
	})

	t.Run("www and default port fold", func(t *testing.T) {
		t.Parallel()

		apex, ok := templateKey("https://example.com/pricing")
		require.True(t, ok)
		www, ok := templateKey("https://www.example.com/pricing")
		require.True(t, ok)
		port, ok := templateKey("https://example.com:443/pricing")
		require.True(t, ok)
		assert.Equal(t, apex, www)
		assert.Equal(t, apex, port)
	})
}

func TestOrigin_FoldsWWWAndPorts(t *testing.T) {
	t.Parallel()

	site, ok := parseOrigin("https://example.com/app")
	require.True(t, ok)
	assert.True(t, site.contains("https://www.Example.com:443/pricing"))
	assert.False(t, site.contains("https://example.com:8443/pricing"))
	assert.False(t, site.contains("https://other.example.com/pricing"))
	assert.Equal(t, "https://example.com/", site.root())
	assert.Equal(t, "https://www.example.com/", site.wwwRoot())

	_, ok = parseOrigin("https://user:pass@example.com")
	assert.False(t, ok)

	_, ok = parseOrigin("javascript:alert(1)")
	assert.False(t, ok)
}

func TestChooseStart_UsesWWWWhenApexDoesNotResolve(t *testing.T) {
	t.Parallel()

	site, ok := parseOrigin("https://example.com")
	require.True(t, ok)

	got, err := chooseStart(site, hostPublic, hostUnknown)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/", got)

	got, err = chooseStart(site, hostUnknown, hostPublic)
	require.NoError(t, err)
	assert.Equal(t, "https://www.example.com/", got)

	_, err = chooseStart(site, hostPrivate, hostPublic)
	require.ErrorIs(t, err, netcheck.ErrBlockedAddress)

	_, err = chooseStart(site, hostUnknown, hostUnknown)
	require.ErrorContains(t, err, "cannot resolve")
}

func TestParseSetCookie_ReadsScannerAttributes(t *testing.T) {
	t.Parallel()

	cookie, ok := parseSetCookie(
		"sid=abc; HttpOnly; Secure; SameSite=Lax; Max-Age=3600; Domain=.Example.com; Partitioned",
		"https://example.com/login?token=secret#frag",
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

	hostOnly, ok := parseSetCookie("id=1; Path=/", "https://example.com/")
	require.True(t, ok)
	assert.Nil(t, hostOnly.Domain)
	require.NotNil(t, hostOnly.HostOnly)
	assert.True(t, *hostOnly.HostOnly)

	_, ok = parseSetCookie("probo_consent=1; Path=/", "https://example.com/")
	assert.False(t, ok)

	deleted, ok := parseSetCookie("gone=1; Max-Age=0; Expires=Wed, 21 Oct 2030 07:28:00 GMT", "https://example.com/")
	require.True(t, ok)
	require.NotNil(t, deleted.MaxAgeSeconds)
	assert.Equal(t, 0, *deleted.MaxAgeSeconds)

	_, ok = parseSetCookie(strings.Repeat("n", maxNameBytes+1)+"=1", "https://example.com/")
	assert.False(t, ok)
}

func TestRequestPolicy_Block(t *testing.T) {
	t.Parallel()

	site, ok := parseOrigin("https://shop.example.com")
	require.True(t, ok)
	api, ok := parseOrigin("https://api.example.com:443/v1")
	require.True(t, ok)

	policy := newRequestPolicy(site, api, true)
	policy.setMainFrame("main")

	ctx := context.Background()
	banner := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	apiURL := "https://api.example.com/v1/" + banner

	assert.True(t, policy.block(ctx, http.MethodPost, apiURL+"/report", "", ""))
	assert.True(t, policy.block(ctx, http.MethodPut, apiURL+"/consents", "", ""))
	assert.True(t, policy.block(ctx, http.MethodDelete, "https://api.example.com/anything", "", ""))
	assert.False(t, policy.block(ctx, http.MethodGet, apiURL+"/config", "", ""))
	assert.False(t, policy.block(ctx, http.MethodHead, apiURL+"/consents/abc", "", ""))
	assert.False(t, policy.block(ctx, http.MethodOptions, apiURL+"/report", "", ""))
	assert.False(t, policy.block(ctx, http.MethodPost, "https://shop.example.com/"+banner+"/report", "", ""))

	assert.True(t, policy.block(ctx, http.MethodGet, "https://calendly.com/demo", string(network.ResourceTypeDocument), "main"))
	assert.False(t, policy.block(ctx, http.MethodGet, "https://calendly.com/embed", string(network.ResourceTypeDocument), "child"))
	assert.False(t, policy.block(ctx, http.MethodGet, "https://cdn.example.com/gtm.js", string(network.ResourceTypeScript), "main"))
	assert.True(t, policy.block(ctx, http.MethodGet, "http://[not a url", "", ""))
	assert.False(t, policy.block(ctx, http.MethodGet, "data:text/plain,hi", "", ""))

	open := newRequestPolicy(site, api, false)
	assert.True(t, open.block(ctx, http.MethodGet, "http://127.0.0.1/admin", "", ""))
	assert.True(t, open.block(ctx, http.MethodGet, "http://169.254.169.254/latest", "", ""))
	assert.False(t, open.block(ctx, http.MethodGet, "http://8.8.8.8/", "", ""))
}

func TestConsentCookie_IsURLEncoded(t *testing.T) {
	t.Parallel()

	params, ok, err := consentCookie(Input{
		BannerID: "banner",
		Consent:  &Consent{Version: 3, CategorySlugs: []string{"analytics"}},
	}, "https://example.com/", "abc")
	require.NoError(t, err)
	require.True(t, ok)
	assert.True(t, strings.HasPrefix(params.Value, "%7B"))
	assert.NotContains(t, params.Value, "{")
	assert.Contains(t, params.Value, "ACCEPT_ALL")
}

func TestCookieHook_HidesBinding(t *testing.T) {
	t.Parallel()

	source := cookieHookSource("b0123456789abcdef0123456789abcdef")
	assert.Contains(t, source, `delete globalThis`)
	assert.Contains(t, source, `"b0123456789abcdef0123456789abcdef"`)
	assert.NotContains(t, source, "__proboScanCookie")
}

func TestRun_RequiresAPIBaseURL(t *testing.T) {
	t.Parallel()

	_, err := Run(context.Background(), nil, Input{
		Origin:             "https://example.com",
		BannerID:           "banner",
		skipPublicURLCheck: true,
	})
	require.ErrorContains(t, err, "API base URL is required")
}

func TestRun_RejectsCredentialsInOrigin(t *testing.T) {
	t.Parallel()

	_, err := Run(context.Background(), nil, Input{
		Origin:   "https://user:pass@example.com",
		BannerID: "banner",
	})
	require.ErrorContains(t, err, "origin is invalid")
}

func TestRun_RejectsPrivateOrigin(t *testing.T) {
	t.Parallel()

	_, err := Run(context.Background(), nil, Input{
		Origin:     "http://127.0.0.1",
		BannerID:   "banner",
		APIBaseURL: "https://api.example.com",
	})
	require.ErrorIs(t, err, netcheck.ErrBlockedAddress)
}

func TestRun_TimeBudgetWithNoPages(t *testing.T) {
	t.Parallel()

	t.Run("no pages is an error", func(t *testing.T) {
		t.Parallel()

		result, err := Run(context.Background(), nil, Input{
			Origin:             "https://example.com",
			BannerID:           "banner",
			APIBaseURL:         "https://api.example.com",
			skipPublicURLCheck: true,
			Limits:             Limits{ScanTimeout: time.Second},
			Checkpoint:         &Checkpoint{ElapsedMS: 2000},
		})
		require.ErrorIs(t, err, ErrNoPageLoaded)
		require.NotNil(t, result)
		assert.True(t, result.Truncated)
		assert.False(t, result.Paused)
		assert.Equal(t, 0, result.PagesScanned)
	})

	t.Run("pages already scanned stay completed", func(t *testing.T) {
		t.Parallel()

		result, err := Run(context.Background(), nil, Input{
			Origin:             "https://example.com",
			BannerID:           "banner",
			APIBaseURL:         "https://api.example.com",
			skipPublicURLCheck: true,
			Limits:             Limits{ScanTimeout: time.Second},
			Checkpoint:         &Checkpoint{ElapsedMS: 2000, PagesScanned: 3},
		})
		require.NoError(t, err)
		assert.True(t, result.Truncated)
		assert.Equal(t, 3, result.PagesScanned)
	})
}

func TestRecorder_DeduplicatesAndDropsAPIOrigin(t *testing.T) {
	t.Parallel()

	site, err := mustOrigin("https://example.com")
	require.NoError(t, err)
	api, err := mustOrigin("https://api.example.com/v1")
	require.NoError(t, err)

	rec := newRecorder(site, api, "banner", "binding", 10)
	rec.beginPage()
	rec.handleEvent(network.EventRequestWillBeSent{
		RequestID: "1",
		Request:   &network.Request{URL: "https://cdn.example.com/gtm.js"},
		Type:      network.ResourceTypeScript,
		FrameID:   "main",
		Initiator: &network.Initiator{URL: "https://example.com/page?x=1"},
	})
	rec.handleEvent(network.EventRequestWillBeSent{
		RequestID: "1",
		Request:   &network.Request{URL: "https://cdn.example.com/gtm.js"},
		Type:      network.ResourceTypeScript,
		FrameID:   "main",
	})
	rec.handleEvent(network.EventRequestWillBeSent{
		RequestID: "2",
		Request:   &network.Request{URL: "https://api.example.com/v1/banner/config"},
		Type:      network.ResourceTypeFetch,
		FrameID:   "main",
	})
	rec.handleEvent(network.EventRequestWillBeSent{
		RequestID: "3",
		Request:   &network.Request{URL: "https://www.example.com/app.js"},
		Type:      network.ResourceTypeScript,
		FrameID:   "main",
	})
	rec.handleEvent(network.EventResponseReceivedExtraInfo{
		RequestID: "1",
		Headers:   network.Headers{"set-cookie": "sid=abc; HttpOnly; Secure; Domain=example.com"},
	})
	rec.handleEvent(runtime.EventBindingCalled{
		Name:    "binding",
		Payload: `{"name":"sid","host":"example.com","domain":"example.com","initiator":"https://cdn.example.com/gtm.js?x=1"}`,
	})
	rec.handleEvent(runtime.EventBindingCalled{
		Name:    "binding",
		Payload: `{"name":"_ga","host":"example.com","initiator":"https://cdn.example.com/gtm.js"}`,
	})

	page := rec.endPage([]*network.Cookie{{
		Name:     "sid",
		Value:    "abc",
		Domain:   ".example.com",
		HTTPOnly: true,
		Secure:   true,
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
	require.NotNil(t, page.Resources[0].InitiatorURL)
	assert.Equal(t, "https://example.com/page", *page.Resources[0].InitiatorURL)

	require.Len(t, page.Cookies, 3)

	byName := map[string]Cookie{}
	for _, cookie := range page.Cookies {
		byName[cookie.Name] = cookie
	}

	sid := byName["sid"]
	assert.Equal(t, coredata.CookieSourceHTTP, sid.Source)
	require.NotNil(t, sid.SetByURL)
	assert.Equal(t, "https://cdn.example.com/gtm.js", *sid.SetByURL)
	require.NotNil(t, sid.HTTPOnly)
	assert.True(t, *sid.HTTPOnly)
	require.NotNil(t, sid.InitiatorURL)
	assert.Equal(t, "https://cdn.example.com/gtm.js", *sid.InitiatorURL)

	ga := byName["_ga"]
	assert.Equal(t, coredata.CookieSourceScript, ga.Source)
	require.NotNil(t, ga.Domain)
	assert.Equal(t, "example.com", *ga.Domain)
	require.NotNil(t, ga.HostOnly)
	assert.True(t, *ga.HostOnly)

	_, apiCookie := byName["api_sid"]
	assert.False(t, apiCookie)
	assert.Contains(t, byName, "extra")

	again := rec.endPage(nil, nil)
	assert.Empty(t, again.Resources)
	assert.Empty(t, again.Cookies)
}

func TestRecorder_HostOnlyHeaderAndJarAreOneCookie(t *testing.T) {
	t.Parallel()

	rec := testRecorder(t)
	rec.beginPage()
	rec.handleEvent(network.EventRequestWillBeSent{
		RequestID: "1",
		Request:   &network.Request{URL: "https://shop.example.com/login"},
	})
	rec.handleEvent(network.EventResponseReceivedExtraInfo{
		RequestID: "1",
		Headers:   network.Headers{"set-cookie": "id=1; HttpOnly"},
	})

	page := rec.endPage([]*network.Cookie{{
		Name:     "id",
		Value:    "1",
		Domain:   "shop.example.com",
		HTTPOnly: true,
	}}, nil)

	require.Len(t, page.Cookies, 1)
	assert.Equal(t, coredata.CookieSourceHTTP, page.Cookies[0].Source)
	require.NotNil(t, page.Cookies[0].HostOnly)
	assert.True(t, *page.Cookies[0].HostOnly)
	require.NotNil(t, page.Cookies[0].Domain)
	assert.Equal(t, "shop.example.com", *page.Cookies[0].Domain)
	require.NotNil(t, page.Cookies[0].SetByURL)
	assert.Equal(t, "https://shop.example.com/login", *page.Cookies[0].SetByURL)
}

func TestScan_PauseKeepsTheObservationCap(t *testing.T) {
	t.Parallel()

	rec := testRecorder(t)
	rec.truncated = true
	s := &scan{
		rec:     rec,
		front:   newFrontier(Checkpoint{}, 10),
		pages:   2,
		started: time.Now(),
	}

	result, err := s.finish(causePaused)
	require.NoError(t, err)
	assert.True(t, result.Paused)
	assert.False(t, result.Truncated)
	assert.True(t, result.Checkpoint.Truncated)
	assert.Equal(t, 2, result.PagesScanned)
}

func TestRun_TruncatedCheckpointStops(t *testing.T) {
	t.Parallel()

	result, err := Run(context.Background(), nil, Input{
		Origin:             "https://example.com",
		BannerID:           "banner",
		APIBaseURL:         "https://api.example.com",
		skipPublicURLCheck: true,
		Checkpoint:         &Checkpoint{Truncated: true, PagesScanned: 2},
	})
	require.NoError(t, err)
	assert.True(t, result.Truncated)
	assert.False(t, result.Paused)
	assert.Equal(t, 2, result.PagesScanned)
}

func TestRecorder_RedirectDoesNotDoubleCountInflight(t *testing.T) {
	t.Parallel()

	rec := testRecorder(t)
	rec.beginPage()
	rec.handleEvent(network.EventRequestWillBeSent{RequestID: "1", Request: &network.Request{URL: "https://cdn.example.com/a"}})
	rec.handleEvent(network.EventRequestWillBeSent{RequestID: "1", Request: &network.Request{URL: "https://cdn.example.com/b"}})
	rec.handleEvent(network.EventRequestServedFromCache{RequestID: "1"})
	assert.False(t, rec.idleFor(0))

	rec.handleEvent(network.EventLoadingFinished{RequestID: "1"})
	assert.True(t, rec.idleFor(0))
}

func TestRecorder_FinishBeforeStartStillGoesIdle(t *testing.T) {
	t.Parallel()

	rec := testRecorder(t)
	rec.beginPage()
	rec.handleEvent(network.EventLoadingFinished{RequestID: "1"})
	assert.False(t, rec.idleFor(0))

	rec.handleEvent(network.EventRequestWillBeSent{
		RequestID: "1",
		Request:   &network.Request{URL: "https://cdn.example.com/a"},
	})
	assert.True(t, rec.idleFor(0))
}

func TestRecorder_SetCookieBeforeRequestKeepsSetter(t *testing.T) {
	t.Parallel()

	rec := testRecorder(t)
	rec.beginPage()
	rec.handleEvent(network.EventResponseReceivedExtraInfo{
		RequestID: "9",
		Headers:   network.Headers{"set-cookie": "sid=abc; Domain=example.com"},
	})
	rec.handleEvent(network.EventRequestWillBeSent{
		RequestID: "9",
		Request:   &network.Request{URL: "https://cdn.example.com/set?token=1"},
	})

	page := rec.endPage([]*network.Cookie{{
		Name:   "sid",
		Value:  "abc",
		Domain: ".example.com",
	}}, nil)

	require.Len(t, page.Cookies, 1)
	assert.Equal(t, coredata.CookieSourceHTTP, page.Cookies[0].Source)
	require.NotNil(t, page.Cookies[0].SetByURL)
	assert.Equal(t, "https://cdn.example.com/set", *page.Cookies[0].SetByURL)
}

func TestRecorder_DropsInvalidBlockedCookies(t *testing.T) {
	t.Parallel()

	rec := testRecorder(t)
	rec.beginPage()
	rec.handleEvent(network.EventRequestWillBeSent{
		RequestID: "1",
		Request:   &network.Request{URL: "https://cdn.example.com/a"},
	})
	rec.handleEvent(network.EventResponseReceivedExtraInfo{
		RequestID: "1",
		Headers:   network.Headers{"set-cookie": "bad=1\ngood=1"},
		BlockedCookies: []*network.BlockedSetCookieWithReason{{
			CookieLine:     "bad=1",
			BlockedReasons: []network.SetCookieBlockedReason{network.SetCookieBlockedReasonSyntaxError},
		}, {
			CookieLine:     "pref=1; Domain=cdn.example.com",
			BlockedReasons: []network.SetCookieBlockedReason{network.SetCookieBlockedReasonUserPreferences},
		}},
	})

	page := rec.endPage(nil, nil)

	names := map[string]bool{}
	for _, cookie := range page.Cookies {
		names[cookie.Name] = true
	}

	assert.False(t, names["bad"])
	assert.True(t, names["good"])
	assert.True(t, names["pref"])
}

func TestRecorder_UsesStackFrameWhenInitiatorURLIsEmpty(t *testing.T) {
	t.Parallel()

	rec := testRecorder(t)
	rec.beginPage()
	rec.handleEvent(network.EventRequestWillBeSent{
		RequestID: "1",
		Request:   &network.Request{URL: "https://cdn.example.com/gtm.js"},
		Type:      network.ResourceTypeScript,
		FrameID:   "main",
		Initiator: &network.Initiator{
			Type: network.InitiatorTypeScript,
			Stack: &runtime.StackTrace{CallFrames: []*runtime.CallFrame{
				{URL: ""},
				{URL: "https://tags.example.com/loader.js?x=1"},
			}},
		},
	})

	page := rec.endPage(nil, nil)
	require.Len(t, page.Resources, 1)
	require.NotNil(t, page.Resources[0].InitiatorURL)
	assert.Equal(t, "https://tags.example.com/loader.js", *page.Resources[0].InitiatorURL)
}

func TestRecorder_DropsConsentStoragePrefix(t *testing.T) {
	t.Parallel()

	rec := testRecorder(t)
	rec.beginPage()
	page := rec.endPage(nil, []StorageItem{
		{Type: coredata.TrackerTypeLocalStorage, Key: "probo_consent:banner:vid", ValueSize: 32},
		{Type: coredata.TrackerTypeLocalStorage, Key: "probo_consent:banner:other", ValueSize: 2},
		{Type: coredata.TrackerTypeLocalStorage, Key: "theme", ValueSize: 4},
	})

	require.Len(t, page.Storage, 1)
	assert.Equal(t, "theme", page.Storage[0].Key)
}

func TestFrontier_FailedTemplateIsNotRequeued(t *testing.T) {
	t.Parallel()

	front := newFrontier(Checkpoint{}, 10)
	front.markVisited("https://example.com/missing")
	front.push(QueueItem{URL: "https://www.example.com/missing"})

	assert.Empty(t, front.checkpoint(0, 0, false).Queue)
}

func TestFrontier_RequeueIgnoresCap(t *testing.T) {
	t.Parallel()

	front := newFrontier(Checkpoint{}, 1)
	front.push(QueueItem{URL: "https://example.com/a"})
	front.requeue(QueueItem{URL: "https://example.com/b"})

	assert.Len(t, front.checkpoint(0, 0, false).Queue, 2)
}

func TestFrontier_CapsQueue(t *testing.T) {
	t.Parallel()

	front := newFrontier(Checkpoint{}, 2)
	front.push(QueueItem{URL: "https://example.com/a"})
	front.push(QueueItem{URL: "https://example.com/b"})
	front.push(QueueItem{URL: "https://example.com/c"})

	assert.Len(t, front.checkpoint(0, 0, false).Queue, 2)
}

func TestCollectSitemapURLs_DoesNotEnqueueIndexAtDepthLimit(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<sitemapindex><sitemap><loc>https://example.com/nested.xml</loc></sitemap></sitemapindex>`))
	}))
	t.Cleanup(server.Close)

	site, ok := parseOrigin(server.URL)
	require.True(t, ok)

	s := &scan{
		in: Input{Limits: Limits{
			MaxSitemapIndexes: 1,
			MaxSitemapURLs:    10,
			MaxSitemapBytes:   1 << 20,
			MaxSitemapFetches: 5,
		}},
		site:   site,
		policy: newRequestPolicy(site, origin{}, true),
		http:   server.Client(),
	}

	var (
		urls    []string
		fetches int
	)
	s.collectSitemapURLs(context.Background(), server.URL, 1, map[string]struct{}{}, &urls, &fetches)
	assert.Empty(t, urls)
	assert.Equal(t, 1, fetches)
}

func testRecorder(t *testing.T) *recorder {
	t.Helper()

	site, err := mustOrigin("https://example.com")
	require.NoError(t, err)

	return newRecorder(site, origin{scheme: "https", host: "api.example.com"}, "banner", "binding", 10)
}

func mustOrigin(raw string) (origin, error) {
	parsed, ok := parseOrigin(raw)
	if !ok {
		return origin{}, errResolve(raw)
	}

	return parsed, nil
}
