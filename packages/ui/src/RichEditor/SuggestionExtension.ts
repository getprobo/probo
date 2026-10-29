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

import { Mark, mergeAttributes } from "@tiptap/core";
import { splitBlock } from "@tiptap/pm/commands";
import type { Mark as ProseMirrorMark } from "@tiptap/pm/model";
import type { Node as ProseMirrorNode } from "@tiptap/pm/model";
import { type EditorState, Plugin, PluginKey, TextSelection, type Transaction } from "@tiptap/pm/state";
import { Decoration, DecorationSet } from "@tiptap/pm/view";
import type { EditorView } from "@tiptap/pm/view";
import type { MutableRefObject } from "react";

import {
  adjacentSuggestionId,
  isTextblockBoundary,
  keepsSuggestionBurst,
  suggestionReassignMeta,
  suggestionSplitAttr,
  suggestionSplitKind,
  suggestionSplitKindAttr,
  suggestionSplitMeta,
} from "./suggestionSplit";

export const suggestionInsertName = "suggestionInsert";
export const suggestionDeleteName = "suggestionDelete";

const suggestingPluginKey = new PluginKey("document-suggesting");
const suggestionFocusKey = new PluginKey<SuggestionFocusState>("document-suggestion-focus");

type SuggestionFocusState = {
  id: string;
  decorations: DecorationSet;
};

export type PendingSuggestionInput = {
  view: EditorView;
  from: number;
  to: number;
  text: string;
};

export type SuggestingController = {
  enabled: boolean;
  suggestionId: string | null;
  pending: PendingSuggestionInput | null;
  request: Promise<string | null> | null;
  onNeedId: () => Promise<string | null>;
  ownsSuggestion?: (id: string) => boolean;
  onBurstFinished: () => void;
  onOpenSuggestion: (id: string) => void;
};

const suggestionMark = {
  inclusive: false,
  excludes: "",
  addAttributes() {
    return {
      id: {
        default: null,
      },
    };
  },
};

export const SuggestionInsert = Mark.create({
  name: suggestionInsertName,
  ...suggestionMark,
  parseHTML() {
    return [{ tag: "span[data-suggestion-insert]" }];
  },
  renderHTML({ HTMLAttributes }) {
    return ["span", mergeAttributes(HTMLAttributes, {
      "data-suggestion-insert": HTMLAttributes.id,
      class: "text-green-11",
    }), 0];
  },
});

export const SuggestionDelete = Mark.create({
  name: suggestionDeleteName,
  ...suggestionMark,
  parseHTML() {
    return [{ tag: "span[data-suggestion-delete]" }];
  },
  renderHTML({ HTMLAttributes }) {
    return ["span", mergeAttributes(HTMLAttributes, {
      "data-suggestion-delete": HTMLAttributes.id,
      class: "text-red-11 line-through",
    }), 0];
  },
});

export function createSuggestingExtension(
  controller: MutableRefObject<SuggestingController>,
) {
  return Mark.create({
    name: "documentSuggesting",
    addProseMirrorPlugins() {
      return [
        new Plugin({
          key: suggestingPluginKey,
          props: {
            handleTextInput(view, from, to, text) {
              return applySuggestingText(view, controller.current, from, to, text);
            },
            handlePaste(view, event) {
              const text = event.clipboardData?.getData("text/plain") ?? "";
              if (!controller.current.enabled || text === "") return false;
              const { from, to } = view.state.selection;
              event.preventDefault();
              return applySuggestingText(view, controller.current, from, to, text);
            },
            handleKeyDown(view, event) {
              return applySuggestingKey(view, controller.current, event);
            },
            handleClick(view, pos) {
              const id = suggestionIdAt(view, pos);
              if (!id) return false;
              controller.current.onOpenSuggestion(id);
              return false;
            },
          },
          appendTransaction(transactions, oldState, newState) {
            if (!controller.current.enabled || !controller.current.suggestionId) {
              return null;
            }
            const ownInput = transactions.some(transaction =>
              transaction.getMeta(suggestingPluginKey)?.input,
            );
            if (ownInput || keepsSuggestionBurst(transactions)) return null;
            const moved = transactions.some(transaction => transaction.selectionSet)
              && !oldState.selection.eq(newState.selection);
            const currentId = controller.current.suggestionId;
            const stillOnSuggestion = currentId != null
              && adjacentSuggestionId(newState) === currentId;
            if (moved && !stillOnSuggestion) {
              controller.current.onBurstFinished();
            }
            return null;
          },
        }),
        suggestionFocusPlugin(),
      ];
    },
  });
}

