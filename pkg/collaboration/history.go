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

		// VersionIDs are the versions tagging this revision, oldest first.
		// A published version is a tag on the revision it was taken at.
		VersionIDs []gid.GID
		// Synthetic marks a version that predates the link to the history,
		// or whose revision no longer exists. It stands alone in the
		// timeline under an ID derived from the version ID and carries no heads.
		Synthetic bool
	}

	// DocumentRevisionFilter narrows the history. Versioned keeps only the
	// revisions tagged by at least one version.
	DocumentRevisionFilter struct {
		Versioned bool
	}

	versionTag struct {
		version *coredata.DocumentVersion
		heads   []automerge.Hash
		last    automerge.Hash
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
	filter *DocumentRevisionFilter,
) (*page.Page[*DocumentRevision, DocumentRevisionOrderField], int, error) {
	sessions, err := s.loadRevisions(ctx, scope, documentID)
	if err != nil {
		return nil, 0, err
	}

	if filter != nil && filter.Versioned {
		sessions = slices.DeleteFunc(
			sessions,
			func(session *DocumentRevision) bool { return len(session.VersionIDs) == 0 },
		)
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
	var versions coredata.DocumentVersions

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return versions.LoadTagsByDocumentID(ctx, conn, scope, documentID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load document version tags: %w", err)
	}

	var changes []automerge.ChangeInfo

	opened, err := s.Open(ctx, scope, documentID)

	switch {
	case err == nil:
		defer func() { _ = opened.Document.Close() }()

		changes, err = opened.Document.ChangeHistory()
		if err != nil {
			return nil, fmt.Errorf("cannot read collaboration change history: %w", err)
		}
	case errors.Is(err, ErrDocumentCollaborationNotFound):
	default:
		return nil, fmt.Errorf("cannot open collaboration for history: %w", err)
	}

	actors, err := s.loadActorAttributions(ctx, scope, documentID, actorIDsFromChanges(changes))
	if err != nil {
		return nil, err
	}

	tags, orphans := resolveVersionTags(versions, changes)

	sessions := groupRevisions(scope.GetTenantID(), changes, actors, tags.splitHashes())
	for _, session := range sessions {
		session.DocumentID = documentID
	}

	tags.attach(sessions)

	for i, session := range sessions {
		session.PreviousHeads = nil
		if i+1 < len(sessions) {
			session.PreviousHeads = slices.Clone(sessions[i+1].Heads)
		}
	}

	for _, orphan := range orphans {
		sessions = append(sessions, syntheticRevision(documentID, orphan))
	}

	slices.SortStableFunc(
		sessions,
		func(a, b *DocumentRevision) int { return b.EndedAt.Compare(a.EndedAt) },
	)

	return sessions, nil
}

type versionTags []versionTag

// resolveVersionTags splits versions into those whose heads are all part of
// the change history and the orphans that are not: versions that predate the
// link, and versions whose history was replaced since.
func resolveVersionTags(
	versions []*coredata.DocumentVersion,
	changes []automerge.ChangeInfo,
) (versionTags, []*coredata.DocumentVersion) {
	order := make(map[automerge.Hash]int, len(changes))
	for i, change := range changes {
		order[change.Hash] = i
	}

	var (
		tags    versionTags
		orphans []*coredata.DocumentVersion
	)

	for _, version := range versions {
		heads, err := decodeHeads(version.CollaborationHeads)
		if err != nil || len(heads) == 0 {
			orphans = append(orphans, version)

			continue
		}

		last := -1
		known := true

		for _, head := range heads {
			index, ok := order[head]
			if !ok {
				known = false

				break
			}

			last = max(last, index)
		}

		if !known {
			orphans = append(orphans, version)

			continue
		}

		tags = append(tags, versionTag{
			version: version,
			heads:   heads,
			last:    changes[last].Hash,
		})
	}

	return tags, orphans
}

// splitHashes are the changes after which a new revision must start, so that
// a tagged change always ends its revision instead of sitting mid-session.
func (t versionTags) splitHashes() map[automerge.Hash]struct{} {
	hashes := make(map[automerge.Hash]struct{}, len(t))

	for _, tag := range t {
		for _, head := range tag.heads {
			hashes[head] = struct{}{}
		}
	}

	return hashes
}

func (t versionTags) attach(sessions []*DocumentRevision) {
	byLast := make(map[automerge.Hash]*DocumentRevision, len(sessions))
	for _, session := range sessions {
		byLast[session.Heads[0]] = session
	}

	for _, tag := range t {
		session, ok := byLast[tag.last]
		if !ok {
			continue
		}

		if len(session.VersionIDs) == 0 && len(tag.heads) > 1 {
			session.Heads = slices.Clone(tag.heads)
		}

		session.VersionIDs = append(session.VersionIDs, tag.version.ID)
	}
}

func syntheticRevision(
	documentID gid.GID,
	version *coredata.DocumentVersion,
) *DocumentRevision {
	at := version.CreatedAt
	if version.PublishedAt != nil {
		at = *version.PublishedAt
	}

	return &DocumentRevision{
		ID:         syntheticRevisionID(version.ID),
		DocumentID: documentID,
		StartedAt:  at,
		EndedAt:    at,
		VersionIDs: []gid.GID{version.ID},
		Synthetic:  true,
	}
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

	opened, err := s.Open(ctx, scope, documentID)
	if err != nil {
		if errors.Is(err, ErrDocumentCollaborationNotFound) {
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
		if errors.Is(err, ErrDocumentCollaborationNotFound) {
			return 0, err
		}

		return 0, fmt.Errorf("cannot open collaboration for history restore: %w", err)
	}

	defer func() { _ = opened.Document.Close() }()

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
	splitAfter map[automerge.Hash]struct{},
) []*DocumentRevision {
	if len(changes) == 0 {
		return nil
	}

	sessions := make([]*DocumentRevision, 0)

	var (
		current *DocumentRevision
		closed  bool
	)

	for _, change := range changes {
		actorKey := string(change.Actor)

		changeTime := change.Time
		if changeTime.IsZero() {
			changeTime = time.Unix(0, 0).UTC()
		}

		startsNew := current == nil ||
			closed ||
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

			_, closed = splitAfter[change.Hash]

			continue
		}

		current.ID = revisionID(tenantID, change.Hash)
		current.Heads = []automerge.Hash{change.Hash}
		current.EndedAt = changeTime
		current.ChangeCount++

		_, closed = splitAfter[change.Hash]
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

// syntheticRevisionID derives the revision ID from the version ID. It must
// differ from the version ID: clients key their caches by ID, and the two
// objects would otherwise overwrite each other.
func syntheticRevisionID(versionID gid.GID) gid.GID {
	id := versionID
	binary.BigEndian.PutUint16(id[8:10], coredata.DocumentCollaborationSessionEntityType)

	return id
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
