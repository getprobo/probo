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

import {
  suggestionDeleteName,
  suggestionInsertName,
} from "@probo/ui/src/RichEditor/SuggestionExtension";
import { structuralSuggestionSplit } from "@probo/ui/src/RichEditor/suggestionStructure";
import { suggestionStyles } from "@probo/ui/src/RichEditor/suggestionStyle";
import type { Node as ProseMirrorNode } from "@tiptap/pm/model";

export type SuggestionPreviewPart = {
  kind: "inserted" | "deleted" | "kept";
  text: string;
};

export type SuggestionStructurePreview = {
  action: "add" | "remove";
  target: "row" | "column" | "table";
  text: string;
};

export type SuggestionFormatPreview = {
  adds: string[];
  removes: string[];
};

export type SuggestionPreview = {
  from: number;
  inserted: string;
  deleted: string;
  parts: SuggestionPreviewPart[];
  structure: SuggestionStructurePreview | null;
  formats: SuggestionFormatPreview | null;
};

type PieceFormat = {
  id: string;
  action: "add" | "remove";
  style: string;
};

type Piece = {
  pos: number;
  text: string;
  insertIds: string[];
  deleteIds: string[];
  formats: PieceFormat[];
};

type BlockChunk = {
  pos: number;
  splitId: string | null;
  splitKind: string | null;
  // The split belongs to an enclosing row, cell or table, not to a line break.
  inherited: boolean;
  pieces: Piece[];
};

// Groups every run of one suggestion in document order. Text that stays
// between two pieces of the same suggestion stays between them on the card.
// A blank inserted line, and a deleted line break, are part of the change.
export function collectSuggestionAnchors(doc: ProseMirrorNode) {
  const chunks: BlockChunk[] = [];

  doc.descendants((node, pos) => {
    if (node.isTextblock) {
      const ownId = suggestionSplitId(node);
      const enclosing = ownId ? null : enclosingSplit(doc, pos);
      chunks.push({
        pos,
        splitId: ownId ?? enclosing?.id ?? null,
        splitKind: ownId ? suggestionSplitKind(node) : enclosing?.kind ?? null,
        inherited: enclosing != null,
        pieces: [],
      });
      return;
    }
    const chunk = chunks[chunks.length - 1];
    if (!chunk || !node.isText || !node.text) return;
    chunk.pieces.push({
      pos,
      text: node.text,
      insertIds: markIds(node, suggestionInsertName),
      deleteIds: markIds(node, suggestionDeleteName),
      formats: formatMarks(node),
    });
  });

  const previews = new Map<string, SuggestionPreview>();
  const ids = new Set<string>();
  for (const block of chunks) {
    if (block.splitId && (block.splitKind === "insertion" || block.splitKind === "deletion")) {
      ids.add(block.splitId);
    }
    for (const piece of block.pieces) {
      for (const id of piece.insertIds) ids.add(id);
      for (const id of piece.deleteIds) ids.add(id);
      for (const format of piece.formats) ids.add(format.id);
    }
  }

  for (const id of ids) {
    const structure = structureFor(doc, id);
    const formats = formatsFor(chunks, id);
    const parts = structure ? [] : previewParts(chunks, id);
    if (parts.length === 0 && !structure && !formats) continue;
    const from = structure?.from
      ?? formatPosition(chunks, id)
      ?? chunks.find(block => blockInvolves(block, id))?.pos;
    if (from == null) continue;
    previews.set(id, {
      from,
      inserted: structure ? "" : parts.filter(part => part.kind === "inserted").map(part => part.text).join(""),
      deleted: structure ? "" : parts.filter(part => part.kind === "deleted").map(part => part.text).join(""),
      parts,
      structure: structure ? { action: structure.action, target: structure.target, text: structure.text } : null,
      formats,
    });
  }
  return previews;
}

function previewParts(blocks: BlockChunk[], id: string): SuggestionPreviewPart[] {
  const first = blocks.findIndex(block => blockInvolves(block, id));
  if (first < 0) return [];
  let last = first;
  for (let index = blocks.length - 1; index > first; index -= 1) {
    if (blockInvolves(blocks[index], id)) {
      last = index;
      break;
    }
  }

  const only = blocks[first];
  if (
    first === last
    && only.splitId === id
    && only.splitKind === "insertion"
    && only.pieces.every(piece => !piece.insertIds.includes(id) && !piece.deleteIds.includes(id))
  ) {
    return [{ kind: "inserted", text: "\n" }];
  }

  const parts: SuggestionPreviewPart[] = [];
  for (let index = first; index <= last; index += 1) {
    const block = blocks[index];
    if (index > first) {
      pushPart(parts, lineKind(block, id), "\n");
    } else if (block.splitId === id && block.splitKind === "deletion" && !block.inherited) {
      pushPart(parts, "deleted", "\n");
    }
    for (const piece of block.pieces) {
      if (piece.insertIds.includes(id)) pushPart(parts, "inserted", piece.text);
      else if (piece.deleteIds.includes(id)) pushPart(parts, "deleted", piece.text);
      else pushPart(parts, "kept", piece.text);
    }
  }

  while (parts[0]?.kind === "kept") parts.shift();
  while (parts[parts.length - 1]?.kind === "kept") parts.pop();
  return parts;
}

function lineKind(block: BlockChunk, id: string): SuggestionPreviewPart["kind"] {
  if (block.splitId !== id) return "kept";
  if (block.splitKind === "insertion") return "inserted";
  if (block.splitKind === "deletion") return "deleted";
  return "kept";
}

