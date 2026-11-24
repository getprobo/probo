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

import { describe, expect, it, vi } from "vitest";

import { type RecordedSuggestionChange, suggestionChange, suggestionChanges } from "./suggestionChanges";

// Vitest has no Relay Babel transform, and these tests never read the fragment.
vi.mock("relay-runtime", async importOriginal => ({
  ...await importOriginal<typeof import("relay-runtime")>(),
  graphql: () => null,
}));

function recorded(change: Partial<RecordedSuggestionChange> & Pick<RecordedSuggestionChange, "kind">): RecordedSuggestionChange {
  return {
    deleted: null,
    inserted: null,
    mark: null,
    formatAction: null,
    href: null,
    previousHref: null,
    before: null,
    after: null,
    node: null,
    nodeAction: null,
    ...change,
  };
}

describe("suggestionChange", () => {
  it("maps a text change", () => {
    expect(suggestionChange(recorded({ kind: "TEXT", deleted: "old", inserted: "new" }))).toEqual({
      kind: "text",
      deleted: "old",
      inserted: "new",
    });
  });

  it("maps a missing side of a text change to an empty string", () => {
    expect(suggestionChange(recorded({ kind: "TEXT", inserted: "new" }))).toEqual({
      kind: "text",
      deleted: "",
      inserted: "new",
    });
  });

  it("maps a link change with its addresses", () => {
    expect(suggestionChange(recorded({
      kind: "FORMAT",
      mark: "LINK",
      formatAction: "ADD",
      href: "https://example.com/new",
      previousHref: "https://example.com/old",
    }))).toEqual({
      kind: "format",
      style: "link",
      action: "add",
      href: "https://example.com/new",
      previousHref: "https://example.com/old",
    });
  });

  it("maps a format removal", () => {
    expect(suggestionChange(recorded({ kind: "FORMAT", mark: "BOLD", formatAction: "REMOVE" }))).toEqual({
      kind: "format",
      style: "bold",
      action: "remove",
      href: null,
      previousHref: null,
    });
  });

  it("maps a heading turned into a paragraph", () => {
    expect(suggestionChange(recorded({
      kind: "BLOCK",
      before: { kind: "HEADING", level: 2 },
      after: { kind: "PARAGRAPH", level: null },
    }))).toEqual({
      kind: "block",
      before: { kind: "heading", level: 2 },
      after: { kind: "paragraph", level: null },
    });
  });

  it("maps an inserted row", () => {
    expect(suggestionChange(recorded({ kind: "NODE", node: "ROW", nodeAction: "INSERT" }))).toEqual({
      kind: "node",
      node: "row",
      action: "insert",
    });
  });

  it("maps a deleted line", () => {
    expect(suggestionChange(recorded({ kind: "NODE", node: "LINE", nodeAction: "DELETE" }))).toEqual({
      kind: "node",
      node: "line",
      action: "delete",
    });
  });

  it("falls back to other for an unknown block or node", () => {
    expect(suggestionChange(recorded({
      kind: "BLOCK",
      before: { kind: "FIGURE" as never, level: null },
      after: { kind: "PARAGRAPH", level: null },
    }))).toEqual({
      kind: "block",
      before: { kind: "other", level: null },
      after: { kind: "paragraph", level: null },
    });
    expect(suggestionChange(recorded({ kind: "NODE", node: "VIDEO" as never, nodeAction: "INSERT" }))).toEqual({
      kind: "node",
      node: "other",
      action: "insert",
    });
  });

  it("skips a change it cannot draw", () => {
    expect(suggestionChange(recorded({ kind: "MOVE" as never }))).toBeNull();
    expect(suggestionChange(recorded({ kind: "FORMAT", mark: "HIGHLIGHT" as never, formatAction: "ADD" }))).toBeNull();
    expect(suggestionChange(recorded({ kind: "FORMAT", mark: "BOLD" }))).toBeNull();
    expect(suggestionChange(recorded({ kind: "BLOCK", after: { kind: "PARAGRAPH", level: null } }))).toBeNull();
    expect(suggestionChange(recorded({ kind: "NODE", node: "ROW" }))).toBeNull();
  });
});

describe("suggestionChanges", () => {
  it("keeps the order and leaves out changes it cannot draw", () => {
    expect(suggestionChanges([
      recorded({ kind: "TEXT", inserted: "a" }),
      recorded({ kind: "MOVE" as never }),
      recorded({ kind: "NODE", node: "DIVIDER", nodeAction: "DELETE" }),
    ])).toEqual([
      { kind: "text", deleted: "", inserted: "a" },
      { kind: "node", node: "divider", action: "delete" },
    ]);
  });
});
