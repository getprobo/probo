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
	"encoding/binary"
	"errors"
	"fmt"
	"slices"
	"time"
	"unicode/utf16"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/automerge"
	automergeprosemirror "go.probo.inc/probo/pkg/automerge/prosemirror"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

const revisionGap = 10 * time.Minute

type (
	DocumentHistoryHighlightKind string

	DocumentRevision struct {
		ID            gid.GID
		DocumentID    gid.GID
		Heads         []automerge.Hash
		PreviousHeads []automerge.Hash
		ActorID       []byte
		IdentityID    *gid.GID
		ProfileID     *gid.GID
		AuthorName    string
		StartedAt     time.Time
		EndedAt       time.Time
		ChangeCount   int
	}

	DocumentHistoryHighlight struct {
		Kind DocumentHistoryHighlightKind
		From int
		To   int
		Text string
	}

	DocumentHistorySnapshot struct {
		Content    string
		Highlights []DocumentHistoryHighlight
	}

	DocumentRevisionOrderField string

	actorAttribution struct {
		IdentityID *gid.GID
		ProfileID  *gid.GID
		AuthorName string
	}
)

const (
	DocumentHistoryHighlightKindInsert DocumentHistoryHighlightKind = "INSERT"
	DocumentHistoryHighlightKindDelete DocumentHistoryHighlightKind = "DELETE"

	DocumentRevisionOrderFieldEndedAt DocumentRevisionOrderField = "ENDED_AT"
)

func (DocumentRevisionOrderField) Column() string {
	return "ended_at"
}

func (f DocumentRevisionOrderField) String() string {
	return string(f)
}

func (f DocumentRevisionOrderField) IsValid() bool {
	return f == DocumentRevisionOrderFieldEndedAt
}

func (s *DocumentRevision) CursorKey(orderBy DocumentRevisionOrderField) page.CursorKey {
	switch orderBy {
	case DocumentRevisionOrderFieldEndedAt:
		return page.NewCursorKey(s.ID, s.EndedAt)
	}

	panic(fmt.Sprintf("unsupported order field: %q", orderBy))
}

func (s *Service) ListRevisions(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	cursor *page.Cursor[DocumentRevisionOrderField],
) (*page.Page[*DocumentRevision, DocumentRevisionOrderField], int, error) {
	sessions, err := s.loadRevisions(ctx, scope, documentID)
	if err != nil {
		return nil, 0, err
	}

	paged := paginateRevisions(sessions, cursor)

	return page.NewPage(paged, cursor), len(sessions), nil
}

func (s *Service) GetRevision(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	sessionID gid.GID,
) (*DocumentRevision, error) {
	sessions, err := s.loadRevisions(ctx, scope, documentID)
	if err != nil {
		return nil, err
	}

	for _, session := range sessions {
		if session.ID == sessionID {
			return session, nil
		}
	}

	return nil, coredata.ErrResourceNotFound
}

func (s *Service) loadRevisions(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
) ([]*DocumentRevision, error) {
	opened, err := s.openIfSeeded(ctx, scope, documentID)
	if err != nil {
		if errors.Is(err, ErrDocumentCollaborationNotSeeded) || errors.Is(err, ErrSeedInProgress) {
			return nil, nil
		}

		return nil, fmt.Errorf("cannot open collaboration for history: %w", err)
	}

	defer func() { _ = opened.Document.Close() }()

	changes, err := opened.Document.ChangeHistory()
	if err != nil {
		return nil, fmt.Errorf("cannot read collaboration change history: %w", err)
	}

	actors, err := s.loadActorAttributions(ctx, scope, documentID, actorIDsFromChanges(changes))
	if err != nil {
		return nil, err
	}

	sessions := groupRevisions(scope.GetTenantID(), changes, actors)
	for _, session := range sessions {
		session.DocumentID = documentID
	}

	return sessions, nil
}

