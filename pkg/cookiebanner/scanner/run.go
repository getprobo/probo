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
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
)

type stopCause int

const (
	causeFinished stopCause = iota
	causePaused
	causeTruncated
	causeBrowser
	causeEvents
)

type scan struct {
	mu             sync.Mutex
	in             Input
	site           origin
	startURL       string
	policy         *requestPolicy
	rec            *recorder
	front          frontier
	http           *http.Client
	parent         context.Context
	scanCtx        context.Context
	tabCtx         context.Context
	cancelTab      context.CancelFunc
	target         *chromedp.Target
	binding        string
	browserContext string
	started        time.Time
	elapsedBase    time.Duration
	pages          int
	firstErr       error
	fatal          error
}

// Run crawls in.Origin in an isolated browser context of allocCtx. A
// cancelled ctx pauses the scan after the last page that finished. The
// scan budget is Limits.ScanTimeout minus checkpoint elapsed time.
func Run(ctx context.Context, allocCtx context.Context, in Input) (*Result, error) {
	in.Limits = in.Limits.withDefaults()

	if in.Origin == "" || in.BannerID == "" {
		return nil, fmt.Errorf("cannot start scan: origin and banner id are required")
	}

	site, ok := parseOrigin(in.Origin)
	if !ok {
		return nil, fmt.Errorf("cannot start scan: origin is invalid")
	}

	api, ok := parseOrigin(in.APIBaseURL)
	if !ok {
		return nil, fmt.Errorf("cannot start scan: API base URL is required")
	}

	policy := newRequestPolicy(site, api, in.skipPublicURLCheck)

	startURL, err := policy.resolveStart(ctx, site)
	if err != nil {
		return nil, err
	}

	start := startingCheckpoint(in.Checkpoint)
	elapsed := time.Duration(start.ElapsedMS) * time.Millisecond
	remaining := in.Limits.ScanTimeout - elapsed

	if ctx.Err() != nil && remaining > 0 && !start.Truncated {
		start.ElapsedMS = elapsed.Milliseconds()

		return &Result{
			PagesScanned: start.PagesScanned,
			Paused:       true,
			Checkpoint:   start,
		}, nil
	}

	if remaining <= 0 || start.Truncated {
		start.Truncated = true
		start.ElapsedMS = elapsed.Milliseconds()

		result := &Result{
			PagesScanned: start.PagesScanned,
			Truncated:    true,
			Checkpoint:   start,
		}
		if start.PagesScanned == 0 {
			return result, ErrNoPageLoaded
		}

		return result, nil
	}

	scanCtx, cancelScan := context.WithTimeout(ctx, remaining)
	defer cancelScan()

	tabCtx, cancelTab := chromedp.NewContext(allocCtx, chromedp.WithNewBrowserContext())
	defer cancelTab()

	go func() {
		select {
		case <-scanCtx.Done():
			cancelTab()
		case <-tabCtx.Done():
		}
	}()

	binding := newBindingName()
	s := &scan{
		in:          in,
		site:        site,
		startURL:    startURL,
		policy:      policy,
		rec:         newRecorder(site, api, in.BannerID, binding, in.Limits.MaxObservations),
		front:       newFrontier(start, in.Limits.MaxQueue),
		http:        in.client(),
		parent:      ctx,
		scanCtx:     scanCtx,
		tabCtx:      tabCtx,
		cancelTab:   cancelTab,
		binding:     binding,
		started:     time.Now(),
		elapsedBase: elapsed,
		pages:       start.PagesScanned,
	}

	if err := chromedp.Do(tabCtx, chromedp.Func(func(actionCtx context.Context, tab *chromedp.Target) error {
		return s.prepareTab(actionCtx, tab)
	})); err != nil {
		if cause, stop := s.stopping(); stop && cause != causeBrowser {
			return s.finish(cause)
		}

		if s.tabCtx.Err() != nil && s.parent.Err() == nil && s.scanCtx.Err() == nil {
			return nil, fmt.Errorf("cannot reach browser: %w", err)
		}

		return nil, fmt.Errorf("cannot prepare browser: %w", err)
	}

	if len(start.VisitedTemplates) == 0 && len(start.Queue) == 0 {
		s.front.push(QueueItem{URL: startURL, Depth: 0})
		s.enqueueSitemapURLs(scanCtx)
	}

	cause, err := s.loop()
	if err != nil {
		return nil, err
	}

	return s.finish(cause)
}

