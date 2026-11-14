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

import { lift, setBlockType, wrapIn } from "@tiptap/pm/commands";
import { Fragment, type Node as ProseMirrorNode, type NodeType, type Schema } from "@tiptap/pm/model";
import { liftListItem, wrapInList } from "@tiptap/pm/schema-list";
import { type Command, type EditorState, TextSelection, type Transaction } from "@tiptap/pm/state";

import { suggestionStructureMeta } from "./suggestionStructure";

export const suggestionBlockAttr = "suggestionBlock";

const listTypes = new Set([
  "unordered-list-item",
  "ordered-list-item",
  "task-list-item",
]);

export type SuggestionBlockMark = {
  id: string;
  type: string;
  parents: string[];
  attrs: Record<string, unknown>;
};

type BlockLocation = {
  pos: number;
  node: ProseMirrorNode;
  parent: ProseMirrorNode | null;
  identity: SuggestionBlockMark;
};

// suggestCommand runs a block command and stores the previous block on the
// node that changed, as its own suggestion. A command that does not change
// the document still returns a transaction so the caller does not treat it
// as an edit the editor must refuse.
export function suggestCommand(
  state: EditorState,
  id: string,
  command: Command,
): Transaction | null {
  const producedBox: { tr: Transaction | null } = { tr: null };
  const ran = command(state, (tr) => {
    producedBox.tr = tr;
  });
  const produced = producedBox.tr;
  if (!ran) return null;
  if (!produced || !produced.docChanged) {
    return state.tr.setMeta(suggestionStructureMeta, true);
  }

  for (const changed of changedBlocks(state, produced)) {
    stampBlock(produced, changed.pos, id, changed.before);
  }
  produced.setMeta(suggestionStructureMeta, true);
  return produced.scrollIntoView();
}

// changedBlocks finds every block the command retyped in the selection, with
// the identity it had before, so rejecting the suggestion restores each one.
function changedBlocks(
  state: EditorState,
  produced: Transaction,
): Array<{ pos: number; before: SuggestionBlockMark }> {
  const result: Array<{ pos: number; before: SuggestionBlockMark }> = [];
  const seen = new Set<number>();
  const inverse = produced.mapping.invert();
  const visit = (pos: number) => {
    const after = blockAt(produced.doc, pos);
    if (!after || seen.has(after.pos)) return;
    seen.add(after.pos);
    const before = blockAt(state.doc, inverse.map(Math.min(after.pos + 1, produced.doc.content.size)));
    if (before && blockChanged(before.identity, after.identity)) {
      result.push({ pos: after.pos, before: before.identity });
    }
  };

  const { from, to } = produced.selection;
  visit(from);
  produced.doc.nodesBetween(from, to, (node, pos) => {
    if (node.isTextblock) visit(pos + 1);
  });
  return result;
}

export function paragraphCommand(): Command {
  return (state, dispatch) => {
    const paragraph = state.schema.nodes.paragraph;
    if (!paragraph) return false;
    return setBlockType(paragraph)(state, dispatch);
  };
}

export function headingCommand(level: number): Command {
  return (state, dispatch) => {
    const heading = state.schema.nodes.heading;
    const paragraph = state.schema.nodes.paragraph;
    if (!heading || !paragraph) return false;
    const parent = state.selection.$from.parent;
    if (parent.type === heading && parent.attrs.level === level) {
      return setBlockType(paragraph)(state, dispatch);
    }
    return setBlockType(heading, { level })(state, dispatch);
  };
}

export function codeCommand(language?: string): Command {
  return (state, dispatch) => {
    const code = state.schema.nodes.codeBlock;
    const paragraph = state.schema.nodes.paragraph;
    if (!code || !paragraph) return false;
    const parent = state.selection.$from.parent;
    if (parent.type === code && !language) {
      return setBlockType(paragraph)(state, dispatch);
    }
    return setBlockType(code, language ? { language } : null)(state, dispatch);
  };
}

export function wrapCommand(typeName: string): Command {
  return (state, dispatch) => {
    const type = state.schema.nodes[typeName];
    if (!type) return false;
    for (let depth = state.selection.$from.depth; depth > 0; depth -= 1) {
      if (state.selection.$from.node(depth).type === type) {
        return lift(state, dispatch);
      }
    }
    return wrapIn(type)(state, dispatch);
  };
}

export function listCommand(listName: string): Command {
  return (state, dispatch) => {
    const list = state.schema.nodes[listName];
    const item = state.schema.nodes[listName === "taskList" ? "taskItem" : "listItem"];
    if (!list || !item) return false;
    for (let depth = state.selection.$from.depth; depth > 0; depth -= 1) {
      if (state.selection.$from.node(depth).type === list) {
        return liftListItem(item)(state, dispatch);
      }
    }
    return wrapInList(list)(state, dispatch);
  };
}

