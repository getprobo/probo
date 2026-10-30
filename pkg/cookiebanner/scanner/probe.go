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
	hosts    []string
	links    []string
	bannerID string
	// complete is true when the HTML was parsed through to the end.
	// A short read leaves it false, so an empty banner id is not
	// proof that the banner is absent.
	complete bool
}

// probeEmbeds reads the static HTML of the start page, the pages it links
// to, and the sitemap URLs, and tags their queue entries with the
// third-party hosts they embed. It only orders the queue: the browser
// still records every tracker. Embeds added by JavaScript are not seen.
// It returns ErrBannerNotInstalled when the homepage HTML was read
// through to the end and has no probo banner, and ErrDifferentBanner
// when it embeds another one, before any other page is fetched.
// confirmed is false when that HTML could not be read through to the
// end; the caller retries next run. A refusal from the site is
// ErrBlocked.
func (s *scan) probeEmbeds(ctx context.Context, sitemapURLs []string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, s.in.Limits.ProbeBudget)
	defer cancel()

	home, err := s.probePage(ctx, s.startURL)
	if home.bannerID != s.in.BannerID {
		gaveUp := ctx.Err() != nil ||
			(s.parent != nil && s.parent.Err() != nil) ||
			(s.scanCtx != nil && s.scanCtx.Err() != nil)

		switch {
		case home.bannerID != "":
			return false, ErrDifferentBanner
		case gaveUp:
			return false, nil
		case errors.Is(err, ErrBlocked):
			return false, err
		case err != nil || !home.complete:
			if err != nil {
				s.warn(ctx, "cannot probe pages", err)
			} else {
				s.warn(ctx, "cannot probe pages", errors.New("homepage HTML was cut short"))
			}

			return false, nil
		default:
			return false, fmt.Errorf("%w: homepage has no probo-cookie-banner", ErrBannerNotInstalled)
		}
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

	return true, nil
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
		// The response finished, and it is not an HTML document, so
		// there is no banner element left to find.
		return staticPage{complete: true}, nil
	}

	page, err := readStaticPage(
		resp.Request.URL,
		io.LimitReader(resp.Body, int64(s.in.Limits.MaxProbeBytes)),
		s.site,
		s.policy.api,
	)
	if err != nil {
		return page, fmt.Errorf("cannot read page: %w", err)
	}

	var extra [1]byte

	n, readErr := resp.Body.Read(extra[:])
	if n > 0 {
		return page, nil
	}

	if readErr != nil && !errors.Is(readErr, io.EOF) {
		return page, fmt.Errorf("cannot read page: %w", readErr)
	}

	page.complete = true

	return page, nil
}

func htmlContent(header string) bool {
	if header == "" {
		return true
	}

	mediaType, _, err := mime.ParseMediaType(header)

	return err == nil && (mediaType == "text/html" || mediaType == "application/xhtml+xml")
}

// readStaticPage collects the hosts of embedded frames, scripts and media,
// including lazy attributes such as data-src, and the page's links. The
// first base href becomes the base for later relative URLs, matching the
// document base URL the browser uses. A read that fails before the end
// returns that error; the caller decides whether the bytes were the
// whole document.
func readStaticPage(base *url.URL, body io.Reader, site, api origin) (staticPage, error) {
	var (
		page    staticPage
		hosts   []string
		doc     = base
		sawBase bool
	)

	tokens := html.NewTokenizer(body)

	for {
		switch tokens.Next() {
		case html.ErrorToken:
			page.hosts = unionHosts(nil, hosts)

			if err := tokens.Err(); err != nil && !errors.Is(err, io.EOF) {
				return page, err
			}

			return page, nil
		case html.StartTagToken, html.SelfClosingTagToken:
			name, more := tokens.TagName()
			tag := string(name)

			for more {
				var key, value []byte

				key, value, more = tokens.TagAttr()
				attr := string(key)

				switch {
				case tag == "probo-cookie-banner" && attr == "banner-id" && page.bannerID == "":
					page.bannerID = strings.TrimSpace(string(value))
				case tag == "base" && attr == "href" && !sawBase:
					sawBase = true

					ref, err := base.Parse(strings.TrimSpace(string(value)))
					if err == nil && (ref.Scheme == "http" || ref.Scheme == "https") && ref.Host != "" {
						doc = ref
					}
				case tag == "a" && attr == "href":
					if link, ok := resolveLink(doc.String(), string(value)); ok {
						page.links = append(page.links, link)
					}
				case embedAttr(tag, attr):
					if host, ok := embedHost(doc, string(value), site, api); ok {
						hosts = addEmbedHost(hosts, host)
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
