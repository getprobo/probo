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
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func (s *CollaborationService) Open(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
) (*DocumentCollaboration, error) {
	var (
		snapshot         []byte
		revision         int64
		snapshotRevision int64
		changeRevision   int64
		seedContent      string
		needsSeed        bool
	)

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			doc, err := lockCollaborationDocument(ctx, tx, scope, documentID)
			if err != nil {
				return err
			}

			snapshot = append([]byte(nil), doc.Snapshot...)

			needsSeed, err = claimSeed(ctx, tx, scope, doc, time.Now())
			if err != nil {
				return fmt.Errorf("cannot claim collaboration seed: %w", err)
			}

			if needsSeed {
				content, err := loadSeedContent(
					ctx,
					scope,
					tx,
					documentID,
				)
				if err != nil {
					return fmt.Errorf("cannot load seed content: %w", err)
				}

				seedContent = content
			}

			revision = doc.Revision
			snapshotRevision = doc.SnapshotRevision
			changeRevision = doc.ChangeRevision

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	document, err := loadAutomergeDocument(ctx, snapshot)
	if err != nil {
		return nil, fmt.Errorf("cannot open collaboration snapshot: %w", err)
	}

	if err := s.loadChanges(
		ctx,
		scope,
		documentID,
		document,
		snapshotRevision,
		changeRevision,
	); err != nil {
		_ = document.Close()

		return nil, fmt.Errorf("cannot load collaboration changes: %w", err)
	}

	return &DocumentCollaboration{
		Document:    document,
		Revision:    revision,
		SeedContent: seedContent,
		NeedsSeed:   needsSeed,
	}, nil
}

func (s *CollaborationService) ReleaseSeed(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			doc := &coredata.CollaborationDocument{}
			if err := doc.LoadByDocumentIDForUpdate(ctx, tx, scope, documentID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return nil
				}

				return fmt.Errorf("cannot lock collaboration document seed: %w", err)
			}

			if doc.Seeded || doc.SeedClaimedAt == nil {
				return nil
			}

			doc.SeedClaimedAt = nil
			doc.Revision++

			doc.UpdatedAt = time.Now()
			if err := doc.Update(ctx, tx, scope); err != nil {
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

func claimSeed(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	doc *coredata.CollaborationDocument,
	now time.Time,
) (bool, error) {
	if doc.Seeded {
		return false, nil
	}

	if doc.SeedClaimedAt != nil && doc.SeedClaimedAt.After(now.Add(-seedLease)) {
		return false, nil
	}

	doc.SeedClaimedAt = new(now)
	doc.Revision++

	doc.UpdatedAt = now
	if err := doc.Update(ctx, tx, scope); err != nil {
		return false, fmt.Errorf("cannot update collaboration seed claim: %w", err)
	}

	return true, nil
}