export function refreshSuggestionFocus(
  editor: {
    isDestroyed?: boolean;
    view?: { dispatch: (tr: import("@tiptap/pm/state").Transaction) => void } | null;
    state: import("@tiptap/pm/state").EditorState;
  },
  id: string | null,
) {
  if (!editorViewReady(editor)) return;
  editor.view.dispatch(editor.state.tr.setMeta(suggestionFocusKey, id ?? ""));
}

function editorViewReady(
  editor: {
    isDestroyed?: boolean;
    view?: { dispatch: (tr: Transaction) => void } | null;
  },
): editor is { view: { dispatch: (tr: Transaction) => void } } {
  if (editor.isDestroyed) return false;
  try {
    return editor.view != null;
  } catch {
    return false;
  }
}

function suggestionFocusPlugin() {
  return new Plugin<SuggestionFocusState>({
    key: suggestionFocusKey,
    state: {
      init: () => ({ id: "", decorations: DecorationSet.empty }),
      apply(transaction, previous) {
        const meta = transaction.getMeta(suggestionFocusKey);
        const id = typeof meta === "string" ? meta : previous.id;
        if (typeof meta !== "string" && !transaction.docChanged) {
          return {
            id,
            decorations: previous.decorations.map(transaction.mapping, transaction.doc),
          };
        }
        return { id, decorations: suggestionFocusDecorations(transaction.doc, id) };
      },
    },
    props: {
      decorations(state) {
        return suggestionFocusKey.getState(state)?.decorations;
      },
    },
  });
}

function suggestionFocusDecorations(doc: ProseMirrorNode, id: string) {
  if (id === "") return DecorationSet.empty;
  const decorations: Decoration[] = [];
  doc.descendants((node, pos) => {
    if (!node.isText) return;
    for (const mark of node.marks) {
      if (mark.attrs.id !== id) continue;
      const focused = mark.type.name === suggestionInsertName
        ? "bg-green-3"
        : mark.type.name === suggestionDeleteName
          ? "bg-red-3"
          : null;
      if (!focused) continue;
      decorations.push(Decoration.inline(pos, pos + node.nodeSize, { class: focused }));
    }
  });
  return DecorationSet.create(doc, decorations);
}

function applySuggestingKey(
  view: EditorView,
  controller: SuggestingController,
  event: KeyboardEvent,
): boolean {
  if (!controller.enabled) return false;
  if ((event.metaKey || event.ctrlKey) && ["b", "i", "u", "e", "k"].includes(event.key.toLowerCase())) {
    return true;
  }
  if (event.key === "Enter" && !event.shiftKey) {
    return applySuggestingEnter(view, controller);
  }
  if (event.key !== "Backspace" && event.key !== "Delete") return false;
  if (isTextblockBoundary(view.state, event.key)) return false;

  const { state } = view;
  let id = suggestionIdForInput(state, controller);
  if (!id) {
    void controller.onNeedId();
    id = controller.suggestionId;
  }
  if (!id) return true;

  const { from, to, empty } = state.selection;
  if (!empty) {
    return applySuggestingText(view, controller, from, to, "");
  }

  const deleteForward = event.key === "Delete";
  const start = deleteForward ? from : from - 1;
  const end = deleteForward ? from + 1 : from;
  if (start < 0 || end > state.doc.content.size) return true;

  const tr = state.tr;
  const covered = nextCodePoint(state.doc, start)?.marks ?? [];
  if (markCovers(covered, suggestionInsertName, id)) {
    tr.delete(start, end);
  } else if (!hasDeletion(covered)) {
    markDeletion(tr, start, end, id);
    tr.setSelection(TextSelection.create(tr.doc, start));
  } else {
    tr.setSelection(TextSelection.create(state.doc, start));
  }

  tr.setMeta(suggestingPluginKey, { input: true });
  view.dispatch(tr.scrollIntoView());
  return true;
}

