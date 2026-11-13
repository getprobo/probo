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

import { ChatCircleIcon } from "@phosphor-icons/react";
import type { RepoCollaborationHandle } from "@probo/collaboration";
import { encodeDocumentAnchor } from "@probo/ui/src/RichEditor/commentAnchors";
import { presenceFromPmSelection, richEditorPresenceAdapter } from "@probo/ui/src/RichEditor/repoPresence";
import {
  refreshSuggestionFocus,
  type SuggestingController,
} from "@probo/ui/src/RichEditor/SuggestionExtension";
import type { Editor } from "@tiptap/core";
import { type MutableRefObject, useCallback, useContext, useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment, useLazyLoadQuery, useRelayEnvironment } from "react-relay";
import { fetchQuery, graphql } from "relay-runtime";

import type { DocumentReviewMargin_document$key } from "#/__generated__/core/DocumentReviewMargin_document.graphql";
import type { DocumentReviewMarginQuery } from "#/__generated__/core/DocumentReviewMarginQuery.graphql";
import { useMentionMembers } from "#/pages/organizations/_lib/useMentionMembers";
import { CurrentUser } from "#/providers/CurrentUser";

import {
  commentBodySignature,
  commentThreadsWithoutSuggestions,
  threadSignature,
} from "../../_lib/reviewMarginLayout";
import type { StartedSuggestion } from "../../_lib/useDocumentReviewSuggestionSession";
import {
  type CommentDraft,
  useReviewMarginLayout,
} from "../../_lib/useReviewMarginLayout";

import { DocumentReviewMarginHistory } from "./DocumentReviewMarginHistory";
import { DocumentReviewMarginList } from "./DocumentReviewMarginList";
import { DocumentReviewSuggestionSession } from "./DocumentReviewSuggestionSession";

interface DocumentReviewMarginProps {
  documentId: string;
  editor: Editor;
  collaborationHandle: RepoCollaborationHandle;
  canUpdate: boolean;
  canComment: boolean;
  suggesting: MutableRefObject<SuggestingController>;
  activeThreadId: string | null;
  activeSuggestionId: string | null;
  reviewVersion: number;
  toolbarHost: HTMLElement | null;
  marginHost: HTMLElement | null;
  onActiveThread: (id: string | null) => void;
  onActiveSuggestion: (id: string | null) => void;
  onThreads: (
    threads: Array<{
      id: string;
      anchorStart: string;
      anchorEnd: string;
      suggestionId: string | null;
    }>,
  ) => void;
  onOwnReview: () => void;
  onOwnReviewFailed: () => void;
}

