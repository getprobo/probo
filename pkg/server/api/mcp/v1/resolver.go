// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

//go:generate go run go.probo.inc/mcpgen generate

package mcp_v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/cookiebanner"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/riskmanagement"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/thirdparty"
)

type Resolver struct {
	proboSvc       *probo.Service
	thirdPartySvc  *thirdparty.Service
	iamSvc         *iam.Service
	accessReview   *accessreview.Service
	cookieBanner   *cookiebanner.Service
	riskManagement *riskmanagement.Service
	logger         *log.Logger
}

func markdownToProseMirrorJSON(markdown string) (string, error) {
	node, err := prosemirror.ParseMarkdown(markdown)
	if err != nil {
		return "", fmt.Errorf("cannot parse markdown: %w", err)
	}

	out, err := json.Marshal(node)
	if err != nil {
		return "", fmt.Errorf("cannot marshal prosemirror node: %w", err)
	}

	return string(out), nil
}

func (r *Resolver) Authorize(ctx context.Context, entityID gid.GID, action iam.Action) (*coredata.Scope, error) {
	identity := authn.IdentityFromContext(ctx)

	scope, err := r.iamSvc.Authorizer.Authorize(
		ctx,
		iam.AuthorizeParams{
			Principal: identity.ID,
			Resource:  entityID,
			Action:    action,
		},
	)
	if err == nil {
		return scope, nil
	}

	if _, ok := errors.AsType[*iam.ErrInsufficientPermissions](err); ok {
		return nil, fmt.Errorf("permission denied")
	}

	if _, ok := errors.AsType[*iam.ErrAssumptionRequired](err); ok {
		return nil, fmt.Errorf("assumption required")
	}

	if errors.Is(err, coredata.ErrResourceNotFound) {
		return nil, fmt.Errorf("resource not found")
	}

	r.logger.ErrorCtx(ctx, "cannot authorize MCP request", log.Error(err))

	return nil, fmt.Errorf("internal server error")
}

func (r *Resolver) AuthorizeBatch(ctx context.Context, entityIDs []gid.GID, action iam.Action) (*coredata.Scope, error) {
	identity := authn.IdentityFromContext(ctx)

	scope, err := r.iamSvc.Authorizer.AuthorizeBatch(
		ctx,
		iam.AuthorizeBatchParams{
			Principal: identity.ID,
			Resources: entityIDs,
			Action:    action,
		},
	)
	if err == nil {
		return scope, nil
	}

	if _, ok := errors.AsType[*iam.ErrInsufficientPermissions](err); ok {
		return nil, fmt.Errorf("permission denied")
	}

	if _, ok := errors.AsType[*iam.ErrAssumptionRequired](err); ok {
		return nil, fmt.Errorf("assumption required")
	}

	if _, ok := errors.AsType[*iam.ErrMixedOrganizationBatch](err); ok {
		return nil, fmt.Errorf("mixed-organization batch")
	}

	if _, ok := errors.AsType[*iam.ErrEmptyResourceBatch](err); ok {
		return nil, fmt.Errorf("empty resource batch")
	}

	if _, ok := errors.AsType[*iam.ErrBatchAuthorizationUnsupportedResourceType](err); ok {
		return nil, fmt.Errorf("batch authorization unsupported for resource type")
	}

	if errors.Is(err, coredata.ErrResourceNotFound) {
		return nil, fmt.Errorf("resource not found")
	}

	r.logger.ErrorCtx(ctx, "cannot batch authorize MCP request", log.Error(err))

	return nil, fmt.Errorf("internal server error")
}
