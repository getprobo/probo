// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package slug

import (
	"regexp"
	"strings"

	"go.probo.inc/probo/pkg/crypto/rand"
)

var (
	regHyphens     = regexp.MustCompile("-+")
	regSpaces      = regexp.MustCompile(" ")
	regUnderscores = regexp.MustCompile("_")
	regLower       = regexp.MustCompile("[^a-z0-9-]")
)

func Make(s string) string {
	s = strings.ToLower(s)
	s = regSpaces.ReplaceAllString(s, "-")
	s = regUnderscores.ReplaceAllString(s, "-")
	s = regLower.ReplaceAllString(s, "")
	s = regHyphens.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")

	return s
}

func MakeWithEntropy(s string) string {
	const maxDNSLabel = 63

	base := Make(s)
	suffix := rand.MustHexString(4)

	if base == "" {
		return suffix
	}

	// DNS labels are capped at 63 octets. Keep the entropy suffix and
	// truncate the name-derived prefix so hostnames stay provisionable.
	maxBase := maxDNSLabel - 1 - len(suffix)
	if maxBase < 1 {
		return suffix
	}
	if len(base) > maxBase {
		base = strings.Trim(base[:maxBase], "-")
		if base == "" {
			return suffix
		}
	}

	return base + "-" + suffix
}