func (in Input) client() *http.Client {
	if in.httpClient != nil {
		return in.httpClient
	}

	return httpclient.DefaultClient(httpclient.WithSSRFProtection())
}

func (s *scan) loop() (stopCause, error) {
	for s.pages < s.in.Limits.MaxPages {
		if cause, stop := s.stopping(); stop {
			return cause, nil
		}

		item, found := s.front.pop()
		if !found {
			return causeFinished, nil
		}

		if !s.site.contains(item.URL) || s.policy.block(s.scanCtx, http.MethodGet, item.URL, "", "") {
			s.front.markVisited(item.URL)

			continue
		}

		page, links, err := s.visit(item)
		if err != nil {
			if cause, stop := s.stopping(); stop {
				if cause == causePaused || cause == causeTruncated {
					s.front.requeue(item)
				}

				return cause, nil
			}

			s.front.markVisited(item.URL)

			if s.firstErr == nil {
				s.firstErr = err
			}

			s.warn(s.scanCtx, "cannot scan page", item.URL, err)

			continue
		}

		s.front.markVisited(item.URL)

		if !item.FromSitemap && item.Depth < s.in.Limits.MaxDepth {
			for _, href := range links {
				link, linkOK := resolveLink(item.URL, href)
				if !linkOK || !s.site.contains(link) || s.policy.block(s.scanCtx, http.MethodGet, link, "", "") {
					continue
				}

				s.front.push(QueueItem{URL: link, Depth: item.Depth + 1})
			}
		}

		s.pages++
		page.URL = item.URL

		checkpoint := s.front.checkpoint(s.pages, s.elapsed(), s.rec.capped())
		if s.in.OnPage != nil {
			if err := s.in.OnPage(context.WithoutCancel(s.parent), page, checkpoint); err != nil {
				return 0, fmt.Errorf("cannot accept scanned page: %w", err)
			}
		}
	}

	return causeFinished, nil
}

func (s *scan) visit(item QueueItem) (Page, []string, error) {
	s.rec.beginPage()

	navCtx, cancel := context.WithTimeout(s.tabCtx, s.in.Limits.NavigationTimeout)
	err := chromedp.Do(navCtx, chromedp.Navigate(item.URL))

	cancel()

	if err != nil {
		if cancelled := s.cancelled(); cancelled != nil {
			return Page{}, nil, cancelled
		}

		return Page{}, nil, classifyNavigation(err)
	}

	s.waitNetworkIdle(s.in.Limits.IdleWindow, s.in.Limits.LoadIdleTimeout)

	if item.Depth == 0 && !item.FromSitemap {
		s.sleep(s.in.Limits.HomepageDwell)
	}

	s.scrollToBottom()
	s.waitNetworkIdle(s.in.Limits.IdleWindow, s.in.Limits.ScrollIdleTimeout)

	if cancelled := s.cancelled(); cancelled != nil {
		return Page{}, nil, cancelled
	}

	if status := s.rec.mainStatus(); blockedStatus(status) {
		return Page{}, nil, fmt.Errorf("%w: HTTP %d", ErrBlocked, status)
	}

	readCtx, cancelRead := context.WithTimeout(s.tabCtx, readTimeout)

	links, err := chromedp.Run(readCtx, chromedp.Evaluate[[]string](linkJS))
	if err != nil {
		s.warn(s.scanCtx, "cannot read links", item.URL, err)

		links = nil
	}

	stored, err := chromedp.Run(readCtx, chromedp.Evaluate[storageRead](storageJS))

	cancelRead()

	if err != nil {
		s.warn(s.scanCtx, "cannot read storage", item.URL, err)

		stored = storageRead{}
	}

	jar, err := s.readCookieJar()
	if err != nil {
		s.warn(s.scanCtx, "cannot read cookie jar", item.URL, err)

		jar = nil
	}

	if cancelled := s.cancelled(); cancelled != nil {
		return Page{}, nil, cancelled
	}

	return s.rec.endPage(jar, stored.items()), links, nil
}

