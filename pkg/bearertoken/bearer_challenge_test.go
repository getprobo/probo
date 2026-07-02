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

package bearertoken

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
)

func TestBearerChallenge(t *testing.T) {
	t.Parallel()

	baseURL := baseurl.MustParse("https://example.com")

	tests := []struct {
		name      string
		errorCode string
		scopes    []coredata.OAuth2Scope
		want      string
	}{
		{
			name:      "discovery",
			errorCode: "",
			want:      `Bearer resource_metadata="https://example.com/.well-known/oauth-protected-resource"`,
		},
		{
			name:      "invalid token",
			errorCode: BearerErrInvalidToken,
			want:      `Bearer error="invalid_token", resource_metadata="https://example.com/.well-known/oauth-protected-resource"`,
		},
		{
			name:      "insufficient scope",
			errorCode: BearerErrInsufficientScope,
			scopes:    []coredata.OAuth2Scope{"v1:org:read", "v1:privacy"},
			want:      `Bearer error="insufficient_scope", scope="v1:org:read v1:privacy", resource_metadata="https://example.com/.well-known/oauth-protected-resource"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := BearerChallenge(baseURL, tt.errorCode, tt.scopes...)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSetBearerChallenge(t *testing.T) {
	t.Parallel()

	baseURL := baseurl.MustParse("https://example.com")
	rec := httptest.NewRecorder()

	SetBearerInvalidToken(rec, baseURL)

	assert.Equal(
		t,
		`Bearer error="invalid_token", resource_metadata="https://example.com/.well-known/oauth-protected-resource"`,
		rec.Header().Get("WWW-Authenticate"),
	)
}

func TestIsAttempt(t *testing.T) {
	t.Parallel()

	assert.True(t, IsAttempt("Bearer token"))
	assert.True(t, IsAttempt("bearer token"))
	assert.True(t, IsAttempt("BEARER"))
	assert.False(t, IsAttempt(""))
	assert.False(t, IsAttempt("Basic dXNlcjpwYXNz"))
	assert.False(t, IsAttempt("Bear token"))
}
