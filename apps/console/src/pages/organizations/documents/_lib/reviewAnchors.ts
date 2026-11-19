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
import { readSuggestionBlock } from "@probo/ui/src/RichEditor/suggestionBlock";
import {
  suggestionDeleteName,
  suggestionInsertName,
} from "@probo/ui/src/RichEditor/SuggestionExtension";
import { structuralSuggestionSplit } from "@probo/ui/src/RichEditor/suggestionStructure";
import { isFormatMark } from "@probo/ui/src/RichEditor/suggestionStyle";
import type { Node as ProseMirrorNode } from "@tiptap/pm/model";

export type SuggestionAnchor = {
  id: string;
  from: number;
  inserted: string;
  deleted: string;
  formatting: boolean;
  structure: boolean;
};

export type ThreadAnchorSource = {
  id: string;
  anchorStart: string;
  anchorEnd: string;
};

type TextRun = {
  text: string;
  parent: ProseMirrorNode | null;
};

type Collected = SuggestionAnchor & {
  insertRun: TextRun;
  deleteRun: TextRun;
};

// suggestionAnchors reads every open suggestion from the document. Text runs
// in different blocks are joined by a line break so a card shows the lines.
export function suggestionAnchors(doc: ProseMirrorNode): Map<string, SuggestionAnchor> {
  const collected = new Map<string, Collected>();

  const touch = (id: string, pos: number): Collected => {
    const current = collected.get(id);
    if (current) return current;
    const next: Collected = {
      id,
      from: pos,
      inserted: "",
      deleted: "",
      formatting: false,
      structure: false,
      insertRun: { text: "", parent: null },
      deleteRun: { text: "", parent: null },
    };
    collected.set(id, next);
    return next;
  };

  doc.descendants((node, pos, parent) => {
    const split = structuralSuggestionSplit(node);
    if (split) {
      const anchor = touch(split.id, pos);
      if (split.kind === "structural" || !node.isTextblock) anchor.structure = true;
    }
    const block = readSuggestionBlock(node);
    if (block) touch(block.id, pos).structure = true;
    if (!node.isText || !node.text) return;

    for (const mark of node.marks) {
      const id = typeof mark.attrs.id === "string" ? mark.attrs.id : "";
      if (id === "") continue;
      if (mark.type.name === suggestionInsertName) {
        appendRun(touch(id, pos).insertRun, node.text, parent);
      } else if (mark.type.name === suggestionDeleteName) {
        appendRun(touch(id, pos).deleteRun, node.text, parent);
      } else if (isFormatMark(mark)) {
        touch(id, pos).formatting = true;
      }
    }
  });

  const anchors = new Map<string, SuggestionAnchor>();
  for (const { insertRun, deleteRun, ...anchor } of collected.values()) {
    anchors.set(anchor.id, { ...anchor, inserted: insertRun.text, deleted: deleteRun.text });
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

function appendRun(run: TextRun, text: string, parent: ProseMirrorNode | null) {
  if (run.text !== "" && run.parent !== parent) run.text += "\n";
  run.text += text;
  run.parent = parent;
}
