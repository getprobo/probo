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

package drivers

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseIdentityCenterLoginEvent(t *testing.T) {
	t.Parallel()

	t.Run(
		"reads user id and totp from additional event data",
		func(t *testing.T) {
			t.Parallel()

			userID, credentialType, ok := parseIdentityCenterLoginEvent(
				`{"userIdentity":{"onBehalfOf":{"userId":"11111111-1111-1111-1111-111111111111"}},"additionalEventData":{"CredentialType":"PASSWORD,TOTP"}}`,
			)
			require.True(t, ok)
			assert.Equal(t, "11111111-1111-1111-1111-111111111111", userID)
			assert.Equal(t, "PASSWORD,TOTP", credentialType)
		},
	)

	t.Run(
		"treats password-only as no mfa signal",
		func(t *testing.T) {
			t.Parallel()

			userID, credentialType, ok := parseIdentityCenterLoginEvent(
				`{"userIdentity":{"onBehalfOf":{"userId":"carol"}},"additionalEventData":{"CredentialType":"PASSWORD"}}`,
			)
			require.True(t, ok)
			assert.Equal(t, "carol", userID)
			assert.Equal(t, "PASSWORD", credentialType)
		},
	)

	t.Run(
		"reads webauthn from user identity additional data",
		func(t *testing.T) {
			t.Parallel()

			userID, credentialType, ok := parseIdentityCenterLoginEvent(
				`{"userIdentity":{"onBehalfOf":{"userId":"bob"},"additionalEventData":{"CredentialType":"WEBAUTHN"}}}`,
			)
			require.True(t, ok)
			assert.Equal(t, "bob", userID)
			assert.Equal(t, "WEBAUTHN", credentialType)
		},
	)

	t.Run(
		"rejects missing user id",
		func(t *testing.T) {
			t.Parallel()

			_, _, ok := parseIdentityCenterLoginEvent(`{"additionalEventData":{"CredentialType":"TOTP"}}`)
			assert.False(t, ok)
		},
	)

	t.Run(
		"rejects invalid json",
		func(t *testing.T) {
			t.Parallel()

			_, _, ok := parseIdentityCenterLoginEvent(`{`)
			assert.False(t, ok)
		},
	)
}

func TestIdentityCenterCredentialUsedMFA(t *testing.T) {
	t.Parallel()

	assert.True(t, identityCenterCredentialUsedMFA("PASSWORD,TOTP"))
	assert.True(t, identityCenterCredentialUsedMFA("webauthn"))
	assert.False(t, identityCenterCredentialUsedMFA("PASSWORD"))
	assert.False(t, identityCenterCredentialUsedMFA("EXTERNAL_IDP"))
	assert.False(t, identityCenterCredentialUsedMFA("EMAIL_OTP"))
	assert.False(t, identityCenterCredentialUsedMFA(""))
}

func TestApplyIdentityCenterActivity(t *testing.T) {
	t.Parallel()

	lastLogin := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	users := []identityCenterUser{
		{ID: "bob"},
		{ID: "carol"},
		{ID: "dave"},
	}

	applyIdentityCenterActivity(
		users,
		map[string]identityCenterLogin{
			"bob":   {at: lastLogin, credentialType: "PASSWORD,TOTP"},
			"carol": {at: lastLogin, credentialType: "PASSWORD"},
		},
		map[string]bool{"dave": true},
	)

	require.NotNil(t, users[0].LastLogin)
	assert.True(t, users[0].LastLogin.Equal(lastLogin))
	require.NotNil(t, users[0].MFAEnabled)
	assert.True(t, *users[0].MFAEnabled)
	assert.Equal(t, "PASSWORD,TOTP", users[0].CredentialType)

	require.NotNil(t, users[1].LastLogin)
	assert.True(t, users[1].LastLogin.Equal(lastLogin))
	assert.Nil(t, users[1].MFAEnabled)
	assert.Equal(t, "PASSWORD", users[1].CredentialType)

	assert.Nil(t, users[2].LastLogin)
	require.NotNil(t, users[2].MFAEnabled)
	assert.True(t, *users[2].MFAEnabled)
	assert.Empty(t, users[2].CredentialType)
}

func TestApplyIdentityCenterActivity_DeviceOverridesPasswordLogin(t *testing.T) {
	t.Parallel()

	lastLogin := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	users := []identityCenterUser{{ID: "bob"}}

	applyIdentityCenterActivity(
		users,
		map[string]identityCenterLogin{"bob": {at: lastLogin, credentialType: "PASSWORD"}},
		map[string]bool{"bob": true},
	)

	require.NotNil(t, users[0].MFAEnabled)
	assert.True(t, *users[0].MFAEnabled)
	require.NotNil(t, users[0].LastLogin)
}

func TestListMfaDevicesOutput_HasDevice(t *testing.T) {
	t.Parallel()

	assert.False(t, listMfaDevicesOutput{}.hasDevice())
	assert.True(t, listMfaDevicesOutput{MfaDevices: []json.RawMessage{[]byte(`{}`)}}.hasDevice())
	assert.True(t, listMfaDevicesOutput{MFADevices: []json.RawMessage{[]byte(`{}`)}}.hasDevice())
	assert.True(t, listMfaDevicesOutput{Devices: []json.RawMessage{[]byte(`{}`)}}.hasDevice())
}

func TestParseAWSJSONError(t *testing.T) {
	t.Parallel()

	err := parseAWSJSONError(
		[]byte(`{"__type":"AccessDeniedException","Message":"not authorized"}`),
	)
	require.True(t, identityCenterAccessDenied(err))
	assert.ErrorContains(t, err, "not authorized")
}
