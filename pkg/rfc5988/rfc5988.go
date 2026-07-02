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

// Package rfc5988 provides a parser for HTTP Link headers as defined in
// RFC 5988 (Web Linking).
package rfc5988

import "strings"

// Link represents a single entry in an HTTP Link header.
type Link struct {
	URL    string
	Params map[string]string
}

// Parse parses an HTTP Link header value into individual Link entries.
// Each entry has the form <URL>; param1="value1"; param2="value2".
func Parse(header string) []Link {
	if header == "" {
		return nil
	}

	var links []Link

	for part := range strings.SplitSeq(header, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		start := strings.Index(part, "<")

		end := strings.Index(part, ">")
		if start == -1 || end == -1 || end <= start {
			continue
		}

		link := Link{
			URL:    part[start+1 : end],
			Params: make(map[string]string),
		}

		rest := part[end+1:]
		for segment := range strings.SplitSeq(rest, ";") {
			segment = strings.TrimSpace(segment)
			if segment == "" {
				continue
			}

			key, value, ok := strings.Cut(segment, "=")
			if !ok {
				continue
			}

			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)
			value = strings.Trim(value, `"`)

			link.Params[key] = value
		}

		links = append(links, link)
	}

	return links
}

// FindByRel returns the URL of the first Link entry whose "rel" parameter
// matches the given value. It returns an empty string if no match is found.
func FindByRel(header string, rel string) string {
	for _, link := range Parse(header) {
		if link.Params["rel"] == rel {
			return link.URL
		}
	}

	return ""
}
