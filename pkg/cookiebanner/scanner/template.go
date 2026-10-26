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
	"net/url"
	"regexp"
	"strings"
)

var (
	uuidSegment = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	numSegment  = regexp.MustCompile(`^[0-9]+$`)
	hexSegment  = regexp.MustCompile(`(?i)^[0-9a-f]{8,}$`)
	hashSegment = regexp.MustCompile(`^[A-Za-z0-9_-]{20,}$`)
)

// templateKey collapses variable path segments so one URL stands for a
// template. The query and fragment are ignored.
func templateKey(raw string) (string, bool) {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return "", false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", false
	}

	u.RawQuery = ""
	u.Fragment = ""
	u.Path = templatePath(u.Path)

	return u.Scheme + "://" + strings.ToLower(u.Host) + u.Path, true
}

func templatePath(path string) string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return "/"
	}

	parts := strings.Split(trimmed, "/")
	for i, part := range parts {
		if numSegment.MatchString(part) || uuidSegment.MatchString(part) || hexSegment.MatchString(part) || hashSegment.MatchString(part) {
			parts[i] = ":id"
		}
	}

	return "/" + strings.Join(parts, "/")
}

func sameOrigin(origin, raw string) bool {
	base, err := url.Parse(origin)
	if err != nil {
		return false
	}

	u, err := url.Parse(raw)
	if err != nil {
		return false
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}

	return strings.EqualFold(base.Scheme, u.Scheme) && strings.EqualFold(base.Host, u.Host)
}

func originRoot(raw string) (string, bool) {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return "", false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", false
	}

	u.Path = "/"
	u.RawQuery = ""
	u.Fragment = ""

	return u.String(), true
}

func requestOrigin(raw string) (string, bool) {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return "", false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", false
	}

	return u.Scheme + "://" + strings.ToLower(u.Host), true
}
