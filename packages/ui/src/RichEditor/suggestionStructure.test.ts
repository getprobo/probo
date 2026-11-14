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

import * as Automerge from "@automerge/automerge";
import type { DocHandle } from "@automerge/prosemirror";
import { pmDocFromSpans } from "@automerge/prosemirror";
import { getSchema } from "@tiptap/core";
import type { Node as ProseMirrorNode } from "@tiptap/pm/model";
import { EditorState, type Plugin, TextSelection, type Transaction } from "@tiptap/pm/state";
import { addRowAfter, CellSelection } from "@tiptap/pm/tables";
import type { EditorView } from "@tiptap/pm/view";
import { describe, expect, it } from "vitest";

import { createAutomergeSyncPlugin } from "./AutomergeSyncPlugin";
import { createSchemaAdapter, type RichEditorAutomergeDocument } from "./collaboration";
import { createRichEditorAutomergeDocument } from "./CollaborativeRichEditor";
import { richEditorCollaborationExtensions } from "./RichEditor";
import {
  applySuggestingStructure,
  createSuggestingExtension,
  discardSuggestionId,
  reassignSuggestionId,
  type SuggestingController,
} from "./SuggestionExtension";
import {
  structuralSuggestionSplit,
  structureSignature,
  suggestDeleteColumn,
  suggestDeleteRow,
  suggestDeleteTable,
  suggestInsertColumn,
  suggestInsertRow,
  suggestInsertTable,
} from "./suggestionStructure";

const schema = getSchema(richEditorCollaborationExtensions);
const id = "suggestion-1";

const permissions = {
  type: "doc",
  content: [
    paragraphJSON("Intro"),
    {
      type: "table",
      content: [
        rowJSON("tableHeader", ["Right", "Access"]),
        rowJSON("tableCell", ["Read", "Yes"]),
      ],
    },
    paragraphJSON("Outro"),
  ],
};

describe("tracked table rows", () => {
  it("inserts a row whose row and cells carry the insertion", () => {
    const state = stateFor(permissions);
    const tr = suggestInsertRow(state, id, { tableStart: tableStart(state), rowIndex: 1 }, false);
    if (!tr) throw new Error("expected a tracked row");
    const next = state.apply(tr);

    expect(grid(next.doc)).toEqual([["H:Right", "H:Access"], ["Read", "Yes"], ["", ""]]);
    const row = table(next.doc).child(2);
    expect(structuralSuggestionSplit(row)).toEqual({ id, kind: "insertion" });
    row.forEach(cell => expect(structuralSuggestionSplit(cell)).toEqual({ id, kind: "insertion" }));
    expect(next.selection.$from.node(-2)).toBe(row);
  });

  it("copies the cell types of the row below a header row", () => {
    const state = stateFor(permissions);
    const below = state.apply(
      suggestInsertRow(state, id, { tableStart: tableStart(state), rowIndex: 0 }, false)!,
    );
    expect(grid(below.doc)[1]).toEqual(["", ""]);

    const above = state.apply(
      suggestInsertRow(state, id, { tableStart: tableStart(state), rowIndex: 0 }, true)!,
    );
    expect(table(above.doc).child(0).child(0).type.name).toBe("tableCell");
  });

  it("redlines every cell and character of a deleted row", () => {
    const state = stateFor(permissions);
    const tr = suggestDeleteRow(state, id, { tableStart: tableStart(state), rowIndex: 1 });
    if (!tr) throw new Error("expected a tracked deletion");
    const next = state.apply(tr);

    expect(grid(next.doc)).toEqual(grid(state.doc));
    const row = table(next.doc).child(1);
    expect(structuralSuggestionSplit(row)).toEqual({ id, kind: "deletion" });
    row.forEach((cell) => {
      expect(structuralSuggestionSplit(cell)).toEqual({ id, kind: "deletion" });
      expect(structuralSuggestionSplit(cell.child(0))).toBeNull();
      cell.descendants((node) => {
        if (!node.isText) return true;
        expect(node.marks.map(mark => [mark.type.name, mark.attrs.id as unknown])).toEqual([
          ["suggestionDelete", id],
        ]);
        return false;
      });
    });
  });

  it("refuses to delete the last row left", () => {
    let state = stateFor(permissions);
    state = state.apply(suggestDeleteRow(state, id, { tableStart: tableStart(state), rowIndex: 1 })!);
    expect(suggestDeleteRow(state, "suggestion-2", { tableStart: tableStart(state), rowIndex: 0 }))
      .toBeNull();
  });

  it("refuses to swallow another suggestion", () => {
    let state = stateFor(permissions);
    state = state.apply(suggestInsertRow(state, "other", { tableStart: tableStart(state), rowIndex: 1 }, false)!);
    expect(suggestDeleteRow(state, id, { tableStart: tableStart(state), rowIndex: 2 })).toBeNull();
  });

  it("refuses tables with merged cells", () => {
    const state = stateFor({
      type: "doc",
      content: [{
        type: "table",
        content: [
          { type: "tableRow", content: [cellJSON("tableCell", "Merged", { colspan: 2 })] },
          rowJSON("tableCell", ["Read", "Yes"]),
        ],
      }],
    });
    const target = { tableStart: tableStart(state), rowIndex: 1 };
    expect(suggestInsertRow(state, id, target, false)).toBeNull();
    expect(suggestDeleteRow(state, id, target)).toBeNull();
  });
});

