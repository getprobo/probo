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

import { describe, expect, it } from "vitest";

import documentCommentHistoryListFragment from "#/__generated__/core/DocumentCommentHistoryList_document.graphql";
import documentReviewPanelFragment from "#/__generated__/core/DocumentReviewPanel_threads.graphql";

import { documentCommentHistoryConnectionId, documentOpenThreadsConnectionId } from "./documentThreadConnections";

function connectionStorageKey(node: unknown, name: string): string {
  const found = findSelection(node, name);
  if (found == null || typeof found.storageKey !== "string") {
    throw new Error(`expected a connection selection named ${name}`);
  }
  return found.storageKey;
}

type Selection = { name?: unknown; storageKey?: unknown; selections?: unknown };

function findSelection(node: unknown, name: string): Selection | null {
  if (typeof node !== "object" || node == null) return null;
  const record = node as Selection;
  if (record.name === name) return record;
  if (!Array.isArray(record.selections)) return null;
  for (const selection of record.selections) {
    const found = findSelection(selection, name);
    if (found != null) return found;
  }
  return null;
}

describe("documentThreadConnections", () => {
  it("matches the open threads connection in DocumentReviewPanel", () => {
    const storageKey = connectionStorageKey(
      documentReviewPanelFragment,
      "__DocumentReviewPanel_commentThreads_connection",
    );

    expect(documentOpenThreadsConnectionId("doc")).toBe(`client:doc:${storageKey}`);
  });

  it("matches the history connection in DocumentCommentHistoryList", () => {
    const storageKey = connectionStorageKey(
      documentCommentHistoryListFragment,
      "__DocumentCommentHistoryList_commentThreads_connection",
    );

    expect(documentCommentHistoryConnectionId("doc")).toBe(`client:doc:${storageKey}`);
  });
});
