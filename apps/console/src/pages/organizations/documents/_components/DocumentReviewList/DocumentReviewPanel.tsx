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
import { useEffect, useMemo, useState } from "react";
import { createPortal } from "react-dom";
import { graphql, usePaginationFragment, useQueryLoader } from "react-relay";

import type { DocumentCommentHistoryQuery } from "#/__generated__/core/DocumentCommentHistoryQuery.graphql";
import type { DocumentReviewPanel_suggestions$key } from "#/__generated__/core/DocumentReviewPanel_suggestions.graphql";
import type { DocumentReviewPanel_threads$key } from "#/__generated__/core/DocumentReviewPanel_threads.graphql";
import type { DocumentReviewPanelSuggestionsPaginationQuery } from "#/__generated__/core/DocumentReviewPanelSuggestionsPaginationQuery.graphql";
import type { DocumentReviewPanelThreadsPaginationQuery } from "#/__generated__/core/DocumentReviewPanelThreadsPaginationQuery.graphql";
import { useLoadAllPages } from "#/lib/relay/useLoadAllPages";

import { documentCommentHistoryConnectionId, type DocumentThreadConnections } from "../../_lib/documentThreadConnections";
import type { ThreadAnchorSource } from "../../_lib/reviewAnchors";
import { scrollToReviewAnchor } from "../../_lib/scrollToReviewAnchor";
import { useDismissActiveReview } from "../../_lib/useDismissActiveReview";
import { useReviewAnchors } from "../../_lib/useReviewAnchors";
import type { ReviewCardAnchor } from "../../_lib/useReviewCardTops";
import { useReviewHighlights } from "../../_lib/useReviewHighlights";
import { documentReviewPanel } from "../../variants";

import { DocumentCommentHistory, documentCommentHistoryQuery } from "./DocumentCommentHistory";
import { DocumentCommentThreadListItem } from "./DocumentCommentThreadListItem";
import { DocumentReviewListColumn } from "./DocumentReviewListColumn";
import { type DocumentReviewView, DocumentReviewViewSwitch } from "./DocumentReviewViewSwitch";
import { DocumentSuggestionListItem } from "./DocumentSuggestionListItem";

const reviewPageSize = 50;

