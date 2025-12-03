// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package bearertoken

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		credentials string
		wantToken   string
		wantErr     error
	}{
		// Valid credentials
		{
			name:        "valid simple token",
			credentials: "Bearer abc123",
			wantToken:   "abc123",
			wantErr:     nil,
		},
		{
			name:        "valid uppercase token",
			credentials: "Bearer ABCXYZ",
			wantToken:   "ABCXYZ",
			wantErr:     nil,
		},
		{
			name:        "valid digits only token",
			credentials: "Bearer 0123456789",
			wantToken:   "0123456789",
			wantErr:     nil,
		},
		{
			name:        "valid base64 token with single padding",
			credentials: "Bearer dXNlcm5hbWU6cGFzc3dvcmQ=",
			wantToken:   "dXNlcm5hbWU6cGFzc3dvcmQ=",
			wantErr:     nil,
		},
		{
			name:        "valid base64 token with double padding",
			credentials: "Bearer YWJj==",
			wantToken:   "YWJj==",
			wantErr:     nil,
		},
		{
			name:        "valid token with all special chars",
			credentials: "Bearer abc-._~+/123",
			wantToken:   "abc-._~+/123",
			wantErr:     nil,
		},
		{
			name:        "valid token with hyphen",
			credentials: "Bearer abc-def",
			wantToken:   "abc-def",
			wantErr:     nil,
		},
		{
			name:        "valid token with dot",
			credentials: "Bearer abc.def",
			wantToken:   "abc.def",
			wantErr:     nil,
		},
		{
			name:        "valid token with underscore",
			credentials: "Bearer abc_def",
			wantToken:   "abc_def",
			wantErr:     nil,
		},
		{
			name:        "valid token with tilde",
			credentials: "Bearer abc~def",
			wantToken:   "abc~def",
			wantErr:     nil,
		},
		{
			name:        "valid token with plus",
			credentials: "Bearer abc+def",
			wantToken:   "abc+def",
			wantErr:     nil,
		},
		{
			name:        "valid token with slash",
			credentials: "Bearer abc/def",
			wantToken:   "abc/def",
			wantErr:     nil,
		},
		{
			name:        "valid token with multiple spaces after scheme",
			credentials: "Bearer    token123",
			wantToken:   "token123",
			wantErr:     nil,
		},
		{
			name:        "valid jwt-like token",
			credentials: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
			wantToken:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
			wantErr:     nil,
		},
		{
			name:        "valid single char token",
			credentials: "Bearer a",
			wantToken:   "a",
			wantErr:     nil,
		},

		// Case insensitive scheme
		{
			name:        "lowercase scheme",
			credentials: "bearer abc123",
			wantToken:   "abc123",
			wantErr:     nil,
		},
		{
			name:        "uppercase scheme",
			credentials: "BEARER abc123",
			wantToken:   "abc123",
			wantErr:     nil,
		},
		{
			name:        "mixed case scheme",
			credentials: "BeArEr abc123",
			wantToken:   "abc123",
			wantErr:     nil,
		},

		// Invalid credentials (scheme errors)
		{
			name:        "empty string",
			credentials: "",
			wantToken:   "",
			wantErr:     ErrInvalidCredentials,
		},
		{
			name:        "only scheme without space",
			credentials: "Bearer",
			wantToken:   "",
			wantErr:     ErrInvalidCredentials,
		},
		{
			name:        "scheme without space before token",
			credentials: "Bearerabc123",
			wantToken:   "",
			wantErr:     ErrInvalidCredentials,
		},
		{
			name:        "wrong scheme Basic",
			credentials: "Basic abc123",
			wantToken:   "",
			wantErr:     ErrInvalidCredentials,
		},
		{
			name:        "wrong scheme Digest",
			credentials: "Digest abc123",
			wantToken:   "",
			wantErr:     ErrInvalidCredentials,
		},
		{
			name:        "partial scheme",
			credentials: "Bear abc123",
			wantToken:   "",
			wantErr:     ErrInvalidCredentials,
		},
		{
			name:        "scheme with tab instead of space",
			credentials: "Bearer\tabc123",
			wantToken:   "",
			wantErr:     ErrInvalidCredentials,
		},

		// Missing token
		{
			name:        "missing token after single space",
			credentials: "Bearer ",
			wantToken:   "",
			wantErr:     ErrMissingToken,
		},
		{
			name:        "missing token after multiple spaces",
			credentials: "Bearer    ",
			wantToken:   "",
			wantErr:     ErrMissingToken,
		},

		// Invalid token characters
		{
			name:        "invalid char @",
			credentials: "Bearer abc@123",
			wantToken:   "",
			wantErr:     ErrInvalidToken,
		},
		{
			name:        "invalid char #",
			credentials: "Bearer abc#123",
			wantToken:   "",
			wantErr:     ErrInvalidToken,
		},
		{
			name:        "invalid char !",
			credentials: "Bearer abc!123",
			wantToken:   "",
			wantErr:     ErrInvalidToken,
		},
		{
			name:        "invalid char ?",
			credentials: "Bearer abc?123",
			wantToken:   "",
			wantErr:     ErrInvalidToken,
		},
		{
			name:        "invalid char *",
			credentials: "Bearer abc*123",
			wantToken:   "",
			wantErr:     ErrInvalidToken,
		},
		{
			name:        "invalid char space in token",
			credentials: "Bearer abc 123",
			wantToken:   "",
			wantErr:     ErrInvalidToken,
		},
		{
			name:        "invalid char tab in token",
			credentials: "Bearer abc\t123",
			wantToken:   "",
			wantErr:     ErrInvalidToken,
		},
		{
			name:        "invalid char newline in token",
			credentials: "Bearer abc\n123",
			wantToken:   "",
			wantErr:     ErrInvalidToken,
		},

		// Invalid padding
		{
			name:        "token starting with equals",
			credentials: "Bearer =abc",
			wantToken:   "",
			wantErr:     ErrInvalidToken,
		},
		{
			name:        "token with equals in middle",
			credentials: "Bearer abc=def",
			wantToken:   "",
			wantErr:     ErrInvalidToken,
		},
		{
			name:        "token only padding",
			credentials: "Bearer ==",
			wantToken:   "",
			wantErr:     ErrInvalidToken,
		},
		{
			name:        "token with char after padding",
			credentials: "Bearer abc==def",
			wantToken:   "",
			wantErr:     ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				gotToken, gotErr := Parse(tt.credentials)

				if tt.wantErr != nil {
					require.ErrorIs(t, gotErr, tt.wantErr)
					assert.Empty(t, gotToken)
				} else {
					require.NoError(t, gotErr)
					assert.Equal(t, tt.wantToken, gotToken)
				}
			},
		)
	}
}
