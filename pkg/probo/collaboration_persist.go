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
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/realtime"
)

func (s *CollaborationService) Persist(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	document *automerge.Document,
) (int64, error) {
	localHeads, err := document.Heads()
	if err != nil {
		return 0, fmt.Errorf("cannot read local collaboration heads: %w", err)
	}

	localSnapshot, err := document.Save()
	if err != nil {
		return 0, fmt.Errorf("cannot save local collaboration document: %w", err)
	}

	if len(localSnapshot) > snapshotMaxBytes {
		return 0, &ErrCollaborationStateTooLarge{Size: len(localSnapshot)}
	}

	var (
		canonicalChangesForLocal []automerge.Change
		revision                 int64
	)

	err = s.pg.WithTx(
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

			canonical, err := loadAutomergeDocument(ctx, doc.Snapshot)
			if err != nil {
				return fmt.Errorf("cannot load canonical collaboration document: %w", err)
			}

			defer func() { _ = canonical.Close() }()

			if err := s.loadChangesFromConn(
				ctx,
				scope,
				tx,
				documentID,
				canonical,
				doc.SnapshotRevision,
				doc.ChangeRevision,
			); err != nil {
				return fmt.Errorf("cannot load canonical collaboration changes: %w", err)
			}

			local, err := loadAutomergeDocument(ctx, localSnapshot)
			if err != nil {
				return fmt.Errorf("cannot load local collaboration document: %w", err)
			}

			defer func() { _ = local.Close() }()

			before, err := canonical.Heads()
			if err != nil {
				return fmt.Errorf("cannot read canonical collaboration heads: %w", err)
			}

			if _, err := canonical.Merge(local); err != nil {
				return fmt.Errorf("cannot merge collaboration changes: %w", err)
			}

			after, err := canonical.Heads()
			if err != nil {
				return fmt.Errorf("cannot read merged collaboration heads: %w", err)
			}

			incrementalChanges, err := canonical.ChangesSince(before)
			if err != nil {
				return fmt.Errorf("cannot read merged collaboration changes: %w", err)
			}

			localChanges, err := canonical.ChangesSince(localHeads)
			if err != nil {
				return fmt.Errorf("cannot read canonical changes for local document: %w", err)
			}

			canonicalChangesForLocal = localChanges

			seeded := doc.Seeded || len(after) > 0
			if !slices.Equal(before, after) || seeded != doc.Seeded {
				now := time.Now()
				nextChangeRevision := doc.ChangeRevision

				for _, change := range incrementalChanges {
					nextChangeRevision++

					storedChange := coredata.CollaborationChange{
						DocumentID:     documentID,
						OrganizationID: organizationID,
						Revision:       nextChangeRevision,
						ChangeHash:     append([]byte(nil), change.Hash[:]...),
						ChangeBytes:    change.Bytes,
						CreatedAt:      now,
					}
					if err := storedChange.Insert(ctx, tx, scope); err != nil {
						return fmt.Errorf(
							"cannot append collaboration change: %w",
							err,
						)
					}
				}

				doc.Revision++
				doc.ChangeRevision = nextChangeRevision
				doc.Heads = encodeHeads(after)

				if doc.ChangeRevision-doc.SnapshotRevision >= compactionChanges {
					snapshot, err := canonical.Save()
					if err != nil {
						return fmt.Errorf("cannot compact collaboration snapshot: %w", err)
					}

					if len(snapshot) > snapshotMaxBytes {
						return &ErrCollaborationStateTooLarge{Size: len(snapshot)}
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
				}

				doc.Seeded = seeded
				if seeded {
					doc.SeedClaimedAt = nil
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

				if seeded {
					content, err := materialize(ctx, canonical)
					if err != nil {
						return fmt.Errorf("cannot materialize collaboration: %w", err)
					}

					if err := writeLiveContent(ctx, scope, tx, documentID, content); err != nil {
						return err
					}
				}
			}

			revision = doc.Revision

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	if err := document.ApplyChanges(canonicalChangesForLocal); err != nil {
		return 0, fmt.Errorf("cannot refresh local collaboration document: %w", err)
	}

	return revision, nil
}

func (s *CollaborationService) Refresh(
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

	canonical, err := loadAutomergeDocument(ctx, doc.Snapshot)
	if err != nil {
		return 0, false, fmt.Errorf("cannot load refreshed collaboration document: %w", err)
	}

	defer func() { _ = canonical.Close() }()

	if len(changes) > 0 {
		if err := canonical.ApplyChanges(changes); err != nil {
			return 0, false, fmt.Errorf("cannot apply refreshed collaboration changes: %w", err)
		}
	}

	if _, err := document.Merge(canonical); err != nil {
		return 0, false, fmt.Errorf("cannot merge refreshed collaboration document: %w", err)
	}

	return doc.Revision, true, nil
}

func (s *CollaborationService) NotifyEphemeral(
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

func (s *CollaborationService) RecordActors(
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
				if err := row.InsertIfAbsent(ctx, conn, scope); err != nil {
					return fmt.Errorf("cannot record collaboration actor: %w", err)
				}
			}

			return nil
		},
	)
}

func (s *CollaborationService) loadChanges(
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

func (s *CollaborationService) loadChangesFromConn(
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

func (s *CollaborationService) collectChanges(
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
func (s *CollaborationService) readConsistentCollaboration(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
) (*coredata.CollaborationDocument, []automerge.Change, error) {
	const attempts = 2

	var lastErr error

	for attempt := 0; attempt < attempts; attempt++ {
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
