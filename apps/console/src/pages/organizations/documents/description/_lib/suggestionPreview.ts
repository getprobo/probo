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
import type { Node as ProseMirrorNode } from "@tiptap/pm/model";

export type SuggestionPreviewPart = {
  kind: "inserted" | "deleted" | "kept";
  text: string;
};

export type SuggestionPreview = {
  from: number;
  inserted: string;
  deleted: string;
  parts: SuggestionPreviewPart[];
};

type Piece = {
  text: string;
  insertIds: string[];
  deleteIds: string[];
};

type BlockChunk = {
  pos: number;
  splitId: string | null;
  splitKind: string | null;
  pieces: Piece[];
};

// Groups every run of one suggestion in document order. Text that stays
// between two pieces of the same suggestion stays between them on the card.
// A blank inserted line is part of the change.
export function collectSuggestionAnchors(doc: ProseMirrorNode) {
  const chunks: BlockChunk[] = [];

  doc.descendants((node, pos) => {
    if (node.isTextblock) {
      chunks.push({
        pos,
        splitId: suggestionSplitId(node),
        splitKind: suggestionSplitKind(node),
        pieces: [],
      });
      return;
    }
    const chunk = chunks[chunks.length - 1];
    if (!chunk || !node.isText || !node.text) return;
    chunk.pieces.push({
      text: node.text,
      insertIds: markIds(node, suggestionInsertName),
      deleteIds: markIds(node, suggestionDeleteName),
    });
  });

  const previews = new Map<string, SuggestionPreview>();
  const ids = new Set<string>();
  for (const block of chunks) {
    if (block.splitId && block.splitKind === "insertion") ids.add(block.splitId);
    for (const piece of block.pieces) {
      for (const id of piece.insertIds) ids.add(id);
      for (const id of piece.deleteIds) ids.add(id);
    }
  }

  for (const id of ids) {
    const parts = previewParts(chunks, id);
    if (parts.length === 0) continue;
    const from = chunks.find(block => blockInvolves(block, id))?.pos;
    if (from == null) continue;
    previews.set(id, {
      from,
      inserted: parts.filter(part => part.kind === "inserted").map(part => part.text).join(""),
      deleted: parts.filter(part => part.kind === "deleted").map(part => part.text).join(""),
      parts,
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
      const ownsLine = block.splitId === id && block.splitKind === "insertion";
      pushPart(parts, ownsLine ? "inserted" : "kept", "\n");
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

function blockInvolves(block: BlockChunk, id: string): boolean {
  if (block.splitId === id && block.splitKind === "insertion") return true;
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

function markIds(node: ProseMirrorNode, name: string): string[] {
  const ids: string[] = [];
  for (const mark of node.marks) {
    if (mark.type.name !== name) continue;
    if (typeof mark.attrs.id === "string" && mark.attrs.id !== "") ids.push(mark.attrs.id);
  }
  return ids;
}

function suggestionSplitId(node: ProseMirrorNode): string | null {
  const unknown = node.attrs.unknownAttrs;
  if (!unknown || typeof unknown !== "object") return null;
  const id = (unknown as { suggestionSplit?: unknown }).suggestionSplit;
  return typeof id === "string" && id !== "" ? id : null;
}

function suggestionSplitKind(node: ProseMirrorNode): string | null {
  const unknown = node.attrs.unknownAttrs;
  if (!unknown || typeof unknown !== "object") return null;
  const kind = (unknown as { splitKind?: unknown }).splitKind;
  return kind === "insertion" || kind === "structural" ? kind : null;
}