describe("tracked table columns", () => {
  it("inserts a cell in every row at the column", () => {
    const state = stateFor(permissions);
    const tr = suggestInsertColumn(state, id, { tableStart: tableStart(state), columnIndex: 0 }, false);
    if (!tr) throw new Error("expected a tracked column");
    const next = state.apply(tr);

    expect(grid(next.doc)).toEqual([["H:Right", "H:", "H:Access"], ["Read", "", "Yes"]]);
    for (let row = 0; row < 2; row += 1) {
      const cell = table(next.doc).child(row).child(1);
      expect(structuralSuggestionSplit(cell)).toEqual({ id, kind: "insertion" });
    }
    expect(structuralSuggestionSplit(table(next.doc).child(0))).toBeNull();
  });

  it("appends a column after the last one", () => {
    const state = stateFor(permissions);
    const next = state.apply(
      suggestInsertColumn(state, id, { tableStart: tableStart(state), columnIndex: 1 }, false)!,
    );
    expect(grid(next.doc)).toEqual([["H:Right", "H:Access", "H:"], ["Read", "Yes", ""]]);
  });

  it("redlines each cell of a deleted column", () => {
    const state = stateFor(permissions);
    const next = state.apply(
      suggestDeleteColumn(state, id, { tableStart: tableStart(state), columnIndex: 1 })!,
    );
    for (let row = 0; row < 2; row += 1) {
      expect(structuralSuggestionSplit(table(next.doc).child(row).child(1)))
        .toEqual({ id, kind: "deletion" });
      expect(structuralSuggestionSplit(table(next.doc).child(row).child(0))).toBeNull();
    }
    expect(suggestDeleteColumn(next, "suggestion-2", { tableStart: tableStart(next), columnIndex: 0 }))
      .toBeNull();
  });
});

describe("tracked tables", () => {
  it("inserts a whole table after the current block", () => {
    const state = stateFor({ type: "doc", content: [paragraphJSON("Scope")] }, 2);
    const next = state.apply(suggestInsertTable(state, id, 2, 2, true)!);

    expect(next.doc.child(0).textContent).toBe("Scope");
    expect(grid(next.doc)).toEqual([["H:", "H:"], ["", ""]]);
    next.doc.child(1).descendants((node) => {
      if (node.isTextblock) return false;
      expect(structuralSuggestionSplit(node)).toEqual({ id, kind: "insertion" });
      return true;
    });
  });

  it("redlines a whole table", () => {
    const state = stateFor(permissions);
    const next = state.apply(suggestDeleteTable(state, id, { tableStart: tableStart(state) })!);
    expect(structuralSuggestionSplit(table(next.doc))).toEqual({ id, kind: "deletion" });
    expect(next.doc.textBetween(0, next.doc.content.size)).toBe(state.doc.textBetween(0, state.doc.content.size));
  });
});

