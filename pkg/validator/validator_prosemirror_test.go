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

package validator

import (
	"strings"
	"testing"
)

func TestProseMirrorDocumentContent(t *testing.T) {
	t.Parallel()

	validDoc := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"hi"}]}]}`

	tests := []struct {
		name      string
		value     any
		wantError bool
	}{
		{"empty string", "", false},
		{"whitespace only", "   \n\t  ", false},
		{"valid doc", validDoc, false},
		{"plain text", "not json", true},
		{"non-doc root", `{"type":"paragraph","content":[]}`, true},
		{"nil value", nil, false},
		{"nil *string", (*string)(nil), false},
		{"non-string", 1, true},
	}

	fn := ProseMirrorDocumentContent()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := fn(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("ProseMirrorDocumentContent() error = %v, wantError %v", err, tt.wantError)
			}

			if err != nil && err.Code != ErrorCodeInvalidFormat {
				t.Errorf("expected code %s, got %s", ErrorCodeInvalidFormat, err.Code)
			}
		})
	}
}

func proseMirrorDoc(text string) string {
	return `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"` + text + `"}]}]}`
}

func TestProseMirrorDocumentMaxTextLength(t *testing.T) {
	t.Parallel()

	const maxLen = 10

	tests := []struct {
		name      string
		value     any
		wantError bool
		wantCode  ErrorCode
	}{
		{"nil value", nil, false, ""},
		{"nil *string", (*string)(nil), false, ""},
		{"empty string", "", false, ""},
		{"whitespace only", "   \n\t  ", false, ""},
		{"under limit", proseMirrorDoc("hello"), false, ""},
		{"at limit", proseMirrorDoc(strings.Repeat("a", maxLen)), false, ""},
		{"over limit", proseMirrorDoc(strings.Repeat("a", maxLen+1)), true, ErrorCodeTooLong},
		{"invalid json skipped", "not json", false, ""},
		{"non-string", 42, true, ErrorCodeInvalidFormat},
	}

	fn := ProseMirrorDocumentMaxTextLength(maxLen)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := fn(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("ProseMirrorDocumentMaxTextLength() error = %v, wantError %v", err, tt.wantError)
			}

			if err != nil && tt.wantCode != "" && err.Code != tt.wantCode {
				t.Errorf("expected code %s, got %s", tt.wantCode, err.Code)
			}
		})
	}
}
