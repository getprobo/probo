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

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/automerge"
	automergeprosemirror "go.probo.inc/probo/pkg/automerge/prosemirror"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/validator"
)

type (
	DocumentCollaborationEdit struct {
		Target      string
		Replacement string
	}

	ApplyCollaborationEditsRequest struct {
		DocumentVersionID gid.GID
		BaseHeads         []automerge.Hash
		Edits             []DocumentCollaborationEdit
		IdentityID        gid.GID
		ProfileID         *gid.GID
	}

	ApplyCollaborationContentRequest struct {
		DocumentVersionID gid.GID
		Content           string
		IdentityID        gid.GID
		ProfileID         *gid.GID
	}

	CollaborationWriteResult struct {
		Content string
		Heads   []automerge.Hash
	}
)

var (
	ErrDocumentCollaborationBaseHeadsUnknown = errors.New(
		"document collaboration base heads are unknown; read the document again",
	)
	ErrDocumentCollaborationContentOnSeeded = errors.New(
		"seeded draft rejects full content; use fragment edits",
	)
	ErrDocumentCollaborationEditsNotSeeded = errors.New(
		"document collaboration is not initialized; send full content first",
	)
)

// CollaborationHeads returns the persisted Automerge heads for a seeded draft.
// An unseeded or missing collaboration state returns an empty slice.
func (s *DocumentService) CollaborationHeads(
	ctx context.Context,
	scope coredata.Scoper,
	documentVersionID gid.GID,
) ([]automerge.Hash, error) {
	state := &coredata.DocumentVersionAutomergeState{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return state.LoadByDocumentVersionID(ctx, conn, scope, documentVersionID)
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("cannot load collaboration heads: %w", err)
	}

	if !state.Seeded {
		return nil, nil
	}

	heads, err := decodeAutomergeHeads(state.Heads)
	if err != nil {
		return nil, fmt.Errorf("cannot decode collaboration heads: %w", err)
	}

	return heads, nil
}

// CollaborationSeeded reports whether the draft has a seeded Automerge body.
func (s *DocumentService) CollaborationSeeded(
	ctx context.Context,
	scope coredata.Scoper,
	documentVersionID gid.GID,
) (bool, error) {
	state := &coredata.DocumentVersionAutomergeState{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return state.LoadByDocumentVersionID(ctx, conn, scope, documentVersionID)
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return false, nil
		}

		return false, fmt.Errorf("cannot load collaboration seed state: %w", err)
	}

	return state.Seeded, nil
}

// ApplyCollaborationContent writes a full ProseMirror document into the live
// Automerge draft. Unseeded drafts are seeded from content. Seeded drafts
// replace the body on the latest heads (GraphQL / CLI / n8n replace path).
func (s *DocumentService) ApplyCollaborationContent(
	ctx context.Context,
	scope coredata.Scoper,
	req ApplyCollaborationContentRequest,
) (*CollaborationWriteResult, error) {
	sanitized, err := prosemirror.SanitizeDocumentJSON(req.Content)
	if err != nil {
		return nil, fmt.Errorf("cannot sanitize collaboration content: %w", err)
	}

	if err := validateCollaborationContentLength(sanitized); err != nil {
		return nil, err
	}

	version, err := s.loadCollaborationVersionForWrite(ctx, scope, req.DocumentVersionID)
	if err != nil {
		return nil, err
	}

	actorID, err := s.collaborationActorForIdentity(
		ctx,
		scope,
		req.DocumentVersionID,
		version.OrganizationID,
		req.IdentityID,
		req.ProfileID,
	)
	if err != nil {
		return nil, err
	}

	collaboration, err := s.openCollaborationWithActor(
		ctx,
		scope,
		req.DocumentVersionID,
		actorID,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot open collaboration for content write: %w", err)
	}

	defer func() { _ = collaboration.Document.Close() }()

	message := "Update document content"
	if collaboration.NeedsSeed {
		message = "Seed collaboration document"
	}

	if err := writeCollaborationSpans(collaboration.Document, sanitized, true); err != nil {
		if collaboration.NeedsSeed {
			_ = s.ReleaseCollaborationSeed(ctx, scope, req.DocumentVersionID)
		}

		return nil, err
	}

	if _, err := collaboration.Document.Commit(message, time.Now()); err != nil {
		if collaboration.NeedsSeed {
			_ = s.ReleaseCollaborationSeed(ctx, scope, req.DocumentVersionID)
		}

		return nil, fmt.Errorf("cannot commit collaboration content: %w", err)
	}

	if err := s.RecordCollaborationActors(
		ctx,
		scope,
		req.DocumentVersionID,
		version.OrganizationID,
		req.IdentityID,
		req.ProfileID,
		[][]byte{actorID[:]},
	); err != nil {
		return nil, fmt.Errorf("cannot record collaboration content actor: %w", err)
	}

	if _, err := s.PersistCollaboration(
		ctx,
		scope,
		req.DocumentVersionID,
		collaboration.Document,
	); err != nil {
		return nil, fmt.Errorf("cannot persist collaboration content: %w", err)
	}

	return collaborationWriteResult(collaboration.Document)
}

