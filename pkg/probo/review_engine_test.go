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

package probo

import "testing"

func TestNormalizeAccountKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		email      string
		externalID string
		want       string
	}{
		{
			name:       "email only",
			email:      "  Jane@Example.com ",
			externalID: "",
			want:       "jane@example.com",
		},
		{
			name:       "email and external id",
			email:      "Jane@Example.com",
			externalID: " 123 ",
			want:       "jane@example.com|123",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := normalizeAccountKey(tt.email, tt.externalID)
			if got != tt.want {
				t.Fatalf("normalizeAccountKey(%q, %q) = %q, want %q", tt.email, tt.externalID, got, tt.want)
			}
		})
	}
}
