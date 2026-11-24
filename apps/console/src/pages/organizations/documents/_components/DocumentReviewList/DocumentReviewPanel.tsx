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

import { ScrollArea } from "@base-ui/react/scroll-area";
import type { RepoCollaborationHandle } from "@probo/collaboration";
import type { Editor } from "@tiptap/core";
import { useEffect, useMemo, useState } from "react";
import { createPortal } from "react-dom";
import { graphql, usePaginationFragment, useQueryLoader } from "react-relay";

import type { DocumentCommentHistoryQuery } from "#/__generated__/core/DocumentCommentHistoryQuery.graphql";
import type { DocumentReviewPanel_threads$key } from "#/__generated__/core/DocumentReviewPanel_threads.graphql";
import type { DocumentReviewPanelThreadsPaginationQuery } from "#/__generated__/core/DocumentReviewPanelThreadsPaginationQuery.graphql";
import { useLoadAllPages } from "#/lib/relay/useLoadAllPages";

import { type DocumentCommentDraft, documentCommentDraftKey } from "../../_lib/documentCommentDraft";
import { documentCommentHistoryConnectionId, type DocumentThreadConnections } from "../../_lib/documentThreadConnections";
import type { ThreadAnchorSource } from "../../_lib/reviewAnchors";
import type { ReviewHistoryTarget } from "../../_lib/reviewHistoryTarget";
import { scrollToReviewAnchor, scrollToReviewPosition } from "../../_lib/scrollToReviewAnchor";
import { useDismissActiveReview } from "../../_lib/useDismissActiveReview";
import { useReviewAnchors } from "../../_lib/useReviewAnchors";
import type { ReviewCardAnchor } from "../../_lib/useReviewCardTops";
import { useReviewHighlights } from "../../_lib/useReviewHighlights";
import { documentReviewPanel } from "../../variants";

import { DocumentCommentDraftCard } from "./DocumentCommentDraftCard";
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
    # Must match documentOpenThreadsConnectionId.
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
          suggestion {
            id
            status
          }
          ...DocumentCommentThreadListItem_thread
          ...DocumentSuggestionListItem_thread
        }
      }
    }
  }
