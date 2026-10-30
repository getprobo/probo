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

import type { Node as ProseMirrorNode } from "@tiptap/pm/model";
import type { EditorState, Transaction } from "@tiptap/pm/state";

export type SuggestionSplitKind = "insertion" | "structural";

export type SuggestionSplit = {
  id: string;
  kind: SuggestionSplitKind;
};

export type SuggestionReassignment = {
  from: string;
  to: string;
};

export const suggestionSplitAttr = "suggestionSplit";
export const suggestionSplitKindAttr = "splitKind";
export const suggestionSplitMeta = "suggestion-split";
export const suggestionReassignMeta = "suggestion-reassign";

// Enter at the end of a block is an insertion. Enter inside existing text is
// only a boundary. The following text stays on the same suggestion either way.
export function suggestionSplitKind(
  doc: ProseMirrorNode,
  pos: number,
): SuggestionSplitKind | null {
  const $pos = doc.resolve(pos);
  if (!$pos.parent.isTextblock) return null;
  if ($pos.parentOffset === $pos.parent.content.size) return "insertion";
  return "structural";
}

export function isTextblockBoundary(state: EditorState, key: string): boolean {
  const { empty, $from } = state.selection;
  if (!empty || !$from.parent.isTextblock) return false;
  if (key === "Backspace") {
    if ($from.parentOffset !== 0) return false;
    return state.doc.resolve($from.before()).nodeBefore?.isTextblock === true;
  }
  if (key === "Delete") {
    if ($from.parentOffset !== $from.parent.content.size) return false;
    return state.doc.resolve($from.after()).nodeAfter?.isTextblock === true;
  }
  return false;
}

export function keepsSuggestionBurst(transactions: readonly Transaction[]): boolean {
  return transactions.some((transaction) => {
    const doc = transaction.docs[0];
    if (!doc) return false;
    return transaction.steps.some(step =>
      textblockSplit(step, doc) !== null
      || textblockJoin(step, doc) !== null
      || textblockRemoval(step, doc) !== null,
    );
  });
}

// The caret is still on this user's suggestion, including a blank line that
// suggestion just inserted. Typing there extends the suggestion.
export function adjacentSuggestionId(state: EditorState): string | null {
  const { empty, $from } = state.selection;
  if (!empty || !$from.parent.isTextblock) return null;

  const split = suggestionIdOnBlock($from.parent);
  if (split) return split;

  const before = suggestionIdOnText($from.nodeBefore);
  if (before) return before;
  const after = suggestionIdOnText($from.nodeAfter);
  if (after) return after;

  if ($from.parentOffset === 0) {
    const previous = state.doc.resolve($from.before()).nodeBefore;
    if (previous?.isTextblock) {
      return suggestionIdOnBlock(previous) ?? suggestionIdOnText(previous.lastChild);
    }
  }
  if ($from.parentOffset === $from.parent.content.size) {
    const next = state.doc.resolve($from.after()).nodeAfter;
    if (next?.isTextblock) {
      return suggestionIdOnBlock(next) ?? suggestionIdOnText(next.firstChild);
    }
  }
  return null;
}

function suggestionIdOnBlock(node: ProseMirrorNode): string | null {
  const unknown = node.attrs.unknownAttrs;
  if (!unknown || typeof unknown !== "object") return null;
  const id = (unknown as { suggestionSplit?: unknown }).suggestionSplit;
  const kind = (unknown as { splitKind?: unknown }).splitKind;
  if (kind !== "insertion") return null;
  return typeof id === "string" && id !== "" ? id : null;
}

function suggestionIdOnText(node: ProseMirrorNode | null | undefined): string | null {
  if (!node?.isText) return null;
  for (const mark of node.marks) {
    if (mark.type.name !== "suggestionInsert") continue;
    if (typeof mark.attrs.id === "string" && mark.attrs.id !== "") return mark.attrs.id;
  }
  return null;
}

export function suggestionSplitFrom(transaction: Transaction): SuggestionSplit | null {
  const value = transaction.getMeta(suggestionSplitMeta) as SuggestionSplit | undefined;
  if (!value || value.id === "") return null;
  if (value.kind !== "insertion" && value.kind !== "structural") return null;
  return value;
}

export function suggestionReassignmentFrom(
  transactions: readonly Transaction[],
): SuggestionReassignment | null {
  for (const transaction of transactions) {
    const value = transaction.getMeta(suggestionReassignMeta) as SuggestionReassignment | undefined;
    if (value && value.from !== "" && value.to !== "") return value;
  }
  return null;
}

export function textblockSplit(
  step: Transaction["steps"][number],
  doc: ProseMirrorNode,
): {
  from: number;
  left: ProseMirrorNode;
  right: ProseMirrorNode;
  ancestors: ProseMirrorNode[];
} | null {
  if (step.toJSON().stepType !== "replace") return null;
  const replace = step as unknown as {
    from: number;
    to: number;
    slice: {
      openStart: number;
      openEnd: number;
      content: {
        childCount: number;
        child: (index: number) => ProseMirrorNode;
      };
    };
  };
  if (replace.from !== replace.to) return null;
  if (replace.slice.openStart !== 1 || replace.slice.openEnd !== 1) return null;
  if (replace.slice.content.childCount !== 2) return null;

  const left = replace.slice.content.child(0);
  const right = replace.slice.content.child(1);
  if (!left.isTextblock || left.content.size !== 0) return null;
  if (!right.isTextblock || right.content.size !== 0) return null;

  const $from = doc.resolve(replace.from);
  if (!$from.parent.isTextblock || left.type !== $from.parent.type) return null;

  const ancestors: ProseMirrorNode[] = [];
  for (let depth = 0; depth < $from.depth; depth += 1) {
    ancestors.push($from.node(depth));
  }
  return { from: replace.from, left, right, ancestors };
}

export function textblockJoin(
  step: Transaction["steps"][number],
  doc: ProseMirrorNode,
): { from: number; to: number } | null {
  if (step.toJSON().stepType !== "replace") return null;
  const replace = step as unknown as {
    from: number;
    to: number;
    slice: { content: { size: number } };
  };
  if (replace.slice.content.size !== 0 || replace.to - replace.from !== 2) return null;
  const $from = doc.resolve(replace.from);
  const $to = doc.resolve(replace.to);
  if ($from.sameParent($to)) return null;
  if (!$from.parent.isTextblock || !$to.parent.isTextblock) return null;
  if ($from.parentOffset !== $from.parent.content.size) return null;
  if ($to.parentOffset !== 0) return null;
  return { from: replace.from, to: replace.to };
}

// Backspace in the paragraph after a blank line deletes that empty paragraph
// instead of joining through it. The step covers the whole empty node.
export function textblockRemoval(
  step: Transaction["steps"][number],
  doc: ProseMirrorNode,
): { at: number } | null {
  if (step.toJSON().stepType !== "replace") return null;
  const replace = step as unknown as {
    from: number;
    to: number;
    slice: { content: { size: number } };
  };
  if (replace.slice.content.size !== 0 || replace.to - replace.from !== 2) return null;
  const node = doc.nodeAt(replace.from);
  if (!node?.isTextblock) return null;
  if (node.content.size !== 0 || node.nodeSize !== 2) return null;
  return { at: replace.from };
}