describe("suggesting safety net", () => {
  it("drops untracked structural edits and reports them", () => {
    let blocked = 0;
    const { state, plugin } = suggestingState(permissions, () => {
      blocked += 1;
    });
    const inCell = state.apply(
      state.tr.setSelection(TextSelection.create(state.doc, textPosition(state.doc, "Read"))),
    );
    let untracked: Transaction | null = null;
    addRowAfter(inCell, (tr) => {
      untracked = tr;
    });
    if (!untracked) throw new Error("expected prosemirror-tables to add a row");

    expect(plugin.spec.filterTransaction?.(untracked, inCell)).toBe(false);
    expect(blocked).toBe(1);

    const tracked = suggestInsertRow(inCell, id, { tableStart: tableStart(inCell), rowIndex: 1 }, false)!;
    expect(plugin.spec.filterTransaction?.(tracked, inCell)).toBe(true);
  });

  it("keeps text edits and edits made while not suggesting", () => {
    const { state, plugin, controller } = suggestingState(permissions, () => {});
    const typed = state.tr.insertText("x", textPosition(state.doc, "Intro"));
    expect(plugin.spec.filterTransaction?.(typed, state)).toBe(true);

    const cleared = state.tr.setSelection(CellSelection.create(
      state.doc,
      textPosition(state.doc, "Read") - 2,
    ));
    expect(plugin.spec.filterTransaction?.(cleared, state)).toBe(true);

    controller.enabled = false;
    let untracked: Transaction | null = null;
    addRowAfter(state.apply(
      state.tr.setSelection(TextSelection.create(state.doc, textPosition(state.doc, "Read"))),
    ), (tr) => {
      untracked = tr;
    });
    expect(plugin.spec.filterTransaction?.(untracked!, state)).toBe(true);
  });

  it("ignores typing, splits and suggestion attrs in the signature", () => {
    const state = stateFor(permissions);
    const signature = structureSignature(state.doc);
    const typed = state.apply(state.tr.insertText("x", textPosition(state.doc, "Read")));
    expect(structureSignature(typed.doc)).toBe(signature);
    const split = typed.apply(typed.tr.split(textPosition(typed.doc, "Intro") + 2));
    expect(structureSignature(split.doc)).toBe(signature);
    const marked = state.apply(suggestDeleteRow(state, id, { tableStart: tableStart(state), rowIndex: 1 })!);
    expect(structureSignature(marked.doc)).toBe(signature);
  });

  it("runs the tracked edit through the editor helper", () => {
    const editor = editorFor(permissions);
    const handled = applySuggestingStructure(editor, (state, suggestionId) =>
      suggestInsertRow(state, suggestionId, { tableStart: tableStart(state), rowIndex: 1 }, false),
    );
    expect(handled).toBe(true);
    expect(grid(editor.state.doc)).toHaveLength(3);
    expect(structuralSuggestionSplit(table(editor.state.doc).child(2))).toEqual({
      id: "local:1",
      kind: "insertion",
    });

    editor.controller.enabled = false;
    expect(applySuggestingStructure(editor, () => null)).toBe(false);
  });

  it("does not open a suggestion for an edit it must refuse", () => {
    const editor = editorFor(permissions);
    editor.controller.suggestionId = null;
    let requested = 0;
    editor.controller.onNeedId = () => {
      requested += 1;
      return Promise.resolve(null);
    };
    let blocked = 0;
    editor.controller.onBlocked = () => {
      blocked += 1;
    };
    expect(applySuggestingStructure(editor, () => null)).toBe(true);
    expect(requested).toBe(0);
    expect(blocked).toBe(1);
  });
});

describe("structural suggestion ids", () => {
  it("moves rows and cells to the server id", () => {
    const editor = editorFor(permissions);
    editor.apply(suggestInsertRow(editor.state, "local:1", { tableStart: tableStart(editor.state), rowIndex: 1 }, false)!);
    reassignSuggestionId(editor, "local:1", "server");
    const row = table(editor.state.doc).child(2);
    expect(structuralSuggestionSplit(row)).toEqual({ id: "server", kind: "insertion" });
    row.forEach(cell => expect(structuralSuggestionSplit(cell)?.id).toBe("server"));
  });

  it("drops inserted structure and restores deleted structure on discard", () => {
    const editor = editorFor(permissions);
    editor.apply(suggestInsertTable(editor.state, id, 2, 2, false)!);
    editor.apply(suggestDeleteRow(editor.state, id, { tableStart: tableStart(editor.state, 1), rowIndex: 1 })!);
    editor.apply(suggestInsertColumn(
      editor.state,
      id,
      { tableStart: tableStart(editor.state, 1), columnIndex: 0 },
      true,
    )!);

    discardSuggestionId(editor, id);
    expect(editor.state.doc.toJSON()).toEqual(stateFor(permissions).doc.toJSON());
  });
});

