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
	"fmt"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/collaboration"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/types"
	"go.probo.inc/probo/pkg/validator"
)

func (r *Resolver) setDocumentCommentThreadResolved(
	ctx context.Context,
	threadID gid.GID,
	resolved bool,
) (*types.DocumentCommentThread, error) {
	scope, err := r.Authorize(ctx, threadID, probo.ActionDocumentComment)
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

func (r *Resolver) decideDocumentSuggestion(
	ctx context.Context,
	suggestionID gid.GID,
	accept bool,
) (*types.DocumentSuggestion, error) {
	scope, err := r.Authorize(ctx, suggestionID, probo.ActionDocumentUpdate)
	if err != nil {
		return nil, err
	}

	identity := authn.IdentityFromContext(ctx)

	suggestion, err := r.proboSvc.Documents.DecideSuggestion(
		ctx,
		scope,
		probo.DecideDocumentSuggestionRequest{
			SuggestionID: suggestionID,
			IdentityID:   identity.ID,
			Accept:       accept,
		},
	)
	if err != nil {
		return nil, r.documentReviewError(ctx, "cannot decide document suggestion", err)
	}

	return types.NewDocumentSuggestion(suggestion), nil
}

func (r *Resolver) documentReviewError(ctx context.Context, message string, err error) error {
	if errors.Is(err, coredata.ErrResourceNotFound) ||
		errors.Is(err, coredata.ErrResourceAlreadyExists) ||
		errors.Is(err, probo.ErrNotDocumentCommentAuthor) ||
		errors.Is(err, probo.ErrDocumentCommentThreadResolved) ||
		errors.Is(err, probo.ErrInvalidDocumentAnchor) ||
		errors.Is(err, collaboration.ErrQuoteNotFound) ||
		errors.Is(err, collaboration.ErrSuggestionNotInDocument) ||
		errors.Is(err, collaboration.ErrOpenSuggestions) {
		return err
	}

	if errArchived, ok := errors.AsType[*collaboration.ErrDocumentArchived](err); ok {
		return errArchived
	}

	if errGenerated, ok := errors.AsType[*probo.ErrDocumentGenerated](err); ok {
		return errGenerated
	}

	if validationErrors, ok := errors.AsType[validator.ValidationErrors](err); ok {
		return validationErrors
	}

	r.logger.ErrorCtx(ctx, message, log.Error(err))

	return fmt.Errorf("internal error")
}
