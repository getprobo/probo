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

package connect_v1

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/saml"
	"go.probo.inc/probo/pkg/mail"
)

func TestRedirectAuthError(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/api/connect/v1/oidc/google/callback", nil)
	rec := httptest.NewRecorder()

	redirectAuthError(rec, req, authErrorPersonalAccountNotAllowed)

	assert.Equal(t, http.StatusFound, rec.Code)
	location, err := rec.Result().Location()
	require.NoError(t, err)
	assert.Equal(t, "/auth/error", location.Path)
	assert.Equal(t, authErrorPersonalAccountNotAllowed, location.Query().Get("error"))
}

func TestAuthErrorCodeFromSAML(t *testing.T) {
	t.Parallel()

	configID := gid.New(gid.NewTenantID(), coredata.SAMLConfigurationEntityType)
	email, err := mail.ParseAddr("user@example.com")
	require.NoError(t, err)

	tests := []struct {
		name string
		err  error
		code string
		ok   bool
	}{
		{
			name: "disabled stays generic",
			err:  saml.NewSAMLDisabledError(),
			code: authErrorAuthenticationFailed,
			ok:   true,
		},
		{
			name: "configuration not found stays generic",
			err:  saml.NewSAMLConfigurationNotFoundError(configID),
			code: authErrorAuthenticationFailed,
			ok:   true,
		},
		{
			name: "email domain mismatch stays generic",
			err:  saml.NewEmailDomainMismatchError(email, "acme.com"),
			code: authErrorAuthenticationFailed,
			ok:   true,
		},
		{
			name: "auto signup disabled stays generic",
			err:  saml.NewSAMLAutoSignupDisabledError(configID),
			code: authErrorAuthenticationFailed,
			ok:   true,
		},
		{
			name: "user inactive stays generic",
			err:  saml.NewUserInactiveError(configID),
			code: authErrorAuthenticationFailed,
			ok:   true,
		},
		{
			name: "subject already in use stays generic",
			err:  saml.NewSAMLSubjectAlreadyInUseError("assertion-1"),
			code: authErrorAuthenticationFailed,
			ok:   true,
		},
		{
			name: "invalid assertion stays generic",
			err:  saml.NewInvalidAssertionError("assertion-1", errors.New("bad signature")),
			code: authErrorAuthenticationFailed,
			ok:   true,
		},
		{
			name: "replay stays generic",
			err:  saml.NewReplayAttackDetectedError("assertion-1"),
			code: authErrorAuthenticationFailed,
			ok:   true,
		},
		{
			name: "unknown error",
			err:  errors.New("boom"),
			ok:   false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				code, ok := authErrorCodeFromSAML(tt.err)
				assert.Equal(t, tt.ok, ok)
				assert.Equal(t, tt.code, code)
			},
		)
	}
}
