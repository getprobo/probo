// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import { Extension } from "@tiptap/core";
import { Plugin, PluginKey } from "@tiptap/pm/state";
import { Decoration, DecorationSet } from "@tiptap/pm/view";

const slashCommandKey = new PluginKey("slashCommand");

export type SlashCommandStorage = {
  active: boolean;
  query: string;
  from: number;
};

export function activateSlashCommand(storage: SlashCommandStorage, from: number) {
  storage.active = true;
  storage.query = "";
  storage.from = from;
}

export function deactivateSlashCommand(storage: SlashCommandStorage) {
  storage.active = false;
  storage.query = "";
  storage.from = 0;
}

export const SlashCommandExtension = Extension.create<object, SlashCommandStorage>({
  name: "slashCommand",

  addStorage() {
    return {
      active: false,
      query: "",
      from: 0,
    };
  },

  addProseMirrorPlugins() {
    const storage = this.storage;

    return [
      new Plugin({
        key: slashCommandKey,

        props: {
          handleTextInput(view, from, _to, text) {
            if (text !== "/") return false;
            if (storage.active) return false;

            const { state } = view;
            const $from = state.doc.resolve(from);

            if ($from.parent.type.name === "codeBlock") return false;
            if ($from.marks().some(m => m.type.name === "code")) return false;

            const blockStart = $from.start($from.depth);
            if (from !== blockStart) return false;
            if ($from.parent.textContent.length !== 0) return false;

            storage.active = true;
            storage.from = from;
            storage.query = "";

            return false;
          },

          handleKeyDown(view, event) {
            if (!storage.active) return false;

            if (event.key === "Escape") {
              const { state } = view;
              const cursorPos = state.selection.from;
              const from = storage.from;

              deactivateSlashCommand(storage);

              if (cursorPos > from) {
                const tr = state.tr.delete(from, cursorPos);
                view.dispatch(tr);
              }

              return true;
            }

            if (event.key === "Backspace") {
              const { state } = view;
              const cursorPos = state.selection.from;

              if (cursorPos <= storage.from + 1) {
                deactivateSlashCommand(storage);
              }
            }

            return false;
          },

          decorations(state) {
            if (!storage.active) return DecorationSet.empty;

            const { from } = storage;
            const cursorPos = state.selection.from;

            try {
              const $from = state.doc.resolve(from);
              const blockStart = $from.start($from.depth);
              const blockEnd = $from.end($from.depth);

              if (cursorPos < blockStart || cursorPos > blockEnd) {
                deactivateSlashCommand(storage);
                return DecorationSet.empty;
              }

              const text = state.doc.textBetween(from, cursorPos);
              if (!text.startsWith("/")) {
                deactivateSlashCommand(storage);
                return DecorationSet.empty;
              }

              storage.query = text.slice(1);

              const decoEnd = Math.max(cursorPos, from + 1);
              const isEmpty = storage.query.length === 0;

              return DecorationSet.create(state.doc, [
                Decoration.inline(from, decoEnd, {
                  "nodeName": "span",
                  "class": "slash-search",
                  "data-placeholder": "Search",
                  "data-empty": isEmpty
                    ? "true"
                    : "false",
                } as Record<string, string>),
              ]);
            } catch {
              deactivateSlashCommand(storage);
              return DecorationSet.empty;
            }
          },
        },
      }),
    ];
  },
});