describe("structural suggestions in automerge", () => {
  it.each([
    ["an inserted row", (state: EditorState) =>
      suggestInsertRow(state, id, { tableStart: tableStart(state), rowIndex: 0 }, true)],
    ["a deleted row", (state: EditorState) =>
      suggestDeleteRow(state, id, { tableStart: tableStart(state), rowIndex: 1 })],
    ["an inserted column", (state: EditorState) =>
      suggestInsertColumn(state, id, { tableStart: tableStart(state), columnIndex: 1 }, false)],
    ["a deleted column", (state: EditorState) =>
      suggestDeleteColumn(state, id, { tableStart: tableStart(state), columnIndex: 0 })],
    ["an inserted table", (state: EditorState) =>
      suggestInsertTable(state, id, 2, 2, true)],
    ["a deleted table", (state: EditorState) =>
      suggestDeleteTable(state, id, { tableStart: tableStart(state) })],
  ])("rolls %s back to the original spans, as the server checks", (_name, build) => {
    const editor = collaborativeEditor(permissions);
    const before = Automerge.spans(editor.document, ["body"]);
    const tr = build(editor.state);
    if (!tr) throw new Error("expected a tracked change");
    editor.apply(tr);
    const after = Automerge.spans(editor.document, ["body"]);
    expect(after).not.toEqual(before);
    expect(rolledBack(after)).toEqual(rolledBack(before));
  });

  it.each([
    ["a first row", (state: EditorState) =>
      suggestInsertRow(state, id, { tableStart: tableStart(state), rowIndex: 0 }, true)],
    ["a first column", (state: EditorState) =>
      suggestInsertColumn(state, id, { tableStart: tableStart(state), columnIndex: 0 }, true)],
    ["a middle column", (state: EditorState) =>
      suggestInsertColumn(state, id, { tableStart: tableStart(state), columnIndex: 1 }, true)],
    ["a last column", (state: EditorState) =>
      suggestInsertColumn(state, id, { tableStart: tableStart(state), columnIndex: 1 }, false)],
  ])("stores %s where the editor shows it", (_name, build) => {
    const editor = collaborativeEditor(permissions);
    editor.apply(build(editor.state)!);

    const reloaded = pmDocFromSpans(editor.adapter, Automerge.spans(editor.document, ["body"]));
    expect(reloaded.toJSON()).toEqual(editor.state.doc.toJSON());
  });

  it.each([
    ["first", 0, true],
    ["middle", 1, true],
    ["last", 1, false],
  ] as const)("stores text typed in a new %s column", (_name, columnIndex, before) => {
    const editor = collaborativeEditor(permissions);
    editor.apply(suggestInsertColumn(editor.state, id, { tableStart: tableStart(editor.state), columnIndex }, before)!);
    const column = before ? columnIndex : columnIndex + 1;
    let header = -1;
    table(editor.state.doc).child(0).forEach((_cell, offset, index) => {
      if (index === column) header = tableStart(editor.state) + 1 + offset + 2;
    });
    editor.apply(editor.state.tr.insertText("Notes", header));

    expect(grid(editor.state.doc)[0]?.[column]).toBe("H:Notes");
    const reloaded = pmDocFromSpans(editor.adapter, Automerge.spans(editor.document, ["body"]));
    expect(reloaded.toJSON()).toEqual(editor.state.doc.toJSON());
  });

  it.each([
    [0, true, ["", "Auditor", ""]],
    [1, true, ["Auditor", "", ""]],
    [1, false, ["Auditor", "", ""]],
  ] as const)("stores column %i (before: %s) added after other table suggestions", (columnIndex, before, expected) => {
    const editor = collaborativeEditor(permissions);
    editor.apply(suggestInsertRow(editor.state, "row", { tableStart: tableStart(editor.state), rowIndex: 0 }, false)!);
    let cell = -1;
    table(editor.state.doc).forEach((_row, offset, index) => {
      if (index === 1) cell = tableStart(editor.state) + offset + 3;
    });
    editor.apply(editor.state.tr.insertText("Auditor", cell));
    editor.apply(suggestDeleteRow(editor.state, "gone", { tableStart: tableStart(editor.state), rowIndex: 2 })!);
    editor.apply(suggestInsertColumn(editor.state, "column", { tableStart: tableStart(editor.state), columnIndex }, before)!);

    expect(grid(editor.state.doc)[1]).toEqual(expected);
    const column = before ? columnIndex : columnIndex + 1;
    table(editor.state.doc).forEach((row) => {
      expect(structuralSuggestionSplit(row.child(column))?.id).toBe("column");
    });
    const reloaded = pmDocFromSpans(editor.adapter, Automerge.spans(editor.document, ["body"]));
    expect(reloaded.toJSON()).toEqual(editor.state.doc.toJSON());
  });

  it("stores an inserted row as split table blocks", () => {
    const editor = collaborativeEditor(permissions);
    editor.apply(suggestInsertRow(editor.state, id, { tableStart: tableStart(editor.state), rowIndex: 1 }, false)!);

    const blocks = blockSpans(editor.document);
    const rows = blocks.filter(block => block.type === "table-row");
    expect(rows).toHaveLength(3);
    expect(rows[2]?.attrs).toMatchObject({ suggestionSplit: id, splitKind: "insertion" });
    const lastCells = blocks.slice(blocks.indexOf(rows[2]) + 1)
      .filter(block => block.type === "table-cell");
    expect(lastCells).toHaveLength(2);
    for (const cell of lastCells) {
      expect(cell.attrs).toMatchObject({ suggestionSplit: id, splitKind: "insertion" });
    }

    const reloaded = pmDocFromSpans(editor.adapter, Automerge.spans(editor.document, ["body"]));
    expect(structuralSuggestionSplit(table(reloaded).child(2))).toEqual({ id, kind: "insertion" });
  });

  it("stores a deleted row as deletion splits and delete marks", () => {
    const editor = collaborativeEditor(permissions);
    editor.apply(suggestDeleteRow(editor.state, id, { tableStart: tableStart(editor.state), rowIndex: 1 })!);

    const marked = Automerge.spans(editor.document, ["body"]).map(span =>
      span.type === "block"
        ? `${scalarText(span.value.type)}:${scalarText(blockAttr(span.value.attrs, "splitKind"))}`
        : `${scalarText(span.value)}:${"suggestion-delete" in (span.marks ?? {}) ? "deleted" : ""}`,
    );
    expect(marked).toEqual([
      "paragraph:",
      "Intro:",
      "table:",
      "table-row:",
      "table-header:",
      "Right:",
      "table-header:",
      "Access:",
      "table-row:deletion",
      "table-cell:deletion",
      "Read:deleted",
      "table-cell:deletion",
      "Yes:deleted",
      "paragraph:",
      "Outro:",
    ]);
  });
});

