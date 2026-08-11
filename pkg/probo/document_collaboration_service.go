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
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/automerge"
	automergeprosemirror "go.probo.inc/probo/pkg/automerge/prosemirror"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/realtime"
)

const (
	documentCollaborationChangeBatchSize   = 1000
	documentCollaborationCompactionChanges = 500
	documentCollaborationSeedLease         = 30 * time.Second
	documentCollaborationSnapshotMaxBytes  = 8 * 1024 * 1024
)

type (
	DocumentCollaboration struct {
		Document    *automerge.Document
		Revision    int64
		SeedContent string
		NeedsSeed   bool
	}

	ErrDocumentCollaborationStateTooLarge struct {
		Size int
	}

	DocumentCollaborationTextEdit struct {
		ExpectedRevision int64
		Index            uint32
		Cursor           automerge.Cursor
		DeleteCount      int32
		Text             string
	}
)

var (
	ErrDocumentCollaborationNotSeeded = errors.New("document collaboration is not initialized")
	ErrDocumentCollaborationStale     = errors.New("document collaboration changed; read it again before editing")
)

func (e ErrDocumentCollaborationStateTooLarge) Error() string {
	return fmt.Sprintf(
		"document collaboration state is too large: %d bytes exceeds %d bytes",
		e.Size,
		documentCollaborationSnapshotMaxBytes,
	)
}

func (s *DocumentService) OpenCollaboration(
	ctx context.Context,
	scope coredata.Scoper,
	documentVersionID gid.GID,
) (*DocumentCollaboration, error) {
	var (
		snapshot         []byte
		revision         int64
		snapshotRevision int64
		changeRevision   int64
		seedContent      string
		needsSeed        bool
	)

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			version, err := s.loadEditableCollaborationVersion(ctx, tx, scope, documentVersionID)
			if err != nil {
				return err
			}

			seedContent = version.Content

			state := &coredata.DocumentVersionAutomergeState{}

			err = state.LoadByDocumentVersionIDForUpdate(ctx, tx, scope, documentVersionID)
			if err == nil {
				snapshot = append([]byte(nil), state.Snapshot...)

				needsSeed, err = claimCollaborationSeed(ctx, tx, scope, state, time.Now())
				if err != nil {
					return fmt.Errorf("cannot claim document collaboration seed: %w", err)
				}

				revision = state.Revision
				snapshotRevision = state.SnapshotRevision
				changeRevision = state.ChangeRevision

				return nil
			}

			if !errors.Is(err, coredata.ErrResourceNotFound) {
				return fmt.Errorf("cannot load document collaboration state: %w", err)
			}

			document, err := newAutomergeDocument(ctx)
			if err != nil {
				return fmt.Errorf("cannot create empty collaboration document: %w", err)
			}

			defer func() { _ = document.Close(context.Background()) }()

			snapshot, err = document.Save(ctx)
			if err != nil {
				return fmt.Errorf("cannot save empty collaboration document: %w", err)
			}

			now := time.Now()
			state = &coredata.DocumentVersionAutomergeState{
				DocumentVersionID: documentVersionID,
				OrganizationID:    version.OrganizationID,
				Snapshot:          snapshot,
				Heads:             encodeAutomergeHeads(nil),
				Revision:          1,
				SnapshotRevision:  1,
				ChangeRevision:    1,
				Seeded:            false,
				SeedClaimedAt:     new(now),
				CreatedAt:         now,
				UpdatedAt:         now,
			}

			inserted, err := state.InsertIfAbsent(ctx, tx, scope)
			if err != nil {
				return fmt.Errorf("cannot initialize document collaboration state: %w", err)
			}

			if !inserted {
				if err := state.LoadByDocumentVersionIDForUpdate(
					ctx,
					tx,
					scope,
					documentVersionID,
				); err != nil {
					return fmt.Errorf("cannot load concurrently initialized collaboration state: %w", err)
				}

				snapshot = append([]byte(nil), state.Snapshot...)

				needsSeed, err = claimCollaborationSeed(ctx, tx, scope, state, now)
				if err != nil {
					return fmt.Errorf("cannot claim concurrently initialized collaboration seed: %w", err)
				}
			} else {
				needsSeed = true
			}

			revision = state.Revision
			snapshotRevision = state.SnapshotRevision
			changeRevision = state.ChangeRevision

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

	if err := s.loadCollaborationChanges(
		ctx,
		scope,
		documentVersionID,
		document,
		snapshotRevision,
		changeRevision,
	); err != nil {
		_ = document.Close(context.Background())

		return nil, fmt.Errorf("cannot load collaboration changes: %w", err)
	}

	return &DocumentCollaboration{
		Document:    document,
		Revision:    revision,
		SeedContent: seedContent,
		NeedsSeed:   needsSeed,
	}, nil
}

