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

package browser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckPDF(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		url       string
		wantError bool
	}{
		{
			name:      "lowercase .pdf returns error",
			url:       "https://example.com/document.pdf",
			wantError: true,
		},
		{
			name:      "uppercase .PDF returns error",
			url:       "https://example.com/document.PDF",
			wantError: true,
		},
		{
			name:      "mixed case .Pdf returns error",
			url:       "https://example.com/document.Pdf",
			wantError: true,
		},
		{
			name:      "normal URL returns nil",
			url:       "https://example.com/page",
			wantError: false,
		},
		{
			name:      "URL with .pdf in path but not at end returns nil",
			url:       "https://example.com/pdf-viewer/document",
			wantError: false,
		},
		{
			name:      "URL with .pdf in query but not at end returns nil",
			url:       "https://example.com/view?file=report.pdf&page=1",
			wantError: false,
		},
		{
			name:      "html URL returns nil",
			url:       "https://example.com/page.html",
			wantError: false,
		},
		{
			name:      "URL ending with .pdf and path segments",
			url:       "https://example.com/files/reports/annual.pdf",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				result := checkPDF(tt.url)

				if tt.wantError {
					require.NotNil(t, result)
					assert.True(t, result.IsError)
					assert.Contains(t, result.Content, "PDF files are not supported")
				} else {
					assert.Nil(t, result)
				}
			},
		)
	}
}
