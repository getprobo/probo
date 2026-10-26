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
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"iter"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/cdproto/storage"
	"github.com/chromedp/chromedp"
	"go.probo.inc/probo/pkg/coredata"
)

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

const cookieHookJS = `(function () {
  var proto = Document.prototype;
  var desc = Object.getOwnPropertyDescriptor(proto, "cookie");
  if (!desc || typeof desc.set !== "function") return;
  var nativeSet = desc.set;
  Object.defineProperty(proto, "cookie", {
    configurable: true,
    enumerable: desc.enumerable,
    get: desc.get,
    set: function (value) {
      try {
        var raw = String(value);
        var name = raw.split("=")[0].trim();
        var initiator = "";
        var stack = "";
        try { stack = (new Error()).stack || ""; } catch (e) {}
        var lines = stack.split("\n");
        for (var i = 0; i < lines.length; i++) {
          var match = lines[i].match(/https?:\/\/[^\s)]+/);
          if (!match) continue;
          initiator = match[0].replace(/:\d+(?::\d+)?$/, "");
          break;
        }
        if (typeof globalThis.__proboScanCookie === "function") {
          globalThis.__proboScanCookie(JSON.stringify({ name: name, initiator: initiator }));
        }
      } catch (e) {}
      return nativeSet.call(this, value);
    }
  });
})();`

const linkJS = `Array.from(document.querySelectorAll("a[href]")).map(function (a) { return a.href; })`

const scrollJS = `(() => {
  var height = Math.max(window.innerHeight || 0, 400);
  var max = Math.max(
    document.body ? document.body.scrollHeight : 0,
    document.documentElement ? document.documentElement.scrollHeight : 0
  );
  var y = window.scrollY || 0;
  if (y + height >= max - 2) {
    window.scrollTo(0, max);
    return true;
  }
  window.scrollTo(0, y + height);
  return false;
})()`

const storageJS = `(() => {
  var read = function (store) {
    var out = [];
    try {
      for (var i = 0; i < store.length; i++) {
        var key = store.key(i);
        var value = store.getItem(key) || "";
        out.push({ key: key, size: value.length });
      }
    } catch (e) {}
    return out;
  };
  return { local: read(localStorage), session: read(sessionStorage) };
})()`

