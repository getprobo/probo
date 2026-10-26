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
	"strings"
	"sync"
	"time"

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

	recorder struct {
		mu         sync.Mutex
		apiOrigin  string
		bannerID   string
		maxObs     int
		events     []any
		requests   map[network.RequestID]requestMeta
		seenCookie map[string]struct{}
		seenRes    map[string]struct{}
		seenStore  map[string]struct{}
		total      int
		truncated  bool
		inflight   int
		quietSince time.Time
		mainFrame  string
	}

	cookieBindingPayload struct {
		Name      string `json:"name"`
		Initiator string `json:"initiator"`
	}
)

func newRecorder(apiBase, bannerID string, maxObs int) *recorder {
	origin, _ := requestOrigin(apiBase)

	return &recorder{
		apiOrigin:  origin,
		bannerID:   bannerID,
		maxObs:     maxObs,
		requests:   map[network.RequestID]requestMeta{},
		seenCookie: map[string]struct{}{},
		seenRes:    map[string]struct{}{},
		seenStore:  map[string]struct{}{},
		quietSince: time.Now(),
	}
}

func (r *recorder) append(ev any) {
	r.mu.Lock()
	defer r.mu.Unlock()

	switch ev.(type) {
	case *network.EventRequestWillBeSent:
		r.inflight++
		r.quietSince = time.Time{}
	case *network.EventLoadingFinished, *network.EventLoadingFailed, *network.EventRequestServedFromCache:
		if r.inflight > 0 {
			r.inflight--
		}
		if r.inflight == 0 {
			r.quietSince = time.Now()
		}
	}

	r.events = append(r.events, ev)
}

func (r *recorder) mark() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.inflight = 0
	r.quietSince = time.Time{}

	return len(r.events)
}

func (r *recorder) quietFor(window time.Duration) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.inflight == 0 && !r.quietSince.IsZero() && time.Since(r.quietSince) >= window
}

func (r *recorder) truncatedObservations() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.truncated
}

func (r *recorder) collect(from int, jar []*network.Cookie, storage []StorageItem) Page {
	r.mu.Lock()
	defer r.mu.Unlock()

	events := append([]any(nil), r.events[from:]...)
	var page Page
	for _, ev := range events {
		switch e := ev.(type) {
		case *network.EventRequestWillBeSent:
			r.noteRequest(e, &page)
		case *network.EventResponseReceivedExtraInfo:
			r.noteSetCookies(e, &page)
		case *runtime.EventBindingCalled:
			r.noteBinding(e, &page)
		}
	}

	r.noteJar(jar, &page)
	for _, item := range storage {
		r.noteStorage(item, &page)
	}

	return page
}

func (r *recorder) noteRequest(ev *network.EventRequestWillBeSent, page *Page) {
	if ev.Request == nil {
		return
	}

	meta := requestMeta{
		url:     ev.Request.URL,
		kind:    ev.Type,
		frameID: string(ev.FrameID),
	}
	if ev.Initiator != nil {
		meta.initiator = ev.Initiator.URL
	}
	r.requests[ev.RequestID] = meta

	if r.mainFrame == "" && ev.Type == network.ResourceTypeDocument && meta.frameID != "" {
		r.mainFrame = meta.frameID
	}
	if r.isAPI(meta.url) {
		return
	}

	kind, ok := mapResourceType(ev.Type, meta.frameID, r.mainFrame)
	if !ok {
		return
	}

	resource, ok := resourceOf(meta.url, meta.initiator, kind)
	if !ok {
		return
	}
	r.addResource(resource, page)
}

func (r *recorder) noteSetCookies(ev *network.EventResponseReceivedExtraInfo, page *Page) {
	meta := r.requests[ev.RequestID]
	if r.isAPI(meta.url) {
		return
	}

	for _, line := range headerLines(ev.Headers, "set-cookie") {
		r.addHTTPCookie(line, meta.url, page)
	}
	for _, blocked := range ev.BlockedCookies {
		if blocked == nil {
			continue
		}
		r.addHTTPCookie(blocked.CookieLine, meta.url, page)
	}
}

