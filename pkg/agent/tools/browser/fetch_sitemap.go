// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package browser

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.probo.inc/probo/pkg/agent"
)

type sitemapParams struct {
	URL string `json:"url" jsonschema:"The full URL of the sitemap to fetch (e.g. https://example.com/sitemap.xml)"`
}

type sitemapResult struct {
	Found       bool     `json:"found"`
	URLs        []string `json:"urls,omitempty"`
	URLCount    int      `json:"url_count"`
	ErrorDetail string   `json:"error_detail,omitempty"`
}

const maxSitemapURLs = 200

func FetchSitemapTool() (agent.Tool, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	return agent.FunctionTool[sitemapParams](
		"fetch_sitemap",
		"Fetch and parse a sitemap XML file. Returns discovered URLs which can reveal pages not linked from the main navigation (trust centers, legal docs, status pages).",
		func(ctx context.Context, p sitemapParams) (agent.ToolResult, error) {
			if err := validatePublicURL(p.URL); err != nil {
				data, _ := json.Marshal(sitemapResult{
					Found:       false,
					ErrorDetail: fmt.Sprintf("URL not allowed: %s", err),
				})
				return agent.ToolResult{Content: string(data), IsError: true}, nil
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.URL, nil)
			if err != nil {
				data, _ := json.Marshal(sitemapResult{
					Found:       false,
					ErrorDetail: fmt.Sprintf("cannot create request: %s", err),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}

			resp, err := client.Do(req)
			if err != nil {
				data, _ := json.Marshal(sitemapResult{
					Found:       false,
					ErrorDetail: fmt.Sprintf("cannot fetch sitemap: %s", err),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				data, _ := json.Marshal(sitemapResult{
					Found:       false,
					ErrorDetail: fmt.Sprintf("sitemap returned status %d", resp.StatusCode),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}

			var reader io.Reader = resp.Body
			if strings.HasSuffix(strings.ToLower(p.URL), ".gz") ||
				resp.Header.Get("Content-Encoding") == "gzip" {
				gz, err := gzip.NewReader(resp.Body)
				if err != nil {
					data, _ := json.Marshal(sitemapResult{
						Found:       false,
						ErrorDetail: fmt.Sprintf("cannot decompress gzipped sitemap: %s", err),
					})
					return agent.ToolResult{Content: string(data)}, nil
				}
				defer gz.Close()
				reader = gz
			}

			// Limit read to 5MB.
			reader = io.LimitReader(reader, 5*1024*1024)

			urls, err := parseSitemapXML(reader)
			if err != nil {
				data, _ := json.Marshal(sitemapResult{
					Found:       false,
					ErrorDetail: fmt.Sprintf("cannot parse sitemap XML: %s", err),
				})
				return agent.ToolResult{Content: string(data)}, nil
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

			data, _ := json.Marshal(result)
			return agent.ToolResult{Content: string(data)}, nil
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
