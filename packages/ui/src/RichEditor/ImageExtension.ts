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

import { mergeAttributes, Node } from "@tiptap/core";
import { Fragment, Slice } from "@tiptap/pm/model";
import { NodeSelection } from "@tiptap/pm/state";
import type { EditorView } from "@tiptap/pm/view";

export type RichEditorPicture = {
  src: string;
  alt?: string;
};

export const pictureContentTypes = ["image/jpeg", "image/png", "image/webp"] as const;

export function isPictureFile(file: File) {
  return pictureContentTypes.some(type => type === file.type);
}

export function insertPicture(view: EditorView, pos: number, picture: RichEditorPicture) {
  const type = view.state.schema.nodes.image;
  if (!type) {
    return;
  }

  const node = type.create({
    src: picture.src,
    alt: picture.alt ?? null,
    title: null,
  });
  const doc = view.state.doc;
  const safePos = Math.max(0, Math.min(pos, doc.content.size));
  const $pos = doc.resolve(safePos);
  let from = safePos;
  let to = safePos;

  if (
    $pos.depth > 0
    && $pos.parent.type.name === "paragraph"
    && $pos.parent.content.size === 0
  ) {
    from = $pos.before($pos.depth);
    to = $pos.after($pos.depth);
  }

  let tr = view.state.tr.replaceRange(
    from,
    to,
    new Slice(Fragment.from(node), 0, 0),
  );
  const insertedAt = tr.mapping.map(from);

  try {
    tr = tr.setSelection(NodeSelection.create(tr.doc, insertedAt));
  } catch {
    // The mapped position is not an image node.
  }

  view.dispatch(tr.scrollIntoView());
}

export const ImageExtension = Node.create({
  name: "image",
  group: "block",
  atom: true,
  draggable: true,
  selectable: true,

  addAttributes() {
    return {
      src: { default: null },
      alt: { default: null },
      title: { default: null },
    };
  },

  parseHTML() {
    return [{ tag: "img[src]" }];
  },

  renderHTML({ HTMLAttributes }) {
    return ["img", mergeAttributes(HTMLAttributes)];
  },
});
