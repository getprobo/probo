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

package coredata_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/coredata"
)

func TestOAuth2Scope_IsValid(t *testing.T) {
	t.Parallel()

	t.Run(
		"offline_access is valid",
		func(t *testing.T) {
			t.Parallel()

			assert.True(t, coredata.OAuth2ScopeOfflineAccess.IsValid())
		},
	)

	t.Run(
		"unknown scope is invalid",
		func(t *testing.T) {
			t.Parallel()

			assert.False(t, coredata.OAuth2Scope("admin").IsValid())
		},
	)
}

func TestOAuth2Scope_UnmarshalText(t *testing.T) {
	t.Parallel()

	t.Run(
		"offline_access unmarshals",
		func(t *testing.T) {
			t.Parallel()

			var scope coredata.OAuth2Scope
			err := scope.UnmarshalText([]byte("offline_access"))
			assert.NoError(t, err)
			assert.Equal(t, coredata.OAuth2ScopeOfflineAccess, scope)
		},
	)

	t.Run(
		"invalid scope returns error",
		func(t *testing.T) {
			t.Parallel()

			var scope coredata.OAuth2Scope
			err := scope.UnmarshalText([]byte("admin"))
			assert.Error(t, err)
		},
	)
}

func TestOAuth2Scopes_Contains(t *testing.T) {
	t.Parallel()

	t.Run(
		"contains offline_access",
		func(t *testing.T) {
			t.Parallel()

			scopes := coredata.OAuth2Scopes{
				coredata.OAuth2ScopeOpenID,
				coredata.OAuth2ScopeOfflineAccess,
			}
			assert.True(t, scopes.Contains(coredata.OAuth2ScopeOfflineAccess))
		},
	)

	t.Run(
		"does not contain offline_access",
		func(t *testing.T) {
			t.Parallel()

			scopes := coredata.OAuth2Scopes{
				coredata.OAuth2ScopeOpenID,
				coredata.OAuth2ScopeProfile,
			}
			assert.False(t, scopes.Contains(coredata.OAuth2ScopeOfflineAccess))
		},
	)
}

func TestOAuth2Scopes_OrDefault(t *testing.T) {
	t.Parallel()

	defaultScopes := coredata.OAuth2Scopes{
		coredata.OAuth2ScopeOpenID,
		coredata.OAuth2ScopeProfile,
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

			scopes := coredata.OAuth2Scopes{coredata.OAuth2ScopeEmail}
			result := scopes.OrDefault(defaultScopes)
			assert.Equal(t, scopes, result)
		},
	)
}
