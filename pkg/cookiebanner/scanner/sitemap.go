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
	"bytes"
	"compress/gzip"
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
)

func (s *scan) enqueueSitemapURLs(ctx context.Context) {
	seeds := []string{s.site.pathURL("/sitemap.xml")}
	seeds = append(seeds, s.sitemapsFromRobots(ctx)...)

	seen := map[string]struct{}{}

	var (
		urls    []string
		fetches int
	)

	for _, seed := range seeds {
		if _, ok := seen[seed]; ok {
			continue
		}

		seen[seed] = struct{}{}
		s.collectSitemapURLs(ctx, seed, 1, seen, &urls, &fetches)
	}

	for _, raw := range urls {
		if !s.site.contains(raw) || s.policy.block(ctx, http.MethodGet, raw, "", "") {
			continue
		}

		s.front.push(QueueItem{URL: raw, FromSitemap: true})
	}
}

func (s *scan) sitemapsFromRobots(ctx context.Context) []string {
	body, ok := s.fetchText(ctx, s.site.pathURL("/robots.txt"), 1<<20)
	if !ok {
		return nil
	}

	var urls []string

	for line := range strings.SplitSeq(body, "\n") {
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

func (s *scan) collectSitemapURLs(
	ctx context.Context,
	loc string,
	depth int,
	seen map[string]struct{},
	urls *[]string,
	fetches *int,
) {
	limits := s.in.Limits
	if ctx.Err() != nil || len(*urls) >= limits.MaxSitemapURLs || *fetches >= limits.MaxSitemapFetches || depth > limits.MaxSitemapIndexes {
		return
	}

	if s.policy.block(ctx, http.MethodGet, loc, string(network.ResourceTypeDocument), "") {
		return
	}

	*fetches++

	body, ok := s.fetchText(ctx, loc, int64(limits.MaxSitemapBytes))
	if !ok {
		return
	}

	locs, index := sitemapLocations(body)
	if index {
		if depth < limits.MaxSitemapIndexes {
			for _, child := range locs {
				if *fetches >= limits.MaxSitemapFetches {
					return
				}

				if _, ok := seen[child]; ok {
					continue
				}

				seen[child] = struct{}{}
				s.collectSitemapURLs(ctx, child, depth+1, seen, urls, fetches)
			}
		}

		return
	}

	for _, raw := range locs {
		if len(*urls) >= limits.MaxSitemapURLs {
			return
		}

		*urls = append(*urls, raw)
	}
}

func sitemapLocations(body string) ([]string, bool) {
	decoder := xml.NewDecoder(strings.NewReader(body))

	var (
		locs  []string
		index bool
	)

	for {
		token, err := decoder.Token()
		if err != nil {
			return locs, index
		}

		start, ok := token.(xml.StartElement)
		if !ok {
			continue
		}

		switch strings.ToLower(start.Name.Local) {
		case "sitemapindex":
			index = true
		case "loc":
			var loc string
			if err := decoder.DecodeElement(&loc, &start); err != nil {
				continue
			}

			if loc = strings.TrimSpace(loc); loc != "" {
				locs = append(locs, loc)
			}
		}
	}
}

func (s *scan) fetchText(ctx context.Context, raw string, limit int64) (string, bool) {
	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, raw, nil)
	if err != nil {
		s.warn(ctx, "cannot fetch text", err)

		return "", false
	}

	resp, err := s.http.Do(req)
	if err != nil {
		s.warn(ctx, "cannot fetch text", err)

		return "", false
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", false
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil || int64(len(body)) > limit {
		if err != nil {
			s.warn(ctx, "cannot read text", err)
		}

		return "", false
	}

	if isGzip(body) {
		decoded, ok := gunzipLimited(body, limit)
		if !ok {
			return "", false
		}

		body = decoded
	}

	return string(body), true
}

func isGzip(body []byte) bool {
	return len(body) >= 2 && body[0] == 0x1f && body[1] == 0x8b
}

func gunzipLimited(body []byte, limit int64) ([]byte, bool) {
	reader, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		return nil, false
	}

	defer func() { _ = reader.Close() }()

	decoded, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil || int64(len(decoded)) > limit {
		return nil, false
	}

	return decoded, true
}