describe("decided structural suggestions from another peer", () => {
  const isDeletion = (attrs: Record<string, unknown>) => attrs.splitKind === "deletion";
  const isSplit = (split: string) => (attrs: Record<string, unknown>) =>
    attrs.suggestionSplit === split;

  it("drops an accepted row deletion", () => {
    const editor = collaborativeEditor(permissions);
    editor.apply(suggestDeleteRow(editor.state, id, { tableStart: tableStart(editor.state), rowIndex: 1 })!);

    editor.decide(isDeletion);

    expect(grid(editor.state.doc)).toEqual([["H:Right", "H:Access"]]);
    expect(editor.state.doc.toJSON())
      .toEqual(pmDocFromSpans(editor.adapter, Automerge.spans(editor.document, ["body"])).toJSON());
  });

  it("drops a rejected column insertion next to an inserted row", () => {
    const editor = collaborativeEditor(permissions);
    editor.apply(suggestInsertRow(editor.state, "row", { tableStart: tableStart(editor.state), rowIndex: 1 }, false)!);
    editor.apply(suggestInsertColumn(editor.state, "column", { tableStart: tableStart(editor.state), columnIndex: 1 }, false)!);

    editor.decide(isSplit("column"));

    expect(grid(editor.state.doc)).toEqual([["H:Right", "H:Access"], ["Read", "Yes"], ["", ""]]);
    expect(editor.state.doc.toJSON())
      .toEqual(pmDocFromSpans(editor.adapter, Automerge.spans(editor.document, ["body"])).toJSON());
  });

  it("keeps accepted row and column insertions", () => {
    const editor = collaborativeEditor(permissions);
    editor.apply(suggestInsertRow(editor.state, "row", { tableStart: tableStart(editor.state), rowIndex: 1 }, false)!);
    editor.apply(suggestInsertColumn(editor.state, "column", { tableStart: tableStart(editor.state), columnIndex: 1 }, false)!);

    editor.settle(isSplit("column"));
    expect(editor.state.doc.toJSON())
      .toEqual(pmDocFromSpans(editor.adapter, Automerge.spans(editor.document, ["body"])).toJSON());
    editor.settle(isSplit("row"));

    expect(grid(editor.state.doc)).toEqual([["H:Right", "H:Access", "H:"], ["Read", "Yes", ""], ["", "", ""]]);
    expect(editor.state.doc.toJSON())
      .toEqual(pmDocFromSpans(editor.adapter, Automerge.spans(editor.document, ["body"])).toJSON());
    table(editor.state.doc).descendants((node) => {
      expect(structuralSuggestionSplit(node)).toBeNull();
    });
  });

  it.each([
    ["first", 0, true],
    ["last", 1, false],
  ] as const)("keeps a %s column in place when another peer renames it", (_name, columnIndex, before) => {
    const editor = collaborativeEditor(permissions);
    editor.apply(suggestInsertRow(editor.state, "row", { tableStart: tableStart(editor.state), rowIndex: 0 }, false)!);
    editor.apply(suggestInsertColumn(editor.state, "local:column", { tableStart: tableStart(editor.state), columnIndex }, before)!);
    const shown = editor.state.doc.toJSON() as unknown;

    editor.settle(isSplit("local:column"), "server-column");

    const column = before ? columnIndex : columnIndex + 1;
    table(editor.state.doc).forEach((row) => {
      expect(structuralSuggestionSplit(row.child(column))?.id).toBe("server-column");
    });
    expect(editor.state.doc.toJSON())
      .toEqual(pmDocFromSpans(editor.adapter, Automerge.spans(editor.document, ["body"])).toJSON());
    const renamedBack = JSON.stringify(editor.state.doc.toJSON()).split("server-column").join("local:column");
    expect(JSON.parse(renamedBack)).toEqual(shown);
  });

  it("drops a rejected table insertion", () => {
    const editor = collaborativeEditor(permissions);
    editor.apply(suggestInsertTable(editor.state, id, 2, 2, true)!);

    editor.decide(isSplit(id));

    expect(editor.state.doc.toJSON())
      .toEqual(pmDocFromSpans(editor.adapter, Automerge.spans(editor.document, ["body"])).toJSON());
    expect(editor.state.doc.childCount).toBe(schema.nodeFromJSON(permissions).childCount);
  });
});

