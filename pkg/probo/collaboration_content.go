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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

const emptyLiveDocument = `{"type":"doc","content":[{"type":"paragraph"}]}`

// MaterializedContent returns the live ProseMirror JSON. A seeded CRDT wins.
// fallback is the document row's content, used until the CRDT exists.
func (s *CollaborationService) MaterializedContent(
	ctx context.Context,
	scope coredata.Scoper,
	conn pg.Querier,
	documentID gid.GID,
	fallback string,
) (string, error) {
	doc := &coredata.CollaborationDocument{}
	err := doc.LoadByDocumentID(ctx, conn, scope, documentID)
	if errors.Is(err, coredata.ErrResourceNotFound) || (err == nil && !doc.Seeded) {
		return fallback, nil
	}

	if err != nil {
		return "", fmt.Errorf("cannot load collaboration document: %w", err)
	}

	automergeDocument, err := loadAutomergeDocument(ctx, doc.Snapshot)
	if err != nil {
		return "", err
	}

	defer func() { _ = automergeDocument.Close() }()

	if err := s.loadChangesFromConn(
		ctx,
		scope,
		conn,
		documentID,
		automergeDocument,
		doc.SnapshotRevision,
		doc.ChangeRevision,
	); err != nil {
		return "", err
	}

	content, err := materialize(ctx, automergeDocument)
	if err != nil {
		return "", fmt.Errorf("cannot materialize collaboration document: %w", err)
	}

	return content, nil
}

// PutContent replaces the live CRDT with content. It is the write used when
// creating a document, regenerating one, or restoring the published snapshot.
func (s *CollaborationService) PutContent(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	documentID gid.GID,
	content string,
) error {
	if strings.TrimSpace(content) == "" {
		content = emptyLiveDocument
	}

	doc, err := lockCollaborationDocument(ctx, tx, scope, documentID)
	if err != nil {
		return err
	}

	automergeDocument, err := loadAutomergeDocument(ctx, doc.Snapshot)
	if err != nil {
		return err
	}

	defer func() { _ = automergeDocument.Close() }()

	if err := s.loadChangesFromConn(
		ctx,
		scope,
		tx,
		documentID,
		automergeDocument,
		doc.SnapshotRevision,
		doc.ChangeRevision,
	); err != nil {
		return err
	}

	if err := writeSpans(automergeDocument, content, !doc.Seeded); err != nil {
		return err
	}

	if _, err := automergeDocument.Commit("Set document content", time.Now()); err != nil {
		return fmt.Errorf("cannot commit collaboration content: %w", err)
	}

	snapshot, err := automergeDocument.Save()
	if err != nil {
		return fmt.Errorf("cannot save collaboration content: %w", err)
	}

	if len(snapshot) > snapshotMaxBytes {
		return &ErrCollaborationStateTooLarge{Size: len(snapshot)}
	}

	heads, err := automergeDocument.Heads()
	if err != nil {
		return fmt.Errorf("cannot read collaboration heads: %w", err)
	}

	if err := coredata.DeleteCollaborationChangesThroughRevision(
		ctx,
		tx,
		scope,
		documentID,
		doc.ChangeRevision,
	); err != nil {
		return err
	}

	now := time.Now()
	doc.Snapshot = snapshot
	doc.Heads = encodeHeads(heads)
	doc.Seeded = true
	doc.SeedClaimedAt = nil
	doc.Revision++
	doc.SnapshotRevision = doc.Revision
	doc.ChangeRevision = doc.Revision
	doc.UpdatedAt = now

	if err := doc.Update(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot store collaboration content: %w", err)
	}

	return nil
}
