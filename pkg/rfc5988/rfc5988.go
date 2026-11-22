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

	for _, part := range splitOutside(header, ',') {
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

		for _, segment := range splitOutside(part[end+1:], ';') {
			segment = strings.TrimSpace(segment)
			if segment == "" {
				continue
			}

			key, value, ok := strings.Cut(segment, "=")
			if !ok {
				continue
			}

			link.Params[strings.TrimSpace(key)] = unquote(strings.TrimSpace(value))
		}

		links = append(links, link)
	}

	return links
}

// splitOutside splits s at every sep that is not inside an angle-bracketed
// URI or a quoted string. Quoted strings may contain backslash escapes. If a
// quote or angle bracket is never closed, s is split at every sep instead so
// one malformed entry does not swallow the entries after it. A "<" inside an
// angle-bracketed URI counts as unclosed too, since URIs cannot contain it.
func splitOutside(s string, sep byte) []string {
	var (
		parts   []string
		start   int
		inAngle bool
		inQuote bool
	)

	for i := 0; i < len(s); i++ {
		c := s[i]

		switch {
		case inQuote:
			switch c {
			case '\\':
				i++
			case '"':
				inQuote = false
			}
		case inAngle:
			switch c {
			case '<':
				return strings.Split(s, string(sep))
			case '>':
				inAngle = false
			}
		case c == '"':
			inQuote = true
		case c == '<':
			inAngle = true
		case c == sep:
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}

	if inAngle || inQuote {
		return strings.Split(s, string(sep))
	}

	return append(parts, s[start:])
}

// unquote strips the surrounding quotes of a quoted-string and resolves its
// backslash escapes. Values that are not quoted are returned unchanged.
func unquote(value string) string {
	if len(value) < 2 || value[0] != '"' || value[len(value)-1] != '"' {
		return strings.Trim(value, `"`)
	}

	value = value[1 : len(value)-1]
	if !strings.Contains(value, `\`) {
		return value
	}

	var b strings.Builder

	for i := 0; i < len(value); i++ {
		if value[i] == '\\' && i+1 < len(value) {
			i++
		}

		b.WriteByte(value[i])
	}

	return b.String()
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