func (r *recorder) noteBinding(ev *runtime.EventBindingCalled, page *Page) {
	if ev.Name != cookieBinding {
		return
	}

	var payload cookieBindingPayload
	if err := json.Unmarshal([]byte(ev.Payload), &payload); err != nil {
		return
	}
	if payload.Name == "" || strings.EqualFold(payload.Name, consentCookieName) || r.knownName(payload.Name) {
		return
	}

	cookie := Cookie{
		Name:   payload.Name,
		Source: coredata.CookieSourceScript,
	}
	if payload.Initiator != "" {
		cookie.InitiatorURL = new(payload.Initiator)
	}
	r.addCookie(cookie, page)
}

func (r *recorder) noteJar(jar []*network.Cookie, page *Page) {
	for _, item := range jar {
		if item == nil || item.Name == "" || strings.EqualFold(item.Name, consentCookieName) {
			continue
		}
		if r.knownCookie(item.Name, item.Domain) {
			continue
		}

		source := coredata.CookieSourceScript
		if item.HTTPOnly {
			source = coredata.CookieSourceHTTP
		}
		domain := strings.TrimPrefix(item.Domain, ".")
		cookie := Cookie{
			Name:      item.Name,
			Source:    source,
			ValueSize: len(item.Value),
			HTTPOnly:  new(item.HTTPOnly),
			Secure:    new(item.Secure),
		}
		if domain != "" {
			cookie.Domain = new(domain)
		}
		if item.PartitionKey != nil {
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
		r.addCookie(cookie, page)
	}
}

func (r *recorder) noteStorage(item StorageItem, page *Page) {
	if item.Key == "" || item.Key == consentCookieName+":"+r.bannerID+":vid" {
		return
	}

	key := string(item.Type) + "\x00" + item.Key
	if _, ok := r.seenStore[key]; ok {
		return
	}
	if !r.room() {
		return
	}

	r.seenStore[key] = struct{}{}
	r.total++
	page.Storage = append(page.Storage, item)
}

func (r *recorder) addHTTPCookie(line, setByURL string, page *Page) {
	cookie, ok := mapSetCookie(line, setByURL)
	if !ok {
		return
	}

	r.addCookie(cookie, page)
}

func (r *recorder) addCookie(cookie Cookie, page *Page) {
	key := cookieIdentity(cookie)
	if _, ok := r.seenCookie[key]; ok {
		return
	}
	if !r.room() {
		return
	}

	r.seenCookie[key] = struct{}{}
	r.total++
	page.Cookies = append(page.Cookies, cookie)
}

func (r *recorder) addResource(resource Resource, page *Page) {
	if r.isAPI(resource.Origin) {
		return
	}

	key := string(resource.Type) + "\x00" + resource.Origin + resource.Path
	if _, ok := r.seenRes[key]; ok {
		return
	}
	if !r.room() {
		return
	}

	r.seenRes[key] = struct{}{}
	r.total++
	page.Resources = append(page.Resources, resource)
}

func (r *recorder) knownCookie(name, domain string) bool {
	name = strings.ToLower(name)
	if _, ok := r.seenCookie[name+"\x00"]; ok {
		return true
	}

	domain = strings.ToLower(strings.TrimPrefix(domain, "."))
	_, ok := r.seenCookie[name+"\x00"+domain]

	return ok
}

func (r *recorder) knownName(name string) bool {
	prefix := strings.ToLower(name) + "\x00"
	for key := range r.seenCookie {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}

	return false
}

func (r *recorder) room() bool {
	if r.total < r.maxObs {
		return true
	}

	r.truncated = true

	return false
}

func (r *recorder) isAPI(raw string) bool {
	if r.apiOrigin == "" {
		return false
	}

	origin, ok := requestOrigin(raw)
	if !ok {
		origin = strings.TrimRight(raw, "/")
	}

	return strings.EqualFold(origin, r.apiOrigin)
}

func cookieIdentity(cookie Cookie) string {
	domain := ""
	if cookie.Domain != nil && (cookie.HostOnly == nil || !*cookie.HostOnly) {
		domain = strings.ToLower(*cookie.Domain)
	}

	return strings.ToLower(cookie.Name) + "\x00" + domain
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
