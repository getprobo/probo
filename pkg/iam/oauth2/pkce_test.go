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

package oauth2_test

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/iam/oauth2"
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

			result := oauth2.ValidateCodeChallenge(
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

			result := oauth2.ValidateCodeChallenge(
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

			result := oauth2.ValidateCodeChallenge(
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

			result := oauth2.ValidateCodeChallenge(
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

			result := oauth2.ValidateCodeChallenge(
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

			result := oauth2.ValidateCodeChallenge(
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

			result := oauth2.ValidateCodeChallenge(
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

			result := oauth2.ValidateCodeChallenge(
				"",
				"",
				coredata.OAuth2CodeChallengeMethodS256,
			)

			assert.False(t, result)
		},
	)
}
