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

package riskmanagement

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapWords(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		width    int
		expected string
	}{
		{
			name:     "empty",
			input:    "",
			width:    28,
			expected: "",
		},
		{
			name:     "short label stays on one line",
			input:    "Store patient records",
			width:    28,
			expected: "Store patient records",
		},
		{
			name:     "exact width boundary does not wrap",
			input:    strings.Repeat("a", 28),
			width:    28,
			expected: strings.Repeat("a", 28),
		},
		{
			name:     "packs words up to width then breaks",
			input:    "Store patient conversation records",
			width:    28,
			expected: "Store patient conversation\nrecords",
		},
		{
			name:     "single oversized word is left intact",
			input:    "Supercalifragilisticexpialidocious",
			width:    28,
			expected: "Supercalifragilisticexpialidocious",
		},
		{
			name:     "non-ascii counted in bytes not runes",
			input:    "café café café café café",
			width:    28,
			expected: "café café café café\ncafé",
		},
		{
			name:     "non-positive width returns input",
			input:    "Store patient conversation records",
			width:    0,
			expected: "Store patient conversation records",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()
				assert.Equal(t, tt.expected, wrapWords(tt.input, tt.width))
			},
		)
	}
}

func TestMermaidEdgeLabel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "short label unchanged",
			input:    "Store patient records",
			expected: "Store patient records",
		},
		{
			name:     "long label uses br breaks",
			input:    "Store patient conversation records",
			expected: "Store patient conversation<br>records",
		},
		{
			name:     "escapes html before inserting br",
			input:    "Store patient <conversation> records now",
			expected: "Store patient<br>&lt;conversation&gt; records<br>now",
		},
		{
			name:     "literal underscore is preserved",
			input:    "Data_Processing pipeline stage one two",
			expected: "Data_Processing pipeline<br>stage one two",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()
				assert.Equal(t, tt.expected, mermaidEdgeLabel(tt.input))
			},
		)
	}
}