export function DocumentReviewMargin(props: DocumentReviewMarginProps) {
  const {
    documentId,
    editor,
    collaborationHandle,
    canUpdate,
    canComment,
    suggesting,
    activeThreadId,
    activeSuggestionId,
    reviewVersion,
    toolbarHost,
    marginHost,
    onActiveThread,
    onActiveSuggestion,
    onThreads,
    onOwnReview,
    onOwnReviewFailed,
  } = props;
  const { t } = useTranslation();
  const currentUser = useContext(CurrentUser);
  const mentionSource = useMentionMembers();
  const environment = useRelayEnvironment();
  const queryData = useLazyLoadQuery<DocumentReviewMarginQuery>(
    reviewQuery,
    { documentId: documentId },
    { fetchPolicy: "store-and-network" },
  );
  const documentKey = queryData.document?.__typename === "Document"
    ? queryData.document
    : null;
  const reviewDocument = useFragment<DocumentReviewMargin_document$key>(
    reviewDocumentFragment,
    documentKey,
  );
  const seenReviewVersion = useRef(reviewVersion);
  useEffect(() => {
    if (seenReviewVersion.current === reviewVersion) return;
    seenReviewVersion.current = reviewVersion;
    const subscription = fetchQuery<DocumentReviewMarginQuery>(
      environment,
      reviewQuery,
      { documentId: documentId },
      { fetchPolicy: "network-only" },
    ).subscribe({});
    return () => subscription.unsubscribe();
  }, [environment, documentId, reviewVersion]);

  const [draft, setDraft] = useState<CommentDraft | null>(null);
  const activeCardKey = `${activeThreadId ?? ""}:${activeSuggestionId ?? ""}`;
  const [trackedActiveCardKey, setTrackedActiveCardKey] = useState(activeCardKey);
  if (trackedActiveCardKey !== activeCardKey) {
    setTrackedActiveCardKey(activeCardKey);
    if (activeThreadId || activeSuggestionId) setDraft(null);
  }
  const [historyOpen, setHistoryOpen] = useState(false);
  const [startedSuggestions, setStartedSuggestions] = useState<StartedSuggestion[]>([]);
  const removeStartedSuggestionRef = useRef<((id: string) => void) | null>(null);
  const onStartedSuggestions = useCallback((suggestions: StartedSuggestion[]) => {
    setStartedSuggestions(suggestions);
  }, []);

  const threads = reviewDocument?.commentThreads.edges.map(edge => edge.node) ?? [];
  const loadedSuggestions = reviewDocument?.suggestions.edges.map(edge => edge.node) ?? [];
  const profileId = currentUser.profile?.id;

  const suggestionsById = new Map<string, {
    id: string;
    suggestionKey: (typeof loadedSuggestions)[number] | null;
    localSuggestion: StartedSuggestion | null;
  }>();
  for (const suggestion of loadedSuggestions) {
    suggestionsById.set(suggestion.id, {
      id: suggestion.id,
      suggestionKey: suggestion,
      localSuggestion: null,
    });
  }
  for (const suggestion of startedSuggestions) {
    if (suggestionsById.has(suggestion.id)) continue;
    suggestionsById.set(suggestion.id, {
      id: suggestion.id,
      suggestionKey: null,
      localSuggestion: suggestion,
    });
  }

  const {
    tops,
    previewMap,
    narrow,
    selectionTop,
    positioned: layoutPositioned,
    suggestionCards: layoutSuggestionCards,
    floatingThreads: layoutFloatingThreads,
    floatingSuggestions: layoutFloatingSuggestions,
    bindCard,
    bumpLayout,
  } = useReviewMarginLayout({
    editor,
    collaborationHandle,
    canComment,
    draft,
    setDraft,
    threads,
    commentThreads: commentThreadsWithoutSuggestions(
      threads,
      new Set(suggestionsById.keys()),
    ),
    suggestions: [...suggestionsById.values()],
    activeKey: draft ? "draft" : activeSuggestionId ?? activeThreadId,
    threadSignature: threadSignature(threads),
    commentBodySignature: commentBodySignature(threads),
    suggestionSignature: [...suggestionsById.keys()].join("|"),
  });

  // Cards follow marks in the document, not the SQL page. Open rows with no
  // remaining redline stay out of the margin.
  const suggestionIds = new Set(suggestionsById.keys());
  const commentThreads = commentThreadsWithoutSuggestions(threads, suggestionIds);
  const openThreadSignature = threadSignature(threads);
  const threadsBySuggestionId = new Map(
    threads
      .filter(thread => thread.suggestionId != null)
      .map(thread => [thread.suggestionId!, thread] as const),
  );

  useEffect(() => {
    const next = commentThreads.map(thread => ({
      id: thread.id,
      anchorStart: thread.anchorStart,
      anchorEnd: thread.anchorEnd,
      suggestionId: null,
    }));
    if (draft) {
      next.push({
        id: "draft",
        anchorStart: draft.anchorStart,
        anchorEnd: draft.anchorEnd,
        suggestionId: null,
      });
    }
    onThreads(next);
    // The signature is the stable view of the open comment list.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [openThreadSignature, draft]);

  useEffect(() => {
    refreshSuggestionFocus(editor, activeSuggestionId);
  }, [activeSuggestionId, editor]);

  const openComposer = () => {
    const { from, to, empty } = editor.state.selection;
    if (empty || !collaborationHandle) return;
    const automerge = collaborationHandle.doc();
    if (!automerge) return;
    const selection = presenceFromPmSelection(richEditorPresenceAdapter(), automerge, from, to);
    if (!selection) return;
    setDraft({
      quote: editor.state.doc.textBetween(from, to, "\n"),
      anchorStart: encodeDocumentAnchor(String(selection.anchor)),
      anchorEnd: encodeDocumentAnchor(String(selection.head)),
      from: Math.min(from, to),
    });
    onActiveThread(null);
    onActiveSuggestion(null);
  };

  const positioned = layoutPositioned.flatMap((item) => {
    const thread = commentThreads.find(candidate => candidate.id === item.key);
    if (!thread) return [];
    return [{ key: item.key, anchor: item.anchor, threadKey: thread }];
  });
  const suggestionCards = layoutSuggestionCards.flatMap((item) => {
    const suggestion = suggestionsById.get(item.key);
    if (!suggestion) return [];
    const attached = threadsBySuggestionId.get(item.key);
    return [{
      key: item.key,
      anchor: item.anchor,
      suggestionKey: suggestion.suggestionKey,
      localSuggestion: suggestion.localSuggestion,
      threadKey: attached ?? null,
      preview: item.preview,
    }];
  });
  const floatingThreads = layoutFloatingThreads.flatMap((thread) => {
    const match = commentThreads.find(candidate => candidate.id === thread.id);
    if (!match) return [];
    return [{ id: thread.id, threadKey: match }];
  });
  const floatingSuggestions = layoutFloatingSuggestions.flatMap((suggestion) => {
    const match = suggestionsById.get(suggestion.id);
    const preview = previewMap.get(suggestion.id);
    if (!match || !preview) return [];
    const attached = threadsBySuggestionId.get(suggestion.id);
    return [{
      id: suggestion.id,
      suggestionKey: match.suggestionKey,
      localSuggestion: match.localSuggestion,
      threadKey: attached ?? null,
      preview,
    }];
  });

  const bubble = selectionTop != null && !draft
    ? (
        <button
          type="button"
          className="absolute right-3 z-20 flex size-8 items-center justify-center rounded-full border border-sand-6 bg-sand-1 text-sand-12 shadow-2 hover:bg-sand-3"
          style={{ top: selectionTop }}
          aria-label={t("documentDescriptionPage.review.comment")}
          onMouseDown={event => event.preventDefault()}
          onClick={openComposer}
        >
          <ChatCircleIcon />
        </button>
      )
    : null;

  if (!reviewDocument) {
    throw new Error("invalid type for node");
  }

  return (
    <>
      <DocumentReviewSuggestionSession
        documentId={documentId}
        editor={editor}
        suggesting={suggesting}
        loadedSuggestions={loadedSuggestions}
        profileId={profileId}
        profileFullName={currentUser.profile?.fullName}
        activeSuggestionId={activeSuggestionId}
        onActiveSuggestion={onActiveSuggestion}
        onOwnReview={onOwnReview}
        onOwnReviewFailed={onOwnReviewFailed}
        onStartedSuggestions={onStartedSuggestions}
        removeStartedSuggestionRef={removeStartedSuggestionRef}
      />
      {toolbarHost
        ? (
            <DocumentReviewMarginHistory
              documentId={documentId}
              toolbarHost={toolbarHost}
              historyOpen={historyOpen}
              onHistoryOpenChange={setHistoryOpen}
              reviewVersion={reviewVersion}
              onOwnReview={onOwnReview}
              onOwnReviewFailed={onOwnReviewFailed}
            />
          )
        : null}
      {bubble}
      <DocumentReviewMarginList
        documentId={documentId}
        narrow={narrow}
        marginHost={marginHost}
        tops={tops}
        draft={draft}
        positioned={positioned}
        suggestionCards={suggestionCards}
        floatingThreads={floatingThreads}
        floatingSuggestions={floatingSuggestions}
        activeThreadId={activeThreadId}
        activeSuggestionId={activeSuggestionId}
        canUpdate={canUpdate}
        canComment={canComment}
        profileId={profileId ?? null}
        mentionSource={mentionSource}
        bindCard={bindCard}
        onActiveThread={onActiveThread}
        onActiveSuggestion={onActiveSuggestion}
        onDraftCancel={() => setDraft(null)}
        onDraftCreated={() => setDraft(null)}
        onSuggestionDecided={(id) => {
          removeStartedSuggestionRef.current?.(id);
          if (activeSuggestionId === id) onActiveSuggestion(null);
        }}
        onOwnReview={onOwnReview}
        onOwnReviewFailed={onOwnReviewFailed}
        onLayout={bumpLayout}
      />
    </>
  );
}

const reviewDocumentFragment = graphql`
  fragment DocumentReviewMargin_document on Document {
    id
    commentThreads(first: 200, filter: { resolved: false })
      @connection(key: "DocumentReviewMargin_commentThreads", filters: ["filter"]) {
      edges {
        node {
          id
          anchorQuote
          anchorStart
          anchorEnd
          resolvedAt
          suggestionId
          comments(first: 100) @connection(key: "DocumentReviewMargin_comments") {
            edges {
              node {
                id
                content
              }
            }
          }
          ...DocumentReviewMarginThread_thread
        }
      }
      pageInfo { hasNextPage endCursor }
    }
    suggestions(
      first: 200
      filter: { status: OPEN }
      orderBy: { field: CREATED_AT, direction: DESC }
    )
      @connection(key: "DocumentReviewMargin_suggestions", filters: ["filter", "orderBy"]) {
      edges {
        node {
          id
          status
          createdAt
          author { id fullName }
          ...DocumentReviewMarginSuggestion_suggestion
        }
      }
      pageInfo { hasNextPage endCursor }
    }
  }
`;

const reviewQuery = graphql`
  query DocumentReviewMarginQuery($documentId: ID!) {
    document: node(id: $documentId) {
      __typename
      ... on Document {
        ...DocumentReviewMargin_document
      }
    }
  }
`;
