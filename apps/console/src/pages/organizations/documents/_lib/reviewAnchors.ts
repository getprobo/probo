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
import { automergeBlockType, readSuggestionBlock } from "@probo/ui/src/RichEditor/suggestionBlock";
import {
  suggestionDeleteName,
  suggestionInsertName,
} from "@probo/ui/src/RichEditor/SuggestionExtension";
import { structuralSuggestionSplit } from "@probo/ui/src/RichEditor/suggestionStructure";
import {
  type SuggestionStyleName,
  suggestionStyles,
} from "@probo/ui/src/RichEditor/suggestionStyle";
import type { Mark, Node as ProseMirrorNode } from "@tiptap/pm/model";

// A card lists this many changes and counts the rest.
export const maxSuggestionChanges = 3;
// Inserted or deleted text past this length is cut, so a pasted page is not
// held in state or drawn on a card.
export const maxChangeTextLength = 280;

export type BlockKind
  = | "paragraph"
    | "heading"
    | "quote"
    | "code"
    | "bulletList"
    | "orderedList"
    | "taskList"
    | "other";

export type BlockLabel = {
  kind: BlockKind;
  level: number | null;
};

export type StructureNode
  = | "table"
    | "row"
    | "column"
    | "image"
    | "divider"
    | "file"
    | "line"
    | "other";

export type SuggestionChange
  = | { kind: "text"; deleted: string; inserted: string }
    | {
      kind: "format";
      style: SuggestionStyleName;
      action: "add" | "remove";
      href: string | null;
      previousHref: string | null;
    }
    | { kind: "block"; before: BlockLabel; after: BlockLabel }
    | { kind: "node"; node: StructureNode; action: "insert" | "delete" };

export type SuggestionAnchor = {
  id: string;
  from: number;
  changes: SuggestionChange[];
  changeCount: number;
};

export type ThreadAnchorSource = {
  id: string;
  anchorStart: string;
  anchorEnd: string;
};

type TextField = {
  text: string;
  parent: ProseMirrorNode | null;
  truncated: boolean;
};

type OpenText = {
  change: { kind: "text"; deleted: string; inserted: string };
  deleted: TextField;
  inserted: TextField;
  lastIndex: number;
};

type Collected = {
  id: string;
  from: number;
  changes: SuggestionChange[];
  keys: Set<string>;
  openText: OpenText | null;
  // Inserted or deleted structure redlines everything inside it too. Those
  // inner marks belong to that one change, so they are skipped.
  coveredUntil: number;
  lineInserted: boolean;
  lineDeleted: boolean;
};

// suggestionAnchors reads every open suggestion from the document as an
// ordered list of distinct changes.
export function suggestionAnchors(doc: ProseMirrorNode): Map<string, SuggestionAnchor> {
  const collected = new Map<string, Collected>();
  let textIndex = 0;

  const touch = (id: string, pos: number): Collected => {
    const current = collected.get(id);
    if (current) return current;
    const next: Collected = {
      id,
      from: pos,
      changes: [],
      keys: new Set(),
      openText: null,
      coveredUntil: -1,
      lineInserted: false,
      lineDeleted: false,
    };
    collected.set(id, next);
    return next;
  };

  doc.descendants((node, pos, parent) => {
    const split = structuralSuggestionSplit(node);
    if (split) {
      const entry = touch(split.id, pos);
      if (pos >= entry.coveredUntil) {
        if (node.isTextblock) {
          if (split.kind === "deletion") entry.lineDeleted = true;
          else entry.lineInserted = true;
        } else {
          const structure = structureNode(node);
          const action = split.kind === "deletion" ? "delete" : "insert";
          addKeyed(entry, `node:${structure}:${action}`, { kind: "node", node: structure, action });
          entry.coveredUntil = pos + node.nodeSize;
        }
      }
    }

    const block = readSuggestionBlock(node);
    if (block) {
      const entry = touch(block.id, pos);
      if (pos >= entry.coveredUntil) {
        const before = blockLabel(block.type, block.attrs.level);
        const after = blockLabel(automergeBlockType(node, parent) ?? "", node.attrs.level);
        addKeyed(entry, `block:${labelKey(before)}:${labelKey(after)}`, { kind: "block", before, after });
      }
    }

    if (!node.isText || !node.text) return;
    textIndex += 1;

    for (const mark of node.marks) {
      const id = typeof mark.attrs.id === "string" ? mark.attrs.id : "";
      if (id === "") continue;
      const isText = mark.type.name === suggestionInsertName || mark.type.name === suggestionDeleteName;
      const style = isText ? null : formatStyle(mark);
      if (!isText && style == null) continue;

      const entry = touch(id, pos);
      if (pos < entry.coveredUntil) continue;

      if (isText) {
        const field = mark.type.name === suggestionDeleteName ? "deleted" : "inserted";
        appendTextRun(entry, field, node.text, parent, textIndex);
      } else if (style != null) {
        const action = mark.attrs.action === "remove" ? "remove" : "add";
        addKeyed(entry, `format:${style}:${action}`, {
          kind: "format",
          style,
          action,
          href: style === "link" ? linkHref(node) : null,
          previousHref: nonEmpty(mark.attrs.href),
        });
      }
    }
  });

  const anchors = new Map<string, SuggestionAnchor>();
  for (const entry of collected.values()) {
    const changes = [...entry.changes];
    if (!changes.some(change => change.kind === "text")) {
      if (entry.lineInserted) changes.push({ kind: "node", node: "line", action: "insert" });
      if (entry.lineDeleted) changes.push({ kind: "node", node: "line", action: "delete" });
    }
    anchors.set(entry.id, {
      id: entry.id,
      from: entry.from,
      changes: changes.slice(0, maxSuggestionChanges),
      changeCount: changes.length,
    });
  }
  return anchors;
}

