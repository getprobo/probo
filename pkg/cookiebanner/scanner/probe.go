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
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

var embedTags = map[string]struct{}{
	"audio":  {},
	"embed":  {},
	"frame":  {},
	"iframe": {},
	"script": {},
	"source": {},
	"video":  {},
}

type staticPage struct {
	hosts []string
	links []string
}

// probeEmbeds reads the static HTML of the start page, the pages it links
// to, and the sitemap URLs, and tags their queue entries with the
// third-party hosts they embed. It only orders the queue: the browser
// still records every tracker. Embeds added by JavaScript are not seen.
func (s *scan) probeEmbeds(ctx context.Context, sitemapURLs []string) {
	ctx, cancel := context.WithTimeout(ctx, s.in.Limits.ProbeBudget)
	defer cancel()

	home, err := s.probePage(ctx, s.startURL)
	if errors.Is(err, ErrBlocked) {
		s.warn(ctx, "cannot probe pages", err)

		return
	}

	if len(home.hosts) > 0 {
		s.front.push(QueueItem{URL: s.startURL, EmbedHosts: home.hosts})
	}

	candidates := s.probeCandidates(ctx, home.links, sitemapURLs)
	hosts := s.probeAll(ctx, cancel, candidates)

	for i, item := range candidates {
		if len(hosts[i]) > 0 {
			item.EmbedHosts = hosts[i]
			s.front.push(item)
		}
	}
}

func (s *scan) probeCandidates(ctx context.Context, links, sitemapURLs []string) []QueueItem {
	seen := map[string]struct{}{}
	if key, ok := templateKey(s.startURL); ok {
		seen[key] = struct{}{}
	}

	limit := s.in.Limits.MaxProbes - 1

	var out []QueueItem

	add := func(item QueueItem) {
		key, ok := templateKey(item.URL)
		if !ok || len(out) >= limit {
			return
		}

		if _, dup := seen[key]; dup {
			return
		}

		seen[key] = struct{}{}

		out = append(out, item)
	}

	for _, link := range links {
		if s.site.contains(link) && !s.policy.block(ctx, http.MethodGet, link, "", "") {
			add(QueueItem{URL: link, Depth: 1})
		}
	}

	for _, raw := range sitemapURLs {
		add(QueueItem{URL: raw, FromSitemap: true})
	}

	return out
}

// probeAll fetches candidates with bounded concurrency. The first refusal
// stops the probe, so it does not get the scan itself blocked.
func (s *scan) probeAll(ctx context.Context, stop context.CancelFunc, candidates []QueueItem) [][]string {
	var (
		hosts   = make([][]string, len(candidates))
		jobs    = make(chan int)
		refused sync.Once
		wg      sync.WaitGroup
	)

	for range min(s.in.Limits.ProbeConcurrency, len(candidates)) {
		wg.Go(
			func() {
				for i := range jobs {
					page, err := s.probePage(ctx, candidates[i].URL)
					if errors.Is(err, ErrBlocked) {
						refused.Do(
							func() {
								s.warn(ctx, "cannot probe pages", err)
								stop()
							},
						)

						continue
					}

					hosts[i] = page.hosts
				}
			},
		)
	}

dispatch:
	for i := range candidates {
		select {
		case <-ctx.Done():
			break dispatch
		case jobs <- i:
		}
	}

	close(jobs)
	wg.Wait()

	return hosts
}

func (s *scan) probePage(ctx context.Context, raw string) (staticPage, error) {
	reqCtx, cancel := context.WithTimeout(ctx, s.in.Limits.ProbeTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, raw, nil)
	if err != nil {
		return staticPage{}, fmt.Errorf("cannot create probe request: %w", err)
	}

	resp, err := s.http.Do(req)
	if err != nil {
		return staticPage{}, fmt.Errorf("cannot fetch page: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if blockedStatus(resp.StatusCode) {
		return staticPage{}, fmt.Errorf("%w: HTTP %d", ErrBlocked, resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK || !htmlContent(resp.Header.Get("Content-Type")) {
		return staticPage{}, nil
	}

	body := io.LimitReader(resp.Body, int64(s.in.Limits.MaxProbeBytes))

	return readStaticPage(resp.Request.URL, body, s.site, s.policy.api), nil
}

func htmlContent(header string) bool {
	if header == "" {
		return true
	}

	mediaType, _, err := mime.ParseMediaType(header)

	return err == nil && (mediaType == "text/html" || mediaType == "application/xhtml+xml")
}

// readStaticPage collects the hosts of embedded frames, scripts and media,
// including lazy attributes such as data-src, and the page's links. A body
// cut short by the size cap or a timeout yields what was read.
func readStaticPage(base *url.URL, body io.Reader, site, api origin) staticPage {
	var (
		page  staticPage
		hosts []string
	)

	tokens := html.NewTokenizer(body)

	for {
		switch tokens.Next() {
		case html.ErrorToken:
			page.hosts = unionHosts(nil, hosts)

			return page
		case html.StartTagToken, html.SelfClosingTagToken:
			name, more := tokens.TagName()
			tag := string(name)

			for more {
				var key, value []byte

				key, value, more = tokens.TagAttr()
				attr := string(key)

				switch {
				case tag == "a" && attr == "href":
					if link, ok := resolveLink(base.String(), string(value)); ok {
						page.links = append(page.links, link)
					}
				case embedAttr(tag, attr):
					if host, ok := embedHost(base, string(value), site, api); ok {
						hosts = append(hosts, host)
					}
				}
			}
		}
	}
}

func embedAttr(tag, attr string) bool {
	if tag == "object" {
		return attr == "data"
	}

	_, ok := embedTags[tag]

	return ok && strings.HasSuffix(attr, "src")
}

func embedHost(base *url.URL, raw string, site, api origin) (string, bool) {
	ref, err := base.Parse(strings.TrimSpace(raw))
	if err != nil || (ref.Scheme != "http" && ref.Scheme != "https") || ref.Hostname() == "" {
		return "", false
	}

	if site.contains(ref.String()) || api.contains(ref.String()) {
		return "", false
	}

	return strings.ToLower(ref.Hostname()), true
}

func resourceHosts(resources []Resource) []string {
	var hosts []string

	for _, resource := range resources {
		parsed, err := url.Parse(resource.Origin)
		if err == nil && parsed.Hostname() != "" {
			hosts = append(hosts, strings.ToLower(parsed.Hostname()))
		}
	}

	return unionHosts(nil, hosts)
}