function applySuggestingEnter(
  view: EditorView,
  controller: SuggestingController,
): boolean {
  const { state } = view;
  const { empty, $from } = state.selection;
  if (!empty) return false;
  const kind = suggestionSplitKind(state.doc, $from.pos);
  if (!kind) return false;

  let id = suggestionIdForInput(state, controller);
  if (!id) {
    void controller.onNeedId();
    id = controller.suggestionId;
  }
  if (!id) return true;

  let transaction = state.tr;
  const split = splitBlock(state, (next) => {
    transaction = next;
  });
  if (!split) return false;

  const blockPos = transaction.selection.$from.before();
  const block = transaction.doc.nodeAt(blockPos);
  if (block?.type.spec.attrs?.unknownAttrs) {
    const unknown = block.attrs.unknownAttrs;
    const current = unknown && typeof unknown === "object" ? unknown : {};
    transaction.setNodeMarkup(blockPos, undefined, {
      ...block.attrs,
      unknownAttrs: {
        ...current,
        [suggestionSplitAttr]: id,
        [suggestionSplitKindAttr]: kind,
      },
    });
  }
  transaction.setMeta(suggestingPluginKey, { input: true });
  transaction.setMeta(suggestionSplitMeta, { id, kind });
  view.dispatch(transaction.scrollIntoView());
  return true;
}

// A keystroke next to this user's own suggestion continues it. A click, or
// typing beside someone else's suggestion, still starts a new one.
function suggestionIdForInput(
  state: EditorState,
  controller: SuggestingController,
): string | null {
  if (controller.suggestionId) return controller.suggestionId;
  const adjacent = adjacentSuggestionId(state);
  if (!adjacent || !controller.ownsSuggestion?.(adjacent)) return null;
  controller.suggestionId = adjacent;
  return adjacent;
}

function applySuggestingText(
  view: EditorView,
  controller: SuggestingController,
  from: number,
  to: number,
  text: string,
): boolean {
  if (!controller.enabled) return false;
  let id = suggestionIdForInput(view.state, controller);
  if (!id) {
    // A reload does not retain the previous burst's controller id. If the
    // incoming text exactly follows a deletion, restore that deletion rather
    // than opening a second suggestion for the same character.
    if (restoresDeletedText(view.state.doc, from, to, text)) {
      const tr = suggestionInputTransaction(view.state, from, to, text, "");
      tr.setMeta(suggestingPluginKey, { input: true });
      view.dispatch(tr.scrollIntoView());
      return true;
    }
    const pendingRequest = controller.onNeedId();
    id = controller.suggestionId;
    if (!id) {
      queuePendingText(controller, view, from, to, text);
      void pendingRequest.then(nextID => {
        if (!nextID) return;
        controller.suggestionId = nextID;
        flushPendingText(controller);
      });
      return true;
    }
  }

  const tr = suggestionInputTransaction(view.state, from, to, text, id);
  tr.setMeta(suggestingPluginKey, { input: true });
  view.dispatch(tr.scrollIntoView());
  return true;
}

// Typing over text this suggestion deleted restores that text. A different
// character is inserted in front of the remaining deletion, as in Docs.
export function suggestionInputTransaction(
  state: EditorState,
  from: number,
  to: number,
  text: string,
  id: string,
): Transaction {
  if (from !== to && text !== "") {
    return replaceSelection(state, from, to, text, id);
  }

  let tr = state.tr;
  if (from !== to) {
    tr = markDeletion(tr, from, to, id);
  }

  let cursor = tr.mapping.map(from);
  for (const char of Array.from(text)) {
    const next = nextCodePoint(tr.doc, cursor);
    if (next?.text === char && hasDeletion(next.marks)) {
      rewriteText(
        tr,
        cursor,
        cursor + char.length,
        next.marks.filter(mark => mark.type.name !== suggestionDeleteName),
      );
      cursor += char.length;
      continue;
    }

    cursor = insertSuggestionText(tr, cursor, char, id);
  }

  tr.setSelection(TextSelection.create(tr.doc, cursor));
  return tr;
}

function restoresDeletedText(
  doc: ProseMirrorNode,
  from: number,
  to: number,
  text: string,
): boolean {
  if (from !== to || text === "") return false;

  let cursor = from;
  for (const char of Array.from(text)) {
    const next = nextCodePoint(doc, cursor);
    if (next?.text !== char || !hasDeletion(next.marks)) return false;
    cursor += char.length;
  }
  return true;
}

