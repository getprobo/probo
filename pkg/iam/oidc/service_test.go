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
	"errors"
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
		)
		_, ok := errors.AsType[*ErrEmailNotVerified](err)
		assert.True(t, ok, "got %T: %v", err, err)
	})
}
