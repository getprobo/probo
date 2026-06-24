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

package iam

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/oauth2"
	"go.probo.inc/probo/pkg/iam/oauth2scope"
)

func TestAuthorizer_checkOAuth2Scope(t *testing.T) {
	t.Parallel()

	const scopeV1OrgRead = coredata.OAuth2Scope("v1:org:read")

	principal := gid.New(gid.NilTenant, coredata.IdentityEntityType)
	action := Action("core:organization:get")

	t.Run("skips when access token is absent", func(t *testing.T) {
		t.Parallel()

		a := &Authorizer{}

		err := a.checkOAuth2Scope(context.Background(), principal, action)
		require.NoError(t, err)
	})

	t.Run("denies before IAM when scopes are not registered", func(t *testing.T) {
		t.Parallel()

		a := &Authorizer{}

		ctx := oauth2.ContextWithAccessToken(
			context.Background(),
			&coredata.OAuth2AccessToken{Scopes: coredata.OAuth2Scopes{scopeV1OrgRead}},
		)

		err := a.checkOAuth2Scope(ctx, principal, action)
		require.Error(t, err)

		scopeErr, ok := errors.AsType[*ErrInsufficientOAuth2Scope](err)
		require.True(t, ok)
		assert.Equal(t, principal, scopeErr.IdentityID)
		assert.Empty(t, scopeErr.Scopes)
	})

	t.Run("reports granting scopes when token lacks authorization", func(t *testing.T) {
		t.Parallel()

		const (
			scopeV1OrgWrite = coredata.OAuth2Scope("v1:org")
			updateAction    = Action("core:organization:update")
		)

		scopeSet := oauth2scope.NewRegistry().Register(
			map[coredata.OAuth2Scope][]string{
				scopeV1OrgRead:  {action},
				scopeV1OrgWrite: {updateAction},
			},
		)

		a := NewAuthorizer(nil, nil, scopeSet)

		ctx := oauth2.ContextWithAccessToken(
			context.Background(),
			&coredata.OAuth2AccessToken{Scopes: coredata.OAuth2Scopes{scopeV1OrgRead}},
		)

		err := a.checkOAuth2Scope(ctx, principal, updateAction)
		require.Error(t, err)

		scopeErr, ok := errors.AsType[*ErrInsufficientOAuth2Scope](err)
		require.True(t, ok)
		assert.Equal(t, principal, scopeErr.IdentityID)
		assert.Equal(t, []coredata.OAuth2Scope{scopeV1OrgWrite}, scopeErr.Scopes)
	})

	t.Run("allows when registered scopes authorize the action", func(t *testing.T) {
		t.Parallel()

		scopeSet := oauth2scope.NewRegistry().Register(
			map[coredata.OAuth2Scope][]string{
				scopeV1OrgRead: {action},
			},
		)

		a := NewAuthorizer(nil, nil, scopeSet)

		ctx := oauth2.ContextWithAccessToken(
			context.Background(),
			&coredata.OAuth2AccessToken{Scopes: coredata.OAuth2Scopes{scopeV1OrgRead}},
		)

		err := a.checkOAuth2Scope(ctx, principal, action)
		require.NoError(t, err)
	})
}