export function suggestDivider(state: EditorState, id: string): Transaction | null {
  const type = state.schema.nodes.horizontalRule;
  if (!type) return null;
  const node = type.create({
    isAmgBlock: true,
    unknownAttrs: { suggestionSplit: id, splitKind: "insertion" },
  });
  const tr = state.tr.replaceSelectionWith(node);
  tr.setMeta(suggestionStructureMeta, true);
  return tr.scrollIntoView();
}

// suggestTurnInto applies a block command to the block at pos, unwrapping a
// list or quote first when the command expects a text block or a new wrapper.
export function suggestTurnInto(
  state: EditorState,
  id: string,
  pos: number,
  node: ProseMirrorNode,
  command: Command,
  targetKind: "textblock" | "wrapper",
): Transaction | null {
  if (node.isTextblock) {
    return suggestCommand(state, id, commandAt(pos, command));
  }

  return suggestCommand(state, id, (current, dispatch) => {
    const schema = current.schema;
    const tr = current.tr;
    // A new wrapper takes every paragraph that came out of the old one.
    let wrappedSize = 0;
    if (targetKind === "textblock") {
      const { targetContent, remaining } = decomposeWrapper(node);
      if (!targetContent) return false;
      const target = schema.nodes.paragraph?.create(null, targetContent);
      if (!target) return false;
      tr.replaceWith(pos, pos + node.nodeSize, [target, ...remaining]);
    } else {
      const textBlocks = collectTextBlocks(node);
      if (textBlocks.length === 0) return false;
      const paragraph = schema.nodes.paragraph;
      if (!paragraph) return false;
      const paragraphs = textBlocks.map(block => paragraph.create(null, block.content));
      tr.replaceWith(pos, pos + node.nodeSize, paragraphs);
      wrappedSize = paragraphs.reduce((size, block) => size + block.nodeSize, 0);
    }
    const next = current.apply(tr);
    const $near = next.doc.resolve(Math.min(pos + 1, next.doc.content.size));
    const selection = wrappedSize > 0
      ? TextSelection.create(next.doc, pos + 1, Math.min(pos + wrappedSize - 1, next.doc.content.size))
      : TextSelection.near($near);
    const selected = next.tr.setSelection(selection);
    const selectedState = next.apply(selected);
    const innerBox: { tr: Transaction | null } = { tr: null };
    const ran = command(selectedState, (nextTr) => {
      innerBox.tr = nextTr;
    });
    // Unwrapping already made the change when the command has nothing left
    // to do, like turning a list item into the paragraph it now is.
    const inner = ran ? innerBox.tr : null;
    if (!dispatch) return true;
    for (const step of selected.steps) dispatchStep(tr, step);
    if (inner) {
      for (const step of inner.steps) dispatchStep(tr, step);
      keepSelection(tr, inner);
    } else {
      tr.setSelection(TextSelection.create(tr.doc, selection.from, selection.to));
    }
    dispatch(tr);
    return true;
  });
}

export function suggestionBlockParents(doc: ProseMirrorNode, pos: number): string[] {
  const $pos = doc.resolve(Math.min(pos, doc.content.size));
  return parentsAt($pos, $pos.depth);
}

export function readSuggestionBlock(node: ProseMirrorNode): SuggestionBlockMark | null {
  const raw = unknownRecord(node)[suggestionBlockAttr];
  if (typeof raw !== "string" || raw === "") return null;
  try {
    const parsed: unknown = JSON.parse(raw);
    if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) return null;
    const record = parsed as Record<string, unknown>;
    if (typeof record.id !== "string" || record.id === "") return null;
    const parents = Array.isArray(record.parents)
      ? record.parents.filter((parent): parent is string => typeof parent === "string")
      : [];
    const attrs = record.attrs && typeof record.attrs === "object" && !Array.isArray(record.attrs)
      ? record.attrs as Record<string, unknown>
      : {};
    return {
      id: record.id,
      type: typeof record.type === "string" ? record.type : "",
      parents,
      attrs,
    };
  } catch {
    return null;
  }
}

export function automergeBlockType(
  node: ProseMirrorNode,
  parent: ProseMirrorNode | null,
): string | null {
  if (node.type.name === "listItem") {
    return parent?.type.name === "orderedList" ? "ordered-list-item" : "unordered-list-item";
  }
  switch (node.type.name) {
    case "paragraph":
      return "paragraph";
    case "heading":
      return "heading";
    case "blockquote":
      return "blockquote";
    case "codeBlock":
      return "code-block";
    case "horizontalRule":
      return "horizontal-rule";
    case "taskItem":
      return "task-list-item";
    case "image":
      return "image";
    case "file":
      return "file";
    case "mention":
      return "mention";
    default:
      return null;
  }
}