func (s *Service) GetHistorySnapshot(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	heads []automerge.Hash,
	previousHeads []automerge.Hash,
) (*DocumentHistorySnapshot, error) {
	if len(heads) == 0 {
		return nil, fmt.Errorf("collaboration history snapshot requires heads")
	}

	opened, err := s.openIfSeeded(ctx, scope, documentID)
	if err != nil {
		if errors.Is(err, ErrDocumentCollaborationNotSeeded) || errors.Is(err, ErrSeedInProgress) {
			return nil, err
		}

		return nil, fmt.Errorf("cannot open collaboration for history snapshot: %w", err)
	}

	defer func() { _ = opened.Document.Close() }()

	hasHeads, err := opened.Document.HasHeads(heads)
	if err != nil {
		return nil, fmt.Errorf("cannot verify collaboration history heads: %w", err)
	}

	if !hasHeads {
		return nil, fmt.Errorf("collaboration history heads are unknown")
	}

	content, err := materializeAtHeads(opened.Document, heads)
	if err != nil {
		return nil, err
	}

	highlights, err := historyHighlights(opened.Document, previousHeads, heads)
	if err != nil {
		return nil, err
	}

	return &DocumentHistorySnapshot{
		Content:    content,
		Highlights: highlights,
	}, nil
}

func (s *Service) RestoreHistory(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	heads []automerge.Hash,
) (int64, error) {
	if len(heads) == 0 {
		return 0, fmt.Errorf("collaboration history restore requires heads")
	}

	var (
		organizationID gid.GID
		identityID     gid.GID
		profileID      *gid.GID
	)

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			orgID, err := assertDraftEditable(
				ctx,
				scope,
				conn,
				documentID,
			)
			if err != nil {
				return err
			}

			organizationID = orgID

			identityID, profileID, err = collaborationProfile(
				ctx,
				conn,
				scope,
				organizationID,
			)
			if err != nil {
				return err
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	// A fresh actor keeps the restore off the editor actor's clock and session.
	actorID, err := automerge.NewActorID()
	if err != nil {
		return 0, fmt.Errorf("cannot create collaboration restore actor: %w", err)
	}

	opened, err := s.openWithActor(ctx, scope, documentID, actorID)
	if err != nil {
		if errors.Is(err, ErrDocumentCollaborationNotSeeded) || errors.Is(err, ErrSeedInProgress) {
			return 0, err
		}

		return 0, fmt.Errorf("cannot open collaboration for history restore: %w", err)
	}

	defer func() { _ = opened.Document.Close() }()

	spans, err := BodySpans(opened.Document)
	if err != nil {
		return 0, err
	}

	if HasOpenSuggestions(spans) {
		return 0, ErrOpenSuggestions
	}

	hasHeads, err := opened.Document.HasHeads(heads)
	if err != nil {
		return 0, fmt.Errorf("cannot verify collaboration history restore heads: %w", err)
	}

	if !hasHeads {
		return 0, fmt.Errorf("collaboration history restore heads are unknown")
	}

	// Record the restore actor only after validation so a rejected restore
	// leaves no actor attribution behind.
	if err := s.RecordActors(
		ctx,
		scope,
		documentID,
		organizationID,
		identityID,
		profileID,
		[][]byte{actorID[:]},
	); err != nil {
		return 0, fmt.Errorf("cannot record collaboration restore actor: %w", err)
	}

	content, err := materializeAtHeads(opened.Document, heads)
	if err != nil {
		return 0, err
	}

	restored, err := automergeprosemirror.ToSpans(content)
	if err != nil {
		return 0, fmt.Errorf("cannot convert restored content to spans: %w", err)
	}

	text, err := opened.Document.Text("body")
	if err != nil {
		return 0, fmt.Errorf("cannot get collaboration body for restore: %w", err)
	}

	if err := text.UpdateSpans(restored, automergeprosemirror.UpdateSpansConfig()); err != nil {
		return 0, fmt.Errorf("cannot write restored collaboration spans: %w", err)
	}

	if _, err := opened.Document.Commit("Restore document history", time.Now()); err != nil {
		return 0, fmt.Errorf("cannot commit collaboration history restore: %w", err)
	}

	revision, err := s.Persist(ctx, scope, documentID, opened.Document)
	if err != nil {
		return 0, fmt.Errorf("cannot persist collaboration history restore: %w", err)
	}

	return revision, nil
}

func actorIDsFromChanges(changes []automerge.ChangeInfo) [][]byte {
	seen := make(map[string]struct{}, len(changes))
	ids := make([][]byte, 0)

	for _, change := range changes {
		if len(change.Actor) == 0 {
			continue
		}

		key := string(change.Actor)
		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}

		ids = append(ids, append([]byte(nil), change.Actor...))
	}

	return ids
}