function paragraphJSON(text: string) {
  return { type: "paragraph", content: text ? [{ type: "text", text }] : [] };
}

function cellJSON(type: string, text: string, attrs: Record<string, unknown> = {}) {
  return { type, attrs, content: [paragraphJSON(text)] };
}

function rowJSON(type: string, cells: string[]) {
  return { type: "tableRow", content: cells.map(text => cellJSON(type, text)) };
}

function stateFor(content: object, selection?: number): EditorState {
  const doc = schema.nodeFromJSON(content);
  return EditorState.create({
    schema,
    doc,
    selection: selection == null ? undefined : TextSelection.create(doc, selection),
  });
}

function table(doc: ProseMirrorNode): ProseMirrorNode {
  let found: ProseMirrorNode | null = null;
  doc.forEach((node) => {
    if (!found && node.type.name === "table") found = node;
  });
  if (!found) throw new Error("expected a table");
  return found;
}

function tableStart(state: EditorState, nth = 0): number {
  const starts: number[] = [];
  state.doc.forEach((node, offset) => {
    if (node.type.name === "table") starts.push(offset + 1);
  });
  return starts[nth] ?? -1;
}

function grid(doc: ProseMirrorNode): string[][] {
  const rows: string[][] = [];
  table(doc).forEach((row) => {
    const cells: string[] = [];
    row.forEach((cell) => {
      cells.push(`${cell.type.name === "tableHeader" ? "H:" : ""}${cell.textContent}`);
    });
    rows.push(cells);
  });
  return rows;
}

