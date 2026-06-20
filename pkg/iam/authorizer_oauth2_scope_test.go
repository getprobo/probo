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
		assert.Equal(t, action, scopeErr.Action)
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