function replaceSelection(
  state: EditorState,
  from: number,
  to: number,
  text: string,
  id: string,
): Transaction {
  const before = state.doc.textBetween(from, to, "\n", "\ufffc");
  const singleCharacter = Array.from(text);
  const { prefix, suffix } = sharedEdges(before, text);
  const tr = state.tr;
  // Browsers report a replacement as one text-input event per keystroke. When
  // the first key appears later in the selected text, restore it in place so
  // subsequent matching keys continue through that existing deletion.
  if (singleCharacter.length === 1) {
    const match = before.indexOf(text);
    if (match >= 0) {
      const matchFrom = from + match;
      const matchTo = matchFrom + text.length;
      markDeletion(tr, from, matchFrom, id);
      markDeletion(tr, matchTo, to, id);
      tr.setSelection(TextSelection.create(tr.doc, matchTo));
      return tr;
    }
  }

  const deleteFrom = from + prefix;
  const deleteTo = to - suffix;
  if (deleteFrom < deleteTo) {
    markDeletion(tr, deleteFrom, deleteTo, id);
  }

  const inserted = text.slice(prefix, text.length - suffix);
  let cursor = tr.mapping.map(deleteTo);
  if (inserted !== "") {
    cursor = insertSuggestionText(tr, cursor, inserted, id);
  }
  tr.setSelection(TextSelection.create(tr.doc, cursor + suffix));
  return tr;
}

function sharedEdges(before: string, after: string): { prefix: number; suffix: number } {
  const oldChars = Array.from(before);
  const newChars = Array.from(after);
  let prefix = 0;
  while (prefix < oldChars.length && prefix < newChars.length && oldChars[prefix] === newChars[prefix]) {
    prefix += 1;
  }
  let suffix = 0;
  while (
    suffix < oldChars.length - prefix
    && suffix < newChars.length - prefix
    && oldChars[oldChars.length - 1 - suffix] === newChars[newChars.length - 1 - suffix]
  ) {
    suffix += 1;
  }
  return {
    prefix: oldChars.slice(0, prefix).join("").length,
    suffix: suffix === 0 ? 0 : oldChars.slice(oldChars.length - suffix).join("").length,
  };
}

function insertSuggestionText(tr: Transaction, cursor: number, text: string, id: string): number {
  const inherited = tr.doc.resolve(cursor).marks().filter(mark =>
    mark.type.name !== suggestionInsertName && mark.type.name !== suggestionDeleteName,
  );
  const mark = tr.doc.type.schema.marks[suggestionInsertName].create({ id });
  tr.replaceWith(cursor, cursor, tr.doc.type.schema.text(text, [...inherited, mark]));
  return cursor + text.length;
}

function suggestionSplitId(node: ProseMirrorNode): string | null {
  const unknown = node.attrs.unknownAttrs;
  if (!unknown || typeof unknown !== "object") return null;
  const id = (unknown as { suggestionSplit?: unknown }).suggestionSplit;
  return typeof id === "string" && id !== "" ? id : null;
}

function rewriteText(
  tr: Transaction,
  from: number,
  to: number,
  marks: readonly ProseMirrorMark[],
) {
  const text = tr.doc.textBetween(from, to);
  if (text === "") return;
  tr.replaceWith(from, to, tr.doc.type.schema.text(text, marks));
}

function markDeletion(
  tr: import("@tiptap/pm/state").Transaction,
  from: number,
  to: number,
  id: string,
) {
  const deleteMark = tr.doc.type.schema.marks[suggestionDeleteName].create({ id });
  const removals: Array<{ from: number; to: number }> = [];
  const rewrites: Array<{ from: number; to: number; marks: ProseMirrorMark[] }> = [];

  tr.doc.nodesBetween(from, to, (node, pos) => {
    if (!node.isText) return;
    const start = Math.max(from, pos);
    const end = Math.min(to, pos + node.nodeSize);
    if (hasDeletion(node.marks)) return;
    if (markCovers(node.marks, suggestionInsertName, id)) {
      removals.push({ from: start, to: end });
      return;
    }
    rewrites.push({
      from: start,
      to: end,
      marks: [...node.marks, deleteMark],
    });
  });

  for (const range of rewrites.reverse()) {
    rewriteText(tr, range.from, range.to, range.marks);
  }

  for (const range of removals.reverse()) {
    tr.delete(range.from, range.to);
  }

  return tr;
}

