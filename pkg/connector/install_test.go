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

package connector_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/statelesstoken"
)

const (
	installStateSecret  = "install-state-signing-key"
	installStateOrgGID  = "gid:organization:AAAAAAAAAAAAAAAAAAAAAA"
	installStateUserGID = "gid:identity:BBBBBBBBBBBBBBBBBBBBBB"
)

func TestInstallState_RoundTrip(t *testing.T) {
	t.Parallel()

	state, err := connector.NewInstallState(
		installStateSecret,
		"CRISP",
		installStateOrgGID,
		installStateUserGID,
	)
	require.NoError(t, err)

	payload, err := connector.ValidateInstallState(installStateSecret, state)
	require.NoError(t, err)

	assert.Equal(t, "CRISP", payload.Data.Provider)
	assert.Equal(t, installStateOrgGID, payload.Data.OrganizationID)
	assert.Equal(t, installStateUserGID, payload.Data.IdentityID)
	assert.NotEmpty(t, payload.Data.Nonce, "the claim ledger keys on the digest of this token")
}

// TestInstallState_NonceMakesEachMintUnique pins what the single-use claim
// ledger rests on: it keys on the SHA-256 of the whole state, so two mints with
// identical inputs have to produce different tokens. Without the nonce, a
// customer's second tab would collide with their first and be refused as a
// replay.
func TestInstallState_NonceMakesEachMintUnique(t *testing.T) {
	t.Parallel()

	first, err := connector.NewInstallState(
		installStateSecret,
		"CRISP",
		installStateOrgGID,
		installStateUserGID,
	)
	require.NoError(t, err)

	second, err := connector.NewInstallState(
		installStateSecret,
		"CRISP",
		installStateOrgGID,
		installStateUserGID,
	)
	require.NoError(t, err)

	assert.NotEqual(t, first, second)
}

func TestInstallState_Rejects(t *testing.T) {
	t.Parallel()

	valid, err := connector.NewInstallState(
		installStateSecret,
		"CRISP",
		installStateOrgGID,
		installStateUserGID,
	)
	require.NoError(t, err)

	t.Run("another deployment's signing key", func(t *testing.T) {
		t.Parallel()

		_, err := connector.ValidateInstallState("a-different-key", valid)
		require.Error(t, err)
	})

	// The state rides through a third party that can read and rewrite it, so a
	// forged organization or identity must not survive the signature.
	t.Run("a tampered payload", func(t *testing.T) {
		t.Parallel()

		encodedPayload, signature, found := strings.Cut(valid, ".")
		require.True(t, found)

		tampered := encodedPayload[:len(encodedPayload)-1] + "X" + "." + signature

		_, err := connector.ValidateInstallState(installStateSecret, tampered)
		require.Error(t, err)
	})

	t.Run("a stripped signature", func(t *testing.T) {
		t.Parallel()

		encodedPayload, _, found := strings.Cut(valid, ".")
		require.True(t, found)

		_, err := connector.ValidateInstallState(installStateSecret, encodedPayload)
		require.Error(t, err)
	})

	// Every stateless token this deployment signs shares one secret, so the
	// type is the only thing keeping a password-reset or invitation token from
	// being spent as an install state.
	t.Run("a token of another type", func(t *testing.T) {
		t.Parallel()

		foreign, err := statelesstoken.NewToken(
			installStateSecret,
			"probo/some/other/purpose",
			10*time.Minute,
			connector.InstallState{
				Provider:       "CRISP",
				OrganizationID: installStateOrgGID,
				IdentityID:     installStateUserGID,
				Nonce:          "nonce",
			},
		)
		require.NoError(t, err)

		_, err = connector.ValidateInstallState(installStateSecret, foreign)
		require.Error(t, err)
	})

	t.Run("an expired state", func(t *testing.T) {
		t.Parallel()

		expired, err := statelesstoken.NewToken(
			installStateSecret,
			connector.TokenTypeConnectorInstall,
			-time.Minute,
			connector.InstallState{
				Provider:       "CRISP",
				OrganizationID: installStateOrgGID,
				IdentityID:     installStateUserGID,
				Nonce:          "nonce",
			},
		)
		require.NoError(t, err)

		_, err = connector.ValidateInstallState(installStateSecret, expired)
		require.Error(t, err)
	})

	t.Run("an empty state", func(t *testing.T) {
		t.Parallel()

		_, err := connector.ValidateInstallState(installStateSecret, "")
		require.Error(t, err)
	})
}
