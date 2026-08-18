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

package provider_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
)

// TestAsanaRegistrationScopes pins the registration that the outage turned on.
// The driver reads /workspaces/{gid}/workspace_memberships, which no granular
// Asana scope reaches: a token scoped users:read + workspaces:read is answered
// with 403 "Full permissions are required to use this endpoint", even though
// the probe, the organization picker and the name resolver all still succeed —
// so the connector connects green and only the campaign sync fails. Narrowing
// these scopes again reintroduces exactly that outage.
func TestAsanaRegistrationScopes(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderAsana)
	require.True(t, ok, "asana provider must be registered")

	assert.Equal(t, []string{"default"}, reg.OAuth2Scopes)
	assert.True(t, reg.ExclusiveScopes, "asana rejects any scope its app no longer offers")
}

// TestApplyOAuth2Defaults_AsanaExclusiveScopes checks the trait survives the
// registration -> connector copy, since only the connector sees it at
// authorize time.
func TestApplyOAuth2Defaults_AsanaExclusiveScopes(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()

	c := &connector.OAuth2Connector{}
	require.NoError(t, r.ApplyOAuth2Defaults(
		string(coredata.ConnectorProviderAsana),
		"https://example.com/cb",
		c,
	))

	assert.True(t, c.ExclusiveScopes)
	assert.Equal(t, []string{"default"}, c.RegisteredScopes)
}
