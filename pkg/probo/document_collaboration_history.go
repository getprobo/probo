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
	"encoding/binary"
	"encoding/hex"
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

const documentCollaborationHistorySessionGap = 10 * time.Minute

type (
	DocumentCollaborationHistoryHighlightKind string

	DocumentCollaborationHistorySession struct {
		ID            gid.GID
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

	DocumentCollaborationHistoryHighlight struct {
		Kind DocumentCollaborationHistoryHighlightKind
		From int
		To   int
		Text string
	}

	DocumentCollaborationHistorySnapshot struct {
		Content    string
		Highlights []DocumentCollaborationHistoryHighlight
	}

	DocumentCollaborationHistorySessionOrderField string

	collaborationActorAttribution struct {
		IdentityID *gid.GID
		ProfileID  *gid.GID
		AuthorName string
	}
)

const (
	DocumentCollaborationHistoryHighlightKindInsert DocumentCollaborationHistoryHighlightKind = "INSERT"
	DocumentCollaborationHistoryHighlightKindDelete DocumentCollaborationHistoryHighlightKind = "DELETE"

	DocumentCollaborationHistorySessionOrderFieldEndedAt DocumentCollaborationHistorySessionOrderField = "ENDED_AT"
)

func (DocumentCollaborationHistorySessionOrderField) Column() string {
	return "ended_at"
}

func (f DocumentCollaborationHistorySessionOrderField) String() string {
	return string(f)
}

func (f DocumentCollaborationHistorySessionOrderField) IsValid() bool {
	return f == DocumentCollaborationHistorySessionOrderFieldEndedAt
}

func (s *DocumentCollaborationHistorySession) CursorKey(
	orderBy DocumentCollaborationHistorySessionOrderField,
) page.CursorKey {
	switch orderBy {
	case DocumentCollaborationHistorySessionOrderFieldEndedAt:
		return page.NewCursorKey(s.ID, s.EndedAt)
	}

	panic(fmt.Sprintf("unsupported order field: %q", orderBy))
}

// ListCollaborationHistorySessions groups Automerge changes into author sessions
// (same actor, gaps under 10 minutes) and returns them newest-first.
func (s *DocumentService) ListCollaborationHistorySessions(
	ctx context.Context,
	scope coredata.Scoper,
	documentVersionID gid.GID,
	cursor *page.Cursor[DocumentCollaborationHistorySessionOrderField],
) (*page.Page[*DocumentCollaborationHistorySession, DocumentCollaborationHistorySessionOrderField], int, error) {
	collaboration, err := s.OpenCollaboration(ctx, scope, documentVersionID)
	if err != nil {
		return nil, 0, fmt.Errorf("cannot open collaboration for history: %w", err)
	}

	defer func() { _ = collaboration.Document.Close() }()

	if collaboration.NeedsSeed {
		return page.NewPage([]*DocumentCollaborationHistorySession{}, cursor), 0, nil
	}

	changes, err := collaboration.Document.ChangeHistory()
	if err != nil {
		return nil, 0, fmt.Errorf("cannot read collaboration change history: %w", err)
	}

	actors, err := s.loadCollaborationActorAttributions(ctx, scope, documentVersionID)
	if err != nil {
		return nil, 0, err
	}

	sessions := groupCollaborationHistorySessions(
		scope.GetTenantID(),
		changes,
		actors,
	)
	paged := paginateCollaborationHistorySessions(sessions, cursor)

	return page.NewPage(paged, cursor), len(sessions), nil
}

// GetCollaborationHistorySnapshot materializes the document at the given heads
// and returns highlight ranges for the edits since previousHeads.
func (s *DocumentService) GetCollaborationHistorySnapshot(
	ctx context.Context,
	scope coredata.Scoper,
	documentVersionID gid.GID,
	heads []automerge.Hash,
	previousHeads []automerge.Hash,
) (*DocumentCollaborationHistorySnapshot, error) {
	if len(heads) == 0 {
		return nil, fmt.Errorf("collaboration history snapshot requires heads")
	}

	collaboration, err := s.OpenCollaboration(ctx, scope, documentVersionID)
	if err != nil {
		return nil, fmt.Errorf("cannot open collaboration for history snapshot: %w", err)
	}

	defer func() { _ = collaboration.Document.Close() }()

	if collaboration.NeedsSeed {
		return nil, ErrDocumentCollaborationNotSeeded
	}

	hasHeads, err := collaboration.Document.HasHeads(heads)
	if err != nil {
		return nil, fmt.Errorf("cannot verify collaboration history heads: %w", err)
	}

	if !hasHeads {
		return nil, fmt.Errorf("collaboration history heads are unknown")
	}

	content, err := materializeCollaborationAtHeads(collaboration.Document, heads)
	if err != nil {
		return nil, err
	}

	highlights, err := collaborationHistoryHighlights(
		collaboration.Document,
		previousHeads,
		heads,
	)
	if err != nil {
		return nil, err
	}

	return &DocumentCollaborationHistorySnapshot{
		Content:    content,
		Highlights: highlights,
	}, nil
}

// RestoreCollaborationHistory replaces the live body with the document state at
// heads by appending one new Automerge change.
func (s *DocumentService) RestoreCollaborationHistory(
	ctx context.Context,
	scope coredata.Scoper,
	documentVersionID gid.GID,
	heads []automerge.Hash,
) (int64, error) {
	if len(heads) == 0 {
		return 0, fmt.Errorf("collaboration history restore requires heads")
	}

	collaboration, err := s.OpenCollaboration(ctx, scope, documentVersionID)
	if err != nil {
		return 0, fmt.Errorf("cannot open collaboration for history restore: %w", err)
	}

	defer func() { _ = collaboration.Document.Close() }()

	if collaboration.NeedsSeed {
		return 0, ErrDocumentCollaborationNotSeeded
	}

	hasHeads, err := collaboration.Document.HasHeads(heads)
	if err != nil {
		return 0, fmt.Errorf("cannot verify collaboration history restore heads: %w", err)
	}

	if !hasHeads {
		return 0, fmt.Errorf("collaboration history restore heads are unknown")
	}

	content, err := materializeCollaborationAtHeads(collaboration.Document, heads)
	if err != nil {
		return 0, err
	}

	spans, err := automergeprosemirror.ToSpans(content)
	if err != nil {
		return 0, fmt.Errorf("cannot convert restored content to spans: %w", err)
	}

	text, err := collaboration.Document.Text("body")
	if err != nil {
		return 0, fmt.Errorf("cannot get collaboration body for restore: %w", err)
	}

	if err := text.UpdateSpans(spans, automergeprosemirror.UpdateSpansConfig()); err != nil {
		return 0, fmt.Errorf("cannot write restored collaboration spans: %w", err)
	}

	if _, err := collaboration.Document.Commit("Restore document history", time.Now()); err != nil {
		return 0, fmt.Errorf("cannot commit collaboration history restore: %w", err)
	}

	revision, err := s.PersistCollaboration(
		ctx,
		scope,
		documentVersionID,
		collaboration.Document,
	)
	if err != nil {
		return 0, fmt.Errorf("cannot persist collaboration history restore: %w", err)
	}

	return revision, nil
}

func (s *DocumentService) loadCollaborationActorAttributions(
	ctx context.Context,
	scope coredata.Scoper,
	documentVersionID gid.GID,
) (map[string]collaborationActorAttribution, error) {
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
		return nil, fmt.Errorf("cannot load collaboration actors: %w", err)
	}

	result := make(map[string]collaborationActorAttribution, len(actors))
	for _, actor := range actors {
		attr := collaborationActorAttribution{
			IdentityID: new(actor.IdentityID),
			ProfileID:  actor.ProfileID,
		}

		if actor.ProfileID != nil {
			profile := &coredata.MembershipProfile{}
			err := s.svc.pg.WithConn(
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
		err := s.svc.pg.WithConn(
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

func groupCollaborationHistorySessions(
	tenantID gid.TenantID,
	changes []automerge.ChangeInfo,
	actors map[string]collaborationActorAttribution,
) []*DocumentCollaborationHistorySession {
	if len(changes) == 0 {
		return nil
	}

	sessions := make([]*DocumentCollaborationHistorySession, 0)
	var current *DocumentCollaborationHistorySession

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
				changeTime.Sub(current.EndedAt) > documentCollaborationHistorySessionGap)

		if startsNew {
			if current != nil {
				sessions = append(sessions, current)
			}

			attr := actors[actorKey]
			current = &DocumentCollaborationHistorySession{
				ID:          collaborationHistorySessionID(tenantID, change.Hash),
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

		current.ID = collaborationHistorySessionID(tenantID, change.Hash)
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

func collaborationHistorySessionID(
	tenantID gid.TenantID,
	changeHash automerge.Hash,
) gid.GID {
	var id gid.GID
	copy(id[0:8], tenantID[:])
	binary.BigEndian.PutUint16(id[8:10], coredata.DocumentCollaborationSessionEntityType)
	copy(id[10:24], changeHash[0:14])

	return id
}

func paginateCollaborationHistorySessions(
	sessions []*DocumentCollaborationHistorySession,
	cursor *page.Cursor[DocumentCollaborationHistorySessionOrderField],
) []*DocumentCollaborationHistorySession {
	if cursor == nil {
		return sessions
	}

	ordered := sessions
	if cursor.OrderBy.Direction == page.OrderDirectionAsc {
		ordered = slices.Clone(sessions)
		slices.Reverse(ordered)
	}

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

	end := start + limit
	if end > len(ordered) {
		end = len(ordered)
	}

	return ordered[start:end]
}

func materializeCollaborationAtHeads(
	document *automerge.Document,
	heads []automerge.Hash,
) (string, error) {
	text, err := document.Text("body")
	if err != nil {
		return "", fmt.Errorf("cannot get collaboration body at heads: %w", err)
	}

	spans, err := text.SpansAt(heads)
	if err != nil {
		return "", fmt.Errorf("cannot read collaboration spans at heads: %w", err)
	}

	content, err := automergeprosemirror.Render(spans)
	if err != nil {
		return "", fmt.Errorf("cannot render collaboration spans at heads: %w", err)
	}

	return content, nil
}

func collaborationHistoryHighlights(
	document *automerge.Document,
	before []automerge.Hash,
	after []automerge.Hash,
) ([]DocumentCollaborationHistoryHighlight, error) {
	patches, err := document.Diff(before, after)
	if err != nil {
		return nil, fmt.Errorf("cannot diff collaboration history session: %w", err)
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

	highlights := make([]DocumentCollaborationHistoryHighlight, 0)
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
				DocumentCollaborationHistoryHighlight{
					Kind: DocumentCollaborationHistoryHighlightKindInsert,
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
				DocumentCollaborationHistoryHighlight{
					Kind: DocumentCollaborationHistoryHighlightKindDelete,
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

// ParseAutomergeHashes decodes lowercase hex change hashes.
func ParseAutomergeHashes(values []string) ([]automerge.Hash, error) {
	hashes := make([]automerge.Hash, len(values))
	for i, value := range values {
		raw, err := hex.DecodeString(value)
		if err != nil {
			return nil, fmt.Errorf("cannot decode Automerge hash %q: %w", value, err)
		}

		if len(raw) != len(automerge.Hash{}) {
			return nil, fmt.Errorf(
				"cannot decode Automerge hash %q: expected %d bytes, got %d",
				value,
				len(automerge.Hash{}),
				len(raw),
			)
		}

		copy(hashes[i][:], raw)
	}

	return hashes, nil
}

// EncodeAutomergeHashes encodes change hashes as lowercase hex strings.
func EncodeAutomergeHashes(hashes []automerge.Hash) []string {
	values := make([]string, len(hashes))
	for i, hash := range hashes {
		values[i] = hash.String()
	}

	return values
}