// ApplyCollaborationEdits replaces exact ProseMirror fragments in a seeded
// draft. Targets are matched against the latest materialized document. Missing
// or ambiguous targets conflict without writing.
func (s *DocumentService) ApplyCollaborationEdits(
	ctx context.Context,
	scope coredata.Scoper,
	req ApplyCollaborationEditsRequest,
) (*CollaborationWriteResult, error) {
	if len(req.BaseHeads) == 0 {
		return nil, ErrDocumentCollaborationBaseHeadsUnknown
	}

	if len(req.Edits) == 0 {
		return nil, fmt.Errorf("collaboration edits require at least one edit")
	}

	version, err := s.loadCollaborationVersionForWrite(ctx, scope, req.DocumentVersionID)
	if err != nil {
		return nil, err
	}

	actorID, err := s.collaborationActorForIdentity(
		ctx,
		scope,
		req.DocumentVersionID,
		version.OrganizationID,
		req.IdentityID,
		req.ProfileID,
	)
	if err != nil {
		return nil, err
	}

	collaboration, err := s.openCollaborationWithActor(
		ctx,
		scope,
		req.DocumentVersionID,
		actorID,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot open collaboration for edits: %w", err)
	}

	defer func() { _ = collaboration.Document.Close() }()

	if collaboration.NeedsSeed {
		_ = s.ReleaseCollaborationSeed(ctx, scope, req.DocumentVersionID)

		return nil, ErrDocumentCollaborationEditsNotSeeded
	}

	hasHeads, err := collaboration.Document.HasHeads(req.BaseHeads)
	if err != nil {
		return nil, fmt.Errorf("cannot verify collaboration base heads: %w", err)
	}

	if !hasHeads {
		return nil, ErrDocumentCollaborationBaseHeadsUnknown
	}

	content, err := materializeCollaboration(ctx, collaboration.Document)
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

	if err := validateCollaborationContentLength(sanitized); err != nil {
		return nil, err
	}

	if err := writeCollaborationSpans(collaboration.Document, sanitized, false); err != nil {
		return nil, err
	}

	if _, err := collaboration.Document.Commit("Agent document edit", time.Now()); err != nil {
		return nil, fmt.Errorf("cannot commit collaboration edits: %w", err)
	}

	if err := s.RecordCollaborationActors(
		ctx,
		scope,
		req.DocumentVersionID,
		version.OrganizationID,
		req.IdentityID,
		req.ProfileID,
		[][]byte{actorID[:]},
	); err != nil {
		return nil, fmt.Errorf("cannot record collaboration edit actor: %w", err)
	}

	if _, err := s.PersistCollaboration(
		ctx,
		scope,
		req.DocumentVersionID,
		collaboration.Document,
	); err != nil {
		return nil, fmt.Errorf("cannot persist collaboration edits: %w", err)
	}

	return collaborationWriteResult(collaboration.Document)
}