const documentReviewPanelThreadsFragment = graphql`
  fragment DocumentReviewPanel_threads on Document
  @refetchable(queryName: "DocumentReviewPanelThreadsPaginationQuery")
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
    ) @connection(key: "DocumentReviewPanel_commentThreads", filters: ["filter", "orderBy"]) {
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

const documentReviewPanelSuggestionsFragment = graphql`
  fragment DocumentReviewPanel_suggestions on Document
  @refetchable(queryName: "DocumentReviewPanelSuggestionsPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    suggestions(
      first: $first
      after: $after
      filter: { status: OPEN }
      orderBy: { field: CREATED_AT, direction: ASC }
    ) @connection(key: "DocumentReviewPanel_suggestions", filters: ["filter", "orderBy"]) {
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

interface DocumentReviewPanelProps {
  documentKey: DocumentReviewPanel_threads$key & DocumentReviewPanel_suggestions$key;
  editor: Editor | null;
  collaborationHandle: RepoCollaborationHandle;
  switchContainer: HTMLElement | null;
  activeKey: string | null;
  onActiveChange: (key: string | null) => void;
}

// Owns what both views share: the open threads and suggestions, where they
// sit in the text, their highlights, and the active item. The open view
// places a card beside each one; the history view stacks every thread.
export function DocumentReviewPanel({
  documentKey,
  editor,
  collaborationHandle,
  switchContainer,
  activeKey,
  onActiveChange,
}: DocumentReviewPanelProps) {
  const threadsPage = usePaginationFragment<
    DocumentReviewPanelThreadsPaginationQuery,
    DocumentReviewPanel_threads$key
  >(documentReviewPanelThreadsFragment, documentKey);
  const suggestionsPage = usePaginationFragment<
    DocumentReviewPanelSuggestionsPaginationQuery,
    DocumentReviewPanel_suggestions$key
  >(documentReviewPanelSuggestionsFragment, documentKey);
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
  const [view, setView] = useState<DocumentReviewView>("open");
  const [barElement, setBarElement] = useState<HTMLDivElement | null>(null);
  const [barHeight, setBarHeight] = useState(0);

  // The switch is portaled out of this panel, so the panel keeps a gap of
  // the same height and the cards stay beside their anchors.
  useEffect(() => {
    if (barElement == null) {
      return;
    }
    const observer = new ResizeObserver(() => {
      setBarHeight(barElement.getBoundingClientRect().height);
    });
    observer.observe(barElement);
    return () => observer.disconnect();
  }, [barElement]);
  const [historyQueryRef, loadHistory] = useQueryLoader<DocumentCommentHistoryQuery>(documentCommentHistoryQuery);
  const [historyAnchor, setHistoryAnchor] = useState<ThreadAnchorSource | null>(null);

  const reviewDocument = threadsPage.data;
  const openConnectionId = reviewDocument.commentThreads.__id;
  const threadConnections = useMemo<DocumentThreadConnections>(
    () => ({ open: openConnectionId, history: documentCommentHistoryConnectionId(reviewDocument.id) }),
    [openConnectionId, reviewDocument.id],
  );
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

  // A thread picked in the history may be resolved or belong to a
  // suggestion, so it is not among the open threads drawn in the text. It
  // joins them while it is active so its text still lights up.
  const highlightSources = useMemo(() => {
    if (
      historyAnchor == null
      || historyAnchor.id !== activeKey
      || threadSources.some(source => source.id === historyAnchor.id)
    ) {
      return threadSources;
    }
    return [...threadSources, historyAnchor];
  }, [activeKey, historyAnchor, threadSources]);
  const activeSuggestionId = activeKey != null && anchors.suggestions.has(activeKey)
    ? activeKey
    : null;
  const activeThreadId = activeSuggestionId == null ? activeKey : null;
  useReviewHighlights(editor, highlightSources, activeThreadId, activeSuggestionId);
  const markPressInside = useDismissActiveReview(activeKey != null, () => onActiveChange(null));

  const threadById = useMemo(() => new Map(threads.map(thread => [thread.id, thread])), [threads]);
  const suggestionById = useMemo(
    () => new Map(suggestions.map(suggestion => [suggestion.id, suggestion])),
    [suggestions],
  );
  const { root, bar, body } = documentReviewPanel({ view });

  function selectView(next: DocumentReviewView) {
    setView(next);
    onActiveChange(null);
    if (next === "all" && historyQueryRef == null) {
      loadHistory({ documentId: reviewDocument.id }, { fetchPolicy: "store-and-network" });
    }
  }

  function activateHistoryThread(thread: ThreadAnchorSource) {
    setHistoryAnchor(thread);
    onActiveChange(thread.id);
    if (editor != null) {
      scrollToReviewAnchor(editor, collaborationHandle, thread);
    }
  }

  function renderCard(key: string) {
    const thread = threadById.get(key);
    if (thread != null) {
      return (
        <DocumentCommentThreadListItem
          threadKey={thread}
          threadConnections={threadConnections}
          active={activeKey === key}
          onActivate={() => onActiveChange(key)}
        />
      );
    }
    const suggestion = suggestionById.get(key);
    const anchor = anchors.suggestions.get(key);
    if (suggestion == null || anchor == null) {
      return null;
    }
    return (
      <DocumentSuggestionListItem
        suggestionKey={suggestion}
        threadKey={suggestionThreads.get(key) ?? null}
        anchor={anchor}
        documentId={reviewDocument.id}
        canComment={reviewDocument.canComment}
        threadConnections={threadConnections}
        active={activeKey === key}
        onActivate={() => onActiveChange(key)}
      />
    );
  }

  return (
    <div className={root()} style={barHeight > 0 ? { paddingTop: barHeight } : undefined}>
      {switchContainer != null && createPortal(
        <div ref={setBarElement} className={bar()}>
          <DocumentReviewViewSwitch view={view} onViewChange={selectView} />
        </div>,
        switchContainer,
      )}
      {view === "open"
        ? (
            <DocumentReviewListColumn
              editor={editor}
              anchors={cardAnchors}
              activeKey={activeKey}
              renderCard={renderCard}
              onPressInside={markPressInside}
            />
          )
        : (
            <div className={body()}>
              <DocumentCommentHistory
                queryRef={historyQueryRef}
                threadConnections={threadConnections}
                activeKey={activeKey}
                onActivate={activateHistoryThread}
                onPressInside={markPressInside}
              />
            </div>
          )}
    </div>
  );
}
