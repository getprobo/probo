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

import {
  documentListSearch,
  normalizeDocumentSearch,
  readDocumentColumn,
  readDocumentMode,
  withDocumentColumn,
  withDocumentMode,
  withDocumentVersion,
} from "./documentSearch";
import type { DocumentMode } from "./useDocumentModes";

const allModes: DocumentMode[] = ["viewing", "editing", "suggesting"];

function search(value: string): URLSearchParams {
  return new URLSearchParams(value);
}

describe("readDocumentMode", () => {
  it("defaults to viewing", () => {
    expect(readDocumentMode(search(""))).toBe("viewing");
    expect(readDocumentMode(search("mode=unknown"))).toBe("viewing");
  });

  it("reads editing and suggesting", () => {
    expect(readDocumentMode(search("mode=editing"))).toBe("editing");
    expect(readDocumentMode(search("mode=suggesting"))).toBe("suggesting");
  });

  it("stays viewing beside a version", () => {
    expect(readDocumentMode(search("version=v1&mode=editing"))).toBe("viewing");
    expect(readDocumentMode(search("version=v1&mode=suggesting"))).toBe("viewing");
  });
});

describe("readDocumentColumn", () => {
  it("defaults to the open comments", () => {
    expect(readDocumentColumn(search(""))).toBe("open");
    expect(readDocumentColumn(search("view=unknown"))).toBe("open");
  });

  it("opens the versions for a version without a view", () => {
    expect(readDocumentColumn(search("version=v1"))).toBe("versions");
  });

  it("keeps the named view beside a version", () => {
    expect(readDocumentColumn(search("version=v1&view=open"))).toBe("open");
    expect(readDocumentColumn(search("version=v1&view=all"))).toBe("all");
  });
});

describe("normalizeDocumentSearch", () => {
  it("drops an editing or suggesting mode beside a version", () => {
    expect(normalizeDocumentSearch(search("version=v1&mode=editing&view=open"), allModes).toString())
      .toBe("version=v1&view=open");
    expect(normalizeDocumentSearch(search("version=v1&mode=suggesting"), allModes).toString())
      .toBe("version=v1");
  });

  it("drops a mode the viewer may not use", () => {
    expect(normalizeDocumentSearch(search("mode=editing&q=a"), ["viewing", "suggesting"]).toString())
      .toBe("q=a");
  });

  it("keeps an allowed mode", () => {
    expect(normalizeDocumentSearch(search("mode=suggesting&view=all"), allModes).toString())
      .toBe("mode=suggesting&view=all");
  });

  it("drops defaults and unknown values", () => {
    expect(normalizeDocumentSearch(search("mode=viewing&view=open"), allModes).toString()).toBe("");
    expect(normalizeDocumentSearch(search("mode=unknown&view=unknown"), allModes).toString()).toBe("");
  });
});

describe("withDocumentMode", () => {
  it("drops the version and closes the versions list when editing", () => {
    expect(withDocumentMode(search("version=v1"), "editing").toString()).toBe("mode=editing");
    expect(withDocumentMode(search("view=versions"), "suggesting").toString()).toBe("mode=suggesting");
  });

  it("keeps the comment view when editing", () => {
    expect(withDocumentMode(search("view=all"), "editing").toString()).toBe("view=all&mode=editing");
  });

  it("omits viewing", () => {
    expect(withDocumentMode(search("mode=editing"), "viewing").toString()).toBe("");
  });
});

describe("withDocumentVersion", () => {
  it("forces viewing", () => {
    expect(withDocumentVersion(search("mode=editing&view=versions"), "v1").toString()).toBe("version=v1");
  });

  it("keeps the versions open when showing the latest", () => {
    expect(withDocumentVersion(search("version=v1"), null).toString()).toBe("view=versions");
  });

  it("keeps a closed list closed when showing the latest", () => {
    expect(withDocumentVersion(search("version=v1&view=open"), null).toString()).toBe("");
  });
});

describe("withDocumentColumn", () => {
  it("omits the default column", () => {
    expect(withDocumentColumn(search("view=all"), "open").toString()).toBe("");
    expect(withDocumentColumn(search("version=v1&view=open"), "versions").toString()).toBe("version=v1");
  });

  it("writes open beside a version", () => {
    expect(withDocumentColumn(search("version=v1"), "open").toString()).toBe("version=v1&view=open");
  });
});

describe("documentListSearch", () => {
  it("keeps only the list params", () => {
    expect(documentListSearch("?q=a&version=v1&mode=editing&view=all")).toBe("?q=a");
    expect(documentListSearch("?version=v1&mode=editing&view=all")).toBe("");
  });
});
