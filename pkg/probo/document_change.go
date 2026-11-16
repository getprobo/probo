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
	"fmt"
	"strings"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/collaboration"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/validator"
)

// DocumentChangeMode applies a change or stores it as a suggestion.
type DocumentChangeMode string

const (
	DocumentChangeModeApply   DocumentChangeMode = "APPLY"
	DocumentChangeModeSuggest DocumentChangeMode = "SUGGEST"
)

// DocumentChangeResult is the document after the change. Suggestion is set
// when the mode is SUGGEST.
type DocumentChangeResult struct {
	Document   *coredata.Document
	Suggestion *coredata.DocumentSuggestion
}

type ChangeDocumentTextRequest struct {
	Mode         DocumentChangeMode
	DocumentID   gid.GID
	SuggestionID gid.GID
	IdentityID   gid.GID
	Quote        string
	Replacement  string
	Occurrence   int
}

type ChangeDocumentFormatRequest struct {
	Mode         DocumentChangeMode
	DocumentID   gid.GID
	SuggestionID gid.GID
	IdentityID   gid.GID
	Mark         collaboration.FormatSuggestionMark
	Action       collaboration.FormatSuggestionAction
	Quote        string
	Occurrence   int
	Href         string
}

type ChangeDocumentTableRequest struct {
	Mode         DocumentChangeMode
	DocumentID   gid.GID
	SuggestionID gid.GID
	IdentityID   gid.GID
	Operation    collaboration.TableSuggestionOperation
	Quote        string
	Occurrence   int
	Before       bool
	Cells        []string
	Rows         [][]string
	HeaderRow    bool
}

type ChangeDocumentBlockRequest struct {
	Mode         DocumentChangeMode
	DocumentID   gid.GID
	SuggestionID gid.GID
	IdentityID   gid.GID
	Operation    collaboration.BlockOperation
	Quote        string
	Occurrence   int
	Before       *bool
	Block        collaboration.BlockKind
	Level        int
	Language     string
	Text         string
	Items        []string
	Checked      *bool
	FileID       gid.GID
	Src          string
	Alt          string
	Width        *int
	ProfileID    gid.GID
}

// ChangeText replaces quoted text, or stores that replacement as a suggestion.
func (s *DocumentService) ChangeText(
	ctx context.Context,
	scope coredata.Scoper,
	req ChangeDocumentTextRequest,
) (*DocumentChangeResult, error) {
	if err := validateChangeMode(req.Mode, req.DocumentID, req.SuggestionID); err != nil {
		return nil, err
	}

	v := validator.New()
	v.Check(req.IdentityID, "identity_id", validator.Required(), validator.GID(coredata.IdentityEntityType))
	v.Check(req.Quote, "quote", validator.Required(), validator.MaxLen(documentQuoteMaxLength))
	v.Check(req.Replacement, "replacement", validator.MaxLen(documentQuoteMaxLength))
	v.Check(req.Occurrence, "quote_occurrence", validator.Min(0))

	if err := v.Error(); err != nil {
		return nil, err
	}

	prepared, err := s.prepareDocumentChange(ctx, scope, req.Mode, req.DocumentID, req.SuggestionID, req.IdentityID, nil)
	if err != nil {
		return nil, err
	}

	return s.commitDocumentChange(ctx, scope, prepared, req.IdentityID, func(suggestionID string) (*automerge.Document, automerge.ActorID, error) {
		return s.paintTextChange(ctx, scope, prepared, req, suggestionID)
	})
}

