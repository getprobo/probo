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

package probo

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/collaboration"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

const (
	documentCommentMaxLength = 10000
	documentQuoteMaxLength   = 2000
	documentAnchorMaxBytes   = 512
)

var (
	ErrNotDocumentCommentAuthor      = errors.New("only the comment author can change this comment")
	ErrDocumentCommentThreadResolved = errors.New("comment thread is resolved")
	ErrInvalidDocumentAnchor         = errors.New("document comment anchor is invalid")
)

type (
	CreateDocumentCommentThreadRequest struct {
		DocumentID      gid.GID
		IdentityID      gid.GID
		Content         string
		Quote           string
		AnchorStart     string
		AnchorEnd       string
		SuggestionID    *gid.GID
		QuoteOccurrence int
	}

	CreateDocumentCommentRequest struct {
		ThreadID   gid.GID
		IdentityID gid.GID
		Content    string
	}

	UpdateDocumentCommentRequest struct {
		CommentID  gid.GID
		IdentityID gid.GID
		Content    string
	}

	ResolveDocumentCommentThreadRequest struct {
		ThreadID   gid.GID
		IdentityID gid.GID
		Resolved   bool
	}

	DecideDocumentSuggestionRequest struct {
		SuggestionID gid.GID
		IdentityID   gid.GID
		Accept       bool
	}
)

func (req *CreateDocumentCommentThreadRequest) Validate() error {
	v := validator.New()
	v.Check(req.DocumentID, "document_id", validator.Required(), validator.GID(coredata.DocumentEntityType))
	v.Check(req.IdentityID, "identity_id", validator.Required(), validator.GID(coredata.IdentityEntityType))
	v.Check(req.Content, "content", validator.Required(), validator.MaxLen(documentCommentMaxLength))
	v.Check(req.Quote, "quote", validator.MaxLen(documentQuoteMaxLength))
	v.Check(req.AnchorStart, "anchor_start", validator.MaxLen(documentAnchorMaxBytes*2))
	v.Check(req.AnchorEnd, "anchor_end", validator.MaxLen(documentAnchorMaxBytes*2))
	v.Check(req.SuggestionID, "suggestion_id", validator.GID(coredata.DocumentSuggestionEntityType))
	v.Check(req.QuoteOccurrence, "quote_occurrence", validator.Min(0))
	return v.Error()
}

func (req *CreateDocumentCommentRequest) Validate() error {
	v := validator.New()
	v.Check(req.ThreadID, "thread_id", validator.Required(), validator.GID(coredata.DocumentCommentThreadEntityType))
	v.Check(req.IdentityID, "identity_id", validator.Required(), validator.GID(coredata.IdentityEntityType))
	v.Check(req.Content, "content", validator.Required(), validator.MaxLen(documentCommentMaxLength))
	return v.Error()
}

func (req *UpdateDocumentCommentRequest) Validate() error {
	v := validator.New()
	v.Check(req.CommentID, "comment_id", validator.Required(), validator.GID(coredata.DocumentCommentEntityType))
	v.Check(req.IdentityID, "identity_id", validator.Required(), validator.GID(coredata.IdentityEntityType))
	v.Check(req.Content, "content", validator.Required(), validator.MaxLen(documentCommentMaxLength))
	return v.Error()
}

