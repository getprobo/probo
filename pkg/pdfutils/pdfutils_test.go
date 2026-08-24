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

package pdfutils

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddWatermarkWithTimestamp_WatermarkTextTooLong(t *testing.T) {
	t.Parallel()

	watermarkText := strings.Repeat("a", MaxWatermarkTextLength+1)

	pdf, err := AddWatermarkWithTimestamp(nil, "PUBLIC", watermarkText)

	require.Error(t, err)
	assert.Nil(t, pdf)
	assert.ErrorContains(t, err, "watermark text must not exceed 64 bytes")
}

func TestAddWatermarkWithTimestamp_WatermarkTextEmpty(t *testing.T) {
	t.Parallel()

	testCases := map[string]string{
		"empty":              "",
		"ASCII whitespace":   " \t\n",
		"Unicode whitespace": "\u0085",
	}

	for name, watermarkText := range testCases {
		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				pdf, err := AddWatermarkWithTimestamp(nil, "PUBLIC", watermarkText)

				require.Error(t, err)
				assert.Nil(t, pdf)
				assert.ErrorContains(t, err, "watermark text is required")
			},
		)
	}
}

func TestBuildWatermarkLines_UsesDocumentClassification(t *testing.T) {
	t.Parallel()

	lines := buildWatermarkLines(
		"PUBLIC",
		"recipient@example.com",
		time.Date(2026, time.August, 24, 10, 43, 0, 0, time.UTC),
	)

	assert.Equal(t, []string{"PUBLIC", "recipient@example.com", "2026-08-24"}, lines)
}

func TestTruncateWatermarkText(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    string
		expected string
	}{
		"within limit": {
			input:    "recipient@example.com",
			expected: "recipient@example.com",
		},
		"ASCII over limit": {
			input:    strings.Repeat("a", MaxWatermarkTextLength+1),
			expected: strings.Repeat("a", MaxWatermarkTextLength),
		},
		"multibyte rune at boundary": {
			input:    strings.Repeat("a", MaxWatermarkTextLength-1) + "é",
			expected: strings.Repeat("a", MaxWatermarkTextLength-1),
		},
	}

	for name, testCase := range testCases {
		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				actual := TruncateWatermarkText(testCase.input)

				assert.Equal(t, testCase.expected, actual)
			},
		)
	}
}