// ChangeFormat sets or clears a style, or stores that change as a suggestion.
func (s *DocumentService) ChangeFormat(
	ctx context.Context,
	scope coredata.Scoper,
	req ChangeDocumentFormatRequest,
) (*DocumentChangeResult, error) {
	if err := validateChangeMode(req.Mode, req.DocumentID, req.SuggestionID); err != nil {
		return nil, err
	}

	v := validator.New()
	v.Check(req.IdentityID, "identity_id", validator.Required(), validator.GID(coredata.IdentityEntityType))
	v.Check(req.Mark, "mark", validator.Required(), validator.OneOfSlice(documentFormatMarks))
	v.Check(req.Action, "action", validator.Required(), validator.OneOfSlice(documentFormatActions))
	v.Check(req.Quote, "quote", validator.Required(), validator.MaxLen(documentQuoteMaxLength))
	v.Check(req.Occurrence, "quote_occurrence", validator.Min(0))
	v.Check(req.Href, "href", validator.MaxLen(documentQuoteMaxLength))

	if err := v.Error(); err != nil {
		return nil, err
	}

	format := collaboration.FormatSuggestion{
		Mark:       req.Mark,
		Action:     req.Action,
		Quote:      req.Quote,
		Occurrence: req.Occurrence,
		Href:       req.Href,
	}

	prepared, err := s.prepareDocumentChange(ctx, scope, req.Mode, req.DocumentID, req.SuggestionID, req.IdentityID, nil)
	if err != nil {
		return nil, err
	}

	return s.commitDocumentChange(ctx, scope, prepared, req.IdentityID, func(suggestionID string) (*automerge.Document, automerge.ActorID, error) {
		return s.paintFormatChange(ctx, scope, prepared, req.IdentityID, suggestionID, format)
	})
}

// ChangeTable edits a table, or stores that edit as a suggestion.
func (s *DocumentService) ChangeTable(
	ctx context.Context,
	scope coredata.Scoper,
	req ChangeDocumentTableRequest,
) (*DocumentChangeResult, error) {
	if err := validateChangeMode(req.Mode, req.DocumentID, req.SuggestionID); err != nil {
		return nil, err
	}

	v := validator.New()
	v.Check(req.IdentityID, "identity_id", validator.Required(), validator.GID(coredata.IdentityEntityType))
	v.Check(req.Operation, "operation", validator.Required(), validator.OneOfSlice(documentTableOperations))
	v.Check(req.Quote, "quote", validator.Required(), validator.MaxLen(documentQuoteMaxLength))
	v.Check(req.Occurrence, "quote_occurrence", validator.Min(0))

	cellLimit := documentTableMaxColumns
	if req.Operation == collaboration.TableSuggestionInsertColumn {
		cellLimit = documentTableMaxRows
	}

	v.Check(req.Cells, "cells", validator.MaxItems(cellLimit))
	v.CheckEach(req.Cells, "cells", func(index int, item any) {
		v.Check(item, fmt.Sprintf("cells[%d]", index), validator.MaxLen(documentQuoteMaxLength))
	})
	v.Check(req.Rows, "rows", validator.MaxItems(documentTableMaxRows))
	v.CheckEach(req.Rows, "rows", func(index int, item any) {
		v.Check(item, fmt.Sprintf("rows[%d]", index), validator.MaxItems(documentTableMaxColumns))
		v.CheckEach(item, fmt.Sprintf("rows[%d]", index), func(cell int, value any) {
			v.Check(value, fmt.Sprintf("rows[%d][%d]", index, cell), validator.MaxLen(documentQuoteMaxLength))
		})
	})

	if err := v.Error(); err != nil {
		return nil, err
	}

	table := collaboration.TableSuggestion{
		Operation:  req.Operation,
		Quote:      req.Quote,
		Occurrence: req.Occurrence,
		Before:     req.Before,
		Cells:      req.Cells,
		Rows:       req.Rows,
		HeaderRow:  req.HeaderRow,
	}

	prepared, err := s.prepareDocumentChange(ctx, scope, req.Mode, req.DocumentID, req.SuggestionID, req.IdentityID, nil)
	if err != nil {
		return nil, err
	}

	return s.commitDocumentChange(ctx, scope, prepared, req.IdentityID, func(suggestionID string) (*automerge.Document, automerge.ActorID, error) {
		return s.paintTableChange(ctx, scope, prepared, req.IdentityID, suggestionID, table)
	})
}

