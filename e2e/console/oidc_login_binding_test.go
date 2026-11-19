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

package console_test

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestOIDCLogin_RejectsStateFromAnotherClient(t *testing.T) {
	t.Parallel()

	env := testutil.StartIsolatedEnv(
		t,
		testutil.IsolatedEnvOptions{EnableOIDC: true},
	)

	for _, provider := range []string{"google", "microsoft"} {
		t.Run(provider, func(t *testing.T) {
			t.Parallel()

			origin := newOIDCBrowser(t)
			state := startOIDCLogin(t, origin, env.BaseURL, provider)

			rejected := getOIDCCallback(t, newCookielessOIDCBrowser(), env.BaseURL, provider, state)
			assert.Equal(t, http.StatusBadRequest, rejected.StatusCode)
			assert.Empty(t, rejected.Header.Values("Set-Cookie"))

			mismatched := getOIDCCallback(
				t,
				newOIDCBrowserWithNonce(t, env.BaseURL, "different-browser"),
				env.BaseURL,
				provider,
				state,
			)
			assert.Equal(t, http.StatusBadRequest, mismatched.StatusCode)
			assert.Empty(t, mismatched.Header.Values("Set-Cookie"))

			continued := getOIDCCallback(t, origin, env.BaseURL, provider, state)
			assert.Equal(t, http.StatusFound, continued.StatusCode)
			assertAuthError(t, continued, "authentication_failed")
			assertNoSessionCookie(t, continued)

			replayed := getOIDCCallback(t, origin, env.BaseURL, provider, state)
			assert.Equal(t, http.StatusFound, replayed.StatusCode)
			assertAuthError(t, replayed, "invalid_state")
		})
	}
}

func refuseOIDCRedirect(*http.Request, []*http.Request) error {
	return http.ErrUseLastResponse
}

func newOIDCBrowser(t *testing.T) *http.Client {
	t.Helper()

	jar, err := cookiejar.New(nil)
	require.NoError(t, err, "cannot create cookie jar")

	return &http.Client{
		Jar:           jar,
		CheckRedirect: refuseOIDCRedirect,
		Timeout:       45 * time.Second,
	}
}

func newCookielessOIDCBrowser() *http.Client {
	return &http.Client{
		CheckRedirect: refuseOIDCRedirect,
		Timeout:       45 * time.Second,
	}
}

func newOIDCBrowserWithNonce(t *testing.T, baseURL string, nonce string) *http.Client {
	t.Helper()

	client := newOIDCBrowser(t)
	endpoint, err := url.Parse(baseURL)
	require.NoError(t, err)

	client.Jar.SetCookies(
		endpoint,
		[]*http.Cookie{{
			Name:  "OIDC_LOGIN",
			Value: nonce,
			Path:  "/api/connect/v1/oidc",
		}},
	)

	return client
}

func startOIDCLogin(t *testing.T, client *http.Client, baseURL string, provider string) string {
	t.Helper()

	resp := getOIDC(t, client, baseURL+"/api/connect/v1/oidc/"+provider+"/login")
	require.Equal(t, http.StatusFound, resp.StatusCode)

	location, err := url.Parse(resp.Header.Get("Location"))
	require.NoError(t, err)

	state := location.Query().Get("state")
	require.NotEmpty(t, state)

	var binding *http.Cookie

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "OIDC_LOGIN" {
			binding = cookie
		}
	}

	require.NotNil(t, binding)
	assert.True(t, binding.HttpOnly)
	assert.Equal(t, http.SameSiteLaxMode, binding.SameSite)
	assert.Equal(t, "/api/connect/v1/oidc", binding.Path)
	assert.Equal(t, 600, binding.MaxAge)
	assert.NotEmpty(t, binding.Value)
	assert.NotContains(t, resp.Header.Get("Location"), binding.Value)
	assertNoSessionCookie(t, resp)

	return state
}

func getOIDCCallback(t *testing.T, client *http.Client, baseURL string, provider string, state string) *http.Response {
	t.Helper()

	endpoint, err := url.Parse(baseURL + "/api/connect/v1/oidc/" + provider + "/callback")
	require.NoError(t, err)

	query := endpoint.Query()
	query.Set("code", "placeholder_code")
	query.Set("state", state)
	endpoint.RawQuery = query.Encode()

	return getOIDC(t, client, endpoint.String())
}

func getOIDC(t *testing.T, client *http.Client, endpoint string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	})

	return resp
}

func assertAuthError(t *testing.T, resp *http.Response, code string) {
	t.Helper()

	location, err := url.Parse(resp.Header.Get("Location"))
	require.NoError(t, err)
	assert.Equal(t, "/auth/error", location.Path)
	assert.Equal(t, code, location.Query().Get("error"))
}

func assertNoSessionCookie(t *testing.T, resp *http.Response) {
	t.Helper()

	for _, header := range resp.Header.Values("Set-Cookie") {
		assert.Truef(t, strings.HasPrefix(header, "OIDC_LOGIN="), "unexpected cookie %q", header)
	}
}
