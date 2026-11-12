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
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/realtime"
	"go.probo.inc/probo/pkg/richtext"
)

const (
	// stableReadAttempts bounds how often persist re-reads a live document
	// that peers keep editing.
	stableReadAttempts = 3
	// persistCatchUpAttempts bounds how often persist catches the live
	// document up with a canonical document other writers keep advancing.
	persistCatchUpAttempts = 3
)

var (
	errCollaborationDocumentBusy = errors.New(
		"collaboration document changed while it was read",
	)
	errCollaborationPersistContended = errors.New(
		"collaboration document kept advancing while it was persisted",
	)
)

// Persist appends the changes of a live document to the canonical change log
// and returns the canonical revision.
func (s *Service) Persist(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	document *automerge.Document,
) (int64, error) {
	for range persistCatchUpAttempts {
		revision, compactAt, persisted, err := s.appendLiveChanges(ctx, scope, documentID, document)
		if err != nil {
			return 0, err
		}

		if persisted {
			if compactAt != nil {
				s.compact(ctx, scope, documentID, document, compactAt)
			}

			return revision, nil
		}

		// Another writer advanced the canonical document. Catch up with no
		// lock held, then try again.
		if err := s.catchUp(ctx, scope, documentID, document); err != nil {
			return 0, err
		}
	}

	return 0, errCollaborationPersistContended
}

// appendLiveChanges persists a live document that already contains everything
// the canonical document has. The merged state is then the live document
// itself, so nothing is loaded while the row is locked: the lock only covers
// SQL and reads of the in-memory document. It reports false, and writes
// nothing, when the canonical document holds changes the live one lacks. When
// the change log is long enough to compact, it also returns the heads the
// canonical document was left at.
func (s *Service) appendLiveChanges(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	document *automerge.Document,
) (int64, []automerge.Hash, bool, error) {
	var (
		revision  int64
		compactAt []automerge.Hash
		persisted bool
	)

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			organizationID, err := assertDraftEditable(
				ctx,
				scope,
				tx,
				documentID,
			)
			if err != nil {
				return err
			}

			doc, err := lockCollaborationDocument(ctx, tx, scope, documentID)
			if err != nil {
				return err
			}

			canonicalHeads, err := decodeHeads(doc.Heads)
			if err != nil {
				return fmt.Errorf("cannot decode canonical collaboration heads: %w", err)
			}

			// Holding every canonical head means holding the whole canonical
			// history, so the live document is the merged document.
			covered, err := document.HasHeads(canonicalHeads)
			if err != nil {
				return fmt.Errorf("cannot inspect local collaboration heads: %w", err)
			}

			if !covered {
				return nil
			}

			heads, changes, stable, err := readStableChanges(document, canonicalHeads)
			if err != nil {
				return err
			}

			if !stable {
				return errCollaborationDocumentBusy
			}

			persisted = true
			revision = doc.Revision

			if len(changes) == 0 {
				// A compaction that failed earlier is retried even though
				// this persist has nothing to append.
				if doc.ChangeRevision-doc.SnapshotRevision >= compactionChanges {
					compactAt = heads
				}

				return nil
			}

			now := time.Now()

			for _, change := range changes {
				doc.ChangeRevision++

				storedChange := coredata.CollaborationChange{
					DocumentID:     documentID,
					OrganizationID: organizationID,
					Revision:       doc.ChangeRevision,
					ChangeHash:     append([]byte(nil), change.Hash[:]...),
					ChangeBytes:    change.Bytes,
					CreatedAt:      now,
				}
				if err := storedChange.Insert(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot append collaboration change: %w", err)
				}
			}

			doc.Revision++
			doc.Heads = encodeHeads(heads)

			if doc.ChangeRevision-doc.SnapshotRevision >= compactionChanges {
				compactAt = heads
			}

			doc.UpdatedAt = now
			if err := doc.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot persist collaboration document: %w", err)
			}

			if _, err := tx.Exec(
				ctx,
				`SELECT pg_notify(@channel, @payload)`,
				pgx.StrictNamedArgs{
					"channel": realtime.CollaborationChannel,
					"payload": documentID.String(),
				},
			); err != nil {
				return fmt.Errorf("cannot notify collaboration change: %w", err)
			}

			content, err := materializeAtHeads(document, heads)
			if err != nil {
				return fmt.Errorf("cannot materialize collaboration: %w", err)
			}

			content, err = richtext.DocumentSchema.Sanitize(content)
			if err != nil {
				return fmt.Errorf("cannot sanitize collaboration content: %w", err)
			}

			if err := writeLiveContent(ctx, scope, tx, documentID, content); err != nil {
				return err
			}

			revision = doc.Revision

			return nil
		},
	)
	if err != nil {
		return 0, nil, false, err
	}

	return revision, compactAt, persisted, nil
}