// ChangeBlock edits a heading, list, image, or other block, or stores that
// edit as a suggestion.
func (s *DocumentService) ChangeBlock(
	ctx context.Context,
	scope coredata.Scoper,
	req ChangeDocumentBlockRequest,
) (*DocumentChangeResult, error) {
	if err := validateChangeMode(req.Mode, req.DocumentID, req.SuggestionID); err != nil {
		return nil, err
	}

	v := validator.New()
	v.Check(req.IdentityID, "identity_id", validator.Required(), validator.GID(coredata.IdentityEntityType))
	v.Check(req.Operation, "operation", validator.Required(), validator.OneOfSlice(documentBlockOperations))
	v.Check(req.Quote, "quote", validator.MaxLen(documentQuoteMaxLength))
	v.Check(req.Text, "text", validator.MaxLen(documentQuoteMaxLength))
	v.Check(req.Occurrence, "quote_occurrence", validator.Min(0))
	v.Check(req.Items, "items", validator.MaxItems(documentTableMaxRows))
	v.CheckEach(req.Items, "items", func(index int, item any) {
		v.Check(item, fmt.Sprintf("items[%d]", index), validator.MaxLen(documentQuoteMaxLength))
	})

	if err := v.Error(); err != nil {
		return nil, err
	}

	var block collaboration.BlockChange

	prepared, err := s.prepareDocumentChange(
		ctx,
		scope,
		req.Mode,
		req.DocumentID,
		req.SuggestionID,
		req.IdentityID,
		func(ctx context.Context, conn pg.Querier, document *coredata.Document) error {
			resolved, err := resolveBlockChange(ctx, conn, scope, document, req)
			if err != nil {
				return err
			}

			block = resolved

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return s.commitDocumentChange(ctx, scope, prepared, req.IdentityID, func(suggestionID string) (*automerge.Document, automerge.ActorID, error) {
		block.SuggestionID = suggestionID

		return s.paintBlockChange(ctx, scope, prepared, req.IdentityID, suggestionID, block)
	})
}

// documentChange is the document and author resolved for one edit. Suggestion
// is set when the edit is stored as a redline instead of applied.
type documentChange struct {
	Document   *coredata.Document
	ProfileID  gid.GID
	Suggestion *coredata.DocumentSuggestion
}

// prepareDocumentChange loads the document, the caller's profile, and any
// extra rows the edit needs on one transaction. A suggestion edit also loads
// the open suggestion and checks that the caller wrote it.
func (s *DocumentService) prepareDocumentChange(
	ctx context.Context,
	scope coredata.Scoper,
	mode DocumentChangeMode,
	documentID gid.GID,
	suggestionID gid.GID,
	identityID gid.GID,
	attach func(ctx context.Context, conn pg.Querier, document *coredata.Document) error,
) (*documentChange, error) {
	prepared := &documentChange{}

	err := s.svc.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		switch mode {
		case DocumentChangeModeSuggest:
			suggestion := &coredata.DocumentSuggestion{}
			if err := suggestion.LoadByIDForUpdate(ctx, tx, scope, suggestionID); err != nil {
				return fmt.Errorf("cannot load document suggestion: %w", err)
			}

			if suggestion.Status != coredata.DocumentSuggestionStatusOpen {
				return ErrDocumentSuggestionClosed
			}

			document, err := s.editableReviewDocument(ctx, tx, scope, suggestion.DocumentID)
			if err != nil {
				return err
			}

			profile, err := s.reviewProfile(ctx, tx, scope, suggestion.OrganizationID, identityID)
			if err != nil {
				return err
			}

			if profile.ID != suggestion.AuthorProfileID {
				return ErrNotDocumentSuggestionAuthor
			}

			prepared.Document = document
			prepared.ProfileID = profile.ID
			prepared.Suggestion = suggestion
		default:
			document, err := s.editableReviewDocument(ctx, tx, scope, documentID)
			if err != nil {
				return err
			}

			profile, err := s.reviewProfile(ctx, tx, scope, document.OrganizationID, identityID)
			if err != nil {
				return err
			}

			prepared.Document = document
			prepared.ProfileID = profile.ID
		}

		if attach == nil {
			return nil
		}

		return attach(ctx, tx, prepared.Document)
	})
	if err != nil {
		return nil, err
	}

	return prepared, nil
}

