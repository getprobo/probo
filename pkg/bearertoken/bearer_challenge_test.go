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
