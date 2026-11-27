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

import { evidencePreviewFormat } from "./evidencePreviewFormat";
import { formatJSONPreview } from "./formatJSONPreview";

describe("evidencePreviewFormat", () => {
  it("previews JSON evidence by MIME type or extension", () => {
    expect(evidencePreviewFormat("application/json", "report.json")).toBe("json");
    expect(evidencePreviewFormat("application/json; charset=utf-8", "report.json")).toBe("json");
    expect(evidencePreviewFormat("APPLICATION/JSON", "report.json")).toBe("json");
    expect(evidencePreviewFormat("text/json", "report.json")).toBe("json");
    expect(evidencePreviewFormat("application/octet-stream", "report.JSON")).toBe("json");
    expect(evidencePreviewFormat("text/plain", "notes.json")).toBe("json");
  });

  it("keeps the existing preview kinds", () => {
    expect(evidencePreviewFormat("text/uri-list", "link.uri")).toBe("uri");
    expect(evidencePreviewFormat("text/uri", "link.uri")).toBe("uri");
    expect(evidencePreviewFormat("image/png", "photo.png")).toBe("image");
    expect(evidencePreviewFormat("application/pdf", "policy.pdf")).toBe("pdf");
    expect(evidencePreviewFormat("text/csv", "rows.csv")).toBe("csv");
    expect(evidencePreviewFormat("application/csv", "rows.csv")).toBe("csv");
    expect(evidencePreviewFormat("application/octet-stream", "rows.csv")).toBe("csv");
    expect(evidencePreviewFormat("text/markdown", "readme.md")).toBe("markdown");
    expect(evidencePreviewFormat("text/plain", "readme.md")).toBe("markdown");
    expect(evidencePreviewFormat("text/plain", "notes.txt")).toBe("text");
    expect(evidencePreviewFormat("text/plain", "app.log")).toBe("text");
    expect(evidencePreviewFormat("application/zip", "archive.zip")).toBe("unavailable");
    expect(evidencePreviewFormat(null, null)).toBe("unavailable");
  });
});

describe("formatJSONPreview", () => {
  it("pretty-prints valid JSON", () => {
    expect(formatJSONPreview("{\"b\":1,\"a\":[true,null]}")).toBe(
      "{\n  \"b\": 1,\n  \"a\": [\n    true,\n    null\n  ]\n}",
    );
    expect(formatJSONPreview(" null ")).toBe("null");
    expect(formatJSONPreview("\uFEFF{\"ok\":true}")).toBe("{\n  \"ok\": true\n}");
  });

  it("shows invalid JSON unchanged", () => {
    const invalid = "{not json";
    expect(formatJSONPreview(invalid)).toBe(invalid);
    expect(formatJSONPreview("")).toBe("");
  });
});
