// Copyright (c) 2026 Probo Inc <hello@probo.com>.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

import { Node } from "@tiptap/core";
import { ReactNodeViewRenderer } from "@tiptap/react";

import { ContentUploadNodeView } from "./ContentUploadNodeView";

export const ContentUploadExtension = Node.create({
  name: "contentUpload",
  group: "block",
  atom: true,
  selectable: false,
  addAttributes() {
    return {
      id: { default: null },
      fileName: { default: "" },
    };
  },
  renderHTML({ node }) {
    const id = typeof node.attrs.id === "string" ? node.attrs.id : "";
    const fileName = typeof node.attrs.fileName === "string" ? node.attrs.fileName : "";

    return ["div", { "data-content-upload": id }, fileName];
  },
  addNodeView() {
    return ReactNodeViewRenderer(ContentUploadNodeView);
  },
});