// Run crawls in.Origin in an isolated browser context of allocCtx. A
// cancelled ctx pauses the scan after the last page that finished. The
// scan budget is Limits.ScanTimeout minus checkpoint elapsed time.
func Run(ctx context.Context, allocCtx context.Context, in Input) (*Result, error) {
	in.Limits = in.Limits.withDefaults()
	root, ok := originRoot(in.Origin)
	if !ok || in.BannerID == "" {
		return nil, fmt.Errorf("cannot start scan: origin and banner id are required")
	}
	if !in.allow(root) {
		return nil, fmt.Errorf("cannot start scan: origin is not a public URL")
	}

	start := checkpointOrEmpty(in.Checkpoint)
	elapsed := time.Duration(start.ElapsedMS) * time.Millisecond
	remaining := in.Limits.ScanTimeout - elapsed
	if remaining <= 0 || start.Truncated {
		start.Truncated = true

		return &Result{
			PagesScanned: start.PagesScanned,
			Truncated:    true,
			Checkpoint:   start,
		}, nil
	}

	scanCtx, cancelScan := context.WithTimeout(ctx, remaining)
	defer cancelScan()
	started := time.Now()

	tabCtx, cancelTab := chromedp.NewContext(allocCtx, chromedp.WithNewBrowserContext())
	defer cancelTab()
	go func() {
		select {
		case <-scanCtx.Done():
			cancelTab()
		case <-tabCtx.Done():
		}
	}()

	rec := newRecorder(in.APIBaseURL, in.BannerID, in.Limits.MaxObservations)
	if err := chromedp.Do(tabCtx, chromedp.Func(func(actionCtx context.Context, target *chromedp.Target) error {
		listen(tabCtx, target, rec, in.BannerID)

		return prepareBrowser(actionCtx, target, in, root)
	})); err != nil {
		if paused, truncated := stopReason(ctx, scanCtx); paused || truncated {
			return stoppedResult(start, elapsed+time.Since(started), paused, truncated), nil
		}

		return nil, fmt.Errorf("cannot prepare browser: %w", err)
	}

	front := newFrontier(start)
	if len(start.VisitedTemplates) == 0 && len(start.Queue) == 0 {
		front.add(QueueItem{URL: root, Depth: 0})
		in.seedSitemaps(scanCtx, root, &front)
	}

	result := &Result{PagesScanned: start.PagesScanned}
	for result.PagesScanned < in.Limits.MaxPages {
		if paused, truncated := stopReason(ctx, scanCtx); paused || truncated {
			return finish(result, front, elapsed+time.Since(started), paused, truncated || rec.truncatedObservations()), nil
		}
		if rec.truncatedObservations() {
			return finish(result, front, elapsed+time.Since(started), false, true), nil
		}

		item, found := front.pop()
		if !found {
			break
		}
		if !sameOrigin(root, item.URL) || !in.allow(item.URL) {
			continue
		}

		visited, loaded := visitPage(scanCtx, tabCtx, rec, in, item)
		if !loaded {
			if paused, truncated := stopReason(ctx, scanCtx); paused || truncated {
				return finish(result, front, elapsed+time.Since(started), paused, truncated || rec.truncatedObservations()), nil
			}

			continue
		}

		front.visit(item.URL)
		if !item.FromSitemap && item.Depth < in.Limits.MaxDepth {
			for _, href := range visited.links {
				link, linkOK := normalizeLink(item.URL, href)
				if !linkOK || !sameOrigin(root, link) || !in.allow(link) {
					continue
				}
				front.add(QueueItem{URL: link, Depth: item.Depth + 1})
			}
		}

		result.PagesScanned++
		result.Cookies = append(result.Cookies, visited.observations.Cookies...)
		result.Storage = append(result.Storage, visited.observations.Storage...)
		result.Resources = append(result.Resources, visited.observations.Resources...)
		checkpoint := front.checkpoint(result.PagesScanned, elapsed+time.Since(started), rec.truncatedObservations())
		visited.observations.URL = item.URL
		if in.OnPage != nil {
			if err := in.OnPage(context.WithoutCancel(ctx), visited.observations, checkpoint); err != nil {
				return nil, fmt.Errorf("cannot accept scanned page: %w", err)
			}
		}
		result.Checkpoint = checkpoint
	}

	_, truncated := stopReason(ctx, scanCtx)
	paused, _ := stopReason(ctx, scanCtx)
	truncated = truncated || rec.truncatedObservations()
	result = finish(result, front, elapsed+time.Since(started), paused, truncated)
	if result.PagesScanned == 0 && !result.Paused {
		return result, ErrNoPageLoaded
	}

	return result, nil
}

type visitedPage struct {
	observations Page
	links        []string
}

func visitPage(scanCtx, tabCtx context.Context, rec *recorder, in Input, item QueueItem) (visitedPage, bool) {
	mark := rec.mark()
	navCtx, cancel := bound(tabCtx, scanCtx, in.Limits.NavigationTimeout)
	err := chromedp.Do(navCtx, chromedp.Navigate(item.URL))
	cancel()
	if err != nil || scanCtx.Err() != nil {
		return visitedPage{}, false
	}

	waitQuiet(scanCtx, rec, in.Limits.QuietWindow, in.Limits.QuietCap)
	if item.Depth == 0 && !item.FromSitemap {
		dwell(scanCtx, in.Limits.HomepageDwell)
	}
	scroll(scanCtx, tabCtx)
	waitQuiet(scanCtx, rec, in.Limits.QuietWindow, in.Limits.ScrollQuietCap)
	if scanCtx.Err() != nil {
		return visitedPage{}, false
	}

	readCtx, cancelRead := bound(tabCtx, scanCtx, 5*time.Second)
	links, err := chromedp.Run(readCtx, chromedp.Evaluate[[]string](linkJS))
	if err != nil {
		links = nil
	}
	stored, err := chromedp.Run(readCtx, chromedp.Evaluate[storageRead](storageJS))
	cancelRead()
	if err != nil {
		stored = storageRead{}
	}

	jar, _ := browserCookies(tabCtx)
	observations := rec.collect(mark, jar, stored.items())

	return visitedPage{observations: observations, links: links}, true
}

