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
	"strings"

	"go.probo.inc/probo/pkg/prosemirror/internal/urlsafe"
)

const linkRelBlankTargetDefault = "noopener noreferrer"

// LinkRelToEmit returns the rel attribute value for a link mark, or empty when
// the attribute should be omitted. When target opens a new browsing context
// (_blank), noopener is always injected to prevent the opened page from
// accessing window.opener, even when the document supplies a custom rel.
func LinkRelToEmit(attrs LinkAttrs) string {
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

// SafeLinkHref returns a value safe to use in link mark attrs and to emit in an
// HTML href attribute. URLs with disallowed schemes (for example javascript: or
// data:) are replaced with "#" so escaped text content cannot be combined with an
// executable URL.
func SafeLinkHref(href string) string {
	return urlsafe.LinkHref(href)
}

func safeImageSrc(src string) string {
	return urlsafe.ImageSrc(src)
}

func persistedImageSrc(src string) string {
	return urlsafe.PersistedImageSrc(src)
}

// SafeImageSrc returns a value safe to emit in an HTML img src attribute.
func SafeImageSrc(src string) string {
	return safeImageSrc(src)
}

// PersistedImageSrc returns the src stored on an image node after sanitize.
func PersistedImageSrc(src string) string {
	return persistedImageSrc(src)
}
