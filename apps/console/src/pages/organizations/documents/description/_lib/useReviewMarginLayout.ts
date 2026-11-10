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
import { automergeRemoteTransaction } from "@probo/ui/src/RichEditor/AutomergeSyncPlugin";
import type { Editor } from "@tiptap/core";
import type { Transaction } from "@tiptap/pm/state";
import { useEffect, useLayoutEffect, useMemo, useRef, useState } from "react";

import { collectSuggestionAnchors, type SuggestionPreview } from "./suggestionPreview";
import {
  editorContentTop,
  placeCardTops,
  threadPosition,
  type ReviewMarginAnchorThread,
} from "./reviewMarginLayout";

export type CommentDraft = {
  quote: string;
  anchorStart: string;
  anchorEnd: string;
  from: number;
};

type UseReviewMarginLayoutArgs = {
  editor: Editor;
  collaborationHandle: RepoCollaborationHandle;
  canComment: boolean;
  draft: CommentDraft | null;
  setDraft: (value: CommentDraft | null | ((current: CommentDraft | null) => CommentDraft | null)) => void;
  threads: ReadonlyArray<ReviewMarginAnchorThread>;
  commentThreads: ReadonlyArray<ReviewMarginAnchorThread>;
  suggestions: ReadonlyArray<{ id: string }>;
  activeKey: string | null;
  threadSignature: string;
  commentBodySignature: string;
  suggestionSignature: string;
};

export function useReviewMarginLayout({
  editor,
  collaborationHandle,
  canComment,
  draft,
  setDraft,
  threads,
  commentThreads,
  suggestions,
  activeKey,
  threadSignature,
  commentBodySignature,
  suggestionSignature,
}: UseReviewMarginLayoutArgs) {
  const [selectionTick, setSelectionTick] = useState(0);
  const [layoutTick, setLayoutTick] = useState(0);
  const [narrow, setNarrow] = useState(() => window.matchMedia("(max-width: 1023px)").matches);
  const [tops, setTops] = useState<Record<string, number>>({});
  const [previewMap, setPreviewMap] = useState<Map<string, SuggestionPreview>>(() => new Map());
  const cardRefs = useRef(new Map<string, HTMLDivElement>());
  const cardHeights = useRef(new Map<string, number>());
  const commentPositions = useRef(new Map<string, number>());
  const suggestionPreviews = useRef(new Map<string, SuggestionPreview>());
  const threadsRef = useRef(threads);

  useLayoutEffect(() => {
    threadsRef.current = threads;
  });

  useLayoutEffect(() => {
    let cardFrame = 0;
    let selectionFrame = 0;
    const scheduleCards = () => {
      if (cardFrame !== 0) return;
      cardFrame = window.requestAnimationFrame(() => {
        cardFrame = 0;
        setPreviewMap(new Map(suggestionPreviews.current));
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
  }, [collaborationHandle, editor, setDraft]);

  useEffect(() => {
    const next = new Map<string, number>();
    for (const thread of threads) {
      const from = threadPosition(collaborationHandle, editor.state.doc, thread);
      if (from != null) next.set(thread.id, from);
    }
    commentPositions.current = next;
    suggestionPreviews.current = collectSuggestionAnchors(editor.state.doc);
    const frame = window.requestAnimationFrame(() => {
      setPreviewMap(new Map(suggestionPreviews.current));
      setLayoutTick(value => value + 1);
    });
    return () => window.cancelAnimationFrame(frame);
    // threadSignature is the open comment list. Cursor resolution waits for it
    // or a remote document change.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [threadSignature, editor]);

  const selectionTop = useMemo(() => {
    const { from, to, empty } = editor.state.selection;
    if (empty || !canComment || draft) return null;
    return editorContentTop(editor, from < to ? from : to);
  }, [canComment, draft, editor, selectionTick]);

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
    const preview = previewMap.get(suggestion.id);
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
    previewMap.has(suggestion.id) && tops[suggestion.id] == null
  ));

  const bindCard = (node: HTMLDivElement) => {
    const key = node.dataset.reviewCard;
    if (!key) return;
    const cards = cardRefs.current;
    cards.set(key, node);
    return () => {
      if (cards.get(key) === node) cards.delete(key);
    };
  };

  return {
    tops,
    previewMap,
    narrow,
    selectionTop,
    positioned,
    suggestionCards,
    floatingThreads,
    floatingSuggestions,
    bindCard,
    bumpLayout: () => setLayoutTick(value => value + 1),
  };
}
