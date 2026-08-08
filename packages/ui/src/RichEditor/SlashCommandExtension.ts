// Copyright (c) 2026 Probo Inc <hello@probo.com>.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

import { Extension } from "@tiptap/core";
import {
  Plugin,
  PluginKey,
  type Transaction,
} from "@tiptap/pm/state";
import type { EditorView } from "@tiptap/pm/view";
import { Decoration, DecorationSet } from "@tiptap/pm/view";

type SlashCommandPluginState = {
  active: boolean;
  query: string;
  from: number;
};

type SlashCommandAction
  = { type: "open"; from: number }
    | { type: "append"; text: string }
    | { type: "backspace" }
    | { type: "close" };

const slashCommandKey = new PluginKey<SlashCommandPluginState>("slashCommand");
const inactiveSlashCommand: SlashCommandPluginState = {
  active: false,
  query: "",
  from: 0,
};

export type SlashCommandStorage = {
  active: boolean;
  query: string;
  from: number;
};

export function activateSlashCommand(view: EditorView, from: number): void {
  dispatchSlashCommand(view, { type: "open", from });
}

export function deactivateSlashCommand(view: EditorView): void {
  dispatchSlashCommand(view, { type: "close" });
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
    return [createSlashCommandPlugin(this.storage)];
  },
});

export function createSlashCommandPlugin(
  storage: SlashCommandStorage,
): Plugin<SlashCommandPluginState> {
  return new Plugin({
    key: slashCommandKey,

    state: {
      init() {
        updateSlashCommandStorage(storage, inactiveSlashCommand);
        return inactiveSlashCommand;
      },

      apply(transaction, current) {
        const action = slashCommandAction(transaction);
        let next = applySlashCommandAction(current, action);
        if (next.active && transaction.docChanged) {
          const mapped = transaction.mapping.mapResult(next.from, 1);
          next = mapped.deleted
            ? inactiveSlashCommand
            : { ...next, from: mapped.pos };
        }
        if (
          next.active
          && transaction.selectionSet
          && transaction.selection.from !== next.from
        ) {
          next = inactiveSlashCommand;
        }

        updateSlashCommandStorage(storage, next);
        return next;
      },
    },

    props: {
      handleTextInput(view, from, _to, text) {
        const current = slashCommandKey.getState(view.state);
        if (current?.active) {
          dispatchSlashCommand(view, { type: "append", text });
          return true;
        }
        if (text !== "/" || !canOpenSlashCommand(view, from)) {
          return false;
        }

        dispatchSlashCommand(view, { type: "open", from });
        return true;
      },

      handleKeyDown(view, event) {
        const current = slashCommandKey.getState(view.state);
        if (!current?.active) return false;

        if (event.key === "Escape") {
          dispatchSlashCommand(view, { type: "close" });
          return true;
        }

        if (event.key === "Backspace") {
          dispatchSlashCommand(view, { type: "backspace" });
          return true;
        }

        return false;
      },

      decorations(state) {
        const current = slashCommandKey.getState(state);
        if (!current?.active) return DecorationSet.empty;

        try {
          state.doc.resolve(current.from);
          return DecorationSet.create(state.doc, [
            Decoration.widget(
              current.from,
              () => createSlashCommandWidget(current.query),
              { key: `slash-command:${current.query}`, side: -1 },
            ),
          ]);
        } catch {
          return DecorationSet.empty;
        }
      },
    },
  });
}

function applySlashCommandAction(
  current: SlashCommandPluginState,
  action: SlashCommandAction | undefined,
): SlashCommandPluginState {
  switch (action?.type) {
    case "open":
      return { active: true, query: "", from: action.from };
    case "append":
      return { ...current, query: current.query + action.text };
    case "backspace":
      return current.query.length === 0
        ? inactiveSlashCommand
        : { ...current, query: current.query.slice(0, -1) };
    case "close":
      return inactiveSlashCommand;
    default:
      return current;
  }
}

function canOpenSlashCommand(view: EditorView, from: number): boolean {
  const $from = view.state.doc.resolve(from);
  if ($from.parent.type.name === "codeBlock") return false;
  if ($from.marks().some(mark => mark.type.name === "code")) return false;

  const blockStart = $from.start($from.depth);
  return from === blockStart && $from.parent.textContent.length === 0;
}

function createSlashCommandWidget(query: string): HTMLElement {
  const widget = document.createElement("span");
  widget.className = "slash-search";
  widget.dataset.placeholder = "Search";
  widget.dataset.empty = query.length === 0 ? "true" : "false";
  widget.contentEditable = "false";
  widget.textContent = `/${query}`;
  return widget;
}

function dispatchSlashCommand(
  view: EditorView,
  action: SlashCommandAction,
): void {
  view.dispatch(view.state.tr.setMeta(slashCommandKey, action));
}

function slashCommandAction(
  transaction: Transaction,
): SlashCommandAction | undefined {
  return transaction.getMeta(slashCommandKey) as SlashCommandAction | undefined;
}

function updateSlashCommandStorage(
  storage: SlashCommandStorage,
  state: SlashCommandPluginState,
): void {
  storage.active = state.active;
  storage.query = state.query;
  storage.from = state.from;
}
