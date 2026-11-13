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
import { commentAnchorRange } from "@probo/ui/src/RichEditor/commentAnchors";
import type { Editor } from "@tiptap/core";

export type ReviewMarginAnchorThread = {
  id: string;
  anchorStart: string;
  anchorEnd: string;
  anchorQuote?: string | null;
  suggestionId?: string | null;
  comments?: {
    edges: ReadonlyArray<{ node: { content: string } }>;
  };
};

export function threadSignature(threads: ReadonlyArray<ReviewMarginAnchorThread>) {
  return threads
    .map(thread => `${thread.id}:${thread.suggestionId ?? ""}:${thread.anchorStart}:${thread.anchorEnd}`)
    .join("|");
}

export function commentBodySignature(threads: ReadonlyArray<ReviewMarginAnchorThread>) {
  return threads
    .map(thread => `${thread.id}:${thread.comments?.edges.map(edge => edge.node.content).join("\n") ?? ""}`)
    .join("|");
}

export function commentThreadsWithoutSuggestions<T extends { suggestionId?: string | null }>(
  threads: ReadonlyArray<T>,
  suggestionIds: ReadonlySet<string>,
) {
  return threads.filter(thread =>
    thread.suggestionId == null || !suggestionIds.has(thread.suggestionId),
  );
}

export function threadPosition(
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
export function quotePosition(doc: Editor["state"]["doc"], quote: string): number | null {
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

export function editorContentTop(editor: Editor, pos: number) {
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

export function placeCardTops(
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
