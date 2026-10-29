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
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"go.gearno.de/kit/httpclient"
)

func (s *scan) sitemapURLs(ctx context.Context) []string {
	seeds := []string{s.site.pathURL("/sitemap.xml")}
	seeds = append(seeds, s.sitemapsFromRobots(ctx)...)

	groups := s.sitemapGroups(ctx, seeds)
	kept := make([][]string, 0, len(groups))

	for _, group := range groups {
		var accepted []string

		for _, raw := range group {
			if s.site.contains(raw) && !s.policy.block(ctx, http.MethodGet, raw, "", "") {
				accepted = append(accepted, raw)
			}
		}

		if len(accepted) > 0 {
			kept = append(kept, accepted)
		}
	}

	return arrangeSitemapURLs(kept, s.in.Limits.MaxSitemapURLs)
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

func (s *scan) sitemapGroups(ctx context.Context, seeds []string) [][]string {
	seen := map[string]struct{}{}

	var (
		groups  [][]string
		fetches int
	)

	for _, seed := range seeds {
		if _, ok := seen[seed]; ok {
			continue
		}

		seen[seed] = struct{}{}
		s.collectSitemapGroups(ctx, seed, 1, seen, &groups, &fetches)
	}

	return groups
}

func (s *scan) collectSitemapGroups(
	ctx context.Context,
	loc string,
	depth int,
	seen map[string]struct{},
	groups *[][]string,
	fetches *int,
) {
	limits := s.in.Limits
	if ctx.Err() != nil || *fetches >= limits.MaxSitemapFetches || depth > limits.MaxSitemapIndexes {
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
				s.collectSitemapGroups(ctx, child, depth+1, seen, groups, fetches)
			}
		}

		return
	}

	if limits.MaxSitemapURLs > 0 && len(locs) > limits.MaxSitemapURLs {
		locs = locs[:limits.MaxSitemapURLs]
	}

	if len(locs) > 0 {
		*groups = append(*groups, locs)
	}
}

// arrangeSitemapURLs round-robins child sitemaps into cap so the first
// file cannot fill it.
func arrangeSitemapURLs(groups [][]string, cap int) []string {
	var (
		out     []string
		indexes = make([]int, len(groups))
	)

	for {
		added := false

		for i, group := range groups {
			if cap > 0 && len(out) >= cap {
				return out
			}

			if indexes[i] >= len(group) {
				continue
			}

			out = append(out, group[indexes[i]])
			indexes[i]++
			added = true
		}

		if !added {
			return out
		}
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

// sameSiteClient follows a redirect only when the target is the same
// site as the request. The crawl already treats www and the apex as
// one origin, but the SSRF client refuses any host change, including
// that one. The dial check still rejects a redirect to a private address.
func sameSiteClient(base *http.Client) *http.Client {
	return &http.Client{
		Transport:     base.Transport,
		Jar:           base.Jar,
		Timeout:       base.Timeout,
		CheckRedirect: sameSiteRedirect,
	}
}

func sameSiteRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}

	if len(via) == 0 {
		return nil
	}

	if sameSiteURL(via[0].URL, req.URL) {
		return nil
	}

	return fmt.Errorf("%w: %s -> %s", httpclient.ErrCrossOriginRedirect, via[0].URL.Scheme+"://"+via[0].URL.Host, req.URL.Scheme+"://"+req.URL.Host)
}

func sameSiteURL(original, target *url.URL) bool {
	left, ok := parseOrigin(original.String())
	if !ok {
		return false
	}

	right, ok := parseOrigin(target.String())

	return ok && left == right
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