func (s *DocumentService) PersistCollaboration(
	ctx context.Context,
	scope coredata.Scoper,
	documentVersionID gid.GID,
	document *automerge.Document,
) (int64, error) {
	localHeads, err := document.Heads(ctx)
	if err != nil {
		return 0, fmt.Errorf("cannot read local collaboration heads: %w", err)
	}

	localSnapshot, err := document.Save(ctx)
	if err != nil {
		return 0, fmt.Errorf("cannot save local collaboration document: %w", err)
	}

	if len(localSnapshot) > documentCollaborationSnapshotMaxBytes {
		return 0, &ErrDocumentCollaborationStateTooLarge{Size: len(localSnapshot)}
	}

	var (
		canonicalChangesForLocal []automerge.Change
		revision                 int64
	)

	err = s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			version, err := s.loadEditableCollaborationVersion(
				ctx,
				tx,
				scope,
				documentVersionID,
			)
			if err != nil {
				return err
			}

			state := &coredata.DocumentVersionAutomergeState{}
			if err := state.LoadByDocumentVersionIDForUpdate(
				ctx,
				tx,
				scope,
				documentVersionID,
			); err != nil {
				return fmt.Errorf("cannot lock document collaboration state: %w", err)
			}

			canonical, err := loadAutomergeDocument(ctx, state.Snapshot)
			if err != nil {
				return fmt.Errorf("cannot load canonical collaboration document: %w", err)
			}

			defer func() { _ = canonical.Close(context.Background()) }()

			if err := s.loadCollaborationChanges(
				ctx,
				scope,
				documentVersionID,
				canonical,
				state.SnapshotRevision,
				state.ChangeRevision,
			); err != nil {
				return fmt.Errorf("cannot load canonical collaboration changes: %w", err)
			}

			local, err := loadAutomergeDocument(ctx, localSnapshot)
			if err != nil {
				return fmt.Errorf("cannot load local collaboration document: %w", err)
			}

			defer func() { _ = local.Close(context.Background()) }()

			before, err := canonical.Heads(ctx)
			if err != nil {
				return fmt.Errorf("cannot read canonical collaboration heads: %w", err)
			}

			if _, err := canonical.Merge(ctx, local); err != nil {
				return fmt.Errorf("cannot merge collaboration changes: %w", err)
			}

			after, err := canonical.Heads(ctx)
			if err != nil {
				return fmt.Errorf("cannot read merged collaboration heads: %w", err)
			}

			incrementalChanges, err := canonical.ChangesSince(ctx, before)
			if err != nil {
				return fmt.Errorf("cannot read merged collaboration changes: %w", err)
			}

			localChanges, err := canonical.ChangesSince(ctx, localHeads)
			if err != nil {
				return fmt.Errorf("cannot read canonical changes for local document: %w", err)
			}

			canonicalChangesForLocal = localChanges

			seeded := state.Seeded || len(after) > 0
			if !slices.Equal(before, after) || seeded != state.Seeded {
				now := time.Now()
				nextChangeRevision := state.ChangeRevision

				for _, change := range incrementalChanges {
					nextChangeRevision++

					storedChange := coredata.DocumentVersionAutomergeChange{
						DocumentVersionID: documentVersionID,
						OrganizationID:    version.OrganizationID,
						Revision:          nextChangeRevision,
						ChangeHash:        append([]byte(nil), change.Hash[:]...),
						ChangeBytes:       change.Bytes,
						CreatedAt:         now,
					}
					if err := storedChange.Insert(ctx, tx, scope); err != nil {
						return fmt.Errorf(
							"cannot append document collaboration change: %w",
							err,
						)
					}
				}

				state.Revision++
				state.ChangeRevision = nextChangeRevision
				state.Heads = encodeAutomergeHeads(after)

				if state.ChangeRevision-state.SnapshotRevision >=
					documentCollaborationCompactionChanges {
					snapshot, err := canonical.Save(ctx)
					if err != nil {
						return fmt.Errorf("cannot compact collaboration snapshot: %w", err)
					}

					if len(snapshot) > documentCollaborationSnapshotMaxBytes {
						return &ErrDocumentCollaborationStateTooLarge{Size: len(snapshot)}
					}

					state.Snapshot = snapshot
					state.SnapshotRevision = state.ChangeRevision

					if err := coredata.DeleteDocumentVersionAutomergeChangesThroughRevision(
						ctx,
						tx,
						scope,
						documentVersionID,
						state.SnapshotRevision,
					); err != nil {
						return fmt.Errorf("cannot delete compacted collaboration changes: %w", err)
					}
				}

				state.Seeded = seeded
				if seeded {
					state.SeedClaimedAt = nil
				}

				state.UpdatedAt = now
				if err := state.Update(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot persist document collaboration state: %w", err)
				}

				if _, err := tx.Exec(
					ctx,
					`SELECT pg_notify(@channel, @payload)`,
					pgx.StrictNamedArgs{
						"channel": realtime.DocumentCollaborationChannel,
						"payload": documentVersionID.String(),
					},
				); err != nil {
					return fmt.Errorf("cannot notify document collaboration change: %w", err)
				}

				if seeded {
					content, err := materializeCollaboration(ctx, canonical)
					if err != nil {
						return fmt.Errorf("cannot materialize document collaboration: %w", err)
					}

					req := UpdateDocumentRequest{
						DocumentID: version.DocumentID,
						Content:    new(content),
					}
					if err := req.Validate(); err != nil {
						return fmt.Errorf("cannot validate materialized collaboration: %w", err)
					}

					version.Content = content

					version.UpdatedAt = now
					if err := version.Update(ctx, tx, scope); err != nil {
						return fmt.Errorf("cannot update materialized document version: %w", err)
					}
				}
			}

			revision = state.Revision

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	if err := document.ApplyChanges(ctx, canonicalChangesForLocal); err != nil {
		return 0, fmt.Errorf("cannot refresh local collaboration document: %w", err)
	}

	return revision, nil
}

