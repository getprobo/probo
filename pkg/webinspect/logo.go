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
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

func FindLogoURL(info *PageInfo) (string, error) {
	head := findElement(info.Root, "head")
	if head == nil {
		return "", fmt.Errorf("cannot find logo: no head element")
	}

	var (
		svgIcon        string
		appleTouchIcon string
		appleTouchSize int
		largestIcon    string
		largestSize    int
		msTileImage    string
	)

	for _, n := range findAllIn(head, "link") {
		rel := strings.ToLower(attrVal(n, "rel"))
		href := attrVal(n, "href")
		if href == "" {
			continue
		}

		switch {
		case strings.Contains(rel, "icon") && !strings.Contains(rel, "apple-touch-icon") && attrVal(n, "type") == "image/svg+xml":
			svgIcon = href
		case strings.Contains(rel, "apple-touch-icon"):
			size := parseSizeAttr(attrVal(n, "sizes"))
			if appleTouchIcon == "" || size > appleTouchSize {
				appleTouchIcon = href
				appleTouchSize = size
			}
		case strings.Contains(rel, "icon") && !strings.Contains(rel, "apple-touch-icon"):
			size := parseSizeAttr(attrVal(n, "sizes"))
			if largestIcon == "" || size > largestSize {
				largestIcon = href
				largestSize = size
			}
		}
	}

	for _, n := range findAllIn(head, "meta") {
		name := strings.ToLower(attrVal(n, "name"))
		content := attrVal(n, "content")
		if name == "msapplication-tileimage" && content != "" {
			msTileImage = content
		}
	}

	candidates := []string{
		svgIcon,
		appleTouchIcon,
		largestIcon,
		msTileImage,
	}

	for _, href := range candidates {
		if href != "" {
			return info.ResolveHref(href), nil
		}
	}

	return "", fmt.Errorf("cannot find logo")
}

func parseSizeAttr(sizes string) int {
	if sizes == "" || strings.EqualFold(sizes, "any") {
		return 0
	}

	best := 0
	for token := range strings.FieldsSeq(sizes) {
		token = strings.ToLower(token)
		parts := strings.SplitN(token, "x", 2)
		if len(parts) != 2 {
			continue
		}

		w, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		if w > best {
			best = w
		}
	}

	return best
}

func ExtensionForMIME(contentType string) string {
	ct := strings.ToLower(contentType)
	if idx := strings.Index(ct, ";"); idx != -1 {
		ct = ct[:idx]
	}
	ct = strings.TrimSpace(ct)

	switch ct {
	case "image/svg+xml":
		return ".svg"
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/x-icon", "image/vnd.microsoft.icon":
		return ".ico"
	default:
		return ".png"
	}
}

// HeadLinks returns all <link> nodes from <head> whose rel attribute
// contains the given value (case-insensitive partial match).
func (p *PageInfo) HeadLinks(rel string) []*html.Node {
	head := findElement(p.Root, "head")
	if head == nil {
		return nil
	}

	rel = strings.ToLower(rel)
	var matches []*html.Node

	for _, n := range findAllIn(head, "link") {
		if strings.Contains(strings.ToLower(attrVal(n, "rel")), rel) {
			matches = append(matches, n)
		}
	}

	return matches
}

// HeadMeta returns the content attribute of the first <meta> tag in <head>
// whose name attribute matches (case-insensitive).
func (p *PageInfo) HeadMeta(name string) (string, bool) {
	head := findElement(p.Root, "head")
	if head == nil {
		return "", false
	}

	name = strings.ToLower(name)

	for _, n := range findAllIn(head, "meta") {
		if strings.EqualFold(attrVal(n, "name"), name) {
			content := attrVal(n, "content")
			if content != "" {
				return content, true
			}
		}
	}

	return "", false
}
