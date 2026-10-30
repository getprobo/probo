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
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"strings"

	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/debugger"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/cdproto/storage"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"go.probo.inc/probo/pkg/coredata"
)

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

//go:embed scripts/cookie_hook.js
var cookieHookJS string

//go:embed scripts/visitor_id.js
var visitorIDJS string

//go:embed scripts/links.js
var linkJS string

//go:embed scripts/scroll.js
var scrollJS string

//go:embed scripts/storage.js
var storageJS string

type (
	storageRead struct {
		Local   []storageKey `json:"local"`
		Session []storageKey `json:"session"`
	}

	storageKey struct {
		Key  string `json:"key"`
		Size int    `json:"size"`
	}

	consentCookieValue struct {
		Bid    string          `json:"bid"`
		V      int             `json:"v"`
		Vid    string          `json:"vid"`
		Action string          `json:"action"`
		Data   map[string]bool `json:"data"`
	}
)

func (s *scan) prepareTab(ctx context.Context, tab *chromedp.Target) error {
	s.target = tab
	s.subscribe(tab)

	if _, err := cdp.Call(ctx, tab, network.Enable, network.EnableParams{}); err != nil {
		return fmt.Errorf("cannot enable network events: %w", err)
	}

	if _, err := cdp.Call(ctx, tab, debugger.Enable, debugger.EnableParams{}); err != nil {
		return fmt.Errorf("cannot enable script stacks: %w", err)
	}

	if _, err := cdp.Call(ctx, tab, network.SetBypassServiceWorker, network.SetBypassServiceWorkerParams{Bypass: true}); err != nil {
		return fmt.Errorf("cannot bypass service workers: %w", err)
	}

	tree, err := cdp.Call(ctx, tab, page.GetFrameTree, cdp.Empty{})
	if err != nil || tree.FrameTree == nil || tree.FrameTree.Frame == nil {
		return fmt.Errorf("cannot read frame tree: %w", err)
	}

	s.policy.setMainFrame(string(tree.FrameTree.Frame.ID))
	s.rec.setMainFrame(string(tree.FrameTree.Frame.ID))

	if _, err := cdp.Call(ctx, tab, runtime.AddBinding, runtime.AddBindingParams{Name: s.binding}); err != nil {
		return fmt.Errorf("cannot add cookie binding: %w", err)
	}

	if _, err := cdp.Call(ctx, tab, page.AddScriptToEvaluateOnNewDocument, page.AddScriptToEvaluateOnNewDocumentParams{
		Source: cookieHookSource(s.binding),
	}); err != nil {
		return fmt.Errorf("cannot install cookie hook: %w", err)
	}

	if _, err := cdp.Call(ctx, tab, emulation.SetUserAgentOverride, emulation.SetUserAgentOverrideParams{
		UserAgent:      userAgent,
		AcceptLanguage: "en-US,en;q=0.9",
	}); err != nil {
		return fmt.Errorf("cannot set user agent: %w", err)
	}

	if err := s.denyDownloads(ctx); err != nil {
		return err
	}

	if _, err := cdp.Call(ctx, tab, fetch.Enable, fetch.EnableParams{}); err != nil {
		return fmt.Errorf("cannot intercept browser requests: %w", err)
	}

	if err := s.watchPopups(ctx); err != nil {
		return err
	}

	vid := newVisitorID()
	if script, ok := visitorIDScript(s.in, s.site, vid); ok {
		if _, err := cdp.Call(ctx, tab, page.AddScriptToEvaluateOnNewDocument, page.AddScriptToEvaluateOnNewDocumentParams{
			Source: script,
		}); err != nil {
			return fmt.Errorf("cannot seed visitor id: %w", err)
		}
	}

	// The banner writes a host-only cookie, so one set on the apex is not
	// sent to www. Both hosts are crawled as one site.
	for _, rawURL := range consentCookieURLs(s.site) {
		params, ok, err := consentCookie(s.in, rawURL, vid)
		if err != nil {
			return err
		}

		if !ok {
			break
		}

		if _, err := cdp.Call(ctx, tab, network.SetCookie, params); err != nil {
			return fmt.Errorf("cannot seed consent cookie: %w", err)
		}
	}

	if err := s.watchChildren(ctx); err != nil {
		return err
	}

	return nil
}