func browserCookies(ctx context.Context) ([]*network.Cookie, error) {
	res, err := chromedp.CallBrowser(ctx, storage.GetCookies, storage.GetCookiesParams{
		BrowserContextID: chromedp.FromContext(ctx).BrowserContextID,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot read cookie jar: %w", err)
	}

	return res.Cookies, nil
}

func waitQuiet(ctx context.Context, rec *recorder, window, cap time.Duration) {
	deadline := time.Now().Add(cap)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		if rec.quietFor(window) || time.Now().After(deadline) {
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func dwell(ctx context.Context, minimum time.Duration) {
	timer := time.NewTimer(minimum)
	defer timer.Stop()

	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

func scroll(ctx context.Context, tabCtx context.Context) {
	for range 12 {
		if ctx.Err() != nil {
			return
		}

		stepCtx, cancel := bound(tabCtx, ctx, time.Second)
		done, err := chromedp.Run(stepCtx, chromedp.Evaluate[bool](scrollJS))
		cancel()
		if err != nil || done {
			return
		}
	}
}

func bound(browser, scan context.Context, limit time.Duration) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(browser)
	timer := time.AfterFunc(limit, cancel)
	stop := context.AfterFunc(scan, cancel)

	return ctx, func() {
		timer.Stop()
		stop()
		cancel()
	}
}

func prepareBrowser(ctx context.Context, target *chromedp.Target, in Input, root string) error {
	if _, err := cdp.Call(ctx, target, network.Enable, network.EnableParams{}); err != nil {
		return fmt.Errorf("cannot enable network events: %w", err)
	}
	if _, err := cdp.Call(ctx, target, runtime.AddBinding, runtime.AddBindingParams{Name: cookieBinding}); err != nil {
		return fmt.Errorf("cannot add cookie binding: %w", err)
	}
	if _, err := cdp.Call(ctx, target, page.AddScriptToEvaluateOnNewDocument, page.AddScriptToEvaluateOnNewDocumentParams{
		Source: cookieHookJS,
	}); err != nil {
		return fmt.Errorf("cannot install cookie hook: %w", err)
	}
	if _, err := cdp.Call(ctx, target, emulation.SetUserAgentOverride, emulation.SetUserAgentOverrideParams{
		UserAgent:      userAgent,
		AcceptLanguage: "en-US,en;q=0.9",
	}); err != nil {
		return fmt.Errorf("cannot set user agent: %w", err)
	}
	if pattern, ok := fetchPattern(in.APIBaseURL); ok {
		if _, err := cdp.Call(ctx, target, fetch.Enable, fetch.EnableParams{
			Patterns: []*fetch.RequestPattern{{
				URLPattern:   pattern,
				RequestStage: fetch.RequestStageRequest,
			}},
		}); err != nil {
			return fmt.Errorf("cannot block sdk posts: %w", err)
		}
	}

	vid := visitorID()
	if script, ok := consentStorageScript(in, root, vid); ok {
		if _, err := cdp.Call(ctx, target, page.AddScriptToEvaluateOnNewDocument, page.AddScriptToEvaluateOnNewDocumentParams{
			Source: script,
		}); err != nil {
			return fmt.Errorf("cannot seed visitor id: %w", err)
		}
	}
	if params, ok := consentCookieParams(in, root, vid); ok {
		if _, err := cdp.Call(ctx, target, network.SetCookie, params); err != nil {
			return fmt.Errorf("cannot seed consent cookie: %w", err)
		}
	}

	return nil
}

func listen(ctx context.Context, target *chromedp.Target, rec *recorder, bannerID string) {
	rangeAppend(cdp.Events(ctx, target, network.RequestWillBeSent), rec)
	rangeAppend(cdp.Events(ctx, target, network.LoadingFinished), rec)
	rangeAppend(cdp.Events(ctx, target, network.LoadingFailed), rec)
	rangeAppend(cdp.Events(ctx, target, network.RequestServedFromCache), rec)
	rangeAppend(cdp.Events(ctx, target, network.ResponseReceivedExtraInfo), rec)
	rangeAppend(cdp.Events(ctx, target, runtime.BindingCalled), rec)

	paused := cdp.Events(ctx, target, fetch.RequestPaused)
	go func() {
		for ev, err := range paused {
			if err != nil {
				return
			}
			go answerFetch(ctx, target, ev, bannerID)
		}
	}()
}

func rangeAppend[E any](events iter.Seq2[E, error], rec *recorder) {
	go func() {
		for ev, err := range events {
			if err != nil {
				return
			}
			rec.append(&ev)
		}
	}()
}

func fetchPattern(apiBase string) (string, bool) {
	origin, ok := requestOrigin(apiBase)
	if !ok {
		return "", false
	}

	return origin + "/*", true
}

func consentStorageScript(in Input, root, vid string) (string, bool) {
	if in.Consent == nil {
		return "", false
	}

	origin, ok := requestOrigin(root)
	if !ok {
		return "", false
	}

	key, _ := json.Marshal(consentCookieName + ":" + in.BannerID + ":vid")
	value, _ := json.Marshal(vid)
	originJSON, _ := json.Marshal(origin)

	return fmt.Sprintf(`(function () {
  try {
    if (location.origin !== %s) return;
    localStorage.setItem(%s, %s);
  } catch (e) {}
})();`, originJSON, key, value), true
}

func consentCookieParams(in Input, root, vid string) (network.SetCookieParams, bool) {
	if in.Consent == nil {
		return network.SetCookieParams{}, false
	}

	data := map[string]bool{}
	for _, slug := range in.Consent.CategorySlugs {
		data[slug] = true
	}
	encoded, err := json.Marshal(map[string]any{
		"bid":    in.BannerID,
		"v":      in.Consent.Version,
		"vid":    vid,
		"action": consentAction,
		"data":   data,
	})
	if err != nil {
		return network.SetCookieParams{}, false
	}

	secure := strings.HasPrefix(root, "https://")

	return network.SetCookieParams{
		Name:     consentCookieName,
		Value:    string(encoded),
		URL:      root,
		Path:     "/",
		Secure:   &secure,
		SameSite: network.CookieSameSiteLax,
	}, true
}

func visitorID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "00000000000000000000000000000000"
	}

	return hex.EncodeToString(buf)
}

func answerFetch(ctx context.Context, target *chromedp.Target, ev fetch.EventRequestPaused, bannerID string) {
	if ev.Request != nil && blockSDKPost(ev.Request.Method, ev.Request.URL, bannerID) {
		_, _ = cdp.Call(ctx, target, fetch.FailRequest, fetch.FailRequestParams{
			RequestID:   ev.RequestID,
			ErrorReason: network.ErrorReasonBlockedByClient,
		})

		return
	}

	_, _ = cdp.Call(ctx, target, fetch.ContinueRequest, fetch.ContinueRequestParams{
		RequestID: ev.RequestID,
	})
}

func checkpointOrEmpty(checkpoint *Checkpoint) Checkpoint {
	if checkpoint == nil {
		return Checkpoint{VisitedTemplates: []string{}, Queue: []QueueItem{}}
	}

	out := *checkpoint
	if out.VisitedTemplates == nil {
		out.VisitedTemplates = []string{}
	}
	if out.Queue == nil {
		out.Queue = []QueueItem{}
	}

	return out
}

func stopReason(parent, scan context.Context) (bool, bool) {
	if parent.Err() != nil {
		return true, false
	}
	if scan.Err() != nil {
		return false, true
	}

	return false, false
}

func stoppedResult(start Checkpoint, elapsed time.Duration, paused, truncated bool) *Result {
	start.ElapsedMS = elapsed.Milliseconds()
	start.Truncated = truncated

	return &Result{
		PagesScanned: start.PagesScanned,
		Truncated:    truncated,
		Paused:       paused,
		Checkpoint:   start,
	}
}

func finish(result *Result, front frontier, elapsed time.Duration, paused, truncated bool) *Result {
	result.Paused = paused
	result.Truncated = truncated
	result.Checkpoint = front.checkpoint(result.PagesScanned, elapsed, truncated)

	return result
}

type storageRead struct {
	Local   []storageKey `json:"local"`
	Session []storageKey `json:"session"`
}

type storageKey struct {
	Key  string `json:"key"`
	Size int    `json:"size"`
}

func (s storageRead) items() []StorageItem {
	items := make([]StorageItem, 0, len(s.Local)+len(s.Session))
	for _, item := range s.Local {
		items = append(items, StorageItem{
			Type:      coredata.TrackerTypeLocalStorage,
			Key:       item.Key,
			ValueSize: item.Size,
		})
	}
	for _, item := range s.Session {
		items = append(items, StorageItem{
			Type:      coredata.TrackerTypeSessionStorage,
			Key:       item.Key,
			ValueSize: item.Size,
		})
	}

	return items
}
