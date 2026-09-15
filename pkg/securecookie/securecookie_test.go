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

package securecookie_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/securecookie"
)

func TestSign_RoundTrip(t *testing.T) {
	t.Parallel()

	signed, err := securecookie.Sign("session-id-123", "test-secret")
	require.NoError(t, err)

	value, err := securecookie.Verify(signed, "test-secret")
	require.NoError(t, err)
	assert.Equal(t, "session-id-123", value)
}

func TestVerify_ValueWithDots(t *testing.T) {
	t.Parallel()

	values := []string{
		"user@example.com",
		"a.b",
		"a.b.c.d",
		"filename.with.many.dots.pdf",
		"https://example.com/path?q=1&r=2",
	}

	for _, value := range values {
		t.Run(
			"value "+value,
			func(t *testing.T) {
				t.Parallel()

				signed, err := securecookie.Sign(value, "test-secret")
				require.NoError(t, err)

				got, err := securecookie.Verify(signed, "test-secret")
				require.NoError(t, err)
				assert.Equal(t, value, got)
			},
		)
	}
}

func TestVerify_Tampered(t *testing.T) {
	t.Parallel()

	signed, err := securecookie.Sign("hello", "test-secret")
	require.NoError(t, err)

	_, err = securecookie.Verify(signed+"tamper", "test-secret")
	require.Error(t, err)

	_, err = securecookie.Verify("evil."+signed[len(signed)-43:], "test-secret")
	require.Error(t, err)
}

func TestVerify_WrongSecret(t *testing.T) {
	t.Parallel()

	signed, err := securecookie.Sign("hello", "correct-secret")
	require.NoError(t, err)

	_, err = securecookie.Verify(signed, "wrong-secret")
	require.ErrorIs(t, err, securecookie.ErrInvalidSignature)
}

func TestSignVerify_EmptySecret(t *testing.T) {
	t.Parallel()

	_, err := securecookie.Sign("hello", "")
	require.Error(t, err)

	_, err = securecookie.Verify("hello.sig", "")
	require.Error(t, err)
}

func TestVerify_InvalidFormat(t *testing.T) {
	t.Parallel()

	_, err := securecookie.Verify("no-separator-at-all", "test-secret")
	require.Error(t, err)
}

func TestGet_MissingCookie(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	_, err := securecookie.Get(req, securecookie.Config{Name: "SSID", Secret: "test-secret"})
	require.ErrorIs(t, err, securecookie.ErrCookieNotFound)
}

func TestSetGet_RoundTrip(t *testing.T) {
	t.Parallel()

	config := securecookie.Config{Name: "SSID", Secret: "test-secret", Path: "/"}

	recorder := httptest.NewRecorder()
	require.NoError(t, securecookie.Set(recorder, config, "user.123@example.com"))

	resp := recorder.Result()
	defer resp.Body.Close()

	cookies := resp.Cookies()
	require.Len(t, cookies, 1)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookies[0])

	value, err := securecookie.Get(req, config)
	require.NoError(t, err)
	assert.Equal(t, "user.123@example.com", value)
}

func TestGet_InvalidSignature(t *testing.T) {
	t.Parallel()

	config := securecookie.Config{Name: "SSID", Secret: "test-secret"}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "SSID", Value: "forged.value"})

	_, err := securecookie.Get(req, config)
	require.ErrorIs(t, err, securecookie.ErrInvalidCookie)
}
