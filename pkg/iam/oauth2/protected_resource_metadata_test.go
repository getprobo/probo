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
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/iam/oauth2"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/uri"
)

func TestNewProtectedResourceMetadata(t *testing.T) {
	t.Parallel()

	apiScopes := []coredata.OAuth2Scope{probo.ScopeV1DocumentRead}

	resource := uri.URI("https://app.example.com")
	authorizationServer := uri.URI("https://app.example.com")

	metadata := oauth2.NewProtectedResourceMetadata(resource, authorizationServer, apiScopes)
	require.NotNil(t, metadata)

	assert.Equal(t, resource, metadata.Resource)
	assert.Equal(t, []uri.URI{authorizationServer}, metadata.AuthorizationServers)
	assert.Equal(t, []string{"header"}, metadata.BearerMethodsSupported)
	assert.Contains(t, metadata.ScopesSupported, oauth2.ScopeOpenID)
	assert.Contains(t, metadata.ScopesSupported, probo.ScopeV1DocumentRead)
	assert.NotContains(t, metadata.ScopesSupported, oauth2.ScopeProfile)
}
