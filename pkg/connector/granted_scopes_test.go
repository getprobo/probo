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

package connector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGrantedScopes(t *testing.T) {
	t.Parallel()

	t.Run("nil connection grants nothing", func(t *testing.T) {
		t.Parallel()

		assert.Nil(t, GrantedScopes(nil))
	})

	t.Run("api key connection grants nothing", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, GrantedScopes(&APIKeyConnection{APIKey: "k"}))
	})

	t.Run("oauth2 connection grants its scope string", func(t *testing.T) {
		t.Parallel()

		assert.Equal(
			t,
			[]string{"User.Read.All", "openid"},
			GrantedScopes(&OAuth2Connection{Scope: "openid User.Read.All"}),
		)
	})

	// Microsoft's v2 token response returns Graph scopes (and often
	// openid/profile) but never offline_access, despite issuing a refresh
	// token. Existing connections may also lack it in Scope.
	t.Run("refresh token proves offline_access", func(t *testing.T) {
		t.Parallel()

		assert.Equal(
			t,
			[]string{"offline_access", "openid"},
			GrantedScopes(&OAuth2Connection{RefreshToken: "rt", Scope: "openid"}),
		)
	})

	t.Run("no refresh token leaves offline_access ungranted", func(t *testing.T) {
		t.Parallel()

		assert.Equal(
			t,
			[]string{"openid"},
			GrantedScopes(&OAuth2Connection{Scope: "openid"}),
		)
	})

	t.Run("slack connection carries its embedded refresh token", func(t *testing.T) {
		t.Parallel()

		conn := &SlackConnection{
			OAuth2Connection: OAuth2Connection{RefreshToken: "rt", Scope: "channels:read"},
		}

		assert.Equal(t, []string{"channels:read", "offline_access"}, GrantedScopes(conn))
	})
}

// The console badge reads a missing scope as RECONNECT_REQUIRED, so the two
// helpers must compose to an empty result whenever the grant already covers
// what the provider registration asks for.
func TestMissingScopesAgainstGrant(t *testing.T) {
	t.Parallel()

	required := []string{
		"openid",
		"https://graph.microsoft.com/AuditLog.Read.All",
		"https://graph.microsoft.com/User.Read.All",
	}

	t.Run("partial grant reports what is missing", func(t *testing.T) {
		t.Parallel()

		granted := GrantedScopes(&OAuth2Connection{Scope: "openid User.Read.All"})

		assert.Equal(
			t,
			[]string{"https://graph.microsoft.com/AuditLog.Read.All"},
			MissingScopes(required, granted),
		)
	})

	t.Run("full grant reports nothing missing", func(t *testing.T) {
		t.Parallel()

		granted := GrantedScopes(&OAuth2Connection{Scope: "openid AuditLog.Read.All User.Read.All"})

		assert.Empty(t, MissingScopes(required, granted))
	})

	t.Run("nil connection is short of everything", func(t *testing.T) {
		t.Parallel()

		assert.Equal(
			t,
			[]string{
				"https://graph.microsoft.com/AuditLog.Read.All",
				"https://graph.microsoft.com/User.Read.All",
				"openid",
			},
			MissingScopes(required, GrantedScopes(nil)),
		)
	})

	t.Run("refresh token satisfies a required offline_access", func(t *testing.T) {
		t.Parallel()

		granted := GrantedScopes(&OAuth2Connection{
			RefreshToken: "rt",
			Scope:        "openid User.Read.All",
		})

		assert.Empty(
			t,
			MissingScopes(
				[]string{"openid", "offline_access", "https://graph.microsoft.com/User.Read.All"},
				granted,
			),
		)
	})
}
