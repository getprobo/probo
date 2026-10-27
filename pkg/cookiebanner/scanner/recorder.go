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
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"go.probo.inc/probo/pkg/coredata"
)

type (
	requestMeta struct {
		url       string
		initiator string
		kind      network.ResourceType
		frameID   string
	}

	cookieHint struct {
		source       coredata.CookieSource
		setBy        *string
		initiator    *string
		maxAge       *int
		fallback     Cookie
		haveFallback bool
	}

	recorder struct {
		mu          sync.Mutex
		site        origin
		api         origin
		bannerID    string
		binding     string
		maxObs      int
		session     string
		requests    map[string]requestMeta
		inflight    map[string]struct{}
		earlyFinish map[string]struct{}
		earlyExtra  map[string]network.EventResponseReceivedExtraInfo
		hints       map[string]cookieHint
		seenCookie  map[string]struct{}
		seenRes     map[string]struct{}
		seenStore   map[string]struct{}
		page        Page
		total       int
		truncated   bool
		idleSince   time.Time
		mainFrame   string
		docStatus   int
	}

	cookieBindingPayload struct {
		Name      string `json:"name"`
		Initiator string `json:"initiator"`
		Host      string `json:"host"`
		Domain    string `json:"domain"`
	}
)

func newRecorder(site, api origin, bannerID, binding string, maxObs int) *recorder {
	return &recorder{
		site:        site,
		api:         api,
		bannerID:    bannerID,
		binding:     binding,
		maxObs:      maxObs,
		requests:    map[string]requestMeta{},
		inflight:    map[string]struct{}{},
		earlyFinish: map[string]struct{}{},
		earlyExtra:  map[string]network.EventResponseReceivedExtraInfo{},
		hints:       map[string]cookieHint{},
		seenCookie:  map[string]struct{}{},
		seenRes:     map[string]struct{}{},
		seenStore:   map[string]struct{}{},
	}
}

func (r *recorder) setMainFrame(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.mainFrame = id
}

func (r *recorder) handleEvent(ev any) {
	r.handleSession("", ev)
}

func (r *recorder) handleSession(session string, ev any) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.session = session

	switch ev := ev.(type) {
	case *network.EventRequestWillBeSent:
		r.recordRequest(ev)
	case network.EventRequestWillBeSent:
		r.recordRequest(&ev)
	case *network.EventLoadingFinished:
		r.trackFinish(r.reqKey(ev.RequestID))
	case network.EventLoadingFinished:
		r.trackFinish(r.reqKey(ev.RequestID))
	case *network.EventLoadingFailed:
		r.trackFinish(r.reqKey(ev.RequestID))
	case network.EventLoadingFailed:
		r.trackFinish(r.reqKey(ev.RequestID))
	case *network.EventResponseReceived:
		r.recordStatus(ev)
	case network.EventResponseReceived:
		r.recordStatus(&ev)
	case *network.EventResponseReceivedExtraInfo:
		r.recordExtra(ev)
	case network.EventResponseReceivedExtraInfo:
		r.recordExtra(&ev)
	case *runtime.EventBindingCalled:
		r.recordBinding(ev)
	case runtime.EventBindingCalled:
		r.recordBinding(&ev)
	}
}

func (r *recorder) beginPage() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.requests = map[string]requestMeta{}
	r.inflight = map[string]struct{}{}
	r.earlyFinish = map[string]struct{}{}
	r.earlyExtra = map[string]network.EventResponseReceivedExtraInfo{}
	r.hints = map[string]cookieHint{}
	r.page = Page{}
	r.docStatus = 0
	r.idleSince = time.Time{}
}

func (r *recorder) idleFor(window time.Duration) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	return len(r.inflight) == 0 && !r.idleSince.IsZero() && time.Since(r.idleSince) >= window
}

func (r *recorder) capped() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.truncated
}

func (r *recorder) mainStatus() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.docStatus
}

func (r *recorder) endPage(jar []*network.Cookie, storage []StorageItem) Page {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.recordJar(jar)

	for _, item := range storage {
		r.recordStorage(item)
	}

	for key, hint := range r.hints {
		if hint.haveFallback {
			r.emitCookie(hint.fallback)
		}

		delete(r.hints, key)
	}

	page := r.page
	r.page = Page{}
	r.requests = map[string]requestMeta{}
	r.earlyExtra = map[string]network.EventResponseReceivedExtraInfo{}
	r.earlyFinish = map[string]struct{}{}

	return page
}

