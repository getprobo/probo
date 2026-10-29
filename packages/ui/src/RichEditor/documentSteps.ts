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

import * as Automerge from "@automerge/automerge";
import type { SchemaAdapter } from "@automerge/prosemirror";
import pmToAm from "@automerge/prosemirror/dist/pmToAm.js";
import { amMarksFromPmMarks } from "@automerge/prosemirror/dist/schema.js";
import { pmRangeToAmRange } from "@automerge/prosemirror/dist/traversal.js";
import type { Node as ProseMirrorNode } from "@tiptap/pm/model";
import type { Transaction } from "@tiptap/pm/state";

import type { RichEditorAutomergeDocument } from "./collaboration";
import {
  type SuggestionSplit,
  suggestionSplitAttr,
  suggestionSplitKindAttr,
  textblockJoin,
  textblockRemoval,
  textblockSplit,
} from "./suggestionSplit";

type AutomergeBlock = {
  type: Automerge.ImmutableString;
  parents: Automerge.ImmutableString[];
  attrs: { [key: string]: Automerge.MaterializeValue };
  isEmbed: boolean;
};

// Applies one editor transaction to the automerge body without rebuilding
// every span. Unhandled steps return false so the caller can fall back.
export function applyDocumentSteps(
  adapter: SchemaAdapter,
  document: Automerge.Doc<RichEditorAutomergeDocument>,
  path: Automerge.Prop[],
  pmDoc: ProseMirrorNode,
  steps: readonly Transaction["steps"][number][],
  suggestion: SuggestionSplit | null,
): boolean {
  let spans = Automerge.spans(document, path);
  for (const step of steps) {
    const marked = applyMappedMarkStep(adapter, document, path, spans, step);
    if (marked === false) return false;
    if (marked !== true) {
      const deletion = textDeletion(step, pmDoc);
      if (deletion) {
        const range = pmRangeToAmRange(adapter, spans, deletion);
        if (!range) return false;
        let { start, end } = range;
        if (start > end) {
          [start, end] = [end, start];
        }
        Automerge.splice(document, path, start, end - start, "");
      } else if (isSingleTextReplace(step)) {
        pmToAm(adapter, spans, [step], document, pmDoc, path);
      } else if (applyTextblockJoin(adapter, document, path, spans, step, pmDoc)) {
        // The boundary marker is gone.
      } else if (applyTextblockRemoval(adapter, document, path, spans, step, pmDoc)) {
        // The empty paragraph marker is gone.
      } else if (applyTextblockSplit(adapter, document, path, spans, step, pmDoc, suggestion)) {
        // The new block marker is in place.
      } else if (isEditorAttributeStep(step, pmDoc)) {
        // isAmgBlock and suggestionSplit live on the editor node. The block
        // marker was already written by the split, so this step must not
        // rebuild every span.
      } else {
        return false;
      }
    }
    const applied = step.apply(pmDoc).doc;
    if (!applied) return false;
    pmDoc = applied;
    spans = Automerge.spans(document, path);
  }
  return true;
}

export function reassignSuggestionBlocks(
  document: Automerge.Doc<RichEditorAutomergeDocument>,
  path: Automerge.Prop[],
  fromId: string,
  toId: string,
) {
  const spans = Automerge.spans(document, path);
  let index = 0;
  const updates: Array<{ index: number; block: { [key: string]: Automerge.MaterializeValue } }> = [];
  for (const span of spans) {
    if (span.type === "block") {
      const block = span.value;
      const attrs = blockAttrs(block);
      if (attrs?.[suggestionSplitAttr] === fromId) {
        updates.push({
          index,
          block: {
            ...block,
            attrs: { ...attrs, [suggestionSplitAttr]: toId },
          },
        });
      }
      index += 1;
    } else {
      index += String(span.value).length;
    }
  }
  for (const update of updates) {
    Automerge.updateBlock(document, path, update.index, update.block);
  }
}

function applyTextblockRemoval(
  adapter: SchemaAdapter,
  document: Automerge.Doc<RichEditorAutomergeDocument>,
  path: Automerge.Prop[],
  spans: ReturnType<typeof Automerge.spans>,
  step: Transaction["steps"][number],
  pmDoc: ProseMirrorNode,
): boolean {
  const removal = textblockRemoval(step, pmDoc);
  if (!removal) return false;
  const applied = step.apply(pmDoc).doc;
  if (!applied) return false;
  // The marker sits where a join of this boundary would, at the end of the
  // previous paragraph. The first paragraph has no previous position.
  const index = removal.at === 0
    ? 0
    : pmRangeToAmRange(adapter, spans, { from: removal.at - 1, to: removal.at - 1 }).start;
  Automerge.joinBlock(document, path, index);
  if (isLoneDefaultParagraph(applied)) {
    const next = Automerge.spans(document, path);
    if (next[0]?.type === "block") Automerge.joinBlock(document, path, 0);
  }
  return true;
}

