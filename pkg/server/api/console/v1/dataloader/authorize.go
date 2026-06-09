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

package dataloader

import (
	"context"
	"errors"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/server/api/authz"
	"go.probo.inc/probo/pkg/server/gqlutils"
)

// NewAuthorizeFunc returns an authz.AuthorizeFunc that batches authorize
// calls through the per-request dataloader. Parallel field resolvers within
// the same request collapse into a single iam.Authorizer.AuthorizeMulti
// call. Requires the dataloader middleware to have populated Loaders in
// context.
func NewAuthorizeFunc(logger *log.Logger) authz.AuthorizeFunc {
	return func(
		ctx context.Context,
		objectID gid.GID,
		action string,
		options ...authz.AuthorizeFuncOption,
	) (*coredata.Scope, error) {
		loaders := FromContext(ctx)

		applied := iam.AuthorizeParams{
			ResourceAttributes: make(map[string]string),
		}
		for _, opt := range options {
			opt(&applied)
		}

		result, err := loaders.Authorize.Load(
			ctx,
			AuthorizeKey{
				ResourceID:          objectID,
				Action:              action,
				ResourceAttributes:  EncodeAuthorizeKeyAttributes(applied.ResourceAttributes),
				DryRun:              applied.DryRun,
				SkipAssumptionCheck: applied.SkipAssumptionCheck,
			},
		)
		if err != nil {
			if _, ok := errors.AsType[*iam.ErrAssumptionRequired](err); ok {
				return nil, gqlutils.AssumptionRequired(ctx, err)
			}

			if _, ok := errors.AsType[*iam.ErrInsufficientPermissions](err); ok {
				return nil, gqlutils.Forbidden(ctx, err)
			}

			if errors.Is(err, coredata.ErrResourceNotFound) {
				return nil, gqlutils.NotFoundf(ctx, "resource not found")
			}

			logger.ErrorCtx(ctx, "cannot authorize", log.Error(err))

			return nil, gqlutils.Internal(ctx)
		}

		return result.Scope, nil
	}
}
