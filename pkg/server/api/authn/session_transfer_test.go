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
