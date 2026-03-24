// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import { Extension } from "@tiptap/core";
import { Plugin, PluginKey } from "@tiptap/pm/state";
import { Decoration, DecorationSet } from "@tiptap/pm/view";

const placeholderKey = new PluginKey("placeholder");

export const PlaceholderExtension = Extension.create({
  name: "placeholder",

  addProseMirrorPlugins() {
    return [
      new Plugin({
        key: placeholderKey,

        props: {
          decorations(state) {
            const { selection } = state;
            if (!selection.empty) return DecorationSet.empty;

            const $pos = selection.$from;
            const node = $pos.parent;

            if (node.type.name !== "paragraph") return DecorationSet.empty;
            if (node.content.size !== 0) return DecorationSet.empty;

            const pos = $pos.before($pos.depth);

            return DecorationSet.create(state.doc, [
              Decoration.node(pos, pos + node.nodeSize, {
                "class": "is-empty-focused",
                "data-placeholder": "Write or type / for commands\u2026",
              }),
            ]);
          },
        },
      }),
    ];
  },
});
