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

import { ChatCircleIcon, CheckIcon, XIcon } from "@phosphor-icons/react";
import type { RepoCollaborationHandle } from "@probo/collaboration";
import { formatError } from "@probo/helpers";
import { automergeRemoteTransaction } from "@probo/ui/src/RichEditor/AutomergeSyncPlugin";
import { commentAnchorRange, encodeDocumentAnchor } from "@probo/ui/src/RichEditor/commentAnchors";
import { presenceFromPmSelection, richEditorPresenceAdapter } from "@probo/ui/src/RichEditor/repoPresence";
import {
  discardSuggestionId,
  reassignSuggestionId,
  refreshSuggestionFocus,
  type SuggestingController,
} from "@probo/ui/src/RichEditor/SuggestionExtension";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Popover } from "@probo/ui/src/v2/Popover/Popover";
import { PopoverPopup } from "@probo/ui/src/v2/Popover/PopoverPopup";
import { PopoverTrigger } from "@probo/ui/src/v2/Popover/PopoverTrigger";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import type { Editor } from "@tiptap/core";
import type { Transaction } from "@tiptap/pm/state";
import { type MutableRefObject, type RefObject, useContext, useEffect, useLayoutEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { useTranslation } from "react-i18next";
import { useFragment, useLazyLoadQuery, useRelayEnvironment } from "react-relay";
import { ConnectionHandler, fetchQuery, graphql, type RecordSourceSelectorProxy } from "relay-runtime";

import type { DocumentReviewMargin_acceptMutation } from "#/__generated__/core/DocumentReviewMargin_acceptMutation.graphql";
import type { DocumentReviewMargin_createThreadMutation } from "#/__generated__/core/DocumentReviewMargin_createThreadMutation.graphql";
import type { DocumentReviewMargin_deleteCommentMutation } from "#/__generated__/core/DocumentReviewMargin_deleteCommentMutation.graphql";
import type { DocumentReviewMargin_document$key } from "#/__generated__/core/DocumentReviewMargin_document.graphql";
import type { DocumentReviewMargin_historyQuery } from "#/__generated__/core/DocumentReviewMargin_historyQuery.graphql";
import type { DocumentReviewMargin_rejectMutation } from "#/__generated__/core/DocumentReviewMargin_rejectMutation.graphql";
import type { DocumentReviewMargin_reopenMutation } from "#/__generated__/core/DocumentReviewMargin_reopenMutation.graphql";
import type { DocumentReviewMargin_replyMutation } from "#/__generated__/core/DocumentReviewMargin_replyMutation.graphql";
import type { DocumentReviewMargin_resolveMutation } from "#/__generated__/core/DocumentReviewMargin_resolveMutation.graphql";
import type { DocumentReviewMargin_startSuggestionMutation } from "#/__generated__/core/DocumentReviewMargin_startSuggestionMutation.graphql";
import type { DocumentReviewMargin_updateCommentMutation } from "#/__generated__/core/DocumentReviewMargin_updateCommentMutation.graphql";
import type { DocumentReviewMarginQuery } from "#/__generated__/core/DocumentReviewMarginQuery.graphql";
import { useMutation } from "#/lib/relay/useMutation";
import { CurrentUser } from "#/providers/CurrentUser";

import { collectSuggestionAnchors, type SuggestionPreview } from "../../_lib/suggestionPreview";

import { DocumentReviewMarginCard } from "./DocumentReviewMarginCard";
import {
  DocumentReviewMarginComments,
  type DocumentReviewMarginThread,
} from "./DocumentReviewMarginComments";

export type CommentDraft = {
  quote: string;
  anchorStart: string;
  anchorEnd: string;
  from: number;
};

type HistoryThread = {
  id: string;
  anchorQuote: string;
  comments: {
    edges: ReadonlyArray<{ node: { content: string } }>;
  };
};

interface DocumentReviewMarginProps {
  documentId: string;
  editor: Editor;
  collaborationHandle: RepoCollaborationHandle;
  canUpdate: boolean;
  canComment: boolean;
  mode: "viewing" | "editing" | "suggesting";
  suggesting: MutableRefObject<SuggestingController>;
  activeThreadId: string | null;
  activeSuggestionId: string | null;
  reviewVersion: number;
  toolbarHost: RefObject<HTMLDivElement | null>;
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

const mutationFeedback = { errorToast: false };

export function DocumentReviewMargin(props: DocumentReviewMarginProps) {
  const {
    documentId,
    editor,
    collaborationHandle,
    canUpdate,
    canComment,
    mode,
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
  const toast = useToast();
  const currentUser = useContext(CurrentUser);
  const environment = useRelayEnvironment();
  const queryData = useLazyLoadQuery<DocumentReviewMarginQuery>(
    reviewQuery,
    { documentId: documentId },
    { fetchPolicy: "store-and-network" },
  );
  const reviewDocument = useFragment<DocumentReviewMargin_document$key>(
    reviewDocumentFragment,
    queryData.document?.__typename === "Document" ? queryData.document : null,
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
  const [createThread, creatingThread] = useMutation<DocumentReviewMargin_createThreadMutation>(
    createThreadMutation,
    mutationFeedback,
  );
  const [reply, replying] = useMutation<DocumentReviewMargin_replyMutation>(
    replyMutation,
    mutationFeedback,
  );
  const [updateComment] = useMutation<DocumentReviewMargin_updateCommentMutation>(
    updateCommentMutation,
    mutationFeedback,
  );
  const [deleteComment] = useMutation<DocumentReviewMargin_deleteCommentMutation>(
    deleteCommentMutation,
    mutationFeedback,
  );
  const [resolveThread] = useMutation<DocumentReviewMargin_resolveMutation>(
    resolveMutation,
    mutationFeedback,
  );
  const [reopenThread] = useMutation<DocumentReviewMargin_reopenMutation>(
    reopenMutation,
    mutationFeedback,
  );
  const [startSuggestion] = useMutation<DocumentReviewMargin_startSuggestionMutation>(
    startSuggestionMutation,
    mutationFeedback,
  );
  const [acceptSuggestion] = useMutation<DocumentReviewMargin_acceptMutation>(
    acceptMutation,
    mutationFeedback,
  );
  const [rejectSuggestion] = useMutation<DocumentReviewMargin_rejectMutation>(
    rejectMutation,
    mutationFeedback,
  );
  const [replyText, setReplyText] = useState("");
  const [draftText, setDraftText] = useState("");
  const [draft, setDraft] = useState<CommentDraft | null>(null);
  const [selectionTop, setSelectionTop] = useState<number | null>(null);
  const [selectionTick, setSelectionTick] = useState(0);
  const [layoutTick, setLayoutTick] = useState(0);
  const [narrow, setNarrow] = useState(() => window.matchMedia("(max-width: 1023px)").matches);
  const [toolbarNode, setToolbarNode] = useState<HTMLElement | null>(null);
  const [tops, setTops] = useState<Record<string, number>>({});
  const [startedSuggestions, setStartedSuggestions] = useState<Array<{
    id: string;
    author: { fullName: string | null | undefined };
  }>>([]);
  const [historyOpen, setHistoryOpen] = useState(false);
  const [resolvedThreads, setResolvedThreads] = useState<HistoryThread[]>([]);
  const cardRefs = useRef(new Map<string, HTMLDivElement>());
  const cardHeights = useRef(new Map<string, number>());
  const commentPositions = useRef(new Map<string, number>());
  const suggestionPreviews = useRef(new Map<string, SuggestionPreview>());
  const threadsRef = useRef<Array<{
    id: string;
    anchorStart: string;
    anchorEnd: string;
    anchorQuote?: string | null;
  }>>([]);
  const ownedSuggestionIds = useRef(new Set<string>());

  const threads = reviewDocument?.commentThreads.edges.map(edge => edge.node) ?? [];
  const loadedSuggestions = reviewDocument?.suggestions.edges.map(edge => edge.node) ?? [];
  const profileId = currentUser.profile?.id;
  for (const suggestion of loadedSuggestions) {
    if (profileId && suggestion.author.id === profileId) {
      ownedSuggestionIds.current.add(suggestion.id);
    }
  }
  for (const suggestion of startedSuggestions) {
    ownedSuggestionIds.current.add(suggestion.id);
  }
  suggesting.current.ownsSuggestion = id => ownedSuggestionIds.current.has(id);
  const suggestionsById = new Map<string, (typeof loadedSuggestions)[number] | {
    id: string;
    status: "OPEN";
    createdAt: string;
    author: { id: string; fullName: string | null | undefined };
  }>();
  for (const suggestion of loadedSuggestions) {
    suggestionsById.set(suggestion.id, suggestion);
  }
  for (const suggestion of startedSuggestions) {
    if (suggestionsById.has(suggestion.id)) continue;
    suggestionsById.set(suggestion.id, {
      id: suggestion.id,
      status: "OPEN",
      createdAt: "",
      author: { id: profileId ?? "", fullName: suggestion.author.fullName },
    });
  }
  // Cards follow marks in the document, not the SQL page. Open rows with no
  // remaining redline stay out of the margin.
  const suggestions = [...suggestionPreviews.current.keys()]
    .map(id => suggestionsById.get(id))
    .filter((suggestion): suggestion is NonNullable<typeof suggestion> => suggestion != null);
  const suggestionIds = new Set(suggestionsById.keys());
  const commentThreads = threads.filter(thread =>
    thread.suggestionId == null || !suggestionIds.has(thread.suggestionId),
  );
  threadsRef.current = threads;

  const threadSignature = threads
    .map(thread => `${thread.id}:${thread.suggestionId ?? ""}:${thread.anchorStart}:${thread.anchorEnd}`)
    .join("|");
  const commentBodySignature = threads
    .map(thread => `${thread.id}:${thread.comments.edges.map(edge => edge.node.content).join("\n")}`)
    .join("|");

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
  }, [threadSignature, draft]);

  useEffect(() => {
    suggesting.current.onNeedId = () => {
      const controller = suggesting.current;
      if (controller.suggestionId) return Promise.resolve(controller.suggestionId);

      const localId = `local:${crypto.randomUUID()}`;
      controller.suggestionId = localId;
      ownedSuggestionIds.current.add(localId);
      const author = { fullName: currentUser.profile?.fullName };
      setStartedSuggestions(current => current.some(suggestion => suggestion.id === localId)
        ? current
        : [...current, { id: localId, author }]);

      const request = new Promise<string | null>((resolve) => {
        onOwnReview();
        void startSuggestion({
          variables: { input: { documentId: documentId } },
          updater: store => insertSuggestionEdge(store, documentId),
          onCompleted: (response, errors) => {
            if (controller.request === request) controller.request = null;
            if (errors?.length) {
              onOwnReviewFailed();
              if (controller.suggestionId === localId) controller.suggestionId = null;
              ownedSuggestionIds.current.delete(localId);
              discardSuggestionId(editor, localId);
              setStartedSuggestions(current => current.filter(suggestion => suggestion.id !== localId));
              toast.add({ title: t("documentDescriptionPage.errors.title"), description: formatError(t("documentDescriptionPage.review.suggestionFailed"), errors), type: "error" });
              resolve(null);
              return;
            }
            const id = response.startDocumentSuggestion.documentSuggestionEdge.node.id;
            ownedSuggestionIds.current.add(id);
            reassignSuggestionId(editor, localId, id);
            if (controller.suggestionId === localId) controller.suggestionId = id;
            setStartedSuggestions(current => current.map(suggestion =>
              suggestion.id === localId ? { id, author } : suggestion,
            ));
            if (activeSuggestionId === localId) onActiveSuggestion(id);
            resolve(id);
          },
          onError: (error) => {
            onOwnReviewFailed();
            if (controller.request === request) controller.request = null;
            if (controller.suggestionId === localId) controller.suggestionId = null;
            ownedSuggestionIds.current.delete(localId);
            discardSuggestionId(editor, localId);
            setStartedSuggestions(current => current.filter(suggestion => suggestion.id !== localId));
            toast.add({ title: t("documentDescriptionPage.errors.title"), description: error.message, type: "error" });
            resolve(null);
          },
        });
      });
      controller.request = request;
      return request;
    };
    suggesting.current.onBurstFinished = () => {
      suggesting.current.suggestionId = null;
    };
  }, [
    activeSuggestionId,
    currentUser.profile?.fullName,
    documentId,
    editor,
    onActiveSuggestion,
    onOwnReview,
    onOwnReviewFailed,
    startSuggestion,
    suggesting,
    t,
    toast,
  ]);

  useEffect(() => {
    if (mode === "editing") {
      suggesting.current.suggestionId = null;
    }
  }, [mode, suggesting]);

  useEffect(() => {
    refreshSuggestionFocus(editor, activeSuggestionId);
  }, [activeSuggestionId, editor]);

  useEffect(() => {
    if (activeThreadId || activeSuggestionId) setDraft(null);
  }, [activeSuggestionId, activeThreadId]);

  useEffect(() => {
    setToolbarNode(toolbarHost.current);
  }, [toolbarHost]);

  useEffect(() => {
    if (!historyOpen) return;
    const subscription = fetchQuery<DocumentReviewMargin_historyQuery>(
      environment,
      historyQuery,
      { documentId: documentId },
      { fetchPolicy: "network-only" },
    ).subscribe({
      next: (data) => {
        const document = data.document?.__typename === "Document" ? data.document : null;
        setResolvedThreads((document?.commentThreads.edges.map(edge => edge.node) ?? []));
      },
    });
    return () => subscription.unsubscribe();
  }, [environment, historyOpen, documentId, reviewVersion]);

  useLayoutEffect(() => {
    let cardFrame = 0;
    let selectionFrame = 0;
    const scheduleCards = () => {
      if (cardFrame !== 0) return;
      cardFrame = window.requestAnimationFrame(() => {
        cardFrame = 0;
        setLayoutTick(value => value + 1);
      });
    };
    const scheduleSelection = () => {
      if (selectionFrame !== 0) return;
      selectionFrame = window.requestAnimationFrame(() => {
        selectionFrame = 0;
        setSelectionTick(value => value + 1);
      });
    };
    const resolveCommentPositions = () => {
      const next = new Map<string, number>();
      for (const thread of threadsRef.current) {
        const from = threadPosition(collaborationHandle, editor.state.doc, thread);
        if (from != null) next.set(thread.id, from);
      }
      commentPositions.current = next;
    };
    const onTransaction = ({ transaction }: { transaction: Transaction }) => {
      if (!transaction.docChanged) {
        if (transaction.selectionSet) scheduleSelection();
        return;
      }
      const unresolved = threadsRef.current.some(thread => !commentPositions.current.has(thread.id));
      if (transaction.getMeta(automergeRemoteTransaction) || unresolved) {
        resolveCommentPositions();
      } else {
        const next = new Map<string, number>();
        for (const [id, from] of commentPositions.current) {
          next.set(id, transaction.mapping.map(from));
        }
        commentPositions.current = next;
        setDraft(current => current == null
          ? current
          : { ...current, from: transaction.mapping.map(current.from) });
      }
      suggestionPreviews.current = collectSuggestionAnchors(editor.state.doc);
      scheduleCards();
    };
    editor.on("transaction", onTransaction);
    resolveCommentPositions();
    suggestionPreviews.current = collectSuggestionAnchors(editor.state.doc);
    scheduleCards();
    const scroll = editor.view.dom.closest("[data-review-scroll]");
    const observer = new ResizeObserver(scheduleCards);
    if (scroll) {
      observer.observe(scroll);
      // Cards are portaled into a non-scrolling sibling <aside>; the editor
      // root scrolls on its own, so reposition them on every scroll frame.
      scroll.addEventListener("scroll", scheduleCards, { passive: true });
    }
    const media = window.matchMedia("(max-width: 1023px)");
    const onMedia = () => {
      setNarrow(media.matches);
      scheduleCards();
    };
    media.addEventListener("change", onMedia);
    return () => {
      if (cardFrame !== 0) window.cancelAnimationFrame(cardFrame);
      if (selectionFrame !== 0) window.cancelAnimationFrame(selectionFrame);
      editor.off("transaction", onTransaction);
      scroll?.removeEventListener("scroll", scheduleCards);
      observer.disconnect();
      media.removeEventListener("change", onMedia);
    };
  }, [collaborationHandle, editor]);

  useEffect(() => {
    const next = new Map<string, number>();
    for (const thread of threads) {
      const from = threadPosition(collaborationHandle, editor.state.doc, thread);
      if (from != null) next.set(thread.id, from);
    }
    commentPositions.current = next;
    suggestionPreviews.current = collectSuggestionAnchors(editor.state.doc);
    setLayoutTick(value => value + 1);
    // threadSignature is the open comment list. Cursor resolution waits for it
    // or a remote document change.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [threadSignature, editor]);

  useEffect(() => {
    const { from, to, empty } = editor.state.selection;
    if (empty || !canComment || draft) {
      setSelectionTop(null);
      return;
    }
    setSelectionTop(editorContentTop(editor, from < to ? from : to));
  }, [draft, selectionTick, canComment, editor]);

  const reportError = (error: Error) => {
    toast.add({
      title: t("documentDescriptionPage.errors.title"),
      description: error.message,
      type: "error",
    });
  };

  const activeKey = draft
    ? "draft"
    : activeSuggestionId ?? activeThreadId;
  const suggestionSignature = suggestions.map(suggestion => suggestion.id).join("|");
  const suggestionAnchor = (id: string) => {
    const preview = suggestionPreviews.current.get(id);
    if (preview) return preview.from;
    const thread = threads.find(item => item.suggestionId === id);
    if (!thread) return null;
    return commentPositions.current.get(thread.id) ?? null;
  };

  const placedTopCount = Object.keys(tops).length;
  useLayoutEffect(() => {
    suggestionPreviews.current = collectSuggestionAnchors(editor.state.doc);
    for (const thread of threads) {
      if (commentPositions.current.has(thread.id)) continue;
      const from = threadPosition(collaborationHandle, editor.state.doc, thread);
      if (from != null) commentPositions.current.set(thread.id, from);
    }
    const cards: Array<{ key: string; anchor: number }> = [];
    for (const thread of commentThreads) {
      const from = commentPositions.current.get(thread.id);
      if (from == null) continue;
      const top = editorContentTop(editor, from);
      if (top != null) cards.push({ key: thread.id, anchor: top });
    }
    for (const suggestion of suggestions) {
      const from = suggestionAnchor(suggestion.id);
      if (from == null) continue;
      const top = editorContentTop(editor, from);
      if (top != null) cards.push({ key: suggestion.id, anchor: top });
    }
    if (draft) {
      const top = editorContentTop(editor, draft.from);
      if (top != null) cards.push({ key: "draft", anchor: top });
    }
    const measured = cards.map(card => ({
      key: card.key,
      anchor: card.anchor,
      height: cardRefs.current.get(card.key)?.offsetHeight ?? 120,
    }));
    const next = placeCardTops(measured, activeKey);
    setTops((current) => {
      const keys = Object.keys(next);
      if (keys.length === Object.keys(current).length && keys.every(key => current[key] === next[key])) {
        return current;
      }
      return next;
    });
    // Positions live in refs. These signals are a document change, a new card,
    // a reply that makes a card taller, or the first paint that mounts the
    // cards so their heights can be measured.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [
    activeKey,
    commentBodySignature,
    draft?.from,
    layoutTick,
    narrow,
    suggestionSignature,
    threadSignature,
    placedTopCount,
  ]);

  const positioned = commentThreads.flatMap((thread) => {
    const top = tops[thread.id];
    if (top == null) return [];
    return [{ key: thread.id, anchor: top, thread }];
  });
  const suggestionCards = suggestions.flatMap((suggestion) => {
    const top = tops[suggestion.id];
    const preview = suggestionPreviews.current.get(suggestion.id);
    if (top == null || !preview) return [];
    return [{ key: suggestion.id, anchor: top, suggestion, preview }];
  });
  const placedCardSignature = [
    ...positioned.map(item => item.key),
    ...suggestionCards.map(item => item.key),
    draft && tops.draft != null ? "draft" : "",
  ].join("|");
  useEffect(() => {
    const observer = new ResizeObserver((entries) => {
      let changed = false;
      for (const entry of entries) {
        const node = entry.target;
        if (!(node instanceof HTMLElement)) continue;
        const key = node.dataset.reviewCard;
        if (!key) continue;
        const height = node.offsetHeight;
        if (cardHeights.current.get(key) === height) continue;
        cardHeights.current.set(key, height);
        changed = true;
      }
      if (changed) setLayoutTick(value => value + 1);
    });
    for (const node of cardRefs.current.values()) observer.observe(node);
    return () => observer.disconnect();
  }, [placedCardSignature]);

  const floatingThreads = commentThreads.filter(thread => tops[thread.id] == null);
  const floatingSuggestions = suggestions.filter(suggestion => (
    suggestionPreviews.current.has(suggestion.id) && tops[suggestion.id] == null
  ));

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

  const bindCard = (key: string) => (node: HTMLDivElement | null) => {
    if (node) cardRefs.current.set(key, node);
    else cardRefs.current.delete(key);
  };

  const commentCard = (thread: DocumentReviewMarginThread, active: boolean) => (
    <DocumentReviewMarginCard
      active={active}
      onActivate={() => {
        onActiveThread(thread.id);
        onActiveSuggestion(null);
      }}
    >
      <DocumentReviewMarginComments
        thread={thread}
        active={active}
        profileId={currentUser.profile?.id ?? null}
        canComment={canComment}
        replyText={active ? replyText : ""}
        onReplyText={setReplyText}
        replying={replying}
        onReply={(content) => {
          onOwnReview();
          void reply({
            variables: { input: { threadId: thread.id, content } },
            updater: store => insertCommentEdge(store, thread.id),
            onCompleted: (_response, errors) => {
              if (errors?.length) {
                onOwnReviewFailed();
                toast.add({ title: t("documentDescriptionPage.errors.title"), description: formatError(t("documentDescriptionPage.review.commentFailed"), errors), type: "error" });
                return;
              }
              setReplyText("");
            },
            onError: (error) => {
              onOwnReviewFailed();
              reportError(error);
            },
          });
        }}
        onUpdate={(commentId, content) => {
          onOwnReview();
          void updateComment({
            variables: { input: { commentId, content } },
            onCompleted: (_response, errors) => {
              if (errors?.length) onOwnReviewFailed();
            },
            onError: (error) => {
              onOwnReviewFailed();
              reportError(error);
            },
          });
        }}
        onDelete={(commentId) => {
          onOwnReview();
          void deleteComment({
            variables: { input: { commentId } },
            updater: store => deleteCommentEdge(store, thread.id, commentId),
            onCompleted: (_response, errors) => {
              if (errors?.length) onOwnReviewFailed();
            },
            onError: (error) => {
              onOwnReviewFailed();
              reportError(error);
            },
          });
        }}
        onLayout={() => setLayoutTick(value => value + 1)}
        onResolve={() => {
          onOwnReview();
          void resolveThread({
            variables: { input: { threadId: thread.id } },
            updater: store => deleteOpenThread(store, documentId, thread.id),
            onCompleted: (_response, errors) => {
              if (errors?.length) {
                onOwnReviewFailed();
                return;
              }
              if (activeThreadId === thread.id) onActiveThread(null);
            },
            onError: (error) => {
              onOwnReviewFailed();
              reportError(error);
            },
          });
        }}
      />
    </DocumentReviewMarginCard>
  );

  const suggestionCard = (
    suggestion: { id: string; author: { fullName: string | null | undefined } },
    preview: Pick<SuggestionPreview, "parts">,
    active: boolean,
  ) => {
    const thread = threads.find(item => item.suggestionId === suggestion.id);
    return (
      <DocumentReviewMarginCard
        active={active}
        onActivate={() => {
          onActiveSuggestion(suggestion.id);
          onActiveThread(null);
        }}
      >
        <div className="flex items-start justify-between gap-2">
          <p className="text-2 font-medium text-sand-12">{suggestion.author.fullName}</p>
          {canUpdate && !suggestion.id.startsWith("local:") && (
            <div className="flex gap-1">
              <Button
                variant="ghost"
                color="green"
                size={1}
                aria-label={t("documentDescriptionPage.review.accept")}
                iconStart={<CheckIcon />}
                onClick={() => {
                  onOwnReview();
                  void acceptSuggestion({
                    variables: { input: { suggestionId: suggestion.id } },
                    updater: store => closeSuggestion(store, documentId, suggestion.id),
                    onCompleted: (_response, errors) => {
                      if (errors?.length) {
                        onOwnReviewFailed();
                        return;
                      }
                      setStartedSuggestions(current => current.filter(item => item.id !== suggestion.id));
                      if (activeSuggestionId === suggestion.id) onActiveSuggestion(null);
                    },
                    onError: (error) => {
                      onOwnReviewFailed();
                      reportError(error);
                    },
                  });
                }}
              />
              <Button
                variant="ghost"
                color="red"
                size={1}
                aria-label={t("documentDescriptionPage.review.reject")}
                iconStart={<XIcon />}
                onClick={() => {
                  onOwnReview();
                  void rejectSuggestion({
                    variables: { input: { suggestionId: suggestion.id } },
                    updater: store => closeSuggestion(store, documentId, suggestion.id),
                    onCompleted: (_response, errors) => {
                      if (errors?.length) {
                        onOwnReviewFailed();
                        return;
                      }
                      setStartedSuggestions(current => current.filter(item => item.id !== suggestion.id));
                      if (activeSuggestionId === suggestion.id) onActiveSuggestion(null);
                    },
                    onError: (error) => {
                      onOwnReviewFailed();
                      reportError(error);
                    },
                  });
                }}
              />
            </div>
          )}
        </div>
        <p className="whitespace-pre-wrap text-2 text-sand-12">
          {preview.parts.map((part, index) =>
            part.kind === "deleted"
              ? (
                  <span key={index} className="text-red-11 line-through">{part.text}</span>
                )
              : part.kind === "inserted"
                ? (
                    <span key={index} className="text-green-11">{part.text}</span>
                  )
                : (
                    <span key={index}>{part.text}</span>
                  ),
          )}
          {preview.parts.length === 0 && thread?.anchorQuote}
        </p>
        <DocumentReviewMarginComments
          thread={thread}
          active={active}
          profileId={currentUser.profile?.id ?? null}
          canComment={canComment && !suggestion.id.startsWith("local:")}
          replyText={active ? replyText : ""}
          onReplyText={setReplyText}
          replying={replying}
          onLayout={() => setLayoutTick(value => value + 1)}
          onReply={(content) => {
            if (thread) {
              onOwnReview();
              void reply({
                variables: { input: { threadId: thread.id, content } },
                updater: store => insertCommentEdge(store, thread.id),
                onCompleted: (_response, errors) => {
                  if (errors?.length) {
                    onOwnReviewFailed();
                    return;
                  }
                  setReplyText("");
                },
                onError: (error) => {
                  onOwnReviewFailed();
                  reportError(error);
                },
              });
              return;
            }
            onOwnReview();
            void createThread({
              variables: { input: { documentId: documentId, suggestionId: suggestion.id, content } },
              updater: store => insertThreadEdge(store, documentId),
              onCompleted: (_response, errors) => {
                if (errors?.length) {
                  onOwnReviewFailed();
                  return;
                }
                setReplyText("");
              },
              onError: (error) => {
                onOwnReviewFailed();
                reportError(error);
              },
            });
          }}
          onUpdate={(commentId, content) => {
            onOwnReview();
            void updateComment({
              variables: { input: { commentId, content } },
              onCompleted: (_response, errors) => {
                if (errors?.length) onOwnReviewFailed();
              },
              onError: (error) => {
                onOwnReviewFailed();
                reportError(error);
              },
            });
          }}
          onDelete={(commentId) => {
            if (!thread) return;
            onOwnReview();
            void deleteComment({
              variables: { input: { commentId } },
              updater: store => deleteCommentEdge(store, thread.id, commentId),
              onCompleted: (_response, errors) => {
                if (errors?.length) onOwnReviewFailed();
              },
              onError: (error) => {
                onOwnReviewFailed();
                reportError(error);
              },
            });
          }}
        />
      </DocumentReviewMarginCard>
    );
  };

  const composer = (
    <DocumentReviewMarginCard active onActivate={() => {}}>
      <form
        className="flex flex-col gap-2"
        onSubmit={(event) => {
          event.preventDefault();
          if (!draft || draftText.trim() === "") return;
          onOwnReview();
          void createThread({
            variables: {
              input: {
                documentId: documentId,
                content: draftText,
                quote: draft.quote,
                anchorStart: draft.anchorStart,
                anchorEnd: draft.anchorEnd,
              },
            },
            updater: store => insertThreadEdge(store, documentId),
            onCompleted: (_response, errors) => {
              if (errors?.length) {
                onOwnReviewFailed();
                toast.add({ title: t("documentDescriptionPage.errors.title"), description: formatError(t("documentDescriptionPage.review.commentFailed"), errors), type: "error" });
                return;
              }
              setDraftText("");
              setDraft(null);
            },
            onError: (error) => {
              onOwnReviewFailed();
              reportError(error);
            },
          });
        }}
      >
        <textarea
          autoFocus
          className="min-h-16 w-full rounded-2 border border-sand-7 bg-sand-1 p-2 text-2 text-sand-12 outline-none focus:border-sand-8"
          value={draftText}
          onChange={event => setDraftText(event.target.value)}
          placeholder={t("documentDescriptionPage.review.commentPlaceholder")}
        />
        <div className="flex justify-end gap-2">
          <Button variant="ghost" color="neutral" size={1} onClick={() => setDraft(null)}>
            {t("documentDescriptionPage.review.cancel")}
          </Button>
          <Button variant="solid" color="indigo" size={1} type="submit" disabled={creatingThread}>
            {t("documentDescriptionPage.review.comment")}
          </Button>
        </div>
      </form>
    </DocumentReviewMarginCard>
  );

  const history = toolbarNode
    ? createPortal(
        <Popover open={historyOpen} onOpenChange={setHistoryOpen}>
          <PopoverTrigger
            render={(
              <Button
                variant="ghost"
                color="neutral"
                size={2}
                aria-label={t("documentDescriptionPage.review.history")}
                iconStart={<ChatCircleIcon />}
              />
            )}
          />
          <PopoverPopup align="end" side="bottom">
            <div className="flex w-72 flex-col gap-3 p-3">
              {resolvedThreads.length === 0 && (
                <p className="text-2 text-sand-11">{t("documentDescriptionPage.review.resolvedEmpty")}</p>
              )}
              {resolvedThreads.map(thread => (
                <div key={thread.id} className="flex flex-col gap-1">
                  <p className="whitespace-pre-wrap text-2 text-sand-12">{thread.anchorQuote}</p>
                  <p className="text-1 text-sand-11">{thread.comments.edges[0]?.node.content}</p>
                  <Button
                    variant="ghost"
                    color="neutral"
                    size={1}
                    onClick={() => {
                      onOwnReview();
                      void reopenThread({
                        variables: { input: { threadId: thread.id } },
                        updater: store => reopenThreadEdge(store, documentId, thread.id),
                        onCompleted: (_response, errors) => {
                          if (errors?.length) {
                            onOwnReviewFailed();
                            return;
                          }
                          setResolvedThreads(current => current.filter(item => item.id !== thread.id));
                        },
                        onError: (error) => {
                          onOwnReviewFailed();
                          reportError(error);
                        },
                      });
                    }}
                  >
                    {t("documentDescriptionPage.review.reopen")}
                  </Button>
                </div>
              ))}
            </div>
          </PopoverPopup>
        </Popover>,
        toolbarNode,
      )
    : null;

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

  const floating = (floatingThreads.length > 0 || floatingSuggestions.length > 0) && (
    <div className="absolute top-14 right-3 z-20 flex w-60 flex-col gap-2">
      {floatingSuggestions.map(suggestion => (
        <div key={suggestion.id}>
          {suggestionCard(
            suggestion,
            suggestionPreviews.current.get(suggestion.id) ?? { inserted: "", deleted: "", parts: [] },
            activeSuggestionId === suggestion.id && !draft,
          )}
        </div>
      ))}
      {floatingThreads.map(thread => (
        <div key={thread.id}>
          {commentCard(thread, activeThreadId === thread.id && !draft)}
        </div>
      ))}
    </div>
  );

  if (narrow) {
    const dockedSuggestion = suggestionCards.find(item => item.key === activeSuggestionId);
    const dockedThread = positioned.find(item => item.key === activeThreadId);
    const docked = draft
      ? composer
      : dockedSuggestion
        ? suggestionCard(dockedSuggestion.suggestion, dockedSuggestion.preview, true)
        : dockedThread
          ? commentCard(dockedThread.thread, true)
          : null;
    return (
      <>
        {history}
        {bubble}
        {docked
          ? <div className="sticky top-2 z-20 mx-3 mb-3">{docked}</div>
          : floating}
      </>
    );
  }

  const desktopCards = (
    <>
      {floating}
      {positioned.map(item => (
        <div
          key={item.key}
          ref={bindCard(item.key)}
          data-review-card={item.key}
          className="absolute right-3 z-10 w-60"
          style={{ top: tops[item.key] ?? item.anchor }}
        >
          {commentCard(item.thread, activeThreadId === item.key && !draft)}
        </div>
      ))}
      {suggestionCards.map(item => (
        <div
          key={item.key}
          ref={bindCard(item.key)}
          data-review-card={item.key}
          className="absolute right-3 z-10 w-60"
          style={{ top: tops[item.key] ?? item.anchor }}
        >
          {suggestionCard(item.suggestion, item.preview, activeSuggestionId === item.key && !draft)}
        </div>
      ))}
      {draft && tops.draft != null && (
        <div
          ref={bindCard("draft")}
          data-review-card="draft"
          className="absolute right-3 z-10 w-60"
          style={{ top: tops.draft }}
        >
          {composer}
        </div>
      )}
    </>
  );

  return (
    <>
      {history}
      {bubble}
      {marginHost
        ? createPortal(desktopCards, marginHost)
        : desktopCards}
    </>
  );
}

function threadPosition(
  handle: RepoCollaborationHandle,
  doc: Editor["state"]["doc"],
  thread: { anchorStart: string; anchorEnd: string; anchorQuote?: string | null },
): number | null {
  const range = commentAnchorRange(handle, doc, thread);
  if (range) return range.from;
  return quotePosition(doc, thread.anchorQuote ?? "");
}

// A suggestion comment stores a server cursor the editor cannot read back.
// The quoted text is still in the document, so the card can sit beside it.
function quotePosition(doc: Editor["state"]["doc"], quote: string): number | null {
  const needle = quote.replaceAll("\r\n", "\n").trim();
  if (needle === "") return null;
  const lines = needle.split("\n").map(line => line.trim()).filter(line => line !== "");
  for (const line of lines) {
    let found: number | null = null;
    doc.descendants((node, pos) => {
      if (found != null || !node.isText || !node.text) return;
      const at = node.text.indexOf(line);
      if (at >= 0) found = pos + at;
    });
    if (found != null) return found;
  }
  return null;
}

function editorContentTop(editor: Editor, pos: number) {
  // Prefer the shared editor+aside row so cards in the margin column and the
  // text share one coordinate space, regardless of which ancestor scrolls.
  const layout = editor.view.dom.closest("[data-review-layout]");
  const scroll = layout instanceof HTMLElement
    ? layout
    : editor.view.dom.closest("[data-review-scroll]");
  if (!(scroll instanceof HTMLElement)) return null;
  const clamped = Math.max(0, Math.min(pos, editor.state.doc.content.size));
  try {
    const coords = editor.view.coordsAtPos(clamped);
    const rect = scroll.getBoundingClientRect();
    return coords.top - rect.top + scroll.scrollTop;
  } catch {
    return null;
  }
}

const openThreadConnection = "DocumentReviewMargin_commentThreads";
const openSuggestionConnection = "DocumentReviewMargin_suggestions";
const commentConnection = "DocumentReviewMargin_comments";

function openThreads(store: RecordSourceSelectorProxy, documentId: string) {
  return store.get(ConnectionHandler.getConnectionID(
    documentId,
    openThreadConnection,
    { filter: { resolved: false } },
  ));
}

function insertThreadEdge(store: RecordSourceSelectorProxy, documentId: string) {
  const edge = store.getRootField("createDocumentCommentThread")?.getLinkedRecord("documentCommentThreadEdge");
  const connection = openThreads(store, documentId);
  if (edge && connection) ConnectionHandler.insertEdgeAfter(connection, edge);
}

function insertCommentEdge(store: RecordSourceSelectorProxy, threadId: string) {
  const edge = store.getRootField("createDocumentComment")?.getLinkedRecord("documentCommentEdge");
  const connection = store.get(ConnectionHandler.getConnectionID(threadId, commentConnection));
  if (edge && connection) ConnectionHandler.insertEdgeAfter(connection, edge);
}

function deleteCommentEdge(store: RecordSourceSelectorProxy, threadId: string, commentId: string) {
  const connection = store.get(ConnectionHandler.getConnectionID(threadId, commentConnection));
  if (connection) ConnectionHandler.deleteNode(connection, commentId);
}

function deleteOpenThread(store: RecordSourceSelectorProxy, documentId: string, threadId: string) {
  const connection = openThreads(store, documentId);
  if (connection) ConnectionHandler.deleteNode(connection, threadId);
}

function insertSuggestionEdge(store: RecordSourceSelectorProxy, documentId: string) {
  const edge = store.getRootField("startDocumentSuggestion")?.getLinkedRecord("documentSuggestionEdge");
  const connection = store.get(ConnectionHandler.getConnectionID(
    documentId,
    openSuggestionConnection,
    { filter: { status: "OPEN" }, orderBy: { direction: "DESC", field: "CREATED_AT" } },
  ));
  if (edge && connection) ConnectionHandler.insertEdgeBefore(connection, edge);
}

function closeSuggestion(store: RecordSourceSelectorProxy, documentId: string, suggestionId: string) {
  const suggestions = store.get(ConnectionHandler.getConnectionID(
    documentId,
    openSuggestionConnection,
    { filter: { status: "OPEN" }, orderBy: { direction: "DESC", field: "CREATED_AT" } },
  ));
  if (suggestions) ConnectionHandler.deleteNode(suggestions, suggestionId);
  const threads = openThreads(store, documentId);
  for (const edge of threads?.getLinkedRecords("edges") ?? []) {
    const node = edge?.getLinkedRecord("node");
    if (node?.getValue("suggestionId") === suggestionId && threads) {
      ConnectionHandler.deleteNode(threads, node.getDataID());
    }
  }
}

function reopenThreadEdge(store: RecordSourceSelectorProxy, documentId: string, threadId: string) {
  const connection = openThreads(store, documentId);
  const node = store.get(threadId);
  if (!connection || !node) return;
  const edge = ConnectionHandler.createEdge(store, connection, node, "DocumentCommentThreadEdge");
  ConnectionHandler.insertEdgeAfter(connection, edge);
}

function placeCardTops(
  cards: Array<{ key: string; anchor: number; height: number }>,
  activeKey: string | null,
) {
  const gap = 8;
  const sorted = [...cards].sort((left, right) => left.anchor - right.anchor || left.key.localeCompare(right.key));
  const tops = new Map<string, number>();
  let floor = 0;
  for (const card of sorted) {
    const top = Math.max(card.anchor, floor);
    tops.set(card.key, top);
    floor = top + card.height + gap;
  }

  const active = activeKey ? sorted.find(card => card.key === activeKey) : undefined;
  const current = activeKey ? tops.get(activeKey) : undefined;
  if (active && current != null && current > active.anchor) {
    const blocked = sorted.some((card) => {
      if (card.key === active.key) return false;
      const top = tops.get(card.key) ?? card.anchor;
      return top < active.anchor + active.height && top + card.height > active.anchor;
    });
    if (!blocked) {
      tops.set(active.key, active.anchor);
      let nextFloor = active.anchor + active.height + gap;
      for (const card of sorted) {
        if (card.key === active.key || card.anchor + card.height <= active.anchor) continue;
        const top = Math.max(tops.get(card.key) ?? card.anchor, nextFloor);
        tops.set(card.key, top);
        nextFloor = top + card.height + gap;
      }
    }
  }

  return Object.fromEntries(tops);
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
                createdAt
                author { id fullName }
              }
            }
            pageInfo { hasNextPage endCursor }
          }
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
        }
      }
      pageInfo { hasNextPage endCursor }
    }
  }
