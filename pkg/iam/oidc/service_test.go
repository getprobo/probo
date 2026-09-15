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

package oidc

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
)

func newTestService(t *testing.T) *Service {
	t.Helper()

	return NewService(
		nil,
		"https://app.probo.test",
		ProviderConfig{ClientID: "google-client", ClientSecret: "s", Enabled: true},
		ProviderConfig{ClientID: "microsoft-client", ClientSecret: "s", Enabled: true},
		log.NewLogger(),
	)
}

// TestMicrosoftRequiresDomainOwnerVerified pins the nOAuth mitigation.
// Microsoft never emits the standard email_verified claim, so trustProviderEmail
// must be true (the email_verified check is skipped); email verification is
// instead enforced through the xms_edov domain-ownership claim, which
// requireEmailDomainOwnerVerified pins.
func TestMicrosoftRequiresDomainOwnerVerified(t *testing.T) {
	t.Parallel()

	s := newTestService(t)

	microsoft := s.providers[coredata.OIDCProviderMicrosoft]
	require.NotNil(t, microsoft)
	assert.True(t, microsoft.trustProviderEmail, "Microsoft does not emit email_verified; rely on xms_edov")
	assert.True(t, microsoft.requireEmailDomainOwnerVerified, "Microsoft must require xms_edov")

	google := s.providers[coredata.OIDCProviderGoogle]
	require.NotNil(t, google)
	assert.False(t, google.requireEmailDomainOwnerVerified, "Google verifies its domains and does not use xms_edov")
}

func TestIsEmailDomainOwnerVerified(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{"bool true", true, true},
		{"bool false", false, false},
		{"string true", "true", true},
		{"string True", "True", true},
		{"string false", "false", false},
		{"absent", nil, false},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				claims := &idTokenClaims{EmailDomainOwnerVerified: tt.value}
				assert.Equal(t, tt.want, claims.isEmailDomainOwnerVerified())
			},
		)
	}
}

func TestValidateIDTokenClaims_PersonalAccounts(t *testing.T) {
	t.Parallel()

	s := newTestService(t)

	t.Run("rejects Google personal account without hosted domain", func(t *testing.T) {
		t.Parallel()

		err := validateIDTokenClaims(
			s.providers[coredata.OIDCProviderGoogle],
			&idTokenClaims{
				Email:         "user@gmail.com",
				EmailVerified: true,
			},
			false,
		)
		_, ok := errors.AsType[*ErrPersonalAccountNotAllowed](err)
		assert.True(t, ok, "got %T: %v", err, err)
	})

	t.Run("rejects Microsoft personal account before xms_edov check", func(t *testing.T) {
		t.Parallel()

		err := validateIDTokenClaims(
			s.providers[coredata.OIDCProviderMicrosoft],
			&idTokenClaims{
				Issuer: "https://login.microsoftonline.com/" + microsoftConsumerTenantID + "/v2.0",
				Email:  "user@outlook.com",
			},
			false,
		)
		_, ok := errors.AsType[*ErrPersonalAccountNotAllowed](err)
		assert.True(t, ok, "got %T: %v", err, err)
	})

	t.Run("accepts Google Workspace account with hosted domain", func(t *testing.T) {
		t.Parallel()

		err := validateIDTokenClaims(
			s.providers[coredata.OIDCProviderGoogle],
			&idTokenClaims{
				Email:         "user@acme.com",
				EmailVerified: true,
				HostedDomain:  "acme.com",
			},
			false,
		)
		assert.NoError(t, err)
	})

	t.Run("rejects Microsoft enterprise account missing xms_edov", func(t *testing.T) {
		t.Parallel()

		err := validateIDTokenClaims(
			s.providers[coredata.OIDCProviderMicrosoft],
			&idTokenClaims{
				Issuer: "https://login.microsoftonline.com/tenant-id/v2.0",
				Email:  "user@acme.com",
			},
			false,
		)
		_, ok := errors.AsType[*ErrEmailNotVerified](err)
		assert.True(t, ok, "got %T: %v", err, err)
	})
}