func (s *DocumentService) ReleaseCollaborationSeed(
	ctx context.Context,
	scope coredata.Scoper,
	documentVersionID gid.GID,
) error {
	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			state := &coredata.DocumentVersionAutomergeState{}
			if err := state.LoadByDocumentVersionIDForUpdate(
				ctx,
				tx,
				scope,
				documentVersionID,
			); err != nil {
				return fmt.Errorf("cannot lock document collaboration seed: %w", err)
			}

			if state.Seeded || state.SeedClaimedAt == nil {
				return nil
			}

			state.SeedClaimedAt = nil
			state.Revision++

			state.UpdatedAt = time.Now()
			if err := state.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot release document collaboration seed: %w", err)
			}

			return nil
		},
	)
}

func (s *DocumentService) RefreshCollaboration(
	ctx context.Context,
	scope coredata.Scoper,
	documentVersionID gid.GID,
	document *automerge.Document,
	knownRevision int64,
) (int64, bool, error) {
	state := &coredata.DocumentVersionAutomergeState{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := state.LoadByDocumentVersionID(
				ctx,
				conn,
				scope,
				documentVersionID,
			); err != nil {
				return fmt.Errorf("cannot load document collaboration state: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, false, err
	}

	if state.Revision == knownRevision {
		return knownRevision, false, nil
	}

	canonical, err := loadAutomergeDocument(ctx, state.Snapshot)
	if err != nil {
		return 0, false, fmt.Errorf("cannot load refreshed collaboration document: %w", err)
	}

	defer func() { _ = canonical.Close(context.Background()) }()

	if err := s.loadCollaborationChanges(
		ctx,
		scope,
		documentVersionID,
		canonical,
		state.SnapshotRevision,
		state.ChangeRevision,
	); err != nil {
		return 0, false, fmt.Errorf("cannot load refreshed collaboration changes: %w", err)
	}

	if _, err := document.Merge(ctx, canonical); err != nil {
		return 0, false, fmt.Errorf("cannot merge refreshed collaboration document: %w", err)
	}

	return state.Revision, true, nil
}

func (s *DocumentService) ApplyCollaborationTextEdit(
	ctx context.Context,
	scope coredata.Scoper,
	documentVersionID gid.GID,
	edit DocumentCollaborationTextEdit,
) (int64, error) {
	if edit.DeleteCount < 0 {
		return 0, fmt.Errorf("collaboration text delete count cannot be negative")
	}

	collaboration, err := s.OpenCollaboration(ctx, scope, documentVersionID)
	if err != nil {
		return 0, fmt.Errorf("cannot open collaboration for text edit: %w", err)
	}

	defer func() { _ = collaboration.Document.Close(context.Background()) }()

	if collaboration.NeedsSeed {
		return 0, ErrDocumentCollaborationNotSeeded
	}

	if edit.ExpectedRevision != collaboration.Revision {
		return 0, ErrDocumentCollaborationStale
	}

	text, err := collaboration.Document.Text(ctx, "body")
	if err != nil {
		return 0, fmt.Errorf("cannot get collaboration body: %w", err)
	}

	index := edit.Index
	if len(edit.Cursor) > 0 {
		index, err = text.CursorPosition(ctx, edit.Cursor)
		if err != nil {
			return 0, fmt.Errorf("cannot resolve collaboration text cursor: %w", err)
		}
	}

	if err := text.Splice(ctx, index, edit.DeleteCount, edit.Text); err != nil {
		return 0, fmt.Errorf("cannot apply collaboration text edit: %w", err)
	}

	content, err := text.String(ctx)
	if err != nil {
		return 0, fmt.Errorf("cannot read edited collaboration body: %w", err)
	}

	if utf8.RuneCountInString(content) > documentContentMaxTextLength {
		return 0, fmt.Errorf(
			"collaboration text exceeds maximum length of %d characters",
			documentContentMaxTextLength,
		)
	}

	if _, err := collaboration.Document.Commit(ctx, "Agent edit", time.Now()); err != nil {
		return 0, fmt.Errorf("cannot commit collaboration text edit: %w", err)
	}

	revision, err := s.PersistCollaboration(
		ctx,
		scope,
		documentVersionID,
		collaboration.Document,
	)
	if err != nil {
		return 0, fmt.Errorf("cannot persist collaboration text edit: %w", err)
	}

	return revision, nil
}

func (s *DocumentService) ReadCollaborationText(
	ctx context.Context,
	scope coredata.Scoper,
	documentVersionID gid.GID,
) (string, int64, error) {
	collaboration, err := s.OpenCollaboration(ctx, scope, documentVersionID)
	if err != nil {
		return "", 0, fmt.Errorf("cannot open collaboration for reading: %w", err)
	}

	defer func() { _ = collaboration.Document.Close(context.Background()) }()

	if collaboration.NeedsSeed {
		return "", 0, ErrDocumentCollaborationNotSeeded
	}

	text, err := collaboration.Document.Text(ctx, "body")
	if err != nil {
		return "", 0, fmt.Errorf("cannot get collaboration body for reading: %w", err)
	}

	content, err := text.String(ctx)
	if err != nil {
		return "", 0, fmt.Errorf("cannot read collaboration body: %w", err)
	}

	return content, collaboration.Revision, nil
}

func (s *DocumentService) loadEditableCollaborationVersion(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	documentVersionID gid.GID,
) (*coredata.DocumentVersion, error) {
	version := &coredata.DocumentVersion{}
	if err := version.LoadByID(ctx, conn, scope, documentVersionID); err != nil {
		return nil, fmt.Errorf("cannot load collaboration document version: %w", err)
	}

	if version.Status != coredata.DocumentVersionStatusDraft {
		return nil, &ErrDocumentVersionNotDraft{}
	}

	document := &coredata.Document{}
	if err := document.LoadByID(ctx, conn, scope, version.DocumentID); err != nil {
		return nil, fmt.Errorf("cannot load collaboration document: %w", err)
	}

	if document.ArchivedAt != nil {
		return nil, &ErrDocumentArchived{}
	}

	if document.WriteMode == coredata.DocumentWriteModeGenerated {
		return nil, &ErrDocumentVersionGenerated{}
	}

	return version, nil
}

func newAutomergeDocument(ctx context.Context) (*automerge.Document, error) {
	actorID, err := automerge.NewActorID()
	if err != nil {
		return nil, fmt.Errorf("cannot generate collaboration actor ID: %w", err)
	}

	document, err := automerge.New(ctx, actorID)
	if err != nil {
		return nil, fmt.Errorf("cannot create Automerge document: %w", err)
	}

	return document, nil
}

func loadAutomergeDocument(ctx context.Context, snapshot []byte) (*automerge.Document, error) {
	actorID, err := automerge.NewActorID()
	if err != nil {
		return nil, fmt.Errorf("cannot generate collaboration actor ID: %w", err)
	}

	document, err := automerge.Load(ctx, snapshot, actorID)
	if err != nil {
		return nil, fmt.Errorf("cannot load Automerge document: %w", err)
	}

	return document, nil
}

func (s *DocumentService) loadCollaborationChanges(
	ctx context.Context,
	scope coredata.Scoper,
	documentVersionID gid.GID,
	document *automerge.Document,
	snapshotRevision int64,
	latestRevision int64,
) error {
	currentRevision := snapshotRevision

	for currentRevision < latestRevision {
		var batch coredata.DocumentVersionAutomergeChanges

		if err := s.svc.pg.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				return batch.LoadAfterRevision(
					ctx,
					conn,
					scope,
					documentVersionID,
					currentRevision,
					documentCollaborationChangeBatchSize,
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

		if err := document.ApplyChanges(ctx, changes); err != nil {
			return fmt.Errorf("cannot apply collaboration change batch: %w", err)
		}
	}

	return nil
}

func encodeAutomergeHeads(heads []automerge.Hash) []byte {
	encoded := make([]byte, 0, len(heads)*32)
	for _, head := range heads {
		encoded = append(encoded, head[:]...)
	}

	return encoded
}

func claimCollaborationSeed(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	state *coredata.DocumentVersionAutomergeState,
	now time.Time,
) (bool, error) {
	if state.Seeded {
		return false, nil
	}

	if state.SeedClaimedAt != nil && state.SeedClaimedAt.After(now.Add(-documentCollaborationSeedLease)) {
		return false, nil
	}

	state.SeedClaimedAt = new(now)
	state.Revision++

	state.UpdatedAt = now
	if err := state.Update(ctx, tx, scope); err != nil {
		return false, fmt.Errorf("cannot update document collaboration seed claim: %w", err)
	}

	return true, nil
}

func materializeCollaboration(
	ctx context.Context,
	document *automerge.Document,
) (string, error) {
	text, err := document.Text(ctx, "body")
	if err != nil {
		return "", fmt.Errorf("cannot get collaboration body: %w", err)
	}

	spans, err := text.Spans(ctx)
	if err != nil {
		return "", fmt.Errorf("cannot get collaboration spans: %w", err)
	}

	content, err := automergeprosemirror.Render(spans)
	if err != nil {
		return "", fmt.Errorf("cannot render collaboration spans: %w", err)
	}

	content, err = prosemirror.SanitizeDocumentJSON(content)
	if err != nil {
		return "", fmt.Errorf("cannot sanitize collaboration content: %w", err)
	}

	return content, nil
}
