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

func TestLinearRegistrationScopes(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderLinear)
	require.True(t, ok, "linear provider must be registered")

	assert.Equal(t, []string{"read"}, reg.OAuth2.Scopes)
	assert.Equal(t, ",", reg.OAuth2.ScopeSeparator)
	assert.Equal(t, "app", reg.OAuth2.ExtraAuthParams["actor"])
	assert.NotNil(t, reg.NewDriver)
}

func TestLinearSyncRegistrationScopes(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderLinearSync)
	require.True(t, ok, "linear sync provider must be registered")

	assert.Equal(t, "Linear Sync", reg.DisplayName)
	assert.Equal(t, []string{"read", "write", "issues:create"}, reg.OAuth2.Scopes)
	assert.Equal(t, ",", reg.OAuth2.ScopeSeparator)
	assert.Equal(t, "app", reg.OAuth2.ExtraAuthParams["actor"])
	assert.Nil(t, reg.NewDriver, "linear sync must stay out of the access-review catalog")
}

func TestApplyOAuth2Defaults_LinearScopeSeparator(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()

	c := &connector.OAuth2Connector{}
	require.NoError(t, r.ApplyOAuth2Defaults(
		string(coredata.ConnectorProviderLinear),
		"https://example.com/cb",
		c,
	))

	assert.Equal(t, []string{"read"}, c.RegisteredScopes)
	assert.Equal(t, ",", c.ScopeSeparator)
}

func TestApplyOAuth2Defaults_LinearSyncScopeSeparator(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()

	c := &connector.OAuth2Connector{}
	require.NoError(t, r.ApplyOAuth2Defaults(
		string(coredata.ConnectorProviderLinearSync),
		"https://example.com/cb",
		c,
	))

	assert.Equal(t, []string{"read", "write", "issues:create"}, c.RegisteredScopes)
	assert.Equal(t, ",", c.ScopeSeparator)
}