function blockInvolves(block: BlockChunk, id: string): boolean {
  if (block.splitId === id && (block.splitKind === "insertion" || block.splitKind === "deletion")) {
    return true;
  }
  return block.pieces.some(piece => piece.insertIds.includes(id) || piece.deleteIds.includes(id));
}

function pushPart(
  parts: SuggestionPreviewPart[],
  kind: SuggestionPreviewPart["kind"],
  text: string,
) {
  if (text === "") return;
  const last = parts[parts.length - 1];
  if (last && last.kind === kind) {
    last.text += text;
    return;
  }
  parts.push({ kind, text });
}

const styleOrder = ["bold", "italic", "underline", "strikethrough", "code", "link"];
const styleByFormat = new Map<string, string>(suggestionStyles.map(style => [style.format, style.label]));

function formatMarks(node: ProseMirrorNode): PieceFormat[] {
  const formats: PieceFormat[] = [];
  for (const mark of node.marks) {
    const style = styleByFormat.get(mark.type.name);
    if (!style || typeof mark.attrs.id !== "string" || mark.attrs.id === "") continue;
    formats.push({
      id: mark.attrs.id,
      action: mark.attrs.action === "remove" ? "remove" : "add",
      style,
    });
  }
  return formats;
}

function formatsFor(blocks: BlockChunk[], id: string): SuggestionFormatPreview | null {
  const adds = new Set<string>();
  const removes = new Set<string>();
  for (const block of blocks) {
    for (const piece of block.pieces) {
      for (const format of piece.formats) {
        if (format.id !== id) continue;
        if (format.action === "remove") removes.add(format.style);
        else adds.add(format.style);
      }
    }
  }
  if (adds.size === 0 && removes.size === 0) return null;
  return {
    adds: styleOrder.filter(style => adds.has(style)),
    removes: styleOrder.filter(style => removes.has(style)),
  };
}

function formatPosition(blocks: BlockChunk[], id: string): number | null {
  for (const block of blocks) {
    for (const piece of block.pieces) {
      if (piece.formats.some(format => format.id === id)) return piece.pos;
    }
  }
  return null;
}

// structureFor describes a suggested table, row, or column. The outermost
// block wins, so an added table is not also reported as its rows.
function structureFor(
  doc: ProseMirrorNode,
  id: string,
): (SuggestionStructurePreview & { from: number }) | null {
  const hits: Array<{ pos: number; rank: number; kind: "insertion" | "deletion"; node: ProseMirrorNode }> = [];
  doc.descendants((node, pos) => {
    const split = structuralSuggestionSplit(node);
    if (!split || split.id !== id || split.kind === "structural") return;
    const rank = structureRank(node.type.name);
    if (rank == null || (split.kind !== "insertion" && split.kind !== "deletion")) return;
    hits.push({ pos, rank, kind: split.kind, node });
  });
  if (hits.length === 0) return null;
  const rank = Math.min(...hits.map(hit => hit.rank));
  const chosen = hits.filter(hit => hit.rank === rank);
  const kind = chosen[0]?.kind;
  if (!kind) return null;
  const texts: string[] = [];
  for (const hit of chosen) {
    hit.node.descendants((child) => {
      if (child.isText && child.text) texts.push(child.text);
    });
  }
  return {
    action: kind === "deletion" ? "remove" : "add",
    target: rank === 0 ? "table" : rank === 1 ? "row" : "column",
    text: texts.join(" ").replace(/\s+/g, " ").trim(),
    from: Math.min(...chosen.map(hit => hit.pos)),
  };
}

function structureRank(name: string): number | null {
  if (name === "table") return 0;
  if (name === "tableRow") return 1;
  if (name === "tableCell" || name === "tableHeader") return 2;
  return null;
}

function markIds(node: ProseMirrorNode, name: string): string[] {
  const ids: string[] = [];
  for (const mark of node.marks) {
    if (mark.type.name !== name) continue;
    if (typeof mark.attrs.id === "string" && mark.attrs.id !== "") ids.push(mark.attrs.id);
  }
  return ids;
}

// enclosingSplit finds the insertion or deletion split of the nearest row,
// cell or table around a textblock. Their first paragraph has no split of its
// own because it is stored inside the enclosing block.
function enclosingSplit(
  doc: ProseMirrorNode,
  pos: number,
): { id: string; kind: string } | null {
  const $pos = doc.resolve(pos);
  for (let depth = $pos.depth; depth > 0; depth -= 1) {
    const node = $pos.node(depth);
    const id = suggestionSplitId(node);
    const kind = suggestionSplitKind(node);
    if (id && (kind === "insertion" || kind === "deletion")) return { id, kind };
  }
  return null;
}

function suggestionSplitId(node: ProseMirrorNode): string | null {
  const unknown = nodeUnknownAttrs(node);
  if (!unknown || typeof unknown !== "object") return null;
  const id = (unknown as { suggestionSplit?: unknown }).suggestionSplit;
  return typeof id === "string" && id !== "" ? id : null;
}

function suggestionSplitKind(node: ProseMirrorNode): string | null {
  const unknown = nodeUnknownAttrs(node);
  if (!unknown || typeof unknown !== "object") return null;
  const kind = (unknown as { splitKind?: unknown }).splitKind;
  return kind === "insertion" || kind === "structural" || kind === "deletion" ? kind : null;
}

function nodeUnknownAttrs(node: { attrs: { unknownAttrs?: unknown } }): unknown {
  return node.attrs.unknownAttrs;
}
