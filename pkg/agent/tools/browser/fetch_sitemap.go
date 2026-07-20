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

package browser

import (
	"compress/gzip"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.gearno.de/kit/httpclient"
	"go.probo.inc/probo/pkg/agent"
)

type (
	sitemapParams struct {
		URL string `json:"url" jsonschema:"The full URL of the sitemap to fetch (e.g. https://example.com/sitemap.xml)"`
	}

	sitemapResult struct {
		Found       bool     `json:"found"`
		URLs        []string `json:"urls,omitempty"`
		URLCount    int      `json:"url_count"`
		ErrorDetail string   `json:"error_detail,omitempty"`
	}
)

const (
	maxSitemapURLs = 200
)

func FetchSitemapTool() agent.Tool {
	client := httpclient.DefaultPooledClient(httpclient.WithSSRFProtection())
	client.Timeout = 15 * time.Second

	return agent.FunctionTool(
		"fetch_sitemap",
		"Fetch and parse a sitemap XML file. Returns discovered URLs which can reveal pages not linked from the main navigation (compliance portals, legal docs, status pages).",
		func(ctx context.Context, p sitemapParams) (agent.ToolResult, error) {
			if err := validatePublicURL(p.URL); err != nil {
				return agent.ResultJSON(
					sitemapResult{
						Found:       false,
						ErrorDetail: fmt.Sprintf("URL not allowed: %s", err),
					},
				), nil
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.URL, nil)
			if err != nil {
				return agent.ResultJSON(
					sitemapResult{
						Found:       false,
						ErrorDetail: fmt.Sprintf("cannot create request: %s", err),
					},
				), nil
			}

			resp, err := client.Do(req)
			if err != nil {
				return agent.ResultJSON(
					sitemapResult{
						Found:       false,
						ErrorDetail: fmt.Sprintf("cannot fetch sitemap: %s", err),
					},
				), nil
			}

			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				return agent.ResultJSON(
					sitemapResult{
						Found:       false,
						ErrorDetail: fmt.Sprintf("sitemap returned status %d", resp.StatusCode),
					},
				), nil
			}

			var reader io.Reader = resp.Body
			if strings.HasSuffix(strings.ToLower(p.URL), ".gz") ||
				resp.Header.Get("Content-Encoding") == "gzip" {
				gz, err := gzip.NewReader(resp.Body)
				if err != nil {
					return agent.ResultJSON(
						sitemapResult{
							Found:       false,
							ErrorDetail: fmt.Sprintf("cannot decompress gzipped sitemap: %s", err),
						},
					), nil
				}

				defer func() { _ = gz.Close() }()

				reader = gz
			}

			// Limit read to 5MB.
			reader = io.LimitReader(reader, 5*1024*1024)

			urls, err := parseSitemapXML(reader)
			if err != nil {
				return agent.ResultJSON(
					sitemapResult{
						Found:       false,
						ErrorDetail: fmt.Sprintf("cannot parse sitemap XML: %s", err),
					},
				), nil
			}

			result := sitemapResult{
				Found:    true,
				URLCount: len(urls),
			}

			if len(urls) > maxSitemapURLs {
				result.URLs = urls[:maxSitemapURLs]
			} else {
				result.URLs = urls
			}

			return agent.ResultJSON(result), nil
		},
	)
}

func parseSitemapXML(r io.Reader) ([]string, error) {
	var urls []string

	decoder := xml.NewDecoder(r)

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}

		if err != nil {
			return urls, err
		}

		if se, ok := tok.(xml.StartElement); ok && se.Name.Local == "loc" {
			var loc string
			if err := decoder.DecodeElement(&loc, &se); err == nil {
				loc = strings.TrimSpace(loc)
				if loc != "" {
					urls = append(urls, loc)
				}
			}
		}
	}

	return urls, nil
}