function textPosition(doc: ProseMirrorNode, text: string): number {
  let found = -1;
  doc.descendants((node, pos) => {
    if (found >= 0) return false;
    if (node.isText && node.text?.includes(text)) found = pos + node.text.indexOf(text);
    return true;
  });
  if (found < 0) throw new Error(`expected ${text}`);
  return found;
}

function controllerFor(onBlocked?: () => void): SuggestingController {
  const controller: SuggestingController = {
    enabled: true,
    suggestionId: null,
    pending: null,
    request: null,
    onNeedId: () => {
      controller.suggestionId = "local:1";
      return Promise.resolve("local:1");
    },
    onBurstFinished: () => {},
    onOpenSuggestion: () => {},
    onBlocked,
  };
  return controller;
}

// Tiptap types addProseMirrorPlugins with the editor context a live Mark
// receives. These tests only read the plugin list, so the mark stands in for
// that context here.
function suggestingState(content: object, onBlocked: () => void) {
  const controller = controllerFor(onBlocked);
  const extension = createSuggestingExtension({ current: controller });
  const addPlugins = extension.config.addProseMirrorPlugins!;
  const plugins: Plugin[] = addPlugins.call(
    extension as unknown as ThisParameterType<typeof addPlugins>,
  );
  const plugin = plugins[0];
  const doc = schema.nodeFromJSON(content);
  return { state: EditorState.create({ schema, doc, plugins }), plugin, controller };
}

function editorFor(content: object) {
  const controller = controllerFor();
  let state = stateFor(content);
  return {
    controller,
    storage: { documentSuggesting: { controller: { current: controller } } },
    get state() {
      return state;
    },
    view: {
      dispatch(transaction: Transaction) {
        state = state.apply(transaction);
      },
    },
    apply(transaction: Transaction) {
      state = state.apply(transaction);
    },
  };
}

type ChangeListener = (payload: {
  doc: Automerge.Doc<RichEditorAutomergeDocument>;
  patches: Automerge.Patch[];
  patchInfo: { before: Automerge.Doc<RichEditorAutomergeDocument> };
}) => void;

