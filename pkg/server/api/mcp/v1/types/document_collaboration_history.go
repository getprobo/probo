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

package types

import (
	"encoding/hex"

	"go.probo.inc/probo/pkg/collaboration"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/server/gqlutils/types/pageinfo"
)

func NewDocumentRevision(
	session *collaboration.DocumentRevision,
) *DocumentRevision {
	var authorName *string
	if session.AuthorName != "" {
		authorName = new(session.AuthorName)
	}

	return &DocumentRevision{
		ID:            session.ID,
		Heads:         collaboration.EncodeAutomergeHashes(session.Heads),
		PreviousHeads: collaboration.EncodeAutomergeHashes(session.PreviousHeads),
		ActorID:       hex.EncodeToString(session.ActorID),
		IdentityID:    session.IdentityID,
		ProfileID:     session.ProfileID,
		AuthorName:    authorName,
		StartedAt:     session.StartedAt,
		EndedAt:       session.EndedAt,
		ChangeCount:   session.ChangeCount,
	}
}

func newPageInfo[T page.Paginable[O], O page.OrderField](p *page.Page[T, O]) *PageInfo {
	data := pageinfo.NewPageInfo(p)

	return &PageInfo{
		HasNextPage:     data.HasNextPage,
		HasPreviousPage: data.HasPreviousPage,
		StartCursor:     data.StartCursor,
		EndCursor:       data.EndCursor,
	}
}

func NewListDocumentRevisionsOutput(
	sessionPage *page.Page[*collaboration.DocumentRevision, collaboration.DocumentRevisionOrderField],
	totalCount int,
) ListDocumentRevisionsOutput {
	sessions := make([]*DocumentRevision, 0, len(sessionPage.Data))
	for _, session := range sessionPage.Data {
		sessions = append(sessions, NewDocumentRevision(session))
	}

	var nextCursor *page.CursorKey

	if len(sessionPage.Data) > 0 {
		cursorKey := sessionPage.Data[len(sessionPage.Data)-1].CursorKey(
			sessionPage.Cursor.OrderBy.Field,
		)
		nextCursor = &cursorKey
	}

	return ListDocumentRevisionsOutput{
		Revisions:  sessions,
		TotalCount: totalCount,
		PageInfo:   newPageInfo(sessionPage),
		NextCursor: nextCursor,
	}
}

func NewGetDocumentHistorySnapshotOutput(
	snapshot *collaboration.DocumentHistorySnapshot,
) GetDocumentHistorySnapshotOutput {
	highlights := make([]*DocumentHistoryHighlight, len(snapshot.Highlights))
	for i := range snapshot.Highlights {
		highlights[i] = &DocumentHistoryHighlight{
			Kind: snapshot.Highlights[i].Kind,
			From: snapshot.Highlights[i].From,
			To:   snapshot.Highlights[i].To,
			Text: snapshot.Highlights[i].Text,
		}
	}

	return GetDocumentHistorySnapshotOutput{
		Content:    snapshot.Content,
		Highlights: highlights,
	}
}