func (r *recorder) recordRequest(ev *network.EventRequestWillBeSent) {
	if ev.Request == nil {
		return
	}

	r.trackStart(r.reqKey(ev.RequestID))

	meta := requestMeta{
		url:     ev.Request.URL,
		kind:    ev.Type,
		frameID: string(ev.FrameID),
	}
	if initiator := initiatorURL(ev.Initiator); initiator != "" {
		meta.initiator = initiator
	}

	key := r.reqKey(ev.RequestID)
	r.requests[key] = meta

	if r.mainFrame == "" && ev.Type == network.ResourceTypeDocument && meta.frameID != "" {
		r.mainFrame = meta.frameID
	}

	if extra, ok := r.earlyExtra[key]; ok {
		r.applyExtra(meta, &extra)
		delete(r.earlyExtra, key)
	}

	if r.api.contains(meta.url) || r.site.contains(meta.url) {
		return
	}

	kind, ok := trackerResourceType(ev.Type, meta.frameID, r.mainFrame)
	if !ok {
		return
	}

	resource, ok := newResource(meta.url, meta.initiator, kind)
	if !ok {
		return
	}

	r.emitResource(resource)
}

func (r *recorder) recordStatus(ev *network.EventResponseReceived) {
	if ev.Type != network.ResourceTypeDocument || string(ev.FrameID) != r.mainFrame || ev.Response == nil {
		return
	}

	r.docStatus = int(ev.Response.Status)
}

func (r *recorder) recordExtra(ev *network.EventResponseReceivedExtraInfo) {
	key := r.reqKey(ev.RequestID)

	meta, ok := r.requests[key]
	if !ok {
		if len(r.earlyExtra) < r.maxObs {
			r.earlyExtra[key] = *ev
		}

		return
	}

	r.applyExtra(meta, ev)
}

func (r *recorder) applyExtra(meta requestMeta, ev *network.EventResponseReceivedExtraInfo) {
	if r.api.contains(meta.url) {
		return
	}

	skip := map[string]struct{}{}

	for _, blocked := range ev.BlockedCookies {
		if blocked == nil {
			continue
		}

		if !keepBlocked(blocked.BlockedReasons) {
			skip[blocked.CookieLine] = struct{}{}

			continue
		}

		r.rememberHTTPCookie(blocked.CookieLine, meta.url)
	}

	for _, line := range headerLines(ev.Headers, "set-cookie") {
		if _, blocked := skip[line]; blocked {
			continue
		}

		r.rememberHTTPCookie(line, meta.url)
	}
}

func (r *recorder) recordBinding(ev *runtime.EventBindingCalled) {
	if ev.Name != r.binding {
		return
	}

	var payload cookieBindingPayload
	if err := json.Unmarshal([]byte(ev.Payload), &payload); err != nil {
		return
	}

	if payload.Name == "" || strings.EqualFold(payload.Name, consentCookieName) || len(payload.Name) > maxNameBytes {
		return
	}

	scope := strings.TrimPrefix(payload.Host, ".")
	hostOnly := true

	if payload.Domain != "" {
		scope = strings.TrimPrefix(payload.Domain, ".")
		hostOnly = false
	}

	scope = strings.ToLower(scope)

	cookie := Cookie{
		Name:     payload.Name,
		Source:   coredata.CookieSourceScript,
		HostOnly: new(hostOnly),
	}
	if scope != "" {
		cookie.Domain = new(scope)
	}

	if payload.Initiator != "" {
		if clean := cleanURL(payload.Initiator); clean != "" {
			cookie.InitiatorURL = new(clean)
		}
	}

	r.rememberHint(cookieScopeKey(payload.Name, scope), cookieHint{
		source:       coredata.CookieSourceScript,
		initiator:    cookie.InitiatorURL,
		fallback:     cookie,
		haveFallback: true,
	})
}

