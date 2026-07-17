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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMake(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "hello world", input: "Hello World", expected: "hello-world"},
		{name: "this is a test", input: "This is a test", expected: "this-is-a-test"},
		{name: "special characters", input: "Special characters: !@#$%^&*()", expected: "special-characters"},
		{name: "multiple hyphens", input: "Multiple---Hyphens", expected: "multiple-hyphens"},
		{name: "trim hyphens", input: "-Trim-Hyphens-", expected: "trim-hyphens"},
		{name: "numbers", input: "123 Numbers", expected: "123-numbers"},
		{name: "spaces", input: "     Spaces     ", expected: "spaces"},
		{name: "empty", input: "", expected: ""},
		{name: "uppercase", input: "UPPERCASE", expected: "uppercase"},
		{name: "underscore", input: "under_score", expected: "under-score"},
		{name: "dots", input: "dots.and.more.dots", expected: "dotsandmoredots"},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tt.expected, Make(tt.input))
			},
		)
	}
}

func TestMakeWithEntropy(t *testing.T) {
	t.Parallel()

	t.Run(
		"empty input returns hex only",
		func(t *testing.T) {
			t.Parallel()

			assert.Regexp(t, `^[0-9a-f]{8}$`, MakeWithEntropy(""))
		},
	)

	t.Run(
		"non-empty input returns base with hex suffix",
		func(t *testing.T) {
			t.Parallel()

			got := MakeWithEntropy("Hello World")
			assert.Regexp(t, `^hello-world-[0-9a-f]{8}$`, got)
		},
	)

	t.Run(
		"successive calls differ",
		func(t *testing.T) {
			t.Parallel()

			first := MakeWithEntropy("Acme Corp")
			second := MakeWithEntropy("Acme Corp")
			assert.NotEqual(t, first, second, "MakeWithEntropy should produce distinct slugs")
		},
	)
}
