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
import {
  type DocHandle,
  pmDocFromSpans,
  syncPlugin,
} from "@automerge/prosemirror";
import { getSchema } from "@tiptap/core";
import { history, undo } from "@tiptap/pm/history";
import { Fragment } from "@tiptap/pm/model";
import { EditorState, type Transaction } from "@tiptap/pm/state";
import { describe, expect, it } from "vitest";

import {
  createSchemaAdapter,
  normalizeStructuralLeafFollowers,
  richEditorAutomergeContent,
  type RichEditorAutomergeDocument,
} from "./collaboration";
import {
  createRichEditorAutomergeDocument,
  richEditorCollaborationExtensions,
  supportsRichEditorCollaboration,
} from "./RichEditor";

describe("RichEditor collaboration", () => {
  it("imports supported ProseMirror content into Automerge rich text", () => {
    const document = createRichEditorAutomergeDocument(
      JSON.stringify({
        type: "doc",
        content: [
          {
            type: "heading",
            attrs: { level: 2 },
            content: [{ type: "text", text: "Policy" }],
          },
          {
            type: "paragraph",
            content: [
              { type: "text", text: "Hello ", marks: [{ type: "bold" }] },
              { type: "text", text: "world" },
            ],
          },
        ],
      }),
    );

    expect(Automerge.spans(document, ["body"])).not.toHaveLength(0);
    expect(document.body).toContain("Hello world");
  });

  it("preserves table and row boundaries", () => {
    const content = tableDocumentJSON();

    expect(supportsRichEditorCollaboration(content)).toBe(true);
    const document = createRichEditorAutomergeDocument(content);
    const spans = Automerge.spans(document, ["body"]);
    expect(
      spans
        .filter(span => span.type === "block")
        .map(span => Automerge.isImmutableString(span.value.type)
          ? span.value.type.val
          : span.value.type),
    ).toEqual([
      "table",
      "table-row",
      "table-cell",
      "table-cell",
      "table-row",
      "table-cell",
      "table-cell",
    ]);

    const handle: DocHandle<RichEditorAutomergeDocument> = {
      doc: () => document,
      change: () => {},
      on: () => {},
      off: () => {},
    };
    const roundTrip = richEditorAutomergeContent(
      handle,
      richEditorCollaborationExtensions,
    ) as {
      content: Array<{
        type: string;
        content: Array<{
          type: string;
          content: unknown[];
        }>;
      }>;
    };
    expect(roundTrip.content[0].type).toBe("table");
    expect(roundTrip.content[0].content).toHaveLength(2);
    expect(roundTrip.content[0].content[0].type).toBe("tableRow");
    expect(roundTrip.content[0].content[0].content).toHaveLength(2);
  });

  it("applies table cell edits through the Automerge sync plugin", () => {
    let document = createRichEditorAutomergeDocument(tableDocumentJSON());
    const handle: DocHandle<RichEditorAutomergeDocument> = {
      doc: () => document,
      change: (change) => {
        document = Automerge.change(document, change);
      },
      on: () => {},
      off: () => {},
    };
    const adapter = createSchemaAdapter(richEditorCollaborationExtensions);
    const pmDocument = pmDocFromSpans(
      adapter,
      Automerge.spans(document, ["body"]),
    );
    const state = EditorState.create({
      schema: adapter.schema,
      doc: pmDocument,
      plugins: [
        syncPlugin({
          adapter,
          handle,
          path: ["body"],
        }),
      ],
    });
    let textPosition: number | undefined;
    state.doc.descendants((node, position) => {
      if (node.isText && node.text === "A") {
        textPosition = position;
        return false;
      }

      return true;
    });
    expect(textPosition).toBeDefined();

    state.applyTransaction(state.tr.insertText("X", textPosition! + 1));

    expect(
      Automerge.spans(document, ["body"])
        .filter(span => span.type === "text")
        .map(span => span.value)
        .join(""),
    ).toContain("AX");
  });

  it("preserves rows inserted through ProseMirror transactions", () => {
    let document = createRichEditorAutomergeDocument(tableDocumentJSON());
    const handle: DocHandle<RichEditorAutomergeDocument> = {
      doc: () => document,
      change: (change) => {
        document = Automerge.change(document, change);
      },
      on: () => {},
      off: () => {},
    };
    const adapter = createSchemaAdapter(richEditorCollaborationExtensions);
    const pmDocument = pmDocFromSpans(
      adapter,
      Automerge.spans(document, ["body"]),
    );
    const state = EditorState.create({
      schema: adapter.schema,
      doc: pmDocument,
      plugins: [
        syncPlugin({
          adapter,
          handle,
          path: ["body"],
        }),
      ],
    });
    const paragraph = adapter.schema.nodes.paragraph.create(
      null,
      adapter.schema.text("E"),
    );
    const cell = adapter.schema.nodes.tableCell.create(
      {
        isAmgBlock: true,
        colspan: 1,
        rowspan: 1,
        colwidth: null,
      },
      paragraph,
    );
    const row = adapter.schema.nodes.tableRow.create(
      { isAmgBlock: true },
      [cell, cell],
    );
    const table = state.doc.firstChild;
    if (!table) throw new Error("expected table node");

    state.applyTransaction(
      state.tr.insert(table.nodeSize - 1, row),
    );

    const spans = Automerge.spans(document, ["body"]);
    expect(
      spans.filter(span =>
        span.type === "block"
        && Automerge.isImmutableString(span.value.type)
        && span.value.type.val === "table-row",
      ),
    ).toHaveLength(3);

    const roundTrip = pmDocFromSpans(adapter, spans);
    expect(roundTrip.firstChild?.childCount).toBe(3);
  });

  it("converges concurrent row insertions", () => {
    const base = createRichEditorAutomergeDocument(tableDocumentJSON());
    const left = insertTableRow(
      Automerge.clone(base, { actor: "01000000000000000000000000000000" }),
      "L",
    );
    const right = insertTableRow(
      Automerge.clone(base, { actor: "02000000000000000000000000000000" }),
      "R",
    );
    const merged = Automerge.merge(left, right);
    const adapter = createSchemaAdapter(richEditorCollaborationExtensions);
    const spans = Automerge.spans(merged, ["body"]);
    const pmDocument = pmDocFromSpans(adapter, spans);

    expect(pmDocument.firstChild?.childCount).toBe(4);
    expect(pmDocument.textContent).toContain("L");
    expect(pmDocument.textContent).toContain("R");
  });

  it("records local table edits in collaborative undo history", () => {
    let document = createRichEditorAutomergeDocument(tableDocumentJSON());
    const handle: DocHandle<RichEditorAutomergeDocument> = {
      doc: () => document,
      change: (change) => {
        document = Automerge.change(document, change);
      },
      on: () => {},
      off: () => {},
    };
    const adapter = createSchemaAdapter(richEditorCollaborationExtensions);
    let state = EditorState.create({
      schema: adapter.schema,
      doc: pmDocFromSpans(adapter, Automerge.spans(document, ["body"])),
      plugins: [
        history(),
        syncPlugin({
          adapter,
          handle,
          path: ["body"],
        }),
      ],
    });
    let textPosition: number | undefined;
    state.doc.descendants((node, position) => {
      if (node.isText && node.text === "A") {
        textPosition = position;
        return false;
      }

      return true;
    });
    if (textPosition === undefined) throw new Error("expected cell text");

    state = state.applyTransaction(
      state.tr.insertText("X", textPosition + 1),
    ).state;
    expect(document.body).toContain("AX");

    let undoTransaction: Transaction | undefined;
    expect(
      undo(state, (transaction) => {
        undoTransaction = transaction;
      }),
    ).toBe(true);
    if (!undoTransaction) throw new Error("expected undo transaction");

    state.applyTransaction(undoTransaction);
    expect(document.body).not.toContain("AX");
  });

  it("inserts a divider as a structural block", () => {
    let document = createRichEditorAutomergeDocument(
      JSON.stringify({
        type: "doc",
        content: [
          {
            type: "paragraph",
            content: [{ type: "text", text: "A" }],
          },
          {
            type: "paragraph",
            content: [{ type: "text", text: "B" }],
          },
        ],
      }),
    );
    const handle: DocHandle<RichEditorAutomergeDocument> = {
      doc: () => document,
      change: (change) => {
        document = Automerge.change(document, change);
      },
      on: () => {},
      off: () => {},
    };
    const adapter = createSchemaAdapter(richEditorCollaborationExtensions);
    const pmDocument = pmDocFromSpans(
      adapter,
      Automerge.spans(document, ["body"]),
    );
    const state = EditorState.create({
      schema: adapter.schema,
      doc: pmDocument,
      plugins: [
        syncPlugin({
          adapter,
          handle,
          path: ["body"],
        }),
      ],
    });
    const firstParagraph = state.doc.firstChild;
    if (!firstParagraph) throw new Error("expected first paragraph");

    state.applyTransaction(
      state.tr.insert(
        firstParagraph.nodeSize,
        adapter.schema.nodes.horizontalRule.create({ isAmgBlock: true }),
      ),
    );

    const spans = Automerge.spans(document, ["body"]);
    expect(
      spans.some(span =>
        span.type === "block"
        && Automerge.isImmutableString(span.value.type)
        && span.value.type.val === "horizontal-rule",
      ),
    ).toBe(true);
    expect(
      pmDocFromSpans(adapter, spans).toJSON(),
    ).toMatchObject({
      content: [
        { type: "paragraph" },
        { type: "horizontalRule" },
        { type: "paragraph" },
      ],
    });
  });

  it("synchronizes text typed after a newly inserted divider", () => {
    let document = createRichEditorAutomergeDocument(
      JSON.stringify({
        type: "doc",
        content: [
          {
            type: "paragraph",
            content: [{ type: "text", text: "Before" }],
          },
        ],
      }),
    );
    const targetSchema = getSchema(richEditorCollaborationExtensions);
    const adapter = createSchemaAdapter(
      richEditorCollaborationExtensions,
      targetSchema,
    );
    const handle: DocHandle<RichEditorAutomergeDocument> = {
      doc: () => document,
      change: (change) => {
        document = Automerge.change(document, (draft) => {
          change(draft);
          normalizeStructuralLeafFollowers(draft, adapter);
        });
      },
      on: () => {},
      off: () => {},
    };
    const initialAdapter = createSchemaAdapter(
      richEditorCollaborationExtensions,
    );
    const pmDocument = targetSchema.nodeFromJSON(
      pmDocFromSpans(
        initialAdapter,
        Automerge.spans(document, ["body"]),
      ).toJSON(),
    );
    let state = EditorState.create({
      schema: targetSchema,
      doc: pmDocument,
      plugins: [
        syncPlugin({
          adapter,
          handle,
          path: ["body"],
        }),
      ],
    });
    const divider = targetSchema.nodes.horizontalRule.create();
    const paragraph = targetSchema.nodes.paragraph.create();
    state = state.applyTransaction(
      state.tr.insert(
        state.doc.content.size,
        Fragment.fromArray([divider, paragraph]),
      ),
    ).state;
    expect(
      Automerge.spans(document, ["body"])
        .filter(span => span.type === "block")
        .map(span => Automerge.isImmutableString(span.value.type)
          ? span.value.type.val
          : span.value.type),
    ).toEqual([
      "paragraph",
      "horizontal-rule",
      "paragraph",
    ]);

    const lastParagraphPosition = state.doc.content.size - paragraph.nodeSize;
    let insertionError: unknown;
    try {
      state.applyTransaction(
        state.tr.insertText("After", lastParagraphPosition + 1),
      );
    } catch (error) {
      insertionError = error;
    }

    const spans = Automerge.spans(document, ["body"]);
    expect(
      spans
        .filter(span => span.type === "block")
        .map(span => Automerge.isImmutableString(span.value.type)
          ? span.value.type.val
          : span.value.type),
    ).toEqual([
      "paragraph",
      "horizontal-rule",
      "paragraph",
    ]);
    expect(
      spans
        .filter(span => span.type === "block")
        .map(span => ({
          type: automergeString(span.value.type),
          isEmbed: span.value.isEmbed,
          parents: automergeStringArray(span.value.parents),
        })),
    ).toEqual([
      { type: "paragraph", isEmbed: false, parents: [] },
      { type: "horizontal-rule", isEmbed: false, parents: [] },
      { type: "paragraph", isEmbed: false, parents: [] },
    ]);
    expect(
      spans
        .filter(span => span.type === "text")
        .map(span => span.value)
        .join(""),
    ).toBe("BeforeAfter");
    expect(
      spans.map(span => span.type === "text"
        ? `text:${span.value}`
        : `block:${automergeString(span.value.type)}`),
    ).toEqual([
      "block:paragraph",
      "text:Before",
      "block:horizontal-rule",
      "block:paragraph",
      "text:After",
    ]);
    expect(insertionError).toBeUndefined();

    const remoteDocument = pmDocFromSpans(initialAdapter, spans);
    expect(remoteDocument.textContent).toBe("BeforeAfter");
    expect(remoteDocument.child(1).type.name).toBe("horizontalRule");
    expect(remoteDocument.child(2).textContent).toBe("After");
  });

  it("preserves Mermaid code-block language", () => {
    const document = createRichEditorAutomergeDocument(
      JSON.stringify({
        type: "doc",
        content: [
          {
            type: "codeBlock",
            attrs: { language: "mermaid" },
            content: [{ type: "text", text: "graph TD; A-->B" }],
          },
        ],
      }),
    );
    const spans = Automerge.spans(document, ["body"]);
    const block = spans.find(span => span.type === "block");
    if (block?.type !== "block") throw new Error("expected block span");
    const attrs = block.value.attrs;
    if (
      attrs === null
      || typeof attrs !== "object"
      || !("language" in attrs)
    ) {
      throw new Error("expected code-block language");
    }
    expect(attrs.language).toBe("mermaid");

    const handle: DocHandle<RichEditorAutomergeDocument> = {
      doc: () => document,
      change: () => {},
      on: () => {},
      off: () => {},
    };
    const roundTrip = richEditorAutomergeContent(
      handle,
      richEditorCollaborationExtensions,
    ) as {
      content: Array<{
        type: string;
        attrs: { language: string | null };
      }>;
    };
    expect(roundTrip.content[0]).toMatchObject({
      type: "codeBlock",
      attrs: { language: "mermaid" },
    });
  });
});

