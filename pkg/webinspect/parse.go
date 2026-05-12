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
