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
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/iam/oauth2"
)

func TestIsValid(t *testing.T) {
	t.Parallel()

	t.Run(
		"offline_access is valid",
		func(t *testing.T) {
			t.Parallel()

			assert.True(t, oauth2.IsValid(oauth2.ScopeOfflineAccess))
		},
	)

	t.Run(
		"unknown scope is invalid",
		func(t *testing.T) {
			t.Parallel()

			assert.False(t, oauth2.IsValid(coredata.OAuth2Scope("admin")))
		},
	)
}

func TestUnmarshalScope(t *testing.T) {
	t.Parallel()

	t.Run(
		"offline_access unmarshals",
		func(t *testing.T) {
			t.Parallel()

			scope, err := oauth2.UnmarshalScope([]byte("offline_access"))
			assert.NoError(t, err)
			assert.Equal(t, oauth2.ScopeOfflineAccess, scope)
		},
	)

	t.Run(
		"invalid scope returns error",
		func(t *testing.T) {
			t.Parallel()

			_, err := oauth2.UnmarshalScope([]byte("admin"))
			assert.Error(t, err)
		},
	)
}

func TestOAuth2ScopesContains(t *testing.T) {
	t.Parallel()

	t.Run(
		"contains offline_access",
		func(t *testing.T) {
			t.Parallel()

			scopes := coredata.OAuth2Scopes{
				oauth2.ScopeOpenID,
				oauth2.ScopeOfflineAccess,
			}
			assert.True(t, scopes.Contains(oauth2.ScopeOfflineAccess))
		},
	)

	t.Run(
		"does not contain offline_access",
		func(t *testing.T) {
			t.Parallel()

			scopes := coredata.OAuth2Scopes{
				oauth2.ScopeOpenID,
				oauth2.ScopeProfile,
			}
			assert.False(t, scopes.Contains(oauth2.ScopeOfflineAccess))
		},
	)
}

func TestOAuth2ScopesOrDefault(t *testing.T) {
	t.Parallel()

	defaultScopes := coredata.OAuth2Scopes{
		oauth2.ScopeOpenID,
		oauth2.ScopeProfile,
	}

	t.Run(
		"returns default when scopes is nil",
		func(t *testing.T) {
			t.Parallel()

			var scopes coredata.OAuth2Scopes

			result := scopes.OrDefault(defaultScopes)
			assert.Equal(t, defaultScopes, result)
		},
	)

	t.Run(
		"returns default when scopes is empty",
		func(t *testing.T) {
			t.Parallel()

			scopes := coredata.OAuth2Scopes{}
			result := scopes.OrDefault(defaultScopes)
			assert.Equal(t, defaultScopes, result)
		},
	)

	t.Run(
		"returns scopes when non-empty",
		func(t *testing.T) {
			t.Parallel()

			scopes := coredata.OAuth2Scopes{oauth2.ScopeEmail}
			result := scopes.OrDefault(defaultScopes)
			assert.Equal(t, scopes, result)
		},
	)
}
