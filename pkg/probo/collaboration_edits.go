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
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/validator"
)

type (
	DocumentCollaborationTextEdit struct {
		ExpectedRevision int64
		Index            uint32
		Cursor           automerge.Cursor
		DeleteCount      int32
		Text             string
	}

	DocumentCollaborationEdit struct {
		Target      string
		Replacement string
	}

	CollaborationApplyEditsRequest struct {
		DocumentID gid.GID
		BaseHeads  []automerge.Hash
		Edits      []DocumentCollaborationEdit
		IdentityID gid.GID
		ProfileID  *gid.GID
	}

	CollaborationApplyContentRequest struct {
		DocumentID gid.GID
		Content    string
		IdentityID gid.GID
		ProfileID  *gid.GID
	}

	CollaborationWriteResult struct {
		Content string
		Heads   []automerge.Hash
	}
)

func (s *CollaborationService) ApplyTextEdit(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	edit DocumentCollaborationTextEdit,
) (int64, error) {
	if edit.DeleteCount < 0 {
		return 0, fmt.Errorf("collaboration text delete count cannot be negative")
	}

	opened, err := s.Open(ctx, scope, documentID)
	if err != nil {
		return 0, fmt.Errorf("cannot open collaboration for text edit: %w", err)
	}

	defer func() { _ = opened.Document.Close() }()

	if opened.NeedsSeed {
		return 0, ErrDocumentCollaborationNotSeeded
	}

	if edit.ExpectedRevision != opened.Revision {
		return 0, ErrDocumentCollaborationStale
	}

	text, err := opened.Document.Text("body")
	if err != nil {
		return 0, fmt.Errorf("cannot get collaboration body: %w", err)
	}

	index := edit.Index
	if len(edit.Cursor) > 0 {
		index, err = text.CursorPosition(edit.Cursor)
		if err != nil {
			return 0, fmt.Errorf("cannot resolve collaboration text cursor: %w", err)
		}
	}

	if err := text.Splice(index, edit.DeleteCount, edit.Text); err != nil {
		return 0, fmt.Errorf("cannot apply collaboration text edit: %w", err)
	}

	content, err := text.String()
	if err != nil {
		return 0, fmt.Errorf("cannot read edited collaboration body: %w", err)
	}

	if utf8.RuneCountInString(content) > contentMaxText {
		return 0, fmt.Errorf(
			"collaboration text exceeds maximum length of %d characters",
			contentMaxText,
		)
	}

	if _, err := opened.Document.Commit("Agent edit", time.Now()); err != nil {
		return 0, fmt.Errorf("cannot commit collaboration text edit: %w", err)
	}

	revision, err := s.Persist(ctx, scope, documentID, opened.Document)
	if err != nil {
		return 0, fmt.Errorf("cannot persist collaboration text edit: %w", err)
	}

	return revision, nil
}

func (s *CollaborationService) ReadText(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
) (string, int64, error) {
	opened, err := s.Open(ctx, scope, documentID)
	if err != nil {
		return "", 0, fmt.Errorf("cannot open collaboration for reading: %w", err)
	}

	defer func() { _ = opened.Document.Close() }()

	if opened.NeedsSeed {
		return "", 0, ErrDocumentCollaborationNotSeeded
	}

	text, err := opened.Document.Text("body")
	if err != nil {
		return "", 0, fmt.Errorf("cannot get collaboration body for reading: %w", err)
	}

	content, err := text.String()
	if err != nil {
		return "", 0, fmt.Errorf("cannot read collaboration body: %w", err)
	}

	return content, opened.Revision, nil
}

func (s *CollaborationService) Heads(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
) ([]automerge.Hash, error) {
	doc := &coredata.CollaborationDocument{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return doc.LoadByDocumentID(ctx, conn, scope, documentID)
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("cannot load collaboration heads: %w", err)
	}

	if !doc.Seeded {
		return nil, nil
	}

	heads, err := decodeHeads(doc.Heads)
	if err != nil {
		return nil, fmt.Errorf("cannot decode collaboration heads: %w", err)
	}

	return heads, nil
}

func (s *CollaborationService) Seeded(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
) (bool, error) {
	doc := &coredata.CollaborationDocument{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return doc.LoadByDocumentID(ctx, conn, scope, documentID)
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return false, nil
		}

		return false, fmt.Errorf("cannot load collaboration seed state: %w", err)
	}

	return doc.Seeded, nil
}

