// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import { hasMarkdown, parseMarkdown } from "@probo/prosemirror";
import { Extension } from "@tiptap/core";
import { Fragment, Slice } from "@tiptap/pm/model";
import { Plugin, PluginKey } from "@tiptap/pm/state";

const markdownPasteKey = new PluginKey("markdownPaste");

export const MarkdownPasteExtension = Extension.create({
  name: "markdownPaste",

  addProseMirrorPlugins() {
    return [
      new Plugin({
        key: markdownPasteKey,
        props: {
          handlePaste(view, event) {
            const text = event.clipboardData?.getData("text/plain") ?? "";
            if (!hasMarkdown(text)) {
              return false;
            }

            const schema = view.state.schema;
            const nodes = parseMarkdown(text, schema);
            if (nodes.length === 0) return false;

            const slice = new Slice(Fragment.from(nodes), 0, 0);
            const tr = view.state.tr.replaceSelection(slice);
            view.dispatch(tr);
            return true;
          },
        },
      }),
    ];
  },
});
