// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package slug

import (
	"testing"
)

func TestMake(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello-world"},
		{"This is a test", "this-is-a-test"},
		{"Special characters: !@#$%^&*()", "special-characters"},
		{"Multiple---Hyphens", "multiple-hyphens"},
		{"-Trim-Hyphens-", "trim-hyphens"},
		{"123 Numbers", "123-numbers"},
		{"     Spaces     ", "spaces"},
		{"", ""},
		{"UPPERCASE", "uppercase"},
		{"under_score", "under-score"},
		{"dots.and.more.dots", "dotsandmoredots"},
	}

	for _, test := range tests {
		slug := Make(test.input)
		if slug != test.expected {
			t.Errorf("Slug(%q) = %q; expected %q", test.input, slug, test.expected)
		}
	}
}
