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

package iam

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

func TestScopeSet_Allows(t *testing.T) {
	t.Parallel()

	const scopeV1OrgRead = coredata.OAuth2Scope("v1:org:read")

	scopeSet := CreateScopeSet(
		map[coredata.OAuth2Scope][]Action{
			scopeV1OrgRead: {"core:organization:get"},
		},
	)

	tokenScopes := coredata.OAuth2Scopes{scopeV1OrgRead}

	assert.True(t, scopeSet.Allows(tokenScopes, "core:organization:get"))
	assert.False(t, scopeSet.Allows(tokenScopes, "core:organization:update"))
}

func TestScopeSet_Merge(t *testing.T) {
	t.Parallel()

	const scopeV1OrgRead = coredata.OAuth2Scope("v1:org:read")

	scopeSet := NewScopeSet().
		Merge(
			CreateScopeSet(
				map[coredata.OAuth2Scope][]Action{
					scopeV1OrgRead: {"core:organization:get"},
				},
			),
		).
		Merge(
			CreateScopeSet(
				map[coredata.OAuth2Scope][]Action{
					scopeV1OrgRead: {"core:organization-context:get"},
				},
			),
		)

	tokenScopes := coredata.OAuth2Scopes{scopeV1OrgRead}

	assert.True(t, scopeSet.Allows(tokenScopes, "core:organization:get"))
	assert.True(t, scopeSet.Allows(tokenScopes, "core:organization-context:get"))
}

func TestAuthorizer_RegisterScopes(t *testing.T) {
	t.Parallel()

	const scopeV1OrgRead = coredata.OAuth2Scope("v1:org:read")

	authorizer := NewAuthorizer(nil, nil)
	authorizer.RegisterScopes(
		CreateScopeSet(
			map[coredata.OAuth2Scope][]Action{
				scopeV1OrgRead: {"core:organization:get"},
			},
		),
	)

	require.NotNil(t, authorizer.oauth2ScopeSet)

	tokenScopes := coredata.OAuth2Scopes{scopeV1OrgRead}
	assert.True(t, authorizer.oauth2ScopeSet.Allows(tokenScopes, "core:organization:get"))
	assert.False(t, authorizer.oauth2ScopeSet.Allows(tokenScopes, "core:organization:update"))
}