func (s *CollaborationService) ApplyContent(
	ctx context.Context,
	scope coredata.Scoper,
	req CollaborationApplyContentRequest,
) (*CollaborationWriteResult, error) {
	sanitized, err := prosemirror.SanitizeDocumentJSON(req.Content)
	if err != nil {
		return nil, fmt.Errorf("cannot sanitize collaboration content: %w", err)
	}

	if err := validateContentLength(sanitized); err != nil {
		return nil, err
	}

	var organizationID gid.GID

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			orgID, err := assertDraftEditable(
				ctx,
				scope,
				conn,
				req.DocumentID,
			)
			if err != nil {
				return err
			}

			organizationID = orgID

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	actorID, err := s.actorForIdentity(
		ctx,
		scope,
		req.DocumentID,
		organizationID,
		req.IdentityID,
		req.ProfileID,
	)
	if err != nil {
		return nil, err
	}

	opened, err := s.openWithActor(ctx, scope, req.DocumentID, actorID)
	if err != nil {
		return nil, fmt.Errorf("cannot open collaboration for content write: %w", err)
	}

	defer func() { _ = opened.Document.Close() }()

	message := "Update document content"
	if opened.NeedsSeed {
		message = "Seed collaboration document"
	}

	if err := writeSpans(opened.Document, sanitized, true); err != nil {
		if opened.NeedsSeed {
			_ = s.ReleaseSeed(ctx, scope, req.DocumentID)
		}

		return nil, err
	}

	if _, err := opened.Document.Commit(message, time.Now()); err != nil {
		if opened.NeedsSeed {
			_ = s.ReleaseSeed(ctx, scope, req.DocumentID)
		}

		return nil, fmt.Errorf("cannot commit collaboration content: %w", err)
	}

	if err := s.RecordActors(
		ctx,
		scope,
		req.DocumentID,
		organizationID,
		req.IdentityID,
		req.ProfileID,
		[][]byte{actorID[:]},
	); err != nil {
		return nil, fmt.Errorf("cannot record collaboration content actor: %w", err)
	}

	if _, err := s.Persist(
		ctx,
		scope,
		req.DocumentID,
		opened.Document,
	); err != nil {
		return nil, fmt.Errorf("cannot persist collaboration content: %w", err)
	}

	return writeResult(opened.Document)
}

func (s *CollaborationService) ApplyEdits(
	ctx context.Context,
	scope coredata.Scoper,
	req CollaborationApplyEditsRequest,
) (*CollaborationWriteResult, error) {
	if len(req.BaseHeads) == 0 {
		return nil, ErrDocumentCollaborationBaseHeadsUnknown
	}

	if len(req.Edits) == 0 {
		return nil, fmt.Errorf("collaboration edits require at least one edit")
	}

	var organizationID gid.GID

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			orgID, err := assertDraftEditable(
				ctx,
				scope,
				conn,
				req.DocumentID,
			)
			if err != nil {
				return err
			}

			organizationID = orgID

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	actorID, err := s.actorForIdentity(
		ctx,
		scope,
		req.DocumentID,
		organizationID,
		req.IdentityID,
		req.ProfileID,
	)
	if err != nil {
		return nil, err
	}

	opened, err := s.openWithActor(ctx, scope, req.DocumentID, actorID)
	if err != nil {
		return nil, fmt.Errorf("cannot open collaboration for edits: %w", err)
	}

	defer func() { _ = opened.Document.Close() }()

	if opened.NeedsSeed {
		_ = s.ReleaseSeed(ctx, scope, req.DocumentID)

		return nil, ErrDocumentCollaborationEditsNotSeeded
	}

	hasHeads, err := opened.Document.HasHeads(req.BaseHeads)
	if err != nil {
		return nil, fmt.Errorf("cannot verify collaboration base heads: %w", err)
	}

	if !hasHeads {
		return nil, ErrDocumentCollaborationBaseHeadsUnknown
	}

	content, err := materialize(ctx, opened.Document)
	if err != nil {
		return nil, fmt.Errorf("cannot materialize collaboration for edits: %w", err)
	}

	doc, err := prosemirror.Parse(content)
	if err != nil {
		return nil, fmt.Errorf("cannot parse collaboration document for edits: %w", err)
	}

	edits := make([]prosemirror.FragmentEdit, 0, len(req.Edits))
	for i, edit := range req.Edits {
		target, err := prosemirror.ParseFragment(edit.Target)
		if err != nil {
			return nil, fmt.Errorf("cannot parse collaboration edit %d target: %w", i, err)
		}

		replacement, err := prosemirror.ParseFragment(edit.Replacement)
		if err != nil {
			return nil, fmt.Errorf("cannot parse collaboration edit %d replacement: %w", i, err)
		}

		edits = append(
			edits,
			prosemirror.FragmentEdit{
				Target:      target,
				Replacement: replacement,
			},
		)
	}

	updated, err := prosemirror.ApplyFragmentEdits(doc, edits)
	if err != nil {
		if errors.Is(err, prosemirror.ErrFragmentEditNotFound) {
			return nil, fmt.Errorf("%w: %w", ErrDocumentCollaborationStale, err)
		}

		if errors.Is(err, prosemirror.ErrFragmentEditAmbiguous) {
			return nil, err
		}

		return nil, fmt.Errorf("cannot apply collaboration edits: %w", err)
	}

	encoded, err := json.Marshal(updated)
	if err != nil {
		return nil, fmt.Errorf("cannot encode edited collaboration document: %w", err)
	}

	sanitized, err := prosemirror.SanitizeDocumentJSON(string(encoded))
	if err != nil {
		return nil, fmt.Errorf("cannot sanitize edited collaboration document: %w", err)
	}

	if err := validateContentLength(sanitized); err != nil {
		return nil, err
	}

	if err := writeSpans(opened.Document, sanitized, false); err != nil {
		return nil, err
	}

	if _, err := opened.Document.Commit("Agent document edit", time.Now()); err != nil {
		return nil, fmt.Errorf("cannot commit collaboration edits: %w", err)
	}

	if err := s.RecordActors(
		ctx,
		scope,
		req.DocumentID,
		organizationID,
		req.IdentityID,
		req.ProfileID,
		[][]byte{actorID[:]},
	); err != nil {
		return nil, fmt.Errorf("cannot record collaboration edit actor: %w", err)
	}

	if _, err := s.Persist(
		ctx,
		scope,
		req.DocumentID,
		opened.Document,
	); err != nil {
		return nil, fmt.Errorf("cannot persist collaboration edits: %w", err)
	}

	return writeResult(opened.Document)
}

