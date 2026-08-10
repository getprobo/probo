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

import { Buffer } from "node:buffer";
import { existsSync, readFileSync, writeFileSync } from "node:fs";
import process from "node:process";
import { fileURLToPath } from "node:url";

import * as Automerge from "@automerge/automerge";
import { pmDocFromSpans } from "@automerge/prosemirror";
import { describe, expect, it } from "vitest";

import {
  createRichEditorAutomergeDocument,
  createSchemaAdapter,
} from "./collaboration";
import { richEditorCollaborationExtensions } from "./RichEditor";

// The Go renderer in pkg/automerge/prosemirror is a second implementation of the
// span -> ProseMirror-document conversion performed by @automerge/prosemirror's
// pmDocFromSpans. This file is the oracle: for a corpus of realistic documents it
// records the Automerge document bytes together with the canonical ProseMirror
// JSON the official library produces, so the Go differential test can assert byte
// parity against upstream. Run with GEN_PROSEMIRROR_PARITY=1 to (re)write the
// fixture; otherwise the test guards the committed fixture against frontend drift.

type CorpusEntry = { name: string; doc: unknown };

type FixtureEntry = { name: string; document: string; expected: unknown };

const fixturePath = fileURLToPath(
  new URL(
    "../../../../pkg/automerge/prosemirror/testdata/upstream-render.json",
    import.meta.url,
  ),
);

function tableCell(text: string): unknown {
  return {
    type: "tableCell",
    attrs: { colspan: 1, rowspan: 1, colwidth: null },
    content: [{ type: "paragraph", content: [{ type: "text", text }] }],
  };
}

function tableHeader(text: string): unknown {
  return {
    type: "tableHeader",
    attrs: { colspan: 1, rowspan: 1, colwidth: null },
    content: [{ type: "paragraph", content: [{ type: "text", text }] }],
  };
}

const corpus: CorpusEntry[] = [
  {
    name: "empty-document",
    doc: { type: "doc", content: [{ type: "paragraph" }] },
  },
  {
    name: "paragraph-plain",
    doc: {
      type: "doc",
      content: [{ type: "paragraph", content: [{ type: "text", text: "Hello world" }] }],
    },
  },
  {
    name: "heading-with-marks",
    doc: {
      type: "doc",
      content: [
        {
          type: "heading",
          attrs: { level: 2 },
          content: [{ type: "text", text: "Policy", marks: [{ type: "bold" }] }],
        },
      ],
    },
  },
  {
    name: "adjacent-mark-runs",
    doc: {
      type: "doc",
      content: [
        {
          type: "paragraph",
          content: [
            { type: "text", text: "A", marks: [{ type: "bold" }, { type: "italic" }] },
            { type: "text", text: "B", marks: [{ type: "italic" }] },
            { type: "text", text: "C" },
          ],
        },
      ],
    },
  },
  {
    name: "all-inline-marks",
    doc: {
      type: "doc",
      content: [
        {
          type: "paragraph",
          content: [
            { type: "text", text: "b", marks: [{ type: "bold" }] },
            { type: "text", text: "i", marks: [{ type: "italic" }] },
            { type: "text", text: "s", marks: [{ type: "strike" }] },
            { type: "text", text: "u", marks: [{ type: "underline" }] },
            { type: "text", text: "c", marks: [{ type: "code" }] },
          ],
        },
      ],
    },
  },
  {
    name: "stacked-marks",
    doc: {
      type: "doc",
      content: [
        {
          type: "paragraph",
          content: [
            {
              type: "text",
              text: "stack",
              marks: [
                { type: "underline" },
                { type: "strike" },
                { type: "italic" },
                { type: "bold" },
                { type: "code" },
              ],
            },
          ],
        },
      ],
    },
  },
  {
    name: "link-with-bold",
    doc: {
      type: "doc",
      content: [
        {
          type: "paragraph",
          content: [
            {
              type: "text",
              text: "site",
              marks: [
                { type: "link", attrs: { href: "https://example.com", title: null } },
                { type: "bold" },
              ],
            },
          ],
        },
      ],
    },
  },
  {
    name: "link-mark",
    doc: {
      type: "doc",
      content: [
        {
          type: "paragraph",
          content: [
            { type: "text", text: "Read " },
            {
              type: "text",
              text: "more",
              marks: [{ type: "link", attrs: { href: "https://example.com", title: "Example" } }],
            },
          ],
        },
      ],
    },
  },
  {
    name: "blockquote",
    doc: {
      type: "doc",
      content: [
        { type: "blockquote", content: [{ type: "paragraph", content: [{ type: "text", text: "Quoted" }] }] },
        { type: "paragraph", content: [{ type: "text", text: "After" }] },
      ],
    },
  },
  {
    name: "code-block-language",
    doc: {
      type: "doc",
      content: [
        { type: "codeBlock", attrs: { language: "mermaid" }, content: [{ type: "text", text: "graph TD; A-->B" }] },
      ],
    },
  },
  {
    name: "code-block-no-language",
    doc: {
      type: "doc",
      content: [
        { type: "codeBlock", attrs: { language: null }, content: [{ type: "text", text: "plain" }] },
      ],
    },
  },
  {
    name: "horizontal-rule-between-paragraphs",
    doc: {
      type: "doc",
      content: [
        { type: "paragraph", content: [{ type: "text", text: "Above" }] },
        { type: "horizontalRule" },
        { type: "paragraph", content: [{ type: "text", text: "Below" }] },
      ],
    },
  },
  {
    name: "hard-break-inside-paragraph",
    doc: {
      type: "doc",
      content: [
        {
          type: "paragraph",
          content: [
            { type: "text", text: "A" },
            { type: "hardBreak" },
            { type: "text", text: "B" },
          ],
        },
      ],
    },
  },
  {
    name: "bullet-list-consecutive-items",
    doc: {
      type: "doc",
      content: [
        {
          type: "bulletList",
          content: [
            { type: "listItem", content: [{ type: "paragraph", content: [{ type: "text", text: "One" }] }] },
            { type: "listItem", content: [{ type: "paragraph", content: [{ type: "text", text: "Two" }] }] },
          ],
        },
      ],
    },
  },
  {
    name: "ordered-list",
    doc: {
      type: "doc",
      content: [
        {
          type: "orderedList",
          content: [
            { type: "listItem", content: [{ type: "paragraph", content: [{ type: "text", text: "First" }] }] },
            { type: "listItem", content: [{ type: "paragraph", content: [{ type: "text", text: "Second" }] }] },
          ],
        },
      ],
    },
  },
  {
    name: "nested-bullet-list",
    doc: {
      type: "doc",
      content: [
        {
          type: "bulletList",
          content: [
            {
              type: "listItem",
              content: [
                { type: "paragraph", content: [{ type: "text", text: "Outer" }] },
                {
                  type: "bulletList",
                  content: [
                    { type: "listItem", content: [{ type: "paragraph", content: [{ type: "text", text: "Inner" }] }] },
                  ],
                },
              ],
            },
          ],
        },
      ],
    },
  },
  {
    name: "table-header-and-cells",
    doc: {
      type: "doc",
      content: [
        {
          type: "table",
          content: [
            { type: "tableRow", content: [tableHeader("H1"), tableHeader("H2")] },
            { type: "tableRow", content: [tableCell("A"), tableCell("B")] },
          ],
        },
      ],
    },
  },
  {
    name: "mixed-document",
    doc: {
      type: "doc",
      content: [
        { type: "heading", attrs: { level: 1 }, content: [{ type: "text", text: "Title" }] },
        { type: "paragraph", content: [{ type: "text", text: "Intro" }] },
        {
          type: "bulletList",
          content: [
            { type: "listItem", content: [{ type: "paragraph", content: [{ type: "text", text: "Item", marks: [{ type: "bold" }] }] }] },
          ],
        },
        { type: "horizontalRule" },
        { type: "codeBlock", attrs: { language: null }, content: [{ type: "text", text: "code()" }] },
      ],
    },
  },
];

