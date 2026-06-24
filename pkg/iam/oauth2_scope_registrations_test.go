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

package iam_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/agentrun"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/iam/oauth2scope"
	"go.probo.inc/probo/pkg/probo"
)

func allRegisteredOAuth2ScopeRegistries() *oauth2scope.Registry {
	return oauth2scope.NewRegistry().
		Register(iam.IAMOAuth2ScopeMappings).
		Register(probo.OAuth2ScopeMappings).
		Register(accessreview.OAuth2ScopeMappings).
		Register(agentrun.OAuth2ScopeMappings)
}

func TestRegisteredOAuth2ScopeRegistries_OrganizationRead(t *testing.T) {
	t.Parallel()

	reg := allRegisteredOAuth2ScopeRegistries()
	tokenScopes := coredata.OAuth2Scopes{probo.ScopeV1OrgRead}

	assert.True(t, reg.Allows(tokenScopes, probo.ActionOrganizationGet))
	assert.False(t, reg.Allows(tokenScopes, probo.ActionOrganizationUpdate))
	assert.False(t, reg.Allows(tokenScopes, probo.ActionThirdPartyList))
}

func TestRegisteredOAuth2ScopeRegistries_UnmappedActionDenies(t *testing.T) {
	t.Parallel()

	reg := allRegisteredOAuth2ScopeRegistries()
	tokenScopes := coredata.OAuth2Scopes{
		probo.ScopeV1OrgRead,
		probo.ScopeV1ThirdPartyRead,
	}

	assert.False(t, reg.Allows(tokenScopes, "core:unmapped:action"))
}