func (r *recorder) recordJar(jar []*network.Cookie) {
	for _, item := range jar {
		if item == nil || item.Name == "" || strings.EqualFold(item.Name, consentCookieName) || len(item.Name) > maxNameBytes {
			continue
		}

		if r.apiCookie(item.Domain) {
			continue
		}

		hostOnly := !strings.HasPrefix(item.Domain, ".")
		scope := strings.ToLower(strings.TrimPrefix(item.Domain, "."))
		key := cookieScopeKey(item.Name, scope)

		seen := cookieSeenKey(item.Name, scope, hostOnly)
		if _, ok := r.seenCookie[seen]; ok {
			delete(r.hints, key)

			continue
		}

		source := coredata.CookieSourceScript
		if item.HTTPOnly {
			source = coredata.CookieSourceHTTP
		}

		cookie := Cookie{
			Name:      item.Name,
			Source:    source,
			ValueSize: len(item.Value),
			HTTPOnly:  new(item.HTTPOnly),
			Secure:    new(item.Secure),
			HostOnly:  new(hostOnly),
		}
		if scope != "" {
			cookie.Domain = new(scope)
		}

		if item.PartitionKey != nil || item.PartitionKeyOpaque {
			cookie.Partitioned = new(true)
		}

		switch item.SameSite {
		case network.CookieSameSiteStrict:
			cookie.SameSite = new(coredata.CookieSameSiteStrict)
		case network.CookieSameSiteLax:
			cookie.SameSite = new(coredata.CookieSameSiteLax)
		case network.CookieSameSiteNone:
			cookie.SameSite = new(coredata.CookieSameSiteNone)
		}

		if hint, ok := r.hints[key]; ok {
			cookie.Source = hint.source
			cookie.SetByURL = hint.setBy
			cookie.InitiatorURL = hint.initiator
			cookie.MaxAgeSeconds = hint.maxAge

			delete(r.hints, key)
		}

		r.emitCookie(cookie)
	}
}

func (r *recorder) recordStorage(item StorageItem) {
	if item.Key == "" || strings.HasPrefix(item.Key, consentCookieName+":") || len(item.Key) > maxNameBytes {
		return
	}

	key := string(item.Type) + "\x00" + item.Key
	if _, ok := r.seenStore[key]; ok {
		return
	}

	if !r.reserve() {
		return
	}

	r.seenStore[key] = struct{}{}
	r.page.Storage = append(r.page.Storage, item)
}

func (r *recorder) rememberHTTPCookie(line, setByURL string) {
	cookie, ok := parseSetCookie(line, setByURL)
	if !ok {
		return
	}

	scope := ""
	if cookie.Domain != nil {
		scope = *cookie.Domain
	}

	if cookie.HostOnly != nil && *cookie.HostOnly {
		scope = requestHost(setByURL)
		if scope != "" {
			cookie.Domain = new(scope)
		}
	}

	scope = strings.ToLower(scope)
	if cookie.Domain != nil {
		lower := strings.ToLower(*cookie.Domain)
		cookie.Domain = new(lower)
	}

	r.rememberHint(cookieScopeKey(cookie.Name, scope), cookieHint{
		source:       coredata.CookieSourceHTTP,
		setBy:        cookie.SetByURL,
		maxAge:       cookie.MaxAgeSeconds,
		fallback:     cookie,
		haveFallback: true,
	})
}

func (r *recorder) rememberHint(key string, hint cookieHint) {
	prev, ok := r.hints[key]
	if !ok {
		if len(r.hints) >= r.maxObs {
			r.truncated = true

			return
		}

		r.hints[key] = hint

		return
	}

	if prev.source != coredata.CookieSourceHTTP && hint.source == coredata.CookieSourceHTTP {
		prev.source = hint.source
		prev.setBy = hint.setBy
		prev.maxAge = hint.maxAge
		prev.fallback = hint.fallback
		prev.haveFallback = true
	}

	if prev.initiator == nil && hint.initiator != nil {
		prev.initiator = hint.initiator
	}

	r.hints[key] = prev
}

func (r *recorder) emitCookie(cookie Cookie) {
	hostOnly := cookie.HostOnly != nil && *cookie.HostOnly

	scope := ""
	if cookie.Domain != nil {
		scope = *cookie.Domain
	}

	key := cookieSeenKey(cookie.Name, scope, hostOnly)
	if _, ok := r.seenCookie[key]; ok {
		return
	}

	if !r.reserve() {
		return
	}

	r.seenCookie[key] = struct{}{}
	r.page.Cookies = append(r.page.Cookies, cookie)
}

