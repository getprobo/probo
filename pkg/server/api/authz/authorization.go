// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package authz

import (
	"context"
	"errors"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/gqlutils"
)

type (
	AuthorizeFuncOption func(*iam.AuthorizeParams)
	AuthorizeFunc       func(context.Context, gid.GID, string, ...AuthorizeFuncOption) error
)

func WithAttr(key, value string) AuthorizeFuncOption {
	return func(params *iam.AuthorizeParams) {
		params.ResourceAttributes[key] = value
	}
}

// Use this option when it makes no sense to check whether the viewer is assuming the org of the accessed resource
// Example: on the viewer memberships page, we're accessing several organization names, but the viewer isn't assuming one yet.
func WithSkipAssumptionCheck() AuthorizeFuncOption {
	return func(params *iam.AuthorizeParams) {
		params.Session = nil
	}
}

func NewAuthorizeFunc(
	svc *iam.Service,
	logger *log.Logger,
) AuthorizeFunc {
	return func(
		ctx context.Context,
		objectID gid.GID,
		action string,
		options ...AuthorizeFuncOption,
	) error {
		identity := authn.IdentityFromContext(ctx)
		session := authn.SessionFromContext(ctx)

		params := iam.AuthorizeParams{
			Principal:          identity.ID,
			Resource:           objectID,
			Action:             action,
			ResourceAttributes: make(map[string]string),
		}
		if session != nil {
			params.Session = &session.ID
		}

		for _, option := range options {
			option(&params)
		}

		if err := svc.Authorizer.Authorize(ctx, params); err != nil {
			var errAssumptionRequired *iam.ErrAssumptionRequired
			if errors.As(err, &errAssumptionRequired) {
				return gqlutils.AssumptionRequired(ctx, err)
			}

			var errInsufficientPermissions *iam.ErrInsufficientPermissions
			if errors.As(err, &errInsufficientPermissions) {
				return gqlutils.Forbidden(ctx, err)
			}

			logger.ErrorCtx(ctx, "cannot authorize", log.Error(err))
			return gqlutils.Internal(ctx)
		}

		return nil
	}
}
