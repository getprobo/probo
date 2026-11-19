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

import type { RepoCollaborationHandle } from "@probo/collaboration";
import type { Editor } from "@tiptap/core";
import { useMemo } from "react";
import { graphql, usePaginationFragment } from "react-relay";

import type { DocumentReviewListColumn_suggestions$key } from "#/__generated__/core/DocumentReviewListColumn_suggestions.graphql";
import type { DocumentReviewListColumn_threads$key } from "#/__generated__/core/DocumentReviewListColumn_threads.graphql";
import type { DocumentReviewListColumnSuggestionsPaginationQuery } from "#/__generated__/core/DocumentReviewListColumnSuggestionsPaginationQuery.graphql";
import type { DocumentReviewListColumnThreadsPaginationQuery } from "#/__generated__/core/DocumentReviewListColumnThreadsPaginationQuery.graphql";
import { useLoadAllPages } from "#/lib/relay/useLoadAllPages";

import { useDismissActiveReview } from "../../_lib/useDismissActiveReview";
import { useReviewAnchors } from "../../_lib/useReviewAnchors";
import { type ReviewCardAnchor, useReviewCardTops } from "../../_lib/useReviewCardTops";
import { useReviewHighlights } from "../../_lib/useReviewHighlights";
import { documentReviewList } from "../../variants";

import { DocumentCommentThreadListItem } from "./DocumentCommentThreadListItem";
import { DocumentSuggestionListItem } from "./DocumentSuggestionListItem";

const reviewPageSize = 50;

const documentReviewListColumnThreadsFragment = graphql`
  fragment DocumentReviewListColumn_threads on Document
  @refetchable(queryName: "DocumentReviewListColumnThreadsPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    id
    canComment: permission(action: "core:document:comment")
    commentThreads(
      first: $first
      after: $after
      filter: { resolved: false }
      orderBy: { field: CREATED_AT, direction: ASC }
    ) @connection(key: "DocumentReviewListColumn_commentThreads", filters: ["filter", "orderBy"]) {
      __id
      edges {
        node {
          id
          anchorStart
          anchorEnd
          suggestionId
          ...DocumentCommentThreadListItem_thread
          ...DocumentSuggestionListItem_thread
        }
      }
    }
  }
`;

const documentReviewListColumnSuggestionsFragment = graphql`
  fragment DocumentReviewListColumn_suggestions on Document
  @refetchable(queryName: "DocumentReviewListColumnSuggestionsPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    suggestions(
      first: $first
      after: $after
      filter: { status: OPEN }
      orderBy: { field: CREATED_AT, direction: ASC }
    ) @connection(key: "DocumentReviewListColumn_suggestions", filters: ["filter", "orderBy"]) {
      edges {
        node {
          id
          status
          ...DocumentSuggestionListItem_suggestion
        }
      }
    }
  }
`;

interface DocumentReviewListColumnProps {
  documentKey: DocumentReviewListColumn_threads$key & DocumentReviewListColumn_suggestions$key;
  editor: Editor | null;
  collaborationHandle: RepoCollaborationHandle;
  activeKey: string | null;
  onActiveChange: (key: string | null) => void;
}