export function reassignBlockSuggestions(tr: Transaction, fromId: string, toId: string) {
  const hits: Array<{ pos: number; mark: SuggestionBlockMark }> = [];
  tr.doc.descendants((node, pos) => {
    const mark = readSuggestionBlock(node);
    if (mark?.id === fromId) hits.push({ pos, mark });
  });
  for (const hit of hits) {
    writeSuggestionBlock(tr, hit.pos, { ...hit.mark, id: toId });
  }
}

export function discardBlockSuggestions(tr: Transaction, id: string) {
  const hits: Array<{ pos: number; node: ProseMirrorNode; mark: SuggestionBlockMark }> = [];
  tr.doc.descendants((node, pos) => {
    const mark = readSuggestionBlock(node);
    if (mark?.id === id) hits.push({ pos, node, mark });
  });
  for (const hit of hits.reverse()) {
    const mapped = tr.mapping.map(hit.pos);
    restoreBlock(tr, mapped, hit.mark);
  }
}

export function insertionSplitAttrs(id: string): Record<string, unknown> {
  return { suggestionSplit: id, splitKind: "insertion" };
}

function commandAt(pos: number, command: Command): Command {
  return (state, dispatch) => {
    const $near = state.doc.resolve(Math.min(pos + 1, state.doc.content.size));
    const selected = state.tr.setSelection(TextSelection.near($near));
    const next = state.apply(selected);
    const innerBox: { tr: Transaction | null } = { tr: null };
    const ran = command(next, (tr) => {
      innerBox.tr = tr;
    });
    const inner = innerBox.tr;
    if (!ran || !inner) return false;
    if (!dispatch) return true;
    for (const step of inner.steps) dispatchStep(selected, step);
    keepSelection(selected, inner);
    dispatch(selected);
    return true;
  };
}

// The stitched transaction replays inner's steps on the same document, so
// inner's selection is valid there. suggestCommand reads it to find the
// blocks the command changed.
function keepSelection(tr: Transaction, inner: Transaction) {
  const size = tr.doc.content.size;
  const from = Math.min(inner.selection.from, size);
  const to = Math.min(inner.selection.to, size);
  tr.setSelection(TextSelection.between(tr.doc.resolve(from), tr.doc.resolve(to)));
}

function dispatchStep(tr: Transaction, step: Transaction["steps"][number]) {
  tr.step(step);
}

function blockAt(doc: ProseMirrorNode, pos: number): BlockLocation | null {
  const $pos = doc.resolve(Math.max(1, Math.min(pos, doc.content.size)));
  for (let depth = $pos.depth; depth > 0; depth -= 1) {
    const node = $pos.node(depth);
    const parent = $pos.node(depth - 1);
    const index = $pos.index(depth - 1);
    if (!persistsAsBlock(node, parent, index)) continue;
    const type = automergeBlockType(node, parent);
    if (!type) continue;
    return {
      pos: $pos.before(depth),
      node,
      parent,
      identity: {
        id: "",
        type,
        parents: parentsAt($pos, depth),
        attrs: persistedAttrs(node),
      },
    };
  }
  return null;
}

function parentsAt($pos: ReturnType<ProseMirrorNode["resolve"]>, depth: number): string[] {
  const parents: string[] = [];
  for (let current = depth - 1; current > 0; current -= 1) {
    const node = $pos.node(current);
    const parent = $pos.node(current - 1);
    const index = $pos.index(current - 1);
    if (!persistsAsBlock(node, parent, index)) continue;
    const type = automergeBlockType(node, parent);
    if (type) parents.unshift(type);
  }
  return parents;
}

function persistsAsBlock(
  node: ProseMirrorNode,
  parent: ProseMirrorNode | null,
  index: number,
): boolean {
  if (node.isInline || !parent) return false;
  if (node.attrs.isAmgBlock === true) return true;
  // A document child and a list item are automerge blocks. The identity
  // plugin marks them later, and a suggestion has to see them first. The
  // first paragraph of a list item or a cell stays folded into its parent.
  if (parent.type.name === "doc") return true;
  if (node.type.name === "listItem" || node.type.name === "taskItem") return true;
  return !(index === 0 && parent.type.contentMatch.defaultType === node.type);
}

function persistedAttrs(node: ProseMirrorNode): Record<string, unknown> {
  const attrs: Record<string, unknown> = {};
  if (node.type.name === "heading" && typeof node.attrs.level === "number") {
    attrs.level = node.attrs.level;
  }
  if (node.type.name === "codeBlock" && typeof node.attrs.language === "string" && node.attrs.language !== "") {
    attrs.language = node.attrs.language;
  }
  if (node.type.name === "taskItem") {
    attrs.checked = node.attrs.checked === true;
  }
  return attrs;
}

function blockChanged(before: SuggestionBlockMark, after: SuggestionBlockMark): boolean {
  return before.type !== after.type
    || before.parents.join("/") !== after.parents.join("/")
    || JSON.stringify(before.attrs) !== JSON.stringify(after.attrs);
}