func (r *recorder) emitResource(resource Resource) {
	if r.api.contains(resource.Origin) || r.site.contains(resource.Origin) {
		return
	}

	key := string(resource.Type) + "\x00" + resource.Origin + resource.Path
	if _, ok := r.seenRes[key]; ok {
		return
	}

	if !r.reserve() {
		return
	}

	r.seenRes[key] = struct{}{}
	r.page.Resources = append(r.page.Resources, resource)
}

func (r *recorder) reqKey(id network.RequestID) string {
	return r.session + "\x00" + string(id)
}

func (r *recorder) dropSession(session string) {
	if session == "" {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	prefix := session + "\x00"
	for key := range r.inflight {
		if strings.HasPrefix(key, prefix) {
			delete(r.inflight, key)
		}
	}

	for key := range r.requests {
		if strings.HasPrefix(key, prefix) {
			delete(r.requests, key)
		}
	}

	for key := range r.earlyFinish {
		if strings.HasPrefix(key, prefix) {
			delete(r.earlyFinish, key)
		}
	}

	for key := range r.earlyExtra {
		if strings.HasPrefix(key, prefix) {
			delete(r.earlyExtra, key)
		}
	}

	if len(r.inflight) == 0 {
		r.idleSince = time.Now()
	}
}

func (r *recorder) trackStart(id string) {
	if _, early := r.earlyFinish[id]; early {
		delete(r.earlyFinish, id)

		if len(r.inflight) == 0 && r.idleSince.IsZero() {
			r.idleSince = time.Now()
		}

		return
	}

	if _, ok := r.inflight[id]; ok {
		return
	}

	r.inflight[id] = struct{}{}
	r.idleSince = time.Time{}
}

func (r *recorder) trackFinish(id string) {
	if _, ok := r.inflight[id]; ok {
		delete(r.inflight, id)

		if len(r.inflight) == 0 {
			r.idleSince = time.Now()
		}

		return
	}

	if _, known := r.requests[id]; known {
		return
	}

	r.earlyFinish[id] = struct{}{}
}

func (r *recorder) reserve() bool {
	if r.total >= r.maxObs {
		r.truncated = true

		return false
	}

	r.total++

	return true
}

func (r *recorder) apiCookie(domain string) bool {
	if r.api.host == "" {
		return false
	}

	return strings.EqualFold(strings.TrimPrefix(domain, "."), r.api.host)
}

func cookieScopeKey(name, scope string) string {
	return strings.ToLower(name) + "\x00" + strings.ToLower(scope)
}

func cookieSeenKey(name, scope string, hostOnly bool) string {
	flag := "d"
	if hostOnly {
		flag = "h"
	}

	return flag + "\x00" + cookieScopeKey(name, scope)
}

func keepBlocked(reasons []network.SetCookieBlockedReason) bool {
	if len(reasons) == 0 {
		return false
	}

	for _, reason := range reasons {
		switch reason {
		case network.SetCookieBlockedReasonUserPreferences, network.SetCookieBlockedReasonThirdPartyPhaseout:
		default:
			return false
		}
	}

	return true
}

func initiatorURL(in *network.Initiator) string {
	if in == nil {
		return ""
	}

	if in.URL != "" {
		return in.URL
	}

	if in.Stack == nil {
		return ""
	}

	for _, frame := range in.Stack.CallFrames {
		if frame == nil || frame.URL == "" {
			continue
		}

		if strings.HasPrefix(frame.URL, "http://") || strings.HasPrefix(frame.URL, "https://") {
			return frame.URL
		}
	}

	return ""
}

func headerLines(headers network.Headers, name string) []string {
	for key, value := range headers {
		if !strings.EqualFold(key, name) {
			continue
		}

		switch v := value.(type) {
		case string:
			return splitHeader(v)
		case []any:
			var lines []string

			for _, item := range v {
				if s, ok := item.(string); ok {
					lines = append(lines, splitHeader(s)...)
				}
			}

			return lines
		}
	}

	return nil
}

func splitHeader(value string) []string {
	parts := strings.Split(value, "\n")

	lines := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			lines = append(lines, part)
		}
	}

	return lines
}

