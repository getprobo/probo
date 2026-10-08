// Copyright (c) 2026 Probo Inc <hello@probo.com>.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

import { mergeAttributes, Node } from "@tiptap/core";
import { ReactNodeViewRenderer } from "@tiptap/react";

import { attachmentPath } from "./attachments";
import { ImageNodeView } from "./ImageNodeView";

export const ImageExtension = Node.create({
  name: "image",
  group: "block",
  atom: true,
  draggable: true,
  addAttributes() {
    return {
      fileId: {
        default: null,
        parseHTML: element => element.getAttribute("data-file-id"),
      },
      src: {
        default: null,
      },
      alt: {
        default: null,
      },
      title: {
        default: null,
      },
    };
  },
  parseHTML() {
    return [{ tag: "img[src]" }];
  },
  renderHTML({ node }) {
    const fileId = typeof node.attrs.fileId === "string" ? node.attrs.fileId : null;
    const src = fileId
      ? attachmentPath(fileId)
      : typeof node.attrs.src === "string" ? node.attrs.src : null;
    const alt = typeof node.attrs.alt === "string" ? node.attrs.alt : null;
    const title = typeof node.attrs.title === "string" ? node.attrs.title : null;

    return ["img", mergeAttributes({
      "src": src,
      "alt": alt,
      "title": title,
      "data-file-id": fileId,
    })];
  },
  addNodeView() {
    return ReactNodeViewRenderer(ImageNodeView);
  },
});