func (s *DocumentService) commitDocumentChange(
	ctx context.Context,
	scope coredata.Scoper,
	prepared *documentChange,
	identityID gid.GID,
	paint func(suggestionID string) (*automerge.Document, automerge.ActorID, error),
) (*DocumentChangeResult, error) {
	suggestionID := ""
	if prepared.Suggestion != nil {
		suggestionID = prepared.Suggestion.ID.String()
	}

	painted, actorID, err := paint(suggestionID)
	if err != nil {
		return nil, err
	}

	defer func() { _ = painted.Close() }()

	if prepared.Suggestion != nil {
		if err := s.persistSuggestionChange(
			ctx,
			scope,
			prepared.Suggestion,
			identityID,
			prepared.ProfileID,
			painted,
			actorID,
		); err != nil {
			return nil, err
		}

		return &DocumentChangeResult{Document: prepared.Document, Suggestion: prepared.Suggestion}, nil
	}

	if err := s.finishDirectEdit(
		ctx,
		scope,
		prepared.Document,
		identityID,
		prepared.ProfileID,
		painted,
		actorID,
	); err != nil {
		return nil, err
	}

	return &DocumentChangeResult{Document: prepared.Document}, nil
}

func (s *DocumentService) paintTextChange(
	ctx context.Context,
	scope coredata.Scoper,
	prepared *documentChange,
	req ChangeDocumentTextRequest,
	suggestionID string,
) (*automerge.Document, automerge.ActorID, error) {
	profileID := prepared.ProfileID

	if suggestionID != "" {
		return s.collaboration.PaintSuggestion(
			ctx,
			scope,
			prepared.Document.ID,
			prepared.Document.OrganizationID,
			req.IdentityID,
			&profileID,
			suggestionID,
			req.Quote,
			req.Replacement,
			req.Occurrence,
		)
	}

	return s.collaboration.PaintDirectText(
		ctx,
		scope,
		prepared.Document.ID,
		prepared.Document.OrganizationID,
		req.IdentityID,
		&profileID,
		req.Quote,
		req.Replacement,
		req.Occurrence,
	)
}

func (s *DocumentService) paintFormatChange(
	ctx context.Context,
	scope coredata.Scoper,
	prepared *documentChange,
	identityID gid.GID,
	suggestionID string,
	format collaboration.FormatSuggestion,
) (*automerge.Document, automerge.ActorID, error) {
	profileID := prepared.ProfileID

	if suggestionID != "" {
		return s.collaboration.PaintFormatSuggestion(
			ctx,
			scope,
			prepared.Document.ID,
			prepared.Document.OrganizationID,
			identityID,
			&profileID,
			suggestionID,
			format,
		)
	}

	return s.collaboration.PaintDirectFormat(
		ctx,
		scope,
		prepared.Document.ID,
		prepared.Document.OrganizationID,
		identityID,
		&profileID,
		format,
	)
}

func (s *DocumentService) paintTableChange(
	ctx context.Context,
	scope coredata.Scoper,
	prepared *documentChange,
	identityID gid.GID,
	suggestionID string,
	table collaboration.TableSuggestion,
) (*automerge.Document, automerge.ActorID, error) {
	profileID := prepared.ProfileID

	if suggestionID != "" {
		return s.collaboration.PaintTableSuggestion(
			ctx,
			scope,
			prepared.Document.ID,
			prepared.Document.OrganizationID,
			identityID,
			&profileID,
			suggestionID,
			table,
		)
	}

	return s.collaboration.PaintDirectTable(
		ctx,
		scope,
		prepared.Document.ID,
		prepared.Document.OrganizationID,
		identityID,
		&profileID,
		table,
	)
}

