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

// Package urlsafe decides which URLs may be stored in a document or emitted
// into HTML.
package urlsafe

import (
	"net/url"
	"strings"
)

// LinkHref returns a value safe to store on a link and to emit in an HTML href
// attribute. URLs with disallowed schemes (for example javascript: or data:)
// are replaced with "#".
func LinkHref(href string) string {
	href = strings.TrimSpace(href)
	if href == "" {
		return "#"
	}

	if href[0] == '#' {
		return href
	}

	if strings.HasPrefix(href, "/") {
		if len(href) > 1 && (href[1] == '/' || href[1] == '\\') {
			return "#"
		}

		return href
	}

	u, err := url.Parse(href)
	if err != nil {
		return "#"
	}

	if u.Scheme != "" {
		switch strings.ToLower(u.Scheme) {
		case "http", "https", "mailto", "tel":
			return href
		default:
			return "#"
		}
	}

	if u.Host != "" {
		return "#"
	}

	return href
}

// ImageSrc returns a value safe to use in an HTML img src attribute.
// http, https, same-origin paths, and base64 image data URLs are kept.
// javascript: and other schemes become an empty string.
func ImageSrc(src string) string {
	src = strings.TrimSpace(src)
	if src == "" {
		return ""
	}

	if strings.HasPrefix(src, "/") {
		if len(src) > 1 && (src[1] == '/' || src[1] == '\\') {
			return ""
		}

		return src
	}

	u, err := url.Parse(src)
	if err != nil {
		return ""
	}

	if u.Scheme != "" {
		switch strings.ToLower(u.Scheme) {
		case "http", "https":
			return src
		case "data":
			if imageDataURL(src) {
				return src
			}

			return ""
		default:
			return ""
		}
	}

	if u.Host != "" {
		return ""
	}

	return src
}

// PersistedImageSrc is the src stored in a document. Data URLs are dropped so
// an upload preview is never saved.
func PersistedImageSrc(src string) string {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(src)), "data:") {
		return ""
	}

	return ImageSrc(src)
}

func imageDataURL(src string) bool {
	lower := strings.ToLower(strings.TrimSpace(src))
	switch {
	case strings.HasPrefix(lower, "data:image/png;"),
		strings.HasPrefix(lower, "data:image/jpeg;"),
		strings.HasPrefix(lower, "data:image/webp;"),
		strings.HasPrefix(lower, "data:image/gif;"),
		strings.HasPrefix(lower, "data:image/svg+xml;"):
		return strings.Contains(lower, ";base64,")
	default:
		return false
	}
}