// readStableChanges returns the changes the live document has on top of
// canonicalHeads together with the heads those changes lead to. Peers edit the
// document concurrently, so the read is accepted only when the heads did not
// move around it.
func readStableChanges(
	document *automerge.Document,
	canonicalHeads []automerge.Hash,
) ([]automerge.Hash, []automerge.Change, bool, error) {
	for range stableReadAttempts {
		before, err := document.Heads()
		if err != nil {
			return nil, nil, false, fmt.Errorf("cannot read local collaboration heads: %w", err)
		}

		changes, err := document.ChangesSince(canonicalHeads)
		if err != nil {
			return nil, nil, false, fmt.Errorf("cannot read local collaboration changes: %w", err)
		}

		after, err := document.Heads()
		if err != nil {
			return nil, nil, false, fmt.Errorf("cannot read local collaboration heads: %w", err)
		}

		if slices.Equal(before, after) {
			return after, changes, true, nil
		}
	}

	return nil, nil, false, nil
}

// compact replaces the canonical snapshot with a save of the live document
// and drops the change log it covers. The document is saved before the row is
// locked, so the lock only covers SQL. The save is used only when both the
// live and the canonical document still sit at heads. Otherwise compaction
// waits for a later persist. The changes are already persisted when it runs,
// so a failure is logged and retried later rather than reported.
func (s *Service) compact(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	document *automerge.Document,
	heads []automerge.Hash,
) {
	snapshot, err := document.Save()
	if err != nil {
		s.logCompactionFailure(ctx, documentID, err)

		return
	}

	current, err := document.Heads()
	if err != nil {
		s.logCompactionFailure(ctx, documentID, err)

		return
	}

	if !slices.Equal(current, heads) {
		return
	}

	if len(snapshot) > snapshotMaxBytes {
		s.logCompactionFailure(
			ctx,
			documentID,
			&ErrCollaborationStateTooLarge{Size: len(snapshot)},
		)

		return
	}

	err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			doc, err := lockCollaborationDocument(ctx, tx, scope, documentID)
			if err != nil {
				return err
			}

			canonicalHeads, err := decodeHeads(doc.Heads)
			if err != nil {
				return fmt.Errorf("cannot decode canonical collaboration heads: %w", err)
			}

			// Another writer moved the canonical document, or compacted it.
			if !slices.Equal(canonicalHeads, heads) ||
				doc.ChangeRevision-doc.SnapshotRevision < compactionChanges {
				return nil
			}

			doc.Snapshot = snapshot
			doc.SnapshotRevision = doc.ChangeRevision

			if err := coredata.DeleteCollaborationChangesThroughRevision(
				ctx,
				tx,
				scope,
				documentID,
				doc.SnapshotRevision,
			); err != nil {
				return fmt.Errorf("cannot delete compacted collaboration changes: %w", err)
			}

			doc.UpdatedAt = time.Now()
			if err := doc.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot store compacted collaboration snapshot: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		s.logCompactionFailure(ctx, documentID, err)
	}
}

