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

	"go.probo.inc/probo/pkg/collaboration"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func NewDocumentCommentThread(thread *coredata.DocumentCommentThread) *DocumentCommentThread {
	return &DocumentCommentThread{
		ID:                  thread.ID,
		DocumentID:          thread.DocumentID,
		SuggestionID:        gidString(thread.SuggestionID),
		AnchorStart:         base64.StdEncoding.EncodeToString(thread.AnchorStart),
		AnchorEnd:           base64.StdEncoding.EncodeToString(thread.AnchorEnd),
		AnchorQuote:         thread.AnchorQuote,
		ResolvedAt:          thread.ResolvedAt,
		ResolvedByProfileID: gidString(thread.ResolvedByProfileID),
		CreatedAt:           thread.CreatedAt,
		UpdatedAt:           thread.UpdatedAt,
	}
}

func NewDocumentComment(comment *coredata.DocumentComment) *DocumentComment {
	return &DocumentComment{
		ID:              comment.ID,
		ThreadID:        comment.ThreadID,
		AuthorProfileID: comment.AuthorProfileID,
		Content:         comment.Content,
		CreatedAt:       comment.CreatedAt,
		UpdatedAt:       comment.UpdatedAt,
	}
}

func NewDocumentSuggestion(suggestion *coredata.DocumentSuggestion, threadID gid.GID) *DocumentSuggestion {
	node := &DocumentSuggestion{
		ID:                  suggestion.ID,
		DocumentID:          suggestion.DocumentID,
		AuthorProfileID:     suggestion.AuthorProfileID,
		Status:              suggestion.Status,
		ThreadID:            threadID,
		ResolvedAt:          suggestion.ResolvedAt,
		ResolvedByProfileID: gidString(suggestion.ResolvedByProfileID),
		CreatedAt:           suggestion.CreatedAt,
		UpdatedAt:           suggestion.UpdatedAt,
	}

	if suggestion.AnchorStart != nil {
		node.AnchorStart = new(base64.StdEncoding.EncodeToString(suggestion.AnchorStart))
		node.AnchorEnd = new(base64.StdEncoding.EncodeToString(suggestion.AnchorEnd))
	}

	if suggestion.Changes != nil {
		node.Changes = new(newDocumentSuggestionChanges(suggestion.Changes.Changes))
		node.ChangeCount = new(suggestion.Changes.ChangeCount)
	}

	return node
}

func newDocumentSuggestionChanges(changes []coredata.DocumentSuggestionChange) []*DocumentSuggestionChange {
	nodes := make([]*DocumentSuggestionChange, len(changes))
	for i, change := range changes {
		nodes[i] = &DocumentSuggestionChange{
			Kind:         change.Kind,
			Deleted:      change.Deleted,
			Inserted:     change.Inserted,
			Href:         change.Href,
			PreviousHref: change.PreviousHref,
			Before:       newDocumentSuggestionBlockLabel(change.Before),
			After:        newDocumentSuggestionBlockLabel(change.After),
		}

		if change.Mark != "" {
			nodes[i].Mark = new(collaboration.FormatSuggestionMark(change.Mark))
		}

		if change.FormatAction != "" {
			nodes[i].FormatAction = new(collaboration.FormatSuggestionAction(change.FormatAction))
		}

		if change.Node != "" {
			nodes[i].Node = new(change.Node)
		}

		if change.NodeAction != "" {
			nodes[i].NodeAction = new(change.NodeAction)
		}
	}

	return nodes
}

func newDocumentSuggestionBlockLabel(label *coredata.DocumentSuggestionBlockLabel) *DocumentSuggestionBlockLabel {
	if label == nil {
		return nil
	}

	return &DocumentSuggestionBlockLabel{
		Kind:  label.Kind,
		Level: label.Level,
	}
}

func NewListDocumentCommentThreadsOutput(
	threadPage *page.Page[*coredata.DocumentCommentThread, coredata.DocumentCommentThreadOrderField],
) ListDocumentCommentThreadsOutput {
	threads := make([]*DocumentCommentThread, 0, len(threadPage.Data))
	for _, thread := range threadPage.Data {
		threads = append(threads, NewDocumentCommentThread(thread))
	}

	return ListDocumentCommentThreadsOutput{
		NextCursor: nextCursor(threadPage.Info.HasNext, threadPage.Data, func(thread *coredata.DocumentCommentThread) page.CursorKey {
			return thread.CursorKey(threadPage.Cursor.OrderBy.Field)
		}),
		Threads: threads,
	}
}

func NewListDocumentCommentsOutput(
	commentPage *page.Page[*coredata.DocumentComment, coredata.DocumentCommentOrderField],
) ListDocumentCommentsOutput {
	comments := make([]*DocumentComment, 0, len(commentPage.Data))
	for _, comment := range commentPage.Data {
		comments = append(comments, NewDocumentComment(comment))
	}

	return ListDocumentCommentsOutput{
		NextCursor: nextCursor(commentPage.Info.HasNext, commentPage.Data, func(comment *coredata.DocumentComment) page.CursorKey {
			return comment.CursorKey(commentPage.Cursor.OrderBy.Field)
		}),
		Comments: comments,
	}
}

func NewListDocumentSuggestionsOutput(
	suggestionPage *page.Page[*coredata.DocumentSuggestion, coredata.DocumentSuggestionOrderField],
	threadIDs map[gid.GID]gid.GID,
) ListDocumentSuggestionsOutput {
	suggestions := make([]*DocumentSuggestion, 0, len(suggestionPage.Data))
	for _, suggestion := range suggestionPage.Data {
		suggestions = append(suggestions, NewDocumentSuggestion(suggestion, threadIDs[suggestion.ID]))
	}

	return ListDocumentSuggestionsOutput{
		NextCursor: nextCursor(suggestionPage.Info.HasNext, suggestionPage.Data, func(suggestion *coredata.DocumentSuggestion) page.CursorKey {
			return suggestion.CursorKey(suggestionPage.Cursor.OrderBy.Field)
		}),
		Suggestions: suggestions,
	}
}

func gidString(id *gid.GID) *string {
	if id == nil {
		return nil
	}

	value := id.String()

	return &value
}

func nextCursor[T any](hasNext bool, items []*T, key func(*T) page.CursorKey) *page.CursorKey {
	if !hasNext || len(items) == 0 {
		return nil
	}

	cursorKey := key(items[len(items)-1])

	return &cursorKey
}
