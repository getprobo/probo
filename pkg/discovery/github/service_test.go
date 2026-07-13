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

package github

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
)

func TestValidateDiscoveryScopes(t *testing.T) {
	t.Parallel()

	t.Run("oauth connector with required scopes passes", func(t *testing.T) {
		t.Parallel()

		err := validateDiscoveryScopes(
			coredata.ConnectorProtocolOAuth2,
			&connector.OAuth2Connection{Scope: "read:org repo"},
		)
		assert.NoError(t, err)
	})

	t.Run("oauth connector missing required scopes fails", func(t *testing.T) {
		t.Parallel()

		err := validateDiscoveryScopes(
			coredata.ConnectorProtocolOAuth2,
			&connector.OAuth2Connection{Scope: "repo"},
		)
		require.Error(t, err)

		var scopeErr *InsufficientScopesError
		require.True(t, errors.As(err, &scopeErr))
		assert.Equal(t, []string{"read:org"}, scopeErr.Missing)
	})

	t.Run("api key connector skips scope validation", func(t *testing.T) {
		t.Parallel()

		err := validateDiscoveryScopes(
			coredata.ConnectorProtocolAPIKey,
			&connector.APIKeyConnection{APIKey: "ghp_test"},
		)
		assert.NoError(t, err)
	})
}
