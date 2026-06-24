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

package authn

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignAndVerifySessionTransfer(t *testing.T) {
	t.Parallel()

	secret := "test-secret-key"
	sessionID := "ses_abc123"
	continueURL := "https://custom.example.com/compliance"

	token, err := SignSessionTransfer(sessionID, continueURL, secret)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := VerifySessionTransfer(token, secret)
	require.NoError(t, err)
	assert.Equal(t, sessionID, claims.SessionID)
	assert.Equal(t, continueURL, claims.ContinueURL)
}

func TestVerifySessionTransfer_WrongSecret(t *testing.T) {
	t.Parallel()

	token, err := SignSessionTransfer("ses_abc123", "https://example.com", "secret-a")
	require.NoError(t, err)

	_, err = VerifySessionTransfer(token, "secret-b")
	assert.ErrorIs(t, err, ErrInvalidSessionTransferToken)
}

func TestVerifySessionTransfer_TamperedToken(t *testing.T) {
	t.Parallel()

	token, err := SignSessionTransfer("ses_abc123", "https://example.com", "secret")
	require.NoError(t, err)

	_, err = VerifySessionTransfer(token+"x", "secret")
	assert.ErrorIs(t, err, ErrInvalidSessionTransferToken)
}

func TestVerifySessionTransfer_MalformedToken(t *testing.T) {
	t.Parallel()

	_, err := VerifySessionTransfer("not-a-valid-token", "secret")
	assert.ErrorIs(t, err, ErrInvalidSessionTransferToken)
}

func TestSignSessionTransfer_EmptySecret(t *testing.T) {
	t.Parallel()

	_, err := SignSessionTransfer("ses_abc123", "https://example.com", "")
	assert.Error(t, err)
}
