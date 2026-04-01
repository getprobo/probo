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

package validator

import "testing"

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