function applyTextblockJoin(
  adapter: SchemaAdapter,
  document: Automerge.Doc<RichEditorAutomergeDocument>,
  path: Automerge.Prop[],
  spans: ReturnType<typeof Automerge.spans>,
  step: Transaction["steps"][number],
  pmDoc: ProseMirrorNode,
): boolean {
  const join = textblockJoin(step, pmDoc);
  if (!join) return false;
  const index = pmRangeToAmRange(adapter, spans, { from: join.from, to: join.from }).start;
  const applied = step.apply(pmDoc).doc;
  if (!applied) return false;
  Automerge.joinBlock(document, path, index);
  if (index !== 0 && isLoneDefaultParagraph(applied)) {
    const next = Automerge.spans(document, path);
    if (next[0]?.type === "block") Automerge.joinBlock(document, path, 0);
  }
  return true;
}

function applyTextblockSplit(
  adapter: SchemaAdapter,
  document: Automerge.Doc<RichEditorAutomergeDocument>,
  path: Automerge.Prop[],
  spans: ReturnType<typeof Automerge.spans>,
  step: Transaction["steps"][number],
  pmDoc: ProseMirrorNode,
  suggestion: SuggestionSplit | null,
): boolean {
  const split = textblockSplit(step, pmDoc);
  if (!split) return false;

  const right = automergeBlock(adapter, split.right, split.ancestors);
  if (!right) return false;
  if (suggestion) {
    right.attrs[suggestionSplitAttr] = suggestion.id;
    right.attrs[suggestionSplitKindAttr] = suggestion.kind;
  }
  const index = pmRangeToAmRange(adapter, spans, {
    from: split.from,
    to: split.from,
  }).start;
  const openingNeeded = needsOpeningMarker(adapter, spans, split.ancestors, index);
  const opening = openingNeeded
    ? automergeBlock(adapter, split.left, split.ancestors)
    : null;
  if (openingNeeded && !opening) return false;

  Automerge.splitBlock(document, path, index, right);
  if (opening) Automerge.splitBlock(document, path, 0, opening);
  return true;
}

function isLoneDefaultParagraph(doc: ProseMirrorNode): boolean {
  if (doc.childCount !== 1) return false;
  const only = doc.firstChild;
  if (!only?.isTextblock) return false;
  return doc.type.contentMatch.defaultType === only.type;
}

function isEditorAttributeStep(
  step: Transaction["steps"][number],
  doc: ProseMirrorNode,
): boolean {
  if (step.toJSON().stepType !== "replaceAround") return false;
  const applied = step.apply(doc).doc;
  if (!applied) return false;
  return samePersistedDocument(doc, applied);
}

function samePersistedDocument(left: ProseMirrorNode, right: ProseMirrorNode): boolean {
  if (left.type !== right.type || left.childCount !== right.childCount) return false;
  if (left.isText) {
    if (left.text !== right.text || left.marks.length !== right.marks.length) return false;
    return left.marks.every((mark, index) => mark.eq(right.marks[index]));
  }
  if (!samePersistedAttrs(left.attrs, right.attrs)) return false;
  for (let index = 0; index < left.childCount; index += 1) {
    if (!samePersistedDocument(left.child(index), right.child(index))) return false;
  }
  return true;
}

function samePersistedAttrs(
  left: Record<string, unknown>,
  right: Record<string, unknown>,
): boolean {
  const names = new Set([...Object.keys(left), ...Object.keys(right)]);
  for (const name of names) {
    if (name === "isAmgBlock" || name === "unknownAttrs") continue;
    if (JSON.stringify(left[name]) !== JSON.stringify(right[name])) return false;
  }
  return true;
}

function textDeletion(
  step: Transaction["steps"][number],
  doc: ProseMirrorNode,
): { from: number; to: number } | null {
  if (step.toJSON().stepType !== "replace") return null;
  const replace = step as unknown as {
    from: number;
    to: number;
    slice: { content: { size: number } };
  };
  if (replace.slice.content.size !== 0 || replace.from === replace.to) return null;
  const from = doc.resolve(replace.from);
  const to = doc.resolve(replace.to);
  if (!from.sameParent(to) || !from.parent.isTextblock) return null;
  return { from: replace.from, to: replace.to };
}