func (s *Service) loadActorAttributions(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	actorIDs [][]byte,
) (map[string]actorAttribution, error) {
	var actors coredata.CollaborationActors

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return actors.LoadByDocumentIDAndActorIDs(
				ctx,
				conn,
				scope,
				documentID,
				actorIDs,
			)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load collaboration actors: %w", err)
	}

	result := make(map[string]actorAttribution, len(actors))
	for _, actor := range actors {
		attr := actorAttribution{
			IdentityID: new(actor.IdentityID),
			ProfileID:  actor.ProfileID,
		}

		if actor.ProfileID != nil {
			profile := &coredata.MembershipProfile{}

			err := s.pg.WithConn(
				ctx,
				func(ctx context.Context, conn pg.Querier) error {
					return profile.LoadByID(ctx, conn, scope, *actor.ProfileID)
				},
			)
			if err == nil && profile.FullName != "" {
				attr.AuthorName = profile.FullName
				result[string(actor.ActorID)] = attr

				continue
			}
		}

		identity := &coredata.Identity{}

		err := s.pg.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				return identity.LoadByID(ctx, conn, actor.IdentityID)
			},
		)
		if err == nil {
			attr.AuthorName = identity.FullName
			if attr.AuthorName == "" {
				attr.AuthorName = identity.EmailAddress.String()
			}
		}

		result[string(actor.ActorID)] = attr
	}

	return result, nil
}

func groupRevisions(
	tenantID gid.TenantID,
	changes []automerge.ChangeInfo,
	actors map[string]actorAttribution,
) []*DocumentRevision {
	if len(changes) == 0 {
		return nil
	}

	sessions := make([]*DocumentRevision, 0)

	var current *DocumentRevision

	for _, change := range changes {
		actorKey := string(change.Actor)

		changeTime := change.Time
		if changeTime.IsZero() {
			changeTime = time.Unix(0, 0).UTC()
		}

		startsNew := current == nil ||
			string(current.ActorID) != actorKey ||
			(!changeTime.Equal(time.Unix(0, 0).UTC()) &&
				!current.EndedAt.Equal(time.Unix(0, 0).UTC()) &&
				changeTime.Sub(current.EndedAt) > revisionGap)

		if startsNew {
			if current != nil {
				sessions = append(sessions, current)
			}

			attr := actors[actorKey]
			current = &DocumentRevision{
				ID:          revisionID(tenantID, change.Hash),
				Heads:       []automerge.Hash{change.Hash},
				ActorID:     append([]byte(nil), change.Actor...),
				IdentityID:  attr.IdentityID,
				ProfileID:   attr.ProfileID,
				AuthorName:  attr.AuthorName,
				StartedAt:   changeTime,
				EndedAt:     changeTime,
				ChangeCount: 1,
			}

			continue
		}

		current.ID = revisionID(tenantID, change.Hash)
		current.Heads = []automerge.Hash{change.Hash}
		current.EndedAt = changeTime
		current.ChangeCount++
	}

	if current != nil {
		sessions = append(sessions, current)
	}

	var previousHeads []automerge.Hash
	for _, session := range sessions {
		session.PreviousHeads = append([]automerge.Hash(nil), previousHeads...)
		previousHeads = append([]automerge.Hash(nil), session.Heads...)
	}

	slices.Reverse(sessions)

	return sessions
}

func revisionID(tenantID gid.TenantID, changeHash automerge.Hash) gid.GID {
	var id gid.GID
	copy(id[0:8], tenantID[:])
	binary.BigEndian.PutUint16(id[8:10], coredata.DocumentCollaborationSessionEntityType)
	copy(id[10:24], changeHash[0:14])

	return id
}

func paginateRevisions(
	sessions []*DocumentRevision,
	cursor *page.Cursor[DocumentRevisionOrderField],
) []*DocumentRevision {
	if cursor == nil {
		return sessions
	}

	ordered := sessions
	if cursor.OrderBy.Direction == page.OrderDirectionAsc {
		ordered = slices.Clone(sessions)
		slices.Reverse(ordered)
	}

	if cursor.Position == page.Tail {
		return paginateRevisionTail(ordered, cursor)
	}

	return paginateRevisionHead(ordered, cursor)
}

