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
import type { Mark as ProseMirrorMark, Node as ProseMirrorNode } from "@tiptap/pm/model";
import { liftListItem, sinkListItem } from "@tiptap/pm/schema-list";
import { type EditorState, Plugin, PluginKey, TextSelection, type Transaction } from "@tiptap/pm/state";
import type { EditorView } from "@tiptap/pm/view";
import { Decoration, DecorationSet } from "@tiptap/pm/view";
import type { MutableRefObject } from "react";

import {
  type AutomergeMarkMapping,
  parseJSONObject,
  readStringAttribute,
} from "./automergeMapping";
import { automergeRemoteTransaction } from "./AutomergeSyncPlugin";
import { linkEditorPluginKey } from "./LinkExtension";
import {
  discardBlockSuggestions,
  readSuggestionBlock,
  reassignBlockSuggestions,
  suggestCommand,
} from "./suggestionBlock";
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
import {
  discardStructuralSuggestion,
  reassignStructuralSuggestion,
  structuralSuggestionNodes,
  structuralSuggestionSplit,
  structureSignature,
  suggestionStructureMeta,
} from "./suggestionStructure";
import {
  isFormatMark,
  marksForSuggestionText,
  reassignFormatMarks,
  revertFormatSuggestion,
  styleMarkStep,
  type SuggestionStyleName,
  suggestLink,
  suggestStyleToggle,
} from "./suggestionStyle";

export const suggestionInsertName = "suggestionInsert";
export const suggestionDeleteName = "suggestionDelete";

const suggestingPluginKey = new PluginKey<SuggestingController>("document-suggesting");
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
  // Called when an edit cannot be tracked as a suggestion and was dropped.
  onBlocked?: () => void;
};

export type SuggestingStorage = {
  controller: MutableRefObject<SuggestingController> | null;
};

const suggestingExtensionName = "documentSuggesting";

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

// suggestionAutomergeMapping stores each suggestion mark in Automerge as a
// JSON object carrying its suggestion id. Kebab-case mark names are the write
// format; the legacy camelCase aliases are added by the collaboration adapter.
function suggestionAutomergeMapping(markName: string): () => AutomergeMarkMapping {
  return () => ({
    markName,
    parsers: {
      fromAutomerge: (value) => {
        if (typeof value !== "string") return { id: "" };
        const parsed = parseJSONObject(value);
        return { id: readStringAttribute(parsed, "id", "") };
      },
      fromProsemirror: (mark: ProseMirrorMark) => JSON.stringify({
        id: readStringAttribute(mark.attrs, "id", ""),
      }),
    },
  });
}

export const SuggestionInsert = Mark.create({
  name: suggestionInsertName,
  ...suggestionMark,
  automerge: suggestionAutomergeMapping("suggestion-insert"),
  parseHTML() {
    return [{ tag: "span[data-suggestion-insert]" }];
  },
  renderHTML({ HTMLAttributes }: { HTMLAttributes: Record<string, unknown> }) {
    return ["span", mergeAttributes(HTMLAttributes, {
      "data-suggestion-insert": stringAttr(HTMLAttributes, "id"),
      "class": "text-green-11",
    }), 0];
  },
});

export const SuggestionDelete = Mark.create({
  name: suggestionDeleteName,
  ...suggestionMark,
  automerge: suggestionAutomergeMapping("suggestion-delete"),
  parseHTML() {
    return [{ tag: "span[data-suggestion-delete]" }];
  },
  renderHTML({ HTMLAttributes }: { HTMLAttributes: Record<string, unknown> }) {
    return ["span", mergeAttributes(HTMLAttributes, {
      "data-suggestion-delete": stringAttr(HTMLAttributes, "id"),
      "class": "text-red-11 line-through",
    }), 0];
  },
});

