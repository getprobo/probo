// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
