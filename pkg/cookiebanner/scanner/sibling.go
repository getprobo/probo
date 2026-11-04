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
	"slices"

	"go.probo.inc/probo/pkg/uri"
)

// siblingOrigin reports the canonical origin of raw when it is an http(s)
// URL under the site's ICANN registrable domain but not the site itself
// and not the API. Hosts on a platform suffix share that suffix. The crawl
// does not follow it. www and default ports are folded.
func siblingOrigin(site, api origin, raw string) (string, bool) {
	other, ok := parseOrigin(raw)
	if !ok || other == site || other == api {
		return "", false
	}

	if uri.ICANNRegistrableDomain(other.host) != uri.ICANNRegistrableDomain(site.host) {
		return "", false
	}

	return other.String(), true
}

func (s *scan) noteSibling(raw string) {
	origin, ok := siblingOrigin(s.site, s.api, raw)
	if !ok {
		return
	}

	if s.front.siblings != nil {
		if _, known := s.front.siblings[origin]; known {
			return
		}
	}

	if slices.Contains(s.freshSiblings, origin) {
		return
	}

	known := 0
	if s.front.siblings != nil {
		known = len(s.front.siblings)
	}

	if known+len(s.freshSiblings) >= MaxSiblingOrigins {
		return
	}

	s.freshSiblings = append(s.freshSiblings, origin)
}

func (s *scan) drainSiblings() []string {
	out := s.freshSiblings
	s.freshSiblings = nil

	for _, origin := range out {
		s.front.rememberSibling(origin)
	}

	return out
}