func (s *scan) waitNetworkIdle(window, cap time.Duration) {
	deadline := time.Now().Add(cap)

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		if s.rec.idleFor(window) || time.Now().After(deadline) || s.cancelled() != nil {
			return
		}

		select {
		case <-s.scanCtx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *scan) sleep(minimum time.Duration) {
	timer := time.NewTimer(minimum)
	defer timer.Stop()

	select {
	case <-s.scanCtx.Done():
	case <-timer.C:
	}
}

func (s *scan) scrollToBottom() {
	for range 12 {
		if s.cancelled() != nil {
			return
		}

		stepCtx, cancel := context.WithTimeout(s.tabCtx, time.Second)
		done, err := chromedp.Run(stepCtx, chromedp.Evaluate[bool](scrollJS))

		cancel()

		if err != nil || done {
			return
		}
	}
}

func (s *scan) elapsed() time.Duration {
	return s.elapsedBase + time.Since(s.started)
}

func (s *scan) cancelled() error {
	if s.parent.Err() != nil {
		return s.parent.Err()
	}

	if s.scanCtx.Err() != nil {
		return s.scanCtx.Err()
	}

	if s.tabCtx.Err() != nil {
		return s.tabCtx.Err()
	}

	return nil
}

func (s *scan) stopping() (stopCause, bool) {
	if err := s.fatalErr(); err != nil {
		return causeEvents, true
	}

	if s.parent.Err() != nil {
		return causePaused, true
	}

	if s.scanCtx.Err() != nil {
		return causeTruncated, true
	}

	if s.tabCtx.Err() != nil {
		return causeBrowser, true
	}

	if s.rec.capped() {
		return causeTruncated, true
	}

	return 0, false
}

func (s *scan) fatalErr() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.fatal
}

func (s *scan) finish(cause stopCause) (*Result, error) {
	switch cause {
	case causeBrowser:
		err := s.tabCtx.Err()
		if err == nil {
			err = errors.New("browser context ended")
		}

		return nil, fmt.Errorf("cannot reach browser: %w", err)
	case causeEvents:
		return nil, fmt.Errorf("cannot read browser events: %w", s.fatalErr())
	}

	paused := cause == causePaused
	capped := cause == causeTruncated || s.rec.capped()

	result := &Result{
		PagesScanned: s.pages,
		Paused:       paused,
		Truncated:    !paused && capped,
		Checkpoint:   s.front.checkpoint(s.pages, s.elapsed(), capped),
	}
	if s.pages == 0 && !paused {
		if s.firstErr != nil {
			return result, fmt.Errorf("%w: %w", ErrNoPageLoaded, s.firstErr)
		}

		return result, ErrNoPageLoaded
	}

	return result, nil
}

func (s *scan) warn(ctx context.Context, msg, rawURL string, err error) {
	if s.in.Logger == nil || err == nil {
		return
	}

	path := rawURL
	if parsed, parseErr := url.Parse(rawURL); parseErr == nil {
		path = parsed.Path
	}

	s.in.Logger.WarnCtx(
		ctx,
		msg,
		log.String("path", path),
		log.Error(err),
	)
}

func startingCheckpoint(checkpoint *Checkpoint) Checkpoint {
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

func classifyNavigation(err error) error {
	if load, ok := errors.AsType[*chromedp.LoadError](err); ok {
		if blockedNetError(load.ErrorText) {
			return fmt.Errorf("%w: %s", ErrBlocked, load.ErrorText)
		}

		return fmt.Errorf("%w: %s", ErrUnreachable, load.ErrorText)
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%w: navigation timed out", ErrUnreachable)
	}

	return fmt.Errorf("%w: %s", ErrUnreachable, err.Error())
}

func blockedNetError(text string) bool {
	switch {
	case strings.Contains(text, "ERR_BLOCKED_BY_CLIENT"),
		strings.Contains(text, "ERR_BLOCKED_BY_RESPONSE"),
		strings.Contains(text, "ERR_ACCESS_DENIED"),
		strings.Contains(text, "ERR_ABORTED"):
		return true
	default:
		return false
	}
}

func blockedStatus(status int) bool {
	switch status {
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusTooManyRequests, http.StatusServiceUnavailable:
		return true
	default:
		return false
	}
}
