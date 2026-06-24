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

package oauth2scope_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/iam/oauth2scope"
)

func TestRegistry_Allows(t *testing.T) {
	t.Parallel()

	const scopeV1OrgRead = coredata.OAuth2Scope("v1:org:read")

	reg := oauth2scope.NewRegistry().Register(
		map[coredata.OAuth2Scope][]string{
			scopeV1OrgRead: {"core:organization:get"},
		},
	)

	tokenScopes := coredata.OAuth2Scopes{scopeV1OrgRead}

	assert.True(t, reg.Allows(tokenScopes, "core:organization:get"))
	assert.False(t, reg.Allows(tokenScopes, "core:organization:update"))
}

func TestRegistry_Register(t *testing.T) {
	t.Parallel()

	const scopeV1OrgRead = coredata.OAuth2Scope("v1:org:read")

	reg := oauth2scope.NewRegistry().
		Register(
			map[coredata.OAuth2Scope][]string{
				scopeV1OrgRead: {"core:organization:get"},
			},
		).
		Register(
			map[coredata.OAuth2Scope][]string{
				scopeV1OrgRead: {"core:organization-context:get"},
			},
		)

	tokenScopes := coredata.OAuth2Scopes{scopeV1OrgRead}

	assert.True(t, reg.Allows(tokenScopes, "core:organization:get"))
	assert.True(t, reg.Allows(tokenScopes, "core:organization-context:get"))
}

func TestRegistry_ValidateScopes(t *testing.T) {
	t.Parallel()

	const scopeV1OrgRead = coredata.OAuth2Scope("v1:org:read")

	reg := oauth2scope.NewRegistry().Register(
		map[coredata.OAuth2Scope][]string{
			scopeV1OrgRead: {"core:organization:get"},
		},
	)

	require.NoError(t, reg.ValidateScopes(coredata.OAuth2Scopes{scopeV1OrgRead}))

	err := reg.ValidateScopes(coredata.OAuth2Scopes{"v1:unknown:read"})
	require.Error(t, err)
	assert.EqualError(t, err, "invalid scope: v1:unknown:read")
}

func TestRegistry_ScopesForAction(t *testing.T) {
	t.Parallel()

	const (
		scopeV1OrgRead  = coredata.OAuth2Scope("v1:org:read")
		scopeV1OrgWrite = coredata.OAuth2Scope("v1:org")
	)

	reg := oauth2scope.NewRegistry().Register(
		map[coredata.OAuth2Scope][]string{
			scopeV1OrgRead:  {"core:organization:get"},
			scopeV1OrgWrite: {"core:organization:get", "core:organization:update"},
		},
	)

	assert.Equal(
		t,
		[]coredata.OAuth2Scope{scopeV1OrgWrite, scopeV1OrgRead},
		reg.ScopesForAction("core:organization:get"),
	)
	assert.Equal(t, []coredata.OAuth2Scope{scopeV1OrgWrite}, reg.ScopesForAction("core:organization:update"))
	assert.Nil(t, reg.ScopesForAction("core:organization:delete"))
}

func TestRegistry_AllWriteScopes(t *testing.T) {
	t.Parallel()

	const (
		scopeV1OrgRead  = coredata.OAuth2Scope("v1:org:read")
		scopeV1OrgWrite = coredata.OAuth2Scope("v1:org")
	)

	reg := oauth2scope.NewRegistry().
		Register(
			map[coredata.OAuth2Scope][]string{
				scopeV1OrgWrite: {"core:organization:update"},
				scopeV1OrgRead:  {"core:organization:get"},
			},
		)

	assert.Equal(t, []coredata.OAuth2Scope{scopeV1OrgWrite}, reg.AllWriteScopes())
}

func TestRegistry_RegisteredScopes(t *testing.T) {
	t.Parallel()

	const (
		scopeV1OrgRead  = coredata.OAuth2Scope("v1:org:read")
		scopeV1OrgWrite = coredata.OAuth2Scope("v1:org")
	)

	reg := oauth2scope.NewRegistry().
		Register(
			map[coredata.OAuth2Scope][]string{
				scopeV1OrgWrite: {"core:organization:update"},
				scopeV1OrgRead:  {"core:organization:get"},
			},
		)

	assert.Equal(
		t,
		[]coredata.OAuth2Scope{scopeV1OrgWrite, scopeV1OrgRead},
		reg.RegisteredScopes(),
	)
}
