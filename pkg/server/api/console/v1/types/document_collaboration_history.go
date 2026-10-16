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
	DocumentCollaborationHistorySessionOrderBy = OrderBy[probo.DocumentCollaborationHistorySessionOrderField]

	DocumentCollaborationHistorySession struct {
		ID            gid.GID   `json:"id"`
		Heads         []string  `json:"heads"`
		PreviousHeads []string  `json:"previousHeads"`
		ActorID       string    `json:"actorID"`
		AuthorName    *string   `json:"authorName"`
		ProfileID     *gid.GID  `json:"-"`
		StartedAt     time.Time `json:"startedAt"`
		EndedAt       time.Time `json:"endedAt"`
		ChangeCount   int       `json:"changeCount"`
	}

	DocumentCollaborationHistorySessionConnection struct {
		TotalCount int
		Edges      []*DocumentCollaborationHistorySessionEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}

	DocumentCollaborationHistoryHighlight struct {
		Kind probo.DocumentCollaborationHistoryHighlightKind `json:"kind"`
		From int                                             `json:"from"`
		To   int                                             `json:"to"`
		Text string                                          `json:"text"`
	}

	DocumentCollaborationHistorySnapshot struct {
		Content    string                                   `json:"content"`
		Highlights []*DocumentCollaborationHistoryHighlight `json:"highlights"`
	}
)

func NewDocumentCollaborationHistorySession(
	session *probo.DocumentCollaborationHistorySession,
) *DocumentCollaborationHistorySession {
	var authorName *string
	if session.AuthorName != "" {
		authorName = new(session.AuthorName)
	}

	return &DocumentCollaborationHistorySession{
		ID:            session.ID,
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

func NewDocumentCollaborationHistorySessionConnection(
	p *page.Page[*probo.DocumentCollaborationHistorySession, probo.DocumentCollaborationHistorySessionOrderField],
	parentType any,
	parentID gid.GID,
	totalCount int,
) *DocumentCollaborationHistorySessionConnection {
	edges := make([]*DocumentCollaborationHistorySessionEdge, len(p.Data))
	for i, session := range p.Data {
		edges[i] = &DocumentCollaborationHistorySessionEdge{
			Cursor: session.CursorKey(p.Cursor.OrderBy.Field),
			Node:   NewDocumentCollaborationHistorySession(session),
		}
	}

	return &DocumentCollaborationHistorySessionConnection{
		TotalCount: totalCount,
		Edges:      edges,
		PageInfo:   *NewPageInfo(p),
		Resolver:   parentType,
		ParentID:   parentID,
	}
}

func NewDocumentCollaborationHistorySnapshot(
	snapshot *probo.DocumentCollaborationHistorySnapshot,
) *DocumentCollaborationHistorySnapshot {
	highlights := make([]*DocumentCollaborationHistoryHighlight, len(snapshot.Highlights))
	for i := range snapshot.Highlights {
		highlights[i] = &DocumentCollaborationHistoryHighlight{
			Kind: snapshot.Highlights[i].Kind,
			From: snapshot.Highlights[i].From,
			To:   snapshot.Highlights[i].To,
			Text: snapshot.Highlights[i].Text,
		}
	}

	return &DocumentCollaborationHistorySnapshot{
		Content:    snapshot.Content,
		Highlights: highlights,
	}
}
