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

package oauth2server_test

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/iam/oauth2server"
)

func computeS256Challenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func TestValidateCodeChallenge(t *testing.T) {
	t.Parallel()

	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := computeS256Challenge(verifier)

	t.Run(
		"valid s256",
		func(t *testing.T) {
			t.Parallel()

			result := oauth2server.ValidateCodeChallenge(
				verifier,
				challenge,
				coredata.OAuth2CodeChallengeMethodS256,
			)

			require.True(t, result)
		},
	)

	t.Run(
		"wrong verifier",
		func(t *testing.T) {
			t.Parallel()

			result := oauth2server.ValidateCodeChallenge(
				"wrong-verifier",
				challenge,
				coredata.OAuth2CodeChallengeMethodS256,
			)

			assert.False(t, result)
		},
	)

	t.Run(
		"wrong challenge",
		func(t *testing.T) {
			t.Parallel()

			result := oauth2server.ValidateCodeChallenge(
				verifier,
				"wrong-challenge",
				coredata.OAuth2CodeChallengeMethodS256,
			)

			assert.False(t, result)
		},
	)

	t.Run(
		"unsupported method",
		func(t *testing.T) {
			t.Parallel()

			result := oauth2server.ValidateCodeChallenge(
				verifier,
				challenge,
				coredata.OAuth2CodeChallengeMethod("plain"),
			)

			assert.False(t, result)
		},
	)

	t.Run(
		"empty method",
		func(t *testing.T) {
			t.Parallel()

			result := oauth2server.ValidateCodeChallenge(
				verifier,
				challenge,
				coredata.OAuth2CodeChallengeMethod(""),
			)

			assert.False(t, result)
		},
	)

	t.Run(
		"empty verifier",
		func(t *testing.T) {
			t.Parallel()

			result := oauth2server.ValidateCodeChallenge(
				"",
				challenge,
				coredata.OAuth2CodeChallengeMethodS256,
			)

			assert.False(t, result)
		},
	)

	t.Run(
		"empty challenge",
		func(t *testing.T) {
			t.Parallel()

			result := oauth2server.ValidateCodeChallenge(
				verifier,
				"",
				coredata.OAuth2CodeChallengeMethodS256,
			)

			assert.False(t, result)
		},
	)

	t.Run(
		"both empty",
		func(t *testing.T) {
			t.Parallel()

			result := oauth2server.ValidateCodeChallenge(
				"",
				"",
				coredata.OAuth2CodeChallengeMethodS256,
			)

			assert.False(t, result)
		},
	)
}
