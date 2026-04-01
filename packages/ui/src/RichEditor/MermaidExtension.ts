// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import { CodeBlock } from "@tiptap/extension-code-block";
import { ReactNodeViewRenderer } from "@tiptap/react";

import { MermaidNodeView } from "./MermaidNodeView";

export const MermaidCodeBlock = CodeBlock.extend({
  addNodeView() {
    return ReactNodeViewRenderer(MermaidNodeView);
  },
});