`;

export const reviewThreadFragment = graphql`
  fragment DocumentReviewMargin_thread on DocumentCommentThread {
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
          createdAt
          author { id fullName }
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

const historyQuery = graphql`
  query DocumentReviewMargin_historyQuery($documentId: ID!) {
    document: node(id: $documentId) {
      __typename
      ... on Document {
        commentThreads(first: 50, filter: { resolved: true }) {
          edges {
            node {
              id
              anchorQuote
              comments(first: 1) {
                edges { node { content } }
              }
            }
          }
        }
      }
    }
  }
`;

const createThreadMutation = graphql`
  mutation DocumentReviewMargin_createThreadMutation($input: CreateDocumentCommentThreadInput!) {
    createDocumentCommentThread(input: $input) {
      documentCommentThreadEdge {
        cursor
        node { ...DocumentReviewMargin_thread }
      }
    }
  }
`;

const replyMutation = graphql`
  mutation DocumentReviewMargin_replyMutation($input: CreateDocumentCommentInput!) {
    createDocumentComment(input: $input) {
      documentCommentEdge {
        cursor
        node { id content createdAt author { id fullName } }
      }
    }
  }
`;

const updateCommentMutation = graphql`
  mutation DocumentReviewMargin_updateCommentMutation($input: UpdateDocumentCommentInput!) {
    updateDocumentComment(input: $input) {
      documentComment { id content }
    }
  }
`;

const deleteCommentMutation = graphql`
  mutation DocumentReviewMargin_deleteCommentMutation($input: DeleteDocumentCommentInput!) {
    deleteDocumentComment(input: $input) {
      deletedDocumentCommentId
    }
  }
`;

const resolveMutation = graphql`
  mutation DocumentReviewMargin_resolveMutation($input: ResolveDocumentCommentThreadInput!) {
    resolveDocumentCommentThread(input: $input) {
      documentCommentThread { id resolvedAt }
    }
  }
`;

const reopenMutation = graphql`
  mutation DocumentReviewMargin_reopenMutation($input: ReopenDocumentCommentThreadInput!) {
    reopenDocumentCommentThread(input: $input) {
      documentCommentThread { ...DocumentReviewMargin_thread }
    }
  }
`;

const startSuggestionMutation = graphql`
  mutation DocumentReviewMargin_startSuggestionMutation($input: StartDocumentSuggestionInput!) {
    startDocumentSuggestion(input: $input) {
      documentSuggestionEdge {
        cursor
        node { id status createdAt author { id fullName } }
      }
    }
  }
`;

const acceptMutation = graphql`
  mutation DocumentReviewMargin_acceptMutation($input: AcceptDocumentSuggestionInput!) {
    acceptDocumentSuggestion(input: $input) {
      documentSuggestion { id status }
    }
  }
`;

const rejectMutation = graphql`
  mutation DocumentReviewMargin_rejectMutation($input: RejectDocumentSuggestionInput!) {
    rejectDocumentSuggestion(input: $input) {
      documentSuggestion { id status }
    }
  }
`;