func (s *scan) denyDownloads(ctx context.Context) error {
	id := chromedp.FromContext(ctx).BrowserContextID
	if id == "" {
		return fmt.Errorf("cannot deny downloads: browser context is missing")
	}

	if _, err := chromedp.CallBrowser(ctx, browser.SetDownloadBehavior, browser.SetDownloadBehaviorParams{
		Behavior:         browser.SetDownloadBehaviorBehaviorDeny,
		BrowserContextID: id,
	}); err != nil {
		return fmt.Errorf("cannot deny downloads: %w", err)
	}

	return nil
}

func (s *scan) watchPopups(ctx context.Context) error {
	info := chromedp.FromContext(ctx)
	if info == nil || info.Target == nil {
		return fmt.Errorf("cannot watch popups: browser target is missing")
	}

	browserContextID := info.BrowserContextID
	self := info.Target.TargetID
	events := chromedp.BrowserEvents(s.tabCtx, target.TargetCreated)

	if _, err := chromedp.CallBrowser(ctx, target.SetDiscoverTargets, target.SetDiscoverTargetsParams{Discover: true}); err != nil {
		return fmt.Errorf("cannot discover browser targets: %w", err)
	}

	go func() {
		for ev, err := range events {
			if err != nil {
				s.noteStreamErr(err)

				return
			}

			info := ev.TargetInfo
			if info == nil || info.OpenerID == "" || info.Type != "page" {
				continue
			}

			if info.BrowserContextID != browserContextID || info.TargetID == self {
				continue
			}

			_, _ = chromedp.CallBrowser(s.tabCtx, target.CloseTarget, target.CloseTargetParams{TargetID: info.TargetID})
		}
	}()

	return nil
}

func (s *scan) subscribe(tab *chromedp.Target) {
	s.forwardEvents(cdp.Events(s.tabCtx, tab, network.RequestWillBeSent))
	s.forwardEvents(cdp.Events(s.tabCtx, tab, network.LoadingFinished))
	s.forwardEvents(cdp.Events(s.tabCtx, tab, network.LoadingFailed))
	s.forwardEvents(cdp.Events(s.tabCtx, tab, network.ResponseReceived))
	s.forwardEvents(cdp.Events(s.tabCtx, tab, network.ResponseReceivedExtraInfo))
	s.forwardEvents(cdp.Events(s.tabCtx, tab, runtime.BindingCalled))

	paused := cdp.Events(s.tabCtx, tab, fetch.RequestPaused)

	go func() {
		for ev, err := range paused {
			if err != nil {
				s.noteStreamErr(err)

				return
			}

			go s.handlePausedRequest(ev)
		}
	}()
}

func (s *scan) forwardEvents[E any](events iter.Seq2[E, error]) {
	go func() {
		for ev, err := range events {
			if err != nil {
				s.noteStreamErr(err)

				return
			}

			s.rec.handleEvent(ev)
		}
	}()
}

func (s *scan) noteStreamErr(err error) {
	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return
	}

	s.mu.Lock()
	if s.fatal == nil {
		s.fatal = err
	}

	s.mu.Unlock()

	s.cancelTab()
}

func (s *scan) handlePausedRequest(ev fetch.EventRequestPaused) {
	block := ev.Request == nil
	if ev.Request != nil {
		block = s.policy.block(
			s.scanCtx,
			ev.Request.Method,
			ev.Request.URL,
			string(ev.ResourceType),
			string(ev.FrameID),
		)
	}

	if block {
		_, _ = cdp.Call(s.tabCtx, s.target, fetch.FailRequest, fetch.FailRequestParams{
			RequestID:   ev.RequestID,
			ErrorReason: network.ErrorReasonBlockedByClient,
		})

		return
	}

	_, _ = cdp.Call(s.tabCtx, s.target, fetch.ContinueRequest, fetch.ContinueRequestParams{
		RequestID: ev.RequestID,
	})
}