func TestPortalAuthorizeStateID(t *testing.T) {
	t.Parallel()

	cimdClientID := "https://trust.example.com/.well-known/oauth-client-metadata"
	stateID := "portal-oauth-state"

	tests := []struct {
		name        string
		continueURL string
		wantState   string
		wantOK      bool
	}{
		{
			name: "authorize plus cimd plus source plus state",
			continueURL: "/api/connect/v1/oauth2/authorize?client_id=" +
				url.QueryEscape(cimdClientID) +
				"&source=compliance-portal&state=" +
				url.QueryEscape(stateID),
			wantState: stateID,
			wantOK:    true,
		},
		{
			name: "absolute authorize plus cimd plus source plus state",
			continueURL: "https://auth.example.com/api/connect/v1/oauth2/authorize?client_id=" +
				url.QueryEscape(cimdClientID) +
				"&source=compliance-portal&state=" +
				url.QueryEscape(stateID),
			wantState: stateID,
			wantOK:    true,
		},
		{
			name:        "authorize plus gid client",
			continueURL: "/api/connect/v1/oauth2/authorize?client_id=gid://probo/oauth2_client/abc&source=compliance-portal&state=" + url.QueryEscape(stateID),
			wantOK:      false,
		},
		{
			name:        "overview path",
			continueURL: "/overview",
			wantOK:      false,
		},
		{
			name:        "path suffix lookalike",
			continueURL: "/evil/oauth2/authorize?client_id=" + url.QueryEscape(cimdClientID) + "&source=compliance-portal&state=" + url.QueryEscape(stateID),
			wantOK:      false,
		},
		{
			name:        "authorize plus cimd without source",
			continueURL: "/api/connect/v1/oauth2/authorize?client_id=" + url.QueryEscape(cimdClientID) + "&state=" + url.QueryEscape(stateID),
			wantOK:      false,
		},
		{
			name:        "authorize plus cimd plus source without state",
			continueURL: "/api/connect/v1/oauth2/authorize?client_id=" + url.QueryEscape(cimdClientID) + "&source=compliance-portal",
			wantOK:      false,
		},
		{
			name:        "login page source alone",
			continueURL: "/auth/login?source=compliance-portal",
			wantOK:      false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				gotState, ok := portalAuthorizeStateID(tt.continueURL)
				assert.Equal(t, tt.wantOK, ok)
				assert.Equal(t, tt.wantState, gotState)
			},
		)
	}
}

func TestValidateIDTokenClaims_AllowPersonal(t *testing.T) {
	t.Parallel()

	s := newTestService(t)

	t.Run("accepts Google personal account with verified email", func(t *testing.T) {
		t.Parallel()

		err := validateIDTokenClaims(
			s.providers[coredata.OIDCProviderGoogle],
			&idTokenClaims{
				Email:         "user@gmail.com",
				EmailVerified: true,
			},
			true,
		)
		assert.NoError(t, err)
	})

	t.Run("rejects Google personal account without verified email", func(t *testing.T) {
		t.Parallel()

		err := validateIDTokenClaims(
			s.providers[coredata.OIDCProviderGoogle],
			&idTokenClaims{
				Email:         "user@gmail.com",
				EmailVerified: false,
			},
			true,
		)
		_, ok := errors.AsType[*ErrEmailNotVerified](err)
		assert.True(t, ok, "got %T: %v", err, err)
	})

	t.Run("accepts Microsoft personal account without xms_edov", func(t *testing.T) {
		t.Parallel()

		err := validateIDTokenClaims(
			s.providers[coredata.OIDCProviderMicrosoft],
			&idTokenClaims{
				Issuer: "https://login.microsoftonline.com/" + microsoftConsumerTenantID + "/v2.0",
				Email:  "user@outlook.com",
			},
			true,
		)
		assert.NoError(t, err)
	})

	t.Run("still requires xms_edov for Microsoft work account", func(t *testing.T) {
		t.Parallel()

		err := validateIDTokenClaims(
			s.providers[coredata.OIDCProviderMicrosoft],
			&idTokenClaims{
				Issuer: "https://login.microsoftonline.com/tenant-id/v2.0",
				Email:  "user@acme.com",
			},
			true,
		)
		_, ok := errors.AsType[*ErrEmailNotVerified](err)
		assert.True(t, ok, "got %T: %v", err, err)
	})
}

func TestParseJWK_EC(t *testing.T) {
	t.Parallel()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	point, err := privateKey.PublicKey.Bytes()
	require.NoError(t, err)

	coordinateLen := (len(point) - 1) / 2

	key, err := parseJWK(
		jwk{
			Kty: "EC",
			Crv: "P-256",
			X:   base64.RawURLEncoding.EncodeToString(point[1 : 1+coordinateLen]),
			Y:   base64.RawURLEncoding.EncodeToString(point[1+coordinateLen:]),
		},
	)
	require.NoError(t, err)

	publicKey, ok := key.(*ecdsa.PublicKey)
	require.True(t, ok, "got %T", key)
	assert.True(t, publicKey.Equal(&privateKey.PublicKey))
}