function collaborativeEditor(content: object) {
  let document = createRichEditorAutomergeDocument(JSON.stringify(content));
  let listener: ChangeListener | null = null;
  const handle = {
    doc: () => document,
    change: (callback: (draft: RichEditorAutomergeDocument) => void) => {
      document = Automerge.change(document, callback);
    },
    on: (_event: string, callback: ChangeListener) => {
      listener = callback;
    },
    off: () => {},
  } as unknown as DocHandle<RichEditorAutomergeDocument>;
  const adapter = createSchemaAdapter(richEditorCollaborationExtensions);
  const plugin = createAutomergeSyncPlugin(adapter, handle, ["body"]);
  let state = EditorState.create({
    schema: adapter.schema,
    doc: pmDocFromSpans(adapter, Automerge.spans(document, ["body"])),
    plugins: [plugin],
  });
  const view = {
    get state() {
      return state;
    },
    dispatch(transaction: Transaction) {
      state = state.apply(transaction);
    },
  } as unknown as EditorView;
  plugin.spec.view?.(view);
  return {
    adapter,
    get document() {
      return document;
    },
    get state() {
      return state;
    },
    apply(transaction: Transaction) {
      state = state.apply(transaction);
    },
    // decide plays a server decision from another peer: every block whose
    // attrs match is removed with its text, the way a subtree drop lands.
    decide(matches: (attrs: Record<string, unknown>) => boolean) {
      const before = document;
      const ranges: Array<[number, number]> = [];
      let index = 0;
      let removing = false;
      for (const span of Automerge.spans(before, ["body"])) {
        const length = span.type === "block" ? 1 : String(span.value).length;
        if (span.type === "block") {
          removing = matches(span.value.attrs as Record<string, unknown>);
        }
        if (removing) ranges.push([index, length]);
        index += length;
      }
      const after = Automerge.change(Automerge.clone(before), (draft) => {
        for (const [start, length] of ranges.reverse()) {
          Automerge.splice(draft, ["body"], start, length);
        }
      });
      const patches = Automerge.diff(after, Automerge.getHeads(before), Automerge.getHeads(after));
      document = after;
      listener?.({ doc: after, patches, patchInfo: { before } });
    },
    // settle plays an accepted insertion from another peer: matching blocks
    // keep their place and lose their split attrs, or move to renamedTo.
    settle(matches: (attrs: Record<string, unknown>) => boolean, renamedTo?: string) {
      const before = document;
      const updates: Array<[number, Record<string, unknown>]> = [];
      let index = 0;
      for (const span of Automerge.spans(before, ["body"])) {
        if (span.type === "block") {
          const attrs = { ...(span.value.attrs as Record<string, unknown>) };
          if (matches(attrs)) {
            if (renamedTo) {
              attrs.suggestionSplit = renamedTo;
            } else {
              delete attrs.suggestionSplit;
              delete attrs.splitKind;
            }
            updates.push([index, attrs]);
          }
          index += 1;
        } else {
          index += String(span.value).length;
        }
      }
      const after = Automerge.change(Automerge.clone(before), (draft) => {
        for (const [at, attrs] of updates) {
          const block = Automerge.block(draft, ["body"], at)!;
          Automerge.updateBlock(draft, ["body"], at, {
            type: new Automerge.ImmutableString(scalarText(block.type)),
            parents: (block.parents as unknown[])
              .map(parent => new Automerge.ImmutableString(scalarText(parent))),
            attrs,
          });
        }
      });
      const patches = Automerge.diff(after, Automerge.getHeads(before), Automerge.getHeads(after));
      document = after;
      listener?.({ doc: after, patches, patchInfo: { before } });
    },
  };
}

function blockSpans(document: Automerge.Doc<RichEditorAutomergeDocument>) {
  return Automerge.spans(document, ["body"]).flatMap(span =>
    span.type === "block"
      ? [{ type: scalarText(span.value.type), attrs: span.value.attrs as Record<string, unknown> }]
      : [],
  );
}

// rolledBack mirrors the server's normalizeSuggestionSpans: suggested blocks
// and insertions vanish, deletion splits and delete marks drop.
function rolledBack(spans: ReturnType<typeof Automerge.spans>): unknown[] {
  const result: Array<{ type: string; value: unknown; marks?: Record<string, unknown> }> = [];
  for (const span of spans) {
    if (span.type === "block") {
      const attrs = { ...(span.value.attrs as Record<string, unknown>) };
      if (typeof attrs.suggestionSplit === "string" && attrs.splitKind !== "deletion") continue;
      delete attrs.suggestionSplit;
      delete attrs.splitKind;
      result.push({ type: "block", value: { ...span.value, attrs } });
      continue;
    }
    if (span.type !== "text") continue;
    const marks = { ...(span.marks ?? {}) } as Record<string, unknown>;
    if ("suggestion-insert" in marks) continue;
    delete marks["suggestion-delete"];
    const last = result[result.length - 1];
    if (last?.type === "text" && JSON.stringify(last.marks) === JSON.stringify(marks)) {
      last.value = String(last.value) + String(span.value);
      continue;
    }
    result.push({ type: "text", value: String(span.value), marks });
  }
  return result;
}

function scalarText(value: unknown): string {
  if (typeof value === "string") return value;
  if (Automerge.isImmutableString(value)) return value.val;
  return JSON.stringify(value);
}

function blockAttr(attrs: unknown, name: string): unknown {
  if (!attrs || typeof attrs !== "object") return "";
  return (attrs as Record<string, unknown>)[name] ?? "";
}
