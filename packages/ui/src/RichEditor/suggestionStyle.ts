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
import type { Mark as ProseMirrorMark, Node as ProseMirrorNode } from "@tiptap/pm/model";
import type { EditorState, Transaction } from "@tiptap/pm/state";

import {
  type AutomergeMarkMapping,
  parseJSONObject,
  readStringAttribute,
} from "./automergeMapping";
const suggestionInsertName = "suggestionInsert";
const suggestionDeleteName = "suggestionDelete";

export const suggestionStyles = [
  { name: "bold", format: "suggestionFormatBold", automerge: "suggestion-format-strong", label: "bold" },
  { name: "italic", format: "suggestionFormatItalic", automerge: "suggestion-format-em", label: "italic" },
  { name: "underline", format: "suggestionFormatUnderline", automerge: "suggestion-format-underline", label: "underline" },
  { name: "strike", format: "suggestionFormatStrike", automerge: "suggestion-format-strike", label: "strikethrough" },
  { name: "code", format: "suggestionFormatCode", automerge: "suggestion-format-code", label: "code" },
  { name: "link", format: "suggestionFormatLink", automerge: "suggestion-format-link", label: "link" },
] as const;

export type SuggestionStyleName = (typeof suggestionStyles)[number]["name"];

const styleNames = new Set<string>(suggestionStyles.map(style => style.name));
const formatNames = new Set<string>(suggestionStyles.map(style => style.format));
const styleByFormat = new Map<string, (typeof suggestionStyles)[number]>(
  suggestionStyles.map(style => [style.format, style]),
);
const styleByName = new Map<string, (typeof suggestionStyles)[number]>(
  suggestionStyles.map(style => [style.name, style]),
);

type TextSlice = {
  from: number;
  to: number;
  marks: readonly ProseMirrorMark[];
};

// suggestionFormatExtensions store one suggested style each. The live mark
// shows the suggested result, and this mark remembers how to accept or reject it.
export const suggestionFormatExtensions = suggestionStyles.map(style => Mark.create({
  name: style.format,
  inclusive: false,
  excludes: "",
  addAttributes() {
    return {
      id: { default: null },
      action: { default: "add" },
      href: { default: null },
    };
  },
  automerge: suggestionFormatMapping(style.automerge),
  parseHTML() {
    return [{ tag: `span[data-suggestion-format-${style.name}]` }];
  },
  renderHTML({ HTMLAttributes }: { HTMLAttributes: Record<string, unknown> }) {
    const action = readStringAttribute(HTMLAttributes, "action", "add");
    return ["span", mergeAttributes(HTMLAttributes, {
      [`data-suggestion-format-${style.name}`]: readStringAttribute(HTMLAttributes, "id", ""),
      class: action === "remove" ? "text-red-11" : "text-green-11",
    }), 0];
  },
}));

export function isFormatMark(mark: ProseMirrorMark): boolean {
  return formatNames.has(mark.type.name);
}

// styleMarkStep reports a transaction that toggles bold, italic, or another
// style directly. Suggesting mode turns those into redlines instead.
export function styleMarkStep(step: { toJSON: () => unknown }): boolean {
  const json = step.toJSON();
  if (typeof json !== "object" || json === null) return false;
  const record = json as { stepType?: unknown; mark?: unknown };
  if (record.stepType !== "addMark" && record.stepType !== "removeMark") return false;
  if (typeof record.mark !== "object" || record.mark === null) return false;
  const markType = (record.mark as { type?: unknown }).type;
  return typeof markType === "string" && styleNames.has(markType);
}

// marksForSuggestionText drops suggestion bookkeeping, and drops a style that
// is only present because it was suggested, so typed text does not inherit it.
export function marksForSuggestionText(marks: readonly ProseMirrorMark[]): ProseMirrorMark[] {
  const suggested = new Set<string>();
  for (const mark of marks) {
    const style = styleByFormat.get(mark.type.name);
    if (style && mark.attrs.action !== "remove") suggested.add(style.name);
  }
  return marks.filter((mark) => {
    if (mark.type.name === suggestionInsertName || mark.type.name === suggestionDeleteName) return false;
    if (isFormatMark(mark)) return false;
    return !suggested.has(mark.type.name);
  });
}

export function suggestStyleToggle(
  state: EditorState,
  id: string,
  styleName: SuggestionStyleName,
  from: number,
  to: number,
): Transaction | null {
  return suggestStyle(state, id, styleName, from, to, null);
}

// suggestLink sets or clears a link. A null href removes it.
export function suggestLink(
  state: EditorState,
  id: string,
  from: number,
  to: number,
  href: string | null,
): Transaction | null {
  return suggestStyle(state, id, "link", from, to, href);
}

export function revertFormatSuggestion(tr: Transaction, id: string) {
  const ranges: TextSlice[] = [];
  tr.doc.descendants((node, pos) => {
    if (!node.isText || !node.text) return;
    if (!node.marks.some(mark => isFormatMark(mark) && mark.attrs.id === id)) return;
    ranges.push({ from: pos, to: pos + node.nodeSize, marks: node.marks });
  });
  for (const range of ranges.reverse()) {
    rewriteSlice(tr, range, revertedMarks(tr.doc, range.marks, id));
  }
}

