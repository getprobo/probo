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

// Package webinspect parses a web page's static HTML to extract metadata
// from its <head> (icons/logos, meta tags, link relations). It is a pure,
// deterministic helper: callers supply an http.Client (e.g. one with SSRF
// protection) so the package itself makes no policy decisions about which
// hosts are reachable. The logo discovery is reused by the common
// third-party enricher to populate logo_file_id without an LLM.
package webinspect

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/net/html"
)

type PageInfo struct {
	URL  *url.URL
	Root *html.Node
}

func Parse(ctx context.Context, client *http.Client, websiteURL string) (*PageInfo, error) {
	parsed, err := url.Parse(websiteURL)
	if err != nil {
		return nil, fmt.Errorf("cannot parse website URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, websiteURL, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch page: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cannot fetch page: status %d", resp.StatusCode)
	}

	const maxHTMLSize = 10 << 20 // 10 MiB

	return ParseHTML(parsed, io.LimitReader(resp.Body, maxHTMLSize))
}

func ParseHTML(baseURL *url.URL, r io.Reader) (*PageInfo, error) {
	root, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("cannot parse HTML: %w", err)
	}

	return &PageInfo{URL: baseURL, Root: root}, nil
}

func (p *PageInfo) ResolveHref(href string) string {
	ref, err := url.Parse(href)
	if err != nil {
		return href
	}

	return p.URL.ResolveReference(ref).String()
}

func findElement(n *html.Node, tag string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tag {
		return n
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findElement(c, tag); found != nil {
			return found
		}
	}

	return nil
}

func findAllIn(parent *html.Node, tag string) []*html.Node {
	var nodes []*html.Node

	for c := parent.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == tag {
			nodes = append(nodes, c)
		}

		nodes = append(nodes, findAllIn(c, tag)...)
	}

	return nodes
}

func attrVal(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}

	return ""
}
