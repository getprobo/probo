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

import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";

import { describe, expect, it } from "vitest";

import { createSchemaAdapter } from "./collaboration";
import { richEditorCollaborationExtensions } from "./RichEditor";

// This test holds the frontend schema adapter to the shared ProseMirror <->
// Automerge ledger. Its Go counterpart (TestSchemaMappingLedger) holds the Go
// renderer to the same file, so the two implementations of the mapping cannot
// drift apart — adding or renaming a block or mark on one side without the other
// fails a test.

type Ledger = {
  blocks: Array<{
    automerge: string;
    prosemirror: string;
    outer?: string;
    isEmbed?: boolean;
  }>;
  marks: Array<{ automerge: string; prosemirror: string }>;
};

const ledgerPath = fileURLToPath(
  new URL(
    "../../../../pkg/automerge/prosemirror/testdata/schema-mapping.json",
    import.meta.url,
  ),
);

function loadLedger(): Ledger {
  return JSON.parse(readFileSync(ledgerPath, "utf8")) as Ledger;
}

describe("ProseMirror schema mapping parity", () => {
  it("keeps the frontend adapter aligned with the shared ledger", () => {
    const ledger = loadLedger();
    const adapter = createSchemaAdapter(richEditorCollaborationExtensions);

    const adapterBlocks = adapter.nodeMappings
      .map(mapping => ({
        automerge: mapping.blockName,
        prosemirror: mapping.content.name,
        outer: mapping.outer?.name ?? undefined,
        isEmbed: mapping.isEmbed ?? false,
      }))
      .sort((a, b) => a.automerge.localeCompare(b.automerge));

    const ledgerBlocks = ledger.blocks
      .map(block => ({
        automerge: block.automerge,
        prosemirror: block.prosemirror,
        outer: block.outer ?? undefined,
        isEmbed: block.isEmbed ?? false,
      }))
      .sort((a, b) => a.automerge.localeCompare(b.automerge));

    expect(adapterBlocks).toEqual(ledgerBlocks);
  });

  it("keeps mark names aligned with the shared ledger", () => {
    const ledger = loadLedger();
    const adapter = createSchemaAdapter(richEditorCollaborationExtensions);

    const adapterMarks = adapter.markMappings
      .map(mapping => ({
        automerge: mapping.automergeMarkName,
        prosemirror: mapping.prosemirrorMark.name,
      }))
      .sort((a, b) => a.automerge.localeCompare(b.automerge));

    const ledgerMarks = [...ledger.marks].sort((a, b) =>
      a.automerge.localeCompare(b.automerge),
    );

    expect(adapterMarks).toEqual(ledgerMarks);
  });

  it("keeps mark render order aligned with the ProseMirror schema rank", () => {
    const ledger = loadLedger();
    const adapter = createSchemaAdapter(richEditorCollaborationExtensions);

    const knownMarks = new Set(ledger.marks.map(mark => mark.prosemirror));
    const schemaRank: string[] = [];
    adapter.schema.spec.marks.forEach((name: string) => {
      if (knownMarks.has(name)) schemaRank.push(name);
    });

    expect(schemaRank).toEqual(ledger.marks.map(mark => mark.prosemirror));
  });
});
