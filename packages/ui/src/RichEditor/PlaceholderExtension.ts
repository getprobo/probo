// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import { Extension } from "@tiptap/core";
import { type EditorState, Plugin, PluginKey } from "@tiptap/pm/state";
import { Decoration, DecorationSet, type EditorView } from "@tiptap/pm/view";

const placeholderKey = new PluginKey("placeholder");

function computeDecorations(
  state: EditorState,
  editable: boolean,
): DecorationSet {
  if (!editable) return DecorationSet.empty;

  const { selection } = state;
  if (!selection.empty) return DecorationSet.empty;

  const $pos = selection.$from;
  const node = $pos.parent;

  if (node.type.name !== "paragraph") return DecorationSet.empty;
  if (node.content.size !== 0) return DecorationSet.empty;

  if ($pos.depth >= 2) {
    const parentName = $pos.node($pos.depth - 1).type.name;
    if (
      parentName === "listItem"
      || parentName === "tableCell"
      || parentName === "tableHeader"
    ) {
      return DecorationSet.empty;
    }
  }

  const pos = $pos.before($pos.depth);

  return DecorationSet.create(state.doc, [
    Decoration.node(pos, pos + node.nodeSize, {
      "class": "is-empty-focused",
      "data-placeholder": "Write or type / for commands\u2026",
    }),
  ]);
}

export const PlaceholderExtension = Extension.create({
  name: "placeholder",

  addProseMirrorPlugins() {
    const { editor } = this;

    return [
      new Plugin({
        key: placeholderKey,

        view() {
          let lastEditable = editor.isEditable;
          return {
            update(view: EditorView) {
              const editable = editor.isEditable;
              if (editable !== lastEditable) {
                lastEditable = editable;
                view.dispatch(
                  view.state.tr.setMeta(placeholderKey, true),
                );
              }
            },
          };
        },

        state: {
          init(_config, state) {
            return computeDecorations(state, editor.isEditable);
          },
          apply(tr, value, _oldState, newState) {
            if (
              !tr.docChanged
              && !tr.selectionSet
              && !tr.getMeta(placeholderKey)
            ) {
              return value;
            }
            return computeDecorations(newState, editor.isEditable);
          },
        },

        props: {
          decorations(state) {
            return placeholderKey.getState(state) as DecorationSet;
          },
        },
      }),
    ];
  },
});
