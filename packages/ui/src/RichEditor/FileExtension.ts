// Copyright (c) 2026 Probo Inc <hello@probo.com>.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

import { Node } from "@tiptap/core";
import { ReactNodeViewRenderer } from "@tiptap/react";

import { attachmentPath } from "./attachments";
import { FileNodeView } from "./FileNodeView";

export const FileExtension = Node.create({
  name: "file",
  group: "block",
  atom: true,
  draggable: true,
  addAttributes() {
    return {
      fileId: {
        default: null,
        parseHTML: element => element.getAttribute("data-file-id"),
      },
      fileName: {
        default: "",
        parseHTML: element => element.textContent,
      },
      mimeType: { default: "" },
      size: { default: 0 },
    };
  },
  parseHTML() {
    return [{ tag: "a[data-file-id]" }];
  },
  renderHTML({ node }) {
    const fileId = typeof node.attrs.fileId === "string" ? node.attrs.fileId : null;
    const fileName = typeof node.attrs.fileName === "string" ? node.attrs.fileName : "";

    return ["a", {
      "href": fileId ? attachmentPath(fileId) : undefined,
      "data-file-id": fileId,
    }, fileName];
  },
  addNodeView() {
    return ReactNodeViewRenderer(FileNodeView);
  },
});
