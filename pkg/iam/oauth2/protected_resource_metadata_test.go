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
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/iam/oauth2"
	"go.probo.inc/probo/pkg/iam/oauth2scope"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/uri"
)

func TestNewProtectedResourceMetadata(t *testing.T) {
	t.Parallel()

	reg := oauth2scope.NewRegistry().Register(
		map[coredata.OAuth2Scope][]string{
			probo.ScopeV1DocumentRead: {"core:document:get"},
			probo.ScopeV1Document:     {"core:document:create"},
		},
	)

	resource := uri.URI("https://app.example.com")
	authorizationServer := uri.URI("https://app.example.com")

	metadata := oauth2.NewProtectedResourceMetadata(resource, authorizationServer, reg.AllWriteScopes())
	require.NotNil(t, metadata)

	assert.Equal(t, resource, metadata.Resource)
	assert.Equal(t, []uri.URI{authorizationServer}, metadata.AuthorizationServers)
	assert.Equal(t, []string{"header"}, metadata.BearerMethodsSupported)
	assert.Contains(t, metadata.ScopesSupported, oauth2.ScopeOpenID)
	assert.Contains(t, metadata.ScopesSupported, probo.ScopeV1Document)
	assert.NotContains(t, metadata.ScopesSupported, probo.ScopeV1DocumentRead)
	assert.NotContains(t, metadata.ScopesSupported, oauth2.ScopeProfile)
}
