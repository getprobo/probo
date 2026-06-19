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
