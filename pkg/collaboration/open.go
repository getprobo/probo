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

package collaboration

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// Open returns the live collaboration document. It takes no row lock: the
// canonical state is read consistently and the CRDT is built outside any
// transaction.
func (s *Service) Open(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
) (*DocumentCollaboration, error) {
	doc, changes, err := s.readConsistentCollaboration(ctx, scope, documentID)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, ErrDocumentCollaborationNotFound
		}

		return nil, err
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

// lockCollaborationDocument returns the locked collaboration document of a
// document.
func lockCollaborationDocument(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	documentID gid.GID,
) (*coredata.CollaborationDocument, error) {
	doc := &coredata.CollaborationDocument{}

	if err := doc.LoadByDocumentIDForUpdate(ctx, tx, scope, documentID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, ErrDocumentCollaborationNotFound
		}

		return nil, fmt.Errorf("cannot load collaboration document: %w", err)
	}

	return doc, nil
}

// lockOrCreateCollaborationDocument returns the locked collaboration document
// of a document, creating an empty one when the document has none yet. The
// caller writes the first content in the same transaction.
func lockOrCreateCollaborationDocument(
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
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if _, err := fresh.Insert(ctx, tx, scope); err != nil {
		return nil, fmt.Errorf("cannot create collaboration document: %w", err)
	}

	// Re-read under the row lock so a concurrent creator's row wins
	// consistently instead of this transaction working from stale values.
	if err := doc.LoadByDocumentIDForUpdate(ctx, tx, scope, documentID); err != nil {
		return nil, fmt.Errorf("cannot lock created collaboration document: %w", err)
	}

	return doc, nil
}