func (s *Service) logCompactionFailure(
	ctx context.Context,
	documentID gid.GID,
	err error,
) {
	if s.logger == nil {
		return
	}

	s.logger.WarnCtx(
		ctx,
		"cannot compact collaboration document",
		log.Error(err),
		log.String("document_id", documentID.String()),
	)
}

func (s *Service) Refresh(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	document *automerge.Document,
	knownRevision int64,
) (int64, bool, error) {
	doc, changes, err := s.readConsistentCollaboration(ctx, scope, documentID)
	if err != nil {
		return 0, false, err
	}

	if doc.Revision == knownRevision {
		return knownRevision, false, nil
	}

	if err := mergeCanonical(ctx, doc, changes, document); err != nil {
		return 0, false, err
	}

	return doc.Revision, true, nil
}

// catchUp merges the canonical document into a live document. It takes no
// row lock: the canonical state is read consistently and rebuilt outside any
// transaction.
func (s *Service) catchUp(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	document *automerge.Document,
) error {
	doc, changes, err := s.readConsistentCollaboration(ctx, scope, documentID)
	if err != nil {
		return err
	}

	return mergeCanonical(ctx, doc, changes, document)
}

// mergeCanonical rebuilds the canonical document from its snapshot and change
// log and merges it into document.
func mergeCanonical(
	ctx context.Context,
	doc *coredata.CollaborationDocument,
	changes []automerge.Change,
	document *automerge.Document,
) error {
	canonical, err := loadAutomergeDocument(ctx, doc.Snapshot)
	if err != nil {
		return fmt.Errorf("cannot load refreshed collaboration document: %w", err)
	}

	defer func() { _ = canonical.Close() }()

	if len(changes) > 0 {
		if err := canonical.ApplyChanges(changes); err != nil {
			return fmt.Errorf("cannot apply refreshed collaboration changes: %w", err)
		}
	}

	if _, err := document.Merge(canonical); err != nil {
		return fmt.Errorf("cannot merge refreshed collaboration document: %w", err)
	}

	return nil
}

func (s *Service) NotifyEphemeral(
	ctx context.Context,
	documentID gid.GID,
	instanceID string,
	frame []byte,
) error {
	payload, err := realtime.EncodeCollaborationEphemeral(realtime.CollaborationEphemeral{
		DocumentID: documentID.String(),
		InstanceID: instanceID,
		Frame:      frame,
	})
	if err != nil {
		return fmt.Errorf("cannot encode collaboration ephemeral: %w", err)
	}

	return s.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		if _, err := conn.Exec(
			ctx,
			`SELECT pg_notify(@channel, @payload)`,
			pgx.StrictNamedArgs{
				"channel": realtime.CollaborationChannel,
				"payload": payload,
			},
		); err != nil {
			return fmt.Errorf("cannot notify collaboration ephemeral: %w", err)
		}

		return nil
	})
}

func (s *Service) RecordActors(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	organizationID gid.GID,
	identityID gid.GID,
	profileID *gid.GID,
	actors [][]byte,
) error {
	if len(actors) == 0 {
		return nil
	}

	now := time.Now()

	return s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			for _, actorID := range actors {
				if len(actorID) == 0 {
					continue
				}

				row := coredata.CollaborationActor{
					DocumentID:     documentID,
					OrganizationID: organizationID,
					ActorID:        append([]byte(nil), actorID...),
					IdentityID:     identityID,
					ProfileID:      profileID,
					CreatedAt:      now,
				}
				if err := row.Insert(ctx, conn, scope); err != nil {
					return fmt.Errorf("cannot record collaboration actor: %w", err)
				}
			}

			return nil
		},
	)
}

