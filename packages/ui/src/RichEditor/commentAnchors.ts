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

import type { DocHandle } from "@automerge/prosemirror";
import { Extension } from "@tiptap/core";
import type { Node as ProseMirrorNode } from "@tiptap/pm/model";
import { Plugin, PluginKey } from "@tiptap/pm/state";
import { Decoration, DecorationSet } from "@tiptap/pm/view";
import type { MutableRefObject } from "react";

import type { RichEditorAutomergeDocument } from "./collaboration";
import { pmSelectionFromPresence, richEditorPresenceAdapter } from "./repoPresence";
import type { TextSelection } from "./repoSelection";

export type CommentAnchor = {
  id: string;
  anchorStart: string;
  anchorEnd: string;
  active?: boolean;
};

const commentAnchorPluginKey = new PluginKey<DecorationSet>("document-comment-anchors");

export function encodeDocumentAnchor(value: string): string {
  const bytes = new TextEncoder().encode(value);
  let binary = "";
  for (const byte of bytes) {
    binary += String.fromCharCode(byte);
  }
  return btoa(binary);
}

export function decodeDocumentAnchor(value: string): string | null {
  try {
    const binary = atob(value);
    const bytes = Uint8Array.from(binary, char => char.charCodeAt(0));
    const decoded = new TextDecoder().decode(bytes);
    return decoded === "" ? null : decoded;
  } catch {
    return null;
  }
}

export function createCommentAnchorExtension(
  handle: DocHandle<RichEditorAutomergeDocument>,
  controller: CommentAnchorController,
) {
  return Extension.create({
    name: "documentCommentAnchors",
    addProseMirrorPlugins() {
      return [
        new Plugin({
          key: commentAnchorPluginKey,
          state: {
            init: () => DecorationSet.empty,
            apply(transaction, decorations) {
              const next = transaction.getMeta(commentAnchorPluginKey) as
                | CommentAnchor[]
                | undefined;
              if (!next) return decorations.map(transaction.mapping, transaction.doc);
              return decorationsForAnchors(handle, transaction.doc, next);
            },
          },
          props: {
            decorations(state) {
              return commentAnchorPluginKey.getState(state);
            },
            handleClick(_view, _pos, event) {
              const target = event.target;
              if (!(target instanceof HTMLElement)) return false;
              const id = target.closest("[data-comment-thread]")?.getAttribute("data-comment-thread");
              if (!id) return false;
              controller.current.onOpenThread(id);
              return false;
            },
          },
        }),
      ];
    },
  });
}

export function refreshCommentAnchors(
  editor: {
    isDestroyed?: boolean;
    state: import("@tiptap/pm/state").EditorState;
    view?: { dispatch: (tr: import("@tiptap/pm/state").Transaction) => void } | null;
  },
  anchors: CommentAnchor[],
) {
  if (editor.isDestroyed) return;
  try {
    if (!editor.view) return;
  } catch {
    return;
  }
  editor.view.dispatch(editor.state.tr.setMeta(commentAnchorPluginKey, anchors));
}

export function commentAnchorRange(
  handle: DocHandle<RichEditorAutomergeDocument>,
  doc: import("@tiptap/pm/model").Node,
  anchor: Pick<CommentAnchor, "anchorStart" | "anchorEnd">,
) {
  const document = handle.doc();
  if (!document) return null;
  const start = decodeDocumentAnchor(anchor.anchorStart);
  const end = decodeDocumentAnchor(anchor.anchorEnd);
  if (!start || !end) return null;
  const selection: TextSelection = { field: "body", anchor: start, head: end };
  let resolved: ReturnType<typeof pmSelectionFromPresence>;
  try {
    resolved = pmSelectionFromPresence(richEditorPresenceAdapter(), document, selection);
  } catch {
    resolved = null;
  }
  if (!resolved) return null;
  const from = Math.min(resolved.anchorPosition, resolved.headPosition);
  const to = Math.max(resolved.anchorPosition, resolved.headPosition);
  if (from === to || from < 0 || to > doc.content.size) return null;
  return { from, to };
}

// One inline highlight per text block. A single inline decoration cannot
// cross a paragraph, and one thread still covers the whole selection.
export function commentHighlightRanges(
  doc: ProseMirrorNode,
  from: number,
  to: number,
): Array<{ from: number; to: number; block: boolean }> {
  if (from >= to) return [];
  const ranges: Array<{ from: number; to: number; block: boolean }> = [];
  doc.nodesBetween(from, to, (node, pos) => {
    if (node.isText) {
      const start = Math.max(from, pos);
      const end = Math.min(to, pos + node.nodeSize);
      if (start < end) ranges.push({ from: start, to: end, block: false });
      return false;
    }
    if (node.isTextblock && node.content.size === 0 && from <= pos && to >= pos + node.nodeSize) {
      ranges.push({ from: pos, to: pos + node.nodeSize, block: true });
      return false;
    }
    return true;
  });
  return ranges;
}

function decorationsForAnchors(
  handle: DocHandle<RichEditorAutomergeDocument>,
  doc: import("@tiptap/pm/model").Node,
  anchors: CommentAnchor[],
) {
  const decorations: Decoration[] = [];

  for (const anchor of anchors) {
    const range = commentAnchorRange(handle, doc, anchor);
    if (!range) continue;
    const attrs = {
      "data-comment-thread": anchor.id,
      class: anchor.active ? "bg-amber-5" : "bg-amber-4",
    };
    for (const highlight of commentHighlightRanges(doc, range.from, range.to)) {
      const spec = { ...attrs };
      decorations.push(highlight.block
        ? Decoration.node(highlight.from, highlight.to, spec)
        : Decoration.inline(highlight.from, highlight.to, spec));
    }
  }

  return DecorationSet.create(doc, decorations);
}

export type CommentAnchorController = MutableRefObject<{
  onOpenThread: (id: string) => void;
}>;