func (s *DocumentService) loadCollaborationVersionForWrite(
	ctx context.Context,
	scope coredata.Scoper,
	documentVersionID gid.GID,
) (*coredata.DocumentVersion, error) {
	var version *coredata.DocumentVersion

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			loaded, err := s.loadEditableCollaborationVersion(
				ctx,
				conn,
				scope,
				documentVersionID,
			)
			if err != nil {
				return err
			}

			version = loaded

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return version, nil
}

func (s *DocumentService) collaborationActorForIdentity(
	ctx context.Context,
	scope coredata.Scoper,
	documentVersionID gid.GID,
	organizationID gid.GID,
	identityID gid.GID,
	profileID *gid.GID,
) (automerge.ActorID, error) {
	var actors coredata.DocumentVersionAutomergeActors

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return actors.LoadByDocumentVersionID(
				ctx,
				conn,
				scope,
				documentVersionID,
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

	if err := s.RecordCollaborationActors(
		ctx,
		scope,
		documentVersionID,
		organizationID,
		identityID,
		profileID,
		[][]byte{actorID[:]},
	); err != nil {
		return automerge.ActorID{}, fmt.Errorf("cannot store collaboration actor: %w", err)
	}

	return actorID, nil
}

func (s *DocumentService) openCollaborationWithActor(
	ctx context.Context,
	scope coredata.Scoper,
	documentVersionID gid.GID,
	actorID automerge.ActorID,
) (*DocumentCollaboration, error) {
	collaboration, err := s.OpenCollaboration(ctx, scope, documentVersionID)
	if err != nil {
		return nil, err
	}

	snapshot, err := collaboration.Document.Save()
	if err != nil {
		_ = collaboration.Document.Close()

		return nil, fmt.Errorf("cannot snapshot collaboration for actor swap: %w", err)
	}

	_ = collaboration.Document.Close()

	document, err := automerge.Load(snapshot, actorID)
	if err != nil {
		return nil, fmt.Errorf("cannot reload collaboration with writer actor: %w", err)
	}

	return &DocumentCollaboration{
		Document:    document,
		Revision:    collaboration.Revision,
		SeedContent: collaboration.SeedContent,
		NeedsSeed:   collaboration.NeedsSeed,
	}, nil
}

func writeCollaborationSpans(
	document *automerge.Document,
	contentJSON string,
	createBody bool,
) error {
	spans, err := automergeprosemirror.ToSpans(contentJSON)
	if err != nil {
		return fmt.Errorf("cannot convert content to collaboration spans: %w", err)
	}

	text, err := document.Text("body")
	if err != nil {
		if !createBody {
			return fmt.Errorf("cannot get collaboration body: %w", err)
		}

		text, err = document.CreateText("body")
		if err != nil {
			return fmt.Errorf("cannot create collaboration body: %w", err)
		}
	}

	if err := text.UpdateSpans(spans, automergeprosemirror.UpdateSpansConfig()); err != nil {
		return fmt.Errorf("cannot write collaboration spans: %w", err)
	}

	return nil
}

func collaborationWriteResult(
	document *automerge.Document,
) (*CollaborationWriteResult, error) {
	content, err := materializeCollaboration(context.Background(), document)
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

func validateCollaborationContentLength(contentJSON string) error {
	v := validator.New()
	v.Check(
		contentJSON,
		"content",
		validator.ProseMirrorDocumentMaxTextLength(documentContentMaxTextLength),
	)

	return v.Error()
}

func decodeAutomergeHeads(encoded []byte) ([]automerge.Hash, error) {
	if len(encoded) == 0 {
		return nil, nil
	}

	if len(encoded)%len(automerge.Hash{}) != 0 {
		return nil, fmt.Errorf(
			"collaboration heads length %d is not a multiple of %d",
			len(encoded),
			len(automerge.Hash{}),
		)
	}

	heads := make([]automerge.Hash, 0, len(encoded)/len(automerge.Hash{}))
	for offset := 0; offset < len(encoded); offset += len(automerge.Hash{}) {
		var head automerge.Hash
		copy(head[:], encoded[offset:offset+len(automerge.Hash{})])
		heads = append(heads, head)
	}

	return heads, nil
}