// threadAnchors resolves each thread cursor to its start in the document. A
// thread whose anchor no longer resolves is left out.
export function threadAnchors(
  handle: RepoCollaborationHandle,
  doc: ProseMirrorNode,
  threads: ReadonlyArray<ThreadAnchorSource>,
): Map<string, number> {
  const positions = new Map<string, number>();
  for (const thread of threads) {
    const range = commentAnchorRange(handle, doc, thread);
    if (range) positions.set(thread.id, range.from);
  }
  return positions;
}

function addKeyed(entry: Collected, key: string, change: SuggestionChange) {
  if (entry.keys.has(key)) return;
  entry.keys.add(key);
  entry.changes.push(change);
}

// Delete and insert runs of one suggestion stay one change while they are
// consecutive text nodes, so a replacement reads as old next to new. Any
// other text in between starts a new change.
function appendTextRun(
  entry: Collected,
  field: "deleted" | "inserted",
  text: string,
  parent: ProseMirrorNode | null,
  textIndex: number,
) {
  const open = entry.openText;
  const contiguous = open != null && textIndex - open.lastIndex <= 1;
  const current = contiguous
    ? open
    : {
        change: { kind: "text" as const, deleted: "", inserted: "" },
        deleted: { text: "", parent: null, truncated: false },
        inserted: { text: "", parent: null, truncated: false },
        lastIndex: textIndex,
      };
  if (!contiguous) {
    entry.changes.push(current.change);
    entry.openText = current;
  }
  appendField(current[field], text, parent);
  current.change[field] = current[field].text;
  current.lastIndex = textIndex;
}

function appendField(field: TextField, text: string, parent: ProseMirrorNode | null) {
  if (field.truncated) return;
  const joined = field.text !== "" && field.parent !== parent
    ? `${field.text}\n${text}`
    : field.text + text;
  field.parent = parent;
  if (joined.length > maxChangeTextLength) {
    field.text = `${joined.slice(0, maxChangeTextLength)}…`;
    field.truncated = true;
    return;
  }
  field.text = joined;
}

function formatStyle(mark: Mark): SuggestionStyleName | null {
  return suggestionStyles.find(style => style.format === mark.type.name)?.name ?? null;
}

function linkHref(node: ProseMirrorNode): string | null {
  return nonEmpty(node.marks.find(mark => mark.type.name === "link")?.attrs.href);
}

function nonEmpty(value: unknown): string | null {
  return typeof value === "string" && value !== "" ? value : null;
}

function structureNode(node: ProseMirrorNode): StructureNode {
  switch (node.type.name) {
    case "table":
      return "table";
    case "tableRow":
      return "row";
    case "tableCell":
    case "tableHeader":
      return "column";
    case "image":
      return "image";
    case "horizontalRule":
      return "divider";
    case "file":
      return "file";
    default:
      return "other";
  }
}

function blockLabel(type: string, level: unknown): BlockLabel {
  switch (type) {
    case "paragraph":
      return { kind: "paragraph", level: null };
    case "heading":
      return { kind: "heading", level: typeof level === "number" ? level : 1 };
    case "blockquote":
      return { kind: "quote", level: null };
    case "code-block":
      return { kind: "code", level: null };
    case "unordered-list-item":
      return { kind: "bulletList", level: null };
    case "ordered-list-item":
      return { kind: "orderedList", level: null };
    case "task-list-item":
      return { kind: "taskList", level: null };
    default:
      return { kind: "other", level: null };
  }
}

function labelKey(label: BlockLabel): string {
  return label.level == null ? label.kind : `${label.kind}${label.level}`;
}