export function createSuggestingExtension(
  controller: MutableRefObject<SuggestingController>,
) {
  return Mark.create<unknown, SuggestingStorage>({
    name: suggestingExtensionName,
    // Paste/key handlers must run before MarkdownPaste (and the base keymap),
    // otherwise those extensions edit the document untracked while suggesting.
    priority: 1_000,
    addStorage() {
      return { controller };
    },
    addProseMirrorPlugins() {
      return [
        new Plugin<SuggestingController>({
          key: suggestingPluginKey,
          state: {
            init: () => controller.current,
            apply: () => controller.current,
          },
          // Menus and commands that change the document structure (tables,
          // images, lists) bypass the text handlers. Unless they were tracked,
          // drop them rather than let an untracked edit through.
          filterTransaction(transaction) {
            if (!controller.current.enabled || !transaction.docChanged) return true;
            if (
              transaction.getMeta(suggestionStructureMeta)
              || metaInput(transactionMeta(transaction, suggestingPluginKey))
              || transaction.getMeta(automergeRemoteTransaction)
              || transaction.getMeta("history$")
            ) {
              return true;
            }
            if (transaction.steps.some(styleMarkStep)) {
              controller.current.onBlocked?.();
              return false;
            }
            if (structureSignature(transaction.before) === structureSignature(transaction.doc)) {
              return true;
            }
            controller.current.onBlocked?.();
            return false;
          },
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
              metaInput(transactionMeta(transaction, suggestingPluginKey)),
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

type SuggestingEditor = {
  isDestroyed?: boolean;
  state: EditorState;
  view?: { dispatch: (tr: Transaction) => void } | null;
  storage: unknown;
};

export function isSuggesting(editor: SuggestingEditor): boolean {
  return suggestingController(editor)?.current.enabled === true;
}

// suggestingChangeId allocates the suggestion a structural edit should use.
// fresh starts a new card instead of extending the text the user is typing.
export function suggestingChangeId(state: EditorState, fresh: boolean): string | null {
  const controller = suggestingPluginKey.getState(state);
  if (!controller?.enabled) return null;
  if (fresh) {
    controller.onBurstFinished();
    controller.suggestionId = null;
  }
  let id = fresh ? null : suggestionIdForInput(state, controller);
  if (!id) {
    void controller.onNeedId();
    id = controller.suggestionId;
  }
  return id;
}

// applySuggestingStructure runs a structural edit as a suggestion when the
// editor is suggesting. It returns false when the editor is not suggesting,
// so the caller runs the plain command instead.
export function applySuggestingStructure(
  editor: SuggestingEditor,
  build: (state: EditorState, id: string) => Transaction | null,
  options?: { fresh?: boolean },
): boolean {
  const controller = suggestingController(editor);
  if (!controller?.current.enabled) return false;

  const current = controller.current;
  if (!build(editor.state, "pending")) {
    current.onBlocked?.();
    return true;
  }
  // A heading or list is its own card. Reusing the open text burst would
  // merge that card with the words typed just before it.
  let id: string | null = null;
  if (options?.fresh) {
    current.onBurstFinished();
    current.suggestionId = null;
  } else {
    id = suggestionIdForInput(editor.state, current);
  }
  if (!id) {
    void current.onNeedId();
    id = current.suggestionId;
  }
  if (!id || !editorViewReady(editor)) {
    current.onBlocked?.();
    return true;
  }

  const tr = build(editor.state, id);
  if (!tr) {
    current.onBlocked?.();
    return true;
  }
  editor.view.dispatch(tr);
  return true;
}

function suggestingController(
  editor: SuggestingEditor,
): MutableRefObject<SuggestingController> | null {
  const storage = (editor.storage as Record<string, SuggestingStorage | undefined>)[
    suggestingExtensionName
  ];
  return storage?.controller ?? null;
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
      init: (_config, state) => ({
        id: "",
        decorations: suggestionFocusDecorations(state.doc, ""),
      }),
      apply(transaction, previous) {
        const meta = transactionMeta(transaction, suggestionFocusKey);
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
  const decorations: Decoration[] = [];
  doc.descendants((node, pos) => {
    if (!node.isText && !node.isTextblock) {
      const structural = structuralSuggestionSplit(node);
      if (structural && structural.kind !== "structural") {
        const focused = id !== "" && structural.id === id;
        const base = structural.kind === "insertion"
          ? "suggestion-inserted-block"
          : "suggestion-deleted-block";
        const highlight = structural.kind === "insertion" ? "bg-green-3" : "bg-red-3";
        decorations.push(Decoration.node(pos, pos + node.nodeSize, {
          "class": focused ? `${base} ${highlight}` : base,
          "data-suggestion-block": structural.id,
        }));
      }
      return;
    }
    const split = node.isTextblock ? suggestionSplitOn(node) : null;
    if (split?.kind === "deletion") {
      const focused = id !== "" && split.id === id;
      decorations.push(Decoration.node(pos, pos + node.nodeSize, {
        class: focused ? "suggestion-deleted-line bg-red-3" : "suggestion-deleted-line",
      }));
    }
    if (!node.isText || id === "") return;
    for (const mark of node.marks) {
      if (mark.attrs.id !== id) continue;
      const focused = mark.type.name === suggestionInsertName
        ? "bg-green-3"
        : mark.type.name === suggestionDeleteName
          ? "bg-red-3"
          : isFormatMark(mark)
            ? mark.attrs.action === "remove" ? "bg-red-3" : "bg-green-3"
            : null;
      if (!focused) continue;
      decorations.push(Decoration.inline(pos, pos + node.nodeSize, { class: focused }));
    }
  });
  return DecorationSet.create(doc, decorations);
}

const suggestingStyleKeys: Record<string, SuggestionStyleName> = {
  b: "bold",
  i: "italic",
  u: "underline",
  e: "code",
};

function suggestingStyleKey(key: string, shift: boolean): SuggestionStyleName | null {
  if (!shift && key in suggestingStyleKeys) return suggestingStyleKeys[key];
  if (shift && key === "s") return "strike";
  return null;
}

export function suggestingControllerFrom(editor: { storage: object }): SuggestingController | null {
  const storage = editor.storage as { documentSuggesting?: SuggestingStorage };
  return storage.documentSuggesting?.controller?.current ?? null;
}

// applySuggestingStyle turns a bold, italic, or other style toggle into a
// suggestion. An empty selection returns false so the stored mark still
// applies to the next typed character.
export function applySuggestingStyle(
  view: EditorView,
  controller: SuggestingController,
  style: SuggestionStyleName,
): boolean {
  return applyFormatChange(view, controller, (state, id, from, to) =>
    suggestStyleToggle(state, id, style, from, to),
  );
}

// applySuggestingLink suggests adding or removing a link. A null href removes it.
export function applySuggestingLink(
  view: EditorView,
  controller: SuggestingController,
  href: string | null,
): boolean {
  return applyFormatChange(view, controller, (state, id, from, to) =>
    suggestLink(state, id, from, to, href),
  );
}

function applyFormatChange(
  view: EditorView,
  controller: SuggestingController,
  change: (state: EditorState, id: string, from: number, to: number) => Transaction | null,
): boolean {
  if (!controller.enabled || view.state.selection.empty) return false;
  let id = suggestionIdForInput(view.state, controller);
  if (!id) {
    void controller.onNeedId();
    id = controller.suggestionId;
  }
  if (!id) {
    controller.onBlocked?.();
    return true;
  }
  const { from, to } = view.state.selection;
  const tr = change(view.state, id, from, to);
  if (!tr) {
    controller.onBlocked?.();
    return true;
  }
  tr.setMeta(suggestingPluginKey, { input: true });
  view.dispatch(tr.scrollIntoView());
  return true;
}

function linkHrefInSelection(state: EditorState): string {
  const { from, to } = state.selection;
  let href = "";
  state.doc.nodesBetween(from, to, (node) => {
    const link = node.marks.find(mark => mark.type.name === "link");
    if (typeof link?.attrs.href === "string" && link.attrs.href !== "") href = link.attrs.href;
  });
  return href;
}

function applySuggestingStructureFromView(
  view: EditorView,
  controller: SuggestingController,
  build: (state: EditorState, id: string) => Transaction | null,
): boolean {
  if (!build(view.state, "pending")) {
    controller.onBlocked?.();
    return true;
  }
  controller.onBurstFinished();
  controller.suggestionId = null;
  void controller.onNeedId();
  const id = controller.suggestionId;
  if (!id) {
    controller.onBlocked?.();
    return true;
  }
  const tr = build(view.state, id);
  if (!tr) {
    controller.onBlocked?.();
    return true;
  }
  view.dispatch(tr);
  return true;
}

function listItemType(state: EditorState) {
  const { $from } = state.selection;
  for (let depth = $from.depth; depth > 0; depth -= 1) {
    const name = $from.node(depth).type.name;
    if (name === "taskItem") return state.schema.nodes.taskItem ?? null;
    if (name === "listItem") return state.schema.nodes.listItem ?? null;
  }
  return null;
}

function applySuggestingKey(
  view: EditorView,
  controller: SuggestingController,
  event: KeyboardEvent,
): boolean {
  if (!controller.enabled) return false;
  if (event.metaKey || event.ctrlKey) {
    const key = event.key.toLowerCase();
    const style = suggestingStyleKey(key, event.shiftKey);
    if (style && !event.altKey) {
      if (view.state.selection.empty) return false;
      return applySuggestingStyle(view, controller, style);
    }
    if (key === "k" && !event.shiftKey && !event.altKey) {
      if (view.state.selection.empty) return true;
      view.dispatch(view.state.tr.setMeta(linkEditorPluginKey, linkHrefInSelection(view.state)));
      return true;
    }
  }
  if (event.key === "Tab" && !event.altKey && !event.metaKey && !event.ctrlKey) {
    const item = listItemType(view.state);
    if (!item) return false;
    const command = event.shiftKey ? liftListItem(item) : sinkListItem(item);
    return applySuggestingStructureFromView(view, controller, (state, id) =>
      suggestCommand(state, id, command),
    );
  }
  if (event.key === "Enter" && !event.shiftKey) {
    return applySuggestingEnter(view, controller);
  }
  // A hard break cannot be tracked as a suggestion, so reject it rather than
  // let the default handler insert one untracked.
  if (event.key === "Enter" && event.shiftKey) return true;
  if (event.key !== "Backspace" && event.key !== "Delete") return false;
  if (isTextblockBoundary(view.state, event.key)) {
    return applySuggestingLineBreak(view, controller, event.key);
  }

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
  if (deleteForward) {
    const next = nextCodePoint(state.doc, from);
    if (!next) return true;
    const end = from + next.text.length;
    const tr = state.tr;
    if (markCovers(next.marks, suggestionInsertName, id)) {
      tr.delete(from, end);
    } else if (!hasDeletion(next.marks)) {
      markDeletion(tr, from, end, id);
      tr.setSelection(TextSelection.create(tr.doc, from));
    }
    tr.setMeta(suggestingPluginKey, { input: true });
    view.dispatch(tr.scrollIntoView());
    return true;
  }

  const prev = prevCodePoint(state.doc, from);
  if (!prev) return true;
  const start = from - prev.length;
  const covered = nextCodePoint(state.doc, start)?.marks ?? [];
  const tr = state.tr;
  if (markCovers(covered, suggestionInsertName, id)) {
    tr.delete(start, from);
  } else if (!hasDeletion(covered)) {
    markDeletion(tr, start, from, id);
    tr.setSelection(TextSelection.create(tr.doc, start));
  } else {
    tr.setSelection(TextSelection.create(state.doc, start));
  }

  tr.setMeta(suggestingPluginKey, { input: true });
  view.dispatch(tr.scrollIntoView());
  return true;
}

// Backspace or Delete at a paragraph boundary suggests removing that line
// break. The paragraphs stay apart until the suggestion is accepted.
function applySuggestingLineBreak(
  view: EditorView,
  controller: SuggestingController,
  key: string,
): boolean {
  const blockPos = lineBreakBlockPos(view.state, key);
  if (blockPos == null) return false;

  let id = suggestionIdForInput(view.state, controller);
  if (!id) {
    void controller.onNeedId();
    id = controller.suggestionId;
  }
  if (!id) return true;

  const block = view.state.doc.nodeAt(blockPos);
  if (!block?.isTextblock) return false;
  const split = suggestionSplitOn(block);
  if (split?.id === id && split.kind === "insertion") return false;
  if (split && split.id !== id) return true;
  if (split?.kind === "deletion") {
    return deleteBesideDeletedLine(view, controller, key, blockPos);
  }
  if (!block.type.spec.attrs?.unknownAttrs) return false;

  const tr = view.state.tr;
  const current = attrRecord(nodeUnknownAttrs(block));
  tr.setNodeMarkup(blockPos, undefined, {
    ...block.attrs,
    unknownAttrs: {
      ...current,
      [suggestionSplitAttr]: id,
      [suggestionSplitKindAttr]: "deletion",
    },
  });
  tr.setMeta(suggestingPluginKey, { input: true });
  view.dispatch(tr.scrollIntoView());
  return true;
}

function lineBreakBlockPos(state: EditorState, key: string): number | null {
  const { empty, $from } = state.selection;
  if (!empty || !$from.parent.isTextblock) return null;
  if (key === "Backspace" && $from.parentOffset === 0) return $from.before();
  if (key === "Delete" && $from.parentOffset === $from.parent.content.size) {
    return $from.after();
  }
  return null;
}

function deleteBesideDeletedLine(
  view: EditorView,
  controller: SuggestingController,
  key: string,
  blockPos: number,
): boolean {
  let id = suggestionIdForInput(view.state, controller);
  if (!id) {
    void controller.onNeedId();
    id = controller.suggestionId;
  }
  if (!id) return true;

  const { doc } = view.state;
  if (key === "Backspace") {
    const prev = doc.resolve(blockPos).nodeBefore;
    if (!prev?.isTextblock) return true;
    const prevPos = blockPos - prev.nodeSize;
    const undeleted = lastUndeletedRange(doc, prevPos, prev);
    if (undeleted) {
      return applySuggestingText(view, controller, undeleted.from, undeleted.to, "");
    }
    return continueDeletedLine(view, id, prevPos, prevPos + 1);
  }

  const block = doc.nodeAt(blockPos);
  if (!block?.isTextblock) return true;
  const undeleted = firstUndeletedRange(doc, blockPos, block);
  if (undeleted) {
    return applySuggestingText(view, controller, undeleted.from, undeleted.to, "");
  }
  const nextPos = blockPos + block.nodeSize;
  const next = doc.nodeAt(nextPos);
  if (!next?.isTextblock) return true;
  return continueDeletedLine(
    view,
    id,
    nextPos,
    blockPos + 1 + block.content.size,
  );
}

// Empty or fully deleted paragraphs keep walking: mark the next break for this
// suggestion, or move onto a break this suggestion already owns.
function continueDeletedLine(
  view: EditorView,
  id: string,
  blockPos: number,
  caret: number,
): boolean {
  const block = view.state.doc.nodeAt(blockPos);
  if (!block?.isTextblock) return true;
  const split = suggestionSplitOn(block);
  if (split && split.id !== id) return true;

  const tr = view.state.tr;
  if (split?.kind === "deletion") {
    tr.setSelection(TextSelection.create(tr.doc, caret));
    tr.setMeta(suggestingPluginKey, { input: true });
    view.dispatch(tr.scrollIntoView());
    return true;
  }
  if (!block.type.spec.attrs?.unknownAttrs) return true;

  const current = attrRecord(nodeUnknownAttrs(block));
  tr.setNodeMarkup(blockPos, undefined, {
    ...block.attrs,
    unknownAttrs: {
      ...current,
      [suggestionSplitAttr]: id,
      [suggestionSplitKindAttr]: "deletion",
    },
  });
  tr.setSelection(TextSelection.create(tr.doc, caret));
  tr.setMeta(suggestingPluginKey, { input: true });
  view.dispatch(tr.scrollIntoView());
  return true;
}

function lastUndeletedRange(
  doc: ProseMirrorNode,
  blockPos: number,
  block: ProseMirrorNode,
): { from: number; to: number } | null {
  let pos = blockPos + block.nodeSize - 1;
  while (pos > blockPos) {
    const char = prevCodePoint(doc, pos);
    if (!char) break;
    const from = pos - char.length;
    const covered = nextCodePoint(doc, from);
    if (covered && !hasDeletion(covered.marks)) {
      return { from, to: from + covered.text.length };
    }
    pos = from;
  }
  return null;
}

function firstUndeletedRange(
  doc: ProseMirrorNode,
  blockPos: number,
  block: ProseMirrorNode,
): { from: number; to: number } | null {
  let pos = blockPos + 1;
  const end = blockPos + block.nodeSize - 1;
  while (pos < end) {
    const covered = nextCodePoint(doc, pos);
    if (!covered) break;
    if (!hasDeletion(covered.marks)) {
      return { from: pos, to: pos + covered.text.length };
    }
    pos += covered.text.length;
  }
  return null;
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
    const current = attrRecord(nodeUnknownAttrs(block));
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
      void pendingRequest.then((nextID) => {
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
  const inherited = marksForSuggestionText(tr.doc.resolve(cursor).marks());
  const mark = tr.doc.type.schema.marks[suggestionInsertName].create({ id });
  tr.replaceWith(cursor, cursor, tr.doc.type.schema.text(text, [...inherited, mark]));
  return cursor + text.length;
}

function suggestionSplitId(node: ProseMirrorNode): string | null {
  return suggestionSplitOn(node)?.id ?? null;
}

function suggestionSplitOn(
  node: ProseMirrorNode,
): { id: string; kind: string } | null {
  const unknown = nodeUnknownAttrs(node);
  if (!unknown || typeof unknown !== "object") return null;
  const id = (unknown as { suggestionSplit?: unknown }).suggestionSplit;
  const kind = (unknown as { splitKind?: unknown }).splitKind;
  if (typeof id !== "string" || id === "") return null;
  if (kind !== "insertion" && kind !== "structural" && kind !== "deletion") return null;
  return { id, kind };
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
  const lineBreaks: number[] = [];

  tr.doc.nodesBetween(from, to, (node, pos) => {
    if (
      node.isTextblock
      && pos >= from
      && pos < to
      && tr.doc.resolve(pos).nodeBefore?.isTextblock
    ) {
      lineBreaks.push(pos);
    }
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
  for (const pos of lineBreaks.reverse()) {
    markLineDeletion(tr, tr.mapping.map(pos), id);
  }

  return tr;
}

function markLineDeletion(tr: Transaction, pos: number, id: string) {
  const block = tr.doc.nodeAt(pos);
  if (!block?.isTextblock || !block.type.spec.attrs?.unknownAttrs) return;
  if (suggestionSplitOn(block)) return;
  const current = attrRecord(nodeUnknownAttrs(block));
  tr.setNodeMarkup(pos, undefined, {
    ...block.attrs,
    unknownAttrs: {
      ...current,
      [suggestionSplitAttr]: id,
      [suggestionSplitKindAttr]: "deletion",
    },
  });
}

// An upload can finish after the server has renamed or dropped its local id.
// Only ids an upload is still waiting on are kept, and they are dropped once
// that upload settles.
const settledSuggestionIds = new Map<string, string | null>();
const retainedSuggestionIds = new Map<string, number>();

export function settledSuggestionId(id: string): string | null {
  let current: string | null = id;
  const seen = new Set<string>();
  while (current !== null && settledSuggestionIds.has(current) && !seen.has(current)) {
    seen.add(current);
    current = settledSuggestionIds.get(current) ?? null;
  }
  return current;
}

export function retainSuggestionId(id: string, count = 1) {
  if (count <= 0) return;
  retainedSuggestionIds.set(id, (retainedSuggestionIds.get(id) ?? 0) + count);
}

export function releaseSuggestionId(id: string) {
  const next = (retainedSuggestionIds.get(id) ?? 0) - 1;
  if (next > 0) {
    retainedSuggestionIds.set(id, next);
    return;
  }
  retainedSuggestionIds.delete(id);
  pruneSettledSuggestions();
}

function noteSettlement(fromId: string, toId: string | null) {
  if (!settlementNeeded(fromId)) return;
  settledSuggestionIds.set(fromId, toId);
}

function settlementNeeded(fromId: string): boolean {
  for (const id of retainedSuggestionIds.keys()) {
    if (chainReaches(id, fromId)) return true;
  }
  return false;
}

function chainReaches(start: string, target: string): boolean {
  let current: string | null = start;
  const seen = new Set<string>();
  while (current !== null && !seen.has(current)) {
    if (current === target) return true;
    seen.add(current);
    if (!settledSuggestionIds.has(current)) return false;
    current = settledSuggestionIds.get(current) ?? null;
  }
  return false;
}

function pruneSettledSuggestions() {
  const keep = new Set<string>();
  for (const id of retainedSuggestionIds.keys()) {
    let current: string | null = id;
    const seen = new Set<string>();
    while (current !== null && settledSuggestionIds.has(current) && !seen.has(current)) {
      seen.add(current);
      keep.add(current);
      current = settledSuggestionIds.get(current) ?? null;
    }
  }
  for (const key of settledSuggestionIds.keys()) {
    if (!keep.has(key)) settledSuggestionIds.delete(key);
  }
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
  noteSettlement(fromId, toId);
  if (!editorViewReady(editor)) return;
  const ranges: Array<{ from: number; to: number; marks: readonly ProseMirrorMark[] }> = [];
  const blocks: number[] = [];
  editor.state.doc.descendants((node, pos) => {
    if (node.isTextblock && suggestionSplitId(node) === fromId) blocks.push(pos);
    if (!node.isText) return;
    if (!node.marks.some(mark => mark.attrs.id === fromId && (isSuggestionMark(mark) || isFormatMark(mark)))) return;
    ranges.push({
      from: pos,
      to: pos + node.nodeSize,
      marks: reassignFormatMarks(node.marks.map(mark =>
        mark.attrs.id === fromId && isSuggestionMark(mark)
          ? mark.type.create({ id: toId })
          : mark,
      ), fromId, toId),
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
    const current = attrRecord(nodeUnknownAttrs(block));
    tr.setNodeMarkup(mapped, undefined, {
      ...block.attrs,
      unknownAttrs: { ...current, [suggestionSplitAttr]: toId },
    });
  }
  reassignStructuralSuggestion(tr, fromId, toId);
  reassignBlockSuggestions(tr, fromId, toId);
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
  noteSettlement(id, null);
  if (!editorViewReady(editor)) return;
  const rewrites: Array<{ from: number; to: number; marks: readonly ProseMirrorMark[] }> = [];
  const removals: Array<{ from: number; to: number }> = [];
  const blocks: number[] = [];
  let hasFormat = false;
  editor.state.doc.descendants((node, pos) => {
    if (node.isTextblock && suggestionSplitId(node) === id) blocks.push(pos);
    if (!node.isText) return;
    const inserted = node.marks.some(mark =>
      mark.type.name === suggestionInsertName && mark.attrs.id === id,
    );
    const deleted = node.marks.some(mark =>
      mark.type.name === suggestionDeleteName && mark.attrs.id === id,
    );
    if (node.marks.some(mark => isFormatMark(mark) && mark.attrs.id === id)) hasFormat = true;
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
  const structural = structuralSuggestionNodes(editor.state.doc, id).length > 0;
  let hasBlock = false;
  editor.state.doc.descendants((node) => {
    if (readSuggestionBlock(node)?.id === id) hasBlock = true;
  });
  const untouched = rewrites.length === 0
    && removals.length === 0
    && blocks.length === 0
    && !structural
    && !hasFormat
    && !hasBlock;
  if (untouched) return;
  const tr = editor.state.tr;
  for (const range of rewrites.reverse()) {
    rewriteText(tr, range.from, range.to, range.marks);
  }
  for (const range of removals.reverse()) {
    tr.delete(range.from, range.to);
  }
  for (const pos of blocks.reverse()) {
    const mapped = tr.mapping.map(pos);
    const block = tr.doc.nodeAt(mapped);
    if (!block?.type.spec.attrs?.unknownAttrs) continue;
    const current = attrRecord(nodeUnknownAttrs(block));
    delete current[suggestionSplitAttr];
    delete current[suggestionSplitKindAttr];
    tr.setNodeMarkup(mapped, undefined, {
      ...block.attrs,
      unknownAttrs: current,
    });
  }
  discardStructuralSuggestion(tr, id);
  discardBlockSuggestions(tr, id);
  revertFormatSuggestion(tr, id);
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

function prevCodePoint(doc: ProseMirrorNode, pos: number): string | null {
  if (pos <= 0 || pos > doc.content.size) return null;
  const $pos = doc.resolve(pos);
  const offset = $pos.parentOffset;
  if (offset <= 0) return null;
  const before = $pos.parent.textBetween(Math.max(0, offset - 2), offset);
  if (before === "") return null;
  const chars = Array.from(before);
  return chars[chars.length - 1] ?? null;
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
    item.type.name === suggestionInsertName
    || item.type.name === suggestionDeleteName
    || isFormatMark(item),
  );
  const id = mark ? markAttr(mark, "id") : null;
  return typeof id === "string" && id !== "" ? id : null;
}

function transactionMeta<Key>(
  transaction: { getMeta: (key: Key) => unknown },
  key: Key,
): unknown {
  return transaction.getMeta(key);
}

function metaInput(meta: unknown): boolean {
  return typeof meta === "object"
    && meta !== null
    && "input" in meta
    && meta.input === true;
}

function stringAttr(attrs: Record<string, unknown>, name: string): string {
  const value = attrs[name];
  return typeof value === "string" ? value : "";
}

function nodeUnknownAttrs(node: { attrs: { unknownAttrs?: unknown } }): unknown {
  return node.attrs.unknownAttrs;
}

function attrRecord(value: unknown): Record<string, unknown> {
  if (!value || typeof value !== "object" || Array.isArray(value)) return {};
  return { ...(value as Record<string, unknown>) };
}

function markAttr(mark: { attrs: { id?: unknown } }, name: "id"): unknown {
  return mark.attrs[name];
}