function stampBlock(tr: Transaction, pos: number, id: string, previous: SuggestionBlockMark) {
  writeSuggestionBlock(tr, pos, { ...previous, id });
}

function writeSuggestionBlock(tr: Transaction, pos: number, mark: SuggestionBlockMark) {
  const node = tr.doc.nodeAt(pos);
  if (!node?.type.spec.attrs?.unknownAttrs) return;
  const unknown = { ...unknownRecord(node) };
  unknown[suggestionBlockAttr] = JSON.stringify({
    id: mark.id,
    type: mark.type,
    parents: mark.parents,
    attrs: mark.attrs,
  });
  tr.setNodeMarkup(pos, undefined, { ...node.attrs, unknownAttrs: unknown });
}

function restoreBlock(tr: Transaction, pos: number, previous: SuggestionBlockMark) {
  const node = tr.doc.nodeAt(pos);
  if (!node) return;
  const schema = tr.doc.type.schema;
  const textblock = textblockType(schema, previous.type);
  if (node.isTextblock && textblock) {
    const attrs = { ...node.attrs, ...textblockAttrs(previous) };
    const unknown = { ...unknownRecord(node) };
    delete unknown[suggestionBlockAttr];
    attrs.unknownAttrs = Object.keys(unknown).length > 0 ? unknown : null;
    tr.setNodeMarkup(pos, textblock, attrs);
    return;
  }
  if ((node.type.name === "listItem" || node.type.name === "taskItem") && previous.type === "paragraph") {
    const $pos = tr.doc.resolve(pos);
    const list = $pos.parent;
    const paragraph = schema.nodes.paragraph;
    if (!paragraph || list.childCount !== 1) {
      clearBlockMark(tr, pos);
      return;
    }
    const content = node.firstChild?.isTextblock ? node.firstChild.content : node.content;
    const restored = paragraph.create(null, content);
    const listPos = $pos.before($pos.depth - 1);
    tr.replaceWith(listPos, listPos + list.nodeSize, restored);
    return;
  }
  clearBlockMark(tr, pos);
}

function clearBlockMark(tr: Transaction, pos: number) {
  const node = tr.doc.nodeAt(pos);
  if (!node) return;
  const unknown = { ...unknownRecord(node) };
  delete unknown[suggestionBlockAttr];
  tr.setNodeMarkup(pos, undefined, {
    ...node.attrs,
    unknownAttrs: Object.keys(unknown).length > 0 ? unknown : null,
  });
}

function textblockType(schema: Schema, automergeType: string): NodeType | null {
  switch (automergeType) {
    case "paragraph":
      return schema.nodes.paragraph ?? null;
    case "heading":
      return schema.nodes.heading ?? null;
    case "code-block":
      return schema.nodes.codeBlock ?? null;
    default:
      return null;
  }
}

function textblockAttrs(previous: SuggestionBlockMark): Record<string, unknown> {
  if (previous.type === "heading") {
    const level = previous.attrs.level;
    return { level: typeof level === "number" ? level : 1 };
  }
  if (previous.type === "code-block" && typeof previous.attrs.language === "string") {
    return { language: previous.attrs.language };
  }
  return {};
}

function unknownRecord(node: ProseMirrorNode): Record<string, unknown> {
  const value = node.attrs.unknownAttrs as unknown;
  if (!value || typeof value !== "object" || Array.isArray(value)) return {};
  return { ...(value as Record<string, unknown>) };
}

function collectTextBlocks(node: ProseMirrorNode): ProseMirrorNode[] {
  const result: ProseMirrorNode[] = [];
  node.forEach((child) => {
    if (child.isTextblock) result.push(child);
    else result.push(...collectTextBlocks(child));
  });
  return result;
}

function decomposeWrapper(node: ProseMirrorNode): {
  targetContent: ProseMirrorNode["content"] | null;
  remaining: ProseMirrorNode[];
} {
  const firstChild = node.firstChild;
  if (!firstChild) return { targetContent: null, remaining: [] };
  let targetContent: ProseMirrorNode["content"] | null = null;
  const promoted: ProseMirrorNode[] = [];
  if (firstChild.isTextblock) {
    targetContent = firstChild.content;
  } else {
    firstChild.forEach((child) => {
      if (targetContent === null && child.isTextblock) targetContent = child.content;
      else promoted.push(child);
    });
  }
  const rest: ProseMirrorNode[] = [];
  node.forEach((child, _offset, index) => {
    if (index > 0) rest.push(child);
  });
  const remaining = [...promoted];
  if (rest.length > 0) remaining.push(node.copy(Fragment.from(rest)));
  return { targetContent, remaining };
}

export function isListBlock(type: string): boolean {
  return listTypes.has(type);
}
