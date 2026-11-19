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

import type { EditorState } from "@tiptap/pm/state";

import type { BlockTypeId } from "../BlockMenu/blockItems";

const wrapperTypes = new Set<string>(["bulletList", "orderedList", "taskList", "blockquote"]);

// currentBlockType names the block the selection starts in, as the toolbar's
// turn-into menu lists it. The innermost list or blockquote wins over the
// text block inside it. Blocks the menu cannot produce, such as a Mermaid
// diagram or a fifth-level heading, return null.
export function currentBlockType(state: EditorState): BlockTypeId | null {
  const { $from } = state.selection;

  for (let depth = $from.depth; depth > 0; depth--) {
    const name = $from.node(depth).type.name;
    if (wrapperTypes.has(name)) {
      return name as BlockTypeId;
    }
  }

  const block = $from.parent;
  switch (block.type.name) {
    case "paragraph":
      return "paragraph";
    case "codeBlock":
      return block.attrs.language === "mermaid" ? null : "codeBlock";
    case "heading": {
      const level = block.attrs.level as number;
      return level >= 1 && level <= 4 ? `heading${level}` as BlockTypeId : null;
    }
    default:
      return null;
  }
}