func (s *CollaborationService) actorForIdentity(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	organizationID gid.GID,
	identityID gid.GID,
	profileID *gid.GID,
) (automerge.ActorID, error) {
	var actors coredata.CollaborationActors

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return actors.LoadByDocumentID(
				ctx,
				conn,
				scope,
				documentID,
				0,
			)
		},
	)
	if err != nil {
		return automerge.ActorID{}, fmt.Errorf("cannot load collaboration actors: %w", err)
	}

	for _, actor := range actors {
		if actor.IdentityID != identityID {
			continue
		}

		if len(actor.ActorID) != len(automerge.ActorID{}) {
			continue
		}

		var actorID automerge.ActorID
		copy(actorID[:], actor.ActorID)

		return actorID, nil
	}

	actorID, err := automerge.NewActorID()
	if err != nil {
		return automerge.ActorID{}, fmt.Errorf("cannot create collaboration actor: %w", err)
	}

	if err := s.RecordActors(
		ctx,
		scope,
		documentID,
		organizationID,
		identityID,
		profileID,
		[][]byte{actorID[:]},
	); err != nil {
		return automerge.ActorID{}, fmt.Errorf("cannot store collaboration actor: %w", err)
	}

	return actorID, nil
}

func (s *CollaborationService) openWithActor(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	actorID automerge.ActorID,
) (*DocumentCollaboration, error) {
	opened, err := s.Open(ctx, scope, documentID)
	if err != nil {
		return nil, err
	}

	snapshot, err := opened.Document.Save()
	if err != nil {
		_ = opened.Document.Close()

		return nil, fmt.Errorf("cannot snapshot collaboration for actor swap: %w", err)
	}

	_ = opened.Document.Close()

	document, err := automerge.Load(snapshot, actorID)
	if err != nil {
		return nil, fmt.Errorf("cannot reload collaboration with writer actor: %w", err)
	}

	return &DocumentCollaboration{
		Document:    document,
		Revision:    opened.Revision,
		SeedContent: opened.SeedContent,
		NeedsSeed:   opened.NeedsSeed,
	}, nil
}

func writeResult(document *automerge.Document) (*CollaborationWriteResult, error) {
	content, err := materialize(context.Background(), document)
	if err != nil {
		return nil, fmt.Errorf("cannot materialize collaboration write result: %w", err)
	}

	heads, err := document.Heads()
	if err != nil {
		return nil, fmt.Errorf("cannot read collaboration write heads: %w", err)
	}

	return &CollaborationWriteResult{
		Content: content,
		Heads:   heads,
	}, nil
}

func validateContentLength(contentJSON string) error {
	v := validator.New()
	v.Check(
		contentJSON,
		"content",
		validator.ProseMirrorDocumentMaxTextLength(contentMaxText),
	)

	return v.Error()
}
