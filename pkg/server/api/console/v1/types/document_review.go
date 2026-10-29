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
	"encoding/base64"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	DocumentCommentThreadOrderBy OrderBy[coredata.DocumentCommentThreadOrderField]
	DocumentCommentOrderBy       OrderBy[coredata.DocumentCommentOrderField]
	DocumentSuggestionOrderBy    OrderBy[coredata.DocumentSuggestionOrderField]

	DocumentCommentThreadConnection struct {
		TotalCount int
		Edges      []*DocumentCommentThreadEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filter   *coredata.DocumentCommentThreadFilter
	}

	DocumentCommentConnection struct {
		TotalCount int
		Edges      []*DocumentCommentEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}

	DocumentSuggestionConnection struct {
		TotalCount int
		Edges      []*DocumentSuggestionEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filter   *coredata.DocumentSuggestionFilter
	}
)

func NewDocumentCommentThreadConnection(
	p *page.Page[*coredata.DocumentCommentThread, coredata.DocumentCommentThreadOrderField],
	parentType any,
	parentID gid.GID,
	filter *coredata.DocumentCommentThreadFilter,
) *DocumentCommentThreadConnection {
	edges := make([]*DocumentCommentThreadEdge, len(p.Data))
	for i := range edges {
		edges[i] = NewDocumentCommentThreadEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &DocumentCommentThreadConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),
		Resolver: parentType,
		ParentID: parentID,
		Filter:   filter,
	}
}

func NewDocumentCommentThreadEdge(
	thread *coredata.DocumentCommentThread,
	orderBy coredata.DocumentCommentThreadOrderField,
) *DocumentCommentThreadEdge {
	return &DocumentCommentThreadEdge{
		Cursor: thread.CursorKey(orderBy),
		Node:   NewDocumentCommentThread(thread),
	}
}

func NewDocumentCommentThread(thread *coredata.DocumentCommentThread) *DocumentCommentThread {
	node := &DocumentCommentThread{
		ID:          thread.ID,
		Document:    &Document{ID: thread.DocumentID},
		AnchorStart: base64.StdEncoding.EncodeToString(thread.AnchorStart),
		AnchorEnd:   base64.StdEncoding.EncodeToString(thread.AnchorEnd),
		AnchorQuote: thread.AnchorQuote,
		ResolvedAt:  thread.ResolvedAt,
		CreatedAt:   thread.CreatedAt,
		UpdatedAt:   thread.UpdatedAt,
	}

	if thread.SuggestionID != nil {
		node.SuggestionID = thread.SuggestionID
		node.Suggestion = &DocumentSuggestion{ID: *thread.SuggestionID}
	}

	if thread.ResolvedByProfileID != nil {
		node.ResolvedBy = &Profile{ID: *thread.ResolvedByProfileID}
	}

	return node
}

func NewDocumentCommentConnection(
	p *page.Page[*coredata.DocumentComment, coredata.DocumentCommentOrderField],
	parentType any,
	parentID gid.GID,
) *DocumentCommentConnection {
	edges := make([]*DocumentCommentEdge, len(p.Data))
	for i := range edges {
		edges[i] = NewDocumentCommentEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &DocumentCommentConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),
		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewDocumentCommentEdge(
	comment *coredata.DocumentComment,
	orderBy coredata.DocumentCommentOrderField,
) *DocumentCommentEdge {
	return &DocumentCommentEdge{
		Cursor: comment.CursorKey(orderBy),
		Node:   NewDocumentComment(comment),
	}
}

func NewDocumentComment(comment *coredata.DocumentComment) *DocumentComment {
	return &DocumentComment{
		ID:        comment.ID,
		Thread:    &DocumentCommentThread{ID: comment.ThreadID},
		Author:    &Profile{ID: comment.AuthorProfileID},
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	}
}

func NewDocumentSuggestionConnection(
	p *page.Page[*coredata.DocumentSuggestion, coredata.DocumentSuggestionOrderField],
	parentType any,
	parentID gid.GID,
	filter *coredata.DocumentSuggestionFilter,
) *DocumentSuggestionConnection {
	edges := make([]*DocumentSuggestionEdge, len(p.Data))
	for i := range edges {
		edges[i] = NewDocumentSuggestionEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &DocumentSuggestionConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),
		Resolver: parentType,
		ParentID: parentID,
		Filter:   filter,
	}
}

func NewDocumentSuggestionEdge(
	suggestion *coredata.DocumentSuggestion,
	orderBy coredata.DocumentSuggestionOrderField,
) *DocumentSuggestionEdge {
	return &DocumentSuggestionEdge{
		Cursor: suggestion.CursorKey(orderBy),
		Node:   NewDocumentSuggestion(suggestion),
	}
}

func NewDocumentSuggestion(suggestion *coredata.DocumentSuggestion) *DocumentSuggestion {
	node := &DocumentSuggestion{
		ID:         suggestion.ID,
		Document:   &Document{ID: suggestion.DocumentID},
		Author:     &Profile{ID: suggestion.AuthorProfileID},
		Status:     suggestion.Status,
		ResolvedAt: suggestion.ResolvedAt,
		CreatedAt:  suggestion.CreatedAt,
		UpdatedAt:  suggestion.UpdatedAt,
	}

	if suggestion.ResolvedByProfileID != nil {
		node.ResolvedBy = &Profile{ID: *suggestion.ResolvedByProfileID}
	}

	return node
}
