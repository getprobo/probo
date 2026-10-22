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

import { Extension } from "@tiptap/core";
import type { Node as ProseMirrorNode } from "@tiptap/pm/model";
import { Plugin, PluginKey } from "@tiptap/pm/state";
import { Decoration, DecorationSet } from "@tiptap/pm/view";

export type RichEditorHistoryHighlight = {
  kind: "INSERT" | "DELETE";
  from: number;
  to: number;
  text: string;
};

const historyPluginKey = new PluginKey("document-history");

const insertStyle = "background-color: #b7e1cd";
const deleteStyle = "background-color: #f6c9c6; text-decoration: line-through; color: #3c4043";

export function createRichEditorHistoryExtension(
  highlights: RichEditorHistoryHighlight[],
): Extension {
  return Extension.create({
    name: "documentHistory",

    addProseMirrorPlugins() {
      return [
        new Plugin({
          key: historyPluginKey,
          props: {
            decorations(state) {
              return decorationsForHighlights(state.doc, highlights);
            },
          },
        }),
      ];
    },
  });
}

export function decorationsForHighlights(
  doc: ProseMirrorNode,
  highlights: RichEditorHistoryHighlight[],
): DecorationSet {
  const decorations: Decoration[] = [];
  const documentEnd = doc.content.size;

  for (const highlight of highlights) {
    if (highlight.kind === "DELETE") {
      if (!highlight.text) continue;

      const at = clampPosition(offsetToPosition(doc, highlight.from), 0, documentEnd);
      decorations.push(
        Decoration.widget(
          at,
          () => {
            const deleted = document.createElement("span");
            deleted.textContent = highlight.text;
            deleted.setAttribute("style", deleteStyle);
            return deleted;
          },
          { side: -1, ignoreSelection: true },
        ),
      );
      continue;
    }

    if (highlight.from >= highlight.to) continue;

    const from = clampPosition(offsetToPosition(doc, highlight.from), 0, documentEnd);
    const to = clampPosition(offsetToPosition(doc, highlight.to), 0, documentEnd);
    if (from >= to) continue;

    decorations.push(
      Decoration.inline(from, to, { style: insertStyle }),
    );
  }

  return DecorationSet.create(doc, decorations);
}

function offsetToPosition(doc: ProseMirrorNode, offset: number): number {
  let remaining = offset;
  let position = doc.content.size;
  let found = false;

  doc.descendants((node, pos) => {
    if (found) return false;

    if (node.isText) {
      const length = node.text?.length ?? 0;
      if (remaining <= length) {
        position = pos + remaining;
        found = true;
      } else {
        remaining -= length;
      }
      return false;
    }

    if (!countsAsBlockBoundary(node)) return;

    if (remaining === 0) {
      position = node.isTextblock ? pos + 1 : pos;
      found = true;
      return false;
    }

    remaining -= 1;
    return !node.isLeaf;
  });

  return position;
}

function countsAsBlockBoundary(node: ProseMirrorNode): boolean {
  return node.isBlock || node.type.name === "hardBreak";
}

function clampPosition(value: number, minimum: number, maximum: number): number {
  return Math.min(Math.max(value, minimum), maximum);
}