func parseSetCookie(line string, setByURL string) (Cookie, bool) {
	parsed, err := http.ParseSetCookie(line)
	if err != nil || parsed.Name == "" || strings.EqualFold(parsed.Name, consentCookieName) || len(parsed.Name) > maxNameBytes {
		return Cookie{}, false
	}

	cookie := Cookie{
		Name:      parsed.Name,
		Source:    coredata.CookieSourceHTTP,
		ValueSize: len(parsed.Value),
		HTTPOnly:  new(parsed.HttpOnly),
		Secure:    new(parsed.Secure),
	}
	if clean := cleanURL(setByURL); clean != "" {
		cookie.SetByURL = new(clean)
	}

	if parsed.MaxAge > 0 {
		cookie.MaxAgeSeconds = new(parsed.MaxAge)
	} else if parsed.MaxAge < 0 {
		cookie.MaxAgeSeconds = new(0)
	} else if !parsed.Expires.IsZero() {
		seconds := max(int(time.Until(parsed.Expires).Seconds()), 0)
		cookie.MaxAgeSeconds = new(seconds)
	}

	if parsed.Domain == "" {
		cookie.HostOnly = new(true)
	} else {
		domain := strings.TrimPrefix(parsed.Domain, ".")
		cookie.Domain = new(domain)
		cookie.HostOnly = new(false)
	}

	switch parsed.SameSite {
	case http.SameSiteStrictMode:
		cookie.SameSite = new(coredata.CookieSameSiteStrict)
	case http.SameSiteLaxMode:
		cookie.SameSite = new(coredata.CookieSameSiteLax)
	case http.SameSiteNoneMode:
		cookie.SameSite = new(coredata.CookieSameSiteNone)
	}

	if parsed.Partitioned {
		cookie.Partitioned = new(true)
	}

	return cookie, true
}

func trackerResourceType(kind network.ResourceType, frameID, mainFrame string) (coredata.TrackerResourceType, bool) {
	switch kind {
	case network.ResourceTypeScript:
		return coredata.TrackerResourceTypeScript, true
	case network.ResourceTypeStylesheet:
		return coredata.TrackerResourceTypeStylesheet, true
	case network.ResourceTypeImage:
		return coredata.TrackerResourceTypeImage, true
	case network.ResourceTypeMedia, network.ResourceTypeTextTrack:
		return coredata.TrackerResourceTypeMedia, true
	case network.ResourceTypeFont:
		return coredata.TrackerResourceTypeFont, true
	case network.ResourceTypeXHR, network.ResourceTypeFetch:
		return coredata.TrackerResourceTypeFetch, true
	case network.ResourceTypePing:
		return coredata.TrackerResourceTypeBeacon, true
	case network.ResourceTypeDocument:
		if mainFrame != "" && frameID != "" && frameID != mainFrame {
			return coredata.TrackerResourceTypeIframe, true
		}
	}

	return "", false
}

func newResource(rawURL, initiator string, kind coredata.TrackerResourceType) (Resource, bool) {
	origin, path, ok := resourceParts(rawURL)
	if !ok {
		return Resource{}, false
	}

	resource := Resource{
		Type:   kind,
		Origin: origin,
		Path:   path,
	}
	if clean := cleanURL(initiator); clean != "" {
		resource.InitiatorURL = new(clean)
	}

	return resource, true
}

func resourceParts(raw string) (string, string, bool) {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || u.User != nil {
		return "", "", false
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return "", "", false
	}

	host := strings.ToLower(u.Hostname())
	if port := normalizedPort(u.Scheme, u.Port()); port != "" {
		host = net.JoinHostPort(host, port)
	}

	path := u.EscapedPath()
	if path == "" {
		path = "/"
	}

	if len(path) > maxURLBytes {
		return "", "", false
	}

	return u.Scheme + "://" + host, path, true
}

func cleanURL(raw string) string {
	if raw == "" {
		return ""
	}

	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return ""
	}

	u.User = nil
	u.RawQuery = ""
	u.Fragment = ""

	return limitBytes(u.String(), maxURLBytes)
}

func limitBytes(s string, n int) string {
	if len(s) <= n {
		return s
	}

	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}

	return s[:n]
}

func requestHost(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}

	return strings.ToLower(u.Hostname())
}
