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

package authz

import (
	"context"
	"errors"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/gqlutils"
)

type (
	AuthorizeFuncOption      func(*iam.AuthorizeParams)
	AuthorizeFunc            func(context.Context, gid.GID, iam.Action, ...AuthorizeFuncOption) (*coredata.Scope, error)
	BatchAuthorizeFuncOption func(*iam.AuthorizeBatchParams)
	BatchAuthorizeFunc       func(context.Context, iam.Action, []gid.GID, ...BatchAuthorizeFuncOption) (*coredata.Scope, error)
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
		params.SkipAssumptionCheck = true
	}
}

func WithDryRun() AuthorizeFuncOption {
	return func(params *iam.AuthorizeParams) {
		params.DryRun = true
	}
}

func WithBatchAttr(key, value string) BatchAuthorizeFuncOption {
	return func(params *iam.AuthorizeBatchParams) {
		params.ResourceAttributes[key] = value
	}
}

func WithBatchSkipAssumptionCheck() BatchAuthorizeFuncOption {
	return func(params *iam.AuthorizeBatchParams) {
		params.SkipAssumptionCheck = true
	}
}

func WithBatchDryRun() BatchAuthorizeFuncOption {
	return func(params *iam.AuthorizeBatchParams) {
		params.DryRun = true
	}
}

func NewAuthorizeFunc(
	svc *iam.Service,
	logger *log.Logger,
) AuthorizeFunc {
	return func(
		ctx context.Context,
		objectID gid.GID,
		action iam.Action,
		options ...AuthorizeFuncOption,
	) (*coredata.Scope, error) {
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

		scope, err := svc.Authorizer.Authorize(ctx, params)
		if err != nil {
			if _, ok := errors.AsType[*iam.ErrAssumptionRequired](err); ok {
				return nil, gqlutils.AssumptionRequired(ctx, err)
			}

			if _, ok := errors.AsType[*iam.ErrInsufficientPermissions](err); ok {
				return nil, gqlutils.Forbidden(ctx, err)
			}

			if _, ok := errors.AsType[*iam.ErrInsufficientOAuth2Scope](err); ok {
				return nil, gqlutils.Forbidden(ctx, err)
			}

			if errors.Is(err, coredata.ErrResourceNotFound) {
				return nil, gqlutils.NotFoundf(ctx, "resource not found")
			}

			logger.ErrorCtx(ctx, "cannot authorize", log.Error(err))

			return nil, gqlutils.Internal(ctx)
		}

		return scope, nil
	}
}

func NewBatchAuthorizeFunc(
	svc *iam.Service,
	logger *log.Logger,
) BatchAuthorizeFunc {
	return func(
		ctx context.Context,
		action iam.Action,
		objectIDs []gid.GID,
		options ...BatchAuthorizeFuncOption,
	) (*coredata.Scope, error) {
		identity := authn.IdentityFromContext(ctx)
		session := authn.SessionFromContext(ctx)

		params := iam.AuthorizeBatchParams{
			Principal:          identity.ID,
			Action:             action,
			Resources:          objectIDs,
			ResourceAttributes: make(map[string]string),
		}
		if session != nil {
			params.Session = &session.ID
		}

		for _, option := range options {
			option(&params)
		}

		scope, err := svc.Authorizer.AuthorizeBatch(ctx, params)
		if err != nil {
			if _, ok := errors.AsType[*iam.ErrAssumptionRequired](err); ok {
				return nil, gqlutils.AssumptionRequired(ctx, err)
			}

			if _, ok := errors.AsType[*iam.ErrInsufficientPermissions](err); ok {
				return nil, gqlutils.Forbidden(ctx, err)
			}

			if _, ok := errors.AsType[*iam.ErrInsufficientOAuth2Scope](err); ok {
				return nil, gqlutils.Forbidden(ctx, err)
			}

			if _, ok := errors.AsType[*iam.ErrMixedOrganizationBatch](err); ok {
				return nil, gqlutils.Invalid(ctx, err)
			}

			if _, ok := errors.AsType[*iam.ErrEmptyResourceBatch](err); ok {
				return nil, gqlutils.Invalid(ctx, err)
			}

			if _, ok := errors.AsType[*iam.ErrBatchAuthorizationUnsupportedResourceType](err); ok {
				return nil, gqlutils.Invalid(ctx, err)
			}

			if errors.Is(err, coredata.ErrResourceNotFound) {
				return nil, gqlutils.NotFoundf(ctx, "resource not found")
			}

			logger.ErrorCtx(ctx, "cannot batch authorize", log.Error(err))

			return nil, gqlutils.Internal(ctx)
		}

		return scope, nil
	}
}
