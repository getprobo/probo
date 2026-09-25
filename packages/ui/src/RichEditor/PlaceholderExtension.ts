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

import { type Editor, Extension } from "@tiptap/core";
import { type EditorState, Plugin, PluginKey } from "@tiptap/pm/state";
import { Decoration, DecorationSet, type EditorView } from "@tiptap/pm/view";

const placeholderKey = new PluginKey("placeholder");

export const defaultPlaceholder = "Write or type / for commands\u2026";

function computeDecorations(
  state: EditorState,
  editable: boolean,
  placeholder: string,
): DecorationSet {
  if (!editable) {
    return DecorationSet.empty;
  }

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
      "data-placeholder": placeholder,
    }),
  ]);
}

type PlaceholderStorage = {
  text: string;
};

function placeholderStorage(editor: Editor): PlaceholderStorage | undefined {
  return (editor.storage as { placeholder?: PlaceholderStorage }).placeholder;
}

function refreshPlaceholder(editor: Editor) {
  if (editor.isDestroyed) {
    return;
  }

  editor.view.dispatch(editor.state.tr.setMeta(placeholderKey, true));
}

export function setPlaceholder(editor: Editor, placeholder?: string) {
  if (editor.isDestroyed) {
    return;
  }

  const storage = placeholderStorage(editor);
  if (storage == null) {
    return;
  }

  const next = placeholder ?? defaultPlaceholder;
  if (storage.text === next) {
    return;
  }

  storage.text = next;
  refreshPlaceholder(editor);
}

export const PlaceholderExtension = Extension.create<
  Record<string, never>,
  PlaceholderStorage
>({
  name: "placeholder",

  addStorage() {
    return {
      text: defaultPlaceholder,
    };
  },

  addProseMirrorPlugins() {
    const { editor } = this;
    const getPlaceholder = () => this.storage.text;

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
            return computeDecorations(state, editor.isEditable, getPlaceholder());
          },
          apply(tr, value, _oldState, newState) {
            if (
              !tr.docChanged
              && !tr.selectionSet
              && !tr.getMeta(placeholderKey)
            ) {
              return value;
            }
            return computeDecorations(newState, editor.isEditable, getPlaceholder());
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