func paginateRevisionHead(
	ordered []*DocumentRevision,
	cursor *page.Cursor[DocumentRevisionOrderField],
) []*DocumentRevision {
	start := 0

	if cursor.Key != nil {
		for i, session := range ordered {
			if session.ID == cursor.Key.ID {
				start = i
				break
			}
		}
	}

	limit := cursor.Size + 1
	if cursor.Key != nil {
		limit++
	}

	end := min(start+limit, len(ordered))

	return ordered[start:end]
}

func paginateRevisionTail(
	ordered []*DocumentRevision,
	cursor *page.Cursor[DocumentRevisionOrderField],
) []*DocumentRevision {
	end := len(ordered)

	if cursor.Key != nil {
		found := false

		for i, session := range ordered {
			if session.ID == cursor.Key.ID {
				end = i + 1
				found = true

				break
			}
		}

		if !found {
			return nil
		}
	}

	limit := cursor.Size + 1
	if cursor.Key != nil {
		limit++
	}

	start := max(end-limit, 0)

	window := slices.Clone(ordered[start:end])
	slices.Reverse(window)

	return window
}

func historyHighlights(
	document *automerge.Document,
	before []automerge.Hash,
	after []automerge.Hash,
) ([]DocumentHistoryHighlight, error) {
	patches, err := document.Diff(before, after)
	if err != nil {
		return nil, fmt.Errorf("cannot diff revision: %w", err)
	}

	text, err := document.Text("body")
	if err != nil {
		return nil, fmt.Errorf("cannot get collaboration body for highlights: %w", err)
	}

	afterSpans, err := text.SpansAt(after)
	if err != nil {
		return nil, fmt.Errorf("cannot read after spans for highlights: %w", err)
	}

	var beforeSpans []automerge.Span
	if len(before) > 0 {
		beforeSpans, err = text.SpansAt(before)
		if err != nil {
			return nil, fmt.Errorf("cannot read before spans for highlights: %w", err)
		}
	}

	highlights := make([]DocumentHistoryHighlight, 0)

	for _, patch := range patches {
		switch patch.Action {
		case automerge.PatchSpliceText:
			if patch.Text == "" {
				continue
			}

			from, ok := automergeIndexToPlainOffset(afterSpans, patch.Index)
			if !ok {
				continue
			}

			length := utf16CodeUnitCount(patch.Text)
			highlights = append(
				highlights,
				DocumentHistoryHighlight{
					Kind: DocumentHistoryHighlightKindInsert,
					From: from,
					To:   from + length,
					Text: patch.Text,
				},
			)

		case automerge.PatchDeleteSeq:
			if patch.Length == 0 {
				continue
			}

			from, ok := automergeIndexToPlainOffset(afterSpans, patch.Index)
			if !ok {
				continue
			}

			deleted := plainTextSlice(beforeSpans, patch.Index, patch.Length)
			highlights = append(
				highlights,
				DocumentHistoryHighlight{
					Kind: DocumentHistoryHighlightKindDelete,
					From: from,
					To:   from,
					Text: deleted,
				},
			)
		}
	}

	return highlights, nil
}

func automergeIndexToPlainOffset(spans []automerge.Span, index uint64) (int, bool) {
	var am uint64

	plain := 0

	for _, span := range spans {
		if span.Type == automerge.SpanTypeBlock || span.Block != nil {
			if am == index {
				return plain, true
			}

			am++
			plain++

			continue
		}

		units := utf16.Encode([]rune(span.Text))
		for range units {
			if am == index {
				return plain, true
			}

			am++
			plain++
		}
	}

	if am == index {
		return plain, true
	}

	return 0, false
}

func plainTextSlice(spans []automerge.Span, index uint64, length uint64) string {
	if length == 0 || len(spans) == 0 {
		return ""
	}

	var am uint64

	collected := make([]uint16, 0, length)

	for _, span := range spans {
		if span.Type == automerge.SpanTypeBlock || span.Block != nil {
			am++

			continue
		}

		units := utf16.Encode([]rune(span.Text))
		for _, unit := range units {
			if am < index {
				am++

				continue
			}

			if am >= index+length {
				return string(utf16.Decode(collected))
			}

			collected = append(collected, unit)
			am++
		}
	}

	return string(utf16.Decode(collected))
}

func utf16CodeUnitCount(value string) int {
	return len(utf16.Encode([]rune(value)))
}