function canonicalAttrs(
  type: string,
  attrs: Record<string, unknown> | undefined,
): Record<string, unknown> | undefined {
  const source = attrs ?? {};
  switch (type) {
    case "heading":
      return { level: source.level };
    case "codeBlock":
      return { language: source.language ?? null };
    case "tableCell":
    case "tableHeader":
      return {
        colspan: source.colspan,
        rowspan: source.rowspan,
        colwidth: source.colwidth ?? null,
      };
    default:
      return undefined;
  }
}

function canonicalMark(mark: {
  type: string;
  attrs?: Record<string, unknown>;
}): unknown {
  if (mark.type === "link") {
    const attrs = mark.attrs ?? {};
    return {
      type: "link",
      attrs: { href: attrs.href ?? "", title: attrs.title ?? null },
    };
  }
  return { type: mark.type };
}

// canonical projects the official ProseMirror JSON onto the shape the Go renderer
// targets: editor-only attributes (isAmgBlock, unknownAttrs, align) and display-only
// link attributes (target, rel, class) are dropped, and empty content/marks/attrs are
// omitted to mirror the Go struct's omitempty encoding.
function canonical(node: {
  type: string;
  attrs?: Record<string, unknown>;
  text?: string;
  marks?: Array<{ type: string; attrs?: Record<string, unknown> }>;
  content?: unknown[];
}): unknown {
  const out: Record<string, unknown> = { type: node.type };
  const attrs = canonicalAttrs(node.type, node.attrs);
  if (attrs !== undefined) out.attrs = attrs;
  if (typeof node.text === "string") out.text = node.text;
  if (node.marks && node.marks.length > 0) out.marks = node.marks.map(canonicalMark);
  if (node.content && node.content.length > 0) {
    out.content = node.content.map(child => canonical(child as never));
  }
  return out;
}

function build(): FixtureEntry[] {
  const adapter = createSchemaAdapter(richEditorCollaborationExtensions);
  return corpus.map(({ name, doc }) => {
    const document = createRichEditorAutomergeDocument(
      JSON.stringify(doc),
      richEditorCollaborationExtensions,
    );
    const spans = Automerge.spans(document, ["body"]);
    const rendered: unknown = pmDocFromSpans(adapter, spans).toJSON();
    return {
      name,
      document: Buffer.from(Automerge.save(document)).toString("base64"),
      expected: canonical(rendered as never),
    };
  });
}

describe("ProseMirror render parity fixture", () => {
  it("matches the committed Go differential fixture", () => {
    const entries = build();

    if (process.env.GEN_PROSEMIRROR_PARITY) {
      writeFileSync(fixturePath, JSON.stringify(entries, null, 2) + "\n");
      return;
    }

    expect(existsSync(fixturePath)).toBe(true);
    const committed = JSON.parse(readFileSync(fixturePath, "utf8")) as FixtureEntry[];

    expect(entries.map(entry => entry.name)).toEqual(
      committed.map(entry => entry.name),
    );
    for (const entry of entries) {
      const match = committed.find(candidate => candidate.name === entry.name);
      expect(match, entry.name).toBeDefined();
      expect(entry.expected, entry.name).toEqual(match!.expected);
    }
  });
});
