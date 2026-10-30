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

package mcp_v1

import (
	"context"
	"errors"

	"go.probo.inc/probo/pkg/collaboration"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/mcp/mcpx"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/types"
)

func (r *Resolver) setDocumentCommentThreadResolved(
	ctx context.Context,
	threadID gid.GID,
	resolved bool,
) (*types.DocumentCommentThread, error) {
	scope, err := r.Authorize(ctx, threadID, probo.ActionDocumentCommentResolve)
	if err != nil {
		return nil, err
	}

	identity := authn.IdentityFromContext(ctx)

	thread, err := r.proboSvc.Documents.SetCommentThreadResolved(
		ctx,
		scope,
		probo.ResolveDocumentCommentThreadRequest{
			ThreadID:   threadID,
			IdentityID: identity.ID,
			Resolved:   resolved,
		},
	)
	if err != nil {
		return nil, r.documentReviewError(ctx, "cannot update document comment thread", err)
	}

	return types.NewDocumentCommentThread(thread), nil
}

func (r *Resolver) documentReviewError(ctx context.Context, message string, err error) error {
	switch {
	case errors.Is(err, probo.ErrNotDocumentCommentAuthor):
		return probo.ErrNotDocumentCommentAuthor
	case errors.Is(err, probo.ErrDocumentCommentThreadResolved):
		return probo.ErrDocumentCommentThreadResolved
	case errors.Is(err, probo.ErrInvalidDocumentAnchor):
		return probo.ErrInvalidDocumentAnchor
	case errors.Is(err, collaboration.ErrQuoteNotFound):
		return collaboration.ErrQuoteNotFound
	case errors.Is(err, collaboration.ErrSuggestionNotInDocument):
		return collaboration.ErrSuggestionNotInDocument
	case errors.Is(err, collaboration.ErrOpenSuggestions):
		return collaboration.ErrOpenSuggestions
	case errors.Is(err, collaboration.ErrDocumentCollaborationNotSeeded):
		return collaboration.ErrDocumentCollaborationNotSeeded
	case errors.Is(err, collaboration.ErrSeedInProgress):
		return collaboration.ErrSeedInProgress
	}

	if errArchived, ok := errors.AsType[*collaboration.ErrDocumentArchived](err); ok {
		return errArchived
	}

	if errGenerated, ok := errors.AsType[*probo.ErrDocumentGenerated](err); ok {
		return errGenerated
	}

	return mcpx.MapError(ctx, r.logger, message, err)
}
