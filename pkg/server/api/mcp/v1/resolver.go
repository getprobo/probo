// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

//go:generate go run go.probo.inc/mcpgen generate

package mcp_v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/cookiebanner"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/resourcealias"
	"go.probo.inc/probo/pkg/riskmanagement"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/thirdparty"
)

type Resolver struct {
	proboSvc       *probo.Service
	resourceAlias  *resourcealias.Service
	thirdPartySvc  *thirdparty.Service
	iamSvc         *iam.Service
	accessReview   *accessreview.Service
	cookieBanner   *cookiebanner.Service
	riskManagement *riskmanagement.Service
	logger         *log.Logger
	fileManager    *filemanager.Service
	baseURL        *baseurl.BaseURL
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

	if _, ok := errors.AsType[*iam.ErrInsufficientOAuth2Scope](err); ok {
		return nil, fmt.Errorf("insufficient scope")
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

	if _, ok := errors.AsType[*iam.ErrInsufficientOAuth2Scope](err); ok {
		return nil, fmt.Errorf("insufficient scope")
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
