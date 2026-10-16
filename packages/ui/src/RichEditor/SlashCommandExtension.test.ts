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

import { Schema } from "@tiptap/pm/model";
import { EditorState, type Transaction } from "@tiptap/pm/state";
import type { EditorView } from "@tiptap/pm/view";
import { describe, expect, it } from "vitest";

import {
  createSlashCommandPlugin,
  type SlashCommandStorage,
} from "./SlashCommandExtension";

const schema = new Schema({
  nodes: {
    doc: { content: "block+" },
    paragraph: { content: "inline*", group: "block" },
    text: { group: "inline" },
  },
});

describe("SlashCommandExtension", () => {
  it("keeps slash queries out of the document", () => {
    const storage = slashCommandStorage();
    const plugin = createSlashCommandPlugin(storage);
    let state = EditorState.create({
      schema,
      doc: schema.node("doc", null, [schema.node("paragraph")]),
      plugins: [plugin],
    });
    const transactions: Transaction[] = [];
    const view = {
      get state() {
        return state;
      },
      dispatch(transaction: Transaction) {
        transactions.push(transaction);
        state = state.apply(transaction);
      },
    } as EditorView;
    const handleTextInput = plugin.props.handleTextInput;
    if (!handleTextInput) throw new Error("expected text input handler");

    expect(
      handleTextInput.call(plugin, view, 1, 1, "/", () => state.tr),
    ).toBe(true);
    expect(
      handleTextInput.call(plugin, view, 1, 1, "head", () => state.tr),
    ).toBe(true);

    expect(storage).toEqual({
      active: true,
      query: "head",
      from: 1,
    });
    expect(state.doc.textContent).toBe("");
    expect(transactions.every(transaction => !transaction.docChanged)).toBe(true);
  });

  it("edits and closes a local query with keyboard commands", () => {
    const storage = slashCommandStorage();
    const plugin = createSlashCommandPlugin(storage);
    let state = EditorState.create({
      schema,
      doc: schema.node("doc", null, [schema.node("paragraph")]),
      plugins: [plugin],
    });
    const view = {
      get state() {
        return state;
      },
      dispatch(transaction: Transaction) {
        state = state.apply(transaction);
      },
    } as EditorView;
    const handleTextInput = plugin.props.handleTextInput;
    const handleKeyDown = plugin.props.handleKeyDown;
    if (!handleTextInput || !handleKeyDown) {
      throw new Error("expected slash command keyboard handlers");
    }

    handleTextInput.call(plugin, view, 1, 1, "/", () => state.tr);
    handleTextInput.call(plugin, view, 1, 1, "ab", () => state.tr);
    expect(
      handleKeyDown.call(
        plugin,
        view,
        { key: "Backspace" } as KeyboardEvent,
      ),
    ).toBe(true);
    expect(storage.query).toBe("a");
    expect(
      handleKeyDown.call(
        plugin,
        view,
        { key: "Escape" } as KeyboardEvent,
      ),
    ).toBe(true);
    expect(storage.active).toBe(false);
    expect(state.doc.textContent).toBe("");
  });
});

function slashCommandStorage(): SlashCommandStorage {
  return {
    active: false,
    query: "",
    from: 0,
  };
}