`;

interface DocumentReviewPanelProps {
  documentKey: DocumentReviewPanel_threads$key;
  editor: Editor | null;
  collaborationHandle: RepoCollaborationHandle;
  switchContainer: HTMLElement | null;
  activeKey: string | null;
  onActiveChange: (key: string | null) => void;
  commentDraft: DocumentCommentDraft | null;
  onCommentDraftChange: (draft: DocumentCommentDraft | null) => void;
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
  commentDraft,
  onCommentDraftChange,
}: DocumentReviewPanelProps) {
  const threadsPage = usePaginationFragment<
    DocumentReviewPanelThreadsPaginationQuery,
    DocumentReviewPanel_threads$key
  >(documentReviewPanelThreadsFragment, documentKey);
  const { retry: retryThreadPages } = useLoadAllPages({
    hasNext: threadsPage.hasNext,
    isLoadingNext: threadsPage.isLoadingNext,
    loadNext: threadsPage.loadNext,
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
  const refetchThreads = threadsPage.refetch;
  const historyLoaded = historyQueryRef != null;

  // A review notice means threads or suggestions changed elsewhere: another
  // reviewer, or the server rejecting a suggestion whose edits were undone.
  useEffect(() => {
    return collaborationHandle.onReview?.(() => {
      refetchThreads({}, { fetchPolicy: "store-and-network", onComplete: retryThreadPages });
      if (historyLoaded) {
        loadHistory({ documentId: reviewDocument.id }, { fetchPolicy: "store-and-network" });
      }
    });
  }, [collaborationHandle, historyLoaded, loadHistory, refetchThreads, retryThreadPages, reviewDocument.id]);
  const openConnectionId = reviewDocument.commentThreads.__id;
  const threadConnections = useMemo<DocumentThreadConnections>(
    () => ({ open: openConnectionId, history: documentCommentHistoryConnectionId(reviewDocument.id) }),
    [openConnectionId, reviewDocument.id],
  );
  const threadEdges = reviewDocument.commentThreads.edges;
  // Comment cards are keyed by thread and suggestion cards by suggestion,
  // which is the id the editor marks carry. A decided suggestion whose thread
  // has not left the connection yet stays off the column.
  const { threads, suggestionThreads } = useMemo(() => {
    const comments: Array<(typeof threadEdges)[number]["node"]> = [];
    const bySuggestion = new Map<string, (typeof threadEdges)[number]["node"]>();
    for (const { node } of threadEdges) {
      if (node.suggestion == null) {
        comments.push(node);
      } else if (node.suggestion.status === "OPEN") {
        bySuggestion.set(node.suggestion.id, node);
      }
    }
    return { threads: comments, suggestionThreads: bySuggestion };
  }, [threadEdges]);
  const threadSources = useMemo(
    () => threads.map(({ id, anchorStart, anchorEnd }) => ({ id, anchorStart, anchorEnd })),
    [threads],
  );

  const anchors = useReviewAnchors(editor, collaborationHandle, threadSources);
  const cardAnchors = useMemo(() => {
    const list: ReviewCardAnchor[] = [];
    for (const thread of threads) {
      list.push({ key: thread.id, pos: anchors.threads.get(thread.id) ?? Number.MAX_SAFE_INTEGER });
    }
    for (const suggestionId of suggestionThreads.keys()) {
      const anchor = anchors.suggestions.get(suggestionId);
      if (anchor) list.push({ key: suggestionId, pos: anchor.from });
    }
    if (commentDraft != null) {
      list.push({ key: documentCommentDraftKey, pos: commentDraft.pos });
    }
    return list;
  }, [anchors, commentDraft, suggestionThreads, threads]);

  // A thread picked in the history may be resolved or belong to a
  // suggestion, so it is not among the open threads drawn in the text. It
  // joins them while it is active so its text still lights up.
  const highlightSources = useMemo(() => {
    if (commentDraft != null) {
      const { anchorStart, anchorEnd } = commentDraft;
      return [...threadSources, { id: documentCommentDraftKey, anchorStart, anchorEnd }];
    }
    if (
      historyAnchor == null
      || historyAnchor.id !== activeKey
      || threadSources.some(source => source.id === historyAnchor.id)
    ) {
      return threadSources;
    }
    return [...threadSources, historyAnchor];
  }, [activeKey, commentDraft, historyAnchor, threadSources]);
  const layoutActiveKey = commentDraft != null ? documentCommentDraftKey : activeKey;
  const activeSuggestionId = activeKey != null && anchors.suggestions.has(activeKey)
    ? activeKey
    : null;
  const activeThreadId = activeSuggestionId == null ? layoutActiveKey : null;
  useReviewHighlights(editor, highlightSources, activeThreadId, activeSuggestionId);
  const markPressInside = useDismissActiveReview(activeKey != null, () => onActiveChange(null));

  const threadById = useMemo(() => new Map(threads.map(thread => [thread.id, thread])), [threads]);
  const { root, bar, body, viewport, content } = documentReviewPanel({ view });

  function selectView(next: DocumentReviewView) {
    setView(next);
    onActiveChange(null);
    onCommentDraftChange(null);
    if (next === "all" && historyQueryRef == null) {
      loadHistory({ documentId: reviewDocument.id }, { fetchPolicy: "store-and-network" });
    }
  }

  function activateHistoryEntry(target: ReviewHistoryTarget) {
    if (target.kind === "suggestion") {
      setHistoryAnchor(null);
      onActiveChange(target.id);
      const anchor = anchors.suggestions.get(target.id);
      if (editor != null && anchor != null) {
        scrollToReviewPosition(editor, anchor.from);
      }
      return;
    }
    setHistoryAnchor(target.anchor);
    onActiveChange(target.anchor.id);
    if (editor != null) {
      scrollToReviewAnchor(editor, collaborationHandle, target.anchor);
    }
  }

  function renderCard(key: string) {
    if (key === documentCommentDraftKey && commentDraft != null) {
      return (
        <DocumentCommentDraftCard
          documentId={reviewDocument.id}
          draft={commentDraft}
          threadConnections={threadConnections}
          onDone={(threadId) => {
            onCommentDraftChange(null);
            if (threadId != null) {
              onActiveChange(threadId);
            }
          }}
        />
      );
    }
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
    const suggestionThread = suggestionThreads.get(key);
    const anchor = anchors.suggestions.get(key);
    if (suggestionThread == null || anchor == null) {
      return null;
    }
    return (
      <DocumentSuggestionListItem
        threadKey={suggestionThread}
        anchor={anchor}
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
              activeKey={layoutActiveKey}
              renderCard={renderCard}
              onPressInside={markPressInside}
            />
          )
        : (
            <ScrollArea.Root className={body()}>
              <ScrollArea.Viewport className={viewport()}>
                <ScrollArea.Content className={content()}>
                  <DocumentCommentHistory
                    queryRef={historyQueryRef}
                    threadConnections={threadConnections}
                    liveSuggestions={anchors.suggestions}
                    activeKey={activeKey}
                    onActivate={activateHistoryEntry}
                    onPressInside={markPressInside}
                  />
                </ScrollArea.Content>
              </ScrollArea.Viewport>
            </ScrollArea.Root>
          )}
    </div>
  );
}