export function reassignFormatMarks(marks: readonly ProseMirrorMark[], fromId: string, toId: string) {
  return marks.map((mark) => {
    if (!isFormatMark(mark) || mark.attrs.id !== fromId) return mark;
    const action = mark.attrs.action === "remove" ? "remove" : "add";
    const href = typeof mark.attrs.href === "string" ? mark.attrs.href : null;
    return mark.type.create({ id: toId, action, href });
  });
}

function suggestStyle(
  state: EditorState,
  id: string,
  styleName: SuggestionStyleName,
  from: number,
  to: number,
  linkHref: string | null,
): Transaction | null {
  const style = styleByName.get(styleName);
  const markType = state.schema.marks[styleName];
  if (!style || !markType || from >= to) return null;

  const slices = textSlices(state.doc, from, to);
  if (slices.length === 0 || slices.some(slice => foreignSuggestion(slice.marks, style.format, id))) {
    return null;
  }
  if (styleName === "code" && slices.some(slice => foreignFormat(slice.marks, id))) {
    return null;
  }

  const editable = slices.filter(slice => suggestionId(slice.marks, suggestionDeleteName) !== id);
  if (editable.length === 0) return null;

  const removing = styleName === "link"
    ? linkHref === null
    : editable.some(slice => effectivelyHas(slice.marks, style));
  if (styleName === "link" && linkHref !== null && editable.every(slice => linkHrefOf(slice.marks) === linkHref)) {
    return null;
  }

  const tr = state.tr;
  let changed = false;
  for (const slice of editable.slice().reverse()) {
    const next = nextMarks(state, slice.marks, style, id, removing, linkHref);
    if (!next || sameMarks(slice.marks, next)) continue;
    rewriteSlice(tr, slice, next);
    changed = true;
  }
  if (!changed) return null;
  return tr;
}

function nextMarks(
  state: EditorState,
  marks: readonly ProseMirrorMark[],
  style: (typeof suggestionStyles)[number],
  id: string,
  removing: boolean,
  linkHref: string | null,
): ProseMirrorMark[] | null {
  const formatType = state.schema.marks[style.format];
  const markType = state.schema.marks[style.name];
  if (!formatType || !markType) return null;

  if (removing && !effectivelyHas(marks, style)) return [...marks];
  if (!removing && style.name === "link" && linkHref != null && linkHrefOf(marks) === linkHref && !formatOf(marks, style.format)) {
    return [...marks];
  }

  const ownInsert = suggestionId(marks, suggestionInsertName) === id;
  const format = formatOf(marks, style.format);
  let next = marks.filter(mark => mark.type !== formatType && mark.type !== markType);

  if (removing) {
    if (format?.id === id && format.action === "add") {
      if (style.name === "link" && format.href) {
        next = [...next, markType.create({ href: format.href })];
      }
      return finishCode(state, marks, next, style, id, ownInsert);
    }
    if (!ownInsert) {
      const href = style.name === "link" ? linkHrefOf(marks) : null;
      next = [...next, formatType.create({ id, action: "remove", href })];
    }
    return finishCode(state, marks, next, style, id, ownInsert);
  }

  if (format?.id === id && format.action === "remove") {
    next = [...next, restoredMark(state, style.name, format.href)];
    return finishCode(state, marks, next, style, id, ownInsert);
  }

  const created = style.name === "link"
    ? markType.create({ href: linkHref ?? "" })
    : markType.create();
  next = [...next, created];
  if (!ownInsert) {
    const href = style.name === "link" ? linkHrefOf(marks) || null : null;
    next = [...next, formatType.create({ id, action: "add", href })];
  }
  return finishCode(state, marks, next, style, id, ownInsert);
}

// Suggesting code replaces the other styles on that text. Rejecting the code
// suggestion puts those styles back.
function finishCode(
  state: EditorState,
  original: readonly ProseMirrorMark[],
  next: ProseMirrorMark[],
  style: (typeof suggestionStyles)[number],
  id: string,
  ownInsert: boolean,
): ProseMirrorMark[] {
  if (style.name !== "code" || !next.some(mark => mark.type.name === "code")) return next;

  const cancelled = new Set<string>();
  for (const mark of next) {
    const format = styleByFormat.get(mark.type.name);
    if (!format || format.name === "code" || mark.attrs.action !== "add" || mark.attrs.id !== id) continue;
    cancelled.add(format.name);
  }

  const kept: ProseMirrorMark[] = [];
  for (const mark of next) {
    const format = styleByFormat.get(mark.type.name);
    if (format && format.name !== "code") continue;
    const other = styleByName.get(mark.type.name);
    if (!other || other.name === "code") {
      kept.push(mark);
      continue;
    }
    if (ownInsert || cancelled.has(other.name)) continue;
    const formatType = state.schema.marks[other.format];
    if (!formatType) continue;
    const href = other.name === "link" ? linkHrefOf(original) : null;
    kept.push(formatType.create({ id, action: "remove", href }));
  }
  return kept;
}