export function DocumentReviewListColumn({
  documentKey,
  editor,
  collaborationHandle,
  activeKey,
  onActiveChange,
}: DocumentReviewListColumnProps) {
  const threadsPage = usePaginationFragment<
    DocumentReviewListColumnThreadsPaginationQuery,
    DocumentReviewListColumn_threads$key
  >(documentReviewListColumnThreadsFragment, documentKey);
  const suggestionsPage = usePaginationFragment<
    DocumentReviewListColumnSuggestionsPaginationQuery,
    DocumentReviewListColumn_suggestions$key
  >(documentReviewListColumnSuggestionsFragment, documentKey);
  useLoadAllPages({
    hasNext: threadsPage.hasNext,
    isLoadingNext: threadsPage.isLoadingNext,
    loadNext: threadsPage.loadNext,
    pageSize: reviewPageSize,
  });
  useLoadAllPages({
    hasNext: suggestionsPage.hasNext,
    isLoadingNext: suggestionsPage.isLoadingNext,
    loadNext: suggestionsPage.loadNext,
    pageSize: reviewPageSize,
  });

  const reviewDocument = threadsPage.data;
  const threadsConnectionId = reviewDocument.commentThreads.__id;
  const suggestionEdges = suggestionsPage.data.suggestions.edges;
  const threadEdges = reviewDocument.commentThreads.edges;
  // An accepted or rejected suggestion stays in the connection until the next
  // fetch, so its status keeps it off the column.
  const suggestions = useMemo(
    () => suggestionEdges.map(edge => edge.node).filter(node => node.status === "OPEN"),
    [suggestionEdges],
  );
  // A comment on an open suggestion belongs to that suggestion, not its own card.
  const { threads, suggestionThreads } = useMemo(() => {
    const suggestionIds = new Set(suggestions.map(suggestion => suggestion.id));
    const standalone: Array<(typeof threadEdges)[number]["node"]> = [];
    const bySuggestion = new Map<string, (typeof threadEdges)[number]["node"]>();
    for (const { node } of threadEdges) {
      if (node.suggestionId != null && suggestionIds.has(node.suggestionId)) {
        if (!bySuggestion.has(node.suggestionId)) bySuggestion.set(node.suggestionId, node);
      } else {
        standalone.push(node);
      }
    }
    return { threads: standalone, suggestionThreads: bySuggestion };
  }, [suggestions, threadEdges]);
  const threadSources = useMemo(
    () => threads.map(({ id, anchorStart, anchorEnd }) => ({ id, anchorStart, anchorEnd })),
    [threads],
  );

  const anchors = useReviewAnchors(editor, collaborationHandle, threadSources);
  const cardAnchors = useMemo(() => {
    const list: ReviewCardAnchor[] = [];
    for (const thread of threads) {
      const pos = anchors.threads.get(thread.id);
      if (pos != null) list.push({ key: thread.id, pos });
    }
    for (const suggestion of suggestions) {
      const anchor = anchors.suggestions.get(suggestion.id);
      if (anchor) list.push({ key: suggestion.id, pos: anchor.from });
    }
    return list;
  }, [anchors, suggestions, threads]);
  const { columnRef, bindCard, tops, height } = useReviewCardTops(editor, cardAnchors, activeKey);
  const activeSuggestionId = activeKey != null && anchors.suggestions.has(activeKey)
    ? activeKey
    : null;
  const activeThreadId = activeSuggestionId == null ? activeKey : null;
  useReviewHighlights(editor, threadSources, activeThreadId, activeSuggestionId);
  const markPressInside = useDismissActiveReview(activeKey != null, () => onActiveChange(null));

  const { column, card } = documentReviewList();

  return (
    <div ref={columnRef} className={column()} style={{ minHeight: height }}>
      {threads
        .filter(thread => anchors.threads.has(thread.id))
        .map(thread => (
          <div
            key={thread.id}
            ref={bindCard}
            data-review-card={thread.id}
            className={card({ placed: tops[thread.id] != null })}
            style={{ top: tops[thread.id] ?? 0 }}
            onPointerDownCapture={markPressInside}
          >
            <DocumentCommentThreadListItem
              threadKey={thread}
              threadsConnectionId={threadsConnectionId}
              active={activeKey === thread.id}
              onActivate={() => onActiveChange(thread.id)}
            />
          </div>
        ))}
      {suggestions.flatMap((suggestion) => {
        const anchor = anchors.suggestions.get(suggestion.id);
        if (anchor == null) return [];
        return [(
          <div
            key={suggestion.id}
            ref={bindCard}
            data-review-card={suggestion.id}
            className={card({ placed: tops[suggestion.id] != null })}
            style={{ top: tops[suggestion.id] ?? 0 }}
            onPointerDownCapture={markPressInside}
          >
            <DocumentSuggestionListItem
              suggestionKey={suggestion}
              threadKey={suggestionThreads.get(suggestion.id) ?? null}
              anchor={anchor}
              documentId={reviewDocument.id}
              canComment={reviewDocument.canComment}
              threadsConnectionId={threadsConnectionId}
              active={activeKey === suggestion.id}
              onActivate={() => onActiveChange(suggestion.id)}
            />
          </div>
        )];
      })}
    </div>
  );
}
