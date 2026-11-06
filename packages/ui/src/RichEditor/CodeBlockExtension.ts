// Copyright (c) 2026 Probo Inc <hello@probo.com>.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

import { CodeBlock } from "@tiptap/extension-code-block";
import { ReactNodeViewRenderer } from "@tiptap/react";

import { readNullableStringAttribute } from "./automergeMapping";
import { MermaidNodeView } from "./MermaidNodeView";

export const CodeBlockExtension = CodeBlock.extend({
  automerge: () => ({
    block: "code-block",
    explicitIdentity: "topLevelBlock",
    attrParsers: {
      fromAutomerge: block => ({
        language: readNullableStringAttribute(block.attrs, "language"),
      }),
      fromProsemirror: node => ({
        language: readNullableStringAttribute(node.attrs, "language"),
      }),
    },
  }),
  addNodeView() {
    return ReactNodeViewRenderer(MermaidNodeView);
  },
});
