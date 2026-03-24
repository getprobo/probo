// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import { Extension } from "@tiptap/core";
import { Plugin, PluginKey } from "@tiptap/pm/state";
import { cellAround } from "@tiptap/pm/tables";
import { Decoration, DecorationSet } from "@tiptap/pm/view";

const focusedCellKey = new PluginKey("focusedCell");

export const FocusedCellExtension = Extension.create({
  name: "focusedCell",

  addProseMirrorPlugins() {
    return [
      new Plugin({
        key: focusedCellKey,

        props: {
          decorations(state) {
            const { selection, doc } = state;
            const $pos = doc.resolve(selection.from);
            const cell = cellAround($pos);
            if (!cell) return DecorationSet.empty;

            const node = doc.nodeAt(cell.pos);
            if (!node) return DecorationSet.empty;

            return DecorationSet.create(doc, [
              Decoration.node(cell.pos, cell.pos + node.nodeSize, {
                class: "focusedCell",
              }),
            ]);
          },
        },
      }),
    ];
  },
});