function isSingleTextReplace(step: Transaction["steps"][number]): boolean {
  if (step.toJSON().stepType !== "replace") return false;
  const content = (step as unknown as {
    slice: { content: { childCount: number; firstChild: { isText: boolean } | null } };
  }).slice.content;
  return content.childCount === 1 && content.firstChild?.isText === true;
}

function needsOpeningMarker(
  adapter: SchemaAdapter,
  spans: ReturnType<typeof Automerge.spans>,
  ancestors: readonly ProseMirrorNode[],
  index: number,
): boolean {
  const parent = ancestors[ancestors.length - 1];
  if (parent?.type !== adapter.schema.nodes.doc) return false;
  return !blockCovers(spans, index);
}

function blockCovers(
  spans: ReturnType<typeof Automerge.spans>,
  index: number,
): boolean {
  let cursor = 0;
  let covered = false;
  for (const span of spans) {
    if (cursor > index) return covered;
    if (span.type === "text") {
      const text = String(span.value);
      if (cursor + text.length > index) return covered;
      cursor += text.length;
    } else {
      covered = true;
      cursor += 1;
    }
  }
  return covered;
}

function automergeBlock(
  adapter: SchemaAdapter,
  node: ProseMirrorNode,
  ancestors: readonly ProseMirrorNode[],
): AutomergeBlock | null {
  const mapping = mappingFor(adapter, node, ancestors);
  if (!mapping || mapping.isEmbed) return null;

  const attrs = { ...(mapping.attrParsers?.fromProsemirror(node) ?? {}) };
  const unknown = node.attrs.unknownAttrs;
  if (unknown && typeof unknown === "object") {
    const values = unknown as Record<string, Automerge.MaterializeValue>;
    for (const [key, value] of Object.entries(values)) {
      attrs[key] = value;
    }
  }

  const parents: Automerge.ImmutableString[] = [];
  for (const [index, ancestor] of ancestors.entries()) {
    const parentMapping = mappingFor(adapter, ancestor, ancestors.slice(0, index));
    if (!parentMapping) continue;
    parents.push(new Automerge.ImmutableString(parentMapping.blockName));
  }

  return {
    type: new Automerge.ImmutableString(mapping.blockName),
    parents,
    attrs,
    isEmbed: false,
  };
}

function mappingFor(
  adapter: SchemaAdapter,
  node: ProseMirrorNode,
  ancestors: readonly ProseMirrorNode[],
) {
  const matches = adapter.nodeMappings.filter(mapping => mapping.content === node.type);
  const parent = ancestors[ancestors.length - 1];
  return matches.find(mapping => parent && mapping.outer === parent.type) ?? matches[0] ?? null;
}

function applyMappedMarkStep(
  adapter: SchemaAdapter,
  document: Automerge.Doc<RichEditorAutomergeDocument>,
  path: Automerge.Prop[],
  spans: ReturnType<typeof Automerge.spans>,
  step: Transaction["steps"][number],
): boolean | null {
  const stepType = step.toJSON().stepType;
  if (stepType !== "addMark" && stepType !== "removeMark") return null;

  const marked = step as unknown as {
    from: number;
    to: number;
    mark: {
      type: { spec: { inclusive?: boolean } };
      attrs: Record<string, unknown>;
    };
  };
  const mapped = amMarksFromPmMarks(adapter, [marked.mark as never]);
  const names = Object.keys(mapped);
  if (names.length !== 1) return false;

  const range = pmRangeToAmRange(adapter, spans, { from: marked.from, to: marked.to });
  if (!range) return false;
  let { start, end } = range;
  if (start > end) {
    [start, end] = [end, start];
  }
  const expand = marked.mark.type.spec.inclusive === false ? "none" : "both";
  const name = names[0];
  if (!name) return false;
  if (stepType === "addMark") {
    Automerge.mark(document, path, { start, end, expand }, name, mapped[name]);
  } else {
    Automerge.unmark(document, path, { start, end, expand }, name);
  }
  return true;
}

function blockAttrs(
  block: { [key: string]: Automerge.MaterializeValue },
): { [key: string]: Automerge.MaterializeValue } | null {
  const attrs = block.attrs;
  if (!attrs || typeof attrs !== "object" || Array.isArray(attrs)) return null;
  return attrs as { [key: string]: Automerge.MaterializeValue };
}