function revertedMarks(doc: ProseMirrorNode, marks: readonly ProseMirrorMark[], id: string): ProseMirrorMark[] {
  const dropped = new Set<string>();
  const restored: ProseMirrorMark[] = [];
  for (const mark of marks) {
    if (!isFormatMark(mark) || mark.attrs.id !== id) continue;
    const style = styleByFormat.get(mark.type.name);
    if (!style) continue;
    dropped.add(style.name);
    if (mark.attrs.action === "remove" || (style.name === "link" && mark.attrs.action === "add" && mark.attrs.href)) {
      if (mark.attrs.action === "add" && style.name !== "link") continue;
      const type = doc.type.schema.marks[style.name];
      if (!type) continue;
      restored.push(style.name === "link"
        ? type.create({ href: typeof mark.attrs.href === "string" ? mark.attrs.href : "" })
        : type.create());
    }
  }
  return [
    ...marks.filter(mark => !(isFormatMark(mark) && mark.attrs.id === id) && !dropped.has(mark.type.name)),
    ...restored,
  ];
}

function restoredMark(state: EditorState, name: string, href: string | null): ProseMirrorMark {
  const type = state.schema.marks[name];
  if (!type) throw new Error(`missing ${name} mark`);
  return name === "link" ? type.create({ href: href ?? "" }) : type.create();
}

function effectivelyHas(marks: readonly ProseMirrorMark[], style: (typeof suggestionStyles)[number]): boolean {
  const format = formatOf(marks, style.format);
  if (format?.action === "add") return true;
  if (format?.action === "remove") return false;
  return marks.some(mark => mark.type.name === style.name);
}

function foreignFormat(marks: readonly ProseMirrorMark[], id: string): boolean {
  return marks.some(mark => isFormatMark(mark) && typeof mark.attrs.id === "string" && mark.attrs.id !== "" && mark.attrs.id !== id);
}

function foreignSuggestion(marks: readonly ProseMirrorMark[], formatName: string, id: string): boolean {
  const insert = suggestionId(marks, suggestionInsertName);
  const deleted = suggestionId(marks, suggestionDeleteName);
  if ((insert && insert !== id) || (deleted && deleted !== id)) return true;
  const format = formatOf(marks, formatName);
  return format != null && format.id !== id;
}

function formatOf(marks: readonly ProseMirrorMark[], formatName: string) {
  const mark = marks.find(item => item.type.name === formatName);
  if (!mark || typeof mark.attrs.id !== "string" || mark.attrs.id === "") return null;
  const action = mark.attrs.action === "remove" ? "remove" : "add";
  const href = typeof mark.attrs.href === "string" && mark.attrs.href !== "" ? mark.attrs.href : null;
  return { id: mark.attrs.id, action, href };
}

function suggestionId(marks: readonly ProseMirrorMark[], name: string): string | null {
  for (const mark of marks) {
    if (mark.type.name !== name) continue;
    if (typeof mark.attrs.id === "string" && mark.attrs.id !== "") return mark.attrs.id;
  }
  return null;
}

function linkHrefOf(marks: readonly ProseMirrorMark[]): string {
  const link = marks.find(mark => mark.type.name === "link");
  return typeof link?.attrs.href === "string" ? link.attrs.href : "";
}

function textSlices(doc: ProseMirrorNode, from: number, to: number): TextSlice[] {
  const slices: TextSlice[] = [];
  doc.nodesBetween(from, to, (node, pos) => {
    if (!node.isText || !node.text) return;
    const start = Math.max(from, pos);
    const end = Math.min(to, pos + node.nodeSize);
    if (start >= end) return;
    slices.push({ from: start, to: end, marks: node.marks });
  });
  return slices;
}

function rewriteSlice(tr: Transaction, slice: TextSlice, marks: readonly ProseMirrorMark[]) {
  const text = tr.doc.textBetween(slice.from, slice.to);
  if (text === "") return;
  tr.replaceWith(slice.from, slice.to, tr.doc.type.schema.text(text, marks));
}

function sameMarks(left: readonly ProseMirrorMark[], right: readonly ProseMirrorMark[]): boolean {
  if (left.length !== right.length) return false;
  return left.every((mark, index) => mark.eq(right[index]));
}

function suggestionFormatMapping(markName: string): () => AutomergeMarkMapping {
  return () => ({
    markName,
    parsers: {
      fromAutomerge: (value) => {
        if (typeof value !== "string") return { id: "", action: "add", href: null };
        const parsed = parseJSONObject(value);
        const action = readStringAttribute(parsed, "action", "add");
        const href = readStringAttribute(parsed, "href", "");
        return {
          id: readStringAttribute(parsed, "id", ""),
          action: action === "remove" ? "remove" : "add",
          href: href === "" ? null : href,
        };
      },
      fromProsemirror: (mark: ProseMirrorMark) => {
        const href = readStringAttribute(mark.attrs, "href", "");
        const payload: Record<string, string> = {
          id: readStringAttribute(mark.attrs, "id", ""),
          action: readStringAttribute(mark.attrs, "action", "add") === "remove" ? "remove" : "add",
        };
        if (href !== "") payload.href = href;
        return JSON.stringify(payload);
      },
    },
  });
}