func (s *Service) loadChanges(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	document *automerge.Document,
	snapshotRevision int64,
	latestRevision int64,
) error {
	currentRevision := snapshotRevision

	for currentRevision < latestRevision {
		var batch coredata.CollaborationChanges

		if err := s.pg.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				return batch.LoadAfterRevision(
					ctx,
					conn,
					scope,
					documentID,
					currentRevision,
					changeBatchSize,
				)
			},
		); err != nil {
			return fmt.Errorf("cannot load collaboration change batch: %w", err)
		}

		if len(batch) == 0 {
			return fmt.Errorf(
				"collaboration change log ends at revision %d, expected %d",
				currentRevision,
				latestRevision,
			)
		}

		changes := make([]automerge.Change, len(batch))
		for i, change := range batch {
			if change.Revision != currentRevision+1 {
				return fmt.Errorf(
					"collaboration change revision is %d, expected %d",
					change.Revision,
					currentRevision+1,
				)
			}

			changes[i] = automerge.Change{Bytes: change.ChangeBytes}
			currentRevision = change.Revision
		}

		if err := document.ApplyChanges(changes); err != nil {
			return fmt.Errorf("cannot apply collaboration change batch: %w", err)
		}
	}

	return nil
}

func (s *Service) loadChangesFromConn(
	ctx context.Context,
	scope coredata.Scoper,
	conn pg.Querier,
	documentID gid.GID,
	document *automerge.Document,
	snapshotRevision int64,
	latestRevision int64,
) error {
	changes, err := s.collectChanges(
		ctx,
		conn,
		scope,
		documentID,
		snapshotRevision,
		latestRevision,
	)
	if err != nil {
		return err
	}

	if len(changes) == 0 {
		return nil
	}

	if err := document.ApplyChanges(changes); err != nil {
		return fmt.Errorf("cannot apply collaboration change batch: %w", err)
	}

	return nil
}

func (s *Service) collectChanges(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	documentID gid.GID,
	snapshotRevision int64,
	latestRevision int64,
) ([]automerge.Change, error) {
	currentRevision := snapshotRevision
	changes := make([]automerge.Change, 0)

	for currentRevision < latestRevision {
		var batch coredata.CollaborationChanges

		if err := batch.LoadAfterRevision(
			ctx,
			conn,
			scope,
			documentID,
			currentRevision,
			changeBatchSize,
		); err != nil {
			return nil, fmt.Errorf("cannot load collaboration change batch: %w", err)
		}

		if len(batch) == 0 {
			return nil, fmt.Errorf(
				"collaboration change log ends at revision %d, expected %d",
				currentRevision,
				latestRevision,
			)
		}

		for _, change := range batch {
			if change.Revision != currentRevision+1 {
				return nil, fmt.Errorf(
					"collaboration change revision is %d, expected %d",
					change.Revision,
					currentRevision+1,
				)
			}

			changes = append(changes, automerge.Change{Bytes: change.ChangeBytes})
			currentRevision = change.Revision
		}
	}

	return changes, nil
}

var errCollaborationViewMoved = errors.New("collaboration snapshot moved")

// readConsistentCollaboration reads the snapshot and the change log in one
// transaction and retries once when compaction commits between the two reads.
func (s *Service) readConsistentCollaboration(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
) (*coredata.CollaborationDocument, []automerge.Change, error) {
	const attempts = 2

	var lastErr error

	for range attempts {
		var (
			doc     coredata.CollaborationDocument
			changes []automerge.Change
		)

		err := s.pg.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				if err := doc.LoadByDocumentID(ctx, tx, scope, documentID); err != nil {
					return err
				}

				loaded, err := s.collectChanges(
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

				var again coredata.CollaborationDocument
				if err := again.LoadByDocumentID(ctx, tx, scope, documentID); err != nil {
					return fmt.Errorf("cannot reload collaboration document: %w", err)
				}

				if again.Revision != doc.Revision ||
					again.SnapshotRevision != doc.SnapshotRevision ||
					again.ChangeRevision != doc.ChangeRevision {
					return errCollaborationViewMoved
				}

				changes = loaded

				return nil
			},
		)
		if err == nil {
			return &doc, changes, nil
		}

		if errors.Is(err, errCollaborationViewMoved) {
			lastErr = err

			continue
		}

		return nil, nil, err
	}

	return nil, nil, fmt.Errorf("cannot read consistent collaboration state: %w", lastErr)
}
