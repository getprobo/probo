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

package prosemirror

import (
	"net/url"
	"strings"
)

const linkRelBlankTargetDefault = "noopener noreferrer"

// linkRelToEmit returns the rel attribute value for a link mark, or empty when
// the attribute should be omitted. When target opens a new browsing context
// (_blank), noopener is always injected to prevent the opened page from
// accessing window.opener, even when the document supplies a custom rel.
func linkRelToEmit(attrs LinkAttrs) string {
	blanksTarget := attrs.Target != nil &&
		strings.EqualFold(strings.TrimSpace(*attrs.Target), "_blank")

	if attrs.Rel != nil {
		if s := strings.TrimSpace(*attrs.Rel); s != "" {
			if blanksTarget {
				return ensureNoopener(s)
			}

			return s
		}
	}

	if blanksTarget {
		return linkRelBlankTargetDefault
	}

	return ""
}

// ensureNoopener returns rel unchanged when it already contains the noopener
// token (case-insensitive check). Otherwise it appends " noopener".
func ensureNoopener(rel string) string {
	for tok := range strings.FieldsSeq(rel) {
		if strings.EqualFold(tok, "noopener") {
			return rel
		}
	}

	return rel + " noopener"
}

// safeLinkHref returns a value safe to use in link mark attrs and to emit in an
// HTML href attribute. URLs with disallowed schemes (for example javascript: or
// data:) are replaced with "#" so escaped text content cannot be combined with an
// executable URL.
func safeLinkHref(href string) string {
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

// safeImageSrc returns a value safe to use in an HTML img src attribute.
// http, https, same-origin paths, and base64 image data URLs are kept.
// javascript: and other schemes become an empty string.
func safeImageSrc(src string) string {
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
			if isSafeImageDataURL(src) {
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

// persistedImageSrc is the src stored in a document. Data URLs are dropped so
// an upload preview is never saved; PDF export injects data URLs only in memory.
func persistedImageSrc(src string) string {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(src)), "data:") {
		return ""
	}

	return safeImageSrc(src)
}

func isSafeImageDataURL(src string) bool {
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
