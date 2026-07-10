// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package oidc

import (
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

// TestMicrosoftRequiresDomainOwnerVerified pins the nOAuth mitigation: the
// Microsoft provider must not trust the email claim on email_verified alone and
// must require the xms_edov domain-ownership claim.
func TestMicrosoftRequiresDomainOwnerVerified(t *testing.T) {
	t.Parallel()

	s := newTestService(t)

	microsoft := s.providers[coredata.OIDCProviderMicrosoft]
	require.NotNil(t, microsoft)
	assert.False(t, microsoft.trustProviderEmail, "Microsoft email must not be trusted unconditionally")
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
