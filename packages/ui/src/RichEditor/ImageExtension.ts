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
import { ReactNodeViewRenderer } from "@tiptap/react";

import { attachmentParentID, attachmentPath, fileIDFromAttachmentURL } from "./attachments";
import { ImageNodeView } from "./ImageNodeView";
import { imageWidthPercent } from "./imageWidth";

export const ImageExtension = Node.create({
  name: "image",
  group: "block",
  atom: true,
  draggable: true,
  addAttributes() {
    return {
      fileId: {
        default: null,
        parseHTML: (element) => {
          const id = element.getAttribute("data-file-id")?.trim() ?? "";

          return id === "" ? null : id;
        },
      },
      src: {
        default: null,
        parseHTML: () => null,
      },
      alt: {
        default: null,
      },
      title: {
        default: null,
      },
      width: {
        default: null,
        parseHTML: element => imageWidthPercent(element.getAttribute("width")),
        renderHTML: (attributes) => {
          if (typeof attributes.width !== "number") {
            return {};
          }

          return { width: `${attributes.width}%` };
        },
      },
    };
  },
  parseHTML() {
    return [{
      tag: "img[src], img[data-file-id]",
      getAttrs: (element) => {
        if (!(element instanceof HTMLElement)) {
          return false;
        }

        const fileId = element.getAttribute("data-file-id")?.trim()
          || fileIDFromAttachmentURL(element.getAttribute("src"));
        if (!fileId) {
          return false;
        }

        return {
          fileId,
          width: imageWidthPercent(element.getAttribute("width")),
        };
      },
    }];
  },
  renderHTML({ node }) {
    const fileId = typeof node.attrs.fileId === "string" ? node.attrs.fileId : null;
    const src = fileId
      ? attachmentPath(fileId, attachmentParentID(this.editor))
      : typeof node.attrs.src === "string" ? node.attrs.src : null;
    const alt = typeof node.attrs.alt === "string" ? node.attrs.alt : null;
    const title = typeof node.attrs.title === "string" ? node.attrs.title : null;
    const width = typeof node.attrs.width === "number" ? `${node.attrs.width}%` : null;

    return ["img", mergeAttributes({
      "src": src,
      "alt": alt,
      "title": title,
      "width": width,
      "data-file-id": fileId,
    })];
  },
  addNodeView() {
    return ReactNodeViewRenderer(ImageNodeView);
  },
});