function tableDocumentJSON(): string {
  return JSON.stringify({
    type: "doc",
    content: [
      {
        type: "table",
        content: [
          {
            type: "tableRow",
            content: [
              tableCell("A"),
              tableCell("B"),
            ],
          },
          {
            type: "tableRow",
            content: [
              tableCell("C"),
              tableCell("D"),
            ],
          },
        ],
      },
    ],
  });
}

function automergeString(value: unknown): string {
  if (typeof value === "string") return value;
  if (Automerge.isImmutableString(value)) return value.val;
  throw new Error("expected Automerge string");
}

function automergeStringArray(value: unknown): string[] {
  if (!Array.isArray(value)) throw new Error("expected Automerge string array");
  return value.map(automergeString);
}

function tableCell(text: string) {
  return {
    type: "tableCell",
    attrs: {
      colspan: 1,
      rowspan: 1,
      colwidth: null,
    },
    content: [
      {
        type: "paragraph",
        content: [{ type: "text", text }],
      },
    ],
  };
}

function insertTableRow(
  initial: Automerge.Doc<RichEditorAutomergeDocument>,
  text: string,
): Automerge.Doc<RichEditorAutomergeDocument> {
  let document = initial;
  const handle: DocHandle<RichEditorAutomergeDocument> = {
    doc: () => document,
    change: (change) => {
      document = Automerge.change(document, change);
    },
    on: () => {},
    off: () => {},
  };
  const adapter = createSchemaAdapter(richEditorCollaborationExtensions);
  const pmDocument = pmDocFromSpans(
    adapter,
    Automerge.spans(document, ["body"]),
  );
  const state = EditorState.create({
    schema: adapter.schema,
    doc: pmDocument,
    plugins: [
      syncPlugin({
        adapter,
        handle,
        path: ["body"],
      }),
    ],
  });
  const paragraph = adapter.schema.nodes.paragraph.create(
    null,
    adapter.schema.text(text),
  );
  const cell = adapter.schema.nodes.tableCell.create(
    {
      isAmgBlock: true,
      colspan: 1,
      rowspan: 1,
      colwidth: null,
    },
    paragraph,
  );
  const row = adapter.schema.nodes.tableRow.create(
    { isAmgBlock: true },
    [cell],
  );
  const table = state.doc.firstChild;
  if (!table) throw new Error("expected table node");

  state.applyTransaction(state.tr.insert(table.nodeSize - 1, row));

  return document;
}