func (s *DocumentService) paintBlockChange(
	ctx context.Context,
	scope coredata.Scoper,
	prepared *documentChange,
	identityID gid.GID,
	suggestionID string,
	block collaboration.BlockChange,
) (*automerge.Document, automerge.ActorID, error) {
	profileID := prepared.ProfileID

	if suggestionID != "" {
		return s.collaboration.PaintBlockSuggestion(
			ctx,
			scope,
			prepared.Document.ID,
			prepared.Document.OrganizationID,
			identityID,
			&profileID,
			block,
		)
	}

	return s.collaboration.PaintDirectBlock(
		ctx,
		scope,
		prepared.Document.ID,
		prepared.Document.OrganizationID,
		identityID,
		&profileID,
		block,
	)
}

func resolveBlockChange(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	document *coredata.Document,
	req ChangeDocumentBlockRequest,
) (collaboration.BlockChange, error) {
	change := collaboration.BlockChange{
		Operation:  req.Operation,
		Quote:      req.Quote,
		Occurrence: req.Occurrence,
		Before:     req.Before,
		Block:      req.Block,
		Level:      req.Level,
		Language:   req.Language,
		Text:       req.Text,
		Items:      req.Items,
		Checked:    req.Checked,
		Src:        req.Src,
		Alt:        req.Alt,
		Width:      req.Width,
	}

	if req.FileID != gid.Nil {
		file := &coredata.File{}
		if err := file.LoadActiveByID(ctx, conn, scope, req.FileID); err != nil {
			return collaboration.BlockChange{}, err
		}

		if file.OrganizationID != document.OrganizationID {
			return collaboration.BlockChange{}, coredata.ErrResourceNotFound
		}

		change.FileID = file.ID.String()
		change.FileName = file.FileName
		change.MimeType = file.MimeType
		change.Size = file.FileSize
	}

	if req.ProfileID != gid.Nil {
		profile := &coredata.MembershipProfile{}
		if err := profile.LoadByID(ctx, conn, scope, req.ProfileID); err != nil {
			return collaboration.BlockChange{}, err
		}

		if profile.OrganizationID != document.OrganizationID || strings.TrimSpace(profile.FullName) == "" {
			return collaboration.BlockChange{}, coredata.ErrResourceNotFound
		}

		change.ProfileID = profile.ID.String()
		change.Label = profile.FullName
	}

	return change, nil
}

func (s *DocumentService) finishDirectEdit(
	ctx context.Context,
	scope coredata.Scoper,
	document *coredata.Document,
	identityID gid.GID,
	profileID gid.GID,
	painted *automerge.Document,
	actorID automerge.ActorID,
) error {
	if err := s.collaboration.RecordActors(
		ctx,
		scope,
		document.ID,
		document.OrganizationID,
		identityID,
		&profileID,
		[][]byte{actorID[:]},
	); err != nil {
		return fmt.Errorf("cannot record document edit actor: %w", err)
	}

	if _, err := s.collaboration.Persist(ctx, scope, document.ID, painted); err != nil {
		return fmt.Errorf("cannot persist document edit: %w", err)
	}

	s.notifyReview(ctx, document.ID)

	return nil
}

func validateChangeMode(mode DocumentChangeMode, documentID gid.GID, suggestionID gid.GID) error {
	v := validator.New()
	v.Check(mode, "mode", validator.Required(), validator.OneOfSlice([]DocumentChangeMode{
		DocumentChangeModeApply,
		DocumentChangeModeSuggest,
	}))

	if mode == DocumentChangeModeApply {
		v.Check(documentID, "document_id", validator.Required(), validator.GID(coredata.DocumentEntityType))
	}

	if mode == DocumentChangeModeSuggest {
		v.Check(suggestionID, "suggestion_id", validator.Required(), validator.GID(coredata.DocumentSuggestionEntityType))
	}

	return v.Error()
}