func (s *scan) readCookieJar() ([]*network.Cookie, error) {
	ctx, cancel := context.WithTimeout(s.tabCtx, readTimeout)
	defer cancel()

	info := chromedp.FromContext(s.tabCtx)
	if info == nil || info.BrowserContextID == "" {
		return nil, fmt.Errorf("cannot read cookie jar: browser context is missing")
	}

	res, err := chromedp.CallBrowser(ctx, storage.GetCookies, storage.GetCookiesParams{
		BrowserContextID: info.BrowserContextID,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot read cookie jar: %w", err)
	}

	return res.Cookies, nil
}

func (stored storageRead) items() []StorageItem {
	items := make([]StorageItem, 0, len(stored.Local)+len(stored.Session))
	for _, item := range stored.Local {
		items = append(items, StorageItem{
			Type:      coredata.TrackerTypeLocalStorage,
			Key:       item.Key,
			ValueSize: item.Size,
		})
	}

	for _, item := range stored.Session {
		items = append(items, StorageItem{
			Type:      coredata.TrackerTypeSessionStorage,
			Key:       item.Key,
			ValueSize: item.Size,
		})
	}

	return items
}

func cookieHookSource(binding string) string {
	return strings.ReplaceAll(cookieHookJS, "__PROBO_BINDING__", jsString(binding))
}

func visitorIDScript(in Input, site origin, vid string) (string, bool) {
	if in.Consent == nil {
		return "", false
	}

	script := visitorIDJS
	script = strings.ReplaceAll(script, "__PROBO_APEX__", jsString(site.String()))
	script = strings.ReplaceAll(script, "__PROBO_WWW__", jsString(strings.TrimRight(site.wwwRoot(), "/")))
	script = strings.ReplaceAll(script, "__PROBO_KEY__", jsString(consentCookieName+":"+in.BannerID+":vid"))
	script = strings.ReplaceAll(script, "__PROBO_VALUE__", jsString(vid))

	return script, true
}

func consentCookieURLs(site origin) []string {
	return []string{site.root(), site.wwwRoot()}
}

func consentCookie(in Input, rawURL, vid string) (network.SetCookieParams, bool, error) {
	if in.Consent == nil {
		return network.SetCookieParams{}, false, nil
	}

	data := map[string]bool{}
	for _, slug := range in.Consent.CategorySlugs {
		data[slug] = true
	}

	encoded, err := json.Marshal(consentCookieValue{
		Bid:    in.BannerID,
		V:      in.Consent.Version,
		Vid:    vid,
		Action: consentAction,
		Data:   data,
	})
	if err != nil {
		return network.SetCookieParams{}, false, fmt.Errorf("cannot encode consent cookie: %w", err)
	}

	secure := strings.HasPrefix(rawURL, "https://")

	return network.SetCookieParams{
		Name:     consentCookieName,
		Value:    encodeURIComponent(string(encoded)),
		URL:      rawURL,
		Path:     "/",
		Secure:   &secure,
		SameSite: network.CookieSameSiteLax,
	}, true, nil
}

func encodeURIComponent(s string) string {
	var b strings.Builder

	for i := 0; i < len(s); i++ {
		c := s[i]
		if isURIUnescaped(c) {
			b.WriteByte(c)

			continue
		}

		fmt.Fprintf(&b, "%%%02X", c)
	}

	return b.String()
}

func isURIUnescaped(c byte) bool {
	switch {
	case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c >= '0' && c <= '9':
		return true
	default:
		switch c {
		case '-', '_', '.', '!', '~', '*', '\'', '(', ')':
			return true
		default:
			return false
		}
	}
}

func jsString(s string) string {
	encoded, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}

	return string(encoded)
}

func newVisitorID() string {
	return randomHex(16)
}

func newBindingName() string {
	return "b" + randomHex(16)
}

func randomHex(n int) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}

	return hex.EncodeToString(buf)
}
