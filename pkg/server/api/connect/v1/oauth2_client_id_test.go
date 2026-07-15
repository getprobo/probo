// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package connect_v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOauth2ClientIDFromContinueURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		continueURL string
		want        string
	}{
		{
			name:        "extracts client_id from authorize URL",
			continueURL: "/api/connect/v1/oauth2/authorize?client_id=https%3A%2F%2Ftrust.example.com%2F.well-known%2Foauth-client-metadata&response_type=code",
			want:        "https://trust.example.com/.well-known/oauth-client-metadata",
		},
		{
			name:        "returns empty for non-authorize URL",
			continueURL: "/overview",
			want:        "",
		},
		{
			name:        "returns empty for path suffix lookalike",
			continueURL: "/evil/oauth2/authorize?client_id=phishing",
			want:        "",
		},
		{
			name:        "returns empty when client_id is missing",
			continueURL: "/api/connect/v1/oauth2/authorize?response_type=code",
			want:        "",
		},
		{
			name:        "extracts client_id from absolute authorize URL",
			continueURL: "https://auth.example.com/api/connect/v1/oauth2/authorize?client_id=gid%3A%2F%2Fprobo%2Foauth2_client%2Fabc",
			want:        "gid://probo/oauth2_client/abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, oauth2ClientIDFromContinueURL(tt.continueURL))
		})
	}
}
