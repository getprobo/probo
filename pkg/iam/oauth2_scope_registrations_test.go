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

package iam_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/agentrun"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
)

func allRegisteredOAuth2ScopeSets() *iam.ScopeSet {
	return iam.NewScopeSet().
		Merge(iam.IAMOAuth2ScopeSet()).
		Merge(probo.OAuth2ScopeSet()).
		Merge(accessreview.OAuth2ScopeSet()).
		Merge(agentrun.OAuth2ScopeSet())
}

func registerAllOAuth2ScopeSets(authorizer *iam.Authorizer) {
	authorizer.RegisterScopes(iam.IAMOAuth2ScopeSet())
	authorizer.RegisterScopes(probo.OAuth2ScopeSet())
	authorizer.RegisterScopes(accessreview.OAuth2ScopeSet())
	authorizer.RegisterScopes(agentrun.OAuth2ScopeSet())
}

func TestRegisteredOAuth2ScopeSets_OrganizationRead(t *testing.T) {
	t.Parallel()

	authorizer := iam.NewAuthorizer(nil, nil)
	registerAllOAuth2ScopeSets(authorizer)

	scopeSet := allRegisteredOAuth2ScopeSets()
	tokenScopes := coredata.OAuth2Scopes{probo.ScopeV1OrgRead}

	assert.True(t, scopeSet.Allows(tokenScopes, probo.ActionOrganizationGet))
	assert.False(t, scopeSet.Allows(tokenScopes, probo.ActionOrganizationUpdate))
	assert.False(t, scopeSet.Allows(tokenScopes, probo.ActionThirdPartyList))
}

func TestRegisteredOAuth2ScopeSets_UnmappedActionDenies(t *testing.T) {
	t.Parallel()

	authorizer := iam.NewAuthorizer(nil, nil)
	registerAllOAuth2ScopeSets(authorizer)

	scopeSet := allRegisteredOAuth2ScopeSets()
	tokenScopes := coredata.OAuth2Scopes{
		probo.ScopeV1OrgRead,
		probo.ScopeV1ThirdPartyRead,
	}

	assert.False(t, scopeSet.Allows(tokenScopes, "core:unmapped:action"))
}
