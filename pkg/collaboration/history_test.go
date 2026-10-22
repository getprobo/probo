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

func TestAutomergeIndexToPlainOffset_CountsBlockBoundaries(t *testing.T) {
	t.Parallel()

	spans := []automerge.Span{
		{Type: automerge.SpanTypeBlock, Block: map[string]any{"type": "paragraph"}},
		{Type: automerge.SpanTypeText, Text: "ab"},
		{Type: automerge.SpanTypeBlock, Block: map[string]any{"type": "paragraph"}},
		{Type: automerge.SpanTypeText, Text: "c"},
	}

	boundary, ok := automergeIndexToPlainOffset(spans, 3)
	require.True(t, ok)
	assert.Equal(t, 3, boundary)

	start, ok := automergeIndexToPlainOffset(spans, 4)
	require.True(t, ok)
	assert.Equal(t, 4, start)
}
