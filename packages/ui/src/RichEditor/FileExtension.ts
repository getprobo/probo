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

import { Node } from "@tiptap/core";
import { ReactNodeViewRenderer } from "@tiptap/react";

import { attachmentParentID, attachmentPath, fileIDFromAttachmentURL } from "./attachments";
import { FileNodeView } from "./FileNodeView";

function fileIdFromHTML(element: { getAttribute(name: string): string | null }): string | null {
  const id = element.getAttribute("data-file-id")?.trim() ?? "";
  if (id !== "") {
    return id;
  }

  return fileIDFromAttachmentURL(element.getAttribute("href"));
}

export const FileExtension = Node.create({
  name: "file",
  group: "block",
  atom: true,
  draggable: true,
  addAttributes() {
    return {
      fileId: {
        default: null,
        parseHTML: element => fileIdFromHTML(element),
      },
      fileName: {
        default: "",
        parseHTML: element => element.textContent,
      },
      mimeType: { default: "" },
      size: {
        default: 0,
        parseHTML: (element) => {
          const raw = element.getAttribute("data-size")?.trim() ?? "";
          const size = Number(raw);
          if (raw === "" || !Number.isFinite(size) || size < 0) {
            return 0;
          }

          return size;
        },
      },
    };
  },
  parseHTML() {
    return [
      {
        tag: "a[data-file-id]",
        // Link matches a[href] first at the default priority, so a copied file
        // anchor would paste as a hyperlink.
        priority: 60,
        getAttrs: fileAttrsFromHTML,
      },
      {
        tag: "a[href]",
        priority: 60,
        getAttrs: fileAttrsFromHTML,
      },
    ];
  },
  renderHTML({ node }) {
    const fileId = typeof node.attrs.fileId === "string" ? node.attrs.fileId : null;
    const fileName = typeof node.attrs.fileName === "string" ? node.attrs.fileName : "";
    const size = Number(node.attrs.size);

    return ["a", {
      "href": fileId ? attachmentPath(fileId, attachmentParentID(this.editor)) : undefined,
      "data-file-id": fileId,
      "data-size": Number.isFinite(size) && size > 0 ? String(size) : undefined,
    }, fileName];
  },
  addNodeView() {
    return ReactNodeViewRenderer(FileNodeView);
  },
});

function fileAttrsFromHTML(element: string | { getAttribute(name: string): string | null }) {
  if (typeof element === "string") {
    return false;
  }

  const fileId = fileIdFromHTML(element);
  // Drop a blank id. A file node needs a file id to save.
  if (fileId === null) {
    return false;
  }

  return { fileId };
}
