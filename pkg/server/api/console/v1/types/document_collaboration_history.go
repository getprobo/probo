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
	"time"

	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/probo"
)

type (
	DocumentRevisionOrderBy = OrderBy[probo.DocumentRevisionOrderField]

	DocumentRevision struct {
		ID            gid.GID   `json:"id"`
		DocumentID    gid.GID   `json:"-"`
		Heads         []string  `json:"heads"`
		PreviousHeads []string  `json:"previousHeads"`
		ActorID       string    `json:"actorID"`
		AuthorName    *string   `json:"authorName"`
		ProfileID     *gid.GID  `json:"-"`
		StartedAt     time.Time `json:"startedAt"`
		EndedAt       time.Time `json:"endedAt"`
		ChangeCount   int       `json:"changeCount"`
	}

	DocumentRevisionConnection struct {
		TotalCount int
		Edges      []*DocumentRevisionEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}

	DocumentHistoryHighlight struct {
		Kind probo.DocumentHistoryHighlightKind `json:"kind"`
		From int                                `json:"from"`
		To   int                                `json:"to"`
		Text string                             `json:"text"`
	}

	DocumentHistorySnapshot struct {
		Content    *DocumentContent            `json:"content"`
		Highlights []*DocumentHistoryHighlight `json:"highlights"`
	}
)

func NewDocumentRevision(
	session *probo.DocumentRevision,
) *DocumentRevision {
	var authorName *string
	if session.AuthorName != "" {
		authorName = new(session.AuthorName)
	}

	return &DocumentRevision{
		ID:            session.ID,
		DocumentID:    session.DocumentID,
		Heads:         probo.EncodeAutomergeHashes(session.Heads),
		PreviousHeads: probo.EncodeAutomergeHashes(session.PreviousHeads),
		ActorID:       hex.EncodeToString(session.ActorID),
		AuthorName:    authorName,
		ProfileID:     session.ProfileID,
		StartedAt:     session.StartedAt,
		EndedAt:       session.EndedAt,
		ChangeCount:   session.ChangeCount,
	}
}

func NewDocumentRevisionConnection(
	p *page.Page[*probo.DocumentRevision, probo.DocumentRevisionOrderField],
	parentType any,
	parentID gid.GID,
	totalCount int,
) *DocumentRevisionConnection {
	edges := make([]*DocumentRevisionEdge, len(p.Data))
	for i, session := range p.Data {
		edges[i] = &DocumentRevisionEdge{
			Cursor: session.CursorKey(p.Cursor.OrderBy.Field),
			Node:   NewDocumentRevision(session),
		}
	}

	return &DocumentRevisionConnection{
		TotalCount: totalCount,
		Edges:      edges,
		PageInfo:   *NewPageInfo(p),
		Resolver:   parentType,
		ParentID:   parentID,
	}
}

func NewDocumentHistorySnapshot(
	snapshot *probo.DocumentHistorySnapshot,
) *DocumentHistorySnapshot {
	highlights := make([]*DocumentHistoryHighlight, len(snapshot.Highlights))
	for i := range snapshot.Highlights {
		highlights[i] = &DocumentHistoryHighlight{
			Kind: snapshot.Highlights[i].Kind,
			From: snapshot.Highlights[i].From,
			To:   snapshot.Highlights[i].To,
			Text: snapshot.Highlights[i].Text,
		}
	}

	return &DocumentHistorySnapshot{
		Content:    NewDocumentContent(snapshot.Content),
		Highlights: highlights,
	}
}
