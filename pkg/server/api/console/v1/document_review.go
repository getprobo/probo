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

package console_v1

import (
	"context"
	"errors"

	"github.com/vikstrous/dataloadgen"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/collaboration"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/console/v1/dataloader"
	"go.probo.inc/probo/pkg/server/api/console/v1/types"
	"go.probo.inc/probo/pkg/server/gqlx"
	"go.probo.inc/probo/pkg/validator"
)

func documentCommentThreadFilter(
	filter *types.DocumentCommentThreadFilter,
) *coredata.DocumentCommentThreadFilter {
	if filter == nil {
		return nil
	}

	return coredata.NewDocumentCommentThreadFilter(filter.Resolved)
}

func documentSuggestionFilter(
	filter *types.DocumentSuggestionFilter,
) *coredata.DocumentSuggestionFilter {
	if filter == nil {
		return nil
	}

	return coredata.NewDocumentSuggestionFilter(filter.Status)
}

func (r *Resolver) loadReviewProfile(ctx context.Context, profileID gid.GID) (*types.Profile, error) {
	if _, err := r.authorize(ctx, profileID, iam.ActionMembershipProfileGet); err != nil {
		return nil, err
	}

	profile, err := dataloader.FromContext(ctx).Profile.Load(ctx, profileID)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) || errors.Is(err, dataloadgen.ErrNotFound) {
			return nil, gqlx.NotFound(ctx, err)
		}

		r.logger.ErrorCtx(ctx, "cannot load document review profile", log.Error(err))

		return nil, gqlx.Internal(ctx)
	}

	return types.NewProfile(profile), nil
}

func (r *Resolver) loadReviewDocument(ctx context.Context, documentID gid.GID) (*types.Document, error) {
	if _, err := r.authorize(ctx, documentID, probo.ActionDocumentGet); err != nil {
		return nil, err
	}

	document, err := dataloader.FromContext(ctx).Document.Load(ctx, documentID)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) || errors.Is(err, dataloadgen.ErrNotFound) {
			return nil, gqlx.NotFound(ctx, err)
		}

		r.logger.ErrorCtx(ctx, "cannot load document review document", log.Error(err))

		return nil, gqlx.Internal(ctx)
	}

	return types.NewDocument(document), nil
}

func (r *mutationResolver) setDocumentCommentThreadResolved(
	ctx context.Context,
	threadID gid.GID,
	resolved bool,
) (*types.ResolveDocumentCommentThreadPayload, error) {
	thread, err := r.setDocumentCommentThread(ctx, threadID, resolved)
	if err != nil {
		return nil, err
	}

	return &types.ResolveDocumentCommentThreadPayload{
		DocumentCommentThread: thread,
	}, nil
}

func (r *mutationResolver) setDocumentCommentThread(
	ctx context.Context,
	threadID gid.GID,
	resolved bool,
) (*types.DocumentCommentThread, error) {
	scope, err := r.authorize(ctx, threadID, probo.ActionDocumentCommentResolve)
	if err != nil {
		return nil, err
	}

	identity := authn.IdentityFromContext(ctx)
	thread, err := r.probo.Documents.SetCommentThreadResolved(
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
	case errors.Is(err, coredata.ErrResourceNotFound):
		return gqlx.NotFound(ctx, err)
	case errors.Is(err, coredata.ErrResourceAlreadyExists),
		errors.Is(err, probo.ErrDocumentCommentThreadResolved),
		errors.Is(err, collaboration.ErrOpenSuggestions),
		errors.Is(err, collaboration.ErrDocumentCollaborationNotSeeded),
		errors.Is(err, collaboration.ErrSeedInProgress):
		return gqlx.Conflict(ctx, err)
	case errors.Is(err, probo.ErrNotDocumentCommentAuthor),
		errors.Is(err, probo.ErrInvalidDocumentAnchor),
		errors.Is(err, collaboration.ErrQuoteNotFound),
		errors.Is(err, collaboration.ErrSuggestionNotInDocument):
		return gqlx.Invalid(ctx, err)
	}

	if errArchived, ok := errors.AsType[*collaboration.ErrDocumentArchived](err); ok {
		return gqlx.Conflict(ctx, errArchived)
	}

	if errGenerated, ok := errors.AsType[*probo.ErrDocumentGenerated](err); ok {
		return gqlx.Conflict(ctx, errGenerated)
	}

	if validationErrors, ok := errors.AsType[validator.ValidationErrors](err); ok {
		return gqlx.InvalidValidationErrors(ctx, validationErrors)
	}

	r.logger.ErrorCtx(ctx, message, log.Error(err))

	return gqlx.Internal(ctx)
}
