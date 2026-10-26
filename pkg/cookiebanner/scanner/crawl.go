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
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"go.gearno.de/kit/httpclient"
	"go.probo.inc/probo/pkg/netcheck"
)

var sitemapLoc = regexp.MustCompile(`(?i)<loc>\s*([^<]+)\s*</loc>`)

type frontier struct {
	visited map[string]struct{}
	queued  map[string]struct{}
	next    []QueueItem
	later   []QueueItem
}

func newFrontier(start Checkpoint) frontier {
	f := frontier{
		visited: map[string]struct{}{},
		queued:  map[string]struct{}{},
	}
	for _, key := range start.VisitedTemplates {
		f.visited[key] = struct{}{}
	}
	for _, item := range start.Queue {
		f.add(item)
	}

	return f
}

func (f *frontier) add(item QueueItem) bool {
	key, ok := templateKey(item.URL)
	if !ok {
		return false
	}
	if _, seen := f.visited[key]; seen {
		return false
	}
	if _, waiting := f.queued[key]; waiting {
		return false
	}

	f.queued[key] = struct{}{}
	if item.FromSitemap {
		f.later = append(f.later, item)
	} else {
		f.next = append(f.next, item)
	}

	return true
}

func (f *frontier) pop() (QueueItem, bool) {
	for {
		var (
			item QueueItem
			ok   bool
		)
		if len(f.next) > 0 {
			item, f.next = f.next[0], f.next[1:]
			ok = true
		} else if len(f.later) > 0 {
			item, f.later = f.later[0], f.later[1:]
			ok = true
		}
		if !ok {
			return QueueItem{}, false
		}

		key, valid := templateKey(item.URL)
		if !valid {
			continue
		}
		delete(f.queued, key)
		if _, seen := f.visited[key]; seen {
			continue
		}

		return item, true
	}
}

func (f *frontier) visit(raw string) {
	key, ok := templateKey(raw)
	if !ok {
		return
	}

	f.visited[key] = struct{}{}
	delete(f.queued, key)
}

func (f *frontier) checkpoint(pages int, elapsed time.Duration, truncated bool) Checkpoint {
	visited := make([]string, 0, len(f.visited))
	for key := range f.visited {
		visited = append(visited, key)
	}
	sort.Strings(visited)

	queue := make([]QueueItem, 0, len(f.next)+len(f.later))
	queue = append(queue, f.next...)
	queue = append(queue, f.later...)

	return Checkpoint{
		VisitedTemplates: visited,
		Queue:            queue,
		PagesScanned:     pages,
		ElapsedMS:        elapsed.Milliseconds(),
		Truncated:        truncated,
	}
}

func (in Input) client() *http.Client {
	if in.httpClient != nil {
		return in.httpClient
	}

	return httpclient.DefaultClient(httpclient.WithSSRFProtection())
}

func (in Input) allow(raw string) bool {
	if in.skipPublicURLCheck {
		return true
	}

	return netcheck.ValidatePublicURL(raw) == nil
}

func (in Input) seedSitemaps(ctx context.Context, origin string, f *frontier) {
	seeds := []string{strings.TrimRight(origin, "/") + "/sitemap.xml"}
	seeds = append(seeds, robotsSitemaps(ctx, in.client(), origin)...)

	seen := map[string]struct{}{}
	var urls []string
	for _, seed := range seeds {
		if _, ok := seen[seed]; ok {
			continue
		}
		seen[seed] = struct{}{}
		in.readSitemap(ctx, seed, 1, seen, &urls)
	}

	for _, raw := range urls {
		if !sameOrigin(origin, raw) || !in.allow(raw) {
			continue
		}
		f.add(QueueItem{URL: raw, FromSitemap: true})
	}
}

func robotsSitemaps(ctx context.Context, client *http.Client, origin string) []string {
	body, ok := getText(ctx, client, strings.TrimRight(origin, "/")+"/robots.txt", 1<<20)
	if !ok {
		return nil
	}

	var urls []string
	for _, line := range strings.Split(body, "\n") {
		field, value, found := strings.Cut(line, ":")
		if !found || !strings.EqualFold(strings.TrimSpace(field), "sitemap") {
			continue
		}
		if loc := strings.TrimSpace(value); loc != "" {
			urls = append(urls, loc)
		}
	}

	return urls
}

func (in Input) readSitemap(ctx context.Context, loc string, depth int, seen map[string]struct{}, urls *[]string) {
	if ctx.Err() != nil || len(*urls) >= in.Limits.MaxSitemapURLs || depth > in.Limits.MaxSitemapIndexes {
		return
	}
	if !in.allow(loc) {
		return
	}

	body, ok := getText(ctx, in.client(), loc, int64(in.Limits.MaxSitemapBytes))
	if !ok {
		return
	}

	locs := sitemapLocations(body)
	index := strings.Contains(strings.ToLower(body), "<sitemapindex")
	if index && depth < in.Limits.MaxSitemapIndexes {
		for _, child := range locs {
			if _, ok := seen[child]; ok {
				continue
			}
			seen[child] = struct{}{}
			in.readSitemap(ctx, child, depth+1, seen, urls)
		}

		return
	}

	for _, raw := range locs {
		if len(*urls) >= in.Limits.MaxSitemapURLs {
			return
		}
		*urls = append(*urls, raw)
	}
}

func sitemapLocations(body string) []string {
	matches := sitemapLoc.FindAllStringSubmatch(body, -1)
	locs := make([]string, 0, len(matches))
	for _, match := range matches {
		loc := strings.TrimSpace(match[1])
		loc = strings.ReplaceAll(loc, "&amp;", "&")
		if loc != "" {
			locs = append(locs, loc)
		}
	}

	return locs
}

func getText(ctx context.Context, client *http.Client, raw string, limit int64) (string, bool) {
	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, raw, nil)
	if err != nil {
		return "", false
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", false
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", false
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil || int64(len(body)) > limit {
		return "", false
	}

	return string(body), true
}

func normalizeLink(pageURL, href string) (string, bool) {
	base, err := url.Parse(pageURL)
	if err != nil {
		return "", false
	}

	ref, err := base.Parse(href)
	if err != nil {
		return "", false
	}
	if ref.Scheme != "http" && ref.Scheme != "https" {
		return "", false
	}

	ref.Fragment = ""

	return ref.String(), true
}
