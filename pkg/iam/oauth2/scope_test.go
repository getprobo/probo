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
