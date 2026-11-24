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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func TestPaginateRevisions_Tail(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()

	sessions := make([]*DocumentRevision, 5)
	for i := range sessions {
		sessions[i] = &DocumentRevision{
			ID:      gid.New(tenantID, 1),
			EndedAt: time.Unix(int64(len(sessions)-i), 0).UTC(),
		}
	}

	orderBy := page.OrderBy[DocumentRevisionOrderField]{
		Field:     DocumentRevisionOrderFieldEndedAt,
		Direction: page.OrderDirectionDesc,
	}

	t.Run(
		"last page is the oldest window",
		func(t *testing.T) {
			t.Parallel()

			cursor := page.NewCursor(2, nil, page.Tail, orderBy)
			paged := page.NewPage(paginateRevisions(sessions, cursor), cursor)

			require.Len(t, paged.Data, 2)
			assert.Equal(t, sessions[3].ID, paged.Data[0].ID)
			assert.Equal(t, sessions[4].ID, paged.Data[1].ID)
			assert.True(t, paged.Info.HasPrev)
			assert.False(t, paged.Info.HasNext)
		},
	)

	t.Run(
		"before the last item returns the preceding page",
		func(t *testing.T) {
			t.Parallel()

			key := sessions[4].CursorKey(DocumentRevisionOrderFieldEndedAt)
			cursor := page.NewCursor(2, &key, page.Tail, orderBy)
			paged := page.NewPage(paginateRevisions(sessions, cursor), cursor)

			require.Len(t, paged.Data, 2)
			assert.Equal(t, sessions[2].ID, paged.Data[0].ID)
			assert.Equal(t, sessions[3].ID, paged.Data[1].ID)
			assert.True(t, paged.Info.HasPrev)
		},
	)

	t.Run(
		"first page stays at the head",
		func(t *testing.T) {
			t.Parallel()

			cursor := page.NewCursor(2, nil, page.Head, orderBy)
			paged := page.NewPage(paginateRevisions(sessions, cursor), cursor)

			require.Len(t, paged.Data, 2)
			assert.Equal(t, sessions[0].ID, paged.Data[0].ID)
			assert.Equal(t, sessions[1].ID, paged.Data[1].ID)
			assert.True(t, paged.Info.HasNext)
			assert.False(t, paged.Info.HasPrev)
		},
	)
}

func testChanges(actor []byte, base time.Time, offsets ...time.Duration) []automerge.ChangeInfo {
	changes := make([]automerge.ChangeInfo, len(offsets))
	for i, offset := range offsets {
		changes[i] = automerge.ChangeInfo{
			Hash:  automerge.Hash{byte(i + 1)},
			Actor: actor,
			Time:  base.Add(offset),
		}
	}

	return changes
}

func testVersion(tenantID gid.TenantID, createdAt time.Time, heads ...automerge.Hash) *coredata.DocumentVersion {
	return &coredata.DocumentVersion{
		ID:                 gid.New(tenantID, coredata.DocumentVersionEntityType),
		CreatedAt:          createdAt,
		CollaborationHeads: encodeHeads(heads),
	}
}

func TestGroupRevisions_SplitsAfterTaggedChange(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	base := time.Unix(1_000_000, 0).UTC()
	changes := testChanges([]byte("alice"), base, 0, time.Minute, 2*time.Minute)

	t.Run(
		"one session without a tag",
		func(t *testing.T) {
			t.Parallel()

			sessions := groupRevisions(tenantID, changes, nil, nil)

			require.Len(t, sessions, 1)
			assert.Equal(t, 3, sessions[0].ChangeCount)
		},
	)

	t.Run(
		"tagged change ends its revision",
		func(t *testing.T) {
			t.Parallel()

			split := map[automerge.Hash]struct{}{changes[1].Hash: {}}
			sessions := groupRevisions(tenantID, changes, nil, split)

			require.Len(t, sessions, 2)
			assert.Equal(t, []automerge.Hash{changes[2].Hash}, sessions[0].Heads)
			assert.Equal(t, []automerge.Hash{changes[1].Hash}, sessions[1].Heads)
			assert.Equal(t, 2, sessions[1].ChangeCount)
			assert.Equal(t, sessions[1].Heads, sessions[0].PreviousHeads)
		},
	)
}

func TestResolveVersionTags(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	base := time.Unix(1_000_000, 0).UTC()
	changes := testChanges([]byte("alice"), base, 0, time.Minute, 2*time.Minute)

	tagged := testVersion(tenantID, base, changes[1].Hash)
	headless := &coredata.DocumentVersion{ID: gid.New(tenantID, coredata.DocumentVersionEntityType), CreatedAt: base}
	unknown := testVersion(tenantID, base, automerge.Hash{0xff})
	merged := testVersion(tenantID, base, changes[0].Hash, changes[2].Hash)

	tags, orphans := resolveVersionTags(
		[]*coredata.DocumentVersion{tagged, headless, unknown, merged},
		changes,
	)

	require.Len(t, tags, 2)
	assert.Equal(t, tagged, tags[0].version)
	assert.Equal(t, changes[1].Hash, tags[0].last)
	assert.Equal(t, merged, tags[1].version)
	assert.Equal(t, changes[2].Hash, tags[1].last)

	assert.Equal(t, []*coredata.DocumentVersion{headless, unknown}, orphans)
}

func TestVersionTags_Attach(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	base := time.Unix(1_000_000, 0).UTC()
	changes := testChanges([]byte("alice"), base, 0, time.Minute, 2*time.Minute)

	first := testVersion(tenantID, base, changes[1].Hash)
	second := testVersion(tenantID, base.Add(time.Hour), changes[1].Hash)
	merged := testVersion(tenantID, base.Add(2*time.Hour), changes[0].Hash, changes[2].Hash)

	tags, orphans := resolveVersionTags([]*coredata.DocumentVersion{first, second, merged}, changes)
	require.Empty(t, orphans)

	sessions := groupRevisions(tenantID, changes, nil, tags.splitHashes())
	tags.attach(sessions)

	require.Len(t, sessions, 3)

	assert.Equal(t, []gid.GID{merged.ID}, sessions[0].VersionIDs)
	assert.Equal(t, []automerge.Hash{changes[0].Hash, changes[2].Hash}, sessions[0].Heads)

	assert.Equal(t, []gid.GID{first.ID, second.ID}, sessions[1].VersionIDs)
	assert.Equal(t, []automerge.Hash{changes[1].Hash}, sessions[1].Heads)

	assert.Empty(t, sessions[2].VersionIDs)
}

func TestSyntheticRevision(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	documentID := gid.New(tenantID, coredata.DocumentEntityType)
	created := time.Unix(1_000_000, 0).UTC()
	published := created.Add(time.Hour)

	version := testVersion(tenantID, created)
	version.PublishedAt = &published

	revision := syntheticRevision(documentID, version)

	assert.True(t, revision.Synthetic)
	assert.NotEqual(t, version.ID, revision.ID)
	assert.Equal(t, uint16(coredata.DocumentCollaborationSessionEntityType), revision.ID.EntityType())
	assert.Equal(t, version.ID.TenantID(), revision.ID.TenantID())
	assert.Equal(t, []gid.GID{version.ID}, revision.VersionIDs)
	assert.Empty(t, revision.Heads)
	assert.Equal(t, published, revision.EndedAt)
}