export function reassignSuggestionId(
  editor: {
    isDestroyed?: boolean;
    state: EditorState;
    view?: { dispatch: (tr: Transaction) => void } | null;
  },
  fromId: string,
  toId: string,
) {
  if (!editorViewReady(editor)) return;
  const ranges: Array<{ from: number; to: number; marks: readonly ProseMirrorMark[] }> = [];
  const blocks: number[] = [];
  editor.state.doc.descendants((node, pos) => {
    if (node.isTextblock && suggestionSplitId(node) === fromId) blocks.push(pos);
    if (!node.isText) return;
    if (!node.marks.some(mark => mark.attrs.id === fromId && isSuggestionMark(mark))) return;
    ranges.push({
      from: pos,
      to: pos + node.nodeSize,
      marks: node.marks.map(mark =>
        mark.attrs.id === fromId && isSuggestionMark(mark)
          ? mark.type.create({ id: toId })
          : mark,
      ),
    });
  });
  const tr = editor.state.tr;
  for (const range of ranges.reverse()) {
    rewriteText(tr, range.from, range.to, range.marks);
  }
  for (const pos of blocks.reverse()) {
    const mapped = tr.mapping.map(pos);
    const block = tr.doc.nodeAt(mapped);
    if (!block?.type.spec.attrs?.unknownAttrs) continue;
    const unknown = block.attrs.unknownAttrs;
    const current = unknown && typeof unknown === "object" ? unknown : {};
    tr.setNodeMarkup(mapped, undefined, {
      ...block.attrs,
      unknownAttrs: { ...current, [suggestionSplitAttr]: toId },
    });
  }
  tr.setMeta(suggestingPluginKey, { input: true });
  tr.setMeta(suggestionReassignMeta, { from: fromId, to: toId });
  tr.setMeta("addToHistory", false);
  editor.view.dispatch(tr);
}

export function discardSuggestionId(
  editor: {
    isDestroyed?: boolean;
    state: EditorState;
    view?: { dispatch: (tr: Transaction) => void } | null;
  },
  id: string,
) {
  if (!editorViewReady(editor)) return;
  const rewrites: Array<{ from: number; to: number; marks: readonly ProseMirrorMark[] }> = [];
  const removals: Array<{ from: number; to: number }> = [];
  editor.state.doc.descendants((node, pos) => {
    if (!node.isText) return;
    const inserted = node.marks.some(mark =>
      mark.type.name === suggestionInsertName && mark.attrs.id === id,
    );
    const deleted = node.marks.some(mark =>
      mark.type.name === suggestionDeleteName && mark.attrs.id === id,
    );
    if (!inserted && !deleted) return;
    const from = pos;
    const to = pos + node.nodeSize;
    if (inserted) {
      removals.push({ from, to });
      return;
    }
    rewrites.push({
      from,
      to,
      marks: node.marks.filter(mark =>
        !(mark.type.name === suggestionDeleteName && mark.attrs.id === id),
      ),
    });
  });
  if (rewrites.length === 0 && removals.length === 0) return;
  const tr = editor.state.tr;
  for (const range of rewrites.reverse()) {
    rewriteText(tr, range.from, range.to, range.marks);
  }
  for (const range of removals.reverse()) {
    tr.delete(range.from, range.to);
  }
  tr.setMeta(suggestingPluginKey, { input: true });
  tr.setMeta("addToHistory", false);
  editor.view.dispatch(tr);
}

function isSuggestionMark(mark: ProseMirrorMark): boolean {
  return mark.type.name === suggestionInsertName || mark.type.name === suggestionDeleteName;
}

function queuePendingText(
  controller: SuggestingController,
  view: EditorView,
  from: number,
  to: number,
  text: string,
) {
  const pending = controller.pending;
  if (pending && pending.view === view && pending.from === from && pending.to === to) {
    pending.text += text;
    return;
  }

  controller.pending = { view, from, to, text };
}

function flushPendingText(controller: SuggestingController) {
  const pending = controller.pending;
  controller.pending = null;
  if (!pending || !controller.suggestionId) return;
  applySuggestingText(pending.view, controller, pending.from, pending.to, pending.text);
}

function nextCodePoint(
  doc: ProseMirrorNode,
  pos: number,
): { text: string; marks: readonly ProseMirrorMark[] } | null {
  if (pos < 0 || pos > doc.content.size) return null;
  const $pos = doc.resolve(pos);
  const child = $pos.parent.maybeChild($pos.index());
  if (!child?.isText || !child.text || $pos.textOffset >= child.text.length) return null;
  const text = Array.from(child.text.slice($pos.textOffset))[0];
  if (!text) return null;
  return { text, marks: child.marks };
}

function markCovers(marks: readonly ProseMirrorMark[], name: string, id: string): boolean {
  return marks.some(mark => mark.type.name === name && mark.attrs.id === id);
}

function hasDeletion(marks: readonly ProseMirrorMark[]): boolean {
  return marks.some(mark => mark.type.name === suggestionDeleteName);
}

function suggestionIdAt(view: EditorView, pos: number): string | null {
  const marks = view.state.doc.resolve(pos).marks();
  const mark = marks.find(item =>
    item.type.name === suggestionInsertName || item.type.name === suggestionDeleteName,
  );
  const id = mark?.attrs.id;
  return typeof id === "string" && id !== "" ? id : null;
}
