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
	"errors"
	"fmt"
	"strings"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func (s *CollaborationService) Open(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
) (*DocumentCollaboration, error) {
	var opened *DocumentCollaboration

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			doc, err := lockCollaborationDocument(ctx, tx, scope, documentID)
			if err != nil {
				return err
			}

			now := time.Now()
			if !doc.Seeded {
				if seedClaimActive(doc, now) {
					return ErrSeedInProgress
				}

				automergeDocument, _, err := s.seedUnderDocumentLock(
					ctx,
					tx,
					scope,
					doc,
					nil,
					nil,
					now,
				)
				if err != nil {
					return err
				}

				opened = &DocumentCollaboration{
					Document: automergeDocument,
					Revision: doc.Revision,
				}

				return nil
			}

			changes, err := s.collectChanges(
				ctx,
				tx,
				scope,
				documentID,
				doc.SnapshotRevision,
				doc.ChangeRevision,
			)
			if err != nil {
				return err
			}

			automergeDocument, err := loadAutomergeDocument(ctx, doc.Snapshot)
			if err != nil {
				return fmt.Errorf("cannot open collaboration snapshot: %w", err)
			}

			if len(changes) > 0 {
				if err := automergeDocument.ApplyChanges(changes); err != nil {
					_ = automergeDocument.Close()

					return fmt.Errorf("cannot apply collaboration changes: %w", err)
				}
			}

			opened = &DocumentCollaboration{
				Document: automergeDocument,
				Revision: doc.Revision,
			}

			return nil
		},
	)
	if err != nil {
		if opened != nil && opened.Document != nil {
			_ = opened.Document.Close()
		}

		return nil, err
	}

	return opened, nil
}

// openIfSeeded loads a collaboration document that is already seeded. It does
// not create a row or take the seed lease.
func (s *CollaborationService) openIfSeeded(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
) (*DocumentCollaboration, error) {
	doc, changes, err := s.readConsistentCollaboration(ctx, scope, documentID)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, ErrDocumentCollaborationNotSeeded
		}

		return nil, err
	}

	if !doc.Seeded {
		if seedClaimActive(doc, time.Now()) {
			return nil, ErrSeedInProgress
		}

		return nil, ErrDocumentCollaborationNotSeeded
	}

	automergeDocument, err := loadAutomergeDocument(ctx, doc.Snapshot)
	if err != nil {
		return nil, fmt.Errorf("cannot open collaboration snapshot: %w", err)
	}

	if len(changes) > 0 {
		if err := automergeDocument.ApplyChanges(changes); err != nil {
			_ = automergeDocument.Close()

			return nil, fmt.Errorf("cannot apply collaboration changes: %w", err)
		}
	}

	return &DocumentCollaboration{
		Document: automergeDocument,
		Revision: doc.Revision,
	}, nil
}

func (s *CollaborationService) ReleaseSeed(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	claimedAt time.Time,
) error {
	if claimedAt.IsZero() {
		return nil
	}

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := coredata.ReleaseCollaborationSeedClaim(
				ctx,
				tx,
				scope,
				documentID,
				claimedAt,
			); err != nil {
				return fmt.Errorf("cannot release collaboration document seed: %w", err)
			}

			return nil
		},
	)
}

