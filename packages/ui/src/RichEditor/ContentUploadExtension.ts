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
  parseHTML() {
    return [{
      tag: "div[data-content-upload]",
      // A copied placeholder is dropped on paste, including its file name.
      ignore: true,
    }];
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
