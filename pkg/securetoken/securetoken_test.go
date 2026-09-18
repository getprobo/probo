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

package securetoken_test

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/securetoken"
)

func TestSign_RoundTrip(t *testing.T) {
	t.Parallel()

	signed, err := securetoken.Sign("api-key-id-123", "test-secret")
	require.NoError(t, err)

	value, err := securetoken.Verify(signed, "test-secret")
	require.NoError(t, err)
	assert.Equal(t, "api-key-id-123", value)
}

func TestVerify_ValueWithDots(t *testing.T) {
	t.Parallel()

	values := []string{
		"user@example.com",
		"a.b",
		"a.b.c.d",
		"token.with.dots",
	}

	for _, value := range values {
		t.Run(
			"value "+value,
			func(t *testing.T) {
				t.Parallel()

				signed, err := securetoken.Sign(value, "test-secret")
				require.NoError(t, err)

				got, err := securetoken.Verify(signed, "test-secret")
				require.NoError(t, err)
				assert.Equal(t, value, got)
			},
		)
	}
}

func TestVerify_Tampered(t *testing.T) {
	t.Parallel()

	signed, err := securetoken.Sign("hello", "test-secret")
	require.NoError(t, err)

	_, err = securetoken.Verify(signed+"tamper", "test-secret")
	require.Error(t, err)
}

func TestVerify_WrongSecret(t *testing.T) {
	t.Parallel()

	signed, err := securetoken.Sign("hello", "correct-secret")
	require.NoError(t, err)

	_, err = securetoken.Verify(signed, "wrong-secret")
	require.ErrorIs(t, err, securetoken.ErrInvalidSignature)
}

func TestSignVerify_EmptySecret(t *testing.T) {
	t.Parallel()

	_, err := securetoken.Sign("hello", "")
	require.Error(t, err)

	_, err = securetoken.Verify("hello.sig", "")
	require.Error(t, err)
}

func TestVerify_InvalidFormat(t *testing.T) {
	t.Parallel()

	_, err := securetoken.Verify("no-separator-at-all", "test-secret")
	require.Error(t, err)
}

func TestGet_MissingHeader(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/", nil)

	_, err := securetoken.Get(req, "test-secret")
	require.ErrorIs(t, err, securetoken.ErrTokenNotFound)
}

func TestGet_RoundTrip(t *testing.T) {
	t.Parallel()

	signed, err := securetoken.Sign("api.key.id.with.dots", "test-secret")
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+signed)

	value, err := securetoken.Get(req, "test-secret")
	require.NoError(t, err)
	assert.Equal(t, "api.key.id.with.dots", value)
}

func TestGet_InvalidToken(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer forged.value")

	_, err := securetoken.Get(req, "test-secret")
	require.ErrorIs(t, err, securetoken.ErrInvalidToken)
}