// lockCollaborationDocument returns the locked collaboration document for a
// document, creating an empty unseeded one the first time the document is
// opened.
func lockCollaborationDocument(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	documentID gid.GID,
) (*coredata.CollaborationDocument, error) {
	doc := &coredata.CollaborationDocument{}

	err := doc.LoadByDocumentIDForUpdate(ctx, tx, scope, documentID)
	if err == nil {
		return doc, nil
	}

	if !errors.Is(err, coredata.ErrResourceNotFound) {
		return nil, fmt.Errorf("cannot load collaboration document: %w", err)
	}

	document := &coredata.Document{}
	if err := document.LoadByID(ctx, tx, scope, documentID); err != nil {
		return nil, fmt.Errorf("cannot load document for collaboration: %w", err)
	}

	empty, err := newAutomergeDocument(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot create empty collaboration document: %w", err)
	}

	defer func() { _ = empty.Close() }()

	snapshot, err := empty.Save()
	if err != nil {
		return nil, fmt.Errorf("cannot save empty collaboration document: %w", err)
	}

	now := time.Now()
	fresh := coredata.CollaborationDocument{
		DocumentID:       documentID,
		OrganizationID:   document.OrganizationID,
		Snapshot:         snapshot,
		Heads:            encodeHeads(nil),
		Revision:         1,
		SnapshotRevision: 1,
		ChangeRevision:   1,
		Seeded:           false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if _, err := fresh.InsertIfAbsent(ctx, tx, scope); err != nil {
		return nil, fmt.Errorf("cannot create collaboration document: %w", err)
	}

	// Re-read under the row lock so a concurrent creator's row wins
	// consistently instead of this transaction working from stale values.
	if err := doc.LoadByDocumentIDForUpdate(ctx, tx, scope, documentID); err != nil {
		return nil, fmt.Errorf("cannot lock created collaboration document: %w", err)
	}

	return doc, nil
}

func seedClaimActive(doc *coredata.CollaborationDocument, now time.Time) bool {
	return doc.SeedClaimedAt != nil && doc.SeedClaimedAt.After(now.Add(-seedLease))
}

// seedUnderDocumentLock builds the initial CRDT from content while the
// document row is locked, and commits that seed before the lock is released.
// A nil content reads the locked row. The returned document stays open.
func (s *CollaborationService) seedUnderDocumentLock(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	doc *coredata.CollaborationDocument,
	content *string,
	actorID *automerge.ActorID,
	now time.Time,
) (*automerge.Document, string, error) {
	document := &coredata.Document{}
	if err := document.LoadByIDForUpdate(ctx, tx, scope, doc.DocumentID); err != nil {
		return nil, "", fmt.Errorf("cannot lock document for collaboration seed: %w", err)
	}

	if document.ArchivedAt != nil {
		return nil, "", &ErrDocumentArchived{}
	}

	body := document.Content
	if content != nil {
		body = *content
	}

	if strings.TrimSpace(body) == "" {
		body = emptyLiveDocument
	}

	var (
		automergeDocument *automerge.Document
		err               error
	)
	if actorID != nil {
		automergeDocument, err = automerge.Load(doc.Snapshot, *actorID)
	} else {
		automergeDocument, err = loadAutomergeDocument(ctx, doc.Snapshot)
	}

	if err != nil {
		return nil, "", err
	}

	if err := writeSpans(automergeDocument, body, true); err != nil {
		_ = automergeDocument.Close()

		return nil, "", err
	}

	if _, err := automergeDocument.Commit("Seed collaboration document", now); err != nil {
		_ = automergeDocument.Close()

		return nil, "", fmt.Errorf("cannot commit collaboration seed: %w", err)
	}

	snapshot, err := automergeDocument.Save()
	if err != nil {
		_ = automergeDocument.Close()

		return nil, "", fmt.Errorf("cannot save collaboration seed: %w", err)
	}

	if len(snapshot) > snapshotMaxBytes {
		_ = automergeDocument.Close()

		return nil, "", &ErrCollaborationStateTooLarge{Size: len(snapshot)}
	}

	heads, err := automergeDocument.Heads()
	if err != nil {
		_ = automergeDocument.Close()

		return nil, "", fmt.Errorf("cannot read collaboration seed heads: %w", err)
	}

	if err := coredata.DeleteCollaborationChangesThroughRevision(
		ctx,
		tx,
		scope,
		doc.DocumentID,
		doc.ChangeRevision,
	); err != nil {
		_ = automergeDocument.Close()

		return nil, "", err
	}

	doc.Snapshot = snapshot
	doc.Heads = encodeHeads(heads)
	doc.Seeded = true
	doc.SeedClaimedAt = nil
	doc.Revision++
	doc.SnapshotRevision = doc.Revision
	doc.ChangeRevision = doc.Revision
	doc.UpdatedAt = now
	if err := doc.Update(ctx, tx, scope); err != nil {
		_ = automergeDocument.Close()

		return nil, "", fmt.Errorf("cannot store collaboration seed: %w", err)
	}

	return automergeDocument, body, nil
}