func (s *DocumentService) ListCommentThreads(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	cursor *page.Cursor[coredata.DocumentCommentThreadOrderField],
	filter *coredata.DocumentCommentThreadFilter,
) (*page.Page[*coredata.DocumentCommentThread, coredata.DocumentCommentThreadOrderField], error) {
	var threads coredata.DocumentCommentThreads

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		if err := threads.LoadByDocumentID(ctx, conn, scope, documentID, cursor, filter); err != nil {
			return fmt.Errorf("cannot load document comment threads: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return page.NewPage(threads, cursor), nil
}

func (s *DocumentService) CountCommentThreads(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	filter *coredata.DocumentCommentThreadFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		var threads coredata.DocumentCommentThreads
		var err error
		count, err = threads.CountByDocumentID(ctx, conn, scope, documentID, filter)
		if err != nil {
			return fmt.Errorf("cannot count document comment threads: %w", err)
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *DocumentService) GetCommentThread(
	ctx context.Context,
	scope coredata.Scoper,
	threadID gid.GID,
) (*coredata.DocumentCommentThread, error) {
	thread := &coredata.DocumentCommentThread{}

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		if err := thread.LoadByID(ctx, conn, scope, threadID); err != nil {
			return fmt.Errorf("cannot load document comment thread: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return thread, nil
}

func (s *DocumentService) ListComments(
	ctx context.Context,
	scope coredata.Scoper,
	threadID gid.GID,
	cursor *page.Cursor[coredata.DocumentCommentOrderField],
) (*page.Page[*coredata.DocumentComment, coredata.DocumentCommentOrderField], error) {
	var comments coredata.DocumentComments

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		if err := comments.LoadByThreadID(ctx, conn, scope, threadID, cursor); err != nil {
			return fmt.Errorf("cannot load document comments: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return page.NewPage(comments, cursor), nil
}

func (s *DocumentService) LoadCommentsByThreadIDs(
	ctx context.Context,
	scope coredata.Scoper,
	threadIDs []gid.GID,
) (coredata.DocumentComments, error) {
	var comments coredata.DocumentComments

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		if err := comments.LoadByThreadIDs(ctx, conn, scope, threadIDs); err != nil {
			return fmt.Errorf("cannot load document comments: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return comments, nil
}

func (s *DocumentService) CountComments(
	ctx context.Context,
	scope coredata.Scoper,
	threadID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		var comments coredata.DocumentComments
		var err error
		count, err = comments.CountByThreadID(ctx, conn, scope, threadID)
		if err != nil {
			return fmt.Errorf("cannot count document comments: %w", err)
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *DocumentService) GetComment(
	ctx context.Context,
	scope coredata.Scoper,
	commentID gid.GID,
) (*coredata.DocumentComment, error) {
	comment := &coredata.DocumentComment{}

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		if err := comment.LoadByID(ctx, conn, scope, commentID); err != nil {
			return fmt.Errorf("cannot load document comment: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return comment, nil
}

func (s *DocumentService) CreateCommentThread(
	ctx context.Context,
	scope coredata.Scoper,
	req CreateDocumentCommentThreadRequest,
) (*coredata.DocumentCommentThread, *coredata.DocumentComment, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, err
	}

	anchorStart, err := decodeDocumentAnchor(req.AnchorStart)
	if err != nil {
		return nil, nil, err
	}

	anchorEnd, err := decodeDocumentAnchor(req.AnchorEnd)
	if err != nil {
		return nil, nil, err
	}

	if (len(anchorStart) == 0) != (len(anchorEnd) == 0) {
		return nil, nil, ErrInvalidDocumentAnchor
	}

	now := time.Now()
	thread := &coredata.DocumentCommentThread{
		ID:          gid.New(scope.GetTenantID(), coredata.DocumentCommentThreadEntityType),
		AnchorStart: anchorStart,
		AnchorEnd:   anchorEnd,
		AnchorQuote: req.Quote,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	comment := &coredata.DocumentComment{
		ID:        gid.New(scope.GetTenantID(), coredata.DocumentCommentEntityType),
		ThreadID:  thread.ID,
		Content:   req.Content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if req.SuggestionID != nil {
		suggestion, err := s.GetSuggestion(ctx, scope, *req.SuggestionID)
		if err != nil {
			return nil, nil, err
		}

		if suggestion.DocumentID != req.DocumentID || suggestion.Status != coredata.DocumentSuggestionStatusOpen {
			return nil, nil, fmt.Errorf("cannot load document suggestion: %w", coredata.ErrResourceNotFound)
		}

		start, end, quote, err := s.suggestionAnchor(ctx, scope, req.DocumentID, suggestion.ID.String())
		if err != nil {
			return nil, nil, err
		}

		thread.SuggestionID = &suggestion.ID
		thread.AnchorStart = start
		thread.AnchorEnd = end
		if quote != "" {
			thread.AnchorQuote = quote
		}
	} else if len(thread.AnchorStart) == 0 {
		if req.Quote == "" {
			return nil, nil, ErrInvalidDocumentAnchor
		}

		start, end, err := s.quoteAnchor(ctx, scope, req.DocumentID, req.Quote, req.QuoteOccurrence)
		if err != nil {
			return nil, nil, err
		}

		thread.AnchorStart = start
		thread.AnchorEnd = end
	}

	err = s.svc.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		document, err := s.editableReviewDocument(ctx, tx, scope, req.DocumentID)
		if err != nil {
			return err
		}

		profile, err := s.reviewProfile(ctx, tx, scope, document.OrganizationID, req.IdentityID)
		if err != nil {
			return err
		}

		if req.SuggestionID != nil {
			// Re-check the suggestion under FOR UPDATE so a concurrent decision
			// cannot attach this thread to an already resolved suggestion.
			suggestion := &coredata.DocumentSuggestion{}
			if err := suggestion.LoadByIDForUpdate(ctx, tx, scope, *req.SuggestionID); err != nil {
				return fmt.Errorf("cannot load document suggestion: %w", err)
			}

			if suggestion.DocumentID != document.ID || suggestion.Status != coredata.DocumentSuggestionStatusOpen {
				return fmt.Errorf("cannot load document suggestion: %w", coredata.ErrResourceNotFound)
			}
		}

		thread.OrganizationID = document.OrganizationID
		thread.DocumentID = document.ID
		comment.OrganizationID = document.OrganizationID
		comment.AuthorProfileID = profile.ID

		if err := thread.Insert(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot insert document comment thread: %w", err)
		}

		if err := comment.Insert(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot insert document comment: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	s.notifyReview(ctx, thread.DocumentID)

	return thread, comment, nil
}

func (s *DocumentService) CreateComment(
	ctx context.Context,
	scope coredata.Scoper,
	req CreateDocumentCommentRequest,
) (*coredata.DocumentComment, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	comment := &coredata.DocumentComment{
		ID:        gid.New(scope.GetTenantID(), coredata.DocumentCommentEntityType),
		ThreadID:  req.ThreadID,
		Content:   req.Content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	var documentID gid.GID

	err := s.svc.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		thread := &coredata.DocumentCommentThread{}
		if err := thread.LoadByIDForUpdate(ctx, tx, scope, req.ThreadID); err != nil {
			return fmt.Errorf("cannot load document comment thread: %w", err)
		}

		if thread.ResolvedAt != nil {
			return ErrDocumentCommentThreadResolved
		}

		if _, err := s.editableReviewDocument(ctx, tx, scope, thread.DocumentID); err != nil {
			return err
		}

		profile, err := s.reviewProfile(ctx, tx, scope, thread.OrganizationID, req.IdentityID)
		if err != nil {
			return err
		}

		comment.OrganizationID = thread.OrganizationID
		comment.AuthorProfileID = profile.ID
		documentID = thread.DocumentID

		if err := comment.Insert(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot insert document comment: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	s.notifyReview(ctx, documentID)

	return comment, nil
}

func (s *DocumentService) UpdateComment(
	ctx context.Context,
	scope coredata.Scoper,
	req UpdateDocumentCommentRequest,
) (*coredata.DocumentComment, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	comment := &coredata.DocumentComment{}
	var documentID gid.GID

	err := s.svc.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := comment.LoadByID(ctx, tx, scope, req.CommentID); err != nil {
			return fmt.Errorf("cannot load document comment: %w", err)
		}

		thread := &coredata.DocumentCommentThread{}
		if err := thread.LoadByID(ctx, tx, scope, comment.ThreadID); err != nil {
			return fmt.Errorf("cannot load document comment thread: %w", err)
		}

		if _, err := s.editableReviewDocument(ctx, tx, scope, thread.DocumentID); err != nil {
			return err
		}

		profile, err := s.reviewProfile(ctx, tx, scope, comment.OrganizationID, req.IdentityID)
		if err != nil {
			return err
		}

		if profile.ID != comment.AuthorProfileID {
			return ErrNotDocumentCommentAuthor
		}

		comment.Content = req.Content
		comment.UpdatedAt = time.Now()
		documentID = thread.DocumentID

		if err := comment.Update(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot update document comment: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	s.notifyReview(ctx, documentID)

	return comment, nil
}

func (s *DocumentService) DeleteComment(
	ctx context.Context,
	scope coredata.Scoper,
	commentID gid.GID,
	identityID gid.GID,
) (*gid.GID, error) {
	var (
		documentID      gid.GID
		deletedThreadID *gid.GID
	)

	err := s.svc.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		comment := &coredata.DocumentComment{}
		if err := comment.LoadByID(ctx, tx, scope, commentID); err != nil {
			return fmt.Errorf("cannot load document comment: %w", err)
		}

		thread := &coredata.DocumentCommentThread{}
		if err := thread.LoadByID(ctx, tx, scope, comment.ThreadID); err != nil {
			return fmt.Errorf("cannot load document comment thread: %w", err)
		}

		if _, err := s.editableReviewDocument(ctx, tx, scope, thread.DocumentID); err != nil {
			return err
		}

		profile, err := s.reviewProfile(ctx, tx, scope, comment.OrganizationID, identityID)
		if err != nil {
			return err
		}

		if profile.ID != comment.AuthorProfileID {
			return ErrNotDocumentCommentAuthor
		}

		if err := comment.Delete(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot delete document comment: %w", err)
		}

		var comments coredata.DocumentComments
		remaining, err := comments.CountByThreadID(ctx, tx, scope, thread.ID)
		if err != nil {
			return fmt.Errorf("cannot count document comments: %w", err)
		}

		if remaining == 0 {
			if err := thread.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete document comment thread: %w", err)
			}

			deletedThreadID = &thread.ID
		}

		documentID = thread.DocumentID

		return nil
	})
	if err != nil {
		return nil, err
	}

	s.notifyReview(ctx, documentID)

	return deletedThreadID, nil
}

func (s *DocumentService) SetCommentThreadResolved(
	ctx context.Context,
	scope coredata.Scoper,
	req ResolveDocumentCommentThreadRequest,
) (*coredata.DocumentCommentThread, error) {
	v := validator.New()
	v.Check(req.ThreadID, "thread_id", validator.Required(), validator.GID(coredata.DocumentCommentThreadEntityType))
	v.Check(req.IdentityID, "identity_id", validator.Required(), validator.GID(coredata.IdentityEntityType))
	if err := v.Error(); err != nil {
		return nil, err
	}

	thread := &coredata.DocumentCommentThread{}

	err := s.svc.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := thread.LoadByID(ctx, tx, scope, req.ThreadID); err != nil {
			return fmt.Errorf("cannot load document comment thread: %w", err)
		}

		if _, err := s.editableReviewDocument(ctx, tx, scope, thread.DocumentID); err != nil {
			return err
		}

		profile, err := s.reviewProfile(ctx, tx, scope, thread.OrganizationID, req.IdentityID)
		if err != nil {
			return err
		}

		now := time.Now()
		if req.Resolved {
			thread.ResolvedAt = &now
			thread.ResolvedByProfileID = &profile.ID
		} else {
			thread.ResolvedAt = nil
			thread.ResolvedByProfileID = nil
		}

		thread.UpdatedAt = now
		if err := thread.Update(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot update document comment thread: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	s.notifyReview(ctx, thread.DocumentID)

	return thread, nil
}

func (s *DocumentService) ListSuggestions(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	cursor *page.Cursor[coredata.DocumentSuggestionOrderField],
	filter *coredata.DocumentSuggestionFilter,
) (*page.Page[*coredata.DocumentSuggestion, coredata.DocumentSuggestionOrderField], error) {
	var suggestions coredata.DocumentSuggestions

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		if err := suggestions.LoadByDocumentID(ctx, conn, scope, documentID, cursor, filter); err != nil {
			return fmt.Errorf("cannot load document suggestions: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return page.NewPage(suggestions, cursor), nil
}

func (s *DocumentService) CountSuggestions(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	filter *coredata.DocumentSuggestionFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		var suggestions coredata.DocumentSuggestions
		var err error
		count, err = suggestions.CountByDocumentID(ctx, conn, scope, documentID, filter)
		if err != nil {
			return fmt.Errorf("cannot count document suggestions: %w", err)
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *DocumentService) GetSuggestion(
	ctx context.Context,
	scope coredata.Scoper,
	suggestionID gid.GID,
) (*coredata.DocumentSuggestion, error) {
	suggestion := &coredata.DocumentSuggestion{}

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		if err := suggestion.LoadByID(ctx, conn, scope, suggestionID); err != nil {
			return fmt.Errorf("cannot load document suggestion: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return suggestion, nil
}

func (s *DocumentService) StartSuggestion(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	identityID gid.GID,
) (*coredata.DocumentSuggestion, error) {
	v := validator.New()
	v.Check(documentID, "document_id", validator.Required(), validator.GID(coredata.DocumentEntityType))
	v.Check(identityID, "identity_id", validator.Required(), validator.GID(coredata.IdentityEntityType))
	if err := v.Error(); err != nil {
		return nil, err
	}

	now := time.Now()
	suggestion := &coredata.DocumentSuggestion{
		ID:        gid.New(scope.GetTenantID(), coredata.DocumentSuggestionEntityType),
		Status:    coredata.DocumentSuggestionStatusOpen,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := s.svc.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		document, err := s.editableReviewDocument(ctx, tx, scope, documentID)
		if err != nil {
			return err
		}

		profile, err := s.reviewProfile(ctx, tx, scope, document.OrganizationID, identityID)
		if err != nil {
			return err
		}

		suggestion.OrganizationID = document.OrganizationID
		suggestion.DocumentID = document.ID
		suggestion.AuthorProfileID = profile.ID

		if err := suggestion.Insert(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot insert document suggestion: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	s.notifyReview(ctx, suggestion.DocumentID)

	return suggestion, nil
}

func (s *DocumentService) DecideSuggestion(
	ctx context.Context,
	scope coredata.Scoper,
	req DecideDocumentSuggestionRequest,
) (*coredata.DocumentSuggestion, error) {
	v := validator.New()
	v.Check(req.SuggestionID, "suggestion_id", validator.Required(), validator.GID(coredata.DocumentSuggestionEntityType))
	v.Check(req.IdentityID, "identity_id", validator.Required(), validator.GID(coredata.IdentityEntityType))
	if err := v.Error(); err != nil {
		return nil, err
	}

	suggestion := &coredata.DocumentSuggestion{}

	err := s.svc.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := suggestion.LoadByIDForUpdate(ctx, tx, scope, req.SuggestionID); err != nil {
			return fmt.Errorf("cannot load document suggestion: %w", err)
		}

		profile, err := s.reviewProfile(ctx, tx, scope, suggestion.OrganizationID, req.IdentityID)
		if err != nil {
			return err
		}

		if suggestion.Status != coredata.DocumentSuggestionStatusOpen {
			// Already decided: reconcile the body with the stored decision in
			// case an earlier attempt committed metadata but not the body.
			if err := s.applySuggestionDecision(ctx, scope, suggestion, suggestion.Status == coredata.DocumentSuggestionStatusAccepted); err != nil &&
				!errors.Is(err, collaboration.ErrSuggestionNotInDocument) {
				return err
			}
			return nil
		}

		now := time.Now()
		if req.Accept {
			suggestion.Status = coredata.DocumentSuggestionStatusAccepted
		} else {
			suggestion.Status = coredata.DocumentSuggestionStatusRejected
		}

		suggestion.ResolvedAt = &now
		suggestion.ResolvedByProfileID = &profile.ID
		suggestion.UpdatedAt = now
		if err := suggestion.UpdateIfOpen(ctx, tx, scope); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				if err := suggestion.LoadByID(ctx, tx, scope, req.SuggestionID); err != nil {
					return fmt.Errorf("cannot reload document suggestion: %w", err)
				}
				return nil
			}
			return fmt.Errorf("cannot update document suggestion: %w", err)
		}

		thread := &coredata.DocumentCommentThread{}
		err = thread.LoadOpenBySuggestionID(ctx, tx, scope, suggestion.ID)
		if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
			return fmt.Errorf("cannot load suggestion comment thread: %w", err)
		}

		if err == nil {
			thread.ResolvedAt = &now
			thread.ResolvedByProfileID = &profile.ID
			thread.UpdatedAt = now
			if err := thread.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot resolve suggestion comment thread: %w", err)
			}
		}

		// Mutate the collaborative body last: authorization and metadata
		// failures roll back the transaction before the body is touched. A
		// body already matching the decision is tolerated for recovery.
		if err := s.applySuggestionDecision(ctx, scope, suggestion, req.Accept); err != nil &&
			!errors.Is(err, collaboration.ErrSuggestionNotInDocument) {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	s.notifyReview(ctx, suggestion.DocumentID)

	return suggestion, nil
}

func (s *DocumentService) applySuggestionDecision(
	ctx context.Context,
	scope coredata.Scoper,
	suggestion *coredata.DocumentSuggestion,
	accept bool,
) error {
	opened, err := s.collaboration.Open(ctx, scope, suggestion.DocumentID)
	if err != nil {
		return fmt.Errorf("cannot open collaboration document: %w", err)
	}

	defer func() { _ = opened.Document.Close() }()

	changed, err := collaboration.ApplySuggestionDecision(opened.Document, suggestion.ID.String(), accept)
	if err != nil {
		return err
	}

	if !changed {
		return collaboration.ErrSuggestionNotInDocument
	}

	if _, err := s.collaboration.Persist(ctx, scope, suggestion.DocumentID, opened.Document); err != nil {
		return fmt.Errorf("cannot persist suggestion decision: %w", err)
	}

	return nil
}

func (s *DocumentService) notifyReview(ctx context.Context, documentID gid.GID) {
	if err := s.collaboration.NotifyReview(ctx, documentID); err != nil {
		s.svc.logger.WarnCtx(
			ctx,
			"cannot notify document review",
			log.String("document_id", documentID.String()),
			log.Error(err),
		)
	}
}

func (s *DocumentService) editableReviewDocument(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	documentID gid.GID,
) (*coredata.Document, error) {
	document := &coredata.Document{}
	if err := document.LoadByID(ctx, tx, scope, documentID); err != nil {
		return nil, fmt.Errorf("cannot load document: %w", err)
	}

	if document.ArchivedAt != nil {
		return nil, &collaboration.ErrDocumentArchived{}
	}

	if document.WriteMode == coredata.DocumentWriteModeGenerated {
		return nil, &ErrDocumentGenerated{}
	}

	return document, nil
}

func (s *DocumentService) reviewProfile(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	organizationID gid.GID,
	identityID gid.GID,
) (*coredata.MembershipProfile, error) {
	profile := &coredata.MembershipProfile{}
	if err := profile.LoadByIdentityIDAndOrganizationID(ctx, tx, scope, identityID, organizationID); err != nil {
		return nil, fmt.Errorf("cannot load membership profile: %w", err)
	}

	return profile, nil
}

func (s *DocumentService) suggestionAnchor(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	suggestionID string,
) ([]byte, []byte, string, error) {
	opened, err := s.collaboration.Open(ctx, scope, documentID)
	if err != nil {
		return nil, nil, "", fmt.Errorf("cannot open collaboration document: %w", err)
	}

	defer func() { _ = opened.Document.Close() }()

	start, end, quote, err := collaboration.SuggestionCursors(opened.Document, suggestionID)
	if err != nil {
		return nil, nil, "", err
	}

	return start, end, quote, nil
}

func (s *DocumentService) quoteAnchor(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	quote string,
	occurrence int,
) ([]byte, []byte, error) {
	opened, err := s.collaboration.Open(ctx, scope, documentID)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot open collaboration document: %w", err)
	}

	defer func() { _ = opened.Document.Close() }()

	start, end, err := collaboration.QuoteCursors(opened.Document, quote, occurrence)
	if err != nil {
		return nil, nil, err
	}

	return start, end, nil
}

func decodeDocumentAnchor(value string) ([]byte, error) {
	if value == "" {
		return nil, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(value)
	}

	if err != nil || len(decoded) == 0 || len(decoded) > documentAnchorMaxBytes {
		return nil, ErrInvalidDocumentAnchor
	}

	return decoded, nil
}
