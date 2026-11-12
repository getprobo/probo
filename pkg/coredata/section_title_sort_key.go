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

package coredata

import "strings"

// sectionTitleSortKey matches PostgreSQL section_title_sort_key, including
// lpad truncation at 10 digits.
func sectionTitleSortKey(title string) string {
	var b strings.Builder
	b.Grow(len(title))

	rest := title
	for {
		i := strings.IndexFunc(rest, isASCIIDigit)
		if i < 0 {
			b.WriteString(rest)
			return b.String()
		}

		b.WriteString(rest[:i])

		j := i
		for j < len(rest) && isASCIIDigit(rune(rest[j])) {
			j++
		}

		b.WriteString(padSectionTitleNumber(rest[i:j]))
		rest = rest[j:]
	}
}

func isASCIIDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func padSectionTitleNumber(number string) string {
	if len(number) >= 10 {
		return number[:10]
	}

	return strings.Repeat("0", 10-len(number)) + number
}
