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

package connect_v1

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/iam/oauth2"
)

func TestRedirectWithCode_IncludesIssuer(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "https://auth.example.com/oauth2/authorize", nil)

	redirectWithCode(
		rec,
		req,
		"https://client.example.com/callback",
		"code-value",
		"state-value",
		"https://auth.example.com",
	)

	location, err := url.Parse(rec.Header().Get("Location"))
	require.NoError(t, err)
	assert.Equal(t, "code-value", location.Query().Get("code"))
	assert.Equal(t, "state-value", location.Query().Get("state"))
	assert.Equal(t, "https://auth.example.com", location.Query().Get("iss"))
}

func TestRedirectWithError_IncludesIssuer(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "https://auth.example.com/oauth2/authorize", nil)

	redirectWithError(
		rec,
		req,
		"https://client.example.com/callback",
		"state-value",
		"https://auth.example.com",
		oauth2.ErrAccessDenied,
	)

	location, err := url.Parse(rec.Header().Get("Location"))
	require.NoError(t, err)
	assert.Equal(t, "access_denied", location.Query().Get("error"))
	assert.Equal(t, "state-value", location.Query().Get("state"))
	assert.Equal(t, "https://auth.example.com", location.Query().Get("iss"))
}
